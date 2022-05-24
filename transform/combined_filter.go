package transform

import (
	"fmt"
	"strings"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	pbtransform "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/transform/v1"
	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
)

var CombinedFilterMessageName = proto.MessageName(&pbtransform.CombinedFilter{})

func CombinedFilterFactory(indexStore dstore.Store, possibleIndexSizes []uint64) *transform.Factory {
	return &transform.Factory{
		Obj: &pbtransform.CombinedFilter{},
		NewFunc: func(message *anypb.Any) (transform.Transform, error) {
			if message.MessageName() != CombinedFilterMessageName {
				return nil, fmt.Errorf("expected type url %q, received %q", CombinedFilterMessageName, message.TypeUrl)
			}

			filter := &pbtransform.CombinedFilter{}
			err := proto.Unmarshal(message.Value, filter)
			if err != nil {
				return nil, fmt.Errorf("unexpected unmarshal error: %w", err)
			}

			// TODO: This needs to check any returned type
			if len(filter.EventOriginFilters) == 0 && len(filter.EventOriginFilters) == 0 && len(filter.MessageTypeFilters) == 0 {
				return nil, fmt.Errorf("a combined filter requires at least one Event Type, Event Origin or Message Type filter")
			}

			//combinedFilterMap := make(map[string]bool)
			//for _, acc := range filter.CombinedFilters {
			//	combinedFilterMap[acc] = true
			//}

			//return &CombinedFilter{
			//	CombinedFilters:    combinedFilterMap,
			//	possibleIndexSizes: possibleIndexSizes,
			//	indexStore:         indexStore,
			//}, nil

			return newCombinedFilter(filter.EventTypeFilters, filter.EventOriginFilters, filter.MessageTypeFilters indexStore, possibleIndexSizes)
		},
	}
}

func newCombinedFilter(pbEventTypeFilter []*pbtransform.EventTypeFilter, pbEventOriginFilter []*pbtransform.EventOriginFilter, pbMessageTypeFilter []*pbtransform.MessageTypeFilter, indexStore dstore.Store, possibleIndexSizes []uint64) (*CombinedFilter, error) {
	var eventTypeFilters []*EventTypeFilter
	if l := len(pbEventTypeFilter); l > 0 {
		eventTypeFilters = make([]*EventTypeFilter, l)
		for i, in := range pbEventTypeFilter {
			// TODO: this should call a new filter check in event_type_filter.go or similar
			f, err := NewEventTypeFilter(in)
			if err != nil {
				return nil, err
			}
			eventTypeFilters[i] = f
		}
	}

	var eventOriginFilters []*EventOriginFilter
	if l := len(pbEventOriginFilter); l > 0 {
		eventOriginFilters = make([]*EventOriginFilter, l)
		for i, in := range pbEventOriginFilter {
			f, err := NewEventOriginFilter(in)
			if err != nil {
				return nil, err
			}
			eventOriginFilters[i] = f
		}
	}

	var messageTypeFilters []*MessageTypeFilter
	if l := len(pbMessageTypeFilter); l > 0 {
		messageTypeFilters = make([]*MessageTypeFilter, l)
		for i, in := range pbMessageTypeFilter {
			f, err := NewMessageTypeFilter(in)
			if err != nil {
				return nil, err
			}
			messageTypeFilters[i] = f
		}
	}

	f := &CombinedFilter{
		EventTypeFilters:    eventTypeFilters,
		EventOriginFilters:  eventOriginFilters,
		MessageTypeFilters: 	messageTypeFilters,
		indexStore:         indexStore,
		possibleIndexSizes: possibleIndexSizes,
	}

	return f, nil
}


type CombinedFilter struct {
	EventTypeFilters 	[]*EventTypeFilter
	EventOriginFilters  []*EventOriginFilter
	MessageTypeFilters 	[]*MessageTypeFilter

	indexStore         dstore.Store
	possibleIndexSizes []uint64
}

func (f *CombinedFilter) String() string {
	eventTypeFilters := make([]string, len(f.EventTypeFilters))
	for i, f := range f.EventTypeFilters {
		// TODO: wat
		eventTypeFilters[i] = addSigString(f)
	}
	eventOriginFilters := make([]string, len(f.EventOriginFilters))
	for i, f := range f.EventOriginFilters {
		// TODO: wat
		eventOriginFilters[i] = addSigString(f)
	}
	messageTypeFilters := make([]string, len(f.MessageTypeFilters))
	for i, f := range f.MessageTypeFilters {
		// TODO: wat
		messageTypeFilters[i] = addSigString(f)
	}

	return fmt.Sprintf("Combined filter: Calls:[%s], Logs:[%s]", strings.Join(eventTypeFilters, ","), strings.Join(eventOriginFilters, ","), strings.Join(messageTypeFilters, ","))
}

func (p *CombinedFilter) Transform(readOnlyBlk *bstream.Block, in transform.Input) (transform.Output, error) {
	block := readOnlyBlk.ToProtocol().(*pbcosmos.Block)


	// TODO: check this as its a match in the eth code
	for _, tx := range block.Transactions {
		tx.Tx.Body.Messages = p.filterCombinedFilters(tx.Tx.Body.Messages)
	}

	return block, nil
}

//func (p *CombinedFilter) filterCombinedFilters(messages []*anypb.Any) []*anypb.Any {
//	var outMessages []*anypb.Any
//
//	for _, message := range messages {
//		if p.MessageTypes[message.TypeUrl] {
//			outMessages = append(outMessages, message)
//		}
//	}
//
//	return outMessages
//}

func (p *CombinedFilter) GetIndexProvider() bstream.BlockIndexProvider {
	if p.indexStore == nil {
		return nil
	}

	if len(p.MessageTypes) == 0 {
		return nil
	}

	return NewCombinedFilterIndexProvider(
		p.indexStore,
		p.possibleIndexSizes,
		p.MessageTypes,
	)
}
