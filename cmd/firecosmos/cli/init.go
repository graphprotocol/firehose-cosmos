package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	initCommand = &cobra.Command{
		Use:   "init",
		Short: "Initialize local configuration",
		RunE:  runInitCommand,
	}
)

func runInitCommand(cmd *cobra.Command, args []string) error {
	if err := mkdirStorePathIfLocal(MustReplaceDataDir(DataDir, viper.GetString("common-merged-blocks-store-url"))); err != nil {
		return err
	}

	if err := mkdirStorePathIfLocal(MustReplaceDataDir(DataDir, viper.GetString("common-oneblock-store-url"))); err != nil {
		return err
	}

	return mkdirStorePathIfLocal(MustReplaceDataDir(DataDir, viper.GetString("merger-state-file")))
}
