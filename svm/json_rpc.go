package svm

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"strconv"
	"strings"
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
		txr := TransactionToJsonRpc(ct.Transaction)
		// addressTableLookups presence keys on version, not emptiness: a v0
		// message carries the key even with zero lookups (MessageToJsonRpc
		// cannot know the version, so the omission is repaired here).
		if ct.Version != nil && *ct.Version != "legacy" {
			if msg, ok := txr["message"].(map[string]interface{}); ok {
				if _, has := msg["addressTableLookups"]; !has {
					msg["addressTableLookups"] = []interface{}{}
				}
			}
		}
		res["transaction"] = txr
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
	return metaToJsonRpc(m, CompiledInstructionToJsonRpc, true)
}

// metaToJsonRpc is the shared meta renderer. Two things differ between
// `json` and `jsonParsed`: the instruction renderer, and loadedAddresses —
// present under `json`, OMITTED under `jsonParsed` (Agave merges the loaded
// keys into the parsed accountKeys instead of repeating them in meta).
func metaToJsonRpc(m *TransactionStatusMeta, renderIx func(*CompiledInstruction) map[string]interface{}, includeLoadedAddresses bool) map[string]interface{} {
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
				instrs = append(instrs, renderIx(ci))
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
	if includeLoadedAddresses {
		res["loadedAddresses"] = map[string]interface{}{
			"writable": base58Slice(m.LoadedWritableAddresses),
			"readonly": base58Slice(m.LoadedReadonlyAddresses),
		}
	}

	// Absent return data omits the key entirely (Agave's OptionSerializer
	// or_skip semantics) — never `returnData: null`. Pinned by the captured
	// node output in testdata/parsed_golden.json.
	if m.ReturnData != nil {
		res["returnData"] = map[string]interface{}{
			"programId": Base58Encode(m.ReturnData.ProgramId),
			// Agave renders return data as a [payload, encoding] tuple.
			"data": []interface{}{
				base64.StdEncoding.EncodeToString(m.ReturnData.Data),
				"base64",
			},
		}
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
	// uiAmountString is the scaled decimal ("1.5"), never the raw integer.
	// Derive it by shifting the decimal point in the digit string rather than
	// via float — token amounts are u64 and lose precision above 2^53.
	if a.UiAmountString != nil {
		res["uiAmountString"] = *a.UiAmountString
	} else {
		res["uiAmountString"] = shiftDecimalString(a.Amount, a.Decimals)
	}
	// Agave also emits the lossy float form; keep it for wire completeness.
	// null when the amount is not parseable AND for a zero amount — in the
	// captured node output (testdata/parsed_golden.json) all 12 zero-amount
	// balances carry uiAmount null while all 68 non-zero ones carry a number.
	if raw, err := strconv.ParseFloat(a.Amount, 64); err == nil && raw != 0 {
		res["uiAmount"] = raw / math.Pow10(int(a.Decimals))
	} else {
		res["uiAmount"] = nil
	}
	return res
}

// shiftDecimalString scales a non-negative integer digit string down by
// 10^decimals exactly, trimming trailing fraction zeros the way Agave's
// uiAmountString does ("1.500000" -> "1.5", "1000000" -> "1").
func shiftDecimalString(amount string, decimals uint32) string {
	if decimals == 0 || amount == "" {
		return amount
	}
	for _, c := range amount {
		if c < '0' || c > '9' {
			return amount // not a plain digit string; pass through untouched
		}
	}

	d := int(decimals)
	if len(amount) <= d {
		amount = strings.Repeat("0", d-len(amount)+1) + amount
	}
	split := len(amount) - d
	intPart, frac := amount[:split], amount[split:]

	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return intPart
	}
	return intPart + "." + frac
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

// ---------------------------------------------------------------------------
// encoding: "jsonParsed"
//
// The parsed form differs from `json` in exactly three places, all inside the
// transaction envelope; block-level fields, meta balances, rewards and token
// balances are byte-identical between the two:
//
//   - message.accountKeys: objects {pubkey, signer, writable, source} over the
//     MERGED key list (static ++ loadedWritable ++ loadedReadonly) instead of
//     base58 strings of the static list. The message-level `header` key is
//     omitted (Agave's UiParsedMessage has no header field).
//   - every instruction (top-level and inner): Agave's ParsedInstruction when
//     the server attached one (CompiledInstruction.parsed, spliced verbatim),
//     else Agave's partiallyDecoded form {programId, accounts, data,
//     stackHeight} with indexes resolved to base58 pubkeys.
//
// The signer/writable flags for static keys follow the message header:
// signer  = i < numRequiredSignatures
// writable= signers:     i < numRequiredSignatures-numReadonlySignedAccounts
//           non-signers: i < len(static)-numReadonlyUnsignedAccounts
// Loaded addresses are never signers; writability is which list they came
// from. This is the header math only — Agave additionally demotes certain
// program-id accounts in newer runtime rules; the golden tests pin any
// divergence when it appears.
// ---------------------------------------------------------------------------

// BlockToJsonRpcParsed renders `encoding: "jsonParsed"`. See BlockToJsonRpc
// for the envelope contract (nil block, signatures mode, rewards).
func BlockToJsonRpcParsed(block *ConfirmedBlock, signatures [][]byte) map[string]interface{} {
	if block == nil {
		return nil
	}
	res := BlockToJsonRpc(block, signatures)
	// Signatures mode carries no transactions — nothing to re-render.
	if signatures != nil || block.Transactions == nil {
		return res
	}
	txs := make([]interface{}, 0, len(block.Transactions))
	for _, t := range block.Transactions {
		txs = append(txs, ConfirmedTransactionToJsonRpcParsed(t))
	}
	res["transactions"] = txs
	return res
}

// ConfirmedTransactionToJsonRpcParsed renders one transactions[] entry in
// parsed form. The merged key list needs meta (loaded addresses live there),
// which is why this cannot be a per-message concern.
func ConfirmedTransactionToJsonRpcParsed(ct *ConfirmedTransaction) map[string]interface{} {
	if ct == nil {
		return nil
	}

	// Merged key list, in Agave's documented order.
	var keys []string
	if ct.Transaction != nil && ct.Transaction.Message != nil {
		m := ct.Transaction.Message
		keys = make([]string, 0, len(m.AccountKeys))
		for _, k := range m.AccountKeys {
			keys = append(keys, Base58Encode(k))
		}
	}
	if ct.Meta != nil {
		for _, k := range ct.Meta.LoadedWritableAddresses {
			keys = append(keys, Base58Encode(k))
		}
		for _, k := range ct.Meta.LoadedReadonlyAddresses {
			keys = append(keys, Base58Encode(k))
		}
	}

	renderIx := func(ci *CompiledInstruction) map[string]interface{} {
		return parsedInstructionToJsonRpc(ci, keys)
	}

	res := map[string]interface{}{}
	if ct.Transaction != nil {
		t := ct.Transaction
		sigs := make([]interface{}, 0, len(t.Signatures))
		for _, s := range t.Signatures {
			sigs = append(sigs, Base58Encode(s))
		}
		tx := map[string]interface{}{"signatures": sigs}
		if t.Message != nil {
			var lw, lr [][]byte
			if ct.Meta != nil {
				lw, lr = ct.Meta.LoadedWritableAddresses, ct.Meta.LoadedReadonlyAddresses
			}
			tx["message"] = messageToJsonRpcParsed(t.Message, lw, lr, ct.Version != nil && *ct.Version != "legacy", renderIx)
		}
		res["transaction"] = tx
	}
	if ct.Meta != nil {
		res["meta"] = metaToJsonRpc(ct.Meta, renderIx, false)
	} else {
		res["meta"] = nil
	}
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

// messageToJsonRpcParsed renders `transaction.message` in parsed form.
func messageToJsonRpcParsed(m *Message, loadedWritable, loadedReadonly [][]byte, emitLookups bool, renderIx func(*CompiledInstruction) map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}

	numStatic := len(m.AccountKeys)
	var reqSigs, roSigned, roUnsigned int
	if m.Header != nil {
		reqSigs = int(m.Header.NumRequiredSignatures)
		roSigned = int(m.Header.NumReadonlySignedAccounts)
		roUnsigned = int(m.Header.NumReadonlyUnsignedAccounts)
	}

	keys := make([]interface{}, 0, numStatic+len(loadedWritable)+len(loadedReadonly))
	for i, k := range m.AccountKeys {
		signer := i < reqSigs
		var writable bool
		if signer {
			writable = i < reqSigs-roSigned
		} else {
			writable = i < numStatic-roUnsigned
		}
		keys = append(keys, map[string]interface{}{
			"pubkey":   Base58Encode(k),
			"signer":   signer,
			"writable": writable,
			"source":   "transaction",
		})
	}
	for _, k := range loadedWritable {
		keys = append(keys, map[string]interface{}{
			"pubkey": Base58Encode(k), "signer": false, "writable": true, "source": "lookupTable",
		})
	}
	for _, k := range loadedReadonly {
		keys = append(keys, map[string]interface{}{
			"pubkey": Base58Encode(k), "signer": false, "writable": false, "source": "lookupTable",
		})
	}

	instrs := make([]interface{}, 0, len(m.Instructions))
	for _, i := range m.Instructions {
		instrs = append(instrs, renderIx(i))
	}

	res := map[string]interface{}{
		"accountKeys":     keys,
		"recentBlockhash": Base58Encode(m.RecentBlockhash),
		"instructions":    instrs,
	}
	// No header key: UiParsedMessage does not carry one.
	//
	// addressTableLookups presence keys on the transaction VERSION, not on
	// emptiness: a v0 message carries the key even with zero lookups, legacy
	// omits it. The repeated proto field cannot express that distinction, so
	// the caller passes it down from ConfirmedTransaction.version.
	if emitLookups {
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

// parsedInstructionToJsonRpc renders one instruction in parsed form: the
// server-attached ParsedInstruction verbatim when present, else Agave's
// partiallyDecoded shape over the merged key list.
func parsedInstructionToJsonRpc(ci *CompiledInstruction, keys []string) map[string]interface{} {
	if ci == nil {
		return nil
	}
	if ci.Parsed != nil {
		var out map[string]interface{}
		if err := json.Unmarshal(ci.Parsed, &out); err == nil {
			return out
		}
		// A malformed attachment falls through to partiallyDecoded rather than
		// dropping the instruction; the shape stays valid either way.
	}
	accounts := make([]interface{}, 0, len(ci.Accounts))
	for _, idx := range ci.Accounts {
		accounts = append(accounts, keyAt(keys, int(idx)))
	}
	res := map[string]interface{}{
		"programId": keyAt(keys, int(ci.ProgramIdIndex)),
		"accounts":  accounts,
		"data":      Base58Encode(ci.Data),
	}
	if ci.StackHeight != nil {
		res["stackHeight"] = *ci.StackHeight
	} else {
		res["stackHeight"] = nil
	}
	return res
}

// keyAt resolves a merged-list index defensively: writer-produced data never
// carries an out-of-range index, but a serving path must not panic on one.
// The empty string is deliberately visible in output rather than silently
// substituting a wrong key.
func keyAt(keys []string, i int) string {
	if i >= 0 && i < len(keys) {
		return keys[i]
	}
	return ""
}
