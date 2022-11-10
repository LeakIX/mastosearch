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

- `MASTODON_SERVER` To watch a single instance, unset watches all
- `ES_SERVER` Elasticsearch address
- `NSQD_SERVER` NSQD server

## Build & run

````sh
# Minimum max_map_acount for elasticsearch :
sudo sysctl vm.max_map_count=262144
# Build and run from docker-compose.yaml
docker-compose up --build
```
