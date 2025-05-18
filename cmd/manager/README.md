# Feed Manager CLI Tool

## Overview

The Feed Manager is a command-line interface (CLI) tool that allows users to create, update, and delete feed generator records on the Bluesky network.

**Pre-Built Docker Image**: `git.aykhans.me/bsky/feedgen-manager:latest`

## Commands

### Create a Feed Generator

```bash
task run-manager create
# or
make run-manager create
```

This will prompt for:

- Your Bluesky handle
- Your Bluesky password
- Feed generator hostname
- Record short name (for the URL)
- Display name
- Description (optional)
- Avatar image path (optional)

### Update a Feed Generator

```bash
task run-manager update
# or
make run-manager update
```

Allows updating the properties of an existing feed generator record.

### Delete a Feed Generator

```bash
task run-manager delete
# or
make run-manager delete
```

Permenantly removes a feed generator record from the Bluesky network.
