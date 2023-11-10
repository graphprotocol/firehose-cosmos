package codec

import (
	pbcosmos "github.com/graphprotocol/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/logging"
	"google.golang.org/protobuf/proto"
)

var zlog, tracer = logging.PackageLogger("firecosmos", "github.com/streamingfast/firehose-cosmos/codec")

func init() {
	logging.InstantiateLoggers()

	// Should be aligned with firecore.Chain as defined in `cmd/firecosmos/main.go``
	bstream.InitGeneric("CSM", 1, func() proto.Message {
		return new(pbcosmos.Block)
	})
}
