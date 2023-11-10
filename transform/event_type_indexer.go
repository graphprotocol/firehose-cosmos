package transform

import (
	pbcosmos "github.com/graphprotocol/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	firecore "github.com/streamingfast/firehose-core"
)

type blockIndexer interface {
	Add(keys []string, blockNum uint64)
}

type EventTypeIndexer struct {
	BlockIndexer blockIndexer
}

func NewEventTypeIndexer(indexStore dstore.Store, indexSize uint64) (firecore.BlockIndexer[*pbcosmos.Block], error) {
	bi := transform.NewBlockIndexer(
		indexStore,
		indexSize,
		EventTypeIndexShortName,
	)

	return &EventTypeIndexer{
		BlockIndexer: bi,
	}, nil
}

func (i *EventTypeIndexer) ProcessBlock(block *pbcosmos.Block) error {
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

	return nil
}

func processEvents(keyMap map[string]bool, events []*pbcosmos.Event) {
	for _, event := range events {
		keyMap[event.EventType] = true
	}
}
