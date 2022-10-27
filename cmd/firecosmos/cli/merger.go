package cli

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dlauncher/launcher"
	mergerApp "github.com/streamingfast/merger/app/merger"
)

func init() {
	flags := func(cmd *cobra.Command) error {
		cmd.Flags().Duration("merger-time-between-store-lookups", 1*time.Second, "Delay between source store polling (should be higher for remote storage)")
		cmd.Flags().Duration("merger-time-between-store-pruning", time.Minute, "Delay between source store pruning loops")
		cmd.Flags().String("merger-grpc-listen-addr", MergerServingAddr, "Address to listen for incoming gRPC requests")
		cmd.Flags().Uint64("merger-stop-block", 0, "if non-zero, merger will trigger shutdown when blocks have been merged up to this block")
		return nil
	}

	initFunc := func(runtime *launcher.Runtime) error {
		return nil
	}

	factoryFunc := func(runtime *launcher.Runtime) (launcher.App, error) {

		mergedBlocksStoreURL, oneBlocksStoreURL, err := GetCommonStoresURLs(runtime.AbsDataDir)
		if err != nil {
			return nil, err
		}
		return mergerApp.New(&mergerApp.Config{

			StorageOneBlockFilesPath:     oneBlocksStoreURL,
			StorageMergedBlocksFilesPath: mergedBlocksStoreURL,
			StorageForkedBlocksFilesPath: ".",
			GRPCListenAddr:               viper.GetString("merger-grpc-listen-addr"),
			PruneForkedBlocksAfter:       99999999999, // no forked blocks here
			StopBlock:                    viper.GetUint64("merger-stop-block"),
			TimeBetweenPruning:           viper.GetDuration("merger-time-between-store-pruning"),
			TimeBetweenPolling:           viper.GetDuration("merger-time-between-store-lookups"),
		}), nil
	}

	launcher.RegisterApp(zlog, &launcher.AppDef{
		ID:            "merger",
		Title:         "Merger",
		Description:   "Produces merged block files from single-block files",
		RegisterFlags: flags,
		InitFunc:      initFunc,
		FactoryFunc:   factoryFunc,
	})
}
