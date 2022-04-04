package tools

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	sftools "github.com/streamingfast/sf-tools"
)

var (
	CheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Various checks for deployment, data integrity & debugging",
	}

	checkMergedBlocksCmd = &cobra.Command{
		Use:     "merged-blocks {store-url}",
		Short:   "Checks for any holes in merged blocks as well as ensuring merged blocks integrity",
		Args:    cobra.ExactArgs(1),
		PreRunE: initFirstStreamable,
		RunE:    checkMergedBlocksE,
	}
)

func checkMergedBlocksE(cmd *cobra.Command, args []string) error {
	storeURL := args[0]
	fileBlockSize := uint32(100)

	blockRange, err := sftools.Flags.GetBlockRange("range")
	if err != nil {
		return err
	}

	printDetails := sftools.PrintNothing
	if viper.GetBool("print-stats") {
		printDetails = sftools.PrintStats
	}

	if viper.GetBool("print-full") {
		printDetails = sftools.PrintFull
	}

	return sftools.CheckMergedBlocks(cmd.Context(), zlog, storeURL, fileBlockSize, blockRange, blockPrinter, printDetails)
}

func blockPrinter(block *bstream.Block) {
	fmt.Printf("Block %s, Prev: %s: %d shards, %d transactions\n",
		block.AsRef(),
		block.PreviousRef(),
		0, 0,
	)
}
