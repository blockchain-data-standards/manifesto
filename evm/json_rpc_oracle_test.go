package evm

// The map converters these writers replaced, verbatim from manifesto
// aea005ce6ae7 with an oracle prefix: the fuzz oracle for the writers.

import "fmt"

// oracleLogToJsonRpc serialises a *Log into a JSON-RPC compatible map[string]interface{}.
func oracleLogToJsonRpc(l *Log) map[string]interface{} {
	if l == nil {
		return nil
	}
	topics := make([]string, len(l.Topics))
	for i, t := range l.Topics {
		topics[i] = BytesToHex(t)
	}

	result := map[string]interface{}{
		"address":          BytesToHex(l.Address),
		"topics":           topics,
		"data":             BytesToHex(l.Data),
		"blockNumber":      fmt.Sprintf("0x%x", l.BlockNumber),
		"transactionHash":  BytesToHex(l.TransactionHash),
		"transactionIndex": fmt.Sprintf("0x%x", l.TransactionIndex),
		"blockHash":        BytesToHex(l.BlockHash),
		"logIndex":         fmt.Sprintf("0x%x", l.LogIndex),
		"removed":          false,
	}

	// Include optional block timestamp when present
	if l.BlockTimestamp != nil {
		result["blockTimestamp"] = fmt.Sprintf("0x%x", *l.BlockTimestamp)
	}

	return result
}

// oracleLogsToJsonRpc converts a slice of *Log to []interface{} for JSON-RPC responses.
func oracleLogsToJsonRpc(logs []*Log) []interface{} {
	out := make([]interface{}, len(logs))
	for i, l := range logs {
		out[i] = oracleLogToJsonRpc(l)
	}
	return out
}

// oracleTransactionToJsonRpc converts a *Transaction into the JSON-RPC transaction
// object a node sends for eth_getTransactionByHash and full blocks.
// sig selects the r/s encoding; pass SignatureEncodingForChain(chainId).
//
// A BDS Transaction also carries its receipt's execution figures (gasUsed,
// effectiveGasPrice, blob gas, L2 fees) and optional fields a chain may not
// have. Nodes send those figures on the receipt only, and omit a field the
// transaction does not have rather than sending null, so this leaves them
// out: a value the record carries is never invented or nulled.
// oracleTransactionRecordToJsonRpc renders every field the record holds.
//
// The shape follows what providers send (2026-10-08 survey: 22 chains, 2-8
// providers each): an empty accessList only on the types that define one, no
// chainId on an OP deposit (0x7e) or a pre-EIP-155 legacy transaction.
func oracleTransactionToJsonRpc(tx *Transaction, sig SignatureEncoding) map[string]interface{} {
	o := oracleTransactionRecordToJsonRpc(tx, sig)
	if o == nil {
		return nil
	}
	for k, v := range o {
		// "to" is null on a contract creation, and nodes send it that way.
		if v == nil && k != "to" {
			delete(o, k)
		}
	}
	for _, k := range oracleReceiptFigures {
		delete(o, k)
	}
	if len(tx.AccessList) == 0 && !oracleAccessListTypes[tx.Type] {
		// Legacy, deposit (0x7e), Arbitrum (0x6a), zkSync (0x71) and other
		// types without one: nodes send no accessList key at all.
		delete(o, "accessList")
	}
	switch tx.Type {
	case 0:
		// A legacy transaction names a chainId only when its signature does
		// (EIP-155: v = chainId*2 + 35 or 36, else 27 or 28).
		if v, ok := o["v"].(string); ok {
			if n, err := HexToUint64(v); err == nil && n < 35 {
				delete(o, "chainId")
			}
		}
	case 0x7e:
		// An OP deposit is not signed by its sender and names no chainId.
		delete(o, "chainId")
	}
	return o
}

// oracleAccessListTypes are the transaction types that define an access list, so
// nodes send one even when it is empty: EIP-2930 (1), EIP-1559 (2), EIP-4844
// (3), EIP-7702 (4) and Celo's CIP-64 (0x7b).
var oracleAccessListTypes = map[uint32]bool{1: true, 2: true, 3: true, 4: true, 0x7b: true}

// oracleReceiptFigures are the execution results a BDS Transaction
// copies from its receipt. Nodes send them on the receipt only.
var oracleReceiptFigures = []string{
	"gasUsed", "effectiveGasPrice", "blobGasUsed", "blobGasPrice",
	"l1Fee", "l1GasUsed", "l1GasPrice", "l1FeeScalar", "l1BlobBaseFee", "l1BlobBaseFeeScalar",
}

// oracleTransactionRecordToJsonRpc renders every field a BDS Transaction holds,
// including its receipt's execution figures. The eth_query* methods do not use
// it: they render a MIP-16 field selection (queryTransactionToJsonRpc).
// sig selects the r/s encoding; pass SignatureEncodingForChain(chainId).
func oracleTransactionRecordToJsonRpc(tx *Transaction, sig SignatureEncoding) map[string]interface{} {
	if tx == nil {
		return nil
	}

	o := map[string]interface{}{
		"hash":  BytesToHex(tx.Hash),
		"nonce": fmt.Sprintf("0x%x", tx.Nonce),
		"from":  BytesToHex(tx.From),
		"gas":   fmt.Sprintf("0x%x", tx.GasLimit),
		"input": BytesToHex(tx.Input),
		"type":  fmt.Sprintf("0x%x", tx.Type),
	}

	if len(tx.To) > 0 {
		o["to"] = BytesToHex(tx.To)
	} else {
		o["to"] = nil
	}

	if tx.BlockHash != nil {
		o["blockHash"] = BytesToHex(tx.BlockHash)
	}
	if tx.BlockNumber != nil {
		o["blockNumber"] = fmt.Sprintf("0x%x", *tx.BlockNumber)
	}
	if tx.TransactionIndex != nil {
		o["transactionIndex"] = fmt.Sprintf("0x%x", *tx.TransactionIndex)
	}
	if tx.BlockTimestamp != nil {
		o["blockTimestamp"] = fmt.Sprintf("0x%x", *tx.BlockTimestamp)
	}

	// Value (decimal -> hex)
	if tx.Value != "" {
		if hex, err := DecimalStringToHex(tx.Value); err == nil {
			o["value"] = hex
		}
	}

	// L2 mint value (decimal -> hex)
	if tx.Mint != nil {
		if hex, err := DecimalStringToHex(*tx.Mint); err == nil {
			o["mint"] = hex
		}
	}

	// Fee related fields – pointers to decimal strings
	if tx.GasPrice != nil {
		if hex, err := DecimalStringToHex(*tx.GasPrice); err == nil {
			o["gasPrice"] = hex
		}
	}
	if tx.MaxFeePerGas != nil {
		if hex, err := DecimalStringToHex(*tx.MaxFeePerGas); err == nil {
			o["maxFeePerGas"] = hex
		}
	}
	if tx.MaxPriorityFeePerGas != nil {
		if hex, err := DecimalStringToHex(*tx.MaxPriorityFeePerGas); err == nil {
			o["maxPriorityFeePerGas"] = hex
		}
	}

	// Execution result fields (only available for mined transactions)
	if tx.GasUsed != nil {
		o["gasUsed"] = fmt.Sprintf("0x%x", *tx.GasUsed)
	}
	if tx.EffectiveGasPrice != nil {
		if hex, err := DecimalStringToHex(*tx.EffectiveGasPrice); err == nil {
			o["effectiveGasPrice"] = hex
		}
	}

	if tx.R != nil {
		o["r"] = sig.hex(tx.R)
	}
	if tx.S != nil {
		o["s"] = sig.hex(tx.S)
	}
	if tx.V != nil {
		// v is QUANTITY
		o["v"] = BytesToQuantityHex(tx.V)
	}

	// Source hash (bytes)
	if len(tx.SourceHash) > 0 {
		o["sourceHash"] = BytesToHex(tx.SourceHash)
	}

	// Chain ID (optional)
	if tx.ChainId != nil {
		o["chainId"] = fmt.Sprintf("0x%x", *tx.ChainId)
	} else {
		o["chainId"] = nil
	}

	// yParity (optional)
	if tx.YParity != nil {
		o["yParity"] = fmt.Sprintf("0x%x", *tx.YParity)
	} else {
		o["yParity"] = nil
	}

	// Access list (EIP-2930)
	accessList := make([]interface{}, 0, len(tx.AccessList))
	for _, item := range tx.AccessList {
		if item == nil {
			continue
		}
		obj := map[string]interface{}{
			"address": BytesToHex(item.Address),
		}
		if len(item.StorageKeys) > 0 {
			keys := make([]string, len(item.StorageKeys))
			for i, k := range item.StorageKeys {
				keys[i] = BytesToHex(k)
			}
			obj["storageKeys"] = keys
		} else {
			obj["storageKeys"] = []interface{}{}
		}
		accessList = append(accessList, obj)
	}
	o["accessList"] = accessList

	// EIP-4844 fields
	if tx.MaxFeePerBlobGas != nil {
		if hex, err := DecimalStringToHex(*tx.MaxFeePerBlobGas); err == nil {
			o["maxFeePerBlobGas"] = hex
		}
	}
	if len(tx.BlobVersionedHashes) > 0 {
		hashes := make([]string, len(tx.BlobVersionedHashes))
		for i, h := range tx.BlobVersionedHashes {
			hashes[i] = BytesToHex(h)
		}
		o["blobVersionedHashes"] = hashes
	}
	if tx.BlobGasUsed != nil {
		o["blobGasUsed"] = fmt.Sprintf("0x%x", *tx.BlobGasUsed)
	}
	if tx.BlobGasPrice != nil {
		if hex, err := DecimalStringToHex(*tx.BlobGasPrice); err == nil {
			o["blobGasPrice"] = hex
		}
	}

	// EIP-7702 authorization list
	if len(tx.AuthorizationList) > 0 {
		authList := make([]interface{}, 0, len(tx.AuthorizationList))
		for _, auth := range tx.AuthorizationList {
			if auth == nil {
				continue
			}
			authItem := map[string]interface{}{
				"chainId": fmt.Sprintf("0x%x", auth.ChainId),
				"address": BytesToHex(auth.Address),
				"nonce":   fmt.Sprintf("0x%x", auth.Nonce),
				"r":       sig.hex(auth.R),
				"s":       sig.hex(auth.S),
				"yParity": fmt.Sprintf("0x%x", auth.YParity),
			}
			// Optional authority (bytes) – include when present
			if len(auth.Authority) > 0 {
				authItem["authority"] = BytesToHex(auth.Authority)
			}
			authList = append(authList, authItem)
		}
		o["authorizationList"] = authList
	}

	// L2 fee breakdown & miscellaneous fields
	if tx.L1Fee != nil {
		if hex, err := DecimalStringToHex(*tx.L1Fee); err == nil {
			o["l1Fee"] = hex
		} else {
			o["l1Fee"] = nil
		}
	} else {
		o["l1Fee"] = nil
	}
	if tx.L1GasUsed != nil {
		if hex, err := DecimalStringToHex(*tx.L1GasUsed); err == nil {
			o["l1GasUsed"] = hex
		}
	}
	if tx.L1GasPrice != nil {
		if hex, err := DecimalStringToHex(*tx.L1GasPrice); err == nil {
			o["l1GasPrice"] = hex
		}
	}
	if tx.L1FeeScalar != nil {
		o["l1FeeScalar"] = *tx.L1FeeScalar
	}
	if tx.L1BlobBaseFee != nil {
		if hex, err := DecimalStringToHex(*tx.L1BlobBaseFee); err == nil {
			o["l1BlobBaseFee"] = hex
		}
	}
	if tx.L1BlobBaseFeeScalar != nil {
		o["l1BlobBaseFeeScalar"] = fmt.Sprintf("0x%x", *tx.L1BlobBaseFeeScalar)
	}
	if tx.GatewayFee != nil {
		if hex, err := DecimalStringToHex(*tx.GatewayFee); err == nil {
			o["gatewayFee"] = hex
		}
	}
	if len(tx.FeeCurrency) > 0 {
		o["feeCurrency"] = BytesToHex(tx.FeeCurrency)
	}
	if len(tx.GatewayFeeRecipient) > 0 {
		o["gatewayFeeRecipient"] = BytesToHex(tx.GatewayFeeRecipient)
	}

	// Arbitrum retryable ticket fields
	if len(tx.Beneficiary) > 0 {
		o["beneficiary"] = BytesToHex(tx.Beneficiary)
	}
	if tx.DepositValue != nil {
		if hex, err := DecimalStringToHex(*tx.DepositValue); err == nil {
			o["depositValue"] = hex
		}
	}
	if tx.L1BaseFee != nil {
		if hex, err := DecimalStringToHex(*tx.L1BaseFee); err == nil {
			o["l1BaseFee"] = hex
		}
	}
	if tx.MaxSubmissionFee != nil {
		if hex, err := DecimalStringToHex(*tx.MaxSubmissionFee); err == nil {
			o["maxSubmissionFee"] = hex
		}
	}
	if len(tx.RefundTo) > 0 {
		o["refundTo"] = BytesToHex(tx.RefundTo)
	}
	if len(tx.RequestId) > 0 {
		o["requestId"] = BytesToHex(tx.RequestId)
	}
	if len(tx.RetryData) > 0 {
		o["retryData"] = BytesToHex(tx.RetryData)
	}
	if len(tx.RetryTo) > 0 {
		o["retryTo"] = BytesToHex(tx.RetryTo)
	}
	if tx.RetryValue != nil {
		if hex, err := DecimalStringToHex(*tx.RetryValue); err == nil {
			o["retryValue"] = hex
		}
	}
	if tx.MaxRefund != nil {
		if hex, err := DecimalStringToHex(*tx.MaxRefund); err == nil {
			o["maxRefund"] = hex
		}
	}
	if tx.SubmissionFeeRefund != nil {
		if hex, err := DecimalStringToHex(*tx.SubmissionFeeRefund); err == nil {
			o["submissionFeeRefund"] = hex
		}
	}
	if len(tx.TicketId) > 0 {
		o["ticketId"] = BytesToHex(tx.TicketId)
	}

	// Base-specific fields
	if tx.IsSystemTx != nil {
		o["isSystemTx"] = *tx.IsSystemTx
	}
	if tx.DepositReceiptVersion != nil {
		if hex, err := DecimalStringToHex(*tx.DepositReceiptVersion); err == nil {
			o["depositReceiptVersion"] = hex
		}
	}

	return o
}

// oracleReceiptsToJsonRpc converts a slice of *Receipt to []interface{} for JSON-RPC responses.
func oracleReceiptsToJsonRpc(receipts []*Receipt) []interface{} {
	out := make([]interface{}, len(receipts))
	for i, r := range receipts {
		out[i] = oracleReceiptToJsonRpc(r)
	}
	return out
}

// oracleReceiptToJsonRpc converts a *Receipt into JSON-RPC representation.
func oracleReceiptToJsonRpc(r *Receipt) map[string]interface{} {
	if r == nil {
		return nil
	}

	out := map[string]interface{}{
		"transactionHash":   BytesToHex(r.TransactionHash),
		"transactionIndex":  fmt.Sprintf("0x%x", r.TransactionIndex),
		"blockHash":         BytesToHex(r.BlockHash),
		"blockNumber":       fmt.Sprintf("0x%x", r.BlockNumber),
		"from":              BytesToHex(r.From),
		"cumulativeGasUsed": fmt.Sprintf("0x%x", r.CumulativeGasUsed),
		"gasUsed":           fmt.Sprintf("0x%x", r.GasUsed),
		"logsBloom":         BytesToHex(r.LogsBloom),
		"logs":              oracleLogsToJsonRpc(r.Logs),
		"contractAddress":   nil,
	}

	// Add "to" field - can be null for contract creation
	if len(r.To) > 0 {
		out["to"] = BytesToHex(r.To)
	} else {
		out["to"] = nil
	}

	if r.Status != nil {
		out["status"] = fmt.Sprintf("0x%x", *r.Status)
	}
	if len(r.ContractAddress) > 0 {
		out["contractAddress"] = BytesToHex(r.ContractAddress)
	}
	out["type"] = fmt.Sprintf("0x%x", r.Type)

	if r.EffectiveGasPrice != "" {
		if hex, err := DecimalStringToHex(r.EffectiveGasPrice); err == nil {
			out["effectiveGasPrice"] = hex
		}
	}

	// Add root field for pre-Byzantium compatibility
	if len(r.Root) > 0 {
		out["root"] = BytesToHex(r.Root)
	}

	// No blockTimestamp: no provider sends one on a receipt (it is on the
	// receipt's logs instead).

	if r.GasUsedForL1 != nil {
		out["gasUsedForL1"] = fmt.Sprintf("0x%x", *r.GasUsedForL1)
	}

	if r.L1BlockNumber != nil {
		out["l1BlockNumber"] = fmt.Sprintf("0x%x", *r.L1BlockNumber)
	}

	// Additional L2 fee breakdown fields, present only on chains that have
	// them: a node elsewhere omits them rather than sending null.
	if r.L1Fee != nil {
		if hex, err := DecimalStringToHex(*r.L1Fee); err == nil {
			out["l1Fee"] = hex
		}
	}
	if r.L1GasUsed != nil {
		if hex, err := DecimalStringToHex(*r.L1GasUsed); err == nil {
			out["l1GasUsed"] = hex
		}
	}
	if r.L1GasPrice != nil {
		if hex, err := DecimalStringToHex(*r.L1GasPrice); err == nil {
			out["l1GasPrice"] = hex
		}
	}
	if r.GatewayFee != nil {
		if hex, err := DecimalStringToHex(*r.GatewayFee); err == nil {
			out["gatewayFee"] = hex
		}
	}

	if r.BlobGasUsed != nil {
		out["blobGasUsed"] = fmt.Sprintf("0x%x", *r.BlobGasUsed)
	}
	if r.BlobGasPrice != nil {
		if hex, err := DecimalStringToHex(*r.BlobGasPrice); err == nil {
			out["blobGasPrice"] = hex
		}
	}
	if r.L1FeeScalar != nil {
		out["l1FeeScalar"] = *r.L1FeeScalar
	}
	if r.L1BaseFeeScalar != nil {
		out["l1BaseFeeScalar"] = fmt.Sprintf("0x%x", *r.L1BaseFeeScalar)
	}
	if r.L1BlobBaseFee != nil {
		if hex, err := DecimalStringToHex(*r.L1BlobBaseFee); err == nil {
			out["l1BlobBaseFee"] = hex
		}
	}
	if r.L1BlobBaseFeeScalar != nil {
		out["l1BlobBaseFeeScalar"] = fmt.Sprintf("0x%x", *r.L1BlobBaseFeeScalar)
	}
	if r.DaFootprintGasScalar != nil {
		out["daFootprintGasScalar"] = fmt.Sprintf("0x%x", *r.DaFootprintGasScalar)
	}
	if r.DepositNonce != nil {
		if hex, err := DecimalStringToHex(*r.DepositNonce); err == nil {
			out["depositNonce"] = hex
		}
	}
	if r.DepositReceiptVersion != nil {
		if hex, err := DecimalStringToHex(*r.DepositReceiptVersion); err == nil {
			out["depositReceiptVersion"] = hex
		}
	}
	if r.Timeboosted != nil {
		out["timeboosted"] = *r.Timeboosted
	}

	return out
}

// oracleWithdrawalToJsonRpc converts a *Withdrawal into JSON-RPC representation.
func oracleWithdrawalToJsonRpc(w *Withdrawal) map[string]interface{} {
	if w == nil {
		return nil
	}

	return map[string]interface{}{
		"index":          fmt.Sprintf("0x%x", w.Index),
		"validatorIndex": fmt.Sprintf("0x%x", w.ValidatorIndex),
		"address":        BytesToHex(w.Address),
		"amount":         fmt.Sprintf("0x%x", w.Amount),
	}
}

// oracleWithdrawalsToJsonRpc converts a slice of *Withdrawal to []interface{} for JSON-RPC responses.
func oracleWithdrawalsToJsonRpc(withdrawals []*Withdrawal) []interface{} {
	out := make([]interface{}, len(withdrawals))
	for i, w := range withdrawals {
		out[i] = oracleWithdrawalToJsonRpc(w)
	}
	return out
}

// oracleBlockToJsonRpc converts an EVM BlockHeader plus optional transactions to JSON-RPC format.
// If fullTxs is supplied it will be used; otherwise transaction hashes are used.
// If withdrawals are supplied, they will be included in the response.
// sig selects the r/s encoding of full transactions; pass SignatureEncodingForChain(chainId).
func oracleBlockToJsonRpc(header *BlockHeader, txHashes [][]byte, fullTxs []*Transaction, withdrawals []*Withdrawal, sig SignatureEncoding) map[string]interface{} {
	if header == nil {
		return nil
	}

	res := map[string]interface{}{
		"number":           fmt.Sprintf("0x%x", header.Number),
		"hash":             BytesToHex(header.Hash),
		"parentHash":       BytesToHex(header.ParentHash),
		"sha3Uncles":       BytesToHex(header.Sha3Uncles),
		"logsBloom":        BytesToHex(header.LogsBloom),
		"transactionsRoot": BytesToHex(header.TransactionsRoot),
		"stateRoot":        BytesToHex(header.StateRoot),
		"receiptsRoot":     BytesToHex(header.ReceiptsRoot),
		"miner":            BytesToHex(header.Miner),
		"extraData":        BytesToHex(header.ExtraData),
		"size":             fmt.Sprintf("0x%x", header.Size),
		"gasLimit":         fmt.Sprintf("0x%x", header.GasLimit),
		"gasUsed":          fmt.Sprintf("0x%x", header.GasUsed),
		"timestamp":        fmt.Sprintf("0x%x", header.Timestamp),
	}

	if header.Nonce != nil {
		res["nonce"] = fmt.Sprintf("0x%016x", *header.Nonce)
	}
	if header.BaseFeePerGas != nil {
		if hex, err := DecimalStringToHex(*header.BaseFeePerGas); err == nil {
			res["baseFeePerGas"] = hex
		}
	}
	if header.Difficulty != nil {
		if hex, err := DecimalStringToHex(*header.Difficulty); err == nil {
			res["difficulty"] = hex
		}
	}
	if header.TotalDifficulty != nil {
		if hex, err := DecimalStringToHex(*header.TotalDifficulty); err == nil {
			res["totalDifficulty"] = hex
		}
	}
	if header.MixHash != nil {
		res["mixHash"] = BytesToHex(header.MixHash)
	}
	if header.WithdrawalsRoot != nil {
		res["withdrawalsRoot"] = BytesToHex(header.WithdrawalsRoot)
	}
	if header.RequestsHash != nil {
		res["requestsHash"] = BytesToHex(header.RequestsHash)
	}
	if header.BlobGasUsed != nil {
		res["blobGasUsed"] = fmt.Sprintf("0x%x", *header.BlobGasUsed)
	}
	if header.ExcessBlobGas != nil {
		res["excessBlobGas"] = fmt.Sprintf("0x%x", *header.ExcessBlobGas)
	}
	if header.ParentBeaconBlockRoot != nil {
		res["parentBeaconBlockRoot"] = BytesToHex(header.ParentBeaconBlockRoot)
	}
	if header.L1BlockNumber != nil {
		res["l1BlockNumber"] = fmt.Sprintf("0x%x", *header.L1BlockNumber)
	}
	if header.SendCount != nil {
		res["sendCount"] = fmt.Sprintf("0x%x", *header.SendCount)
	}
	if header.SendRoot != nil {
		res["sendRoot"] = BytesToHex(header.SendRoot)
	}
	if header.Epoch != nil {
		res["epoch"] = fmt.Sprintf("0x%x", *header.Epoch)
	}
	if header.Slot != nil {
		res["slot"] = fmt.Sprintf("0x%x", *header.Slot)
	}
	if header.ProposerIndex != nil {
		res["proposerIndex"] = fmt.Sprintf("0x%x", *header.ProposerIndex)
	}
	if header.TransactionCount != nil {
		res["transactionCount"] = fmt.Sprintf("0x%x", *header.TransactionCount)
	}
	if header.ProposerPublicKey != nil {
		res["proposerPublicKey"] = *header.ProposerPublicKey
	}
	// A header with withdrawalsRoot always carries a withdrawals list, which
	// proto decodes as nil when empty; JSON-RPC nodes return [] for it.
	if withdrawals != nil || header.WithdrawalsRoot != nil {
		res["withdrawals"] = oracleWithdrawalsToJsonRpc(withdrawals)
	}
	if header.CanonicalRlp != nil {
		res["canonicalRlp"] = BytesToHex(header.CanonicalRlp)
	}

	// Uncles
	if len(header.Uncles) > 0 {
		uncles := make([]string, len(header.Uncles))
		for i, u := range header.Uncles {
			uncles[i] = BytesToHex(u)
		}
		res["uncles"] = uncles
	} else {
		res["uncles"] = []interface{}{}
	}

	// Transactions
	switch {
	case len(fullTxs) > 0:
		txs := make([]interface{}, len(fullTxs))
		for i, t := range fullTxs {
			txs[i] = oracleTransactionToJsonRpc(t, sig)
		}
		res["transactions"] = txs
	case len(txHashes) > 0:
		hashes := make([]string, len(txHashes))
		for i, h := range txHashes {
			hashes[i] = BytesToHex(h)
		}
		res["transactions"] = hashes
	default:
		res["transactions"] = []interface{}{}
	}

	return res
}
