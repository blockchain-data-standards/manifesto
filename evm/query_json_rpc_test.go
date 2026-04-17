package evm

import (
	"encoding/json"
	"testing"
)

func TestQueryBlocksJsonRpcRoundTrip(t *testing.T) {
	req, err := QueryBlocksRequestFromJsonRpc(json.RawMessage(`{
		"fromBlock":"0x1",
		"toBlock":"0x10",
		"order":"desc",
		"limit":"0x2",
		"chainId":"0x1",
		"cursor":{"number":"0x5","hash":"0xaa","parentHash":"0xbb"},
		"fields":{"blocks":["hash","timestamp"]}
	}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := req.GetFromBlock(); got != "0x1" {
		t.Fatalf("unexpected fromBlock: %s", got)
	}
	if got := req.GetToBlock(); got != "0x10" {
		t.Fatalf("unexpected toBlock: %s", got)
	}
	if got := req.GetOrder(); got != SortOrder_DESC {
		t.Fatalf("unexpected order: %v", got)
	}
	if got := req.GetLimit(); got != 2 {
		t.Fatalf("unexpected limit: %d", got)
	}
	if req.BlockFields == nil || !req.BlockFields.Hash || !req.BlockFields.Timestamp {
		t.Fatalf("expected block field selection to be populated: %+v", req.BlockFields)
	}

	resp := QueryBlocksResponseToJsonRpc(&QueryBlocksResponse{
		Blocks: []*BlockHeader{{
			Number:    1,
			Hash:      MustHexToBytes("0xaa"),
			ParentHash: MustHexToBytes("0xbb"),
			Timestamp: 2,
		}},
		FromBlock:   &CursorBlock{Number: 1, Hash: MustHexToBytes("0xaa"), ParentHash: MustHexToBytes("0xbb")},
		ToBlock:     &CursorBlock{Number: 1, Hash: MustHexToBytes("0xaa"), ParentHash: MustHexToBytes("0xbb")},
		CursorBlock: nil,
	})
	data := resp["data"].(map[string]interface{})
	if got := len(data["blocks"].([]interface{})); got != 1 {
		t.Fatalf("unexpected block count: %d", got)
	}
	if resp["cursorBlock"] != nil {
		t.Fatalf("expected nil cursorBlock, got %#v", resp["cursorBlock"])
	}
}

func TestQueryLogsJsonRpcRequestConversion(t *testing.T) {
	req, err := QueryLogsRequestFromJsonRpc(json.RawMessage(`{
		"fromBlock":"0x1",
		"toBlock":"0x2",
		"filter":{
			"address":["0x0000000000000000000000000000000000000001"],
			"topics":[["0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"], null]
		},
		"fields":{
			"logs":["address","data"],
			"transactions":true
		}
	}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Filter == nil || len(req.Filter.Address) != 1 || len(req.Filter.Topics) != 2 {
		t.Fatalf("unexpected log filter: %+v", req.Filter)
	}
	if req.LogFields == nil || !req.LogFields.Address || !req.LogFields.Data {
		t.Fatalf("unexpected log field selection: %+v", req.LogFields)
	}
	if req.TransactionFields != nil {
		t.Fatalf("expected nil transaction field selection for true/all")
	}
}
