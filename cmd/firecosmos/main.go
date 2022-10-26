package main

import (
	"github.com/figment-networks/firehose-cosmos/cmd/firecosmos/cli"
)

// Commit sha1 value, injected via go build `ldflags` at build time
var commit = ""

// Version value, injected via go build `ldflags` at build time
var version = "0.6.0"

// Date value, injected via go build `ldflags` at build time
var date = ""

func init() {
	cli.RootCmd.Version = cli.VersionString(version, commit, date)
}

func main() {
	cli.Main()
}
