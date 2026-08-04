package svm

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"
)

// Regression pins from the 7-block mainnet live differential (2026-08-04,
// 9,908 transactions vs node getBlock output) for the three renderer seams it
// fixed or deliberately froze:
//
//  1. writable demotion in parsed accountKeys (117 live divergences) —
//     is_maybe_writable / is_writable_internal from solana-message 2.4.0
//     (legacy.rs:676-700, versions/v0/loaded.rs:118-175) with
//     ReservedAccountKeys::new_all_activated (agave-reserved-account-keys
//     2.3.13);
//  2. UseNumber on the spliced Parsed envelope (u64::MAX lamports exist in
//     the wild and must not round through float64);
//  3. uiAmount as f64(amount)/10^decimals, double rounding included
//     (solana-account-decoder 2.3.13 token_amount_to_ui_amount_v3 plain path).

// -----------------------------------------------------------------------
// 1. Writable demotion
// -----------------------------------------------------------------------

// demotionTx hand-builds the minimal transaction the demotion rules read:
// static keys, an all-writable header (1 signer, no readonly sections, so
// every static index is header-writable and only demotion can flip one),
// top-level instructions, and optional meta for loaded addresses / inners.
func demotionTx(version string, static [][]byte, ixs []*CompiledInstruction, meta *TransactionStatusMeta) *ConfirmedTransaction {
	return &ConfirmedTransaction{
		Version: str(version),
		Transaction: &Transaction{
			Signatures: [][]byte{bytes.Repeat([]byte{9}, 64)},
			Message: &Message{
				Header:          &MessageHeader{NumRequiredSignatures: 1},
				AccountKeys:     static,
				RecentBlockhash: bytes.Repeat([]byte{7}, 32),
				Instructions:    ixs,
			},
		},
		Meta: meta,
	}
}

func renderedKeys(t *testing.T, ct *ConfirmedTransaction) []map[string]interface{} {
	t.Helper()
	msg := messageOf(t, 0, ConfirmedTransactionToJsonRpcParsed(ct))
	raw, ok := msg["accountKeys"].([]interface{})
	if !ok {
		t.Fatalf("accountKeys = %#v, want array", msg["accountKeys"])
	}
	out := make([]map[string]interface{}, len(raw))
	for i, k := range raw {
		m, ok := k.(map[string]interface{})
		if !ok {
			t.Fatalf("accountKeys[%d] = %#v, want object", i, k)
		}
		out[i] = m
	}
	return out
}

// checkKey asserts one accountKeys entry; signer is always checked so any
// demotion bug that leaks into the signer flag trips these tests too.
func checkKey(t *testing.T, keys []map[string]interface{}, i int, signer, writable bool, source string) {
	t.Helper()
	if i >= len(keys) {
		t.Fatalf("accountKeys has %d entries, want index %d", len(keys), i)
	}
	k := keys[i]
	if k["signer"] != signer || k["writable"] != writable || k["source"] != source {
		t.Fatalf("accountKeys[%d] = %#v, want signer=%v writable=%v source=%q",
			i, k, signer, writable, source)
	}
}

func TestParsedWritableDemotion(t *testing.T) {
	key := func(fill byte) []byte { return bytes.Repeat([]byte{fill}, 32) }
	loader := b58(t, BpfUpgradeableLoaderID)
	callIx := func(program byte) *CompiledInstruction {
		return &CompiledInstruction{ProgramIdIndex: uint32(program), Accounts: []byte{0}, Data: []byte{1}, StackHeight: u32(1)}
	}

	// (a) A static key at a header-writable index that a TOP-LEVEL
	// instruction names as program is demoted to writable false.
	t.Run("top-level program id demoted", func(t *testing.T) {
		keys := renderedKeys(t, demotionTx("legacy",
			[][]byte{key(1), key(2), key(3)},
			[]*CompiledInstruction{callIx(1)}, nil))
		checkKey(t, keys, 0, true, true, "transaction")   // fee payer untouched
		checkKey(t, keys, 1, false, false, "transaction") // called as program -> demoted
		checkKey(t, keys, 2, false, true, "transaction")  // bystander untouched
	})

	// (b) The upgradeable loader among the STATIC keys suppresses program-id
	// demotion (upgradeable programs write to themselves during upgrades).
	// The loader itself is a reserved key, so it stays writable false even at
	// a header-writable index.
	t.Run("static upgradeable loader suppresses demotion", func(t *testing.T) {
		keys := renderedKeys(t, demotionTx("legacy",
			[][]byte{key(1), key(2), loader},
			[]*CompiledInstruction{callIx(1)}, nil))
		checkKey(t, keys, 0, true, true, "transaction")
		checkKey(t, keys, 1, false, true, "transaction")  // demotion suppressed
		checkKey(t, keys, 2, false, false, "transaction") // loader: reserved
	})

	// (b, v0) The loader counts even when it arrives via a lookup table:
	// upgradeable_loader_present scans every key the message loads.
	t.Run("loaded upgradeable loader suppresses demotion", func(t *testing.T) {
		keys := renderedKeys(t, demotionTx("0",
			[][]byte{key(1), key(2)},
			[]*CompiledInstruction{callIx(1)},
			&TransactionStatusMeta{LoadedWritableAddresses: [][]byte{loader}}))
		checkKey(t, keys, 0, true, true, "transaction")
		checkKey(t, keys, 1, false, true, "transaction")  // demotion suppressed
		checkKey(t, keys, 2, false, false, "lookupTable") // loader: reserved
	})

	// (c) Reserved keys (ReservedAccountKeys::new_all_activated) are demoted
	// at header-writable indexes even when nothing calls them as program.
	t.Run("reserved keys demoted without program call", func(t *testing.T) {
		keys := renderedKeys(t, demotionTx("legacy",
			[][]byte{key(1), b58(t, "Vote111111111111111111111111111111111111111"), b58(t, "SysvarC1ock11111111111111111111111111111111")},
			nil, nil))
		checkKey(t, keys, 0, true, true, "transaction")
		checkKey(t, keys, 1, false, false, "transaction") // vote program: reserved
		checkKey(t, keys, 2, false, false, "transaction") // clock sysvar: reserved
	})

	// (d) A lookup-table-sourced key named as program id is demoted too, and
	// demotion never rewrites its source.
	t.Run("lookup-sourced program id demoted keeps source", func(t *testing.T) {
		keys := renderedKeys(t, demotionTx("0",
			[][]byte{key(1)},
			[]*CompiledInstruction{callIx(1)}, // merged index 1 = loadedWritable[0]
			&TransactionStatusMeta{LoadedWritableAddresses: [][]byte{key(2)}}))
		checkKey(t, keys, 0, true, true, "transaction")
		checkKey(t, keys, 1, false, false, "lookupTable")
	})

	// (e) Only TOP-LEVEL message.instructions feed called_as_program. A key
	// that appears as program id solely inside meta.innerInstructions stays
	// writable.
	t.Run("inner instruction program ids do not demote", func(t *testing.T) {
		keys := renderedKeys(t, demotionTx("legacy",
			[][]byte{key(1), key(2), key(3)},
			[]*CompiledInstruction{callIx(2)},
			&TransactionStatusMeta{InnerInstructions: []*InnerInstructions{{
				Index:        0,
				Instructions: []*CompiledInstruction{{ProgramIdIndex: 1, Accounts: []byte{0}, StackHeight: u32(2)}},
			}}}))
		checkKey(t, keys, 0, true, true, "transaction")
		checkKey(t, keys, 1, false, true, "transaction")  // inner-only program: NOT demoted
		checkKey(t, keys, 2, false, false, "transaction") // top-level program: demoted
	})
}

// -----------------------------------------------------------------------
// 2. Parsed-envelope splice precision
// -----------------------------------------------------------------------

// The spliced envelope must survive render -> re-marshal with its u64 digits
// intact: a plain json.Unmarshal would round 18446744073709551615 through
// float64 into 1.8446744073709552e19. Malformed bytes still fall back to the
// partiallyDecoded shape through the same full render path.
func TestParsedSpliceU64Digits(t *testing.T) {
	key := func(fill byte) []byte { return bytes.Repeat([]byte{fill}, 32) }
	envelope := `{"program":"system","programId":"11111111111111111111111111111111","parsed":{"type":"transfer","info":{"lamports":18446744073709551615,"nested":{"space":18446744073709551614}}},"stackHeight":2}`

	ct := &ConfirmedTransaction{
		Version: str("legacy"),
		Transaction: &Transaction{
			Signatures: [][]byte{bytes.Repeat([]byte{9}, 64)},
			Message: &Message{
				Header:          &MessageHeader{NumRequiredSignatures: 1},
				AccountKeys:     [][]byte{key(1), key(2)},
				RecentBlockhash: key(7),
				Instructions: []*CompiledInstruction{
					{ProgramIdIndex: 1, Parsed: []byte(envelope)},
					{ProgramIdIndex: 1, Accounts: []byte{0}, Data: []byte{3}, StackHeight: u32(1), Parsed: []byte(`{"program": mangled`)},
				},
			},
		},
	}

	res := ConfirmedTransactionToJsonRpcParsed(ct)
	out, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("re-marshal rendered map: %v", err)
	}
	if !bytes.Contains(out, []byte("18446744073709551615")) {
		t.Fatalf("u64::MAX digits mangled in output: %s", out)
	}
	if !bytes.Contains(out, []byte("18446744073709551614")) {
		t.Fatalf("nested large number mangled in output: %s", out)
	}

	msg := messageOf(t, 0, res)
	instrs, _ := msg["instructions"].([]interface{})
	if len(instrs) != 2 {
		t.Fatalf("instructions = %#v, want 2 entries", instrs)
	}

	// Spliced instruction: lamports round-trips as json.Number, never float64.
	spliced, _ := instrs[0].(map[string]interface{})
	parsed, _ := spliced["parsed"].(map[string]interface{})
	info, _ := parsed["info"].(map[string]interface{})
	if info == nil {
		t.Fatalf("instructions[0] = %#v, want spliced envelope with parsed.info", spliced)
	}
	lam, ok := info["lamports"].(json.Number)
	if !ok || lam.String() != "18446744073709551615" {
		t.Fatalf("lamports = %#v (%T), want json.Number 18446744073709551615", info["lamports"], info["lamports"])
	}

	// Malformed attachment: partiallyDecoded shape, no parsed key, no panic.
	fallback, _ := instrs[1].(map[string]interface{})
	if _, has := fallback["parsed"]; has {
		t.Fatalf("instructions[1] = %#v, malformed bytes must not surface a parsed key", fallback)
	}
	if fallback["programId"] != Base58Encode(key(2)) {
		t.Fatalf("fallback programId = %v, want %v", fallback["programId"], Base58Encode(key(2)))
	}
}

// -----------------------------------------------------------------------
// 3. uiAmount division pin
// -----------------------------------------------------------------------

// uiAmount is f64(amount) / 10^decimals — token_amount_to_ui_amount_v3's
// plain path VERBATIM, double rounding included. Do NOT "fix" this to parse
// the exact shifted decimal string: live differential 2026-08-04 — division
// matches node getBlock 126/154 and getTransaction 28/28; residual is
// provider noise.
func TestUiAmountDivisionPin(t *testing.T) {
	uiAmount := func(amount string, decimals uint32) interface{} {
		return uiTokenAmountToJsonRpc(&UiTokenAmount{Amount: amount, Decimals: decimals})["uiAmount"]
	}

	// The differential's pin value (>2^53, so f64 conversion already rounds).
	amt, err := strconv.ParseFloat("999999997327427901", 64)
	if err != nil {
		t.Fatal(err)
	}
	want := amt / 1e9
	if s := strconv.FormatFloat(want, 'f', -1, 64); s != "999999997.3274279" {
		t.Fatalf("division shortest-form = %s, want 999999997.3274279", s)
	}
	if got := uiAmount("999999997327427901", 9); got != want {
		t.Fatalf("uiAmount(999999997327427901, 9) = %v, want %v", got, want)
	}

	// Adjacent live-range value where division and exact-decimal string-parse
	// produce DIFFERENT floats — the tripwire that catches a string-parse
	// "fix" (for ...901 above the two happen to coincide).
	amt2, err := strconv.ParseFloat("999999997327427800", 64)
	if err != nil {
		t.Fatal(err)
	}
	div := amt2 / 1e9
	parse, err := strconv.ParseFloat("999999997.327427800", 64)
	if err != nil {
		t.Fatal(err)
	}
	if div == parse {
		t.Fatal("sentinel value no longer discriminates division from string-parse; pick another")
	}
	if got := uiAmount("999999997327427800", 9); got != div {
		t.Fatalf("uiAmount(999999997327427800, 9) = %v, want division result %v, not string-parse %v", got, div, parse)
	}

	// Zero amount renders uiAmount null (all 12 zero-amount balances in the
	// captured node output carry null).
	if got := uiAmount("0", 9); got != nil {
		t.Fatalf("uiAmount(0, 9) = %v, want nil", got)
	}

	// Below 2^53 division and exact parse agree; both must equal 1.5.
	if got := uiAmount("1500000", 6); got != 1.5 {
		t.Fatalf("uiAmount(1500000, 6) = %v, want 1.5", got)
	}
}
