package transform

import (
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
)

const EventOriginIndexShortName = "eventorigin"

func NewEventOriginIndexProvider(
	store dstore.Store,
	possibleIndexSizes []uint64,
	eventOrigins map[string]bool,
) *transform.GenericBlockIndexProvider {
	return transform.NewGenericBlockIndexProvider(
		store,
		EventOriginIndexShortName,
		possibleIndexSizes,
		getFilterFunc(eventOrigins),
	)
}
