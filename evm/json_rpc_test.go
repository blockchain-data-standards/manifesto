package evm

import (
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
