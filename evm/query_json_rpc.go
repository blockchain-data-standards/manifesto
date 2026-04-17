package evm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func QueryBlocksRequestFromJsonRpc(params json.RawMessage) (*QueryBlocksRequest, error) {
	var req QueryBlocksRequest
	obj, err := unwrapQueryParams(params)
	if err != nil {
		return nil, err
	}
	if err := applyCommonQueryRequest(obj, &req.FromBlock, &req.ToBlock, &req.Order, &req.Limit, &req.Cursor, &req.ChainId); err != nil {
		return nil, err
	}
	req.BlockFields = blockFieldSelectionFromJsonRpc(fieldGroup(obj, "blocks"))
	return &req, nil
}

func QueryTransactionsRequestFromJsonRpc(params json.RawMessage) (*QueryTransactionsRequest, error) {
	var req QueryTransactionsRequest
	obj, err := unwrapQueryParams(params)
	if err != nil {
		return nil, err
	}
	if err := applyCommonQueryRequest(obj, &req.FromBlock, &req.ToBlock, &req.Order, &req.Limit, &req.Cursor, &req.ChainId); err != nil {
		return nil, err
	}
	req.Filter = transactionFilterFromJsonRpc(obj["filter"])
	req.TransactionFields = transactionFieldSelectionFromJsonRpc(fieldGroup(obj, "transactions"))
	req.BlockFields = blockFieldSelectionFromJsonRpc(fieldGroup(obj, "blocks"))
	return &req, nil
}

func QueryLogsRequestFromJsonRpc(params json.RawMessage) (*QueryLogsRequest, error) {
	var req QueryLogsRequest
	obj, err := unwrapQueryParams(params)
	if err != nil {
		return nil, err
	}
	if err := applyCommonQueryRequest(obj, &req.FromBlock, &req.ToBlock, &req.Order, &req.Limit, &req.Cursor, &req.ChainId); err != nil {
		return nil, err
	}
	req.Filter, err = logFilterFromJsonRpc(obj["filter"])
	if err != nil {
		return nil, err
	}
	req.LogFields = logFieldSelectionFromJsonRpc(fieldGroup(obj, "logs"))
	req.TransactionFields = transactionFieldSelectionFromJsonRpc(fieldGroup(obj, "transactions"))
	req.BlockFields = blockFieldSelectionFromJsonRpc(fieldGroup(obj, "blocks"))
	return &req, nil
}

func QueryTracesRequestFromJsonRpc(params json.RawMessage) (*QueryTracesRequest, error) {
	var req QueryTracesRequest
	obj, err := unwrapQueryParams(params)
	if err != nil {
		return nil, err
	}
	if err := applyCommonQueryRequest(obj, &req.FromBlock, &req.ToBlock, &req.Order, &req.Limit, &req.Cursor, &req.ChainId); err != nil {
		return nil, err
	}
	req.Filter = traceFilterFromJsonRpc(obj["filter"])
	req.TraceFields = traceFieldSelectionFromJsonRpc(fieldGroup(obj, "traces"))
	req.TransactionFields = transactionFieldSelectionFromJsonRpc(fieldGroup(obj, "transactions"))
	req.BlockFields = blockFieldSelectionFromJsonRpc(fieldGroup(obj, "blocks"))
	return &req, nil
}

func QueryTransfersRequestFromJsonRpc(params json.RawMessage) (*QueryTransfersRequest, error) {
	var req QueryTransfersRequest
	obj, err := unwrapQueryParams(params)
	if err != nil {
		return nil, err
	}
	if err := applyCommonQueryRequest(obj, &req.FromBlock, &req.ToBlock, &req.Order, &req.Limit, &req.Cursor, &req.ChainId); err != nil {
		return nil, err
	}
	req.Filter = transferFilterFromJsonRpc(obj["filter"])
	req.TransferFields = transferFieldSelectionFromJsonRpc(fieldGroup(obj, "transfers"))
	req.TransactionFields = transactionFieldSelectionFromJsonRpc(fieldGroup(obj, "transactions"))
	req.BlockFields = blockFieldSelectionFromJsonRpc(fieldGroup(obj, "blocks"))
	return &req, nil
}

func QueryBlocksResponseToJsonRpc(resp *QueryBlocksResponse) map[string]interface{} {
	blocks := make([]interface{}, 0, len(resp.GetBlocks()))
	for _, block := range resp.GetBlocks() {
		blocks = append(blocks, BlockToJsonRpc(block, nil, nil, nil))
	}
	return map[string]interface{}{
		"data":        map[string]interface{}{"blocks": blocks},
		"fromBlock":   cursorBlockToJsonRpc(resp.GetFromBlock()),
		"toBlock":     cursorBlockToJsonRpc(resp.GetToBlock()),
		"cursorBlock": cursorBlockToJsonRpc(resp.GetCursorBlock()),
	}
}

func QueryTransactionsResponseToJsonRpc(resp *QueryTransactionsResponse) map[string]interface{} {
	txs := make([]interface{}, 0, len(resp.GetTransactions()))
	for _, tx := range resp.GetTransactions() {
		txs = append(txs, TransactionToJsonRpc(tx))
	}
	blocks := make([]interface{}, 0, len(resp.GetBlocks()))
	for _, block := range resp.GetBlocks() {
		blocks = append(blocks, BlockToJsonRpc(block, nil, nil, nil))
	}
	return map[string]interface{}{
		"data": map[string]interface{}{
			"transactions": txs,
			"blocks":       blocks,
		},
		"fromBlock":   cursorBlockToJsonRpc(resp.GetFromBlock()),
		"toBlock":     cursorBlockToJsonRpc(resp.GetToBlock()),
		"cursorBlock": cursorBlockToJsonRpc(resp.GetCursorBlock()),
	}
}

func QueryLogsResponseToJsonRpc(resp *QueryLogsResponse) map[string]interface{} {
	logs := make([]interface{}, 0, len(resp.GetLogs()))
	for _, log := range resp.GetLogs() {
		logs = append(logs, LogToJsonRpc(log))
	}
	txs := make([]interface{}, 0, len(resp.GetTransactions()))
	for _, tx := range resp.GetTransactions() {
		txs = append(txs, TransactionToJsonRpc(tx))
	}
	blocks := make([]interface{}, 0, len(resp.GetBlocks()))
	for _, block := range resp.GetBlocks() {
		blocks = append(blocks, BlockToJsonRpc(block, nil, nil, nil))
	}
	return map[string]interface{}{
		"data": map[string]interface{}{
			"logs":         logs,
			"transactions": txs,
			"blocks":       blocks,
		},
		"fromBlock":   cursorBlockToJsonRpc(resp.GetFromBlock()),
		"toBlock":     cursorBlockToJsonRpc(resp.GetToBlock()),
		"cursorBlock": cursorBlockToJsonRpc(resp.GetCursorBlock()),
	}
}

func QueryTracesResponseToJsonRpc(resp *QueryTracesResponse) map[string]interface{} {
	traces := make([]interface{}, 0, len(resp.GetTraces()))
	for _, trace := range resp.GetTraces() {
		traces = append(traces, traceToJsonRpc(trace))
	}
	txs := make([]interface{}, 0, len(resp.GetTransactions()))
	for _, tx := range resp.GetTransactions() {
		txs = append(txs, TransactionToJsonRpc(tx))
	}
	blocks := make([]interface{}, 0, len(resp.GetBlocks()))
	for _, block := range resp.GetBlocks() {
		blocks = append(blocks, BlockToJsonRpc(block, nil, nil, nil))
	}
	return map[string]interface{}{
		"data": map[string]interface{}{
			"traces":       traces,
			"transactions": txs,
			"blocks":       blocks,
		},
		"fromBlock":   cursorBlockToJsonRpc(resp.GetFromBlock()),
		"toBlock":     cursorBlockToJsonRpc(resp.GetToBlock()),
		"cursorBlock": cursorBlockToJsonRpc(resp.GetCursorBlock()),
	}
}

func QueryTransfersResponseToJsonRpc(resp *QueryTransfersResponse) map[string]interface{} {
	transfers := make([]interface{}, 0, len(resp.GetTransfers()))
	for _, transfer := range resp.GetTransfers() {
		transfers = append(transfers, transferToJsonRpc(transfer))
	}
	txs := make([]interface{}, 0, len(resp.GetTransactions()))
	for _, tx := range resp.GetTransactions() {
		txs = append(txs, TransactionToJsonRpc(tx))
	}
	blocks := make([]interface{}, 0, len(resp.GetBlocks()))
	for _, block := range resp.GetBlocks() {
		blocks = append(blocks, BlockToJsonRpc(block, nil, nil, nil))
	}
	return map[string]interface{}{
		"data": map[string]interface{}{
			"transfers":    transfers,
			"transactions": txs,
			"blocks":       blocks,
		},
		"fromBlock":   cursorBlockToJsonRpc(resp.GetFromBlock()),
		"toBlock":     cursorBlockToJsonRpc(resp.GetToBlock()),
		"cursorBlock": cursorBlockToJsonRpc(resp.GetCursorBlock()),
	}
}

func blockFieldSelectionFromJsonRpc(v interface{}) *BlockFieldSelection {
	var sel BlockFieldSelection
	if !applyFieldSelection(v, &sel) {
		return nil
	}
	return &sel
}

func transactionFieldSelectionFromJsonRpc(v interface{}) *TransactionFieldSelection {
	var sel TransactionFieldSelection
	if !applyFieldSelection(v, &sel) {
		return nil
	}
	return &sel
}

func logFieldSelectionFromJsonRpc(v interface{}) *LogFieldSelection {
	var sel LogFieldSelection
	if !applyFieldSelection(v, &sel) {
		return nil
	}
	return &sel
}

func traceFieldSelectionFromJsonRpc(v interface{}) *TraceFieldSelection {
	var sel TraceFieldSelection
	if !applyFieldSelection(v, &sel) {
		return nil
	}
	return &sel
}

func transferFieldSelectionFromJsonRpc(v interface{}) *TransferFieldSelection {
	var sel TransferFieldSelection
	if !applyFieldSelection(v, &sel) {
		return nil
	}
	return &sel
}

func unwrapQueryParams(params json.RawMessage) (map[string]interface{}, error) {
	if len(params) == 0 || string(params) == "null" {
		return map[string]interface{}{}, nil
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(params, &obj); err == nil {
		return obj, nil
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(params, &arr); err == nil {
		if len(arr) == 0 {
			return map[string]interface{}{}, nil
		}
		return arr[0], nil
	}

	var mixed []interface{}
	if err := json.Unmarshal(params, &mixed); err != nil {
		return nil, err
	}
	if len(mixed) == 0 {
		return map[string]interface{}{}, nil
	}
	first, ok := mixed[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid query params")
	}
	return first, nil
}

func applyCommonQueryRequest(
	obj map[string]interface{},
	fromBlock **string,
	toBlock **string,
	order **SortOrder,
	limit **uint32,
	cursor **CursorBlock,
	chainID **uint64,
) error {
	if v, ok := obj["fromBlock"].(string); ok && v != "" {
		*fromBlock = StringPtr(v)
	}
	if v, ok := obj["toBlock"].(string); ok && v != "" {
		*toBlock = StringPtr(v)
	}
	if v, ok := obj["order"].(string); ok && v != "" {
		switch strings.ToLower(v) {
		case "asc":
			o := SortOrder_ASC
			*order = &o
		case "desc":
			o := SortOrder_DESC
			*order = &o
		default:
			return fmt.Errorf("invalid order: %s", v)
		}
	}
	if raw, ok := obj["limit"]; ok {
		parsed, err := parseUint32Value(raw)
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		*limit = &parsed
	}
	if raw, ok := obj["chainId"]; ok {
		parsed, err := parseUint64Value(raw)
		if err != nil {
			return fmt.Errorf("invalid chainId: %w", err)
		}
		*chainID = &parsed
	}
	if raw, ok := obj["cursor"]; ok {
		cur, err := parseCursorBlock(raw)
		if err != nil {
			return err
		}
		*cursor = cur
	} else if raw, ok := obj["cursorBlock"]; ok {
		cur, err := parseCursorBlock(raw)
		if err != nil {
			return err
		}
		*cursor = cur
	}
	return nil
}

func parseCursorBlock(raw interface{}) (*CursorBlock, error) {
	if raw == nil {
		return nil, nil
	}
	obj, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid cursor block")
	}
	cur := &CursorBlock{}
	if v, ok := obj["number"]; ok {
		n, err := parseUint64Value(v)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor number: %w", err)
		}
		cur.Number = n
	}
	if v, ok := obj["hash"].(string); ok && v != "" {
		b, err := HexToBytes(v)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor hash: %w", err)
		}
		cur.Hash = b
	}
	if v, ok := obj["parentHash"].(string); ok && v != "" {
		b, err := HexToBytes(v)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor parentHash: %w", err)
		}
		cur.ParentHash = b
	}
	return cur, nil
}

func transactionFilterFromJsonRpc(raw interface{}) *TransactionFilter {
	obj, _ := raw.(map[string]interface{})
	if obj == nil {
		return nil
	}
	return &TransactionFilter{
		From:     byteListFromJson(obj["from"]),
		To:       byteListFromJson(obj["to"]),
		Selector: byteListFromJson(obj["selector"]),
	}
}

func logFilterFromJsonRpc(raw interface{}) (*LogFilter, error) {
	obj, _ := raw.(map[string]interface{})
	if obj == nil {
		return nil, nil
	}
	filter := &LogFilter{
		Address: byteListFromJson(obj["address"]),
	}
	if rawTopics, ok := obj["topics"].([]interface{}); ok {
		filter.Topics = make([]*TopicFilter, 0, len(rawTopics))
		for _, rawTopic := range rawTopics {
			values := byteListFromJson(rawTopic)
			filter.Topics = append(filter.Topics, &TopicFilter{Values: values})
		}
	}
	return filter, nil
}

func traceFilterFromJsonRpc(raw interface{}) *TraceFilter {
	obj, _ := raw.(map[string]interface{})
	if obj == nil {
		return nil
	}
	filter := &TraceFilter{
		From:     byteListFromJson(obj["from"]),
		To:       byteListFromJson(obj["to"]),
		Selector: byteListFromJson(obj["selector"]),
	}
	if v, ok := obj["isTopLevel"].(bool); ok {
		filter.IsTopLevel = BoolPtr(v)
	}
	return filter
}

func transferFilterFromJsonRpc(raw interface{}) *TransferFilter {
	obj, _ := raw.(map[string]interface{})
	if obj == nil {
		return nil
	}
	filter := &TransferFilter{
		From: byteListFromJson(obj["from"]),
		To:   byteListFromJson(obj["to"]),
	}
	if v, ok := obj["isTopLevel"].(bool); ok {
		filter.IsTopLevel = BoolPtr(v)
	}
	return filter
}

func fieldGroup(obj map[string]interface{}, key string) interface{} {
	fields, ok := obj["fields"].(map[string]interface{})
	if !ok {
		return nil
	}
	return fields[key]
}

func applyFieldSelection(raw interface{}, target interface{}) bool {
	if raw == nil {
		return false
	}
	if b, ok := raw.(bool); ok && b {
		return false
	}
	names, ok := raw.([]interface{})
	if !ok {
		return false
	}
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return false
	}
	structVal := rv.Elem()
	structType := structVal.Type()
	setCount := 0
	for _, item := range names {
		name, ok := item.(string)
		if !ok {
			continue
		}
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			if strings.EqualFold(field.Name, name) {
				structVal.Field(i).SetBool(true)
				setCount++
				break
			}
		}
	}
	return setCount > 0
}

func parseUint32Value(raw interface{}) (uint32, error) {
	switch v := raw.(type) {
	case string:
		return NumberishToUint32(v)
	case float64:
		return uint32(v), nil
	default:
		return 0, fmt.Errorf("unsupported numeric value")
	}
}

func parseUint64Value(raw interface{}) (uint64, error) {
	switch v := raw.(type) {
	case string:
		return NumberishToUint64(v)
	case float64:
		return uint64(v), nil
	default:
		return 0, fmt.Errorf("unsupported numeric value")
	}
}

func byteListFromJson(raw interface{}) [][]byte {
	if raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case string:
		b, err := HexToBytes(v)
		if err != nil {
			return nil
		}
		return [][]byte{b}
	case []interface{}:
		out := make([][]byte, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				continue
			}
			b, err := HexToBytes(s)
			if err != nil {
				continue
			}
			out = append(out, b)
		}
		return out
	default:
		return nil
	}
}

func cursorBlockToJsonRpc(cur *CursorBlock) interface{} {
	if cur == nil {
		return nil
	}
	return map[string]interface{}{
		"number":     fmt.Sprintf("0x%x", cur.Number),
		"hash":       BytesToHex(cur.Hash),
		"parentHash": BytesToHex(cur.ParentHash),
	}
}

func traceToJsonRpc(trace *Trace) map[string]interface{} {
	if trace == nil {
		return nil
	}
	traceAddress := make([]string, 0, len(trace.TraceAddress))
	for _, v := range trace.TraceAddress {
		traceAddress = append(traceAddress, fmt.Sprintf("0x%x", v))
	}
	out := map[string]interface{}{
		"traceType":        strings.ToLower(strings.TrimPrefix(trace.TraceType.String(), "TRACE_")),
		"callType":         strings.ToLower(strings.TrimPrefix(trace.CallType.String(), "TRACE_CALL_")),
		"from":             BytesToHex(trace.From),
		"value":            trace.Value,
		"input":            BytesToHex(trace.Input),
		"output":           BytesToHex(trace.Output),
		"gas":              fmt.Sprintf("0x%x", trace.Gas),
		"gasUsed":          fmt.Sprintf("0x%x", trace.GasUsed),
		"subtraces":        fmt.Sprintf("0x%x", trace.Subtraces),
		"traceAddress":     traceAddress,
		"transactionHash":  BytesToHex(trace.TransactionHash),
		"transactionIndex": fmt.Sprintf("0x%x", trace.TransactionIndex),
		"blockNumber":      fmt.Sprintf("0x%x", trace.BlockNumber),
		"blockHash":        BytesToHex(trace.BlockHash),
	}
	if len(trace.To) > 0 {
		out["to"] = BytesToHex(trace.To)
	} else {
		out["to"] = nil
	}
	if trace.Error != nil {
		out["error"] = *trace.Error
	}
	if trace.BlockTimestamp != nil {
		out["blockTimestamp"] = fmt.Sprintf("0x%x", *trace.BlockTimestamp)
	}
	return out
}

func transferToJsonRpc(transfer *NativeTransfer) map[string]interface{} {
	if transfer == nil {
		return nil
	}
	traceAddress := make([]string, 0, len(transfer.TraceAddress))
	for _, v := range transfer.TraceAddress {
		traceAddress = append(traceAddress, fmt.Sprintf("0x%x", v))
	}
	out := map[string]interface{}{
		"from":             BytesToHex(transfer.From),
		"to":               BytesToHex(transfer.To),
		"value":            transfer.Value,
		"transactionHash":  BytesToHex(transfer.TransactionHash),
		"transactionIndex": fmt.Sprintf("0x%x", transfer.TransactionIndex),
		"blockNumber":      fmt.Sprintf("0x%x", transfer.BlockNumber),
		"blockHash":        BytesToHex(transfer.BlockHash),
		"traceAddress":     traceAddress,
	}
	if transfer.BlockTimestamp != nil {
		out["blockTimestamp"] = fmt.Sprintf("0x%x", *transfer.BlockTimestamp)
	}
	return out
}
