# Example Production Deployment

This is an example of a production deployment for the Feed Generator.

## Architecture

The production setup includes the following services:

- **MongoDB**: Database for storing posts and feed data
- **Consumer**: Service that consumes AT Protocol firehose data
- **Feed Generator (AZ)**: Generates feeds for Azerbaijan-related content
- **API**: REST API service for serving feeds
- **Caddy**: Reverse proxy

## Quick Start

1. **Configure the environment**:
   ```bash
   make configure
   ```
   This will:
   - Copy all example configuration files
   - Prompt for MongoDB credentials
   - Prompt for domain name and AT Protocol DID
   - Update configuration files with your values

2. **Start the services**:
   ```bash
   docker compose up -d
   ```

3. **Check service status**:
   ```bash
   docker compose ps
   docker compose logs
   ```

## Configuration Files

### Application Configuration
- `config/app/.api.env` - API service configuration
- `config/app/.consumer.env` - Consumer service configuration  
- `config/app/.mongodb.env` - MongoDB connection settings
- `config/app/feedgen/.az.env` - Azerbaijan feed generator settings

### Infrastructure Configuration
- `config/caddy/.env` - Caddy reverse proxy settings
- `config/caddy/Caddyfile` - Caddy server configuration
- `config/mongodb/.env` - MongoDB initialization settings

## Environment Variables

### API Service
- `FEEDGEN_HOSTNAME` - Public hostname for the feed generator
- `FEEDGEN_PUBLISHER_DID` - Your AT Protocol DID
- `API_PORT` - Port for the API service (default: 8421)

### Consumer Service
- `POST_MAX_DATE` - Maximum age of posts to store (default: 720h/30 days)
- `POST_COLLECTION_CUTOFF_CRON_DELAY` - Cleanup interval (default: 30m)
- `POST_COLLECTION_CUTOFF_CRON_MAX_DOCUMENT` - Max documents before cleanup (default: 1M)

### AZ Feed Generator
- `FEED_AZ_GENERATER_CRON_DELAY` - Feed generation interval (default: 1m)
- `FEED_AZ_COLLECTION_CUTOFF_CRON_DELAY` - Cleanup interval (default: 30m)
- `FEED_AZ_COLLECTION_CUTOFF_CRON_MAX_DOCUMENT` - Max documents before cleanup (default: 500K)

### MongoDB
- `MONGODB_HOST` - MongoDB hostname (default: mongodb)
- `MONGODB_PORT` - MongoDB port (default: 27017)
- `MONGODB_USERNAME` - Database username
- `MONGODB_PASSWORD` - Database password

### Caddy
- `DOMAIN` - Your domain name
- `API_HOST` - Internal API service URL (default: http://api:8421)
