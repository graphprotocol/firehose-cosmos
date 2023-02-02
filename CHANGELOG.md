# Changelog

## UNRELEASED

### Added

* Added support for "requester pays" buckets on Google Storage in url, ex: `gs://my-bucket/path?project=my-project-id`

## v0.6.0

### Added

* Exposing new GRPC endpoints for firehose: `sf.firehose.v2.Stream/Blocks` and `sf.firehose.v2.Fetch/Block`

### Breaking changes
* One-block-file naming is different: you need to remove existing one-block-files when transitioning

### Changed naming and flags
* Changed binary name: `firehose-cosmos` becomes `firecosmos`
* `firehose-block-index-sizes` becomes `common-block-index-sizes`
* `firehose-block-index-url` becomes `common-index-store-url`
* `common-blocks-store-url` becomes `common-merged-blocks-store-url`
* `common-blockstream-addr` becomes `common-live-blocks-addr`
* `common-oneblock-store-url` becomes `common-one-block-store-url`

### Removed flags
* `firehose-real-time-tolerance`, `firehose-rpc-head-tracker-url`, `firehose-static-head-tracker`, `firehose-tracker-offset` are removed
* `merger-state-file`, `merger-max-one-block-operations-batch-size`, `merger-next-exclusive-highest-block-limit`, `merger-one-block-deletion-threads`, `merger-writers-leeway` are removed
* `ingestor-merge-threshold-block-age`, `reader-wait-upload-complete-on-shutdown`
* `relayer-buffer-size`, `relayer-merger-addr`, `relayer-min-start-offset`, `relayer-source-request-burst`

### Added flags
* `merger-stop-block`, `merger-time-between-store-pruning`
* `reader-oneblock-suffix`, `reader-readiness-max-latency`
