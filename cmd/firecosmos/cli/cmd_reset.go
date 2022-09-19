package main

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	resetCommand = &cobra.Command{
		Use:   "reset",
		Short: "Reset local data directory",
		RunE:  runResetCommand,
	}
)

func runResetCommand(cmd *cobra.Command, args []string) error {
	if !dirExists(DataDir) {
		zlog.Info("data directory does not exist, skipping", zap.String("dir", DataDir))
		return nil
	}

	zlog.Info("removing data directory", zap.String("dir", DataDir))
	return os.RemoveAll(DataDir)
}
