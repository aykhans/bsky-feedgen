package config

import (
	"fmt"
	"net/url"
	"slices"

	"github.com/aykhans/bsky-feedgen/pkg/types"
	"github.com/aykhans/bsky-feedgen/pkg/utils"
	"github.com/whyrusleeping/go-did"
)

type APIConfig struct {
	FeedgenHostname     *url.URL
	ServiceDID          *did.DID
	FeedgenPublisherDID *did.DID
	APIPort             uint16
}

func NewAPIConfig() (*APIConfig, types.ErrMap) {
	errs := make(types.ErrMap)

	defaultHostname, _ := url.Parse("http://localhost")
	feedgenHostname, err := utils.GetEnvOr("FEEDGEN_HOSTNAME", defaultHostname)
	if err != nil {
		errs["FEEDGEN_HOSTNAME"] = err
	} else {
		if !slices.Contains([]string{"", "http", "https"}, feedgenHostname.Scheme) {
			errs["FEEDGEN_HOSTNAME"] = fmt.Errorf(
				"invalid schema '%s' for FEEDGEN_HOSTNAME. Accepted schemas are: '', 'http', 'https'",
				feedgenHostname.Scheme,
			)
		}
	}

	serviceDID, err := did.ParseDID("did:web:" + feedgenHostname.Hostname())
	if err != nil {
		errs["SERVICE_DID"] = fmt.Errorf("failed to parse service DID: %w", err)
	}

	defaultDID, _ := did.ParseDID("did:plc:development")
	feedgenPublisherDID, err := utils.GetEnvOr("FEEDGEN_PUBLISHER_DID", &defaultDID)
	if err != nil {
		errs["FEEDGEN_PUBLISHER_DID"] = err
	}

	apiPort, err := utils.GetEnv[uint16]("API_PORT")
	if err != nil {
		errs["API_PORT"] = err
	}

	if len(errs) > 0 {
		return nil, errs
	}

	if feedgenHostname.Scheme == "" {
		if feedgenHostname.Host == "" {
			feedgenHostname, _ = url.Parse("https://" + feedgenHostname.String())
		} else {
			feedgenHostname.Scheme = "https://"
		}
	}

	return &APIConfig{
		FeedgenHostname:     feedgenHostname,
		ServiceDID:          &serviceDID,
		FeedgenPublisherDID: feedgenPublisherDID,
		APIPort:             apiPort,
	}, nil
}
