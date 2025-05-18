package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/aykhans/bsky-feedgen/pkg/api/response"
	"github.com/whyrusleeping/go-did"
)

type BaseHandler struct {
	WellKnownDIDDoc did.Document
}

func NewBaseHandler(serviceEndpoint *url.URL, serviceDID *did.DID) (*BaseHandler, error) {
	serviceID, err := did.ParseDID("#bsky_fg")
	if err != nil {
		return nil, fmt.Errorf("service ID parse error: %v", err)
	}

	return &BaseHandler{
		WellKnownDIDDoc: did.Document{
			Context: []string{did.CtxDIDv1},
			ID:      *serviceDID,
			Service: []did.Service{
				{
					ID:              serviceID,
					Type:            "BskyFeedGenerator",
					ServiceEndpoint: serviceEndpoint.String(),
				},
			},
		},
	}, nil
}

type WellKnownDidResponse struct {
	Context []string      `json:"@context"`
	ID      string        `json:"id"`
	Service []did.Service `json:"service"`
}

func (handler *BaseHandler) GetWellKnownDIDDoc(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, 200, WellKnownDidResponse{
		Context: handler.WellKnownDIDDoc.Context,
		ID:      handler.WellKnownDIDDoc.ID.String(),
		Service: handler.WellKnownDIDDoc.Service,
	})
}
