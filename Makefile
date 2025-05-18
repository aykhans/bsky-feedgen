# Equivalent Makefile for Taskfile.yaml

.PHONY: ftl fmt tidy lint run-consumer run-feedgen-az run-api run-manager generate-env

# Default value for ARGS if not provided on the command line
ARGS ?=

# Runs fmt, tidy, and lint sequentially
ftl:
	$(MAKE) fmt
	$(MAKE) tidy
	$(MAKE) lint

# Format Go code
fmt:
	gofmt -w -d .

# Tidy Go modules
tidy:
	go mod tidy

# Run golangci-lint
lint:
	golangci-lint run

# Run the consumer application, loading environment from dotenv files
run-consumer:
	set -a; \
	. config/app/.consumer.env; \
	. config/app/.mongodb.env; \
	set +a; \
	go run cmd/consumer/main.go $(ARGS)

# Run the feedgen-az application, loading environment from dotenv files
run-feedgen-az:
	set -a; \
	. config/app/feedgen/.az.env; \
	. config/app/.mongodb.env; \
	set +a; \
	go run cmd/feedgen/az/main.go $(ARGS)

# Run the api application, loading environment from dotenv files
run-api:
	set -a; \
	. config/app/.api.env; \
	. config/app/.mongodb.env; \
	set +a; \
	go run cmd/api/main.go

# Run the manager application with arguments (no dotenv)
run-manager:
	go run cmd/manager/main.go $(ARGS)

# Generate env files from templates
generate-env:
	cp config/app/consumer.env.example config/app/.consumer.env
	cp config/app/api.env.example config/app/.api.env
	cp config/app/mongodb.env.example config/app/.mongodb.env
	cp config/app/feedgen/az.env.example config/app/feedgen/.az.env
	cp config/mongodb/env.example config/mongodb/.env
