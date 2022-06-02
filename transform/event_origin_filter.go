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

type EventOrigin string

const (
	DeliverTx  EventOrigin = "DeliverTx"
	BeginBlock EventOrigin = "BeginBlock"
	EndBlock   EventOrigin = "EndBlock"
)

var EventOriginFilterMessageName = proto.MessageName(&pbtransform.EventOriginFilter{})

func EventOriginFilterFactory(indexStore dstore.Store, possibleIndexSizes []uint64) *transform.Factory {
	return &transform.Factory{
		Obj: &pbtransform.EventOriginFilter{},
		NewFunc: func(message *anypb.Any) (transform.Transform, error) {
			if message.MessageName() != EventOriginFilterMessageName {
				return nil, fmt.Errorf("expected type url %q, received %q", EventOriginFilterMessageName, message.TypeUrl)
			}

			filter := &pbtransform.EventOriginFilter{}
			err := proto.Unmarshal(message.Value, filter)
			if err != nil {
				return nil, fmt.Errorf("unexpected unmarshal error: %w", err)
			}

			if len(filter.EventOrigins) == 0 {
				return nil, fmt.Errorf("event origin filter requires at least one event origin")
			}

			eventOriginMap := make(map[EventOrigin]bool)
			for _, acc := range filter.EventOrigins {
				eventOriginMap[EventOrigin(acc)] = true
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
	EventOrigins map[EventOrigin]bool

	indexStore         dstore.Store
	possibleIndexSizes []uint64
}

func (p *EventOriginFilter) String() string {
	return fmt.Sprintf("%v", p.EventOrigins)
}

func (p *EventOriginFilter) Transform(readOnlyBlk *bstream.Block, in transform.Input) (transform.Output, error) {
	block := readOnlyBlk.ToProtocol().(*pbcosmos.Block)

	// if filter doesn't pass these Event Origins, nullify the objects in the block
	if !p.EventOrigins[BeginBlock] {
		block.ResultBeginBlock.Events = []*pbcosmos.Event{}
	}
	if !p.EventOrigins[EndBlock] {
		block.ResultEndBlock.Events = []*pbcosmos.Event{}
	}
	if !p.EventOrigins[DeliverTx] {
		for _, tx := range block.Transactions {
			tx.Result.Events = []*pbcosmos.Event{}
		}
	}

	return block, nil
}

func (p *EventOriginFilter) GetIndexProvider() bstream.BlockIndexProvider {
	if p.indexStore == nil {
		return nil
	}

	if len(p.EventOrigins) == 0 {
		return nil
	}

	return NewEventOriginIndexProvider(
		p.indexStore,
		p.possibleIndexSizes,
		p.EventOrigins,
	)
}
