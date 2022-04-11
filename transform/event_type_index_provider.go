package transform

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
)

const EventTypeIndexShortName = "eventtype"

func NewEventTypeIndexProvider(
	store dstore.Store,
	possibleIndexSizes []uint64,
	eventTypes map[string]bool,
) *transform.GenericBlockIndexProvider {
	return transform.NewGenericBlockIndexProvider(
		store,
		EventTypeIndexShortName,
		possibleIndexSizes,
		getFilterFunc(eventTypes),
	)
}

func getFilterFunc(eventTypes map[string]bool) func(transform.BitmapGetter) []uint64 {
	return func(getBitmap transform.BitmapGetter) (matchingBlocks []uint64) {
		out := roaring64.NewBitmap()
		for et := range eventTypes {
			if bm := getBitmap(et); bm != nil {
				out.Or(bm)
			}
		}
		return nilIfEmpty(out.ToArray())
	}
}

func nilIfEmpty(in []uint64) []uint64 {
	if len(in) == 0 {
		return nil
	}
	return in
}
