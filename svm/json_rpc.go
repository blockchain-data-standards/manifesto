package svm

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"strconv"
)

// Mapping helpers from bds.svm protobuf messages to Solana JSON-RPC wire
// shapes, mirroring evm/json_rpc.go.
//
// Scope: these render the `encoding: "json"` form — instructions stay compiled
// (`programIdIndex` / `accounts` / base58 `data`). `jsonParsed` is deliberately
// NOT produced here: resolving instructions into `{program, parsed:{type,info}}`
// requires per-program knowledge (SPL Token, System, Stake, Vote, ATA, Memo, …)
// that lives above the data layer. A serving layer that must answer
// `jsonParsed` has to run those parsers itself.
//
// Null-vs-empty fidelity: Agave distinguishes `null` from `[]` on several meta
// fields for pre-recording history. The proto carries that as `*None` booleans,
// and the helpers below honour them — a `*None` flag emits a JSON null while an
// empty-but-present list emits `[]`.

// BlockToJsonRpc converts a ConfirmedBlock to the Solana `getBlock` result
// shape. signatures is the flat list carried by GetBlockResponse when
// transactionDetails is SIGNATURES (the block itself then holds no
// transactions); pass nil otherwise.
//
// Returns nil for a nil block, which is the correct result body for a skipped
// slot (SlotStatus SLOT_SKIPPED) — callers translate that to Agave's -32007 /
// -32009 as appropriate for the requested commitment.
//
// `rewards` is emitted whenever the proto carries them, matching Agave's
// default of `rewards: true`. A caller honouring `excludeRewards` should drop
// the key.
func BlockToJsonRpc(block *ConfirmedBlock, signatures [][]byte) map[string]interface{} {
	if block == nil {
		return nil
	}

	res := map[string]interface{}{
		"blockhash":         Base58Encode(block.Blockhash),
		"previousBlockhash": Base58Encode(block.PreviousBlockhash),
		"parentSlot":        block.ParentSlot,
	}

	// Agave always emits these two keys, using null when unknown.
	if block.BlockHeight != nil {
		res["blockHeight"] = *block.BlockHeight
	} else {
		res["blockHeight"] = nil
	}
	if block.BlockTime != nil {
		res["blockTime"] = *block.BlockTime
	} else {
		res["blockTime"] = nil
	}

	switch {
	case signatures != nil:
		sigs := make([]interface{}, 0, len(signatures))
		for _, s := range signatures {
			sigs = append(sigs, Base58Encode(s))
		}
		res["signatures"] = sigs
	case block.Transactions != nil:
		txs := make([]interface{}, 0, len(block.Transactions))
		for _, t := range block.Transactions {
			txs = append(txs, ConfirmedTransactionToJsonRpc(t))
		}
		res["transactions"] = txs
	}

	if block.Rewards != nil {
		res["rewards"] = rewardsToJsonRpc(block.Rewards)
	}
	if block.NumRewardPartitions != nil {
		res["numRewardPartitions"] = *block.NumRewardPartitions
	}

	return res
}

// ConfirmedTransactionToJsonRpc renders one entry of `getBlock.transactions`.
func ConfirmedTransactionToJsonRpc(ct *ConfirmedTransaction) map[string]interface{} {
	if ct == nil {
		return nil
	}

	res := map[string]interface{}{}
	if ct.Transaction != nil {
		res["transaction"] = TransactionToJsonRpc(ct.Transaction)
	}
	if ct.Meta != nil {
		res["meta"] = MetaToJsonRpc(ct.Meta)
	} else {
		res["meta"] = nil
	}

	// Agave renders legacy as the string "legacy" and versioned as the numeric
	// version (0), not "0". The key is omitted entirely when the client did not
	// opt in via maxSupportedTransactionVersion, which is the caller's call.
	if ct.Version != nil {
		if *ct.Version == "legacy" {
			res["version"] = "legacy"
		} else if n, err := strconv.ParseUint(*ct.Version, 10, 32); err == nil {
			res["version"] = n
		} else {
			res["version"] = *ct.Version
		}
	}

	return res
}

// TransactionToJsonRpc renders the `transaction` object (signatures + message).
func TransactionToJsonRpc(t *Transaction) map[string]interface{} {
	if t == nil {
		return nil
	}

	sigs := make([]interface{}, 0, len(t.Signatures))
	for _, s := range t.Signatures {
		sigs = append(sigs, Base58Encode(s))
	}

	res := map[string]interface{}{"signatures": sigs}
	if t.Message != nil {
		res["message"] = MessageToJsonRpc(t.Message)
	}
	return res
}

// MessageToJsonRpc renders `transaction.message`.
func MessageToJsonRpc(m *Message) map[string]interface{} {
	if m == nil {
		return nil
	}

	keys := make([]interface{}, 0, len(m.AccountKeys))
	for _, k := range m.AccountKeys {
		keys = append(keys, Base58Encode(k))
	}

	instrs := make([]interface{}, 0, len(m.Instructions))
	for _, i := range m.Instructions {
		instrs = append(instrs, CompiledInstructionToJsonRpc(i))
	}

	res := map[string]interface{}{
		"accountKeys":     keys,
		"recentBlockhash": Base58Encode(m.RecentBlockhash),
		"instructions":    instrs,
	}
	if m.Header != nil {
		res["header"] = map[string]interface{}{
			"numRequiredSignatures":       m.Header.NumRequiredSignatures,
			"numReadonlySignedAccounts":   m.Header.NumReadonlySignedAccounts,
			"numReadonlyUnsignedAccounts": m.Header.NumReadonlyUnsignedAccounts,
		}
	}
	// Legacy transactions carry no lookups; Agave omits the key for them.
	if len(m.AddressTableLookups) > 0 {
		lookups := make([]interface{}, 0, len(m.AddressTableLookups))
		for _, l := range m.AddressTableLookups {
			lookups = append(lookups, map[string]interface{}{
				"accountKey":      Base58Encode(l.AccountKey),
				"writableIndexes": byteIndexes(l.WritableIndexes),
				"readonlyIndexes": byteIndexes(l.ReadonlyIndexes),
			})
		}
		res["addressTableLookups"] = lookups
	}
	return res
}

// CompiledInstructionToJsonRpc renders one compiled instruction. Instruction
// data is base58 under `encoding: "json"`.
func CompiledInstructionToJsonRpc(ci *CompiledInstruction) map[string]interface{} {
	if ci == nil {
		return nil
	}
	res := map[string]interface{}{
		"programIdIndex": ci.ProgramIdIndex,
		"accounts":       byteIndexes(ci.Accounts),
		"data":           Base58Encode(ci.Data),
	}
	if ci.StackHeight != nil {
		res["stackHeight"] = *ci.StackHeight
	} else {
		res["stackHeight"] = nil
	}
	return res
}

// MetaToJsonRpc renders `transaction.meta`, preserving Agave's null-vs-empty
// distinctions via the proto's *None flags.
func MetaToJsonRpc(m *TransactionStatusMeta) map[string]interface{} {
	if m == nil {
		return nil
	}

	res := map[string]interface{}{
		"fee":          m.Fee,
		"preBalances":  uint64Slice(m.PreBalances),
		"postBalances": uint64Slice(m.PostBalances),
	}

	// err is carried as the JSON-stringified object; re-inflate it so the wire
	// shape is an object, not a string. status is derived from it.
	if m.Err != nil {
		var errObj interface{}
		if json.Unmarshal([]byte(*m.Err), &errObj) == nil {
			res["err"] = errObj
			res["status"] = map[string]interface{}{"Err": errObj}
		} else {
			res["err"] = *m.Err
			res["status"] = map[string]interface{}{"Err": *m.Err}
		}
	} else {
		res["err"] = nil
		res["status"] = map[string]interface{}{"Ok": nil}
	}

	if m.InnerInstructionsNone {
		res["innerInstructions"] = nil
	} else {
		inner := make([]interface{}, 0, len(m.InnerInstructions))
		for _, ii := range m.InnerInstructions {
			instrs := make([]interface{}, 0, len(ii.Instructions))
			for _, ci := range ii.Instructions {
				instrs = append(instrs, CompiledInstructionToJsonRpc(ci))
			}
			inner = append(inner, map[string]interface{}{
				"index":        ii.Index,
				"instructions": instrs,
			})
		}
		res["innerInstructions"] = inner
	}

	if m.LogMessagesNone {
		res["logMessages"] = nil
	} else {
		logs := make([]interface{}, 0, len(m.LogMessages))
		for _, l := range m.LogMessages {
			logs = append(logs, l)
		}
		res["logMessages"] = logs
	}

	if m.PreTokenBalancesNone {
		res["preTokenBalances"] = nil
	} else {
		res["preTokenBalances"] = tokenBalancesToJsonRpc(m.PreTokenBalances)
	}
	if m.PostTokenBalancesNone {
		res["postTokenBalances"] = nil
	} else {
		res["postTokenBalances"] = tokenBalancesToJsonRpc(m.PostTokenBalances)
	}

	if m.RewardsNone {
		res["rewards"] = nil
	} else {
		res["rewards"] = rewardsToJsonRpc(m.Rewards)
	}

	// loadedAddresses is a v0 concept; Agave emits it whenever the client opted
	// into versioned transactions. Emitting the empty pair for legacy matches
	// Agave's own behaviour under maxSupportedTransactionVersion.
	res["loadedAddresses"] = map[string]interface{}{
		"writable": base58Slice(m.LoadedWritableAddresses),
		"readonly": base58Slice(m.LoadedReadonlyAddresses),
	}

	if m.ReturnData != nil {
		res["returnData"] = map[string]interface{}{
			"programId": Base58Encode(m.ReturnData.ProgramId),
			// Agave renders return data as a [payload, encoding] tuple.
			"data": []interface{}{
				base64.StdEncoding.EncodeToString(m.ReturnData.Data),
				"base64",
			},
		}
	} else {
		res["returnData"] = nil
	}

	if m.ComputeUnitsConsumed != nil {
		res["computeUnitsConsumed"] = *m.ComputeUnitsConsumed
	}
	if m.CostUnits != nil {
		res["costUnits"] = *m.CostUnits
	}

	return res
}

// RewardToJsonRpc renders one reward entry.
func RewardToJsonRpc(r *Reward) map[string]interface{} {
	if r == nil {
		return nil
	}
	res := map[string]interface{}{
		"pubkey":   Base58Encode(r.Pubkey),
		"lamports": r.Lamports,
	}
	if r.PostBalance != nil {
		res["postBalance"] = *r.PostBalance
	}
	if r.RewardType != nil {
		res["rewardType"] = *r.RewardType
	} else {
		res["rewardType"] = nil
	}
	if r.Commission != nil {
		res["commission"] = *r.Commission
	} else {
		res["commission"] = nil
	}
	return res
}

// TokenBalanceToJsonRpc renders one pre/postTokenBalances entry.
func TokenBalanceToJsonRpc(tb *TokenBalance) map[string]interface{} {
	if tb == nil {
		return nil
	}
	res := map[string]interface{}{
		"accountIndex": tb.AccountIndex,
		"mint":         Base58Encode(tb.Mint),
	}
	if tb.Owner != nil {
		res["owner"] = Base58Encode(tb.Owner)
	}
	if tb.ProgramId != nil {
		res["programId"] = Base58Encode(tb.ProgramId)
	}
	if tb.UiTokenAmount != nil {
		res["uiTokenAmount"] = uiTokenAmountToJsonRpc(tb.UiTokenAmount)
	}
	return res
}

func uiTokenAmountToJsonRpc(a *UiTokenAmount) map[string]interface{} {
	res := map[string]interface{}{
		"amount":   a.Amount,
		"decimals": a.Decimals,
	}
	if a.UiAmountString != nil {
		res["uiAmountString"] = *a.UiAmountString
	} else {
		res["uiAmountString"] = a.Amount
	}
	// Agave also emits the lossy float form. Derive it from the raw amount so
	// the wire shape is complete; null when the amount is not parseable.
	if raw, err := strconv.ParseFloat(a.Amount, 64); err == nil {
		res["uiAmount"] = raw / math.Pow10(int(a.Decimals))
	} else {
		res["uiAmount"] = nil
	}
	return res
}

func rewardsToJsonRpc(rs []*Reward) []interface{} {
	out := make([]interface{}, 0, len(rs))
	for _, r := range rs {
		out = append(out, RewardToJsonRpc(r))
	}
	return out
}

func tokenBalancesToJsonRpc(tbs []*TokenBalance) []interface{} {
	out := make([]interface{}, 0, len(tbs))
	for _, tb := range tbs {
		out = append(out, TokenBalanceToJsonRpc(tb))
	}
	return out
}

// byteIndexes renders a packed account-index list as JSON numbers. The proto
// carries them as bytes; Solana's wire form is an array of u8 indexes.
func byteIndexes(b []byte) []interface{} {
	out := make([]interface{}, 0, len(b))
	for _, v := range b {
		out = append(out, uint32(v))
	}
	return out
}

func uint64Slice(vs []uint64) []interface{} {
	out := make([]interface{}, 0, len(vs))
	for _, v := range vs {
		out = append(out, v)
	}
	return out
}

func base58Slice(bs [][]byte) []interface{} {
	out := make([]interface{}, 0, len(bs))
	for _, b := range bs {
		out = append(out, Base58Encode(b))
	}
	return out
}
