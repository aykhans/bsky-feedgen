package config

import (
	"github.com/aykhans/bsky-feedgen/pkg/types"
	"github.com/aykhans/bsky-feedgen/pkg/utils"
)

const MongoDBBaseDB = "main"

type MongoDBConfig struct {
	Host     string
	Port     uint16
	Username string
	Password string
}

func NewMongoDBConfig() (*MongoDBConfig, types.ErrMap) {
	errs := make(types.ErrMap)
	host, err := utils.GetEnv[string]("MONGODB_HOST")
	if err != nil {
		errs["host"] = err
	}
	port, err := utils.GetEnv[uint16]("MONGODB_PORT")
	if err != nil {
		errs["port"] = err
	}
	username, err := utils.GetEnvOr("MONGODB_USERNAME", "")
	if err != nil {
		errs["username"] = err
	}
	password, err := utils.GetEnvOr("MONGODB_PASSWORD", "")
	if err != nil {
		errs["password"] = err
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return &MongoDBConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}, nil
}
