# AzPulse Feed Generator Service

## Overview

The AzPulse Feed Generator service processes posts stored by the Consumer service and generates feed content that will be served by the API service. It implements the logic for the "AzPulse" feed, which showcases selected content from the Bluesky network.

**Pre-Built Docker Image**: `git.aykhans.me/bsky/feedgen-generator-az:latest`

## Features

- Processes posts from MongoDB
- Applies custom feed generation logic for the AzPulse feed
- Stores feed results in MongoDB for API service to access
- Manages feed data lifecycle with automatic pruning
- Runs as a background service with cron jobs

## Command Line Options

- `-cursor`: Specify the starting point for feed data generation
    - `last-generated`: Resume from the last generated data (default)
    - `first-post`: Start from the beginning of the posts collection

## Running the Service

### Docker

```bash
docker build -f cmd/feedgen/az/Dockerfile -t bsky-feedgen-az .
docker --env-file config/app/feedgen/.az.env --enf-file config/app/.mongodb.env run bsky-feedgen-az
```

### Local Development

```bash
task run-feedgen-az
# or
make run-feedgen-az
```
