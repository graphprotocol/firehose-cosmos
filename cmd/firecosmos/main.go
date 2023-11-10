package main

import (
	"github.com/graphprotocol/firehose-cosmos/codec"
	"github.com/graphprotocol/firehose-cosmos/transform"
	pbcosmos "github.com/graphprotocol/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/spf13/pflag"
	firecore "github.com/streamingfast/firehose-core"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func init() {
	firecore.UnsafePayloadKind = pbbstream.Protocol_COSMOS
	firecore.UnsafeJsonBytesEncoder = "hex"
}

func main() {
	firecore.Main(Chain)
}

var Chain = &firecore.Chain[*pbcosmos.Block]{
	ShortName:            "cosmos",
	LongName:             "Cosmos",
	ExecutableName:       "",
	FullyQualifiedModule: "github.com/streamingfast/firehose-ethereum",
	Version:              version,

	Protocol:        "CSM",
	ProtocolVersion: 1,

	BlockFactory: func() firecore.Block { return new(pbcosmos.Block) },

	BlockIndexerFactories: map[string]firecore.BlockIndexerFactory[*pbcosmos.Block]{
		transform.EventOriginIndexShortName: transform.NewEventOriginIndexer,
		transform.EventTypeIndexShortName:   transform.NewEventTypeIndexer,
		transform.MessageTypeIndexShortName: transform.NewMessageTypeIndexer,
	},

	BlockTransformerFactories: map[protoreflect.FullName]firecore.BlockTransformerFactory{
		transform.EventOriginFilterMessageName: transform.NewEventOriginFilterFactory,
		transform.EventTypeFilterMessageName:   transform.NewEventTypeFilterFactory,
		transform.MessageTypeFilterMessageName: transform.NewMessageTypeFilterFactory,
	},

	ConsoleReaderFactory: codec.NewConsoleReader,

	RegisterExtraStartFlags: func(flags *pflag.FlagSet) {
	},

	Tools: &firecore.ToolsConfig[*pbcosmos.Block]{
		BlockPrinter: printBlock,

		// FIXME: Implement so that `firecosmos tools firehose-client` have the correct flags for providing transforms
		// TransformFlags: &firecore.TransformFlags{
		// 	Register: func(flags *pflag.FlagSet) {
		// 		flags.Bool("header-only", false, "Apply the HeaderOnly transform sending back Block's header only (with few top-level fields), exclusive option")
		// 		flags.String("call-filters", "", "call filters (format: '[address1[+address2[+...]]]:[eventsig1[+eventsig2[+...]]]")
		// 		flags.String("log-filters", "", "log filters (format: '[address1[+address2[+...]]]:[eventsig1[+eventsig2[+...]]]")
		// 		flags.Bool("send-all-block-headers", false, "ask for all the blocks to be sent (header-only if there is no match)")
		// 	},

		// 	Parse: parseTransformFlags,
		// },
	},
}

// Version value, injected via go build `ldflags` at build time, **must** not be removed or inlined
var version = "dev"
