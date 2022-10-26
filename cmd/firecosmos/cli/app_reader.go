package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream/blockstream"
	dgrpcserver "github.com/streamingfast/dgrpc/server"
	dgrpcfactory "github.com/streamingfast/dgrpc/server/factory"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/logging"
	nodeManager "github.com/streamingfast/node-manager"
	"github.com/streamingfast/node-manager/metrics"
	"github.com/streamingfast/node-manager/mindreader"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	pbheadinfo "github.com/streamingfast/pbgo/sf/headinfo/v1"

	"github.com/streamingfast/shutter"
	"go.uber.org/zap"

	"github.com/figment-networks/firehose-cosmos/codec"
)

const (
	defaultLineBufferSize = 10 * 1024 * 1024

	modeLogs  = "logs"  // Consume events from the log file(s)
	modeStdin = "stdin" // Consume events from the STDOUT of another process
	modeNode  = "node"  // Consume events from the spawned node process
)

var readerLogger, readerTracer = logging.PackageLogger("reader", "github.com/figment-network/firehose-cosmos/noderunner")

func init() {
	appLogger := readerLogger
	appTracer := readerTracer

	registerFlags := func(cmd *cobra.Command) error {
		flags := cmd.Flags()

		flags.String("reader-mode", modeStdin, "Mode of operation, one of (stdin, logs, node)")
		flags.String("reader-logs-dir", "", "Event logs source directory")
		flags.String("reader-logs-pattern", "\\.log(\\.[\\d]+)?", "Logs file pattern")
		flags.Int("reader-line-buffer-size", defaultLineBufferSize, "Buffer size in bytes for the line reader")
		flags.Duration("reader-readiness-max-latency", 2*time.Minute, "Determine the maximum head block latency at which the instance will be determined healthy. Some chains have more regular block production than others.")
		flags.String("reader-working-dir", "{fh-data-dir}/workdir", "Path where reader will stores its files")
		flags.String("reader-grpc-listen-addr", BlockStreamServingAddr, "GRPC server listen address")
		flags.String("reader-node-path", "", "Path to node binary")
		flags.String("reader-node-dir", "", "Node working directory")
		flags.String("reader-node-args", "", "Node process arguments")
		flags.String("reader-node-env", "", "Node process env vars")
		flags.String("reader-node-logs-filter", "", "Node process log filter expression")
		flags.String("reader-oneblock-suffix", "default", "suffix appended to one-block-files to identify a specific reader process in redundant mode")

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

		_, oneBlockStoreURL, _, err := GetCommonStoresURLs(runtime.AbsDataDir)
		workingDir := MustReplaceDataDir(sfDataDir, viper.GetString("reader-working-dir"))
		gprcListenAdrr := viper.GetString("reader-grpc-listen-addr")
		batchStartBlockNum := viper.GetUint64("reader-start-block-num")
		batchStopBlockNum := viper.GetUint64("reader-stop-block-num")
		oneBlockFileSuffix := viper.GetString("reader-oneblock-suffix")
		blocksChanCapacity := viper.GetInt("reader-blocks-chan-capacity")
		readinessMaxLatency := viper.GetDuration("reader-readiness-max-latency")

		consoleReaderFactory := func(lines chan string) (mindreader.ConsolerReader, error) {
			return codec.NewConsoleReader(lines, zlog)
		}

		blockStreamServer := blockstream.NewUnmanagedServer(blockstream.ServerOptionWithLogger(appLogger))
		healthCheck := func(ctx context.Context) (isReady bool, out interface{}, err error) {
			return blockStreamServer.Ready(), nil, nil
		}

		server := dgrpcfactory.ServerFromOptions(
			dgrpcserver.WithLogger(appLogger),
			dgrpcserver.WithHealthCheck(dgrpcserver.HealthCheckOverGRPC|dgrpcserver.HealthCheckOverHTTP, healthCheck),
		)
		//		err := a.modules.RegisterGRPCService(gs.ServiceRegistrar())

		sr := server.ServiceRegistrar()
		sr.RegisterService(&pbheadinfo.HeadInfo_ServiceDesc, blockStreamServer)
		sr.RegisterService(&pbbstream.BlockStream_ServiceDesc, blockStreamServer)

		metricsAndReadinessManager := buildMetricsAndReadinessManager("reader", readinessMaxLatency)

		mrp, err := mindreader.NewMindReaderPlugin(
			oneBlockStoreURL,
			workingDir,
			consoleReaderFactory,
			batchStartBlockNum,
			batchStopBlockNum,
			blocksChanCapacity,
			metricsAndReadinessManager.UpdateHeadBlock,
			func(error) {},
			oneBlockFileSuffix,
			blockStreamServer,
			appLogger,
			appTracer,
		)
		if err != nil {
			zlog.Error("error initialising reader", zap.Error(err))
			return nil, err
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

	launcher.RegisterApp(zlog, &launcher.AppDef{
		ID:            "reader",
		Title:         "Reader",
		Description:   "Reads the log files produced by the instrumented node",
		RegisterFlags: registerFlags,
		InitFunc:      initFunc,
		FactoryFunc:   factoryFunc,
	})
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

func buildMetricsAndReadinessManager(name string, maxLatency time.Duration) *nodeManager.MetricsAndReadinessManager {
	headBlockTimeDrift := metrics.NewHeadBlockTimeDrift(name)
	headBlockNumber := metrics.NewHeadBlockNumber(name)
	appReadiness := metrics.NewAppReadiness(name)

	metricsAndReadinessManager := nodeManager.NewMetricsAndReadinessManager(
		headBlockTimeDrift,
		headBlockNumber,
		appReadiness,
		maxLatency,
	)
	return metricsAndReadinessManager
}
