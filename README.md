# Mastosearch

This is a 2 part toolset to index the Mastodon federated network.

## cmd/watcher

- Downloads a list of Mastodon server and connects to their public stream to watch for local status.
- Send status update and delete requests to a NSQ server
- Keeps track of timeline download process in a local sqlite datatbase

## cmd/indexer

- Connects to NSQ
- Indexes and deletes status received from `cmd/watcher`

## Environment variables

- `ES_SERVER` Elasticsearch address
- `NSQD_SERVER` NSQD server

## Requirement
- NSQ server
- Elasticsearch server
