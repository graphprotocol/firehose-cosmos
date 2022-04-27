package transform

import (
	pbcodec "github.com/figment-networks/tendermint-protobuf-def/pb/fig/tendermint/codec/v1"
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

func (i *EventTypeIndexer) ProcessBlock(block *pbcodec.EventList) {
	keyMap := make(map[string]bool)

	processEvents(keyMap, block.NewBlock.ResultBeginBlock.Events)
	processEvents(keyMap, block.NewBlock.ResultEndBlock.Events)

	for _, tx := range block.Transaction {
		processEvents(keyMap, tx.TxResult.Result.Events)
	}

	var keys []string
	for key := range keyMap {
		keys = append(keys, key)
	}

	i.BlockIndexer.Add(keys, block.NewBlock.Block.Header.Height)
}

func processEvents(keyMap map[string]bool, events []*pbcodec.Event) {
	for _, event := range events {
		keyMap[event.EventType] = true
	}
}
