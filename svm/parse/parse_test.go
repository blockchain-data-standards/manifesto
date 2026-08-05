package parse_test

// Layer 2: hand-built wire vectors for the arms the captured block never
// exercises. Wire recipes, per program:
//
//	system   bincode: u32 LE discriminant, u64 LE ints, raw 32-byte pubkeys,
//	         strings as u64 LE length + utf8. Exact consumption required.
//	token    u8 discriminant, u64 LE amounts, COption<Pubkey> = 0x00 | 0x01+32.
//	         Trailing bytes tolerated (spl unpack_u64 discards _rest).
//	ata      borsh unit enum: EMPTY data = create, else exactly one byte.
//	memo     raw utf8; the parsed form is a BARE JSON string.
//
// Every expected `parsed` value is written as a JSON literal and compared
// decoded, so key order and 0-vs-0.0 never matter, and every failure prints
// want-vs-got JSON.

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/blockchain-data-standards/manifesto/svm"
	"github.com/blockchain-data-standards/manifesto/svm/parse"
)

// ---------------------------------------------------------------------------
// Wire-building helpers.
// ---------------------------------------------------------------------------

func le32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

func le64(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return b
}

func cat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// kb is a deterministic 32-byte pubkey (every byte = b); k is its base58.
func kb(b byte) []byte {
	out := make([]byte, 32)
	for i := range out {
		out[i] = b
	}
	return out
}

func k(b byte) string { return svm.Base58Encode(kb(b)) }

// bincodeStr is bincode's string layout: u64 LE byte length + utf8 bytes.
func bincodeStr(s string) []byte { return cat(le64(uint64(len(s))), []byte(s)) }

func u32p(v uint32) *uint32 { return &v }

// accs builds an account list from key seeds, in instruction order.
func accs(seeds ...byte) []string {
	out := make([]string, len(seeds))
	for i, s := range seeds {
		out[i] = k(s)
	}
	return out
}

// ---------------------------------------------------------------------------
// Positive vectors.
// ---------------------------------------------------------------------------

type vector struct {
	name      string
	programID string
	data      []byte
	accounts  []string
	program   string // envelope "program"
	parsed    string // expected "parsed", as a JSON literal
}

func runVectors(t *testing.T, vectors []vector) {
	t.Helper()
	for _, v := range vectors {
		env, err := parse.Parse(v.programID, v.data, v.accounts, nil)
		if err != nil {
			t.Errorf("%s: Parse() error = %v, want success", v.name, err)
			continue
		}
		if env.Program != v.program {
			t.Errorf("%s: program = %q, want %q", v.name, env.Program, v.program)
		}
		if env.ProgramID != v.programID {
			t.Errorf("%s: programId = %q, want %q", v.name, env.ProgramID, v.programID)
		}
		if env.StackHeight != nil {
			t.Errorf("%s: stackHeight = %d, want nil passthrough", v.name, *env.StackHeight)
		}
		want := decodeJSON(t, []byte(v.parsed))
		got := decodeJSON(t, env.Parsed)
		if path, diff, ok := firstDiff("", want, got); !ok {
			t.Errorf("%s: parsed mismatch at %q: %s\nwant: %s\ngot:  %s",
				v.name, path, diff, v.parsed, env.Parsed)
		}
	}
}

func TestSystemVectors(t *testing.T) {
	// A multibyte utf8 seed round-trips: bincode length is BYTES, not runes.
	seed := "søl:seed-∂"
	runVectors(t, []vector{
		{
			name:      "createAccount",
			programID: parse.SystemProgramID,
			data:      cat(le32(0), le64(1_000_000), le64(165), kb(9)),
			accounts:  accs(1, 2),
			program:   "system",
			parsed: fmt.Sprintf(`{"type":"createAccount","info":{
				"source":%q,"newAccount":%q,"lamports":1000000,"space":165,"owner":%q}}`,
				k(1), k(2), k(9)),
		},
		{
			name:      "createAccountWithSeed",
			programID: parse.SystemProgramID,
			data:      cat(le32(3), kb(7), bincodeStr(seed), le64(2_039_280), le64(165), kb(8)),
			accounts:  accs(1, 2),
			program:   "system",
			parsed: fmt.Sprintf(`{"type":"createAccountWithSeed","info":{
				"source":%q,"newAccount":%q,"base":%q,"seed":%q,
				"lamports":2039280,"space":165,"owner":%q}}`,
				k(1), k(2), k(7), seed, k(8)),
		},
		{
			name:      "withdrawFromNonce",
			programID: parse.SystemProgramID,
			data:      cat(le32(5), le64(42)),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "system",
			parsed: fmt.Sprintf(`{"type":"withdrawFromNonce","info":{
				"nonceAccount":%q,"destination":%q,"recentBlockhashesSysvar":%q,
				"rentSysvar":%q,"nonceAuthority":%q,"lamports":42}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
		{
			name:      "transferWithSeed",
			programID: parse.SystemProgramID,
			data:      cat(le32(11), le64(123), bincodeStr("seed"), kb(6)),
			accounts:  accs(1, 2, 3),
			program:   "system",
			parsed: fmt.Sprintf(`{"type":"transferWithSeed","info":{
				"source":%q,"sourceBase":%q,"destination":%q,
				"lamports":123,"sourceSeed":"seed","sourceOwner":%q}}`,
				k(1), k(2), k(3), k(6)),
		},
		{
			name:      "upgradeNonce",
			programID: parse.SystemProgramID,
			data:      le32(12),
			accounts:  accs(1),
			program:   "system",
			parsed:    fmt.Sprintf(`{"type":"upgradeNonce","info":{"nonceAccount":%q}}`, k(1)),
		},
		{
			// bincode trailing-bytes tolerance, system family (bincode
			// 1.3.3 allow_trailing_bytes).
			name:      "transfer trailing bytes tolerated",
			programID: parse.SystemProgramID,
			data:      cat(le32(2), le64(9), []byte{0xAA, 0xBB}),
			accounts:  accs(1, 2),
			program:   "system",
			parsed: fmt.Sprintf(`{"type":"transfer","info":{
				"source":%q,"destination":%q,"lamports":9}}`, k(1), k(2)),
		},
	})
}

func TestTokenVectors(t *testing.T) {
	runVectors(t, []vector{
		// freezeAuthority present vs ABSENT: absent means the key is omitted
		// entirely (Agave map_coption_pubkey semantics), never present-null.
		// firstDiff flags any extra key, so the absent case pins the omission.
		{
			name:      "initializeMint freezeAuthority present",
			programID: parse.TokenProgramID,
			data:      cat([]byte{0, 6}, kb(4), []byte{1}, kb(5)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeMint","info":{
				"mint":%q,"decimals":6,"mintAuthority":%q,"rentSysvar":%q,"freezeAuthority":%q}}`,
				k(1), k(4), k(2), k(5)),
		},
		{
			name:      "initializeMint freezeAuthority absent omits key",
			programID: parse.TokenProgramID,
			data:      cat([]byte{0, 0}, kb(4), []byte{0}),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeMint","info":{
				"mint":%q,"decimals":0,"mintAuthority":%q,"rentSysvar":%q}}`,
				k(1), k(4), k(2)),
		},
		{
			name:      "initializeMint2 drops rentSysvar",
			programID: parse.TokenProgramID,
			data:      cat([]byte{20, 9}, kb(4), []byte{0}),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeMint2","info":{
				"mint":%q,"decimals":9,"mintAuthority":%q}}`, k(1), k(4)),
		},
		// setAuthority: the first account's key name depends on the authority
		// class (mint authorities name "mint", account authorities name
		// "account"), and newAuthority is PRESENT-and-null when unset —
		// the opposite presence rule from freezeAuthority above.
		{
			name:      "setAuthority mintTokens newAuthority null",
			programID: parse.TokenProgramID,
			data:      []byte{6, 0, 0},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"setAuthority","info":{
				"mint":%q,"authorityType":"mintTokens","newAuthority":null,"authority":%q}}`,
				k(1), k(2)),
		},
		{
			name:      "setAuthority accountOwner owns account",
			programID: parse.TokenProgramID,
			data:      []byte{6, 2, 0},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"setAuthority","info":{
				"account":%q,"authorityType":"accountOwner","newAuthority":null,"authority":%q}}`,
				k(1), k(2)),
		},
		{
			name:      "setAuthority closeAccount set + multisig",
			programID: parse.TokenProgramID,
			data:      cat([]byte{6, 3, 1}, kb(5)),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"setAuthority","info":{
				"account":%q,"authorityType":"closeAccount","newAuthority":%q,
				"multisigAuthority":%q,"signers":[%q,%q]}}`,
				k(1), k(5), k(2), k(3), k(4)),
		},
		// approve/revoke: accounts beyond the owner position flip the field
		// to multisig* plus a signers array — positions, not signature flags.
		{
			name:      "approve single owner",
			programID: parse.TokenProgramID,
			data:      cat([]byte{4}, le64(555)),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"approve","info":{
				"source":%q,"delegate":%q,"amount":"555","owner":%q}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "approve multisig owner",
			programID: parse.TokenProgramID,
			data:      cat([]byte{4}, le64(555)),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"approve","info":{
				"source":%q,"delegate":%q,"amount":"555",
				"multisigOwner":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
		{
			name:      "revoke single owner",
			programID: parse.TokenProgramID,
			data:      []byte{5},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed:    fmt.Sprintf(`{"type":"revoke","info":{"source":%q,"owner":%q}}`, k(1), k(2)),
		},
		{
			name:      "revoke multisig owner",
			programID: parse.TokenProgramID,
			data:      []byte{5},
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"revoke","info":{
				"source":%q,"multisigOwner":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4)),
		},
		// tokenAmount object: uiAmount is a NUMBER, amount a STRING, and
		// uiAmountString the exact trimmed decimal (string math, no float).
		{
			name:      "transferChecked 1500000/6 -> 1.5",
			programID: parse.TokenProgramID,
			data:      cat([]byte{12}, le64(1_500_000), []byte{6}),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"transferChecked","info":{
				"source":%q,"mint":%q,"destination":%q,"authority":%q,
				"tokenAmount":{"uiAmount":1.5,"decimals":6,"amount":"1500000","uiAmountString":"1.5"}}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			// Same wire under the token-2022 program id: identical layout,
			// different programId in the envelope (runVectors checks it).
			name:      "transferChecked under token-2022",
			programID: parse.Token2022ID,
			data:      cat([]byte{12}, le64(1_500_000), []byte{6}),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"transferChecked","info":{
				"source":%q,"mint":%q,"destination":%q,"authority":%q,
				"tokenAmount":{"uiAmount":1.5,"decimals":6,"amount":"1500000","uiAmountString":"1.5"}}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			// Zero amount: uiAmount 0, NOT null — instruction tokenAmounts use
			// Agave's Some(0.0) path, unlike meta token balances. Pinned here
			// so nobody "unifies" the two rules.
			name:      "transferChecked zero amount",
			programID: parse.TokenProgramID,
			data:      cat([]byte{12}, le64(0), []byte{6}),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"transferChecked","info":{
				"source":%q,"mint":%q,"destination":%q,"authority":%q,
				"tokenAmount":{"uiAmount":0,"decimals":6,"amount":"0","uiAmountString":"0"}}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "mintToChecked",
			programID: parse.TokenProgramID,
			data:      cat([]byte{14}, le64(250), []byte{2}),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"mintToChecked","info":{
				"mint":%q,"account":%q,"mintAuthority":%q,
				"tokenAmount":{"uiAmount":2.5,"decimals":2,"amount":"250","uiAmountString":"2.5"}}}`,
				k(1), k(2), k(3)),
		},
		{
			// 1000000/6: fraction trims away entirely — "1", never "1.000000".
			name:      "burnChecked trims to integer string",
			programID: parse.TokenProgramID,
			data:      cat([]byte{15}, le64(1_000_000), []byte{6}),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"burnChecked","info":{
				"account":%q,"mint":%q,"authority":%q,
				"tokenAmount":{"uiAmount":1,"decimals":6,"amount":"1000000","uiAmountString":"1"}}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "initializeMultisig carries m byte",
			programID: parse.TokenProgramID,
			data:      []byte{2, 2},
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeMultisig","info":{
				"multisig":%q,"rentSysvar":%q,"signers":[%q,%q],"m":2}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "syncNative",
			programID: parse.TokenProgramID,
			data:      []byte{17},
			accounts:  accs(1),
			program:   "spl-token",
			parsed:    fmt.Sprintf(`{"type":"syncNative","info":{"account":%q}}`, k(1)),
		},
		{
			name:      "getAccountDataSize no extensions",
			programID: parse.TokenProgramID,
			data:      []byte{21},
			accounts:  accs(1),
			program:   "spl-token",
			parsed:    fmt.Sprintf(`{"type":"getAccountDataSize","info":{"mint":%q}}`, k(1)),
		},
		{
			// amount is a STRING — Agave calls .to_string() on the u64.
			name:      "amountToUiAmount stringifies amount",
			programID: parse.TokenProgramID,
			data:      cat([]byte{23}, le64(4242)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed:    fmt.Sprintf(`{"type":"amountToUiAmount","info":{"mint":%q,"amount":"4242"}}`, k(1)),
		},
		{
			// uiAmount is the raw utf8 payload, passed through verbatim.
			name:      "uiAmountToAmount raw string",
			programID: parse.TokenProgramID,
			data:      append([]byte{24}, []byte("42.42")...),
			accounts:  accs(1),
			program:   "spl-token",
			parsed:    fmt.Sprintf(`{"type":"uiAmountToAmount","info":{"mint":%q,"uiAmount":"42.42"}}`, k(1)),
		},
		{
			name:      "initializeAccount2 keeps rentSysvar",
			programID: parse.TokenProgramID,
			data:      cat([]byte{16}, kb(5)),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeAccount2","info":{
				"account":%q,"mint":%q,"owner":%q,"rentSysvar":%q}}`,
				k(1), k(2), k(5), k(3)),
		},
		{
			name:      "initializeAccount3 drops rentSysvar",
			programID: parse.TokenProgramID,
			data:      cat([]byte{18}, kb(5)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeAccount3","info":{
				"account":%q,"mint":%q,"owner":%q}}`, k(1), k(2), k(5)),
		},
		{
			// spl unpack_u64 takes rest.get(..8): trailing bytes are legal on
			// token instructions — unlike bincode's exact-consumption rule.
			name:      "transfer tolerates trailing bytes",
			programID: parse.TokenProgramID,
			data:      cat([]byte{3}, le64(777), []byte{0xde, 0xad}),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"transfer","info":{
				"source":%q,"destination":%q,"amount":"777","authority":%q}}`,
				k(1), k(2), k(3)),
		},
	})
}

func TestAssociatedTokenVectors(t *testing.T) {
	createParsed := fmt.Sprintf(`{"type":"create","info":{
		"source":%q,"account":%q,"wallet":%q,"mint":%q,"systemProgram":%q,"tokenProgram":%q}}`,
		k(1), k(2), k(3), k(4), k(5), k(6))
	runVectors(t, []vector{
		// EMPTY data is the original pre-idempotent Create deployment quirk.
		{
			name:      "create with empty data",
			programID: parse.AssociatedID,
			data:      nil,
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "spl-associated-token-account",
			parsed:    createParsed,
		},
		{
			name:      "create with explicit discriminant",
			programID: parse.AssociatedID,
			data:      []byte{0},
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "spl-associated-token-account",
			parsed:    createParsed,
		},
		{
			name:      "createIdempotent",
			programID: parse.AssociatedID,
			data:      []byte{1},
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "spl-associated-token-account",
			parsed: fmt.Sprintf(`{"type":"createIdempotent","info":{
				"source":%q,"account":%q,"wallet":%q,"mint":%q,"systemProgram":%q,"tokenProgram":%q}}`,
				k(1), k(2), k(3), k(4), k(5), k(6)),
		},
		{
			name:      "recoverNested",
			programID: parse.AssociatedID,
			data:      []byte{2},
			accounts:  accs(1, 2, 3, 4, 5, 6, 7),
			program:   "spl-associated-token-account",
			parsed: fmt.Sprintf(`{"type":"recoverNested","info":{
				"nestedSource":%q,"nestedMint":%q,"destination":%q,"nestedOwner":%q,
				"ownerMint":%q,"wallet":%q,"tokenProgram":%q}}`,
				k(1), k(2), k(3), k(4), k(5), k(6), k(7)),
		},
	})
}

func TestMemoVectors(t *testing.T) {
	// Memo's parsed form is a BARE JSON string — no {type, info} envelope.
	runVectors(t, []vector{
		{
			name:      "memo v3 utf8",
			programID: parse.MemoV3ID,
			data:      []byte("hello, memo ☀"),
			accounts:  nil,
			program:   "spl-memo",
			parsed:    `"hello, memo ☀"`,
		},
		{
			name:      "memo v1 shares the parser",
			programID: parse.MemoV1ID,
			data:      []byte("gm"),
			accounts:  nil,
			program:   "spl-memo",
			parsed:    `"gm"`,
		},
	})
}

// Envelope stackHeight is a passthrough: nil marshals as an explicit null
// (the key is never omitted), a value marshals as the number.
func TestStackHeightPassthrough(t *testing.T) {
	env, err := parse.Parse(parse.MemoV3ID, []byte("x"), nil, u32p(3))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	obj := decodeJSON(t, raw).(map[string]any)
	if n, ok := obj["stackHeight"].(json.Number); !ok || n.String() != "3" {
		t.Fatalf("stackHeight = %v, want 3", obj["stackHeight"])
	}

	env, err = parse.Parse(parse.MemoV3ID, []byte("x"), nil, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	raw, err = json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	obj = decodeJSON(t, raw).(map[string]any)
	v, present := obj["stackHeight"]
	if !present || v != nil {
		t.Fatalf("stackHeight = %v (present=%v), want explicit null", v, present)
	}
}

// ---------------------------------------------------------------------------
// Negative vectors: every anomaly must refuse with ErrNotParsable — never a
// partial or guessed object. Truncations walk every field boundary of the
// representative arms (system createAccountWithSeed, token initializeMint).
// ---------------------------------------------------------------------------

func TestNotParsableVectors(t *testing.T) {
	cases := []struct {
		name      string
		programID string
		data      []byte
		accounts  []string
	}{
		// Unknown program: phase-1 covers exactly four registries.
		{"unknown program id", k(0x77), le32(2), accs(1, 2)},

		// system: truncated at every boundary of createAccountWithSeed.
		{"system empty data", parse.SystemProgramID, nil, accs(1)},
		{"system truncated discriminant", parse.SystemProgramID, []byte{3, 0, 0}, accs(1, 2)},
		{"system cAWS missing base", parse.SystemProgramID, le32(3), accs(1, 2)},
		{"system cAWS truncated base", parse.SystemProgramID, cat(le32(3), kb(7)[:16]), accs(1, 2)},
		{"system cAWS missing seed length", parse.SystemProgramID, cat(le32(3), kb(7)), accs(1, 2)},
		{"system cAWS seed length exceeds data", parse.SystemProgramID,
			cat(le32(3), kb(7), le64(4), []byte("ab")), accs(1, 2)},
		{"system cAWS non-utf8 seed", parse.SystemProgramID,
			cat(le32(3), kb(7), le64(2), []byte{0xff, 0xfe}, le64(1), le64(2), kb(8)), accs(1, 2)},
		{"system cAWS missing lamports", parse.SystemProgramID,
			cat(le32(3), kb(7), bincodeStr("s")), accs(1, 2)},
		{"system cAWS truncated space", parse.SystemProgramID,
			cat(le32(3), kb(7), bincodeStr("s"), le64(1), le64(2)[:4]), accs(1, 2)},
		{"system cAWS truncated owner", parse.SystemProgramID,
			cat(le32(3), kb(7), bincodeStr("s"), le64(1), le64(2), kb(8)[:31]), accs(1, 2)},

		// Unknown discriminants.
		{"system unknown discriminant 13", parse.SystemProgramID, cat(le32(13)), accs(1)},
		{"token disc 25 bad COption tag falls to TLV and refuses", parse.TokenProgramID,
			cat([]byte{25}, kb(5)), accs(1)},
		{"token unknown discriminant 255", parse.TokenProgramID, []byte{255}, accs(1)},
		{"ata unknown discriminant 3", parse.AssociatedID, []byte{3}, accs(1, 2, 3, 4, 5, 6, 7)},

		// Too-few accounts (check_num_accounts is a minimum).
		{"system transfer 1 account", parse.SystemProgramID, cat(le32(2), le64(1)), accs(1)},
		{"system withdrawFromNonce 4 accounts", parse.SystemProgramID,
			cat(le32(5), le64(1)), accs(1, 2, 3, 4)},
		{"token transferChecked 3 accounts", parse.TokenProgramID,
			cat([]byte{12}, le64(1), []byte{6}), accs(1, 2, 3)},
		{"token revoke 1 account", parse.TokenProgramID, []byte{5}, accs(1)},
		{"token initializeMint 1 account", parse.TokenProgramID,
			cat([]byte{0, 6}, kb(4), []byte{0}), accs(1)},
		{"ata create 5 accounts", parse.AssociatedID, nil, accs(1, 2, 3, 4, 5)},
		{"ata recoverNested 6 accounts", parse.AssociatedID, []byte{2}, accs(1, 2, 3, 4, 5, 6)},

		// token: truncated at every boundary of initializeMint.
		{"token empty data", parse.TokenProgramID, nil, accs(1)},
		{"token initializeMint missing decimals", parse.TokenProgramID, []byte{0}, accs(1, 2)},
		{"token initializeMint truncated authority", parse.TokenProgramID,
			cat([]byte{0, 6}, kb(4)[:31]), accs(1, 2)},
		{"token initializeMint missing COption tag", parse.TokenProgramID,
			cat([]byte{0, 6}, kb(4)), accs(1, 2)},
		{"token initializeMint truncated COption key", parse.TokenProgramID,
			cat([]byte{0, 6}, kb(4), []byte{1}, kb(5)[:31]), accs(1, 2)},
		{"token initializeMint bad COption tag", parse.TokenProgramID,
			cat([]byte{0, 6}, kb(4), []byte{2}, kb(5)), accs(1, 2)},

		// token: other field-boundary and value anomalies.
		{"token transfer truncated amount", parse.TokenProgramID,
			cat([]byte{3}, le64(1)[:7]), accs(1, 2, 3)},
		{"token transferChecked missing decimals", parse.TokenProgramID,
			cat([]byte{12}, le64(1)), accs(1, 2, 3, 4)},
		{"token setAuthority unknown authority type", parse.TokenProgramID,
			[]byte{6, 17, 0}, accs(1, 2)},
		{"token getAccountDataSize odd extension bytes", parse.TokenProgramID,
			[]byte{21, 6}, accs(1)},
		{"token uiAmountToAmount non-utf8", parse.TokenProgramID,
			[]byte{24, 0xff, 0xfe}, accs(1)},

		// ata: borsh exact consumption — anything past one byte refuses.
		{"ata two-byte data", parse.AssociatedID, []byte{0, 0}, accs(1, 2, 3, 4, 5, 6)},
		{"ata idempotent with trailing byte", parse.AssociatedID, []byte{1, 1}, accs(1, 2, 3, 4, 5, 6)},

		// memo: invalid utf8 falls back rather than re-encoding lossily.
		{"memo invalid utf8", parse.MemoV3ID, []byte{0xff, 0xfe, 0xfd}, nil},
	}

	for _, tc := range cases {
		env, err := parse.Parse(tc.programID, tc.data, tc.accounts, nil)
		if err == nil {
			t.Errorf("%s: Parse() succeeded, want ErrNotParsable; parsed: %s", tc.name, env.Parsed)
			continue
		}
		if !errors.Is(err, parse.ErrNotParsable) {
			t.Errorf("%s: error %v does not wrap ErrNotParsable", tc.name, err)
		}
		if env != nil {
			t.Errorf("%s: non-nil envelope alongside error", tc.name)
		}
	}
}
