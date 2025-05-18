# Bluesky Feed Generator

A Go-based feed generator service for Bluesky social platform that includes components for streaming and processing posts, generating feeds, and exposing them via an API.

## Overview

This project is a custom feed generator for Bluesky (ATProto), implementing the "AzPulse" feed. The application consists of several components working together:

- [**API**](./cmd/api): Serves feed data to Bluesky clients
- [**Consumer**](./cmd/consumer): Streams and processes posts from the Bluesky firehose
- [**FeedGen**](./cmd/feedgen): Processes posts and generates the feed content
- [**Manager**](./cmd/manager): CLI tool for managing feed generator records on Bluesky

## Architecture

The system follows a microservices architecture with the following components:

- MongoDB for data storage
- Multiple services communicating through the database
- Each service deployed as a separate container

## Pre-Built docker images

- **API**: `git.aykhans.me/bsky/feedgen-api:latest`
- **Consumer**: `git.aykhans.me/bsky/feedgen-consumer:latest`
- **FeedGen Az**: `git.aykhans.me/bsky/feedgen-generator-az:latest`
- **Manager**: `git.aykhans.me/bsky/feedgen-manager:latest`

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.24+

### Running with Docker Compose

```bash
docker compose up
```

This will start all necessary services:

- MongoDB database
- Consumer service (streams posts from Bluesky)
- FeedGen service (generates the AZ feed)
- API service (serves feed data to clients)

## Development

For local development without Docker you can checkout the Taskfile and Makefile for development tasks.

## Configuration

All services are configured via environment variables. See the docker-compose.yml file for examples and default values.

## License

This project is licensed under the AGPL-3.0 License, please see the [LICENSE](./LICENSE) file for details.
