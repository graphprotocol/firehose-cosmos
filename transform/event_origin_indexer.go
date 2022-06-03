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
	keyMap := make(map[EventOrigin]bool)

	if len(block.ResultBeginBlock.Events) > 0 {
		keyMap[BeginBlock] = true
	}

	if len(block.ResultEndBlock.Events) > 0 {
		keyMap[EndBlock] = true
	}

	for _, tx := range block.Transactions {
		if len(tx.Result.Events) > 0 {
			keyMap[DeliverTx] = true
			break //exit the loop as soon as we hit this once
		}
	}

	var keys []string
	for key := range keyMap {
		keys = append(keys, string(key))
	}

	i.BlockIndexer.Add(keys, block.Header.Height)
}
