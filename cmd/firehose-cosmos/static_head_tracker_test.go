package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticHeadTracker(t *testing.T) {
	tracker := staticHeadTracker(12345, "HASH")
	ref, err := tracker(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(12345), ref.Num())
	assert.Equal(t, "HASH", ref.ID())
}
