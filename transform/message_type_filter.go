package transform

import (
	"fmt"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	pbtransform "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/transform/v1"
	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
)

var MessageTypeFilterMessageName = proto.MessageName(&pbtransform.MessageTypeFilter{})

func MessageTypeFilterFactory(indexStore dstore.Store, possibleIndexSizes []uint64) *transform.Factory {
	return &transform.Factory{
		Obj: &pbtransform.MessageTypeFilter{},
		NewFunc: func(message *anypb.Any) (transform.Transform, error) {
			if message.MessageName() != MessageTypeFilterMessageName {
				return nil, fmt.Errorf("expected type url %q, received %q", MessageTypeFilterMessageName, message.TypeUrl)
			}

			filter := &pbtransform.MessageTypeFilter{}
			err := proto.Unmarshal(message.Value, filter)
			if err != nil {
				return nil, fmt.Errorf("unexpected unmarshal error: %w", err)
			}

			if len(filter.MessageTypes) == 0 {
				return nil, fmt.Errorf("message filter requires at least one message type")
			}

			messageTypeMap := make(map[string]bool)
			for _, acc := range filter.MessageTypes {
				messageTypeMap[acc] = true
			}

			return &MessageTypeFilter{
				MessageTypes:       messageTypeMap,
				possibleIndexSizes: possibleIndexSizes,
				indexStore:         indexStore,
			}, nil
		},
	}
}

type MessageTypeFilter struct {
	MessageTypes map[string]bool

	indexStore         dstore.Store
	possibleIndexSizes []uint64
}

func (p *MessageTypeFilter) String() string {
	return fmt.Sprintf("%v", p.MessageTypes)
}

func (p *MessageTypeFilter) Transform(readOnlyBlk *bstream.Block, in transform.Input) (transform.Output, error) {
	block := readOnlyBlk.ToProtocol().(*pbcosmos.Block)

	for _, tx := range block.Transactions {
		tx.Tx.Body.Messages = p.filterMessages(tx.Tx.Body.Messages)
	}

	return block, nil
}

func (p *MessageTypeFilter) filterMessages(messages []*anypb.Any) []*anypb.Any {
	var outMessages []*anypb.Any

	for _, message := range messages {
		if p.MessageTypes[message.TypeUrl] {
			outMessages = append(outMessages, message)
		}
	}

	return outMessages
}

func (p *MessageTypeFilter) GetIndexProvider() bstream.BlockIndexProvider {
	if p.indexStore == nil {
		return nil
	}

	if len(p.MessageTypes) == 0 {
		return nil
	}

	return NewMessageTypeIndexProvider(
		p.indexStore,
		p.possibleIndexSizes,
		p.MessageTypes,
	)
}
