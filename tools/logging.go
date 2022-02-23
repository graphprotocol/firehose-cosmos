package tools

import (
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var (
	traceEnabled = logging.IsTraceEnabled("tools", "github.com/figment-networks/firehose-tendermint/tools")
	zlog         = zap.NewNop()
)

func init() {
	logging.Register("github.com/figment-networks/firehose-tendermint/tools", &zlog)
}
