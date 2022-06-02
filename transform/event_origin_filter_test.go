package transform

import (
	pbtransform "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/transform/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventOriginFilterString(t *testing.T) {
	eof := pbtransform.EventOriginFilter{} // we don't use these?

	eof = pbtransform.EventOriginFilter{}
	assert.Equal(t, "", eof.String())

	eof = pbtransform.EventOriginFilter{
		EventOrigins: []string{},
	}
	assert.Equal(t, "", eof.String())

	eof = pbtransform.EventOriginFilter{
		EventOrigins: []string{""},
	}
	assert.Equal(t, "event_origins:\"\"", eof.String())

	eof = pbtransform.EventOriginFilter{
		EventOrigins: []string{"DeliverTx"},
	}
	assert.Equal(t, "event_origins:\"DeliverTx\"", eof.String())
}
