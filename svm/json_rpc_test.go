package svm

import (
	"encoding/json"
	"testing"
)

func u64(v uint64) *uint64 { return &v }
func i64(v int64) *int64   { return &v }
func str(v string) *string { return &v }
func u32(v uint32) *uint32 { return &v }

// A nil block is the correct body for a skipped slot; callers turn that into
// Agave's -32007/-32009 rather than an empty object.
func TestBlockToJsonRpcNilBlock(t *testing.T) {
	if got := BlockToJsonRpc(nil, nil); got != nil {
		t.Fatalf("BlockToJsonRpc(nil) = %v, want nil", got)
	}
}

// Agave always emits blockHeight and blockTime, using null when unknown. A
// caller distinguishing "absent" from "zero" depends on the key existing.
func TestBlockToJsonRpcNullableHeaderFields(t *testing.T) {
	res := BlockToJsonRpc(&ConfirmedBlock{Slot: 7, ParentSlot: 6}, nil)

	for _, k := range []string{"blockHeight", "blockTime"} {
		v, ok := res[k]
		if !ok {
			t.Fatalf("key %q missing; Agave always emits it", k)
		}
		if v != nil {
			t.Fatalf("key %q = %v, want nil when unset", k, v)
		}
	}

	res = BlockToJsonRpc(&ConfirmedBlock{BlockHeight: u64(11), BlockTime: i64(1700000000)}, nil)
	if res["blockHeight"] != uint64(11) {
		t.Fatalf("blockHeight = %v, want 11", res["blockHeight"])
	}
	if res["blockTime"] != int64(1700000000) {
		t.Fatalf("blockTime = %v, want 1700000000", res["blockTime"])
	}
}

// transactionDetails=SIGNATURES puts a flat signature list on the response and
// leaves the block without transactions.
func TestBlockToJsonRpcSignaturesOnly(t *testing.T) {
	sig := make([]byte, 64)
	sig[0] = 9
	res := BlockToJsonRpc(&ConfirmedBlock{}, [][]byte{sig})

	sigs, ok := res["signatures"].([]interface{})
	if !ok || len(sigs) != 1 {
		t.Fatalf("signatures = %v, want one entry", res["signatures"])
	}
	if sigs[0] != Base58Encode(sig) {
		t.Fatalf("signature not base58 encoded: %v", sigs[0])
	}
	if _, present := res["transactions"]; present {
		t.Fatal("transactions must be absent in signatures-only mode")
	}
}

// The proto carries err as a JSON string; the wire form is an object, and
// status is derived from it.
func TestMetaErrorInflationAndStatus(t *testing.T) {
	meta := MetaToJsonRpc(&TransactionStatusMeta{
		Err: str(`{"InstructionError":[0,{"Custom":1}]}`),
		Fee: 5000,
	})

	if _, isString := meta["err"].(string); isString {
		t.Fatal("err must be re-inflated to an object, not left as a string")
	}
	errObj, ok := meta["err"].(map[string]interface{})
	if !ok || errObj["InstructionError"] == nil {
		t.Fatalf("err = %#v, want inflated InstructionError object", meta["err"])
	}
	status, ok := meta["status"].(map[string]interface{})
	if !ok {
		t.Fatalf("status = %#v, want object", meta["status"])
	}
	if _, hasErr := status["Err"]; !hasErr {
		t.Fatalf("status = %#v, want an Err key mirroring err", status)
	}

	ok2 := MetaToJsonRpc(&TransactionStatusMeta{Fee: 5000})
	if ok2["err"] != nil {
		t.Fatalf("err = %v, want nil on success", ok2["err"])
	}
	st, _ := ok2["status"].(map[string]interface{})
	if _, hasOk := st["Ok"]; !hasOk {
		t.Fatalf("status = %#v, want {Ok:null} on success", st)
	}
}

// Agave distinguishes null from [] on these meta fields for pre-recording
// history. The *None flags carry that distinction and must survive.
func TestMetaNullVersusEmptyDistinction(t *testing.T) {
	none := MetaToJsonRpc(&TransactionStatusMeta{
		InnerInstructionsNone: true,
		LogMessagesNone:       true,
		PreTokenBalancesNone:  true,
		PostTokenBalancesNone: true,
		RewardsNone:           true,
	})
	for _, k := range []string{"innerInstructions", "logMessages", "preTokenBalances", "postTokenBalances", "rewards"} {
		if none[k] != nil {
			t.Fatalf("%q = %v, want nil when the *None flag is set", k, none[k])
		}
	}

	empty := MetaToJsonRpc(&TransactionStatusMeta{})
	for _, k := range []string{"innerInstructions", "logMessages", "preTokenBalances", "postTokenBalances", "rewards"} {
		v, ok := empty[k].([]interface{})
		if !ok {
			t.Fatalf("%q = %#v, want an empty slice when the flag is unset", k, empty[k])
		}
		if len(v) != 0 {
			t.Fatalf("%q = %v, want empty", k, v)
		}
	}
}

// Agave renders legacy as the string "legacy" and v0 as the number 0.
func TestTransactionVersionRendering(t *testing.T) {
	legacy := ConfirmedTransactionToJsonRpc(&ConfirmedTransaction{Version: str("legacy")})
	if legacy["version"] != "legacy" {
		t.Fatalf("version = %#v, want \"legacy\"", legacy["version"])
	}

	v0 := ConfirmedTransactionToJsonRpc(&ConfirmedTransaction{Version: str("0")})
	if v0["version"] != uint64(0) {
		t.Fatalf("version = %#v (%T), want numeric 0", v0["version"], v0["version"])
	}

	// No opt-in recorded: the key is omitted entirely.
	none := ConfirmedTransactionToJsonRpc(&ConfirmedTransaction{})
	if _, present := none["version"]; present {
		t.Fatal("version key must be absent when the proto carries no version")
	}
}

// Under encoding=json instruction data is base58 and accounts are numeric
// indexes — not the parsed form.
func TestCompiledInstructionRendering(t *testing.T) {
	ci := CompiledInstructionToJsonRpc(&CompiledInstruction{
		ProgramIdIndex: 3,
		Accounts:       []byte{1, 2, 250},
		Data:           []byte{0xde, 0xad},
		StackHeight:    u32(2),
	})

	if ci["data"] != Base58Encode([]byte{0xde, 0xad}) {
		t.Fatalf("data = %v, want base58", ci["data"])
	}
	accs, ok := ci["accounts"].([]interface{})
	if !ok || len(accs) != 3 || accs[2] != uint32(250) {
		t.Fatalf("accounts = %#v, want numeric indexes [1 2 250]", ci["accounts"])
	}
	if ci["stackHeight"] != uint32(2) {
		t.Fatalf("stackHeight = %v, want 2", ci["stackHeight"])
	}

	noStack := CompiledInstructionToJsonRpc(&CompiledInstruction{})
	if v, present := noStack["stackHeight"]; !present || v != nil {
		t.Fatalf("stackHeight = %#v, want explicit nil when unset", v)
	}
}

// Return data is a [payload, encoding] tuple in Agave's wire form.
func TestReturnDataTuple(t *testing.T) {
	meta := MetaToJsonRpc(&TransactionStatusMeta{
		ReturnData: &ReturnData{ProgramId: make([]byte, 32), Data: []byte{0x01, 0x02}},
	})
	rd, ok := meta["returnData"].(map[string]interface{})
	if !ok {
		t.Fatalf("returnData = %#v, want object", meta["returnData"])
	}
	tuple, ok := rd["data"].([]interface{})
	if !ok || len(tuple) != 2 || tuple[1] != "base64" {
		t.Fatalf("returnData.data = %#v, want [payload, \"base64\"]", rd["data"])
	}
	if tuple[0] != "AQI=" {
		t.Fatalf("returnData payload = %v, want base64 of {1,2}", tuple[0])
	}

	// Absent return data omits the key entirely (Agave or_skip semantics, as
	// captured from a live node in testdata/parsed_golden.json).
	absent := MetaToJsonRpc(&TransactionStatusMeta{})
	if v, present := absent["returnData"]; present {
		t.Fatalf("returnData = %#v, want the key omitted when absent", v)
	}
}

// uiAmount is the lossy float Agave still emits; uiAmountString is authoritative.
func TestUiTokenAmountRendering(t *testing.T) {
	tb := TokenBalanceToJsonRpc(&TokenBalance{
		AccountIndex:  1,
		Mint:          make([]byte, 32),
		UiTokenAmount: &UiTokenAmount{Amount: "1500000", Decimals: 6, UiAmountString: str("1.5")},
	})
	amt, ok := tb["uiTokenAmount"].(map[string]interface{})
	if !ok {
		t.Fatalf("uiTokenAmount = %#v, want object", tb["uiTokenAmount"])
	}
	if amt["amount"] != "1500000" {
		t.Fatalf("amount must stay a string, got %#v", amt["amount"])
	}
	if amt["uiAmountString"] != "1.5" {
		t.Fatalf("uiAmountString = %v, want 1.5", amt["uiAmountString"])
	}
	if f, ok := amt["uiAmount"].(float64); !ok || f != 1.5 {
		t.Fatalf("uiAmount = %#v, want 1.5", amt["uiAmount"])
	}
}

// Legacy transactions carry no address table lookups; Agave omits the key.
func TestMessageAddressTableLookups(t *testing.T) {
	legacy := MessageToJsonRpc(&Message{RecentBlockhash: make([]byte, 32)})
	if _, present := legacy["addressTableLookups"]; present {
		t.Fatal("addressTableLookups must be absent for legacy messages")
	}

	v0 := MessageToJsonRpc(&Message{
		RecentBlockhash: make([]byte, 32),
		AddressTableLookups: []*AddressTableLookup{{
			AccountKey:      make([]byte, 32),
			WritableIndexes: []byte{0, 1},
			ReadonlyIndexes: []byte{2},
		}},
	})
	lookups, ok := v0["addressTableLookups"].([]interface{})
	if !ok || len(lookups) != 1 {
		t.Fatalf("addressTableLookups = %#v, want one entry", v0["addressTableLookups"])
	}
}

// The whole shape must survive encoding/json without panicking on unsupported
// types — this is what the serving layer ultimately does with it.
func TestBlockRoundTripsThroughEncodingJson(t *testing.T) {
	block := &ConfirmedBlock{
		Slot:              100,
		Blockhash:         make([]byte, 32),
		PreviousBlockhash: make([]byte, 32),
		ParentSlot:        99,
		BlockHeight:       u64(90),
		BlockTime:         i64(1700000000),
		Transactions: []*ConfirmedTransaction{{
			Transaction: &Transaction{
				Signatures: [][]byte{make([]byte, 64)},
				Message: &Message{
					Header:          &MessageHeader{NumRequiredSignatures: 1},
					AccountKeys:     [][]byte{make([]byte, 32)},
					RecentBlockhash: make([]byte, 32),
					Instructions:    []*CompiledInstruction{{ProgramIdIndex: 0, Accounts: []byte{0}, Data: []byte{1}}},
				},
			},
			Meta:    &TransactionStatusMeta{Fee: 5000, PreBalances: []uint64{1}, PostBalances: []uint64{0}},
			Version: str("0"),
		}},
		Rewards: []*Reward{{Pubkey: make([]byte, 32), Lamports: -25, PostBalance: u64(10)}},
	}

	out, err := json.Marshal(BlockToJsonRpc(block, nil))
	if err != nil {
		t.Fatalf("result is not JSON-serialisable: %v", err)
	}

	var back map[string]interface{}
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("round trip failed: %v", err)
	}
	if back["parentSlot"].(float64) != 99 {
		t.Fatalf("parentSlot = %v, want 99", back["parentSlot"])
	}
	// Negative reward lamports (rent debits in old history) must survive.
	rewards := back["rewards"].([]interface{})
	first := rewards[0].(map[string]interface{})
	if first["lamports"].(float64) != -25 {
		t.Fatalf("reward lamports = %v, want -25", first["lamports"])
	}
}

// uiAmountString is the SCALED decimal, never the raw integer. Falling back to
// the raw amount misreports the balance by 10^decimals, and deriving it via
// float loses precision above 2^53 — token amounts are u64.
func TestUiAmountStringDerivedWhenAbsent(t *testing.T) {
	cases := []struct {
		amount   string
		decimals uint32
		want     string
	}{
		{"1500000", 6, "1.5"},
		{"1000000", 6, "1"},  // trailing fraction zeros trimmed
		{"1", 6, "0.000001"}, // shorter than decimals -> zero padded
		{"0", 6, "0"},
		{"123", 0, "123"}, // decimals=0 passes through
		{"18446744073709551615", 9, "18446744073.709551615"}, // u64 max, exact
	}

	for _, tc := range cases {
		got := uiTokenAmountToJsonRpc(&UiTokenAmount{Amount: tc.amount, Decimals: tc.decimals})
		if got["uiAmountString"] != tc.want {
			t.Fatalf("amount=%s decimals=%d: uiAmountString = %v, want %q",
				tc.amount, tc.decimals, got["uiAmountString"], tc.want)
		}
		if got["amount"] != tc.amount {
			t.Fatalf("raw amount must be preserved verbatim, got %v", got["amount"])
		}
	}

	// An explicit uiAmountString from the source always wins.
	explicit := uiTokenAmountToJsonRpc(&UiTokenAmount{
		Amount: "1500000", Decimals: 6, UiAmountString: str("1.500"),
	})
	if explicit["uiAmountString"] != "1.500" {
		t.Fatalf("source uiAmountString must be preserved, got %v", explicit["uiAmountString"])
	}
}
