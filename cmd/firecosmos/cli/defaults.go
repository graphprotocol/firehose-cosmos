package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/streamingfast/dlauncher/launcher"
	"go.uber.org/zap"
)

var (
	// Data dir (for local operations only)
	DataDir = "./fh-data"

	// GRPC Service Addresses
	BlockStreamServingAddr  = "0.0.0.0:9000"
	RelayerServingAddr      = "0.0.0.0:9010"
	MergerServingAddr       = "0.0.0.0:9020"
	FirehoseGRPCServingAddr = "0.0.0.0:9030"
	MetricsServingAddr      = "0.0.0.0:9102"

	// Blocks store
	MergedBlocksStoreURL string = "file://{fh-data-dir}/storage/merged-blocks"
	OneBlockStoreURL     string = "file://{fh-data-dir}/storage/one-blocks"

	IndexStoreURL   string = "file://{fh-data-dir}/storage/block-index"
	IndexStoreSizes        = []int{100000, 10000, 1000}

	// Protocol defaults
	FirstStreamableBlock uint64 = 0
)

func init() {
	launcher.RegisterCommonFlags = func(zlog *zap.Logger, cmd *cobra.Command) error {
		initCommonFlags(cmd.Flags())
		return nil
	}
}

func initCommonFlags(flags *pflag.FlagSet) {
	// General
	flags.IntP("verbose", "v", 3, "Enables verbose output (-vvvv for max verbosity)")
	flags.String("log-format", "text", "Logging format")
	flags.StringP("config", "c", "firehose.yml", "Configuration file for the Firehose")
	flags.StringP("data-dir", "d", DataDir, "Path to data storage for all components of Firehose")

	// Analytics, metrics
	flags.String("metrics-listen-addr", MetricsServingAddr, "If non-empty, the process will listen on this address to server Prometheus metrics")
	flags.String("pprof-listen-addr", "", "If non-empty, the process will listen on this address for pprof analysis (see https://golang.org/pkg/net/http/pprof/)")

	// Common stores configuration flags
	flags.String("common-merged-blocks-store-url", MergedBlocksStoreURL, "Store URL (with prefix) where to read/write")
	flags.String("common-one-block-store-url", OneBlockStoreURL, "Store URL (with prefix) to read/write one-block files")
	flags.String("common-live-blocks-addr", RelayerServingAddr, "GRPC endpoint to get real-time blocks")
	flags.Uint64("common-first-streamable-block", FirstStreamableBlock, "First streamable block number")

	flags.String("common-index-store-url", IndexStoreURL, "[COMMON] Store URL (with prefix) to read/write index files.")
	flags.IntSlice("common-block-index-sizes", []int{100000, 100000, 10000, 1000}, "index bundle sizes that that are considered valid when looking for block indexes")

	// Authentication, metering and rate limiter plugins
	flags.String("common-auth-plugin", "null://", "Auth plugin URI, see streamingfast/dauth repository")
	flags.String("common-metering-plugin", "null://", "Metering plugin URI, see streamingfast/dmetering repository")

	// System Behavior
	flags.Duration("common-startup-delay", 0, "Delay before launching Firehose process")
	flags.Duration("common-shutdown-delay", 5, "Add a delay between receiving SIGTERM signal and shutting down apps. Apps will respond negatively to /healthz during this period")
}
