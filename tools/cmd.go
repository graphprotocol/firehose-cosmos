package tools

import (
	"fmt"

	"github.com/figment-networks/firehose-cosmos/codec"
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

func initFirstStreamable(cmd *cobra.Command, args []string) error {
	codec.SetFirstStreamableBlock(mustGetUint64(cmd, "common-first-streamable-block"))
	return codec.Validate()
}
func mustGetString(cmd *cobra.Command, flagName string) string {
	val, err := cmd.Flags().GetString(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}
func mustGetInt64(cmd *cobra.Command, flagName string) int64 {
	val, err := cmd.Flags().GetInt64(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}
func mustGetUint64(cmd *cobra.Command, flagName string) uint64 {
	val, err := cmd.Flags().GetUint64(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}
func mustGetBool(cmd *cobra.Command, flagName string) bool {
	val, err := cmd.Flags().GetBool(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}
