package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/streamingfast/bstream"
)

type statusResponse struct {
	Result struct {
		SyncInfo struct {
			LatestBlockHash   string `json:"latest_block_hash"`
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"sync_info"`
	} `json:"result"`
}

func rpcHeadTracker(endpoint string) bstream.BlockRefGetter {
	var lock sync.Mutex

	return func(ctx context.Context) (bstream.BlockRef, error) {
		lock.Lock()
		defer lock.Unlock()

		reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.WithContext(reqCtx)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		status := statusResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return nil, err
		}

		height, err := strconv.Atoi(status.Result.SyncInfo.LatestBlockHeight)
		if err != nil {
			return nil, err
		}

		return bstream.NewBlockRef(status.Result.SyncInfo.LatestBlockHash, uint64(height)), nil
	}
}
