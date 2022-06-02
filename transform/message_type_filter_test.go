package transform

import (
	pbtransform "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/transform/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessageTypeFilterString(t *testing.T) {
	mf := pbtransform.MessageTypeFilter{} // we don't use these?

	mf = pbtransform.MessageTypeFilter{}
	assert.Equal(t, "", mf.String())

	mf = pbtransform.MessageTypeFilter{
		MessageTypes: []string{},
	}
	assert.Equal(t, "", mf.String())

	mf = pbtransform.MessageTypeFilter{
		MessageTypes: []string{""},
	}
	assert.Equal(t, "message_types:\"\"", mf.String())

	mf = pbtransform.MessageTypeFilter{
		MessageTypes: []string{"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"},
	}
	assert.Equal(t, "message_types:\"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\"", mf.String())
}
