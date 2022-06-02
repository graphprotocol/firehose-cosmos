package transform

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
)

const EventOriginIndexShortName = "eventorigin"

func NewEventOriginIndexProvider(
	store dstore.Store,
	possibleIndexSizes []uint64,
	eventOrigins map[EventOrigin]bool,
) *transform.GenericBlockIndexProvider {
	return transform.NewGenericBlockIndexProvider(
		store,
		EventOriginIndexShortName,
		possibleIndexSizes,
		getEventOriginFilterFunc(eventOrigins),
	)
}

func getEventOriginFilterFunc(eventOrigins map[EventOrigin]bool) func(transform.BitmapGetter) []uint64 {
	return func(getBitmap transform.BitmapGetter) (matchingBlocks []uint64) {
		out := roaring64.NewBitmap()
		for et := range eventOrigins {
			if bm := getBitmap(string(et)); bm != nil {
				out.Or(bm)
			}
		}
		return nilIfEmpty(out.ToArray())
	}
}
