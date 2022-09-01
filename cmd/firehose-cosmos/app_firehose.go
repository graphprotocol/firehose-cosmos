package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	sftransform "github.com/figment-networks/firehose-cosmos/transform"
	"github.com/streamingfast/bstream/transform"
	dauthAuthenticator "github.com/streamingfast/dauth/authenticator"
	_ "github.com/streamingfast/dauth/authenticator/null"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/dmetering"
	"github.com/streamingfast/dmetrics"
	firehoseApp "github.com/streamingfast/firehose/app/firehose"
	"github.com/streamingfast/logging"
)

var (
	metricset           = dmetrics.NewSet()
	headBlockNumMetric  = metricset.NewHeadBlockNumber("firehose")
	headTimeDriftmetric = metricset.NewHeadTimeDrift("firehose")
)

func init() {
	appLogger := zap.NewNop()
	logging.Register("firehose", &appLogger)

	registerFlags := func(cmd *cobra.Command) error {
		cmd.Flags().String("firehose-grpc-listen-addr", FirehoseGRPCServingAddr, "Address on which the firehose will listen")
		return nil
	}

	factoryFunc := func(runtime *launcher.Runtime) (launcher.App, error) {
		// Configure block stream connection (to node or relayer)
		blockstreamAddr := viper.GetString("common-live-blocks-addr")
		if blockstreamAddr == "-" {
			blockstreamAddr = ""
		}

		// Configure authentication (default is no auth)
		authenticator, err := dauthAuthenticator.New(viper.GetString("common-auth-plugin"))
		if err != nil {
			return nil, fmt.Errorf("unable to initialize dauth: %w", err)
		}

		// Metering? (TODO: what is this for?)
		metering, err := dmetering.New(viper.GetString("common-metering-plugin"))
		if err != nil {
			return nil, fmt.Errorf("unable to initialize dmetering: %w", err)
		}
		dmetering.SetDefaultMeter(metering)

		mergedBlocksStoreURL, oneBlocksStoreURL, forkedBlocksStoreURL, err := GetCommonStoresURLs(runtime.AbsDataDir)
		if err != nil {
			return nil, err
		}
		indexStore, possibleIndexSizes, err := GetIndexStore(runtime.AbsDataDir)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize indexes: %w", err)
		}

		// Configure graceful shutdown
		shutdownSignalDelay := viper.GetDuration("common-system-shutdown-signal-delay")
		grcpShutdownGracePeriod := time.Duration(0)
		if shutdownSignalDelay.Seconds() > 5 {
			grcpShutdownGracePeriod = shutdownSignalDelay - (5 * time.Second)
		}

		registry := transform.NewRegistry()
		registry.Register(sftransform.EventOriginFilterFactory(indexStore, possibleIndexSizes))
		registry.Register(sftransform.EventTypeFilterFactory(indexStore, possibleIndexSizes))
		registry.Register(sftransform.MessageTypeFilterFactory(indexStore, possibleIndexSizes))

		return firehoseApp.New(appLogger,
			&firehoseApp.Config{
				MergedBlocksStoreURL:    mergedBlocksStoreURL,
				OneBlocksStoreURL:       oneBlocksStoreURL,
				ForkedBlocksStoreURL:    forkedBlocksStoreURL,
				BlockStreamAddr:         blockstreamAddr,
				GRPCListenAddr:          viper.GetString("firehose-grpc-listen-addr"),
				GRPCShutdownGracePeriod: grcpShutdownGracePeriod,
			},
			&firehoseApp.Modules{
				Authenticator:         authenticator,
				HeadTimeDriftMetric:   headTimeDriftmetric,
				HeadBlockNumberMetric: headBlockNumMetric,
				TransformRegistry:     registry,
			}), nil
	}

	launcher.RegisterApp(&launcher.AppDef{
		ID:            "firehose",
		Title:         "Block Firehose",
		Description:   "Provides on-demand filtered blocks, depends on common-merged-blocks-store-url, common-forked-blocks-store-url and common-live-blocks-addr",
		MetricsID:     "firehose",
		Logger:        launcher.NewLoggingDef("firehose.*", nil),
		RegisterFlags: registerFlags,
		FactoryFunc:   factoryFunc,
	})
}
