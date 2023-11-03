package cli

import (
	"fmt"

	"github.com/graphprotocol/firehose-cosmos/tools"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dlauncher/flags"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/logging"
)

var (
	zlog, _  = logging.RootLogger("firecosmos", "github.com/graphprotocol/firehose-cosmos/cmd/firecosmos")
	allFlags = map[string]bool{}

	RootCmd = &cobra.Command{
		Use:   "firecosmos",
		Short: "Firehose services for Cosmos blockchains",
		// Version:  // set by cmd/main.go
	}
)

func Main() {
	cobra.OnInitialize(func() {
		allFlags = flags.AutoBind(RootCmd, "FH")
	})

	initCommonFlags(RootCmd.PersistentFlags())
	RootCmd.PersistentPreRunE = preRun

	RootCmd.AddCommand(
		startCommand,
		resetCommand,
		tools.Cmd,
	)

	derr.Check("registering application flags", launcher.RegisterFlags(zlog, startCommand))
	derr.Check("executing root command", RootCmd.Execute())
}

func preRun(cmd *cobra.Command, args []string) error {
	DataDir = viper.GetString("data-dir")

	if launcher.Config == nil {
		launcher.Config = map[string]*launcher.CommandConfig{}
	}

	if configFile := viper.GetString("config"); configFile != "" {
		if fileExists(configFile) {
			if err := launcher.LoadConfigFile(configFile); err != nil {
				return err
			}
		}
	}

	cfg := launcher.Config[cmd.Name()]
	if cfg != nil {
		for k, v := range cfg.Flags {
			validFlag := false
			if _, ok := allFlags[k]; ok {
				viper.SetDefault(k, v)
				validFlag = true
			}
			if !validFlag {
				return fmt.Errorf("invalid flag %v", k)
			}
		}
	}

	launcher.SetupLogger(zlog, &launcher.LoggingOptions{
		WorkingDir:    viper.GetString("data-dir"),
		Verbosity:     viper.GetInt("verbose"),
		LogFormat:     viper.GetString("log-format"),
		LogToFile:     viper.GetBool("log-to-file"),
		LogListenAddr: viper.GetString("log-level-switcher-listen-addr"),
	})

	launcher.SetupTracing("firecosmos")
	launcher.SetupAnalyticsMetrics(zlog, viper.GetString("metrics-listen-addr"), viper.GetString("pprof-listen-addr"))

	return nil
}
