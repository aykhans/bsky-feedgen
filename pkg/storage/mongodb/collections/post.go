package collections

import (
	"context"
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostCollection struct {
	Collection *mongo.Collection
}

func NewPostCollection(client *mongo.Client) (*PostCollection, error) {
	client.Database(config.MongoDBBaseDB).Collection("")
	coll := client.Database(config.MongoDBBaseDB).Collection("post")
	_, err := coll.Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{
				Keys: bson.D{{Key: "sequence", Value: -1}},
			},
			{
				Keys: bson.D{{Key: "created_at", Value: 1}},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return &PostCollection{Collection: coll}, nil
}

type Post struct {
	ID        string    `bson:"_id"`
	Sequence  int64     `bson:"sequence"`
	DID       string    `bson:"did"`
	RecordKey string    `bson:"record_key"`
	CreatedAt time.Time `bson:"created_at"`
	Langs     []string  `bson:"langs"`
	Tags      []string  `bson:"tags"`
	Text      string    `bson:"text"`
	Facets    *Facets   `bson:"facets"`
	Reply     *Reply    `bson:"reply"`
}

type Facets struct {
	Tags     []string `bson:"tags"`
	Links    []string `bson:"links"`
	Mentions []string `bson:"mentions"`
}

type Reply struct {
	RootURI   string `bson:"root_uri"`
	ParentURI string `bson:"parent_uri"`
}

func (p PostCollection) CutoffByCount(
	ctx context.Context,
	maxDocumentCount int64,
) (int64, error) {
	count, err := p.Collection.CountDocuments(ctx, bson.M{})
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

	cursor, err := p.Collection.Find(ctx, bson.M{}, findOpts)
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
		result, err := p.Collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": batch}})
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

func (p PostCollection) GetMaxSequence(ctx context.Context) (*int64, error) {
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

	cursor, err := p.Collection.Aggregate(ctx, pipeline)
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

func (p PostCollection) Insert(ctx context.Context, overwrite bool, posts ...*Post) error {
	switch len(posts) {
	case 0:
		return nil
	case 1:
		if overwrite == false {
			_, err := p.Collection.InsertOne(ctx, posts[0])
			return err
		}
		_, err := p.Collection.ReplaceOne(
			ctx,
			bson.M{"_id": posts[0].ID},
			posts[0],
			options.Replace().SetUpsert(true),
		)
		return err
	default:
		if overwrite == false {
			documents := make([]any, len(posts))
			for i, post := range posts {
				documents[i] = post
			}

			_, err := p.Collection.InsertMany(ctx, documents)
			return err
		}
		var models []mongo.WriteModel

		for _, post := range posts {
			filter := bson.M{"_id": post.ID}
			model := mongo.NewReplaceOneModel().
				SetFilter(filter).
				SetReplacement(post).
				SetUpsert(true)
			models = append(models, model)
		}

		opts := options.BulkWrite().SetOrdered(false)
		_, err := p.Collection.BulkWrite(ctx, models, opts)
		if err != nil {
			return err
		}

		return nil
	}
}
