package svm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Golden-fidelity tests for encoding=jsonParsed: testdata/parsed_golden.json
// holds one real mainnet block fetched twice from a live Agave node — once
// with encoding=json and once with encoding=jsonParsed, transactions
// index-aligned. The json side is reversed back into the protos a server
// would hold (base58/base64 decoded, indexes repacked), the jsonParsed side's
// parsed instructions are attached as CompiledInstruction.parsed exactly the
// way prism does via Agave's parser crate, and the renderer's output must
// deep-equal what the node itself returned.

// ---------------------------------------------------------------------------
// Fixture shapes (the node's encoding=json output, typed for reversal).
// ---------------------------------------------------------------------------

type goldenFixture struct {
	Slot       uint64      `json:"slot"`
	Json       goldenBlock `json:"json"`
	JsonParsed goldenBlock `json:"jsonParsed"`
}

type goldenBlock struct {
	Transactions []json.RawMessage `json:"transactions"`
}

type gjTx struct {
	Version     json.RawMessage `json:"version"` // "legacy" or the number 0
	Transaction struct {
		Signatures []string  `json:"signatures"`
		Message    gjMessage `json:"message"`
	} `json:"transaction"`
	Meta *gjMeta `json:"meta"`
}

type gjMessage struct {
	AccountKeys []string `json:"accountKeys"`
	Header      *struct {
		NumRequiredSignatures       uint32 `json:"numRequiredSignatures"`
		NumReadonlySignedAccounts   uint32 `json:"numReadonlySignedAccounts"`
		NumReadonlyUnsignedAccounts uint32 `json:"numReadonlyUnsignedAccounts"`
	} `json:"header"`
	RecentBlockhash     string     `json:"recentBlockhash"`
	Instructions        []gjIx     `json:"instructions"`
	AddressTableLookups []gjLookup `json:"addressTableLookups"`
}

type gjIx struct {
	ProgramIdIndex uint32   `json:"programIdIndex"`
	Accounts       []uint32 `json:"accounts"`
	Data           string   `json:"data"`
	StackHeight    *uint32  `json:"stackHeight"`
}

type gjLookup struct {
	AccountKey      string   `json:"accountKey"`
	WritableIndexes []uint32 `json:"writableIndexes"`
	ReadonlyIndexes []uint32 `json:"readonlyIndexes"`
}

// Pointer-to-slice fields keep Agave's null-vs-[] distinction so the *None
// proto flags can be reconstructed faithfully.
type gjMeta struct {
	Err               json.RawMessage   `json:"err"`
	Fee               uint64            `json:"fee"`
	PreBalances       []uint64          `json:"preBalances"`
	PostBalances      []uint64          `json:"postBalances"`
	InnerInstructions *[]gjInner        `json:"innerInstructions"`
	LogMessages       *[]string         `json:"logMessages"`
	PreTokenBalances  *[]gjTokenBalance `json:"preTokenBalances"`
	PostTokenBalances *[]gjTokenBalance `json:"postTokenBalances"`
	Rewards           *[]gjReward       `json:"rewards"`
	LoadedAddresses   *struct {
		Writable []string `json:"writable"`
		Readonly []string `json:"readonly"`
	} `json:"loadedAddresses"`
	ReturnData *struct {
		ProgramId string   `json:"programId"`
		Data      []string `json:"data"`
	} `json:"returnData"`
	ComputeUnitsConsumed *uint64 `json:"computeUnitsConsumed"`
	CostUnits            *uint64 `json:"costUnits"`
}

type gjInner struct {
	Index        uint32 `json:"index"`
	Instructions []gjIx `json:"instructions"`
}

type gjTokenBalance struct {
	AccountIndex  uint32  `json:"accountIndex"`
	Mint          string  `json:"mint"`
	Owner         *string `json:"owner"`
	ProgramId     *string `json:"programId"`
	UiTokenAmount struct {
		Amount         string  `json:"amount"`
		Decimals       uint32  `json:"decimals"`
		UiAmountString *string `json:"uiAmountString"`
	} `json:"uiTokenAmount"`
}

type gjReward struct {
	Pubkey      string  `json:"pubkey"`
	Lamports    int64   `json:"lamports"`
	PostBalance *uint64 `json:"postBalance"`
	RewardType  *string `json:"rewardType"`
	Commission  *uint32 `json:"commission"`
}

// gpTx picks just the instruction containers out of the jsonParsed side, for
// attaching CompiledInstruction.parsed.
type gpTx struct {
	Transaction struct {
		Message struct {
			Instructions []json.RawMessage `json:"instructions"`
		} `json:"message"`
	} `json:"transaction"`
	Meta struct {
		InnerInstructions []struct {
			Index        uint32            `json:"index"`
			Instructions []json.RawMessage `json:"instructions"`
		} `json:"innerInstructions"`
	} `json:"meta"`
}

func loadGoldenFixture(t *testing.T) *goldenFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "parsed_golden.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	var g goldenFixture
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatalf("decoding fixture: %v", err)
	}
	if len(g.Json.Transactions) == 0 || len(g.Json.Transactions) != len(g.JsonParsed.Transactions) {
		t.Fatalf("fixture has %d json vs %d jsonParsed transactions", len(g.Json.Transactions), len(g.JsonParsed.Transactions))
	}
	return &g
}

// ---------------------------------------------------------------------------
// Reversal: fixture json side -> protos.
// ---------------------------------------------------------------------------

func b58(t *testing.T, s string) []byte {
	t.Helper()
	b, err := Base58Decode(s)
	if err != nil {
		t.Fatalf("base58 %q: %v", s, err)
	}
	return b
}

// indexBytes repacks numeric account indexes into the proto's byte form.
func indexBytes(idxs []uint32) []byte {
	out := make([]byte, len(idxs))
	for i, v := range idxs {
		out[i] = byte(v)
	}
	return out
}

func buildInstruction(t *testing.T, ix *gjIx) *CompiledInstruction {
	t.Helper()
	return &CompiledInstruction{
		ProgramIdIndex: ix.ProgramIdIndex,
		Accounts:       indexBytes(ix.Accounts),
		Data:           b58(t, ix.Data),
		StackHeight:    ix.StackHeight,
	}
}

func versionString(t *testing.T, raw json.RawMessage) *string {
	t.Helper()
	if len(raw) == 0 {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return &s // "legacy"
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		s = n.String()
		return &s
	}
	t.Fatalf("unrecognized version %s", raw)
	return nil
}

func buildConfirmedTransaction(t *testing.T, raw json.RawMessage) *ConfirmedTransaction {
	t.Helper()
	var tx gjTx
	if err := json.Unmarshal(raw, &tx); err != nil {
		t.Fatalf("decoding json transaction: %v", err)
	}

	gm := &tx.Transaction.Message
	msg := &Message{RecentBlockhash: b58(t, gm.RecentBlockhash)}
	if h := gm.Header; h != nil {
		msg.Header = &MessageHeader{
			NumRequiredSignatures:       h.NumRequiredSignatures,
			NumReadonlySignedAccounts:   h.NumReadonlySignedAccounts,
			NumReadonlyUnsignedAccounts: h.NumReadonlyUnsignedAccounts,
		}
	}
	for _, k := range gm.AccountKeys {
		msg.AccountKeys = append(msg.AccountKeys, b58(t, k))
	}
	for i := range gm.Instructions {
		msg.Instructions = append(msg.Instructions, buildInstruction(t, &gm.Instructions[i]))
	}
	for _, l := range gm.AddressTableLookups {
		msg.AddressTableLookups = append(msg.AddressTableLookups, &AddressTableLookup{
			AccountKey:      b58(t, l.AccountKey),
			WritableIndexes: indexBytes(l.WritableIndexes),
			ReadonlyIndexes: indexBytes(l.ReadonlyIndexes),
		})
	}

	ct := &ConfirmedTransaction{
		Transaction: &Transaction{Message: msg},
		Version:     versionString(t, tx.Version),
	}
	for _, s := range tx.Transaction.Signatures {
		ct.Transaction.Signatures = append(ct.Transaction.Signatures, b58(t, s))
	}
	if tx.Meta != nil {
		ct.Meta = buildMeta(t, tx.Meta)
	}
	return ct
}

func buildMeta(t *testing.T, m *gjMeta) *TransactionStatusMeta {
	t.Helper()
	meta := &TransactionStatusMeta{
		Fee:                  m.Fee,
		PreBalances:          m.PreBalances,
		PostBalances:         m.PostBalances,
		ComputeUnitsConsumed: m.ComputeUnitsConsumed,
		CostUnits:            m.CostUnits,
	}
	if len(m.Err) > 0 && string(m.Err) != "null" {
		var buf bytes.Buffer
		if err := json.Compact(&buf, m.Err); err != nil {
			t.Fatalf("compacting err: %v", err)
		}
		meta.Err = str(buf.String())
	}
	if m.InnerInstructions == nil {
		meta.InnerInstructionsNone = true
	} else {
		for _, g := range *m.InnerInstructions {
			inner := &InnerInstructions{Index: g.Index}
			for i := range g.Instructions {
				inner.Instructions = append(inner.Instructions, buildInstruction(t, &g.Instructions[i]))
			}
			meta.InnerInstructions = append(meta.InnerInstructions, inner)
		}
	}
	if m.LogMessages == nil {
		meta.LogMessagesNone = true
	} else {
		meta.LogMessages = *m.LogMessages
	}
	if m.PreTokenBalances == nil {
		meta.PreTokenBalancesNone = true
	} else {
		meta.PreTokenBalances = buildTokenBalances(t, *m.PreTokenBalances)
	}
	if m.PostTokenBalances == nil {
		meta.PostTokenBalancesNone = true
	} else {
		meta.PostTokenBalances = buildTokenBalances(t, *m.PostTokenBalances)
	}
	if m.Rewards == nil {
		meta.RewardsNone = true
	} else {
		for _, r := range *m.Rewards {
			meta.Rewards = append(meta.Rewards, &Reward{
				Pubkey:      b58(t, r.Pubkey),
				Lamports:    r.Lamports,
				PostBalance: r.PostBalance,
				RewardType:  r.RewardType,
				Commission:  r.Commission,
			})
		}
	}
	if m.LoadedAddresses != nil {
		for _, k := range m.LoadedAddresses.Writable {
			meta.LoadedWritableAddresses = append(meta.LoadedWritableAddresses, b58(t, k))
		}
		for _, k := range m.LoadedAddresses.Readonly {
			meta.LoadedReadonlyAddresses = append(meta.LoadedReadonlyAddresses, b58(t, k))
		}
	}
	if m.ReturnData != nil {
		if len(m.ReturnData.Data) != 2 || m.ReturnData.Data[1] != "base64" {
			t.Fatalf("returnData tuple = %v, want [payload, base64]", m.ReturnData.Data)
		}
		payload, err := base64.StdEncoding.DecodeString(m.ReturnData.Data[0])
		if err != nil {
			t.Fatalf("returnData payload: %v", err)
		}
		meta.ReturnData = &ReturnData{ProgramId: b58(t, m.ReturnData.ProgramId), Data: payload}
	}
	return meta
}

func buildTokenBalances(t *testing.T, tbs []gjTokenBalance) []*TokenBalance {
	t.Helper()
	out := make([]*TokenBalance, 0, len(tbs))
	for _, tb := range tbs {
		p := &TokenBalance{
			AccountIndex: tb.AccountIndex,
			Mint:         b58(t, tb.Mint),
			UiTokenAmount: &UiTokenAmount{
				Amount:         tb.UiTokenAmount.Amount,
				Decimals:       tb.UiTokenAmount.Decimals,
				UiAmountString: tb.UiTokenAmount.UiAmountString,
			},
		}
		if tb.Owner != nil {
			p.Owner = b58(t, *tb.Owner)
		}
		if tb.ProgramId != nil {
			p.ProgramId = b58(t, *tb.ProgramId)
		}
		out = append(out, p)
	}
	return out
}

// spliceParsed attaches CompiledInstruction.parsed for every instruction the
// node rendered in parsed form — the compact bytes of the whole jsonParsed
// instruction object {program, programId, parsed, stackHeight}, exactly what
// prism attaches via Agave's parser crate. partiallyDecoded instructions
// (no "parsed" key) get none.
func spliceParsed(t *testing.T, ct *ConfirmedTransaction, parsedRaw json.RawMessage) {
	t.Helper()
	var p gpTx
	if err := json.Unmarshal(parsedRaw, &p); err != nil {
		t.Fatalf("decoding jsonParsed transaction: %v", err)
	}
	top := p.Transaction.Message.Instructions
	if got, want := len(ct.Transaction.Message.Instructions), len(top); got != want {
		t.Fatalf("instruction count mismatch: json %d vs jsonParsed %d", got, want)
	}
	for i, raw := range top {
		attachParsed(t, ct.Transaction.Message.Instructions[i], raw)
	}
	if ct.Meta == nil {
		return
	}
	if got, want := len(ct.Meta.InnerInstructions), len(p.Meta.InnerInstructions); got != want {
		t.Fatalf("inner instruction group mismatch: json %d vs jsonParsed %d", got, want)
	}
	for g, group := range p.Meta.InnerInstructions {
		pg := ct.Meta.InnerInstructions[g]
		if pg.Index != group.Index || len(pg.Instructions) != len(group.Instructions) {
			t.Fatalf("inner group %d mismatch between fixture sides", g)
		}
		for k, raw := range group.Instructions {
			attachParsed(t, pg.Instructions[k], raw)
		}
	}
}

func attachParsed(t *testing.T, ci *CompiledInstruction, raw json.RawMessage) {
	t.Helper()
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("decoding jsonParsed instruction: %v", err)
	}
	if _, isParsed := obj["parsed"]; !isParsed {
		return
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		t.Fatalf("compacting parsed instruction: %v", err)
	}
	ci.Parsed = buf.Bytes()
}

// ---------------------------------------------------------------------------
// Normalization and structural diff.
// ---------------------------------------------------------------------------

// jsonNormalize forces a rendered value through encoding/json so both sides
// of the golden comparison carry the same dynamic types (json.Number for all
// numerics, plain maps and slices for the rest).
func jsonNormalize(t *testing.T, v interface{}) interface{} {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshaling rendered value: %v", err)
	}
	return decodeNumbers(t, b)
}

func decodeNumbers(t *testing.T, b []byte) interface{} {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var out interface{}
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("decoding for comparison: %v", err)
	}
	return out
}

// firstDiff walks want/got and reports the JSON pointer of the first
// difference, or ok=true when deep-equal. Integer literals compare exactly
// (u64 magnitudes never round through float64); anything with a fraction or
// exponent compares by float64 value so formatting variance ("1.0" vs "1")
// does not count as a difference.
func firstDiff(path string, want, got interface{}) (string, string, bool) {
	switch w := want.(type) {
	case map[string]interface{}:
		g, ok := got.(map[string]interface{})
		if !ok {
			return path, fmt.Sprintf("want object, got %s", describeJson(got)), false
		}
		keySet := make(map[string]struct{}, len(w)+len(g))
		for k := range w {
			keySet[k] = struct{}{}
		}
		for k := range g {
			keySet[k] = struct{}{}
		}
		keys := make([]string, 0, len(keySet))
		for k := range keySet {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			wv, inW := w[k]
			gv, inG := g[k]
			if !inW {
				return path + "/" + k, fmt.Sprintf("unexpected key, got %s", describeJson(gv)), false
			}
			if !inG {
				return path + "/" + k, fmt.Sprintf("missing key, want %s", describeJson(wv)), false
			}
			if p, d, ok := firstDiff(path+"/"+k, wv, gv); !ok {
				return p, d, false
			}
		}
		return "", "", true
	case []interface{}:
		g, ok := got.([]interface{})
		if !ok {
			return path, fmt.Sprintf("want array, got %s", describeJson(got)), false
		}
		for i := 0; i < len(w) && i < len(g); i++ {
			if p, d, ok := firstDiff(path+"/"+strconv.Itoa(i), w[i], g[i]); !ok {
				return p, d, false
			}
		}
		if len(w) != len(g) {
			return path, fmt.Sprintf("array length %d, want %d", len(g), len(w)), false
		}
		return "", "", true
	case json.Number:
		g, ok := got.(json.Number)
		if !ok || !numbersEqual(w, g) {
			return path, fmt.Sprintf("want %s, got %s", w, describeJson(got)), false
		}
		return "", "", true
	default:
		if want != got {
			return path, fmt.Sprintf("want %s, got %s", describeJson(want), describeJson(got)), false
		}
		return "", "", true
	}
}

func numbersEqual(a, b json.Number) bool {
	if a.String() == b.String() {
		return true
	}
	if isIntegralLiteral(a.String()) && isIntegralLiteral(b.String()) {
		return false // both exact integers with different digits
	}
	af, aerr := a.Float64()
	bf, berr := b.Float64()
	return aerr == nil && berr == nil && af == bf
}

func isIntegralLiteral(s string) bool {
	return !strings.ContainsAny(s, ".eE")
}

func describeJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%#v", v)
	}
	if len(b) > 160 {
		b = append(b[:160], []byte("...")...)
	}
	return string(b)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// Golden parity: reversing the node's json output into protos, attaching the
// node's parsed instructions, and rendering jsonParsed must reproduce the
// node's jsonParsed output byte-for-byte (modulo numeric formatting).
func TestParsedGoldenTransactionParity(t *testing.T) {
	g := loadGoldenFixture(t)
	for i := range g.Json.Transactions {
		t.Run(fmt.Sprintf("tx%d", i), func(t *testing.T) {
			ct := buildConfirmedTransaction(t, g.Json.Transactions[i])
			spliceParsed(t, ct, g.JsonParsed.Transactions[i])
			got := jsonNormalize(t, ConfirmedTransactionToJsonRpcParsed(ct))
			want := decodeNumbers(t, g.JsonParsed.Transactions[i])
			if p, d, ok := firstDiff("", want, got); !ok {
				t.Fatalf("first difference at %q: %s", p, d)
			}
		})
	}
}

// Envelope rules, asserted independently of the golden and in BOTH encodings:
// parsed meta omits loadedAddresses, parsed message omits header, and
// addressTableLookups presence keys on version (v0 keeps it even when empty,
// legacy omits it).
func TestParsedEnvelopeRules(t *testing.T) {
	g := loadGoldenFixture(t)
	var sawLegacy, sawEmptyLookupsV0 bool
	for i := range g.Json.Transactions {
		ct := buildConfirmedTransaction(t, g.Json.Transactions[i])
		parsed := ConfirmedTransactionToJsonRpcParsed(ct)
		plain := ConfirmedTransactionToJsonRpc(ct)
		pmsg := messageOf(t, i, parsed)
		jmsg := messageOf(t, i, plain)

		if _, has := pmsg["header"]; has {
			t.Errorf("tx%d: parsed message must omit header", i)
		}
		if _, has := jmsg["header"]; !has {
			t.Errorf("tx%d: json message must keep header", i)
		}

		if ct.Meta != nil {
			pmeta, _ := parsed["meta"].(map[string]interface{})
			jmeta, _ := plain["meta"].(map[string]interface{})
			if _, has := pmeta["loadedAddresses"]; has {
				t.Errorf("tx%d: parsed meta must omit loadedAddresses", i)
			}
			if _, has := jmeta["loadedAddresses"]; !has {
				t.Errorf("tx%d: json meta must keep loadedAddresses", i)
			}
		}

		switch {
		case ct.Version == nil:
			t.Errorf("tx%d: fixture transaction carries no version", i)
		case *ct.Version == "legacy":
			sawLegacy = true
			for _, m := range []map[string]interface{}{pmsg, jmsg} {
				if _, has := m["addressTableLookups"]; has {
					t.Errorf("tx%d: legacy message must omit addressTableLookups", i)
				}
			}
		default:
			empty := len(ct.Transaction.Message.AddressTableLookups) == 0
			for _, m := range []map[string]interface{}{pmsg, jmsg} {
				v, has := m["addressTableLookups"]
				if !has {
					t.Errorf("tx%d: v0 message must keep addressTableLookups even when empty", i)
					continue
				}
				if l, ok := v.([]interface{}); ok && empty && len(l) != 0 {
					t.Errorf("tx%d: addressTableLookups = %v, want empty", i, l)
				}
			}
			if empty {
				sawEmptyLookupsV0 = true
			}
		}
	}
	if !sawLegacy {
		t.Error("fixture lost its legacy transaction; the legacy envelope rule went unexercised")
	}
	if !sawEmptyLookupsV0 {
		t.Error("fixture lost its empty-lookups v0 transaction; the v0 envelope rule went unexercised")
	}
}

func messageOf(t *testing.T, i int, res map[string]interface{}) map[string]interface{} {
	t.Helper()
	tx, ok := res["transaction"].(map[string]interface{})
	if !ok {
		t.Fatalf("tx%d: transaction = %#v, want object", i, res["transaction"])
	}
	msg, ok := tx["message"].(map[string]interface{})
	if !ok {
		t.Fatalf("tx%d: message = %#v, want object", i, tx["message"])
	}
	return msg
}

// The fixture block has loaded addresses only in one large transaction; this
// synthetic v0 transaction pins the merged-key contract in isolation: order
// static ++ loadedWritable ++ loadedReadonly, header-derived flags on static
// keys, list-derived writability with source "lookupTable" on loaded keys,
// partiallyDecoded account indexes resolving into the loaded range, and
// loadedAddresses present under json but absent under jsonParsed for the SAME
// meta.
func TestParsedLookupTableKeys(t *testing.T) {
	key := func(fill byte) []byte { return bytes.Repeat([]byte{fill}, 32) }
	static := [][]byte{key(1), key(2), key(3), key(10)}
	w0, w1, r0 := key(4), key(5), key(6)

	ct := &ConfirmedTransaction{
		Version: str("0"),
		Transaction: &Transaction{
			Signatures: [][]byte{bytes.Repeat([]byte{9}, 64), bytes.Repeat([]byte{8}, 64)},
			Message: &Message{
				// All four static classes: index 0 writable signer
				// (0 < 2-1), index 1 readonly signer, index 2 writable
				// non-signer (2 < 4-1), index 3 readonly non-signer.
				Header: &MessageHeader{
					NumRequiredSignatures:       2,
					NumReadonlySignedAccounts:   1,
					NumReadonlyUnsignedAccounts: 1,
				},
				AccountKeys:     static,
				RecentBlockhash: key(7),
				Instructions: []*CompiledInstruction{{
					ProgramIdIndex: 3,
					Accounts:       []byte{4, 5, 6}, // entirely in the loaded range
					Data:           []byte{0xaa},
					StackHeight:    u32(1),
				}},
			},
		},
		Meta: &TransactionStatusMeta{
			LoadedWritableAddresses: [][]byte{w0, w1},
			LoadedReadonlyAddresses: [][]byte{r0},
		},
	}

	res := ConfirmedTransactionToJsonRpcParsed(ct)
	msg := messageOf(t, 0, res)
	keys, ok := msg["accountKeys"].([]interface{})
	if !ok {
		t.Fatalf("accountKeys = %#v, want array", msg["accountKeys"])
	}
	wantKeys := []struct {
		pubkey           string
		signer, writable bool
		source           string
	}{
		{Base58Encode(static[0]), true, true, "transaction"},
		{Base58Encode(static[1]), true, false, "transaction"},
		{Base58Encode(static[2]), false, true, "transaction"},
		{Base58Encode(static[3]), false, false, "transaction"},
		{Base58Encode(w0), false, true, "lookupTable"},
		{Base58Encode(w1), false, true, "lookupTable"},
		{Base58Encode(r0), false, false, "lookupTable"},
	}
	if len(keys) != len(wantKeys) {
		t.Fatalf("accountKeys has %d entries, want %d (static ++ writable ++ readonly)", len(keys), len(wantKeys))
	}
	for i, want := range wantKeys {
		got, ok := keys[i].(map[string]interface{})
		if !ok {
			t.Fatalf("accountKeys[%d] = %#v, want object", i, keys[i])
		}
		if got["pubkey"] != want.pubkey || got["signer"] != want.signer ||
			got["writable"] != want.writable || got["source"] != want.source {
			t.Fatalf("accountKeys[%d] = %#v, want %+v", i, got, want)
		}
	}

	instrs, _ := msg["instructions"].([]interface{})
	ix, ok := instrs[0].(map[string]interface{})
	if !ok {
		t.Fatalf("instructions[0] = %#v, want object", instrs[0])
	}
	if ix["programId"] != Base58Encode(static[3]) {
		t.Fatalf("programId = %v, want %v", ix["programId"], Base58Encode(static[3]))
	}
	accs, _ := ix["accounts"].([]interface{})
	wantAccs := []interface{}{Base58Encode(w0), Base58Encode(w1), Base58Encode(r0)}
	if len(accs) != len(wantAccs) {
		t.Fatalf("accounts = %#v, want %v", accs, wantAccs)
	}
	for i := range wantAccs {
		if accs[i] != wantAccs[i] {
			t.Fatalf("accounts[%d] = %v, want %v (index %d in the loaded range)", i, accs[i], wantAccs[i], i+4)
		}
	}

	pmeta, ok := res["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("meta = %#v, want object", res["meta"])
	}
	if _, has := pmeta["loadedAddresses"]; has {
		t.Fatal("parsed meta must omit loadedAddresses (keys are merged into accountKeys)")
	}
	jmeta := MetaToJsonRpc(ct.Meta)
	la, ok := jmeta["loadedAddresses"].(map[string]interface{})
	if !ok {
		t.Fatalf("json meta loadedAddresses = %#v, want object", jmeta["loadedAddresses"])
	}
	writable, _ := la["writable"].([]interface{})
	readonly, _ := la["readonly"].([]interface{})
	if len(writable) != 2 || writable[0] != Base58Encode(w0) || writable[1] != Base58Encode(w1) {
		t.Fatalf("loadedAddresses.writable = %#v, want [%s %s]", writable, Base58Encode(w0), Base58Encode(w1))
	}
	if len(readonly) != 1 || readonly[0] != Base58Encode(r0) {
		t.Fatalf("loadedAddresses.readonly = %#v, want [%s]", readonly, Base58Encode(r0))
	}
}

// A malformed parsed attachment must fall back to partiallyDecoded, not drop
// the instruction or leak the bad bytes.
func TestParsedSpliceFallbackOnInvalidBytes(t *testing.T) {
	keys := []string{Base58Encode(bytes.Repeat([]byte{1}, 32)), Base58Encode(bytes.Repeat([]byte{2}, 32))}
	res := parsedInstructionToJsonRpc(&CompiledInstruction{
		ProgramIdIndex: 0,
		Accounts:       []byte{1},
		Data:           []byte{0x01, 0x02},
		StackHeight:    u32(2),
		Parsed:         []byte(`{"program": not json`),
	}, keys)

	if _, has := res["parsed"]; has {
		t.Fatalf("result = %#v, malformed bytes must not surface a parsed key", res)
	}
	if res["programId"] != keys[0] {
		t.Fatalf("programId = %v, want %v", res["programId"], keys[0])
	}
	accs, _ := res["accounts"].([]interface{})
	if len(accs) != 1 || accs[0] != keys[1] {
		t.Fatalf("accounts = %#v, want [%v]", accs, keys[1])
	}
	if res["data"] != Base58Encode([]byte{0x01, 0x02}) {
		t.Fatalf("data = %v, want base58 of the raw bytes", res["data"])
	}
	if res["stackHeight"] != uint32(2) {
		t.Fatalf("stackHeight = %#v, want 2", res["stackHeight"])
	}
}

// An out-of-range account index renders as "" rather than panicking or
// substituting a wrong key.
func TestParsedKeyAtOutOfRange(t *testing.T) {
	keys := []string{Base58Encode(bytes.Repeat([]byte{1}, 32))}
	res := parsedInstructionToJsonRpc(&CompiledInstruction{
		ProgramIdIndex: 250,
		Accounts:       []byte{0, 200},
	}, keys)

	if res["programId"] != "" {
		t.Fatalf("programId = %q, want empty string for out-of-range index", res["programId"])
	}
	accs, _ := res["accounts"].([]interface{})
	if len(accs) != 2 || accs[0] != keys[0] || accs[1] != "" {
		t.Fatalf("accounts = %#v, want [%v \"\"]", accs, keys[0])
	}
}
