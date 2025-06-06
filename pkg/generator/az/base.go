package az

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

type Generator struct {
	postCollection   *collections.PostCollection
	feedAzCollection *collections.FeedAzCollection
	textRegex        *regexp.Regexp
}

func NewGenerator(
	postCollection *collections.PostCollection,
	feedAzCollection *collections.FeedAzCollection,
) *Generator {
	return &Generator{
		postCollection:   postCollection,
		feedAzCollection: feedAzCollection,
		textRegex:        regexp.MustCompile("(?i)(azerbaijan|azərbaycan|aзербайджан|azerbaycan)"),
	}
}

func (generator *Generator) Start(ctx context.Context, cursorOption types.GeneratorCursor, batchSize int) error {
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

func (generator *Generator) IsValid(post *collections.Post) bool {
	// Skip posts that are deep replies (not direct replies to original posts)
	if post.Reply != nil && post.Reply.RootURI != post.Reply.ParentURI {
		return false
	}

	// Check if the user who created this post is in our pre-defined list
	// This allows for explicit inclusion/exclusion of specific users
	if isValidUser := Users.IsValid(post.DID); isValidUser != nil {
		return *isValidUser
	}

	// A post is considered valid if it meets either of the following criteria:
	// 1. It's primarily in Azerbaijani (language code "az") with less than 3 detected languages
	//    (to filter out multi-language spam)
	// 2. It contains Azerbaijan-related keywords in the text AND has at least one valid language
	//    from our approved language list
	if (slices.Contains(post.Langs, "az") && len(post.Langs) < 3) ||
		(generator.textRegex.MatchString(post.Text) && Langs.IsExistsAny(post.Langs)) {
		return true
	}

	return false
}
