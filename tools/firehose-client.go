package tools

import (
	sftools "github.com/streamingfast/sf-tools"
)

func init() {
	firehoseClientCmd := sftools.GetFirehoseClientCmd(zlog, tracer, nil)
	Cmd.AddCommand(firehoseClientCmd)

	firehoseSingleBlockClientCmd := sftools.GetFirehoseSingleBlockClientCmd(zlog, tracer)
	Cmd.AddCommand(firehoseSingleBlockClientCmd)
}
