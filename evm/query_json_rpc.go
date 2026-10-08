package evm

// JSON-RPC codec of the eth_query* methods, as Monad MIP-16 (JSON-RPC Query
// Methods) defines them: strict request parsing, responses that carry exactly
// the selected fields, and the error codes of the methods. Servers own tag
// resolution, availability, budgets and paging.

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"

	"github.com/blockchain-data-standards/manifesto/common"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// JSON-RPC error codes of the eth_query* methods. See MIP-16 Errors.
const (
	// JsonRpcCodeInvalidParams: a malformed request or an inverted block range.
	JsonRpcCodeInvalidParams = -32602
	// JsonRpcCodeInternalError: a failure that has no MIP-16 code.
	JsonRpcCodeInternalError = -32603
	// JsonRpcCodeResourceNotFound: the resolved range is partly or wholly outside the availability window.
	JsonRpcCodeResourceNotFound = -32001
	// JsonRpcCodeMethodNotSupported: the node recognizes the method but does not serve it.
	JsonRpcCodeMethodNotSupported = -32004
	// JsonRpcCodeLimitExceeded: a server-imposed limit was exceeded, for example the budget before fromBlock completed.
	JsonRpcCodeLimitExceeded = -32005
)

// QueryParamsError is a malformed eth_query* request. Its JSON-RPC code is -32602.
type QueryParamsError struct {
	Message string
}

func (e *QueryParamsError) Error() string { return "invalid params: " + e.Message }

// Code returns the JSON-RPC error code of the error, -32602.
func (e *QueryParamsError) Code() int { return JsonRpcCodeInvalidParams }

func invalidParams(format string, args ...interface{}) error {
	return &QueryParamsError{Message: fmt.Sprintf(format, args...)}
}

// QueryErrorCode returns the JSON-RPC error code of an eth_query* failure:
// -32602 for a *QueryParamsError, and the MIP-16 code of a BDS
// common.ErrorCode carried by a *common.BaseError or a gRPC status. Any other
// error is -32603. A nil error is 0.
func QueryErrorCode(err error) int {
	if err == nil {
		return 0
	}
	var paramsErr *QueryParamsError
	if errors.As(err, &paramsErr) {
		return JsonRpcCodeInvalidParams
	}
	var baseErr *common.BaseError
	if errors.As(err, &baseErr) {
		return queryErrorCodeOf(baseErr.Code)
	}
	if st, ok := status.FromError(err); ok {
		if baseErr, ok := common.FromGRPCStatus(st); ok {
			return queryErrorCodeOf(baseErr.Code)
		}
	}
	return JsonRpcCodeInternalError
}

func queryErrorCodeOf(code common.ErrorCode) int {
	switch code {
	case common.ErrorCode_INVALID_REQUEST, common.ErrorCode_INVALID_PARAMETER, common.ErrorCode_UNSUPPORTED_BLOCK_TAG:
		return JsonRpcCodeInvalidParams
	case common.ErrorCode_RANGE_OUTSIDE_AVAILABLE, common.ErrorCode_DATA_NOT_FOUND:
		return JsonRpcCodeResourceNotFound
	case common.ErrorCode_UNSUPPORTED_METHOD:
		return JsonRpcCodeMethodNotSupported
	case common.ErrorCode_RANGE_TOO_LARGE, common.ErrorCode_TIMEOUT_ERROR, common.ErrorCode_RATE_LIMITED:
		return JsonRpcCodeLimitExceeded
	default:
		return JsonRpcCodeInternalError
	}
}

// ValidateQueryRange returns a -32602 error if the resolved range is inverted
// for order: fromBlock above toBlock in asc order, or below it in desc order.
// Servers call it after they resolve the block tags of a request.
func ValidateQueryRange(order SortOrder, from, to uint64) error {
	if order == SortOrder_DESC {
		if from < to {
			return invalidParams("inverted range: fromBlock %#x is below toBlock %#x in desc order", from, to)
		}
		return nil
	}
	if from > to {
		return invalidParams("inverted range: fromBlock %#x is above toBlock %#x in asc order", from, to)
	}
	return nil
}

// QuerySchema lists the JSON field names a chain supports for each MIP-16
// object type. Parsing rejects a requested name outside it, and "all" selects
// every name in it. The slices are shared; callers must not modify them.
type QuerySchema struct {
	Blocks       []string
	Transactions []string
	Logs         []string
	Traces       []string
	Transfers    []string
}

// monadTraceFields is the MIP-16 traces schema; transfers use the same schema.
var monadTraceFields = []string{
	"type", "from", "to", "value", "gas", "gasUsed", "input", "output", "error", "reverted",
	"blockHash", "blockNumber", "transactionHash", "transactionIndex", "traceAddress",
}

// monadQuerySchema is MIP-16 Appendix 1 (Monad response schemas).
var monadQuerySchema = QuerySchema{
	Blocks: []string{
		"number", "hash", "parentHash", "timestamp", "miner", "nonce", "mixHash", "sha3Uncles", "logsBloom",
		"transactionsRoot", "stateRoot", "receiptsRoot", "difficulty", "totalDifficulty", "extraData", "size",
		"gasLimit", "gasUsed", "baseFeePerGas", "withdrawalsRoot", "blobGasUsed", "excessBlobGas",
		"parentBeaconBlockRoot", "requestsHash",
	},
	Transactions: []string{
		"hash", "blockHash", "blockNumber", "blockTimestamp", "transactionIndex", "type", "from", "to", "nonce",
		"input", "value", "gas", "gasPrice", "maxFeePerGas", "maxPriorityFeePerGas", "chainId", "accessList",
		"authorizationList", "v", "yParity", "r", "s", "status", "gasUsed", "cumulativeGasUsed",
		"effectiveGasPrice", "contractAddress", "logsBloom",
	},
	Logs: []string{
		"address", "blockHash", "blockNumber", "blockTimestamp", "transactionHash", "transactionIndex",
		"logIndex", "topics", "data", "removed",
	},
	Traces:    monadTraceFields,
	Transfers: monadTraceFields,
}

// genericQuerySchema is every field the proto field selections define. It is
// built on first use: package variables initialize before the proto file
// descriptors register.
var genericQuerySchema = sync.OnceValue(func() QuerySchema {
	return QuerySchema{
		Blocks:       selectionFieldNames(&BlockFieldSelection{}),
		Transactions: selectionFieldNames(&TransactionFieldSelection{}),
		Logs:         selectionFieldNames(&LogFieldSelection{}),
		Traces:       selectionFieldNames(&TraceFieldSelection{}),
		Transfers:    selectionFieldNames(&TransferFieldSelection{}),
	}
})

// QuerySchemaForChain returns the MIP-16 field sets of chainID: Appendix 1
// for Monad mainnet (143) and testnet (10143), and every field of the proto
// field selections for any other chain.
func QuerySchemaForChain(chainID uint64) QuerySchema {
	switch chainID {
	case 143, 10143:
		return monadQuerySchema
	}
	return genericQuerySchema()
}

func selectionFieldNames(sel protoreflect.ProtoMessage) []string {
	fields := sel.ProtoReflect().Descriptor().Fields()
	names := make([]string, fields.Len())
	for i := range names {
		names[i] = string(fields.Get(i).Name())
	}
	return names
}

// Request parsing

// queryMethodSpec is the request shape of one eth_query* method.
type queryMethodSpec struct {
	primary    string
	relations  []string
	filterKeys []string
}

var (
	queryBlocksSpec       = queryMethodSpec{primary: "blocks", filterKeys: []string{"miner"}}
	queryTransactionsSpec = queryMethodSpec{primary: "transactions", relations: []string{"blocks"}, filterKeys: []string{"from", "to", "selector"}}
	queryLogsSpec         = queryMethodSpec{primary: "logs", relations: []string{"transactions", "blocks"}, filterKeys: []string{"address", "topics"}}
	queryTracesSpec       = queryMethodSpec{primary: "traces", relations: []string{"transactions", "blocks"}, filterKeys: []string{"from", "to", "selector", "includeReverted"}}
	queryTransfersSpec    = queryMethodSpec{primary: "transfers", relations: []string{"transactions", "blocks"}, filterKeys: []string{"from", "to", "includeReverted"}}
)

// queryRequest is a decoded MIP-16 request object. filter and fields are nil
// when the request omits them.
type queryRequest struct {
	fromBlock *string
	toBlock   *string
	order     *SortOrder
	target    *uint32
	filter    map[string]json.RawMessage
	fields    map[string]json.RawMessage
}

// QueryBlocksRequestFromJsonRpc parses the params of an eth_queryBlocks
// request for chainID. Every error is a *QueryParamsError (-32602). The
// block field selection is always set; omitted fields select the whole schema.
func QueryBlocksRequestFromJsonRpc(chainID uint64, params json.RawMessage) (*QueryBlocksRequest, error) {
	q, err := decodeQueryRequest(params, &queryBlocksSpec)
	if err != nil {
		return nil, err
	}
	schema := QuerySchemaForChain(chainID)
	req := &QueryBlocksRequest{FromBlock: q.fromBlock, ToBlock: q.toBlock, Order: q.order, Target: q.target, ChainId: chainIDPtr(chainID)}
	if q.filter != nil {
		req.Filter = &BlockFilter{}
		if req.Filter.Miner, err = filterDataList(q.filter, "miner", AddressLength); err != nil {
			return nil, err
		}
	}
	req.BlockFields = &BlockFieldSelection{}
	if err := primarySelection(q.fields, "blocks", schema.Blocks, req.BlockFields); err != nil {
		return nil, err
	}
	return req, nil
}

// QueryTransactionsRequestFromJsonRpc parses the params of an
// eth_queryTransactions request for chainID. Every error is a
// *QueryParamsError (-32602). The transaction field selection is always set;
// the block selection is set only when the request joins blocks.
func QueryTransactionsRequestFromJsonRpc(chainID uint64, params json.RawMessage) (*QueryTransactionsRequest, error) {
	q, err := decodeQueryRequest(params, &queryTransactionsSpec)
	if err != nil {
		return nil, err
	}
	schema := QuerySchemaForChain(chainID)
	req := &QueryTransactionsRequest{FromBlock: q.fromBlock, ToBlock: q.toBlock, Order: q.order, Target: q.target, ChainId: chainIDPtr(chainID)}
	if q.filter != nil {
		req.Filter = &TransactionFilter{}
		if req.Filter.From, err = filterDataList(q.filter, "from", AddressLength); err != nil {
			return nil, err
		}
		if req.Filter.To, err = filterDataList(q.filter, "to", AddressLength); err != nil {
			return nil, err
		}
		if req.Filter.Selector, err = filterDataList(q.filter, "selector", selectorLength); err != nil {
			return nil, err
		}
	}
	req.TransactionFields = &TransactionFieldSelection{}
	if err := primarySelection(q.fields, "transactions", schema.Transactions, req.TransactionFields); err != nil {
		return nil, err
	}
	if raw, ok := q.fields["blocks"]; ok {
		req.BlockFields = &BlockFieldSelection{}
		if err := parseSelection(raw, "blocks", schema.Blocks, req.BlockFields); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// QueryLogsRequestFromJsonRpc parses the params of an eth_queryLogs request
// for chainID. Every error is a *QueryParamsError (-32602). The log field
// selection is always set; the transaction and block selections are set only
// when the request joins them.
func QueryLogsRequestFromJsonRpc(chainID uint64, params json.RawMessage) (*QueryLogsRequest, error) {
	q, err := decodeQueryRequest(params, &queryLogsSpec)
	if err != nil {
		return nil, err
	}
	schema := QuerySchemaForChain(chainID)
	req := &QueryLogsRequest{FromBlock: q.fromBlock, ToBlock: q.toBlock, Order: q.order, Target: q.target, ChainId: chainIDPtr(chainID)}
	if q.filter != nil {
		req.Filter = &LogFilter{}
		if req.Filter.Address, err = filterDataList(q.filter, "address", AddressLength); err != nil {
			return nil, err
		}
		if req.Filter.Topics, err = filterTopics(q.filter["topics"]); err != nil {
			return nil, err
		}
	}
	req.LogFields = &LogFieldSelection{}
	if err := primarySelection(q.fields, "logs", schema.Logs, req.LogFields); err != nil {
		return nil, err
	}
	if req.TransactionFields, err = transactionRelation(q.fields, schema); err != nil {
		return nil, err
	}
	if req.BlockFields, err = blockRelation(q.fields, schema); err != nil {
		return nil, err
	}
	return req, nil
}

// QueryTracesRequestFromJsonRpc parses the params of an eth_queryTraces
// request for chainID. Every error is a *QueryParamsError (-32602). The trace
// field selection is always set; the transaction and block selections are set
// only when the request joins them.
func QueryTracesRequestFromJsonRpc(chainID uint64, params json.RawMessage) (*QueryTracesRequest, error) {
	q, err := decodeQueryRequest(params, &queryTracesSpec)
	if err != nil {
		return nil, err
	}
	schema := QuerySchemaForChain(chainID)
	req := &QueryTracesRequest{FromBlock: q.fromBlock, ToBlock: q.toBlock, Order: q.order, Target: q.target, ChainId: chainIDPtr(chainID)}
	if q.filter != nil {
		req.Filter = &TraceFilter{}
		if req.Filter.From, err = filterDataList(q.filter, "from", AddressLength); err != nil {
			return nil, err
		}
		if req.Filter.To, err = filterDataList(q.filter, "to", AddressLength); err != nil {
			return nil, err
		}
		if req.Filter.Selector, err = filterDataList(q.filter, "selector", selectorLength); err != nil {
			return nil, err
		}
		if req.Filter.IncludeReverted, err = filterBool(q.filter, "includeReverted"); err != nil {
			return nil, err
		}
	}
	req.TraceFields = &TraceFieldSelection{}
	if err := primarySelection(q.fields, "traces", schema.Traces, req.TraceFields); err != nil {
		return nil, err
	}
	if req.TransactionFields, err = transactionRelation(q.fields, schema); err != nil {
		return nil, err
	}
	if req.BlockFields, err = blockRelation(q.fields, schema); err != nil {
		return nil, err
	}
	return req, nil
}

// QueryTransfersRequestFromJsonRpc parses the params of an eth_queryTransfers
// request for chainID. Every error is a *QueryParamsError (-32602). The
// transfer field selection is always set; the transaction and block
// selections are set only when the request joins them.
func QueryTransfersRequestFromJsonRpc(chainID uint64, params json.RawMessage) (*QueryTransfersRequest, error) {
	q, err := decodeQueryRequest(params, &queryTransfersSpec)
	if err != nil {
		return nil, err
	}
	schema := QuerySchemaForChain(chainID)
	req := &QueryTransfersRequest{FromBlock: q.fromBlock, ToBlock: q.toBlock, Order: q.order, Target: q.target, ChainId: chainIDPtr(chainID)}
	if q.filter != nil {
		req.Filter = &TransferFilter{}
		if req.Filter.From, err = filterDataList(q.filter, "from", AddressLength); err != nil {
			return nil, err
		}
		if req.Filter.To, err = filterDataList(q.filter, "to", AddressLength); err != nil {
			return nil, err
		}
		if req.Filter.IncludeReverted, err = filterBool(q.filter, "includeReverted"); err != nil {
			return nil, err
		}
	}
	req.TransferFields = &TransferFieldSelection{}
	if err := primarySelection(q.fields, "transfers", schema.Transfers, req.TransferFields); err != nil {
		return nil, err
	}
	if req.TransactionFields, err = transactionRelation(q.fields, schema); err != nil {
		return nil, err
	}
	if req.BlockFields, err = blockRelation(q.fields, schema); err != nil {
		return nil, err
	}
	return req, nil
}

const selectorLength = 4

func chainIDPtr(chainID uint64) *uint64 {
	if chainID == 0 {
		return nil
	}
	return &chainID
}

// decodeQueryRequest checks the request envelope of MIP-16 Request: params is
// an array with exactly one element, an object with only the request keys.
func decodeQueryRequest(params json.RawMessage, spec *queryMethodSpec) (*queryRequest, error) {
	var arr []json.RawMessage
	if !jsonIs(params, '[') || json.Unmarshal(params, &arr) != nil || len(arr) != 1 {
		return nil, invalidParams("params must be an array with exactly one element")
	}
	obj, err := jsonObject(arr[0], "params[0]")
	if err != nil {
		return nil, err
	}
	var q queryRequest
	for key, raw := range obj {
		switch key {
		case "fromBlock":
			if q.fromBlock, err = blockParam(raw, key); err != nil {
				return nil, err
			}
		case "toBlock":
			if q.toBlock, err = blockParam(raw, key); err != nil {
				return nil, err
			}
		case "order":
			s, ok := jsonString(raw)
			var order SortOrder
			switch {
			case ok && s == "asc":
				order = SortOrder_ASC
			case ok && s == "desc":
				order = SortOrder_DESC
			default:
				return nil, invalidParams(`order must be "asc" or "desc"`)
			}
			q.order = &order
		case "target":
			s, _ := jsonString(raw)
			n, ok := parseQuantity(s)
			if !ok || n == 0 {
				return nil, invalidParams("target must be a QUANTITY of at least 0x1")
			}
			// A target beyond the proto's uint32 is clamped: no page holds that many objects.
			target := uint32(min(n, math.MaxUint32))
			q.target = &target
		case "filter":
			if q.filter, err = jsonObject(raw, "filter"); err != nil {
				return nil, err
			}
			for k := range q.filter {
				if !contains(spec.filterKeys, k) {
					return nil, invalidParams("unknown filter key %q", k)
				}
			}
		case "fields":
			if q.fields, err = jsonObject(raw, "fields"); err != nil {
				return nil, err
			}
			if _, ok := q.fields[spec.primary]; !ok {
				return nil, invalidParams("fields must contain the primary key %q", spec.primary)
			}
			for k := range q.fields {
				if k != spec.primary && !contains(spec.relations, k) {
					return nil, invalidParams("fields key %q is not an object type or relation of this method", k)
				}
			}
		default:
			return nil, invalidParams("unknown request key %q", key)
		}
	}
	return &q, nil
}

// blockParam parses fromBlock or toBlock: a QUANTITY, stored in canonical
// form, or one of the MIP-16 tags.
func blockParam(raw json.RawMessage, key string) (*string, error) {
	s, _ := jsonString(raw)
	switch s {
	case "latest", "earliest", "safe", "finalized":
		return &s, nil
	}
	n, ok := parseQuantity(s)
	if !ok {
		return nil, invalidParams(`%s must be a QUANTITY or one of "latest", "earliest", "safe", "finalized"`, key)
	}
	canonical := quantity(n)
	return &canonical, nil
}

// parseQuantity parses a 64-bit QUANTITY: 0x and hex digits without leading zeros.
func parseQuantity(s string) (uint64, bool) {
	if len(s) < 3 || len(s) > 18 || s[0] != '0' || s[1] != 'x' || (s[2] == '0' && len(s) > 3) {
		return 0, false
	}
	n, err := strconv.ParseUint(s[2:], 16, 64)
	return n, err == nil
}

// filterDataList parses a filter field of DATA values of size bytes: null or
// absent for no constraint, a single value, or an array without null entries.
func filterDataList(filter map[string]json.RawMessage, key string, size int) ([][]byte, error) {
	raw, ok := filter[key]
	if !ok || jsonIsNull(raw) {
		return nil, nil
	}
	if s, ok := jsonString(raw); ok {
		b, err := filterData(s, key, size)
		if err != nil {
			return nil, err
		}
		return [][]byte{b}, nil
	}
	return filterDataArray(raw, key, size)
}

func filterDataArray(raw json.RawMessage, key string, size int) ([][]byte, error) {
	var items []json.RawMessage
	if !jsonIs(raw, '[') || json.Unmarshal(raw, &items) != nil {
		return nil, invalidParams("filter.%s must be DATA or DATA[]", key)
	}
	if len(items) == 0 {
		return nil, nil
	}
	out := make([][]byte, len(items))
	for i, item := range items {
		s, ok := jsonString(item)
		if !ok {
			return nil, invalidParams("filter.%s[%d] must be DATA, not %s", key, i, item)
		}
		b, err := filterData(s, key, size)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func filterData(s, key string, size int) ([]byte, error) {
	if len(s) != 2+2*size || s[0] != '0' || s[1] != 'x' {
		return nil, invalidParams("filter.%s values must be %d-byte DATA", key, size)
	}
	b, err := hex.DecodeString(s[2:])
	if err != nil {
		return nil, invalidParams("filter.%s values must be %d-byte DATA", key, size)
	}
	return b, nil
}

// filterTopics parses the positional topics filter: up to 4 entries, each
// DATA, DATA[] or null. A null or empty entry is a wildcard (empty values).
func filterTopics(raw json.RawMessage) ([]*TopicFilter, error) {
	if raw == nil || jsonIsNull(raw) {
		return nil, nil
	}
	var positions []json.RawMessage
	if !jsonIs(raw, '[') || json.Unmarshal(raw, &positions) != nil {
		return nil, invalidParams("filter.topics must be an array")
	}
	if len(positions) > 4 {
		return nil, invalidParams("filter.topics has %d positions; at most 4 are allowed", len(positions))
	}
	topics := make([]*TopicFilter, len(positions))
	for i, position := range positions {
		topic := &TopicFilter{}
		switch {
		case jsonIsNull(position):
		case jsonIs(position, '"'):
			s, _ := jsonString(position)
			b, err := filterData(s, "topics", HashLength)
			if err != nil {
				return nil, err
			}
			topic.Values = [][]byte{b}
		default:
			values, err := filterDataArray(position, "topics", HashLength)
			if err != nil {
				return nil, err
			}
			topic.Values = values
		}
		topics[i] = topic
	}
	return topics, nil
}

// filterBool parses a single-boolean filter field; null or absent is nil.
func filterBool(filter map[string]json.RawMessage, key string) (*bool, error) {
	raw, ok := filter[key]
	if !ok || jsonIsNull(raw) {
		return nil, nil
	}
	var v bool
	if json.Unmarshal(raw, &v) != nil {
		return nil, invalidParams("filter.%s must be a boolean", key)
	}
	return &v, nil
}

// primarySelection fills sel from fields[key], or with the whole schema when
// the request omits fields. decodeQueryRequest has checked that the key is
// present when fields is.
func primarySelection(fields map[string]json.RawMessage, key string, names []string, sel protoreflect.ProtoMessage) error {
	if fields == nil {
		selectAll(sel, names)
		return nil
	}
	return parseSelection(fields[key], key, names, sel)
}

func transactionRelation(fields map[string]json.RawMessage, schema QuerySchema) (*TransactionFieldSelection, error) {
	raw, ok := fields["transactions"]
	if !ok {
		return nil, nil
	}
	sel := &TransactionFieldSelection{}
	if err := parseSelection(raw, "transactions", schema.Transactions, sel); err != nil {
		return nil, err
	}
	return sel, nil
}

func blockRelation(fields map[string]json.RawMessage, schema QuerySchema) (*BlockFieldSelection, error) {
	raw, ok := fields["blocks"]
	if !ok {
		return nil, nil
	}
	sel := &BlockFieldSelection{}
	if err := parseSelection(raw, "blocks", schema.Blocks, sel); err != nil {
		return nil, err
	}
	return sel, nil
}

// parseSelection sets the fields of sel that a fields value names: a
// non-empty array of names in the schema, or "all" for the whole schema.
func parseSelection(raw json.RawMessage, key string, names []string, sel protoreflect.ProtoMessage) error {
	if s, ok := jsonString(raw); ok {
		if s != "all" {
			return invalidParams(`fields.%s must be an array of field names or "all"`, key)
		}
		selectAll(sel, names)
		return nil
	}
	var list []json.RawMessage
	if !jsonIs(raw, '[') || json.Unmarshal(raw, &list) != nil {
		return invalidParams(`fields.%s must be an array of field names or "all"`, key)
	}
	if len(list) == 0 {
		return invalidParams("fields.%s must not be empty", key)
	}
	m := sel.ProtoReflect()
	descriptors := m.Descriptor().Fields()
	for _, item := range list {
		name, ok := jsonString(item)
		if !ok {
			return invalidParams("fields.%s must contain only field names", key)
		}
		fd := descriptors.ByName(protoreflect.Name(name))
		if fd == nil || !contains(names, name) {
			return invalidParams("unknown field %q in fields.%s", name, key)
		}
		m.Set(fd, protoreflect.ValueOfBool(true))
	}
	return nil
}

func selectAll(sel protoreflect.ProtoMessage, names []string) {
	m := sel.ProtoReflect()
	descriptors := m.Descriptor().Fields()
	for _, name := range names {
		if fd := descriptors.ByName(protoreflect.Name(name)); fd != nil {
			m.Set(fd, protoreflect.ValueOfBool(true))
		}
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// jsonIs reports whether the first non-space byte of raw is c.
func jsonIs(raw json.RawMessage, c byte) bool {
	raw = bytes.TrimLeft(raw, " \t\r\n")
	return len(raw) > 0 && raw[0] == c
}

func jsonIsNull(raw json.RawMessage) bool {
	return string(bytes.TrimSpace(raw)) == "null"
}

func jsonString(raw json.RawMessage) (string, bool) {
	var s string
	if !jsonIs(raw, '"') || json.Unmarshal(raw, &s) != nil {
		return "", false
	}
	return s, true
}

func jsonObject(raw json.RawMessage, what string) (map[string]json.RawMessage, error) {
	var obj map[string]json.RawMessage
	if !jsonIs(raw, '{') || json.Unmarshal(raw, &obj) != nil {
		return nil, invalidParams("%s must be an object", what)
	}
	return obj, nil
}

// Response rendering

// QueryBlocksResponseToJsonRpc renders an eth_queryBlocks result for chainID.
// Block objects carry exactly the fields req selects; a nil selection is the
// whole schema of the chain.
func QueryBlocksResponseToJsonRpc(chainID uint64, req *QueryBlocksRequest, resp *QueryBlocksResponse) map[string]interface{} {
	sel := req.GetBlockFields()
	if sel == nil {
		sel = &BlockFieldSelection{}
		selectAll(sel, QuerySchemaForChain(chainID).Blocks)
	}
	data := map[string]interface{}{"blocks": queryBlocksToJsonRpc(resp.GetBlocks(), sel)}
	return queryResultToJsonRpc(data, resp.GetFromBlock(), resp.GetToBlock(), resp.GetCursorBlock())
}

// QueryTransactionsResponseToJsonRpc renders an eth_queryTransactions result
// for chainID. Objects carry exactly the fields req selects; a nil
// transaction selection is the whole schema of the chain, and blocks appear
// only when req joins them.
func QueryTransactionsResponseToJsonRpc(chainID uint64, req *QueryTransactionsRequest, resp *QueryTransactionsResponse) map[string]interface{} {
	sel := req.GetTransactionFields()
	if sel == nil {
		sel = &TransactionFieldSelection{}
		selectAll(sel, QuerySchemaForChain(chainID).Transactions)
	}
	data := map[string]interface{}{
		"transactions": queryTransactionsToJsonRpc(resp.GetTransactions(), sel, SignatureEncodingForChain(chainID)),
	}
	if blockSel := req.GetBlockFields(); blockSel != nil {
		data["blocks"] = queryBlocksToJsonRpc(resp.GetBlocks(), blockSel)
	}
	return queryResultToJsonRpc(data, resp.GetFromBlock(), resp.GetToBlock(), resp.GetCursorBlock())
}

// QueryLogsResponseToJsonRpc renders an eth_queryLogs result for chainID.
// Objects carry exactly the fields req selects; a nil log selection is the
// whole schema of the chain, and relations appear only when req joins them.
func QueryLogsResponseToJsonRpc(chainID uint64, req *QueryLogsRequest, resp *QueryLogsResponse) map[string]interface{} {
	sel := req.GetLogFields()
	if sel == nil {
		sel = &LogFieldSelection{}
		selectAll(sel, QuerySchemaForChain(chainID).Logs)
	}
	logs := make([]interface{}, len(resp.GetLogs()))
	for i, l := range resp.GetLogs() {
		logs[i] = queryLogToJsonRpc(l, sel)
	}
	data := map[string]interface{}{"logs": logs}
	addQueryRelations(data, chainID, req.GetTransactionFields(), resp.GetTransactions(), req.GetBlockFields(), resp.GetBlocks())
	return queryResultToJsonRpc(data, resp.GetFromBlock(), resp.GetToBlock(), resp.GetCursorBlock())
}

// QueryTracesResponseToJsonRpc renders an eth_queryTraces result for chainID.
// Objects carry exactly the fields req selects; a nil trace selection is the
// whole schema of the chain, and relations appear only when req joins them.
func QueryTracesResponseToJsonRpc(chainID uint64, req *QueryTracesRequest, resp *QueryTracesResponse) map[string]interface{} {
	sel := req.GetTraceFields()
	if sel == nil {
		sel = &TraceFieldSelection{}
		selectAll(sel, QuerySchemaForChain(chainID).Traces)
	}
	traces := make([]interface{}, len(resp.GetTraces()))
	for i, t := range resp.GetTraces() {
		traces[i] = traceFrameToJsonRpc(sel, traceFrame{
			typ: traceFrameType(t), from: t.From, to: t.To, value: t.Value, gas: t.Gas, gasUsed: t.GasUsed,
			input: t.Input, output: t.Output, err: t.Error, reverted: t.Reverted, blockHash: t.BlockHash,
			blockNumber: t.BlockNumber, transactionHash: t.TransactionHash, transactionIndex: t.TransactionIndex,
			traceAddress: t.TraceAddress,
		})
	}
	data := map[string]interface{}{"traces": traces}
	addQueryRelations(data, chainID, req.GetTransactionFields(), resp.GetTransactions(), req.GetBlockFields(), resp.GetBlocks())
	return queryResultToJsonRpc(data, resp.GetFromBlock(), resp.GetToBlock(), resp.GetCursorBlock())
}

// QueryTransfersResponseToJsonRpc renders an eth_queryTransfers result for
// chainID. Objects carry exactly the fields req selects; a nil transfer
// selection is the whole schema of the chain, and relations appear only when
// req joins them.
func QueryTransfersResponseToJsonRpc(chainID uint64, req *QueryTransfersRequest, resp *QueryTransfersResponse) map[string]interface{} {
	sel := req.GetTransferFields()
	if sel == nil {
		sel = &TransferFieldSelection{}
		selectAll(sel, QuerySchemaForChain(chainID).Transfers)
	}
	transfers := make([]interface{}, len(resp.GetTransfers()))
	for i, t := range resp.GetTransfers() {
		transfers[i] = traceFrameToJsonRpc(sel, traceFrame{
			typ: t.Type, from: t.From, to: t.To, value: t.Value, gas: t.Gas, gasUsed: t.GasUsed,
			input: t.Input, output: t.Output, err: t.Error, reverted: t.Reverted, blockHash: t.BlockHash,
			blockNumber: t.BlockNumber, transactionHash: t.TransactionHash, transactionIndex: t.TransactionIndex,
			traceAddress: t.TraceAddress,
		})
	}
	data := map[string]interface{}{"transfers": transfers}
	addQueryRelations(data, chainID, req.GetTransactionFields(), resp.GetTransactions(), req.GetBlockFields(), resp.GetBlocks())
	return queryResultToJsonRpc(data, resp.GetFromBlock(), resp.GetToBlock(), resp.GetCursorBlock())
}

func addQueryRelations(data map[string]interface{}, chainID uint64, txSel *TransactionFieldSelection, txs []*Transaction, blockSel *BlockFieldSelection, blocks []*BlockHeader) {
	if txSel != nil {
		data["transactions"] = queryTransactionsToJsonRpc(txs, txSel, SignatureEncodingForChain(chainID))
	}
	if blockSel != nil {
		data["blocks"] = queryBlocksToJsonRpc(blocks, blockSel)
	}
}

func queryResultToJsonRpc(data map[string]interface{}, from, to, cursor *CursorBlock) map[string]interface{} {
	return map[string]interface{}{
		"data":        data,
		"fromBlock":   blockRefToJsonRpc(from),
		"toBlock":     blockRefToJsonRpc(to),
		"cursorBlock": blockRefToJsonRpc(cursor),
	}
}

func blockRefToJsonRpc(ref *CursorBlock) interface{} {
	if ref == nil {
		return nil
	}
	return map[string]interface{}{
		"number":     quantity(ref.Number),
		"hash":       BytesToHex(ref.Hash),
		"parentHash": BytesToHex(ref.ParentHash),
	}
}

func quantity(n uint64) string {
	return "0x" + strconv.FormatUint(n, 16)
}

// setQuantityString sets o[key] to the canonical QUANTITY of a decimal or hex
// string; a nil or malformed value leaves the field absent.
func setQuantityString(o map[string]interface{}, key string, v *string) {
	if v == nil {
		return
	}
	if q, err := DecimalStringToHex(*v); err == nil {
		o[key] = q
	}
}

func addressOrNull(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return BytesToHex(b)
}

func queryBlocksToJsonRpc(blocks []*BlockHeader, sel *BlockFieldSelection) []interface{} {
	out := make([]interface{}, len(blocks))
	for i, b := range blocks {
		out[i] = queryBlockToJsonRpc(b, sel)
	}
	return out
}

// queryBlockToJsonRpc renders the selected fields of a MIP-16 blocks object.
// A selected field whose optional value is absent (requestsHash before Prague
// or MONAD_FOUR, for example) is omitted.
func queryBlockToJsonRpc(h *BlockHeader, sel *BlockFieldSelection) map[string]interface{} {
	o := make(map[string]interface{}, 24)
	if sel.Number {
		o["number"] = quantity(h.Number)
	}
	if sel.Hash {
		o["hash"] = BytesToHex(h.Hash)
	}
	if sel.ParentHash {
		o["parentHash"] = BytesToHex(h.ParentHash)
	}
	if sel.Timestamp {
		o["timestamp"] = quantity(h.Timestamp)
	}
	if sel.Miner {
		o["miner"] = BytesToHex(h.Miner)
	}
	if sel.Nonce && h.Nonce != nil {
		o["nonce"] = fmt.Sprintf("0x%016x", *h.Nonce)
	}
	if sel.MixHash && h.MixHash != nil {
		o["mixHash"] = BytesToHex(h.MixHash)
	}
	if sel.Sha3Uncles {
		o["sha3Uncles"] = BytesToHex(h.Sha3Uncles)
	}
	if sel.LogsBloom {
		o["logsBloom"] = BytesToHex(h.LogsBloom)
	}
	if sel.TransactionsRoot {
		o["transactionsRoot"] = BytesToHex(h.TransactionsRoot)
	}
	if sel.StateRoot {
		o["stateRoot"] = BytesToHex(h.StateRoot)
	}
	if sel.ReceiptsRoot {
		o["receiptsRoot"] = BytesToHex(h.ReceiptsRoot)
	}
	if sel.Difficulty {
		setQuantityString(o, "difficulty", h.Difficulty)
	}
	if sel.TotalDifficulty {
		setQuantityString(o, "totalDifficulty", h.TotalDifficulty)
	}
	if sel.ExtraData {
		o["extraData"] = BytesToHex(h.ExtraData)
	}
	if sel.Size {
		o["size"] = quantity(h.Size)
	}
	if sel.GasLimit {
		o["gasLimit"] = quantity(h.GasLimit)
	}
	if sel.GasUsed {
		o["gasUsed"] = quantity(h.GasUsed)
	}
	if sel.BaseFeePerGas {
		setQuantityString(o, "baseFeePerGas", h.BaseFeePerGas)
	}
	if sel.WithdrawalsRoot && h.WithdrawalsRoot != nil {
		o["withdrawalsRoot"] = BytesToHex(h.WithdrawalsRoot)
	}
	if sel.BlobGasUsed && h.BlobGasUsed != nil {
		o["blobGasUsed"] = quantity(*h.BlobGasUsed)
	}
	if sel.ExcessBlobGas && h.ExcessBlobGas != nil {
		o["excessBlobGas"] = quantity(*h.ExcessBlobGas)
	}
	if sel.ParentBeaconBlockRoot && h.ParentBeaconBlockRoot != nil {
		o["parentBeaconBlockRoot"] = BytesToHex(h.ParentBeaconBlockRoot)
	}
	if sel.RequestsHash && h.RequestsHash != nil {
		o["requestsHash"] = BytesToHex(h.RequestsHash)
	}
	return o
}

func queryTransactionsToJsonRpc(txs []*Transaction, sel *TransactionFieldSelection, sig SignatureEncoding) []interface{} {
	out := make([]interface{}, len(txs))
	for i, tx := range txs {
		out[i] = queryTransactionToJsonRpc(tx, sel, sig)
	}
	return out
}

// queryTransactionToJsonRpc renders the selected fields of a MIP-16
// transactions object (transaction and receipt fields). Fields that do not
// apply to the transaction type are omitted even when selected, as MIP-16
// Appendix 1 specifies: maxFeePerGas and maxPriorityFeePerGas only on types
// 0x2-0x4, accessList on 0x1-0x4, authorizationList on 0x4, the blob fields
// on 0x3, yParity on typed transactions, chainId except on unprotected 0x0.
// Types above 0x4 (chain-specific) carry each such field when it is present.
func queryTransactionToJsonRpc(tx *Transaction, sel *TransactionFieldSelection, sig SignatureEncoding) map[string]interface{} {
	typ := tx.Type
	standard := typ <= 4
	o := make(map[string]interface{}, 32)
	if sel.Hash {
		o["hash"] = BytesToHex(tx.Hash)
	}
	if sel.BlockHash && tx.BlockHash != nil {
		o["blockHash"] = BytesToHex(tx.BlockHash)
	}
	if sel.BlockNumber && tx.BlockNumber != nil {
		o["blockNumber"] = quantity(*tx.BlockNumber)
	}
	if sel.BlockTimestamp && tx.BlockTimestamp != nil {
		o["blockTimestamp"] = quantity(*tx.BlockTimestamp)
	}
	if sel.TransactionIndex && tx.TransactionIndex != nil {
		o["transactionIndex"] = quantity(uint64(*tx.TransactionIndex))
	}
	if sel.Type {
		o["type"] = quantity(uint64(typ))
	}
	if sel.From {
		o["from"] = BytesToHex(tx.From)
	}
	if sel.To {
		o["to"] = addressOrNull(tx.To)
	}
	if sel.Nonce {
		o["nonce"] = quantity(tx.Nonce)
	}
	if sel.Input {
		o["input"] = BytesToHex(tx.Input)
	}
	if sel.Value {
		v := tx.Value
		setQuantityString(o, "value", &v)
	}
	if sel.Gas {
		o["gas"] = quantity(tx.GasLimit)
	}
	if sel.GasPrice {
		setQuantityString(o, "gasPrice", tx.GasPrice)
	}
	if !standard || typ >= 2 {
		if sel.MaxFeePerGas {
			setQuantityString(o, "maxFeePerGas", tx.MaxFeePerGas)
		}
		if sel.MaxPriorityFeePerGas {
			setQuantityString(o, "maxPriorityFeePerGas", tx.MaxPriorityFeePerGas)
		}
	}
	if sel.ChainId && tx.ChainId != nil && !(typ == 0 && unprotectedLegacy(tx.V)) {
		o["chainId"] = quantity(*tx.ChainId)
	}
	if sel.AccessList && ((standard && typ >= 1) || (!standard && len(tx.AccessList) > 0)) {
		o["accessList"] = queryAccessListToJsonRpc(tx.AccessList)
	}
	if sel.AuthorizationList && ((standard && typ == 4) || (!standard && len(tx.AuthorizationList) > 0)) {
		o["authorizationList"] = queryAuthorizationListToJsonRpc(tx.AuthorizationList, sig)
	}
	if !standard || typ == 3 {
		if sel.MaxFeePerBlobGas {
			setQuantityString(o, "maxFeePerBlobGas", tx.MaxFeePerBlobGas)
		}
		if sel.BlobVersionedHashes && (typ == 3 || len(tx.BlobVersionedHashes) > 0) {
			hashes := make([]string, len(tx.BlobVersionedHashes))
			for i, h := range tx.BlobVersionedHashes {
				hashes[i] = BytesToHex(h)
			}
			o["blobVersionedHashes"] = hashes
		}
	}
	if sel.V && tx.V != nil {
		o["v"] = BytesToQuantityHex(tx.V)
	}
	if sel.YParity && tx.YParity != nil && typ != 0 {
		o["yParity"] = quantity(uint64(*tx.YParity))
	}
	if sel.R {
		o["r"] = sig.hex(tx.R)
	}
	if sel.S {
		o["s"] = sig.hex(tx.S)
	}
	if sel.Status && tx.Status != nil {
		o["status"] = quantity(uint64(*tx.Status))
	}
	if sel.GasUsed && tx.GasUsed != nil {
		o["gasUsed"] = quantity(*tx.GasUsed)
	}
	if sel.CumulativeGasUsed && tx.CumulativeGasUsed != nil {
		o["cumulativeGasUsed"] = quantity(*tx.CumulativeGasUsed)
	}
	if sel.EffectiveGasPrice {
		setQuantityString(o, "effectiveGasPrice", tx.EffectiveGasPrice)
	}
	if sel.ContractAddress {
		o["contractAddress"] = addressOrNull(tx.ContractAddress)
	}
	if sel.LogsBloom {
		o["logsBloom"] = BytesToHex(tx.LogsBloom)
	}
	return o
}

// unprotectedLegacy reports whether a legacy signature's v is 27 or 28, which
// means the signature carries no chain ID (pre-EIP-155).
func unprotectedLegacy(v []byte) bool {
	if len(v) == 0 {
		return false
	}
	n := new(big.Int).SetBytes(v)
	return n.IsUint64() && (n.Uint64() == 27 || n.Uint64() == 28)
}

func queryAccessListToJsonRpc(items []*AccessListItem) []interface{} {
	out := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		keys := make([]string, len(item.StorageKeys))
		for i, k := range item.StorageKeys {
			keys[i] = BytesToHex(k)
		}
		out = append(out, map[string]interface{}{"address": BytesToHex(item.Address), "storageKeys": keys})
	}
	return out
}

// queryAuthorizationListToJsonRpc renders EIP-7702 tuples with exactly the
// MIP-16 keys: chainId, address, nonce, yParity, r, s.
func queryAuthorizationListToJsonRpc(items []*AuthorizationListItem, sig SignatureEncoding) []interface{} {
	out := make([]interface{}, 0, len(items))
	for _, auth := range items {
		if auth == nil {
			continue
		}
		out = append(out, map[string]interface{}{
			"chainId": quantity(auth.ChainId),
			"address": BytesToHex(auth.Address),
			"nonce":   quantity(auth.Nonce),
			"yParity": quantity(uint64(auth.YParity)),
			"r":       sig.hex(auth.R),
			"s":       sig.hex(auth.S),
		})
	}
	return out
}

// queryLogToJsonRpc renders the selected fields of a MIP-16 logs object.
func queryLogToJsonRpc(l *Log, sel *LogFieldSelection) map[string]interface{} {
	o := make(map[string]interface{}, 10)
	if sel.Address {
		o["address"] = BytesToHex(l.Address)
	}
	if sel.BlockHash {
		o["blockHash"] = BytesToHex(l.BlockHash)
	}
	if sel.BlockNumber {
		o["blockNumber"] = quantity(l.BlockNumber)
	}
	if sel.BlockTimestamp && l.BlockTimestamp != nil {
		o["blockTimestamp"] = quantity(*l.BlockTimestamp)
	}
	if sel.TransactionHash {
		o["transactionHash"] = BytesToHex(l.TransactionHash)
	}
	if sel.TransactionIndex {
		o["transactionIndex"] = quantity(uint64(l.TransactionIndex))
	}
	if sel.LogIndex {
		o["logIndex"] = quantity(uint64(l.LogIndex))
	}
	if sel.Topics {
		topics := make([]string, len(l.Topics))
		for i, t := range l.Topics {
			topics[i] = BytesToHex(t)
		}
		o["topics"] = topics
	}
	if sel.Data {
		o["data"] = BytesToHex(l.Data)
	}
	if sel.Removed {
		o["removed"] = false
	}
	return o
}

// traceSelection is the MIP-16 traces schema, which TraceFieldSelection and
// TransferFieldSelection both implement.
type traceSelection interface {
	GetType() bool
	GetFrom() bool
	GetTo() bool
	GetValue() bool
	GetGas() bool
	GetGasUsed() bool
	GetInput() bool
	GetOutput() bool
	GetError() bool
	GetReverted() bool
	GetBlockHash() bool
	GetBlockNumber() bool
	GetTransactionHash() bool
	GetTransactionIndex() bool
	GetTraceAddress() bool
}

// traceFrame is one call frame as a Trace or NativeTransfer carries it.
type traceFrame struct {
	typ              string
	from, to         []byte
	value            string
	gas, gasUsed     uint64
	input, output    []byte
	err              *string
	reverted         bool
	blockHash        []byte
	blockNumber      uint64
	transactionHash  []byte
	transactionIndex uint32
	traceAddress     []uint32
}

// traceFrameToJsonRpc renders the selected fields of a MIP-16 traces or
// transfers object.
func traceFrameToJsonRpc(sel traceSelection, f traceFrame) map[string]interface{} {
	o := make(map[string]interface{}, 15)
	if sel.GetType() {
		o["type"] = f.typ
	}
	if sel.GetFrom() {
		o["from"] = BytesToHex(f.from)
	}
	if sel.GetTo() {
		o["to"] = addressOrNull(f.to)
	}
	if sel.GetValue() {
		setQuantityString(o, "value", &f.value)
	}
	if sel.GetGas() {
		o["gas"] = quantity(f.gas)
	}
	if sel.GetGasUsed() {
		o["gasUsed"] = quantity(f.gasUsed)
	}
	if sel.GetInput() {
		o["input"] = BytesToHex(f.input)
	}
	if sel.GetOutput() {
		o["output"] = BytesToHex(f.output)
	}
	if sel.GetError() {
		if f.err != nil {
			o["error"] = *f.err
		} else {
			o["error"] = nil
		}
	}
	if sel.GetReverted() {
		o["reverted"] = f.reverted
	}
	if sel.GetBlockHash() {
		o["blockHash"] = BytesToHex(f.blockHash)
	}
	if sel.GetBlockNumber() {
		o["blockNumber"] = quantity(f.blockNumber)
	}
	if sel.GetTransactionHash() {
		o["transactionHash"] = BytesToHex(f.transactionHash)
	}
	if sel.GetTransactionIndex() {
		o["transactionIndex"] = quantity(uint64(f.transactionIndex))
	}
	if sel.GetTraceAddress() {
		// JSON numbers; a nil slice must still render as [].
		traceAddress := f.traceAddress
		if traceAddress == nil {
			traceAddress = []uint32{}
		}
		o["traceAddress"] = traceAddress
	}
	return o
}
