package main

import (
	"context"

	"github.com/streamingfast/bstream"
)

func staticHeadTracker(height uint64, hash string) bstream.BlockRefGetter {
	return func(ctx context.Context) (bstream.BlockRef, error) {
		return bstream.NewBlockRef(hash, height), nil
	}
}
