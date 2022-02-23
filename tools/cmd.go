package tools

import (
	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:   "tools",
		Short: "Developer tools",
	}
)

func init() {
	Cmd.AddCommand(CheckCmd)

	CheckCmd.AddCommand(checkMergedBlocksCmd)
	CheckCmd.PersistentFlags().StringP("range", "r", "", "Block range to use for the check")

	checkMergedBlocksCmd.Flags().BoolP("print-stats", "s", false, "Natively decode each block in the segment and print statistics about it, ensuring it contains the required blocks")
	checkMergedBlocksCmd.Flags().BoolP("print-full", "f", false, "Natively decode each block and print the full JSON representation of the block, should be used with a small range only if you don't want to be overwhelmed")
}
