package transform

import (
	pbcosmos "github.com/graphprotocol/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	firecore "github.com/streamingfast/firehose-core"
)

func NewEventOriginIndexer(indexStore dstore.Store, indexSize uint64) (firecore.BlockIndexer[*pbcosmos.Block], error) {
	bi := transform.NewBlockIndexer(
		indexStore,
		indexSize,
		EventOriginIndexShortName,
	)

	return &EventOriginIndexer{
		BlockIndexer: bi,
	}, nil
}

type EventOriginIndexer struct {
	BlockIndexer blockIndexer
}

func (i *EventOriginIndexer) ProcessBlock(block *pbcosmos.Block) error {
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

	return nil
}
