package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCHeadTracker(t *testing.T) {
	testStatusResp := `
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "sync_info": {
      "latest_block_hash": "03BAE86E2E0BAD1BBED598F2565AD9669DF36ED77E367F86BC0FCC51B7F178AE",
      "latest_app_hash": "94C6600C41AA5AB759A1491F7A150A15D20372DE0704D90ACFE11751D4EA0A29",
      "latest_block_height": "4493725",
      "latest_block_time": "2022-05-11T13:20:31.23374692Z",
      "earliest_block_hash": "3B6989296D0844863DC8957FF145AE7071B9970F91A608B71F20BA17A68162BD",
      "earliest_app_hash": "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
      "earliest_block_height": "3215230",
      "earliest_block_time": "2021-06-18T17:00:00Z",
      "catching_up": false
    }
  }
}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bad":
			w.WriteHeader(400)
		case "/empty":
			fmt.Fprintf(w, "{}")
		default:
			fmt.Fprintf(w, testStatusResp)
		}
	}))
	defer server.Close()

	tracker := rpcHeadTracker(server.URL + "/good")
	ref, err := tracker(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(4493725), ref.Num())
	assert.Equal(t, "03BAE86E2E0BAD1BBED598F2565AD9669DF36ED77E367F86BC0FCC51B7F178AE", ref.ID())

	tracker = rpcHeadTracker(server.URL + "/bad")
	ref, err = tracker(context.Background())
	assert.Equal(t, "endpoint returned status code 400", err.Error())
	assert.Nil(t, ref)

	tracker = rpcHeadTracker(server.URL + "/empty")
	ref, err = tracker(context.Background())
	assert.Equal(t, "latest block height is not available", err.Error())
	assert.Nil(t, ref)
}
