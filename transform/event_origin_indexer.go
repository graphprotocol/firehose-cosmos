package transform

import (
	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
)

type EventOriginIndexer struct {
	BlockIndexer blockIndexer
}

func NewEventOriginIndexer(indexStore dstore.Store, indexSize uint64, startBlock uint64) *EventOriginIndexer {
	bi := transform.NewBlockIndexer(
		indexStore,
		indexSize,
		EventOriginIndexShortName,
		transform.WithDefinedStartBlock(startBlock),
	)

	return &EventOriginIndexer{
		BlockIndexer: bi,
	}
}

func (i *EventOriginIndexer) ProcessBlock(block *pbcosmos.Block) {
	keyMap := make(map[string]bool)

	if len(block.Transactions) > 0 {
		keyMap["DeliverTx"] = true
	}

	if len(block.ResultBeginBlock.Events) > 0 {
		keyMap["BeginBlock"] = true
	}

	if len(block.ResultEndBlock.Events) > 0 {
		keyMap["EndBlock"] = true
	}

	var keys []string
	for key := range keyMap {
		keys = append(keys, key)
	}

	i.BlockIndexer.Add(keys, block.Header.Height)
}
