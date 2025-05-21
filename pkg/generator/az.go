package generator

import (
	"context"
	"fmt"
	"regexp"
	"slices"

	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb/collections"
	"github.com/aykhans/bsky-feedgen/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var azInvalidUsers []string = []string{
	"did:plc:5zww7zorx2ajw7hqrhuix3ba",
	"did:plc:c4vhz47h566t2ntgd7gtawen",
	"did:plc:lc7j7xdq67gn7vc6vzmydfqk",
}

var azValidUsers []string = []string{
	"did:plc:jbt4qi6psd7rutwzedtecsq7",
	"did:plc:yzgdpxsklrmfgqmjghdvw3ti",
	"did:plc:g7ebgiai577ln3avsi2pt3sn",
	"did:plc:phtq2rhgbwipyx5ie3apw44j",
	"did:plc:jfdvklrs5n5qv7f25v6swc5h",
	"did:plc:u5ez5w6qslh6advti4wyddba",
	"did:plc:cs2cbzojm6hmx5lfxiuft3mq",
}

type FeedGeneratorAz struct {
	postCollection   *collections.PostCollection
	feedAzCollection *collections.FeedAzCollection
	textRegex        *regexp.Regexp
}

func NewFeedGeneratorAz(
	postCollection *collections.PostCollection,
	feedAzCollection *collections.FeedAzCollection,
) *FeedGeneratorAz {
	return &FeedGeneratorAz{
		postCollection:   postCollection,
		feedAzCollection: feedAzCollection,
		textRegex:        regexp.MustCompile("(?i)(azerbaijan|azərbaycan|aзербайджан|azerbaycan)"),
	}
}

func (generator *FeedGeneratorAz) Start(ctx context.Context, cursorOption types.GeneratorCursor, batchSize int) error {
	var mongoCursor *mongo.Cursor
	switch cursorOption {
	case types.GeneratorCursorLastGenerated:
		sequenceCursor, err := generator.feedAzCollection.GetMaxSequence(ctx)
		if err != nil {
			return err
		}

		if sequenceCursor == nil {
			mongoCursor, err = generator.postCollection.Collection.Find(
				ctx,
				bson.D{},
				options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}),
			)
		} else {
			mongoCursor, err = generator.postCollection.Collection.Find(
				ctx,
				bson.M{"sequence": bson.M{"$gt": *sequenceCursor}},
			)
		}
		if err != nil {
			return err
		}
	case types.GeneratorCursorFirstPost:
		var err error
		mongoCursor, err = generator.postCollection.Collection.Find(
			ctx,
			bson.D{},
			options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}),
		)
		if err != nil {
			return err
		}
	}

	defer func() { _ = mongoCursor.Close(ctx) }()

	feedAzBatch := []*collections.FeedAz{}
	for mongoCursor.Next(ctx) {
		var doc *collections.Post
		if err := mongoCursor.Decode(&doc); err != nil {
			return fmt.Errorf("mongodb cursor decode error: %v", err)
		}

		if generator.IsValid(doc) == false {
			continue
		}

		feedAzBatch = append(
			feedAzBatch,
			&collections.FeedAz{
				ID:        doc.ID,
				Sequence:  doc.Sequence,
				DID:       doc.DID,
				RecordKey: doc.RecordKey,
				CreatedAt: doc.CreatedAt,
			},
		)

		if len(feedAzBatch)%batchSize == 0 {
			err := generator.feedAzCollection.Insert(ctx, true, feedAzBatch...)
			if err != nil {
				return fmt.Errorf("insert FeedAz error: %v", err)
			}
			feedAzBatch = []*collections.FeedAz{}
		}
	}

	if len(feedAzBatch) > 0 {
		err := generator.feedAzCollection.Insert(ctx, true, feedAzBatch...)
		if err != nil {
			return fmt.Errorf("insert FeedAz error: %v", err)
		}
	}

	return nil
}

func (generator *FeedGeneratorAz) IsValid(post *collections.Post) bool {
	if post.Reply != nil && post.Reply.RootURI != post.Reply.ParentURI {
		return false
	}

	if slices.Contains(azInvalidUsers, post.DID) {
		return false
	}

	if slices.Contains(azValidUsers, post.DID) || // Posts from always-valid users
		(slices.Contains(post.Langs, "az") && len(post.Langs) < 3) || // Posts in Azerbaijani language with fewer than 3 languages
		generator.textRegex.MatchString(post.Text) { // Posts containing Azerbaijan-related keywords
		return true
	}

	return false
}
