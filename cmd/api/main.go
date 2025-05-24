package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aykhans/bsky-feedgen/pkg/api"
	"github.com/aykhans/bsky-feedgen/pkg/config"
	"github.com/aykhans/bsky-feedgen/pkg/feed"
	"github.com/aykhans/bsky-feedgen/pkg/logger"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb"
	"github.com/aykhans/bsky-feedgen/pkg/storage/mongodb/collections"
	_ "go.uber.org/automaxprocs"
)

type flags struct {
	version bool
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go listenForTermination(func() { cancel() })

	flags := getFlags()
	if flags.version == true {
		fmt.Printf("API version: %v\n", version)
		os.Exit(0)
	}

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

func getFlags() *flags {
	flags := &flags{}

	flag.Usage = func() {
		fmt.Println(
			`Usage:

consumer [flags]

Flags:
    -version         version information
    -h, -help        Display this help message`)
	}

	flag.BoolVar(&flags.version, "version", false, "print version information")
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

	return flags
}
