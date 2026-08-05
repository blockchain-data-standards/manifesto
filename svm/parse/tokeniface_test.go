package parse_test

// TokenGroup / TokenMetadata TLV interface vectors: all 9 instructions.
// Discriminators are the first 8 bytes of sha256 over the exact
// #[discriminator_hash_input] seeds in spl-token-group-interface 0.6.0 and
// spl-token-metadata-interface 0.7.0 — the byte literals below are
// hand-computed from those seeds, independent of the implementation.
//
// Quirks pinned here, all verified against the Rust source:
//   - The TLV chain is REACHABLE only through TokenInstruction::unpack
//     failure: every vector's first byte is an unknown token discriminant
//     (all nine discriminators start above 44), which is exactly Agave's
//     fall-through. An account-count failure on a VALID token instruction
//     is final and never reaches the TLV parsers.
//   - Both interfaces REJECT trailing bytes (bytemuck try_from_bytes and
//     borsh try_from_slice demand exact consumption) — the opposite of the
//     bincode programs.
//   - InitializeMember's payload is a zero-sized Pod: exactly the 8
//     discriminator bytes.
//   - OptionalNonZeroPubkey renders present-and-null when all-zero, while
//     Emit's start/end are OMITTED when None.
//   - UpdateField's custom key passes through raw as the "field" value.
//   - RemoveKey's wire order is idempotent-then-key; borsh bools are
//     strictly 0/1.

import (
	"errors"
	"fmt"
	"testing"

	"github.com/blockchain-data-standards/manifesto/svm/parse"
)

// TLV discriminators: sha256(seed)[..8], hand-computed.
var (
	tlvInitGroup    = []byte{0x79, 0x71, 0x6C, 0x27, 0x36, 0x33, 0x00, 0x04} // spl_token_group_interface:initialize_token_group
	tlvUpdateMax    = []byte{0x6C, 0x25, 0xAB, 0x8F, 0xF8, 0x1E, 0x12, 0x6E} // spl_token_group_interface:update_group_max_size
	tlvUpdateGrAuth = []byte{0xA1, 0x69, 0x58, 0x01, 0xED, 0xDD, 0xD8, 0xCB} // spl_token_group_interface:update_authority
	tlvInitMember   = []byte{0x98, 0x20, 0xDE, 0xB0, 0xDF, 0xED, 0x74, 0x86} // spl_token_group_interface:initialize_member
	tlvInitMeta     = []byte{0xD2, 0xE1, 0x1E, 0xA2, 0x58, 0xB8, 0x4D, 0x8D} // spl_token_metadata_interface:initialize_account
	tlvUpdateField  = []byte{0xDD, 0xE9, 0x31, 0x2D, 0xB5, 0xCA, 0xDC, 0xC8} // spl_token_metadata_interface:updating_field
	tlvRemoveKey    = []byte{0xEA, 0x12, 0x20, 0x38, 0x59, 0x8D, 0x25, 0xB5} // spl_token_metadata_interface:remove_key_ix
	tlvUpdateMeAuth = []byte{0xD7, 0xE4, 0xA6, 0xE4, 0x54, 0x64, 0x56, 0x7B} // spl_token_metadata_interface:update_the_authority
	tlvEmit         = []byte{0xFA, 0xA6, 0xB4, 0xFA, 0x0D, 0x0C, 0xB8, 0x46} // spl_token_metadata_interface:emitter
)

// borshStr is borsh's string layout: u32 LE byte length + utf8 bytes.
func borshStr(s string) []byte { return cat(le32(uint32(len(s))), []byte(s)) }

func TestTokenInterfaceVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			name:      "initializeTokenGroup authority set",
			programID: parse.Token2022ID,
			data:      cat(tlvInitGroup, kb(0x95), le64(500)),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeTokenGroup","info":{
				"group":%q,"maxSize":500,"mint":%q,"mintAuthority":%q,
				"updateAuthority":%q}}`, k(1), k(2), k(3), k(0x95)),
		},
		{
			name:      "initializeTokenGroup zero authority is present-null",
			programID: parse.Token2022ID,
			data:      cat(tlvInitGroup, rep(32, 0), le64(0)),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeTokenGroup","info":{
				"group":%q,"maxSize":0,"mint":%q,"mintAuthority":%q,
				"updateAuthority":null}}`, k(1), k(2), k(3)),
		},
		{
			// The TLV chain also runs under the CLASSIC token id: both ids
			// share Agave's parser.
			name:      "updateTokenGroupMaxSize under classic token id",
			programID: parse.TokenProgramID,
			data:      cat(tlvUpdateMax, le64(123)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateTokenGroupMaxSize","info":{
				"group":%q,"maxSize":123,"updateAuthority":%q}}`, k(1), k(2)),
		},
		{
			name:      "updateTokenGroupAuthority set",
			programID: parse.Token2022ID,
			data:      cat(tlvUpdateGrAuth, kb(0x96)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateTokenGroupAuthority","info":{
				"group":%q,"updateAuthority":%q,"newAuthority":%q}}`, k(1), k(2), k(0x96)),
		},
		{
			name:      "updateTokenGroupAuthority zero is present-null",
			programID: parse.Token2022ID,
			data:      cat(tlvUpdateGrAuth, rep(32, 0)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateTokenGroupAuthority","info":{
				"group":%q,"updateAuthority":%q,"newAuthority":null}}`, k(1), k(2)),
		},
		{
			// Zero-sized Pod: the instruction is EXACTLY the 8 disc bytes.
			name:      "initializeTokenGroupMember",
			programID: parse.Token2022ID,
			data:      tlvInitMember,
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeTokenGroupMember","info":{
				"member":%q,"memberMint":%q,"memberMintAuthority":%q,
				"group":%q,"groupUpdateAuthority":%q}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
		{
			// Multibyte utf8 in a borsh string: length is BYTES.
			name:      "initializeTokenMetadata",
			programID: parse.Token2022ID,
			data: cat(tlvInitMeta, borshStr("Tokén"), borshStr("MTK"),
				borshStr("https://example.com/t.json")),
			accounts: accs(1, 2, 3, 4),
			program:  "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeTokenMetadata","info":{
				"metadata":%q,"updateAuthority":%q,"mint":%q,"mintAuthority":%q,
				"name":"Tokén","symbol":"MTK","uri":"https://example.com/t.json"}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "updateTokenMetadataField well-known name",
			programID: parse.Token2022ID,
			data:      cat(tlvUpdateField, []byte{0}, borshStr("New Name")),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateTokenMetadataField","info":{
				"metadata":%q,"updateAuthority":%q,"field":"name","value":"New Name"}}`,
				k(1), k(2)),
		},
		{
			// Field::Key(custom): the raw key string IS the field value.
			name:      "updateTokenMetadataField custom key passthrough",
			programID: parse.Token2022ID,
			data:      cat(tlvUpdateField, []byte{3}, borshStr("twitter"), borshStr("@example")),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateTokenMetadataField","info":{
				"metadata":%q,"updateAuthority":%q,"field":"twitter","value":"@example"}}`,
				k(1), k(2)),
		},
		{
			// Wire order is idempotent THEN key; JSON renders key first.
			name:      "removeTokenMetadataKey idempotent true",
			programID: parse.Token2022ID,
			data:      cat(tlvRemoveKey, []byte{1}, borshStr("twitter")),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"removeTokenMetadataKey","info":{
				"metadata":%q,"updateAuthority":%q,"key":"twitter","idempotent":true}}`,
				k(1), k(2)),
		},
		{
			name:      "removeTokenMetadataKey idempotent false",
			programID: parse.Token2022ID,
			data:      cat(tlvRemoveKey, []byte{0}, borshStr("k")),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"removeTokenMetadataKey","info":{
				"metadata":%q,"updateAuthority":%q,"key":"k","idempotent":false}}`,
				k(1), k(2)),
		},
		{
			name:      "updateTokenMetadataAuthority set",
			programID: parse.Token2022ID,
			data:      cat(tlvUpdateMeAuth, kb(0x97)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateTokenMetadataAuthority","info":{
				"metadata":%q,"updateAuthority":%q,"newAuthority":%q}}`, k(1), k(2), k(0x97)),
		},
		{
			name:      "updateTokenMetadataAuthority zero is present-null",
			programID: parse.Token2022ID,
			data:      cat(tlvUpdateMeAuth, rep(32, 0)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateTokenMetadataAuthority","info":{
				"metadata":%q,"updateAuthority":%q,"newAuthority":null}}`, k(1), k(2)),
		},
		{
			// Unlike the null-rendering authority fields above, Emit OMITS
			// its None keys entirely.
			name:      "emitTokenMetadata both None omit keys",
			programID: parse.Token2022ID,
			data:      cat(tlvEmit, []byte{0}, []byte{0}),
			accounts:  accs(1),
			program:   "spl-token",
			parsed:    fmt.Sprintf(`{"type":"emitTokenMetadata","info":{"metadata":%q}}`, k(1)),
		},
		{
			name:      "emitTokenMetadata start only",
			programID: parse.Token2022ID,
			data:      cat(tlvEmit, []byte{1}, le64(5), []byte{0}),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"emitTokenMetadata","info":{
				"metadata":%q,"start":5}}`, k(1)),
		},
		{
			name:      "emitTokenMetadata both set",
			programID: parse.Token2022ID,
			data:      cat(tlvEmit, []byte{1}, le64(5), []byte{1}, le64(9)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"emitTokenMetadata","info":{
				"metadata":%q,"start":5,"end":9}}`, k(1)),
		},
	})
}

// TestTokenInterfaceNotParsable: borsh/bytemuck exact-consumption refusals —
// one trailing-byte case per TLV instruction — plus borsh strictness (bool
// byte 2, Option tag 2, invalid utf8) and the fall-through boundary.
func TestTokenInterfaceNotParsable(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		accounts []string
	}{
		{"initializeTokenGroup trailing byte", cat(tlvInitGroup, rep(32, 1), le64(1), []byte{9}), accs(1, 2, 3)},
		{"updateTokenGroupMaxSize trailing byte", cat(tlvUpdateMax, le64(1), []byte{9}), accs(1, 2)},
		{"updateTokenGroupAuthority trailing byte", cat(tlvUpdateGrAuth, rep(32, 1), []byte{9}), accs(1, 2)},
		{"initializeTokenGroupMember trailing byte", cat(tlvInitMember, []byte{9}), accs(1, 2, 3, 4, 5)},
		{"initializeTokenMetadata trailing byte",
			cat(tlvInitMeta, borshStr("a"), borshStr("b"), borshStr("c"), []byte{9}), accs(1, 2, 3, 4)},
		{"updateTokenMetadataField trailing byte",
			cat(tlvUpdateField, []byte{0}, borshStr("v"), []byte{9}), accs(1, 2)},
		{"removeTokenMetadataKey trailing byte",
			cat(tlvRemoveKey, []byte{1}, borshStr("k"), []byte{9}), accs(1, 2)},
		{"updateTokenMetadataAuthority trailing byte",
			cat(tlvUpdateMeAuth, rep(32, 1), []byte{9}), accs(1, 2)},
		{"emitTokenMetadata trailing byte",
			cat(tlvEmit, []byte{0}, []byte{0}, []byte{9}), accs(1)},

		// Borsh strictness.
		{"removeTokenMetadataKey bool byte 2", cat(tlvRemoveKey, []byte{2}, borshStr("k")), accs(1, 2)},
		{"emitTokenMetadata Option tag 2", cat(tlvEmit, []byte{2}, []byte{0}), accs(1)},
		{"emitTokenMetadata truncated Some", cat(tlvEmit, []byte{1}, le64(1)[:7]), accs(1)},
		{"initializeTokenMetadata invalid utf8 name",
			cat(tlvInitMeta, le32(2), []byte{0xFF, 0xFE}, borshStr("b"), borshStr("c")), accs(1, 2, 3, 4)},
		{"updateTokenMetadataField unknown field tag",
			cat(tlvUpdateField, []byte{4}, borshStr("v")), accs(1, 2)},
		{"initializeTokenMetadata string length exceeds data",
			cat(tlvInitMeta, le32(200), []byte("ab")), accs(1, 2, 3, 4)},

		// Truncated payloads fall through both interfaces.
		{"initializeTokenGroup short payload", cat(tlvInitGroup, rep(39, 1)), accs(1, 2, 3)},
		{"seven byte discriminator prefix", tlvInitGroup[:7], accs(1)},

		// Once unpack succeeds, an account-count error is FINAL.
		{"initializeTokenGroupMember unpack ok account count final", tlvInitMember, accs(1, 2, 3, 4)},
		{"initializeTokenGroup 2 accounts", cat(tlvInitGroup, rep(32, 1), le64(1)), accs(1, 2)},

		// An account-count failure on a VALID TokenInstruction never falls
		// through to TLV (Agave only tries the interfaces when unpack itself
		// failed) — no nine-byte TLV disc starts with a valid discriminant,
		// so a success here could only come from an (incorrect) fall-through.
		{"valid token transfer short accounts stays final", cat([]byte{3}, le64(1)), accs(1, 2)},
		{"valid reallocate short accounts stays final", cat([]byte{29}, le16(1)), accs(1, 2, 3)},
	}
	for _, tc := range cases {
		env, err := parse.Parse(parse.Token2022ID, tc.data, tc.accounts, nil)
		if err == nil {
			t.Errorf("%s: Parse() succeeded, want ErrNotParsable; parsed: %s", tc.name, env.Parsed)
			continue
		}
		if !errors.Is(err, parse.ErrNotParsable) {
			t.Errorf("%s: error %v does not wrap ErrNotParsable", tc.name, err)
		}
	}
}

// TestTokenInterfaceFallbackReachability pins the unpack boundary from the
// positive side: a payload whose FIRST byte is an unknown TokenInstruction
// discriminant but whose 8-byte prefix matches a TLV discriminator must
// parse as the interface instruction — proof the fall-through chain runs.
func TestTokenInterfaceFallbackReachability(t *testing.T) {
	data := cat(tlvUpdateMax, le64(7))
	if data[0] <= 44 {
		t.Fatalf("test premise broken: disc byte %d is a valid TokenInstruction", data[0])
	}
	env, err := parse.Parse(parse.Token2022ID, data, accs(1, 2), nil)
	if err != nil {
		t.Fatalf("TLV fallback did not run: %v", err)
	}
	want := decodeJSON(t, []byte(fmt.Sprintf(
		`{"type":"updateTokenGroupMaxSize","info":{"group":%q,"maxSize":7,"updateAuthority":%q}}`,
		k(1), k(2))))
	if path, diff, ok := firstDiff("", want, decodeJSON(t, env.Parsed)); !ok {
		t.Errorf("fallback envelope mismatch at %q: %s\ngot: %s", path, diff, env.Parsed)
	}
}
