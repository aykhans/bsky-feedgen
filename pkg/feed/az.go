package feed

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aykhans/bsky-feedgen/pkg/logger"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb/collections"
	"github.com/aykhans/bsky-feedgen/pkg/types"
	"github.com/aykhans/bsky-feedgen/pkg/utils"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/whyrusleeping/go-did"
)

type FeedAz struct {
	name             string
	did              *did.DID
	feedAzCollection *collections.FeedAzCollection
}

func NewFeedAz(name string, publisherDID *did.DID, feedAzCollection *collections.FeedAzCollection) *FeedAz {
	return &FeedAz{
		name:             name,
		did:              publisherDID,
		feedAzCollection: feedAzCollection,
	}
}

func (f FeedAz) GetName(_ context.Context) string {
	return f.name
}

func (f *FeedAz) Describe(_ context.Context) bsky.FeedDescribeFeedGenerator_Feed {
	return bsky.FeedDescribeFeedGenerator_Feed{
		Uri: "at://" + f.did.String() + "/app.bsky.feed.generator/" + f.name,
	}
}

func (f *FeedAz) GetPage(
	ctx context.Context,
	_ string, // user did
	limit int64,
	cursor string,
) ([]*bsky.FeedDefs_SkeletonFeedPost, *string, error) {
	var cursorInt int64 = 0
	if cursor != "" {
		var err error
		cursorInt, err = strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("cursor is not an integer: %w", err)
		}
	}

	feedAzItems, err := f.feedAzCollection.GetByCreatedAt(ctx, cursorInt, limit+1)
	if err != nil {
		logger.Log.Error("failed to get feedAzCollection items", "error", err)
		return nil, nil, types.ErrInternal
	}

	var newCursor *string

	if feedAzItemsLen := int64(len(feedAzItems)); limit >= feedAzItemsLen {
		posts := make([]*bsky.FeedDefs_SkeletonFeedPost, feedAzItemsLen)
		for i, feedItem := range feedAzItems {
			posts[i] = &bsky.FeedDefs_SkeletonFeedPost{
				Post: "at://" + feedItem.DID + "/app.bsky.feed.post/" + feedItem.RecordKey,
			}
		}
		return posts, newCursor, nil
	} else {
		posts := make([]*bsky.FeedDefs_SkeletonFeedPost, feedAzItemsLen-1)
		for i, feedItem := range feedAzItems[:feedAzItemsLen-1] {
			posts[i] = &bsky.FeedDefs_SkeletonFeedPost{
				Post: "at://" + feedItem.DID + "/app.bsky.feed.post/" + feedItem.RecordKey,
			}
		}
		return posts, utils.ToPtr(strconv.FormatInt(cursorInt+limit, 10)), nil
	}
}
