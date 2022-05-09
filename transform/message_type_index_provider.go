package transform

import (
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
)

const MessageTypeIndexShortName = "messagetype"

func NewMessageTypeIndexProvider(
	store dstore.Store,
	possibleIndexSizes []uint64,
	messageTypes map[string]bool,
) *transform.GenericBlockIndexProvider {
	return transform.NewGenericBlockIndexProvider(
		store,
		MessageTypeIndexShortName,
		possibleIndexSizes,
		getFilterFunc(messageTypes),
	)
}
