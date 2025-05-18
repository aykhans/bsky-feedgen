# Consumer Service

## Overview

The Consumer service is responsible for connecting to the Bluesky firehose, processing incoming posts, and storing them in MongoDB for later use by the feed generator.

**Pre-Built Docker Image**: `git.aykhans.me/bsky/feedgen-consumer:latest`

## Features

- Connects to the Bluesky firehose websocket
- Processes and filters incoming posts
- Stores relevant post data in MongoDB
- Includes data management via cron jobs
    - Implements collection size limits
    - Prunes older data to prevent storage issues

## Command Line Options

- `-cursor`: Specify the starting point for data consumption
    - `last-consumed`: Resume from the last processed data (default)
    - `first-stream`: Start from the beginning of the firehose
    - `current-stream`: Start from the current position in the firehose

## Running the Service

### Docker

```bash
docker build -f cmd/consumer/Dockerfile -t bsky-feedgen-consumer .
docker --env-file config/app/.consumer.env --env-file config/app/.mongodb.env run bsky-feedgen-consumer
```

### Local Development

```bash
task run-consumer
# or
make run-consumer
```
