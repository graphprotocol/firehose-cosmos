package transform

import (
	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
)

type blockIndexer interface {
	Add(keys []string, blockNum uint64)
}

type EventTypeIndexer struct {
	BlockIndexer blockIndexer
}

func NewEventTypeIndexer(indexStore dstore.Store, indexSize uint64, startBlock uint64) *EventTypeIndexer {
	bi := transform.NewBlockIndexer(
		indexStore,
		indexSize,
		EventTypeIndexShortName,
		transform.WithDefinedStartBlock(startBlock),
	)

	return &EventTypeIndexer{
		BlockIndexer: bi,
	}
}

func (i *EventTypeIndexer) ProcessBlock(block *pbcosmos.Block) {
	keyMap := make(map[string]bool)

	processEvents(keyMap, block.ResultBeginBlock.Events)
	processEvents(keyMap, block.ResultEndBlock.Events)

	for _, tx := range block.Transactions {
		processEvents(keyMap, tx.Result.Events)
	}

	var keys []string
	for key := range keyMap {
		keys = append(keys, key)
	}

	i.BlockIndexer.Add(keys, block.Header.Height)
}

func processEvents(keyMap map[string]bool, events []*pbcosmos.Event) {
	for _, event := range events {
		keyMap[event.EventType] = true
	}
}
