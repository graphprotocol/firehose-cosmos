package transform

import (
	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	"google.golang.org/protobuf/types/known/anypb"
)

type MessageTypeIndexer struct {
	BlockIndexer blockIndexer
}

func NewMessageTypeIndexer(indexStore dstore.Store, indexSize uint64, startBlock uint64) *MessageTypeIndexer {
	bi := transform.NewBlockIndexer(
		indexStore,
		indexSize,
		MessageTypeIndexShortName,
		transform.WithDefinedStartBlock(startBlock),
	)

	return &MessageTypeIndexer{
		BlockIndexer: bi,
	}
}

func (i *MessageTypeIndexer) ProcessBlock(block *pbcosmos.Block) {
	keyMap := make(map[string]bool)

	for _, tx := range block.Transactions {
		processMessages(keyMap, tx.Tx.Body.Messages)
	}

	var keys []string
	for key := range keyMap {
		keys = append(keys, key)
	}

	i.BlockIndexer.Add(keys, block.Header.Height)
}

func processMessages(keyMap map[string]bool, messages []*anypb.Any) {
	for _, message := range messages {
		keyMap[message.TypeUrl] = true
	}
}
