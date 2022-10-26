package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/blockstream"
	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/node-manager/mindreader"

	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	pbheadinfo "github.com/streamingfast/pbgo/sf/headinfo/v1"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/figment-networks/firehose-cosmos/codec"
)

const (
	defaultLineBufferSize = 10 * 1024 * 1024

	modeLogs  = "logs"  // Consume events from the log file(s)
	modeStdin = "stdin" // Consume events from the STDOUT of another process
	modeNode  = "node"  // Consume events from the spawned node process
)

func init() {
	appLogger := zap.NewNop()
	logging.Register("reader", &appLogger)

	registerFlags := func(cmd *cobra.Command) error {
		flags := cmd.Flags()

		flags.String("reader-mode", modeStdin, "Mode of operation, one of (stdin, logs, node)")
		flags.String("reader-logs-dir", "", "Event logs source directory")
		flags.String("reader-logs-pattern", "\\.log(\\.[\\d]+)?", "Logs file pattern")
		flags.Int("reader-line-buffer-size", defaultLineBufferSize, "Buffer size in bytes for the line reader")
		flags.String("reader-working-dir", "{fh-data-dir}/workdir", "Path where reader will stores its files")
		flags.String("reader-grpc-listen-addr", BlockStreamServingAddr, "GRPC server listen address")
		flags.Duration("reader-merge-threshold-block-age", time.Duration(math.MaxInt64), "When processing blocks with a blocktime older than this threshold, they will be automatically merged")
		flags.String("reader-node-path", "", "Path to node binary")
		flags.String("reader-node-dir", "", "Node working directory")
		flags.String("reader-node-args", "", "Node process arguments")
		flags.String("reader-node-env", "", "Node process env vars")
		flags.String("reader-node-logs-filter", "", "Node process log filter expression")
		flags.Duration("reader-wait-upload-complete-on-shutdown", 10*time.Second, "When the reader is shutting down, it will wait up to that amount of time for the archiver to finish uploading the blocks before leaving anyway")

		return nil
	}

	initFunc := func(runtime *launcher.Runtime) (err error) {
		mode := viper.GetString("reader-mode")

		switch mode {
		case modeStdin:
			return nil
		case modeNode:
			return checkNodeBinPath(viper.GetString("reader-node-path"))
		case modeLogs:
			return checkLogsSource(viper.GetString("reader-logs-dir"))
		default:
			return fmt.Errorf("invalid mode: %v", mode)
		}
	}

	factoryFunc := func(runtime *launcher.Runtime) (launcher.App, error) {
		sfDataDir := runtime.AbsDataDir

		oneBlockStoreURL := mustReplaceDataDir(sfDataDir, viper.GetString("common-oneblock-store-url"))
		mergedBlockStoreURL := mustReplaceDataDir(sfDataDir, viper.GetString("common-merged-blocks-store-url"))
		workingDir := mustReplaceDataDir(sfDataDir, viper.GetString("reader-working-dir"))
		gprcListenAdrr := viper.GetString("reader-grpc-listen-addr")
		mergeAndStoreDirectly := viper.GetBool("reader-merge-and-store-directly")
		mergeThresholdBlockAge := viper.GetDuration("reader-merge-threshold-block-age")
		batchStartBlockNum := viper.GetUint64("reader-start-block-num")
		batchStopBlockNum := viper.GetUint64("reader-stop-block-num")
		waitTimeForUploadOnShutdown := viper.GetDuration("reader-wait-upload-complete-on-shutdown")
		oneBlockFileSuffix := viper.GetString("reader-oneblock-suffix")
		blocksChanCapacity := viper.GetInt("reader-blocks-chan-capacity")

		tracker := bstream.NewTracker(50)
		tracker.AddResolver(bstream.OffsetStartBlockResolver(100))

		consoleReaderFactory := func(lines chan string) (mindreader.ConsolerReader, error) {
			return codec.NewConsoleReader(lines, zlog)
		}

		consoleReaderTransformer := func(obj interface{}) (*bstream.Block, error) {
			return codec.FromProto(obj)
		}

		blockStreamServer := blockstream.NewUnmanagedServer(blockstream.ServerOptionWithLogger(appLogger))
		healthCheck := func(ctx context.Context) (isReady bool, out interface{}, err error) {
			return blockStreamServer.Ready(), nil, nil
		}

		server := dgrpc.NewServer2(
			dgrpc.WithLogger(appLogger),
			dgrpc.WithHealthCheck(dgrpc.HealthCheckOverGRPC|dgrpc.HealthCheckOverHTTP, healthCheck),
		)
		server.RegisterService(func(gs *grpc.Server) {
			pbheadinfo.RegisterHeadInfoServer(gs, blockStreamServer)
			pbbstream.RegisterBlockStreamServer(gs, blockStreamServer)
		})

		mrp, err := mindreader.NewMindReaderPlugin(
			oneBlockStoreURL,
			mergedBlockStoreURL,
			mergeAndStoreDirectly,
			mergeThresholdBlockAge,
			workingDir,
			consoleReaderFactory,
			consoleReaderTransformer,
			tracker,
			batchStartBlockNum,
			batchStopBlockNum,
			blocksChanCapacity,
			headBlockUpdater,
			func(error) {},
			true,
			waitTimeForUploadOnShutdown,
			oneBlockFileSuffix,
			blockStreamServer,
			appLogger,
		)
		if err != nil {
			log.Fatal("error initialising reader", zap.Error(err))
			return nil, nil
		}

		return &ReaderApp{
			Shutter:          shutter.New(),
			mrp:              mrp,
			mode:             viper.GetString("reader-mode"),
			lineBufferSize:   viper.GetInt("reader-line-buffer-size"),
			nodeBinPath:      viper.GetString("reader-node-path"),
			nodeDir:          viper.GetString("reader-node-dir"),
			nodeArgs:         viper.GetString("reader-node-args"),
			nodeEnv:          viper.GetString("reader-node-env"),
			nodeLogsFilter:   viper.GetString("reader-node-logs-filter"),
			logsDir:          viper.GetString("reader-logs-dir"),
			logsFilePattern:  viper.GetString("reader-logs-pattern"),
			server:           server,
			serverListenAddr: gprcListenAdrr,
		}, nil
	}

	launcher.RegisterApp(&launcher.AppDef{
		ID:            "reader",
		Title:         "Reader",
		Description:   "Reads the log files produced by the instrumented node",
		MetricsID:     "reader",
		Logger:        launcher.NewLoggingDef("reader.*", nil),
		RegisterFlags: registerFlags,
		InitFunc:      initFunc,
		FactoryFunc:   factoryFunc,
	})
}

func headBlockUpdater(uint64, string, time.Time) {
	// TODO: will need to be implemented somewhere
}

func checkLogsSource(dir string) error {
	if dir == "" {
		return errors.New("reader logs dir must be set")
	}

	dir, err := expandDir(dir)
	if err != nil {
		return err
	}

	if !dirExists(dir) {
		return errors.New("reader logs dir must exist")
	}

	return nil
}

func checkNodeBinPath(binPath string) error {
	if binPath == "" {
		return errors.New("node path must be set")
	}

	stat, err := os.Stat(binPath)
	if err != nil {
		return fmt.Errorf("cant inspect node path: %w", err)
	}

	if stat.IsDir() {
		return fmt.Errorf("path %v is a directory", binPath)
	}

	return nil
}
