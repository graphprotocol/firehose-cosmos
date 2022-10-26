package cli

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dlauncher/launcher"
	relayerApp "github.com/streamingfast/relayer/app/relayer"
)

func init() {
	registerFlags := func(cmd *cobra.Command) error {
		cmd.Flags().String("relayer-grpc-listen-addr", RelayerServingAddr, "Address to listen for incoming gRPC requests")
		cmd.Flags().StringSlice("relayer-source", []string{BlockStreamServingAddr}, "List of Blockstream sources (ingestors) to connect to for live block feeds (repeat flag as needed)")
		cmd.Flags().Duration("relayer-max-source-latency", 999999*time.Hour, "Max latency tolerated to connect to a source. A performance optimization for when you have redundant sources and some may not have caught up")
		return nil
	}

	factoryFunc := func(runtime *launcher.Runtime) (launcher.App, error) {
		sources := viper.GetStringSlice("relayer-source")
		if len(sources) == 0 {
			return nil, errors.New("relayer sources are empty")
		}
		_, oneBlocksStoreURL, _, err := GetCommonStoresURLs(runtime.AbsDataDir)
		if err != nil {
			return nil, err
		}
		return relayerApp.New(&relayerApp.Config{
			SourcesAddr:      sources,
			OneBlocksURL:     oneBlocksStoreURL,
			GRPCListenAddr:   viper.GetString("relayer-grpc-listen-addr"),
			MaxSourceLatency: viper.GetDuration("relayer-max-source-latency"),
		}), nil

	}

	launcher.RegisterApp(zlog, &launcher.AppDef{
		ID:            "relayer",
		Title:         "Relayer",
		Description:   "Serves blocks as a stream, with a buffer",
		RegisterFlags: registerFlags,
		FactoryFunc:   factoryFunc,
	})
}
