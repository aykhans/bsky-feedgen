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

	"github.com/aykhans/bsky-feedgen/pkg/generator"
	"github.com/aykhans/bsky-feedgen/pkg/types"

	"github.com/aykhans/bsky-feedgen/pkg/config"
	"github.com/aykhans/bsky-feedgen/pkg/logger"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb/collections"
	_ "go.uber.org/automaxprocs"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go listenForTermination(func() { cancel() })

	flag.Usage = func() {
		fmt.Println(
			`Usage:

feedgen-az [flags]

Flags:
    -h, -help        Display this help message
    -cursor string   Specify the starting point for feed data generation (default: last-generated)
        Options:
       	    last-generated: Resume from the last generated data in storage
       	    first-post: Start from the beginning of the posts`)
	}

	var cursorOption types.GeneratorCursor
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

	feedGenAzConfig, errMap := config.NewFeedGenAzConfig()
	if errMap != nil {
		logger.Log.Error("feedGenAzConfig ENV error", "error", errMap.ToStringMap())
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

	feedAzCollection, err := collections.NewFeedAzCollection(client)
	if err != nil {
		logger.Log.Error(err.Error())
		os.Exit(1)
	}

	feedGeneratorAz := generator.NewFeedGeneratorAz(postCollection, feedAzCollection)

	startCrons(ctx, feedGenAzConfig, feedGeneratorAz, feedAzCollection, cursorOption)
	logger.Log.Info("Cron jobs started")

	<-ctx.Done()
}

func startCrons(
	ctx context.Context,
	feedGenAzConfig *config.FeedGenAzConfig,
	feedGeneratorAz *generator.FeedGeneratorAz,
	feedAzCollection *collections.FeedAzCollection,
	cursorOption types.GeneratorCursor,
) {
	// Feed az generator
	go func() {
		for {
			startTime := time.Now()
			err := feedGeneratorAz.Start(ctx, cursorOption, 1)
			if err != nil {
				logger.Log.Error("Feed az generator cron error", "error", err)
			}
			elapsedTime := time.Since(startTime)
			logger.Log.Info("Feed az generator cron completed", "time", elapsedTime)

			time.Sleep(feedGenAzConfig.GeneratorCronDelay)
		}
	}()

	// feed_az collection cutoff
	go func() {
		for {
			startTime := time.Now()
			deleteCount, err := feedAzCollection.CutoffByCount(ctx, feedGenAzConfig.CollectionMaxDocument)
			if err != nil {
				logger.Log.Error("feed_az collection cutoff cron error", "error", err)
			}
			elapsedTime := time.Since(startTime)
			logger.Log.Info("feed_az collection cutoff cron completed", "count", deleteCount, "time", elapsedTime)

			time.Sleep(feedGenAzConfig.CutoffCronDelay)
		}
	}()
}

func listenForTermination(do func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	do()
}
