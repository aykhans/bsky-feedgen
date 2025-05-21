package consumer

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/logger"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb/collections"
	"github.com/aykhans/bsky-feedgen/pkg/types"
	"github.com/aykhans/bsky-feedgen/pkg/utils"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/events/schedulers/parallel"
	lexutil "github.com/bluesky-social/indigo/lex/util"

	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/gorilla/websocket"
)

type CallbackData struct {
	Sequence  int64
	DID       syntax.DID
	RecordKey syntax.RecordKey
	Post      bsky.FeedPost
}

type CallbackFunc func(int64, syntax.DID, syntax.RecordKey, bsky.FeedPost)

func RunFirehoseConsumer(
	ctx context.Context,
	relayHost string,
	callbackFunc CallbackFunc,
	cursor *int64,
) error {
	dialer := websocket.DefaultDialer
	u, err := url.Parse(relayHost)
	if err != nil {
		return fmt.Errorf("invalid relayHost URI: %w", err)
	}

	u.Path = "xrpc/com.atproto.sync.subscribeRepos"
	if cursor != nil {
		q := url.Values{}
		q.Set("cursor", strconv.FormatInt(*cursor, 10))
		u.RawQuery = q.Encode()
	}

	logger.Log.Info("subscribing to repo event stream", "upstream", relayHost)
	con, _, err := dialer.Dial(u.String(), http.Header{
		"User-Agent": []string{"Firehose-Consumer"},
	})
	if err != nil {
		return fmt.Errorf("subscribing to firehose failed (dialing): %w", err)
	}

	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *comatproto.SyncSubscribeRepos_Commit) error {
			return HandleRepoCommit(ctx, evt, callbackFunc)
		},
	}

	var scheduler events.Scheduler
	parallelism := 8
	scheduler = parallel.NewScheduler(
		parallelism,
		100_000,
		relayHost,
		rsc.EventHandler,
	)
	logger.Log.Info("firehose scheduler configured", "workers", parallelism)

	err = events.HandleRepoStream(ctx, con, scheduler, logger.Log)
	if err != nil {
		return fmt.Errorf("repoStream error: %v", err)
	}
	return nil
}

func HandleRepoCommit(
	ctx context.Context,
	evt *comatproto.SyncSubscribeRepos_Commit,
	postCallback CallbackFunc,
) error {
	localLogger := logger.Log.With("event", "commit", "did", evt.Repo, "rev", evt.Rev, "seq", evt.Seq)

	if evt.TooBig {
		localLogger.Warn("skipping tooBig events for now")
		return nil
	}

	did, err := syntax.ParseDID(evt.Repo)
	if err != nil {
		localLogger.Error("bad DID syntax in event", "err", err)
		return nil
	}

	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
	if err != nil {
		localLogger.Error("failed to read repo from car", "err", err)
		return nil
	}

	for _, op := range evt.Ops {
		localLogger = localLogger.With("eventKind", op.Action, "path", op.Path)
		collection, rkey, err := syntax.ParseRepoPath(op.Path)
		if err != nil {
			localLogger.Error("invalid path in repo op")
			return nil
		}

		ek := repomgr.EventKind(op.Action)
		switch ek {
		case repomgr.EvtKindCreateRecord, repomgr.EvtKindUpdateRecord:
			// read the record bytes from blocks, and verify CID
			rc, recordCBOR, err := rr.GetRecordBytes(ctx, op.Path)
			if err != nil {
				localLogger.Error("reading record from event blocks (CAR)", "err", err)
				continue
			}
			if op.Cid == nil || lexutil.LexLink(rc) != *op.Cid {
				localLogger.Error("mismatch between commit op CID and record block", "recordCID", rc, "opCID", op.Cid)
				continue
			}

			switch collection {
			case "app.bsky.feed.post":
				var post bsky.FeedPost
				if err := post.UnmarshalCBOR(bytes.NewReader(*recordCBOR)); err != nil {
					localLogger.Error("failed to parse app.bsky.feed.post record", "err", err)
					continue
				}
				postCallback(evt.Seq, did, rkey, post)
			}
		}
	}

	return nil
}

func ConsumeAndSaveToMongoDB(
	ctx context.Context,
	postCollection *collections.PostCollection,
	relayHost string,
	cursorOption types.ConsumerCursor,
	oldestPostDuration time.Duration,
	batchFlushTime time.Duration,
) error {
	firehoseDataChan := make(chan CallbackData, 500)
	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var sequenceCursor *int64
	switch cursorOption {
	case types.ConsumerCursorLastConsumed:
		var err error
		sequenceCursor, err = postCollection.GetMaxSequence(ctx)
		if err != nil {
			return err
		}
	case types.ConsumerCursorFirstStream:
		sequenceCursor = utils.ToPtr[int64](0)
	case types.ConsumerCursorCurrentStream:
		sequenceCursor = nil
	}

	consumerLastFlushingTime := time.Now()
	go func() {
		defer cancel()
		for {
			err := RunFirehoseConsumer(
				localCtx,
				relayHost,
				func(sequence int64, did syntax.DID, recordKey syntax.RecordKey, post bsky.FeedPost) {
					firehoseDataChan <- CallbackData{sequence, did, recordKey, post}
				},
				sequenceCursor,
			)

			if err != nil {
				if localCtx.Err() != nil {
					break
				}
				logger.Log.Error(err.Error())
				if !strings.HasPrefix(err.Error(), "repoStream error") {
					break
				}
			}

			sequenceCursor, err = postCollection.GetMaxSequence(ctx)
			if err != nil {
				logger.Log.Error(err.Error())
				break
			}
		}
	}()

	postBatch := []*collections.Post{}
	ticker := time.NewTicker(batchFlushTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("consumer shutting down")
			return nil

		case <-localCtx.Done():
			logger.Log.Error("inactive firehose consumer error")
			return nil

		case data := <-firehoseDataChan:
			facets := &collections.Facets{}
			for _, facet := range data.Post.Facets {
				for _, feature := range facet.Features {
					if feature.RichtextFacet_Mention != nil {
						facets.Mentions = append(facets.Mentions, feature.RichtextFacet_Mention.Did)
					}
					if feature.RichtextFacet_Link != nil {
						facets.Links = append(facets.Links, feature.RichtextFacet_Link.Uri)
					}
					if feature.RichtextFacet_Tag != nil {
						facets.Tags = append(facets.Tags, feature.RichtextFacet_Tag.Tag)
					}
				}
			}

			reply := &collections.Reply{}
			if data.Post.Reply != nil {
				if data.Post.Reply.Root != nil {
					reply.RootURI = data.Post.Reply.Root.Uri
				}
				if data.Post.Reply.Parent != nil {
					reply.ParentURI = data.Post.Reply.Parent.Uri
				}
			}

			createdAt, _ := time.Parse(time.RFC3339, data.Post.CreatedAt)
			if createdAt.After(time.Now().UTC().Add(-oldestPostDuration)) {
				postItem := &collections.Post{
					ID:        fmt.Sprintf("%s/%s", data.DID, data.RecordKey),
					Sequence:  data.Sequence,
					DID:       data.DID.String(),
					RecordKey: data.RecordKey.String(),
					CreatedAt: createdAt,
					Langs:     data.Post.Langs,
					Tags:      data.Post.Tags,
					Text:      data.Post.Text,
					Facets:    facets,
					Reply:     reply,
				}
				postBatch = append(postBatch, postItem)
			}

		case <-ticker.C:
			if len(postBatch) > 0 {
				consumerLastFlushingTime = time.Now()
				// logger.Log.Info("flushing post batch", "count", len(postBatch))
				err := postCollection.Insert(ctx, true, postBatch...)
				if err != nil {
					return fmt.Errorf("mongodb post insert error: %v", err)
				}
				postBatch = []*collections.Post{} // Clear batch after insert
			} else {
				// If we haven't seen any data for 25 seconds, cancel the consumer connection
				if consumerLastFlushingTime.Add(time.Second * 25).Before(time.Now()) {
					cancel()
				}
			}
		}
	}
}
