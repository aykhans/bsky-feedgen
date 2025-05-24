package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/api/handler"
	"github.com/aykhans/bsky-feedgen/pkg/api/middleware"
	"github.com/aykhans/bsky-feedgen/pkg/config"
	"github.com/aykhans/bsky-feedgen/pkg/feed"
	"github.com/aykhans/bsky-feedgen/pkg/logger"
)

func Run(
	ctx context.Context,
	apiConfig *config.APIConfig,
	feeds []feed.Feed,
) error {
	baseHandler, err := handler.NewBaseHandler(apiConfig.FeedgenHostname, apiConfig.ServiceDID)
	if err != nil {
		return err
	}
	feedHandler := handler.NewFeedHandler(feeds, apiConfig.FeedgenPublisherDID)

	authMiddleware := middleware.NewAuth(apiConfig.ServiceDID)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /.well-known/did.json", baseHandler.GetWellKnownDIDDoc)
	mux.HandleFunc("GET /xrpc/app.bsky.feed.describeFeedGenerator", feedHandler.DescribeFeeds)
	mux.Handle(
		"GET /xrpc/app.bsky.feed.getFeedSkeleton",
		authMiddleware.JWTAuthMiddleware(http.HandlerFunc(feedHandler.GetFeedSkeleton)),
	)
	mux.HandleFunc("GET /{feed}/users", feedHandler.GetAllUsers)
	mux.HandleFunc("GET /{feed}/users/valid/", feedHandler.GetValidUsers)
	mux.HandleFunc("GET /{feed}/users/invalid/", feedHandler.GetInvalidUsers)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", apiConfig.APIPort),
		Handler: mux,
	}

	listenerErrChan := make(chan error)

	logger.Log.Info(fmt.Sprintf("Starting server on port %d", apiConfig.APIPort))
	go func() {
		listenerErrChan <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-listenerErrChan:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("error while serving http: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("error while shutting down http server: %v", err)
		}
	}
	logger.Log.Info(fmt.Sprintf("Server on port %d stopped", apiConfig.APIPort))

	return nil
}
