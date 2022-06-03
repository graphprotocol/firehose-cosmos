# firehose-cosmos

Firehose integration for Cosmos chains

## Getting Started

To get started, first clone the repository and install all dependencies:

```bash
git clone https://github.com/figment-networks/firehose-cosmos.git
go mod download
```

Once done, let's build the development binary:

```bash
make build
```

You should be able to use the `./build/firehose-cosmos` binary moving forward.

To install the binary globally, run:

```bash
make install
```

Alternatively, use a prebuilt binary from [Releases Page](https://github.com/figment-networks/firehose-cosmos/releases)

### Docker

You can use our official Docker images: https://hub.docker.com/r/figmentnetworks/firehose-cosmos/tags

```
docker pull figmentnetworks/firehose-cosmos:0.4.0
```

Execute with:

```
docker run --rm -it figmentnetworks/firehose-cosmos:0.4.0 /app/firehose help
```

## Usage

To view usage and flags, run: `./build/firehose-cosmos help`.

```
Usage:
  firehose-cosmos [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Initialize local configuration
  reset       Reset local data directory
  start       Starts all services at once
  tools       Developer tools

Flags:
      --common-auth-plugin string            Auth plugin URI, see streamingfast/dauth repository (default "null://")
      --common-blocks-store-url string       Store URL (with prefix) where to read/write (default "file://{fh-data-dir}/storage/merged-blocks")
      --common-blockstream-addr string       GRPC endpoint to get real-time blocks (default "0.0.0.0:9010")
      --common-first-streamable-block uint   First streamable block number
      --common-metering-plugin string        Metering plugin URI, see streamingfast/dmetering repository (default "null://")
      --common-oneblock-store-url string     Store URL (with prefix) to read/write one-block files (default "file://{fh-data-dir}/storage/one-blocks")
      --common-shutdown-delay duration       Add a delay between receiving SIGTERM signal and shutting down apps. Apps will respond negatively to /healthz during this period (default 5ns)
      --common-startup-delay duration        Delay before launching firehose process
  -c, --config string                        Configuration file for the firehose (default "firehose.yml")
  -d, --data-dir string                      Path to data storage for all components of firehose (default "./fh-data")
  -h, --help                                 help for firehose-cosmos
      --log-format string                    Logging format (default "text")
      --metrics-listen-addr string           If non-empty, the process will listen on this address to server Prometheus metrics (default "0.0.0.0:9102")
      --pprof-listen-addr string             If non-empty, the process will listen on this address for pprof analysis (see https://golang.org/pkg/net/http/pprof/)
  -v, --verbose int                          Enables verbose output (-vvvv for max verbosity) (default 3)

Use "firehose-cosmos [command] --help" for more information about a command.
```

## Configuration

If you wish to use a configuration file instead of setting all CLI flags, you may create a new `firehose.yml`
file in your current working directory.

Example:

```yml
start:
  args:
    - ingestor
    - merger
    - relayer
    - firehose
  flags:
    # Common flags
    common-first-streamable-block: 1

    # Ingestor specific flags
    ingestor-mode: node
    ingestor-node-path: path/to/node/bin
    ingestor-node-args: start --x-crisis-skip-assert-invariants
    ingestor-node-env: "KEY=VALUE,KEY=VALUE"
```

### Logs input mode

It's possible to run the firehose ingestor from the static logs, mostly for development/testing purposes.

Example config:

```yml
start:
  args:
    - ingestor
    # ... other services
  flags:
    # ... other config options

    # Ingestor specific flags
    ingestor-mode: logs
    ingestor-logs-dir: /path/to/logs/dir

    # Configure the pattern if not using .log extension
    # ingestor-logs-pattern: *.log
```

## Supported networks

We provide scripts for running firehose for these networks:

- [Cosmoshub4](devel/cosmoshub4/)
- [Osmosis1](devel/osmosis1/)

### Service Ports

By default, `firehose-cosmos` will start all available services, each providing a GRPC interface.

- `9000` - Ingestor
- `9010` - Relayer
- `9020` - Merger
- `9030` - Firehose

## License

Apache License 2.0
