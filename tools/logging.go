package tools

import (
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var (
	traceEnabled = logging.IsTraceEnabled("tools", "tools")
	zlog         = zap.NewNop()
)

func init() {
	logging.Register("tools", &zlog)
}
