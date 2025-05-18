package feed

import (
	"context"

	"github.com/bluesky-social/indigo/api/bsky"
)

type Feed interface {
	GetPage(ctx context.Context, userDID string, limit int64, cursor string) (feedPosts []*bsky.FeedDefs_SkeletonFeedPost, newCursor *string, err error)
	GetName(ctx context.Context) string
	Describe(ctx context.Context) bsky.FeedDescribeFeedGenerator_Feed
}
