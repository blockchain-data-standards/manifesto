package evm

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestReceiptL1BlobBaseFeeConversion(t *testing.T) {
	// Test JsonRpcReceipt to Proto conversion with l1BlobBaseFee
	jsonReceipt := &JsonRpcReceipt{
		BlockHash:         "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		BlockNumber:       "0x1",
		TransactionHash:   "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		TransactionIndex:  "0x0",
		From:              "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb7",
		To:                "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb8",
		GasUsed:           "0x5208",
		CumulativeGasUsed: "0x5208",
		EffectiveGasPrice: "0x3b9aca00",
		LogsBloom:         "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		Status:            "0x1",
		Type:              "0x2",
		L1BlobBaseFee:     "0x1234",
		L1BaseFeeScalar:   "0x8dd",
		Logs:              []*JsonRpcLog{},
	}

	// Convert to proto
	protoReceipt, err := jsonReceipt.ToProto()
	if err != nil {
		t.Fatalf("Failed to convert JsonRpcReceipt to proto: %v", err)
	}

	// Check that l1BlobBaseFee was properly converted
	if protoReceipt.L1BlobBaseFee == nil {
		t.Error("Expected L1BlobBaseFee to be set in proto receipt")
	} else if *protoReceipt.L1BlobBaseFee != "0x1234" {
		t.Errorf("Expected L1BlobBaseFee to be '0x1234', got '%s'", *protoReceipt.L1BlobBaseFee)
	}

	// Test Proto to JsonRpc conversion
	jsonRpcMap := ReceiptToJsonRpc(protoReceipt)

	// Check that l1BlobBaseFee is properly converted back to hex
	if l1BlobBaseFee, ok := jsonRpcMap["l1BlobBaseFee"]; !ok {
		t.Error("Expected l1BlobBaseFee in JSON-RPC output")
	} else if l1BlobBaseFee != "0x1234" {
		t.Errorf("Expected l1BlobBaseFee to be '0x1234', got '%v'", l1BlobBaseFee)
	}
}

// Test that l1BlobBaseFee is omitted when not present
func TestReceiptL1BlobBaseFeeOmitted(t *testing.T) {
	jsonReceipt := &JsonRpcReceipt{
		BlockHash:         "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		BlockNumber:       "0x1",
		TransactionHash:   "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		TransactionIndex:  "0x0",
		From:              "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb7",
		To:                "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb8",
		GasUsed:           "0x5208",
		CumulativeGasUsed: "0x5208",
		EffectiveGasPrice: "0x3b9aca00",
		LogsBloom:         "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		Status:            "0x1",
		Type:              "0x0",
		Logs:              []*JsonRpcLog{},
		// Note: L1BlobBaseFee is not set
	}

	// Convert to proto
	protoReceipt, err := jsonReceipt.ToProto()
	if err != nil {
		t.Fatalf("Failed to convert JsonRpcReceipt to proto: %v", err)
	}

	// Check that l1BlobBaseFee is nil when not provided
	if protoReceipt.L1BlobBaseFee != nil {
		t.Errorf("Expected L1BlobBaseFee to be nil, got '%s'", *protoReceipt.L1BlobBaseFee)
	}

	// Test Proto to JsonRpc conversion
	jsonRpcMap := ReceiptToJsonRpc(protoReceipt)

	// Check that l1BlobBaseFee is not in the output when nil
	if _, ok := jsonRpcMap["l1BlobBaseFee"]; ok {
		t.Error("Expected l1BlobBaseFee to be omitted from JSON-RPC output when nil")
	}
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSignatureRAndSRenderAsQuantities(t *testing.T) {
	addr := "c20699185c15d0a2fd65779bb5d69f5b0b113c00"
	fullR := "0a" + strings.Repeat("1b", 31)
	for _, tc := range []struct {
		name         string
		r, s         string
		wantR, wantS string
	}{
		{"minimal bytes", "01", addr, "0x1", "0x" + addr},
		{"left-padded to 32 bytes", strings.Repeat("00", 31) + "01", strings.Repeat("00", 12) + addr, "0x1", "0x" + addr},
		{"full width with leading zero nibble", fullR, fullR, "0x" + fullR[1:], "0x" + fullR[1:]},
		{"zero", strings.Repeat("00", 32), "00", "0x0", "0x0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, s := mustHex(t, tc.r), mustHex(t, tc.s)
			tx := &Transaction{
				R: r,
				S: s,
				AuthorizationList: []*AuthorizationListItem{{
					ChainId: 1,
					Address: mustHex(t, addr),
					R:       r,
					S:       s,
				}},
			}
			out := TransactionToJsonRpc(tx)
			auth := out["authorizationList"].([]interface{})[0].(map[string]interface{})
			for _, c := range []struct {
				field string
				got   interface{}
				want  string
			}{
				{"r", out["r"], tc.wantR},
				{"s", out["s"], tc.wantS},
				{"authorizationList[0].r", auth["r"], tc.wantR},
				{"authorizationList[0].s", auth["s"], tc.wantS},
			} {
				if c.got != c.want {
					t.Errorf("%s = %v, want %s", c.field, c.got, c.want)
				}
			}
		})
	}
}
