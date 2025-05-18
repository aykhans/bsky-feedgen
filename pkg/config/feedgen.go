package config

import (
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/types"
	"github.com/aykhans/bsky-feedgen/pkg/utils"
)

type FeedGenAzConfig struct {
	CollectionMaxDocument int64
	GeneratorCronDelay    time.Duration
	CutoffCronDelay       time.Duration
}

func NewFeedGenAzConfig() (*FeedGenAzConfig, types.ErrMap) {
	errs := make(types.ErrMap)
	maxDocument, err := utils.GetEnv[int64]("FEED_AZ_COLLECTION_CUTOFF_CRON_MAX_DOCUMENT")
	if err != nil {
		errs["FEED_AZ_COLLECTION_CUTOFF_CRON_MAX_DOCUMENT"] = err
	}
	generatorCronDelay, err := utils.GetEnv[time.Duration]("FEED_AZ_GENERATER_CRON_DELAY")
	if err != nil {
		errs["FEED_AZ_GENERATER_CRON_DELAY"] = err
	}
	cutoffCronDelay, err := utils.GetEnv[time.Duration]("FEED_AZ_COLLECTION_CUTOFF_CRON_DELAY")
	if err != nil {
		errs["FEED_AZ_COLLECTION_CUTOFF_CRON_DELAY"] = err
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return &FeedGenAzConfig{
		CollectionMaxDocument: maxDocument,
		GeneratorCronDelay:    generatorCronDelay,
		CutoffCronDelay:       cutoffCronDelay,
	}, nil
}
