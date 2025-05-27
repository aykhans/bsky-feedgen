package handler

import (
	"net/http"

	"github.com/aykhans/bsky-feedgen/pkg/api/response"
	generatorAz "github.com/aykhans/bsky-feedgen/pkg/generator/az"
)

type GeneratorHandler struct{}

func NewGeneratorHandler() *GeneratorHandler {
	return &GeneratorHandler{}
}

func (handler *GeneratorHandler) GetValidUsers(w http.ResponseWriter, r *http.Request) {
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

func (handler *GeneratorHandler) GetInvalidUsers(w http.ResponseWriter, r *http.Request) {
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

func (handler *GeneratorHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	feed := r.PathValue("feed")

	responseData := response.M{"feed": feed}
	switch feed {
	case "AzPulse":
		responseData["valid_users"] = generatorAz.Users.GetValidUsers()
		responseData["invalid_users"] = generatorAz.Users.GetInvalidUsers()
	}

	response.JSON(w, 200, responseData)
}
