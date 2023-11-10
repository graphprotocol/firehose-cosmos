package cli

import (
	"github.com/graphprotocol/firehose-cosmos/codec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dlauncher/launcher"
	"go.uber.org/zap"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Starts all services at once",
		RunE:  runStartCommand,
		Args:  cobra.ArbitraryArgs,
	}
)

func runStartCommand(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	var err error
	tracker := bstream.NewTracker(50)

	codec.SetFirstStreamableBlock(viper.GetUint64("common-first-streamable-block"))
	if err := codec.Validate(); err != nil {
		return err
	}

	modules := &launcher.Runtime{
		AbsDataDir: DataDir,
		Tracker:    tracker,
	}

	launch := launcher.NewLauncher(zlog, modules)
	zlog.Debug("launcher created")

	apps := launcher.ParseAppsFromArgs(args, startAppByDefault)
	if len(args) == 0 {
		if launcher.Config[cmd.Name()] == nil {
			launcher.Config[cmd.Name()] = &launcher.CommandConfig{}
		}
		apps = launcher.ParseAppsFromArgs(launcher.Config[cmd.Name()].Args, startAppByDefault)
	}

	if err := launch.Launch(apps); err != nil {
		return err
	}

	signalHandler := derr.SetupSignalHandler(viper.GetDuration("common-shutdown-delay"))
	select {
	case <-signalHandler:
		zlog.Info("Received termination signal, quitting")
		go launch.Close()
	case appID := <-launch.Terminating():
		if launch.Err() == nil {
			zlog.Info("Application %s triggered a clean shutdown, quitting", zap.String("app_id", appID))
		} else {
			zlog.Info("Application %s shutdown unexpectedly, quitting", zap.String("app_id", appID))
			err = launch.Err()
		}
	}

	launch.WaitForTermination()
	return err
}

func startAppByDefault(app string) bool {
	return true
}
