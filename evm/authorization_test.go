package evm

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

// The Sepolia value this widening exists for: block 8542703, type-0x4
// transaction at index 271 — 20 bytes in what used to be a 64-bit field.
const wideChainId = "0xf6a0be9433ee09f5ba0d5784b102833333333333"

func authTxJson(chainId interface{}) map[string]interface{} {
	auth := map[string]interface{}{
		"address": "0x1111111111111111111111111111111111111111",
		"nonce":   "0x1",
		"yParity": "0x1",
		"r":       "0x9913bfae03d1a071190d025f6af8ffe4597d77668a89170a8a4a866c92d19644",
		"s":       "0x40c47e2a25e4a25c725172969f680d3eca1f36aa21bccdbc6e9237315e07e0c7",
	}
	if chainId != nil {
		auth["chainId"] = chainId
	}
	return map[string]interface{}{
		"hash":              "0x" + strings.Repeat("11", 32),
		"nonce":             "0x1",
		"gas":               "0x5208",
		"from":              "0x2222222222222222222222222222222222222222",
		"value":             "0x0",
		"type":              "0x4",
		"authorizationList": []interface{}{auth},
	}
}

func parseAuth(t *testing.T, chainId interface{}) (*AuthorizationListItem, error) {
	t.Helper()
	tx, err := ParseJsonRpcTransaction(authTxJson(chainId), nil)
	if err != nil {
		return nil, err
	}
	if len(tx.AuthorizationList) != 1 {
		t.Fatalf("expected 1 authorization, got %d", len(tx.AuthorizationList))
	}
	return tx.AuthorizationList[0], nil
}

// A chain id wider than 64 bits must survive verbatim: the block is canonical
// and immutable, so an indexer that narrows it cannot store the block at all.
func TestAuthorizationChainIdSurvivesBeyondUint64(t *testing.T) {
	auth, err := parseAuth(t, wideChainId)
	if err != nil {
		t.Fatalf("wide chainId must parse: %v", err)
	}
	if auth.ChainId != wideChainId {
		t.Fatalf("stored %q, want %q", auth.ChainId, wideChainId)
	}
	if auth.LegacyChainId != 0 {
		t.Fatalf("legacy mirror = %d, want 0 — the value does not fit 64 bits", auth.LegacyChainId)
	}
}

// Regression: ChainId is a string, so fmt.Sprintf("0x%x", …) would hex the
// string's bytes and render chain 1 as "0x31". go vet does not flag it.
func TestAuthorizationChainIdRendersAsNumberNotBytes(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"0x1", "0x1"},
		{"11155111", "0xaa36a7"},
		{"0x0", "0x0"},
		{wideChainId, wideChainId},
	} {
		auth, err := parseAuth(t, tc.in)
		if err != nil {
			t.Fatalf("chainId %q: %v", tc.in, err)
		}
		tx := &Transaction{AuthorizationList: []*AuthorizationListItem{auth}}
		out := TransactionToJsonRpc(tx)["authorizationList"].([]interface{})[0].(map[string]interface{})
		if got := out["chainId"]; got != tc.want {
			t.Fatalf("chainId %q rendered as %v, want %v", tc.in, got, tc.want)
		}
	}
}

// Absent, null, or non-string chainId is malformed, not zero: 0 means "valid
// on any chain" in EIP-7702. The Rust binding errors here too.
func TestAuthorizationChainIdIsRequired(t *testing.T) {
	for _, in := range []interface{}{nil, float64(1), "", true} {
		if _, err := parseAuth(t, in); err == nil {
			t.Fatalf("chainId %#v was accepted, want error", in)
		}
	}
}

// EIP-7702 asserts auth.chain_id < 2**256; past that the enclosing transaction
// is structurally invalid and can never appear in a canonical block.
func TestAuthorizationChainIdIsBoundedByUint256(t *testing.T) {
	maxUint256 := "0x" + strings.Repeat("f", 64)
	for _, tc := range []struct {
		in string
		ok bool
	}{
		{maxUint256, true},
		{"0x1" + strings.Repeat("0", 64), false},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639936", false},
		{"-1", false},
		{"0x-1", false},
		{"not-a-number", false},
	} {
		_, err := parseAuth(t, tc.in)
		if (err == nil) != tc.ok {
			t.Fatalf("chainId %q: err=%v, want ok=%v", tc.in, err, tc.ok)
		}
	}
}

// A pre-widening payload carries only the deprecated 64-bit field; reading it
// must yield that chain id rather than silently reading back as "any chain".
func TestAuthorizationChainIdFallsBackToLegacyField(t *testing.T) {
	auth := &AuthorizationListItem{LegacyChainId: 11155111}
	if got := AuthorizationChainId(auth); got != "11155111" {
		t.Fatalf("legacy fallback = %q, want \"11155111\"", got)
	}
	tx := &Transaction{AuthorizationList: []*AuthorizationListItem{auth}}
	out := TransactionToJsonRpc(tx)["authorizationList"].([]interface{})[0].(map[string]interface{})
	if got := out["chainId"]; got != "0xaa36a7" {
		t.Fatalf("legacy payload rendered %v, want 0xaa36a7", got)
	}
}

// Dual-write: an old reader keeps seeing a chain id for every value that fits.
func TestAuthorizationChainIdMirrorsIntoLegacyField(t *testing.T) {
	auth, err := parseAuth(t, "11155111")
	if err != nil {
		t.Fatal(err)
	}
	if auth.LegacyChainId != 11155111 {
		t.Fatalf("legacy mirror = %d, want 11155111", auth.LegacyChainId)
	}
}

// Negative quantities are malformed everywhere, not just here: the shared
// helper used to normalize "-1" into the invalid QUANTITY "0x-1".
func TestDecimalStringToHexRejectsNegatives(t *testing.T) {
	for _, in := range []string{"-1", "0x-1", "-0x1"} {
		if got, err := DecimalStringToHex(in); err == nil {
			t.Fatalf("DecimalStringToHex(%q) = %q, want error", in, got)
		}
	}
}

// Wire migration, old -> new: a payload written before the widening carries
// only field 1. The new reader must recover that chain id; reading it as 0
// would silently turn a chain-scoped authorization into an any-chain one.
func TestAuthorizationWireCompatOldPayloadKeepsChainId(t *testing.T) {
	old := protowire.AppendVarint(protowire.AppendTag(nil, 1, protowire.VarintType), 11155111)
	old = protowire.AppendBytes(protowire.AppendTag(old, 2, protowire.BytesType), []byte{0xaa, 0xbb})

	var item AuthorizationListItem
	if err := proto.Unmarshal(old, &item); err != nil {
		t.Fatalf("old payload must decode: %v", err)
	}
	if got := AuthorizationChainId(&item); got != "11155111" {
		t.Fatalf("chain id read back as %q, want \"11155111\"", got)
	}
}

// Wire migration, new -> old: a reader still on the pre-widening schema reads
// field 1, so the writer must keep populating it whenever the value fits.
func TestAuthorizationWireCompatNewPayloadCarriesLegacyField(t *testing.T) {
	auth, err := parseAuth(t, "11155111")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := proto.Marshal(auth)
	if err != nil {
		t.Fatal(err)
	}
	if got := varintField(t, encoded, 1); got != 11155111 {
		t.Fatalf("field 1 on the wire = %d, want 11155111", got)
	}
}

// varintField returns the varint value of the given field number, scanning the
// encoded message the way a reader on the old schema would.
func varintField(t *testing.T, buf []byte, want protowire.Number) uint64 {
	t.Helper()
	for len(buf) > 0 {
		num, typ, n := protowire.ConsumeTag(buf)
		if n < 0 {
			t.Fatalf("malformed tag: %v", protowire.ParseError(n))
		}
		buf = buf[n:]
		if num == want && typ == protowire.VarintType {
			v, n := protowire.ConsumeVarint(buf)
			if n < 0 {
				t.Fatalf("malformed varint: %v", protowire.ParseError(n))
			}
			return v
		}
		n = protowire.ConsumeFieldValue(num, typ, buf)
		if n < 0 {
			t.Fatalf("malformed field: %v", protowire.ParseError(n))
		}
		buf = buf[n:]
	}
	t.Fatalf("field %d absent", want)
	return 0
}
