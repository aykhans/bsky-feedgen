package collections

import (
	"context"
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FeedAzCollection struct {
	Collection *mongo.Collection
}

func NewFeedAzCollection(client *mongo.Client) (*FeedAzCollection, error) {
	client.Database(config.MongoDBBaseDB).Collection("")
	coll := client.Database(config.MongoDBBaseDB).Collection("feed_az")

	_, err := coll.Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{
				Keys: bson.D{{Key: "sequence", Value: -1}},
			},
			{
				Keys: bson.D{{Key: "created_at", Value: -1}},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return &FeedAzCollection{Collection: coll}, nil
}

type FeedAz struct {
	ID        string    `bson:"_id"`
	Sequence  int64     `bson:"sequence"`
	DID       string    `bson:"did"`
	RecordKey string    `bson:"record_key"`
	CreatedAt time.Time `bson:"created_at"`
}

func (f FeedAzCollection) GetByCreatedAt(ctx context.Context, skip int64, limit int64) ([]*FeedAz, error) {
	cursor, err := f.Collection.Find(
		ctx, bson.D{},
		options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(skip).
			SetLimit(limit),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var feedAzItems []*FeedAz
	if err = cursor.All(ctx, &feedAzItems); err != nil {
		return nil, err
	}

	return feedAzItems, nil
}

func (f FeedAzCollection) GetMaxSequence(ctx context.Context) (*int64, error) {
	pipeline := mongo.Pipeline{
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "maxSequence", Value: bson.D{
					{Key: "$max", Value: "$sequence"},
				},
				},
			},
			},
		},
	}

	cursor, err := f.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var result struct {
		MaxSequence int64 `bson:"maxSequence"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		return &result.MaxSequence, err
	}

	return nil, nil
}

func (f FeedAzCollection) Insert(ctx context.Context, overwrite bool, feedAz ...*FeedAz) error {
	switch len(feedAz) {
	case 0:
		return nil
	case 1:
		if overwrite == false {
			_, err := f.Collection.InsertOne(ctx, feedAz[0])
			return err
		}
		_, err := f.Collection.ReplaceOne(
			ctx,
			bson.M{"_id": feedAz[0].ID},
			feedAz[0],
			options.Replace().SetUpsert(true),
		)
		return err
	default:
		if overwrite == false {
			documents := make([]any, len(feedAz))
			for i, feed := range feedAz {
				documents[i] = feed
			}

			_, err := f.Collection.InsertMany(ctx, documents)
			return err
		}
		var models []mongo.WriteModel

		for _, feed := range feedAz {
			filter := bson.M{"_id": feed.ID}
			model := mongo.NewReplaceOneModel().
				SetFilter(filter).
				SetReplacement(feed).
				SetUpsert(true)
			models = append(models, model)
		}

		opts := options.BulkWrite().SetOrdered(false)
		_, err := f.Collection.BulkWrite(ctx, models, opts)
		if err != nil {
			return err
		}

		return nil
	}
}

func (f FeedAzCollection) CutoffByCount(
	ctx context.Context,
	maxDocumentCount int64,
) (int64, error) {
	count, err := f.Collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	if count <= maxDocumentCount {
		return 0, nil
	}

	deleteCount := count - maxDocumentCount

	findOpts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(deleteCount).
		SetProjection(bson.M{"_id": 1})

	cursor, err := f.Collection.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		return 0, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	// Process documents in batches to avoid potential memory issues
	const batchSize = 10000
	var totalDeleted int64 = 0

	for {
		batch := make([]string, 0, batchSize)
		batchCount := 0

		for cursor.Next(ctx) && batchCount < batchSize {
			var doc struct {
				ID string `bson:"_id"`
			}
			if err = cursor.Decode(&doc); err != nil {
				return totalDeleted, err
			}
			batch = append(batch, doc.ID)
			batchCount++
		}

		if len(batch) == 0 {
			break
		}

		// Delete the batch
		result, err := f.Collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": batch}})
		if err != nil {
			return totalDeleted, err
		}

		totalDeleted += result.DeletedCount

		if cursor.Err() != nil {
			return totalDeleted, cursor.Err()
		}

		// If we didn't fill the batch, we're done
		if batchCount < batchSize {
			break
		}
	}

	return totalDeleted, nil
}
