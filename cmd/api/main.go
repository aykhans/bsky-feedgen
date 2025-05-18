package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aykhans/bsky-feedgen/pkg/api"
	"github.com/aykhans/bsky-feedgen/pkg/config"
	"github.com/aykhans/bsky-feedgen/pkg/feed"
	"github.com/aykhans/bsky-feedgen/pkg/logger"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb/collections"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go listenForTermination(func() { cancel() })

	apiConfig, errMap := config.NewAPIConfig()
	if errMap != nil {
		logger.Log.Error("API ENV error", "error", errMap.ToStringMap())
		os.Exit(1)
	}

	mongoDBConfig, errMap := config.NewMongoDBConfig()
	if errMap != nil {
		logger.Log.Error("mongodb ENV error", "error", errMap.ToStringMap())
		os.Exit(1)
	}

	client, err := mongodb.NewDB(ctx, mongoDBConfig)
	if err != nil {
		logger.Log.Error("mongodb connection error", "error", err)
		os.Exit(1)
	}

	feedAzCollection, err := collections.NewFeedAzCollection(client)
	if err != nil {
		logger.Log.Error(err.Error())
		os.Exit(1)
	}

	feeds := []feed.Feed{
		feed.NewFeedAz("AzPulse", apiConfig.FeedgenPublisherDID, feedAzCollection),
	}

	if err := api.Run(ctx, apiConfig, feeds); err != nil {
		logger.Log.Error("API error", "error", err)
	}
}

func listenForTermination(do func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	do()
}
