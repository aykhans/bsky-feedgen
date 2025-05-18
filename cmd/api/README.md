# API Service

## Overview

The API service is responsible for serving custom Bluesky feeds to clients. It implements the necessary endpoints required by the Bluesky protocol to deliver feed content.

**Pre-Built Docker Image**: `git.aykhans.me/bsky/feedgen-api:latest`

## API Endpoints

- `GET /.well-known/did.json`: DID configuration
- `GET /xrpc/app.bsky.feed.describeFeedGenerator`: Describe the feed generator
- `GET /xrpc/app.bsky.feed.getFeedSkeleton`: Main feed endpoint

## Running the Service

### Docker

```bash
docker build -f cmd/api/Dockerfile -t bsky-feedgen-api .
docker run --env-file config/app/.api.env --env-file config/app/.mongodb.env -p 8421:8421 bsky-feedgen-api
```

### Local Development

```bash
task run-api
# or
make run-api
```
