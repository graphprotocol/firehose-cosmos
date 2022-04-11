package transform

import (
	"fmt"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	pbcodec "github.com/figment-networks/tendermint-protobuf-def/pb/fig/tendermint/codec/v1"
	pbtransform "github.com/figment-networks/tendermint-protobuf-def/pb/fig/tendermint/transform/v1"
)

var EventFilterMessageName = proto.MessageName(&pbtransform.EventFilter{})

func EventFilterFactory(indexStore dstore.Store, possibleIndexSizes []uint64) *transform.Factory {
	return &transform.Factory{
		Obj: &pbtransform.EventFilter{},
		NewFunc: func(message *anypb.Any) (transform.Transform, error) {
			if message.MessageName() != EventFilterMessageName {
				return nil, fmt.Errorf("expected type url %q, received %q", EventFilterMessageName, message.TypeUrl)
			}

			filter := &pbtransform.EventFilter{}
			err := proto.Unmarshal(message.Value, filter)
			if err != nil {
				return nil, fmt.Errorf("unexpected unmarshal error: %w", err)
			}

			if len(filter.EventTypes) == 0 {
				return nil, fmt.Errorf("event filter requires at least one event type")
			}

			eventTypeMap := make(map[string]bool)
			for _, acc := range filter.EventTypes {
				eventTypeMap[acc] = true
			}

			return &EventFilter{
				EventTypes:         eventTypeMap,
				possibleIndexSizes: possibleIndexSizes,
				indexStore:         indexStore,
			}, nil
		},
	}
}

type EventFilter struct {
	EventTypes map[string]bool

	indexStore         dstore.Store
	possibleIndexSizes []uint64
}

func (p *EventFilter) String() string {
	return fmt.Sprintf("%v", p.EventTypes)
}

func (p *EventFilter) Transform(readOnlyBlk *bstream.Block, in transform.Input) (transform.Output, error) {
	block := readOnlyBlk.ToProtocol().(*pbcodec.EventList)

	block.NewBlock.ResultBeginBlock.Events = p.filterEvents(block.NewBlock.ResultBeginBlock.Events)
	block.NewBlock.ResultEndBlock.Events = p.filterEvents(block.NewBlock.ResultEndBlock.Events)

	for _, tx := range block.Transaction {
		tx.TxResult.Result.Events = p.filterEvents(tx.TxResult.Result.Events)
	}

	return block, nil
}

func (p *EventFilter) filterEvents(events []*pbcodec.Event) []*pbcodec.Event {
	var outEvents []*pbcodec.Event

	for _, event := range events {
		if p.EventTypes[event.EventType] {
			outEvents = append(outEvents, event)
		}
	}

	return outEvents
}

func (p *EventFilter) GetIndexProvider() bstream.BlockIndexProvider {
	if p.indexStore == nil {
		return nil
	}

	if len(p.EventTypes) == 0 {
		return nil
	}

	return NewEventTypeIndexProvider(
		p.indexStore,
		p.possibleIndexSizes,
		p.EventTypes,
	)
}
