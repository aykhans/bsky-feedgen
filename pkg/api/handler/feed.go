package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/api/middleware"
	"github.com/aykhans/bsky-feedgen/pkg/api/response"
	"github.com/aykhans/bsky-feedgen/pkg/feed"
	generatorAz "github.com/aykhans/bsky-feedgen/pkg/generator/az"
	"github.com/aykhans/bsky-feedgen/pkg/types"
	"github.com/aykhans/bsky-feedgen/pkg/utils"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/whyrusleeping/go-did"
)

type FeedHandler struct {
	feedsOutput  []*bsky.FeedDescribeFeedGenerator_Feed
	feedsMap     map[string]feed.Feed
	publisherDID *did.DID
}

func NewFeedHandler(feeds []feed.Feed, publisherDID *did.DID) *FeedHandler {
	ctx := context.Background()

	feedsMap := make(map[string]feed.Feed)
	for _, feed := range feeds {
		feedsMap[feed.GetName(ctx)] = feed
	}

	feedsOutput := make([]*bsky.FeedDescribeFeedGenerator_Feed, len(feeds))
	for i, f := range feeds {
		feedsOutput[i] = utils.ToPtr(f.Describe(ctx))
	}

	return &FeedHandler{
		feedsOutput:  feedsOutput,
		feedsMap:     feedsMap,
		publisherDID: publisherDID,
	}
}

func (handler *FeedHandler) DescribeFeeds(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, 200, bsky.FeedDescribeFeedGenerator_Output{
		Did:   handler.publisherDID.String(),
		Feeds: handler.feedsOutput,
	})
}

func (handler *FeedHandler) GetFeedSkeleton(w http.ResponseWriter, r *http.Request) {
	userDID, _ := middleware.GetValue[string](r, middleware.UserDIDKey)

	feedQuery := r.URL.Query().Get("feed")
	if feedQuery == "" {
		response.JSON(w, 400, response.M{"error": "feed query parameter is required"})
		return
	}

	feedNameStartingIndex := strings.LastIndex(feedQuery, "/")
	if feedNameStartingIndex == -1 {
		response.JSON(w, 400, response.M{"error": "feed query parameter is invalid"})
	}

	feedName := feedQuery[feedNameStartingIndex+1:]
	feed := handler.feedsMap[feedName]
	if feed == nil {
		response.JSON(w, 400, response.M{"error": "feed not found"})
		return
	}

	limitQuery := r.URL.Query().Get("limit")
	var limit int64 = 50
	if limitQuery != "" {
		parsedLimit, err := strconv.ParseInt(limitQuery, 10, 64)
		if err == nil && parsedLimit >= 1 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	cursor := r.URL.Query().Get("cursor")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	feedItems, newCursor, err := feed.GetPage(ctx, userDID, limit, cursor)
	if err != nil {
		if err == types.ErrInternal {
			response.JSON500(w)
			return
		}
		response.JSON(w, 400, response.M{"error": err.Error()})
		return
	}

	response.JSON(w, 200, bsky.FeedGetFeedSkeleton_Output{
		Feed:   feedItems,
		Cursor: newCursor,
	})
}

func (handler *FeedHandler) GetValidUsers(w http.ResponseWriter, r *http.Request) {
	feed := r.PathValue("feed")

	validUsers := make([]string, 0)
	switch feed {
	case "AzPulse":
		validUsers = generatorAz.Users.GetValidUsers()
	}

	response.JSON(w, 200, response.M{
		"feed":  feed,
		"users": validUsers,
	})
}

func (handler *FeedHandler) GetInvalidUsers(w http.ResponseWriter, r *http.Request) {
	feed := r.PathValue("feed")

	invalidUsers := make([]string, 0)
	switch feed {
	case "AzPulse":
		invalidUsers = generatorAz.Users.GetInvalidUsers()
	}

	response.JSON(w, 200, response.M{
		"feed":  feed,
		"users": invalidUsers,
	})
}

func (handler *FeedHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	feed := r.PathValue("feed")

	responseData := response.M{"feed": feed}
	switch feed {
	case "AzPulse":
		responseData["valid_users"] = generatorAz.Users.GetValidUsers()
		responseData["invalid_users"] = generatorAz.Users.GetInvalidUsers()
	}

	response.JSON(w, 200, responseData)
}
