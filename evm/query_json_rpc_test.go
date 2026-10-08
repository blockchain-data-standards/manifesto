package evm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/blockchain-data-standards/manifesto/common"
)

// TestQueryJsonRpcMip16 walks eth_query* requests through the JSON-RPC codec
// as a server uses it: parse the params, let a small in-memory server build
// the proto response, render it, and encode it as JSON. It covers each
// -32602 request class of MIP-16, the server-side error codes through
// QueryErrorCode and ValidateQueryRange, one success per method, Monad and
// generic schemas, relation joins, desc order, paging by target and budget,
// and checks that every rendered object carries exactly the selected keys.
func TestQueryJsonRpcMip16(t *testing.T) {
	const monad = 143
	server := newFakeQueryServer()
	a, b := hexOf(0xaa, 20), hexOf(0xbb, 20)
	t1 := hexOf(0x71, 32)

	cases := []struct {
		name    string
		method  string
		chainID uint64
		budget  int // primary objects per page; 0 is unlimited
		params  string
		code    int // expected JSON-RPC error code; 0 is success
		check   func(t *testing.T, result map[string]interface{})
	}{
		// Request envelope (MIP-16 Request).
		{name: "params is an object", method: "eth_queryLogs", params: `{"fromBlock":"0x10"}`, code: -32602},
		{name: "params is empty", method: "eth_queryLogs", params: `[]`, code: -32602},
		{name: "params has two elements", method: "eth_queryLogs", params: `[{},{}]`, code: -32602},
		{name: "params element is not an object", method: "eth_queryLogs", params: `["0x10"]`, code: -32602},
		{name: "unknown request key", method: "eth_queryBlocks", params: `[{"notAField":1}]`, code: -32602},
		{name: "old limit key", method: "eth_queryTransactions", params: `[{"limit":"0x1"}]`, code: -32602},
		{name: "old cursor key", method: "eth_queryTraces", params: `[{"cursor":{"number":"0x10"}}]`, code: -32602},
		{name: "chainId is not a JSON key", method: "eth_queryTransfers", params: `[{"chainId":"0x8f"}]`, code: -32602},
		{name: "pending tag", method: "eth_queryLogs", params: `[{"toBlock":"pending"}]`, code: -32602},
		{name: "block number with a leading zero", method: "eth_queryLogs", params: `[{"fromBlock":"0x010"}]`, code: -32602},
		{name: "block number as JSON number", method: "eth_queryLogs", params: `[{"fromBlock":16}]`, code: -32602},
		{name: "order ascending", method: "eth_queryBlocks", params: `[{"order":"ascending"}]`, code: -32602},
		{name: "order uppercase", method: "eth_queryBlocks", params: `[{"order":"DESC"}]`, code: -32602},
		{name: "target 0x0", method: "eth_queryTransactions", params: `[{"target":"0x0"}]`, code: -32602},
		{name: "target as JSON number", method: "eth_queryLogs", params: `[{"target":5}]`, code: -32602},
		{name: "target malformed", method: "eth_queryLogs", params: `[{"target":"0xg"}]`, code: -32602},

		// Filters (MIP-16 Filters and each method's Filter).
		{name: "filter is not an object", method: "eth_queryLogs", params: `[{"filter":[]}]`, code: -32602},
		{name: "unknown filter key", method: "eth_queryBlocks", params: `[{"filter":{"from":"` + a + `"}}]`, code: -32602},
		{name: "transfers have no selector filter", method: "eth_queryTransfers", params: `[{"filter":{"selector":"0x12345678"}}]`, code: -32602},
		{name: "miner of 19 bytes", method: "eth_queryBlocks", params: `[{"filter":{"miner":"` + hexOf(0xc0, 19) + `"}}]`, code: -32602},
		{name: "selector of 3 bytes", method: "eth_queryTransactions", params: `[{"filter":{"selector":["0x123456"]}}]`, code: -32602},
		{name: "address that is not hex", method: "eth_queryLogs", params: `[{"filter":{"address":"0x` + strings.Repeat("zz", 20) + `"}}]`, code: -32602},
		{name: "null inside a filter array", method: "eth_queryTransactions", params: `[{"filter":{"to":["` + a + `",null]}}]`, code: -32602},
		{name: "five topic positions", method: "eth_queryLogs", params: `[{"filter":{"topics":[null,null,null,null,null]}}]`, code: -32602},
		{name: "topic of 31 bytes", method: "eth_queryLogs", params: `[{"filter":{"topics":["` + hexOf(0x71, 31) + `"]}}]`, code: -32602},
		{name: "includeReverted as a string", method: "eth_queryTraces", params: `[{"filter":{"includeReverted":"true"}}]`, code: -32602},
		{name: "includeReverted as an array", method: "eth_queryTransfers", params: `[{"filter":{"includeReverted":[true]}}]`, code: -32602},
		{name: "old isTopLevel filter", method: "eth_queryTraces", params: `[{"filter":{"isTopLevel":true}}]`, code: -32602},

		// Fields (MIP-16 Fields and relations).
		{name: "fields without the primary key", method: "eth_queryLogs", params: `[{"fields":{"blocks":["number"]}}]`, code: -32602},
		{name: "unknown fields key", method: "eth_queryLogs", params: `[{"fields":{"logs":"all","receipts":"all"}}]`, code: -32602},
		{name: "blocks join no relation", method: "eth_queryBlocks", params: `[{"fields":{"blocks":"all","transactions":"all"}}]`, code: -32602},
		{name: "one-to-many relation", method: "eth_queryTransactions", params: `[{"fields":{"transactions":"all","logs":"all"}}]`, code: -32602},
		{name: "fields value some", method: "eth_queryTraces", params: `[{"fields":{"traces":"some"}}]`, code: -32602},
		{name: "fields value true", method: "eth_queryTraces", params: `[{"fields":{"traces":true}}]`, code: -32602},
		{name: "empty primary fields", method: "eth_queryTransfers", params: `[{"fields":{"transfers":[]}}]`, code: -32602},
		{name: "empty relation fields", method: "eth_queryLogs", params: `[{"fields":{"logs":["data"],"blocks":[]}}]`, code: -32602},
		{name: "unknown field name", method: "eth_queryLogs", params: `[{"fields":{"logs":["data","notAField"]}}]`, code: -32602},
		{name: "field of another schema", method: "eth_queryBlocks", params: `[{"fields":{"blocks":["transactionIndex"]}}]`, code: -32602},
		{name: "relation field of another schema", method: "eth_queryTraces", params: `[{"fields":{"traces":["type"],"transactions":["traceAddress"]}}]`, code: -32602},
		{name: "blocks never carry transactions", method: "eth_queryBlocks", params: `[{"fields":{"blocks":["transactions"]}}]`, code: -32602},
		{name: "old transactionCount block field", method: "eth_queryBlocks", params: `[{"fields":{"blocks":["transactionCount"]}}]`, code: -32602},
		{name: "old traceType trace field", method: "eth_queryTraces", params: `[{"fields":{"traces":["traceType"]}}]`, code: -32602},
		{name: "Monad has no blob fields", method: "eth_queryTransactions", chainID: monad, params: `[{"fields":{"transactions":["hash","maxFeePerBlobGas"]}}]`, code: -32602},

		// Server-resolved errors: range, availability, method support, budget.
		{name: "inverted asc range", method: "eth_queryBlocks", params: `[{"fromBlock":"0x11","toBlock":"0x10"}]`, code: -32602},
		{name: "inverted desc range", method: "eth_queryLogs", params: `[{"fromBlock":"0x10","toBlock":"latest","order":"desc"}]`, code: -32602},
		{name: "block that does not exist yet", method: "eth_queryTransactions", params: `[{"fromBlock":"0x10","toBlock":"0xffffffff"}]`, code: -32001},
		{name: "omitted fromBlock is earliest, which is pruned", method: "eth_queryBlocks", params: `[{"toBlock":"0x11"}]`, code: -32001},
		{name: "traces not served on this chain", method: "eth_queryTraces", chainID: 1, params: `[{"fromBlock":"0x10","toBlock":"0x10"}]`, code: -32004},
		{name: "budget exhausted before fromBlock completes", method: "eth_queryTransactions", budget: 2, params: `[{"fromBlock":"0x11","toBlock":"0x10","order":"desc"}]`, code: -32005},

		// eth_queryBlocks
		{
			name: "blocks: explicit fields, target ends the page", method: "eth_queryBlocks", chainID: monad,
			params: `[{"fromBlock":"0x10","toBlock":"0x11","target":"0x1","fields":{"blocks":["number","hash","miner"]}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "blocks", "number", "0x10")
				expectRef(t, r, "toBlock", "0x11")
				expectRef(t, r, "cursorBlock", "0x10")
			},
		},
		{
			name: "blocks: desc from latest, all fields, requestsHash only after the fork", method: "eth_queryBlocks", chainID: monad,
			params: `[{"fromBlock":"latest","toBlock":"0x10","order":"desc","fields":{"blocks":"all"}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "blocks", "number", "0x11", "0x10")
				expectValues(t, r, "blocks", "requestsHash", hexOf(0xe7, 32), nil)
				expectValues(t, r, "blocks", "nonce", "0x0000000000000000", "0x0000000000000000")
				expectRef(t, r, "fromBlock", "0x11")
				expectRef(t, r, "cursorBlock", "0x10")
			},
		},
		{
			name: "blocks: miner filter, fields omitted", method: "eth_queryBlocks", chainID: monad,
			params: `[{"fromBlock":"0x10","toBlock":"0x11","filter":{"miner":"` + hexOf(0xc1, 20) + `"}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "blocks", "number", "0x11")
			},
		},

		// eth_queryTransactions
		{
			name: "transactions: Monad all, type-dependent fields, blocks join", method: "eth_queryTransactions", chainID: monad,
			params: `[{"fromBlock":"0x11","toBlock":"0x11","fields":{"transactions":"all","blocks":["number","timestamp"]}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "transactions", "type", "0x4", "0x0", "0x3")
				expectValues(t, r, "transactions", "chainId", "0x8f", "0x8f", "0x8f")
				expectValues(t, r, "transactions", "yParity", "0x0", nil, "0x1")
				expectValues(t, r, "transactions", "maxFeePerGas", "0x64", nil, "0x64")
				expectValues(t, r, "transactions", "accessList", []interface{}{}, nil, []interface{}{})
				expectValues(t, r, "transactions", "maxFeePerBlobGas", nil, nil, nil)
				expectValues(t, r, "transactions", "to", b, nil, b)
				expectValues(t, r, "transactions", "contractAddress", nil, hexOf(0xcc, 20), nil)
				expectValues(t, r, "transactions", "status", "0x1", "0x1", "0x0")
				expectValues(t, r, "transactions", "cumulativeGasUsed", "0x5208", "0xa410", "0xf618")
				auth := objects(t, r, "transactions")[0]["authorizationList"].([]interface{})
				expectKeys(t, auth[0].(map[string]interface{}), "chainId", "address", "nonce", "yParity", "r", "s")
				expectValues(t, r, "blocks", "number", "0x11")
				expectValues(t, r, "blocks", "timestamp", "0x3e9")
			},
		},
		{
			name: "transactions: generic chain all keeps blob fields", method: "eth_queryTransactions", chainID: 1,
			params: `[{"fromBlock":"0x11","toBlock":"0x11","filter":{"to":["` + b + `"]},"fields":{"transactions":"all"}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "transactions", "type", "0x4", "0x3")
				expectValues(t, r, "transactions", "maxFeePerBlobGas", nil, "0x1")
				expectValues(t, r, "transactions", "blobVersionedHashes", nil, []interface{}{hexOf(0x01, 32)})
			},
		},
		{
			name: "transactions: budget ends the page at the previous block", method: "eth_queryTransactions", chainID: monad, budget: 4,
			params: `[{"fromBlock":"0x10","toBlock":"0x11","fields":{"transactions":["hash","type","chainId","yParity","status","cumulativeGasUsed","logsBloom"]}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectRef(t, r, "cursorBlock", "0x10")
				expectValues(t, r, "transactions", "type", "0x0", "0x2")
				// An unprotected legacy transaction has neither chainId nor yParity;
				// its receipt fields come from MergeReceipt.
				expectKeys(t, objects(t, r, "transactions")[0], "hash", "type", "status", "cumulativeGasUsed", "logsBloom")
				expectValues(t, r, "transactions", "chainId", nil, "0x8f")
				expectValues(t, r, "transactions", "status", "0x1", "0x1")
				expectValues(t, r, "transactions", "logsBloom", hexOf(0, 256), hexOf(0, 256))
			},
		},

		// eth_queryLogs
		{
			name: "logs: desc, address and topic filter, transactions join", method: "eth_queryLogs", chainID: monad,
			params: `[{"fromBlock":"0x11","toBlock":"0x10","order":"desc","filter":{"address":"` + a + `","topics":["` + t1 + `"]},"fields":{"logs":["blockNumber","logIndex","topics"],"transactions":["hash","status"]}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "logs", "blockNumber", "0x11", "0x10")
				expectValues(t, r, "logs", "logIndex", "0x1", "0x0")
				expectValues(t, r, "logs", "topics", []interface{}{t1, hexOf(0x99, 32)}, []interface{}{t1})
				expectValues(t, r, "transactions", "hash", txHashHex(0x11, 0), txHashHex(0x10, 1))
			},
		},
		{
			name: "logs: omitted fields select every log field and no relation", method: "eth_queryLogs", chainID: monad,
			params: `[{"fromBlock":"0x10","toBlock":"0x11","filter":{"address":null,"topics":[]}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "logs", "removed", false, false, false)
				expectValues(t, r, "logs", "blockTimestamp", "0x3e8", "0x3e9", "0x3e9")
			},
		},

		// eth_queryTraces
		{
			name: "traces: explicit fields, reverted frames excluded by default", method: "eth_queryTraces", chainID: monad,
			params: `[{"fromBlock":"0x10","toBlock":"0x11","fields":{"traces":["type","from","to","value","error","reverted","traceAddress"]}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				// The Parity child [1,0] of the failed call [1] is reverted too, so it is excluded.
				expectValues(t, r, "traces", "type", "CALL", "DELEGATECALL", "SELFDESTRUCT", "CALL", "STATICCALL")
				expectValues(t, r, "traces", "traceAddress", []interface{}{}, []interface{}{float64(0)}, []interface{}{float64(2)}, []interface{}{}, []interface{}{float64(0)})
				expectValues(t, r, "traces", "error", nil, nil, nil, nil, nil)
				expectValues(t, r, "traces", "reverted", false, false, false, false, false)
				expectValues(t, r, "traces", "value", "0x5", "0x5", "0x7", "0x0", "0x0")
				// A Parity suicide frame: from is the destroyed contract, to the refund address.
				expectValues(t, r, "traces", "from", hexOf(0x01, 20), a, a, hexOf(0x01, 20), b)
				expectValues(t, r, "traces", "to", a, b, hexOf(0xdd, 20), b, a)
			},
		},
		{
			name: "traces: includeReverted, all fields, blocks join", method: "eth_queryTraces", chainID: monad,
			params: `[{"fromBlock":"0x10","toBlock":"0x11","filter":{"includeReverted":true,"to":null},"fields":{"traces":"all","blocks":["number"]}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "traces", "type", "CALL", "DELEGATECALL", "CALL", "CALL", "SELFDESTRUCT", "CALL", "STATICCALL", "CREATE")
				expectValues(t, r, "traces", "error", nil, nil, "execution reverted", nil, nil, nil, nil, "out of gas")
				expectValues(t, r, "traces", "reverted", false, false, true, true, false, false, false, true)
				expectValues(t, r, "traces", "to", a, b, b, a, hexOf(0xdd, 20), b, a, nil)
				// callTracer frames carry the txHash of their debug_traceBlockByNumber item and the given index.
				expectValues(t, r, "traces", "transactionHash", txHashHex(0x10, 1), txHashHex(0x10, 1), txHashHex(0x10, 1),
					txHashHex(0x10, 1), txHashHex(0x10, 1), txHashHex(0x11, 0), txHashHex(0x11, 0), txHashHex(0x11, 1))
				expectValues(t, r, "traces", "transactionIndex", "0x1", "0x1", "0x1", "0x1", "0x1", "0x0", "0x0", "0x1")
				expectValues(t, r, "blocks", "number", "0x10", "0x11")
			},
		},

		// eth_queryTransfers
		{
			name: "transfers: no DELEGATECALL or zero value, transactions join", method: "eth_queryTransfers", chainID: monad,
			params: `[{"fromBlock":"0x10","toBlock":"0x11","filter":{"includeReverted":true},"fields":{"transfers":"all","transactions":["hash"]}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "transfers", "type", "CALL", "CALL", "CALL", "SELFDESTRUCT", "CREATE")
				expectValues(t, r, "transfers", "value", "0x5", "0x1", "0x3", "0x7", "0x2")
				expectValues(t, r, "transfers", "traceAddress", []interface{}{}, []interface{}{float64(1)},
					[]interface{}{float64(1), float64(0)}, []interface{}{float64(2)}, []interface{}{})
				// NativeTransfersFromTraces saw the Parity frames before PropagateParityReverted ran.
				expectValues(t, r, "transfers", "reverted", false, true, true, false, true)
				expectValues(t, r, "transfers", "gasUsed", "0x5208", "0x0", "0x5208", "0x0", "0x5208")
				expectValues(t, r, "transactions", "hash", txHashHex(0x10, 1), txHashHex(0x11, 1))
			},
		},
		{
			name: "transfers: reverted frames excluded by default", method: "eth_queryTransfers", chainID: monad,
			params: `[{"fromBlock":"0x10","toBlock":"0x11","fields":{"transfers":["value","to","reverted"]}}]`,
			check: func(t *testing.T, r map[string]interface{}) {
				expectValues(t, r, "transfers", "value", "0x5", "0x7")
				expectValues(t, r, "transfers", "to", a, hexOf(0xdd, 20))
				expectValues(t, r, "transfers", "reverted", false, false)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := server.run(c.method, c.chainID, c.budget, json.RawMessage(c.params))
			if c.code != 0 {
				if got := QueryErrorCode(err); got != c.code {
					t.Fatalf("error code %d, want %d (error: %v)", got, c.code, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expectSelectedFields(t, c.method, c.chainID, c.params, result)
			c.check(t, result)
		})
	}
}

// conditionalQueryFields are the fields MIP-16 Appendix 1 marks type- or
// fork-dependent: selected, they may still be absent from an object.
var conditionalQueryFields = map[string]bool{
	"maxFeePerGas": true, "maxPriorityFeePerGas": true, "chainId": true, "accessList": true,
	"authorizationList": true, "yParity": true, "requestsHash": true,
	"maxFeePerBlobGas": true, "blobVersionedHashes": true,
}

// expectSelectedFields checks that data has exactly the primary key and the
// selected relations, and that every object has exactly the selected fields,
// less the conditional fields that do not apply to it.
func expectSelectedFields(t *testing.T, method string, chainID uint64, params string, result map[string]interface{}) {
	t.Helper()
	primary := strings.ToLower(strings.TrimPrefix(method, "eth_query")[:1]) + strings.TrimPrefix(method, "eth_query")[1:]
	var p []struct {
		Fields map[string]json.RawMessage `json:"fields"`
	}
	if err := json.Unmarshal([]byte(params), &p); err != nil {
		t.Fatal(err)
	}
	fields := p[0].Fields
	if fields == nil {
		fields = map[string]json.RawMessage{primary: json.RawMessage(`"all"`)}
	}
	schema := QuerySchemaForChain(chainID)
	schemaOf := map[string][]string{
		"blocks": schema.Blocks, "transactions": schema.Transactions, "logs": schema.Logs,
		"traces": schema.Traces, "transfers": schema.Transfers,
	}
	data := result["data"].(map[string]interface{})
	if len(data) != len(fields) {
		t.Fatalf("data has keys %v, want the keys of %v", keysOf(data), keysOf(fields))
	}
	for key, raw := range fields {
		names := schemaOf[key]
		if string(raw) != `"all"` {
			names = nil
			if err := json.Unmarshal(raw, &names); err != nil {
				t.Fatal(err)
			}
		}
		selected := make(map[string]bool, len(names))
		for _, n := range names {
			selected[n] = true
		}
		rows := objects(t, result, key)
		if len(rows) == 0 {
			t.Fatalf("data.%s is empty", key)
		}
		for i, row := range rows {
			for k := range row {
				if !selected[k] {
					t.Errorf("data.%s[%d] has %q, which the request does not select", key, i, k)
				}
			}
			for _, n := range names {
				if _, ok := row[n]; !ok && !conditionalQueryFields[n] {
					t.Errorf("data.%s[%d] lacks the selected field %q", key, i, n)
				}
			}
		}
	}
	for _, ref := range []string{"fromBlock", "toBlock", "cursorBlock"} {
		expectKeys(t, result[ref].(map[string]interface{}), "number", "hash", "parentHash")
	}
}

func objects(t *testing.T, result map[string]interface{}, key string) []map[string]interface{} {
	t.Helper()
	rows, ok := result["data"].(map[string]interface{})[key].([]interface{})
	if !ok {
		t.Fatalf("data.%s is not an array", key)
	}
	out := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		out[i] = row.(map[string]interface{})
	}
	return out
}

// expectValues checks field of each data[key] object in order; nil expects
// the field to be absent or null.
func expectValues(t *testing.T, result map[string]interface{}, key, field string, want ...interface{}) {
	t.Helper()
	rows := objects(t, result, key)
	got := make([]interface{}, len(rows))
	for i, row := range rows {
		got[i] = row[field]
	}
	if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
		t.Errorf("data.%s[].%s = %v, want %v", key, field, got, want)
	}
}

func expectRef(t *testing.T, result map[string]interface{}, ref, number string) {
	t.Helper()
	if got := result[ref].(map[string]interface{})["number"]; got != number {
		t.Errorf("%s.number = %v, want %s", ref, got, number)
	}
}

func expectKeys(t *testing.T, obj map[string]interface{}, want ...string) {
	t.Helper()
	got := keysOf(obj)
	if len(got) != len(want) {
		t.Errorf("keys %v, want %v", got, want)
		return
	}
	for _, k := range want {
		if _, ok := obj[k]; !ok {
			t.Errorf("keys %v, want %v", got, want)
			return
		}
	}
}

func keysOf[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func hexOf(b byte, n int) string {
	return BytesToHex(bytes.Repeat([]byte{b}, n))
}

func txHash(block uint64, index uint32) []byte {
	h := bytes.Repeat([]byte{byte(block)}, 32)
	h[31] = byte(index)
	return h
}

func txHashHex(block uint64, index uint32) string { return BytesToHex(txHash(block, index)) }

// fakeQueryServer stands in for erpc and prism: it builds its records with the
// public conversion helpers (ParseJsonRpcTransaction, MergeReceipt,
// TraceFromParity, PropagateParityReverted, TraceFromGethDebug,
// NativeTransfersFromTraces), resolves the range, checks it with
// ValidateQueryRange, applies its availability window (blocks 0x10 and 0x11),
// filters, pages block by block, joins relations, and returns BDS errors that
// QueryErrorCode maps to JSON-RPC codes.
type fakeQueryServer struct {
	blocks    []*BlockHeader // ascending, from block 0x10
	txs       []*Transaction
	logs      []*Log
	traces    []*Trace
	transfers []*NativeTransfer
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func newFakeQueryServer() *fakeQueryServer {
	addr := func(b byte) []byte { return bytes.Repeat([]byte{b}, 20) }
	hash := func(b byte) []byte { return bytes.Repeat([]byte{b}, 32) }
	block := func(n uint64, miner byte) *BlockHeader {
		return &BlockHeader{
			Number: n, Timestamp: 1000 + n - 0x10, GasLimit: 200_000_000, GasUsed: 105_000, Size: 1234,
			Hash: hash(byte(n)), ParentHash: hash(byte(n - 1)), StateRoot: hash(0x51), TransactionsRoot: hash(0x52),
			ReceiptsRoot: hash(0x53), Sha3Uncles: hash(0x54), Miner: addr(miner), LogsBloom: make([]byte, 256),
			ExtraData: []byte{}, Nonce: Uint64Ptr(0), BlobGasUsed: Uint64Ptr(0), ExcessBlobGas: Uint64Ptr(0),
			MixHash: hash(0x55), ParentBeaconBlockRoot: hash(0x56), WithdrawalsRoot: hash(0x57),
			BaseFeePerGas: StringPtr("100"), Difficulty: StringPtr("0"), TotalDifficulty: StringPtr("0"),
		}
	}
	b16, b17 := block(0x10, 0xc0), block(0x11, 0xc1)
	b17.RequestsHash = hash(0xe7)

	tx := func(blk *BlockHeader, index uint32, typ uint32, to []byte) *Transaction {
		return &Transaction{
			Hash: txHash(blk.Number, index), Nonce: uint64(index), From: addr(0x01), To: to, Value: "1000",
			Input: []byte{0x12, 0x34, 0x56, 0x78}, Type: typ, GasLimit: 50_000, GasPrice: StringPtr("100"),
			GasUsed: Uint64Ptr(21_000), CumulativeGasUsed: Uint64Ptr(21_000 * uint64(index+1)),
			EffectiveGasPrice: StringPtr("100"), Status: Uint32Ptr(1), LogsBloom: make([]byte, 256),
			R: hash(0x0a), S: hash(0x0b), BlockNumber: Uint64Ptr(blk.Number), BlockHash: blk.Hash,
			TransactionIndex: Uint32Ptr(index), BlockTimestamp: Uint64Ptr(blk.Timestamp),
		}
	}
	typed := func(t *Transaction, yParity uint32) *Transaction {
		t.ChainId, t.YParity, t.V = Uint64Ptr(143), Uint32Ptr(yParity), []byte{byte(yParity)}
		t.MaxFeePerGas, t.MaxPriorityFeePerGas = StringPtr("100"), StringPtr("1")
		return t
	}
	// A pre-EIP-155 legacy transaction as a node sends it, merged with its
	// receipt as a server does before it renders MIP-16 transaction objects.
	legacy := must(ParseJsonRpcTransaction(map[string]interface{}{
		"hash": txHashHex(0x10, 0), "nonce": "0x0", "from": BytesToHex(addr(0x01)), "to": BytesToHex(addr(0xaa)),
		"value": "0x3e8", "input": "0x12345678", "type": "0x0", "gas": "0xc350", "gasPrice": "0x64",
		"r": BytesToHex(hash(0x0a)), "s": BytesToHex(hash(0x0b)), "v": "0x1b",
		"blockNumber": "0x10", "blockHash": BytesToHex(b16.Hash), "transactionIndex": "0x0",
	}, b16))
	MergeReceipt(legacy, must((&JsonRpcReceipt{
		BlockHash: BytesToHex(b16.Hash), BlockNumber: "0x10", TransactionHash: txHashHex(0x10, 0), TransactionIndex: "0x0",
		From: BytesToHex(addr(0x01)), To: BytesToHex(addr(0xaa)), Type: "0x0", Status: "0x1", GasUsed: "0x5208",
		CumulativeGasUsed: "0x5208", EffectiveGasPrice: "0x64", LogsBloom: BytesToHex(make([]byte, 256)),
	}).ToProto()))
	dynamic := typed(tx(b16, 1, 2, addr(0xaa)), 1)
	dynamic.AccessList = []*AccessListItem{{Address: addr(0xaa), StorageKeys: [][]byte{hash(0x01)}}}
	setCode := typed(tx(b17, 0, 4, addr(0xbb)), 0)
	setCode.AuthorizationList = []*AuthorizationListItem{{ChainId: 143, Address: addr(0xdd), Nonce: 7, R: hash(0x0c), S: hash(0x0d), YParity: 1, Authority: addr(0xee)}}
	// A contract creation as one merged transaction+receipt object.
	creation := must(ParseJsonRpcTransaction(map[string]interface{}{
		"hash": txHashHex(0x11, 1), "nonce": "0x1", "from": BytesToHex(addr(0x01)), "to": nil,
		"value": "0x3e8", "input": "0x12345678", "type": "0x0", "gas": "0xc350", "gasPrice": "0x64",
		"r": BytesToHex(hash(0x0a)), "s": BytesToHex(hash(0x0b)), "v": "0x141", "chainId": "0x8f",
		"blockNumber": "0x11", "blockHash": BytesToHex(b17.Hash), "transactionIndex": "0x1",
		"status": "0x1", "gasUsed": "0x5208", "cumulativeGasUsed": "0xa410", "effectiveGasPrice": "0x64",
		"contractAddress": BytesToHex(addr(0xcc)), "logsBloom": BytesToHex(make([]byte, 256)),
	}, b17))
	blob := typed(tx(b17, 2, 3, addr(0xbb)), 1)
	blob.Status = Uint32Ptr(0)
	blob.MaxFeePerBlobGas, blob.BlobVersionedHashes = StringPtr("1"), [][]byte{hash(0x01)}

	log := func(blk *BlockHeader, txIndex, logIndex uint32, address byte, topics ...[]byte) *Log {
		return &Log{
			Address: addr(address), Topics: topics, Data: []byte{0x01}, BlockNumber: blk.Number, BlockHash: blk.Hash,
			TransactionHash: txHash(blk.Number, txIndex), TransactionIndex: txIndex, LogIndex: logIndex,
			BlockTimestamp: Uint64Ptr(blk.Timestamp),
		}
	}
	// Block 0x10 comes from trace_block (Parity): transaction 1 calls aa,
	// which delegates, makes a call that fails, and selfdestructs. The failed
	// call's child succeeds on its own but is reverted with its parent.
	parityFrame := func(typ string, traceAddress []interface{}, action map[string]interface{}, result interface{}, errText string) map[string]interface{} {
		raw := map[string]interface{}{
			"type": typ, "action": action, "result": result, "traceAddress": traceAddress,
			"subtraces": float64(0), "transactionHash": txHashHex(0x10, 1), "transactionPosition": float64(1),
		}
		if errText != "" {
			raw["error"] = errText
		}
		return raw
	}
	call := func(callType string, from, to byte, value string) map[string]interface{} {
		return map[string]interface{}{
			"callType": callType, "from": BytesToHex(addr(from)), "to": BytesToHex(addr(to)), "value": value,
			"gas": "0x7530", "input": "0x",
		}
	}
	ok := map[string]interface{}{"gasUsed": "0x5208", "output": "0x"}
	var traces []*Trace
	for _, raw := range []map[string]interface{}{
		parityFrame("call", []interface{}{}, call("call", 0x01, 0xaa, "0x5"), ok, ""),
		parityFrame("call", []interface{}{float64(0)}, call("delegatecall", 0xaa, 0xbb, "0x5"), ok, ""),
		parityFrame("call", []interface{}{float64(1)}, call("call", 0xaa, 0xbb, "0x1"), nil, "execution reverted"),
		parityFrame("call", []interface{}{float64(1), float64(0)}, call("call", 0xbb, 0xaa, "0x3"), ok, ""),
		parityFrame("suicide", []interface{}{float64(2)}, map[string]interface{}{
			"address": BytesToHex(addr(0xaa)), "refundAddress": BytesToHex(addr(0xdd)), "balance": "0x7",
		}, nil, ""),
	} {
		traces = append(traces, must(TraceFromParity(raw, b16.Number, b16.Hash, &b16.Timestamp)))
	}
	// Block 0x11 comes from debug_traceBlockByNumber with the callTracer.
	for i, item := range []map[string]interface{}{
		{"txHash": txHashHex(0x11, 0), "result": map[string]interface{}{
			"type": "CALL", "from": BytesToHex(addr(0x01)), "to": BytesToHex(addr(0xbb)), "value": "0x0",
			"gas": "0x7530", "gasUsed": "0x5208", "input": "0x", "output": "0x",
			"calls": []interface{}{map[string]interface{}{
				"type": "STATICCALL", "from": BytesToHex(addr(0xbb)), "to": BytesToHex(addr(0xaa)),
				"gas": "0x7530", "gasUsed": "0x5208", "input": "0x", "output": "0x",
			}},
		}},
		{"txHash": txHashHex(0x11, 1), "result": map[string]interface{}{
			"type": "CREATE", "from": BytesToHex(addr(0x01)), "value": "0x2", "gas": "0x7530", "gasUsed": "0x5208",
			"input": "0x6080", "error": "out of gas",
		}},
	} {
		traces = append(traces, must(TraceFromGethDebug(item, uint32(i), b17.Number, b17.Hash, &b17.Timestamp))...)
	}
	if _, err := TraceFromGethDebug(map[string]interface{}{"result": map[string]interface{}{"type": "CALL"}}, 0, b17.Number, b17.Hash, nil); err == nil {
		panic("TraceFromGethDebug accepted an item without txHash")
	}
	// Transfers come from the frames before the Parity pass: the helper must
	// find the reverted child of the failed call on its own.
	transfers := NativeTransfersFromTraces(traces)
	PropagateParityReverted(traces)

	return &fakeQueryServer{
		blocks: []*BlockHeader{b16, b17},
		txs:    []*Transaction{legacy, dynamic, setCode, creation, blob},
		logs: []*Log{
			log(b16, 1, 0, 0xaa, hash(0x71)),
			log(b17, 0, 0, 0xbb, hash(0x72)),
			log(b17, 0, 1, 0xaa, hash(0x71), hash(0x99)),
		},
		traces:    traces,
		transfers: transfers,
	}
}

// run serves one JSON-RPC request and returns its result as a client decodes it.
func (s *fakeQueryServer) run(method string, chainID uint64, budget int, params json.RawMessage) (map[string]interface{}, error) {
	var rendered map[string]interface{}
	switch method {
	case "eth_queryBlocks":
		req, err := QueryBlocksRequestFromJsonRpc(chainID, params)
		if err != nil {
			return nil, err
		}
		match := func(b *BlockHeader) bool { return anyOf(req.GetFilter().GetMiner(), b.Miner) }
		resp := &QueryBlocksResponse{}
		page, err := s.page(req.FromBlock, req.ToBlock, req.Order, req.Target, budget, func(n uint64) []interface{} {
			if b := s.blocks[n-0x10]; match(b) {
				return []interface{}{b}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		for _, o := range page.objects {
			resp.Blocks = append(resp.Blocks, o.(*BlockHeader))
		}
		resp.FromBlock, resp.ToBlock, resp.CursorBlock = page.from, page.to, page.cursor
		rendered = QueryBlocksResponseToJsonRpc(chainID, req, resp)

	case "eth_queryTransactions":
		req, err := QueryTransactionsRequestFromJsonRpc(chainID, params)
		if err != nil {
			return nil, err
		}
		f := req.GetFilter()
		resp := &QueryTransactionsResponse{}
		page, err := s.page(req.FromBlock, req.ToBlock, req.Order, req.Target, budget, func(n uint64) []interface{} {
			var out []interface{}
			for _, tx := range s.txs {
				if tx.GetBlockNumber() == n && anyOf(f.GetFrom(), tx.From) && anyOf(f.GetTo(), tx.To) &&
					(len(f.GetSelector()) == 0 || len(tx.Input) >= 4 && anyOf(f.GetSelector(), tx.Input[:4])) {
					out = append(out, tx)
				}
			}
			return out
		})
		if err != nil {
			return nil, err
		}
		for _, o := range page.objects {
			tx := o.(*Transaction)
			resp.Transactions = append(resp.Transactions, tx)
			resp.Blocks = s.joinBlock(resp.Blocks, tx.GetBlockNumber())
		}
		resp.FromBlock, resp.ToBlock, resp.CursorBlock = page.from, page.to, page.cursor
		rendered = QueryTransactionsResponseToJsonRpc(chainID, req, resp)

	case "eth_queryLogs":
		req, err := QueryLogsRequestFromJsonRpc(chainID, params)
		if err != nil {
			return nil, err
		}
		f := req.GetFilter()
		resp := &QueryLogsResponse{}
		page, err := s.page(req.FromBlock, req.ToBlock, req.Order, req.Target, budget, func(n uint64) []interface{} {
			var out []interface{}
			for _, l := range s.logs {
				if l.BlockNumber == n && anyOf(f.GetAddress(), l.Address) && topicsMatch(f.GetTopics(), l.Topics) {
					out = append(out, l)
				}
			}
			return out
		})
		if err != nil {
			return nil, err
		}
		for _, o := range page.objects {
			l := o.(*Log)
			resp.Logs = append(resp.Logs, l)
			resp.Transactions = s.joinTransaction(resp.Transactions, l.TransactionHash)
			resp.Blocks = s.joinBlock(resp.Blocks, l.BlockNumber)
		}
		resp.FromBlock, resp.ToBlock, resp.CursorBlock = page.from, page.to, page.cursor
		rendered = QueryLogsResponseToJsonRpc(chainID, req, resp)

	case "eth_queryTraces":
		if chainID == 1 { // this fake indexes no traces for chain 1
			return nil, common.NewError(common.ErrorCode_UNSUPPORTED_METHOD, "traces are not indexed on this chain")
		}
		req, err := QueryTracesRequestFromJsonRpc(chainID, params)
		if err != nil {
			return nil, err
		}
		f := req.GetFilter()
		resp := &QueryTracesResponse{}
		page, err := s.page(req.FromBlock, req.ToBlock, req.Order, req.Target, budget, func(n uint64) []interface{} {
			var out []interface{}
			for _, tr := range s.traces {
				if tr.BlockNumber == n && (f.GetIncludeReverted() || !tr.Reverted) && anyOf(f.GetFrom(), tr.From) && anyOf(f.GetTo(), tr.To) {
					out = append(out, tr)
				}
			}
			return out
		})
		if err != nil {
			return nil, err
		}
		for _, o := range page.objects {
			tr := o.(*Trace)
			resp.Traces = append(resp.Traces, tr)
			resp.Transactions = s.joinTransaction(resp.Transactions, tr.TransactionHash)
			resp.Blocks = s.joinBlock(resp.Blocks, tr.BlockNumber)
		}
		resp.FromBlock, resp.ToBlock, resp.CursorBlock = page.from, page.to, page.cursor
		rendered = QueryTracesResponseToJsonRpc(chainID, req, resp)

	case "eth_queryTransfers":
		req, err := QueryTransfersRequestFromJsonRpc(chainID, params)
		if err != nil {
			return nil, err
		}
		f := req.GetFilter()
		transfers := s.transfers
		resp := &QueryTransfersResponse{}
		page, err := s.page(req.FromBlock, req.ToBlock, req.Order, req.Target, budget, func(n uint64) []interface{} {
			var out []interface{}
			for _, tr := range transfers {
				if tr.BlockNumber == n && (f.GetIncludeReverted() || !tr.Reverted) && anyOf(f.GetFrom(), tr.From) && anyOf(f.GetTo(), tr.To) {
					out = append(out, tr)
				}
			}
			return out
		})
		if err != nil {
			return nil, err
		}
		for _, o := range page.objects {
			tr := o.(*NativeTransfer)
			resp.Transfers = append(resp.Transfers, tr)
			resp.Transactions = s.joinTransaction(resp.Transactions, tr.TransactionHash)
			resp.Blocks = s.joinBlock(resp.Blocks, tr.BlockNumber)
		}
		resp.FromBlock, resp.ToBlock, resp.CursorBlock = page.from, page.to, page.cursor
		rendered = QueryTransfersResponseToJsonRpc(chainID, req, resp)
	}

	encoded, err := json.Marshal(rendered)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	return result, json.Unmarshal(encoded, &result)
}

type fakePage struct {
	objects          []interface{}
	from, to, cursor *CursorBlock
}

// page resolves the range and scans it one block at a time, as MIP-16
// Block-aligned pagination describes. primaries returns a block's matching
// primary objects in ascending order.
func (s *fakeQueryServer) page(fromTag, toTag *string, order *SortOrder, target *uint32, budget int, primaries func(uint64) []interface{}) (*fakePage, error) {
	desc := order != nil && *order == SortOrder_DESC
	resolve := func(tag *string, def string) uint64 {
		v := def
		if tag != nil {
			v = *tag
		}
		switch v {
		case "earliest":
			return 0
		case "finalized":
			return 0x10
		case "latest", "safe":
			return 0x11
		}
		n, _ := strconv.ParseUint(v[2:], 16, 64)
		return n
	}
	from, to := resolve(fromTag, "earliest"), resolve(toTag, "latest")
	if desc {
		from, to = resolve(fromTag, "latest"), resolve(toTag, "earliest")
	}
	sortOrder := SortOrder_ASC
	if desc {
		sortOrder = SortOrder_DESC
	}
	if err := ValidateQueryRange(sortOrder, from, to); err != nil {
		return nil, err
	}
	if min(from, to) < 0x10 || max(from, to) > 0x11 {
		return nil, common.NewError(common.ErrorCode_RANGE_OUTSIDE_AVAILABLE, "range is outside the availability window")
	}
	ref := func(n uint64) *CursorBlock {
		b := s.blocks[n-0x10]
		return &CursorBlock{Number: b.Number, Hash: b.Hash, ParentHash: b.ParentHash}
	}
	p := &fakePage{from: ref(from), to: ref(to)}
	for n := from; ; {
		objs := primaries(n)
		if budget > 0 && len(p.objects)+len(objs) > budget {
			if p.cursor == nil {
				// As a remote server returns it: a gRPC status with BDS details.
				return nil, common.NewError(common.ErrorCode_RANGE_TOO_LARGE, "budget exhausted").ToGRPCStatus().Err()
			}
			break
		}
		if desc {
			for i, j := 0, len(objs)-1; i < j; i, j = i+1, j-1 {
				objs[i], objs[j] = objs[j], objs[i]
			}
		}
		p.objects = append(p.objects, objs...)
		p.cursor = ref(n)
		if n == to || (target != nil && len(p.objects) >= int(*target)) {
			break
		}
		if desc {
			n--
		} else {
			n++
		}
	}
	return p, nil
}

func (s *fakeQueryServer) joinBlock(blocks []*BlockHeader, n uint64) []*BlockHeader {
	for _, b := range blocks {
		if b.Number == n {
			return blocks
		}
	}
	return append(blocks, s.blocks[n-0x10])
}

func (s *fakeQueryServer) joinTransaction(txs []*Transaction, hash []byte) []*Transaction {
	for _, tx := range txs {
		if bytes.Equal(tx.Hash, hash) {
			return txs
		}
	}
	for _, tx := range s.txs {
		if bytes.Equal(tx.Hash, hash) {
			return append(txs, tx)
		}
	}
	return txs
}

// anyOf reports whether v equals a value of list; an empty list matches all.
func anyOf(list [][]byte, v []byte) bool {
	if len(list) == 0 {
		return true
	}
	for _, x := range list {
		if bytes.Equal(x, v) {
			return true
		}
	}
	return false
}

// topicsMatch follows eth_getLogs: a log needs at least as many topics as
// there are positions, and an empty position is a wildcard.
func topicsMatch(filter []*TopicFilter, topics [][]byte) bool {
	if len(filter) > len(topics) {
		return false
	}
	for i, position := range filter {
		if !anyOf(position.GetValues(), topics[i]) {
			return false
		}
	}
	return true
}
