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

var EventOriginFilterMessageName = proto.MessageName(&pbtransform.EventOriginFilter{})

func EventOriginFilterFactory(indexStore dstore.Store, possibleIndexSizes []uint64) *transform.Factory {
	return &transform.Factory{
		Obj: &pbtransform.EventOriginFilter{},
		NewFunc: func(message *anypb.Any) (transform.Transform, error) {
			if message.MessageName() != EventOriginFilterMessageName {
				return nil, fmt.Errorf("expected type url %q, received %q", EventOriginFilterMessageName, message.TypeUrl) // double check this
			}

			filter := &pbtransform.EventOriginFilter{}
			err := proto.Unmarshal(message.Value, filter)
			if err != nil {
				return nil, fmt.Errorf("unexpected unmarshal error: %w", err)
			}

			if len(filter.EventOrigin) == 0 {
				return nil, fmt.Errorf("event origin filter requires at least one event origin")
			}

			eventOriginMap := make(map[string]bool)
			for _, acc := range filter.EventOrigin {
				eventOriginMap[acc] = true
			}

			return &EventOriginFilter{
				EventOrigins:       eventOriginMap,
				possibleIndexSizes: possibleIndexSizes,
				indexStore:         indexStore,
			}, nil
		},
	}
}

type EventOriginFilter struct {
	EventOrigins map[string]bool

	indexStore         dstore.Store
	possibleIndexSizes []uint64
}

func (p *EventOriginFilter) String() string {
	return fmt.Sprintf("%v", p.EventOrigins)
}

func (p *EventOriginFilter) Transform(readOnlyBlk *bstream.Block, in transform.Input) (transform.Output, error) {
	block := readOnlyBlk.ToProtocol().(*pbcosmos.Block)

	block = p.filterEventsOrigins(block)

	return block, nil
}

func (p *EventOriginFilter) filterEventsOrigins(block *pbcosmos.Block) *pbcosmos.Block {

	if p.EventOrigins["BeginBlock"] == false {
		block.ResultBeginBlock.Events = []*pbcosmos.Event{}
	}
	if p.EventOrigins["EndBlock"] == false {
		block.ResultEndBlock.Events = []*pbcosmos.Event{}
	}
	if p.EventOrigins["DeliverTx"] == false {
		block.Transactions = []*pbcosmos.TxResult{}
	}

	return block
}

func (p *EventOriginFilter) GetIndexProvider() bstream.BlockIndexProvider {
	if p.indexStore == nil {
		return nil
	}

	if len(p.EventOrigins) == 0 {
		return nil
	}

	return NewEventTypeIndexProvider(
		p.indexStore,
		p.possibleIndexSizes,
		p.EventOrigins,
	)
}
