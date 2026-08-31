package evm

import (
	"strings"
	"testing"
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
func TestAuthorizationChainIdBeyondUint64FallsBackToZero(t *testing.T) {
	auth, err := parseAuth(t, wideChainId)
	if err != nil {
		t.Fatalf("the block must still parse: %v", err)
	}
	if auth.ChainId != 0 {
		t.Fatalf("chainId = %d, want the documented 0 fallback", auth.ChainId)
	}
	tx := &Transaction{AuthorizationList: []*AuthorizationListItem{auth}}
	out := TransactionToJsonRpc(tx)["authorizationList"].([]interface{})[0].(map[string]interface{})
	if got := out["chainId"]; got != "0x0" {
		t.Fatalf("chainId rendered %v, want 0x0", got)
	}
}

// The representable path is untouched: ordinary chain ids still render as
// minimal hex quantities.
func TestAuthorizationChainIdRendersAsNumberNotBytes(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"0x1", "0x1"},
		{"11155111", "0xaa36a7"},
		{"0x0", "0x0"},
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

// Negative quantities are malformed everywhere, not just here: the shared
// helper used to normalize "-1" into the invalid QUANTITY "0x-1".
func TestDecimalStringToHexRejectsNegatives(t *testing.T) {
	for _, in := range []string{"-1", "0x-1", "-0x1"} {
		if got, err := DecimalStringToHex(in); err == nil {
			t.Fatalf("DecimalStringToHex(%q) = %q, want error", in, got)
		}
	}
}
