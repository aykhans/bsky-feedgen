package config

import (
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/types"
	"github.com/aykhans/bsky-feedgen/pkg/utils"
)

type ConsumerConfig struct {
	PostMaxDate                         time.Duration
	PostCollectionCutoffCronDelay       time.Duration
	PostCollectionCutoffCronMaxDocument int64
}

func NewConsumerConfig() (*ConsumerConfig, types.ErrMap) {
	errs := make(types.ErrMap)
	maxDate, err := utils.GetEnv[time.Duration]("POST_MAX_DATE")
	if err != nil {
		errs["POST_MAX_DATE"] = err
	}
	cronDelay, err := utils.GetEnv[time.Duration]("POST_COLLECTION_CUTOFF_CRON_DELAY")
	if err != nil {
		errs["POST_COLLECTION_CUTOFF_CRON_DELAY"] = err
	}
	cronMaxDocument, err := utils.GetEnv[int64]("POST_COLLECTION_CUTOFF_CRON_MAX_DOCUMENT")
	if err != nil {
		errs["POST_COLLECTION_CUTOFF_CRON_MAX_DOCUMENT"] = err
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return &ConsumerConfig{
		PostMaxDate:                         maxDate,
		PostCollectionCutoffCronDelay:       cronDelay,
		PostCollectionCutoffCronMaxDocument: cronMaxDocument,
	}, nil
}
