package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/consumer"
	"github.com/aykhans/bsky-feedgen/pkg/types"

	"github.com/aykhans/bsky-feedgen/pkg/config"
	"github.com/aykhans/bsky-feedgen/pkg/logger"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb/collections"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go listenForTermination(func() { cancel() })

	flag.Usage = func() {
		fmt.Println(
			`Usage:

consumer [flags]

Flags:
    -h, -help        Display this help message
    -cursor string   Specify the starting point for data consumption (default: last-consumed)
        Options:
       	    last-consumed: Resume from the last processed data in storage
       	    first-stream: Start from the beginning of the firehose
       	    current-stream: Start from the current position in the firehose stream`)
	}

	var cursorOption types.ConsumerCursor
	flag.Var(&cursorOption, "cursor", "")
	flag.Parse()

	if args := flag.Args(); len(args) > 0 {
		if len(args) == 1 {
			fmt.Printf("unexpected argument: %s\n\n", args[0])
		} else {
			fmt.Printf("unexpected arguments: %v\n\n", strings.Join(args, ", "))
		}
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	if cursorOption == "" {
		_ = cursorOption.Set("")
	}

	consumerConfig, errMap := config.NewConsumerConfig()
	if errMap != nil {
		logger.Log.Error("consumer ENV error", "error", errMap.ToStringMap())
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

	postCollection, err := collections.NewPostCollection(client)
	if err != nil {
		logger.Log.Error(err.Error())
		os.Exit(1)
	}

	startCrons(ctx, consumerConfig, postCollection)
	logger.Log.Info("Cron jobs started")

	err = consumer.ConsumeAndSaveToMongoDB(
		ctx,
		postCollection,
		"wss://bsky.network",
		cursorOption,
		consumerConfig.PostMaxDate, // Save only posts created before PostMaxDate
		10*time.Second,             // Save consumed data to MongoDB every 10 seconds
	)
	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func startCrons(ctx context.Context, consumerConfig *config.ConsumerConfig, postCollection *collections.PostCollection) {
	// Post collection cutoff
	go func() {
		for {
			startTime := time.Now()
			deleteCount, err := postCollection.CutoffByCount(ctx, consumerConfig.PostCollectionCutoffCronMaxDocument)
			if err != nil {
				logger.Log.Error("Post collection cutoff cron error", "error", err)
			}
			elapsedTime := time.Since(startTime)
			logger.Log.Info("Post collection cutoff cron completed", "count", deleteCount, "time", elapsedTime)

			time.Sleep(consumerConfig.PostCollectionCutoffCronDelay)
		}
	}()
}

func listenForTermination(do func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	do()
}
