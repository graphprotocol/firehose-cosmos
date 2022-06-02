package transform

import (
	pbtransform "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/transform/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventTypeFilterString(t *testing.T) {
	etf := pbtransform.EventTypeFilter{} // we don't use these?

	etf = pbtransform.EventTypeFilter{}
	assert.Equal(t, "", etf.String())

	etf = pbtransform.EventTypeFilter{
		EventTypes: []string{},
	}
	assert.Equal(t, "", etf.String())

	etf = pbtransform.EventTypeFilter{
		EventTypes: []string{""},
	}
	assert.Equal(t, "event_types:\"\"", etf.String())

	etf = pbtransform.EventTypeFilter{
		EventTypes: []string{"transfer"},
	}
	assert.Equal(t, "event_types:\"transfer\"", etf.String())
}
