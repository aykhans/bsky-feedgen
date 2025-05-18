package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewDB(ctx context.Context, dbConfig *config.MongoDBConfig) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(
		fmt.Sprintf("mongodb://%s:%v/", dbConfig.Host, dbConfig.Port),
	)

	if dbConfig.Username != "" {
		clientOptions.SetAuth(options.Credential{
			Username: dbConfig.Username,
			Password: dbConfig.Password,
		})
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
