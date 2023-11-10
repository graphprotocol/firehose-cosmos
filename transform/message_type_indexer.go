package transform

import (
	pbcosmos "github.com/graphprotocol/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	firecore "github.com/streamingfast/firehose-core"
	"google.golang.org/protobuf/types/known/anypb"
)

type MessageTypeIndexer struct {
	BlockIndexer blockIndexer
}

func NewMessageTypeIndexer(indexStore dstore.Store, indexSize uint64) (firecore.BlockIndexer[*pbcosmos.Block], error) {
	bi := transform.NewBlockIndexer(
		indexStore,
		indexSize,
		MessageTypeIndexShortName,
	)

	return &MessageTypeIndexer{
		BlockIndexer: bi,
	}, nil
}

func (i *MessageTypeIndexer) ProcessBlock(block *pbcosmos.Block) error {
	keyMap := make(map[string]bool)

	for _, tx := range block.Transactions {
		processMessages(keyMap, tx.Tx.Body.Messages)
	}

	var keys []string
	for key := range keyMap {
		keys = append(keys, key)
	}

	i.BlockIndexer.Add(keys, block.Header.Height)

	return nil
}

func processMessages(keyMap map[string]bool, messages []*anypb.Any) {
	for _, message := range messages {
		keyMap[message.TypeUrl] = true
	}
}
