package evm

import (
	"testing"
)

func TestComprehensiveFieldCoverage(t *testing.T) {
	// Test Transaction with all fields including execution results
	t.Run("Transaction with execution result fields", func(t *testing.T) {
		txMap := map[string]interface{}{
			"hash":                  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"nonce":                 "0x1",
			"from":                  "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb7",
			"to":                    "0x4200000000000000000000000000000000000015",
			"value":                 "0x100",
			"input":                 "0xabcdef",
			"gas":                   "0x5208",
			"gasPrice":              "0x3b9aca00",
			"type":                  "0x2",
			"r":                     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"s":                     "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"v":                     "0x1b",
			"chainId":               "0x1",
			"blockNumber":           "0x1000",
			"blockHash":             "0xblockhash1234567890abcdef1234567890abcdef1234567890abcdef12345678",
			"transactionIndex":      "0x5",
			// Execution result fields
			"gasUsed":               "0x5000",
			"effectiveGasPrice":     "0x3b9aca00",
			"blobGasUsed":           "0x20000",
			"blobGasPrice":          "0x1000",
		}

		tx, err := ParseJsonRpcTransaction(txMap, nil)
		if err != nil {
			t.Fatalf("Failed to parse transaction: %v", err)
		}

		// Check execution result fields
		if tx.GasUsed == nil || *tx.GasUsed != 0x5000 {
			t.Error("Expected gasUsed to be 0x5000")
		}
		if tx.EffectiveGasPrice == nil || *tx.EffectiveGasPrice != "0x3b9aca00" {
			t.Error("Expected effectiveGasPrice to be '0x3b9aca00'")
		}
		if tx.BlobGasUsed == nil || *tx.BlobGasUsed != 0x20000 {
			t.Error("Expected blobGasUsed to be 0x20000")
		}
		if tx.BlobGasPrice == nil || *tx.BlobGasPrice != "0x1000" {
			t.Error("Expected blobGasPrice to be '0x1000'")
		}

		// Test conversion back to JSON-RPC
		jsonRpc := TransactionToJsonRpc(tx)
		
		// Verify execution result fields in output
		if gasUsed, ok := jsonRpc["gasUsed"]; !ok || gasUsed != "0x5000" {
			t.Errorf("Expected gasUsed in JSON-RPC output to be '0x5000', got '%v'", gasUsed)
		}
		if effectiveGasPrice, ok := jsonRpc["effectiveGasPrice"]; !ok || effectiveGasPrice != "0x3b9aca00" {
			t.Errorf("Expected effectiveGasPrice in JSON-RPC output to be '0x3b9aca00', got '%v'", effectiveGasPrice)
		}
		if blobGasUsed, ok := jsonRpc["blobGasUsed"]; !ok || blobGasUsed != "0x20000" {
			t.Errorf("Expected blobGasUsed in JSON-RPC output to be '0x20000', got '%v'", blobGasUsed)
		}
		if blobGasPrice, ok := jsonRpc["blobGasPrice"]; !ok || blobGasPrice != "0x1000" {
			t.Errorf("Expected blobGasPrice in JSON-RPC output to be '0x1000', got '%v'", blobGasPrice)
		}
		
		// Check blockTimestamp is output when present
		if tx.BlockTimestamp != nil {
			if blockTimestamp, ok := jsonRpc["blockTimestamp"]; !ok {
				t.Error("Expected blockTimestamp in JSON-RPC output when present")
			} else {
				t.Logf("blockTimestamp output: %v", blockTimestamp)
			}
		}
	})

	// Test Receipt with blockTimestamp
	t.Run("Receipt with blockTimestamp", func(t *testing.T) {
		jsonReceipt := &JsonRpcReceipt{
			BlockHash:         "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			BlockNumber:       "0x1000",
			BlockTimestamp:    "0x65000000",
			TransactionHash:   "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			TransactionIndex:  "0x5",
			From:              "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb7",
			To:                "0x4200000000000000000000000000000000000015",
			GasUsed:           "0x5000",
			CumulativeGasUsed: "0x10000",
			EffectiveGasPrice: "0x3b9aca00",
			LogsBloom:         "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			Status:            "0x1",
			Type:              "0x2",
			Logs:              []*JsonRpcLog{},
		}

		protoReceipt, err := jsonReceipt.ToProto()
		if err != nil {
			t.Fatalf("Failed to convert receipt to proto: %v", err)
		}

		// Check blockTimestamp was parsed
		if protoReceipt.BlockTimestamp == nil {
			t.Error("Expected BlockTimestamp to be set in proto receipt")
		} else if *protoReceipt.BlockTimestamp != 0x65000000 {
			t.Errorf("Expected BlockTimestamp to be 0x65000000, got 0x%x", *protoReceipt.BlockTimestamp)
		}

		// Test conversion back to JSON-RPC
		jsonRpcMap := ReceiptToJsonRpc(protoReceipt)
		
		// Check blockTimestamp is in output
		if blockTimestamp, ok := jsonRpcMap["blockTimestamp"]; !ok {
			t.Error("Expected blockTimestamp in JSON-RPC output")
		} else if blockTimestamp != "0x65000000" {
			t.Errorf("Expected blockTimestamp to be '0x65000000', got '%v'", blockTimestamp)
		}
	})

	// Test all Base chain specific fields together
	t.Run("All Base chain fields", func(t *testing.T) {
		// Block with requestsHash
		block := &JsonRpcBlock{
			Number:           "0x1000",
			Hash:             "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ParentHash:       "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Timestamp:        "0x65000000",
			GasLimit:         "0x1000000",
			GasUsed:          "0x100000",
			Size:             "0x1000",
			LogsBloom:        "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			TransactionsRoot: "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
			StateRoot:        "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
			ReceiptsRoot:     "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
			Sha3Uncles:       "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
			Miner:            "0x0000000000000000000000000000000000000000",
			ExtraData:        "0x",
			RequestsHash:     "0x7685abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456",
			Transactions:     []interface{}{},
		}

		protoBlock, err := block.ToProto()
		if err != nil {
			t.Fatalf("Failed to convert block: %v", err)
		}

		// Verify requestsHash
		if protoBlock.Header.RequestsHash == nil {
			t.Error("Expected RequestsHash to be set")
		}

		// Transaction with Base-specific fields
		txMap := map[string]interface{}{
			"hash":                  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"nonce":                 "0x1",
			"from":                  "0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001",
			"to":                    "0x4200000000000000000000000000000000000015",
			"value":                 "0x0",
			"input":                 "0x",
			"gas":                   "0x5208",
			"type":                  "0x7e",
			"isSystemTx":            true,
			"depositReceiptVersion": "0x1",
			"r":                     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"s":                     "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"v":                     "0x1b",
		}

		tx, err := ParseJsonRpcTransaction(txMap, nil)
		if err != nil {
			t.Fatalf("Failed to parse transaction: %v", err)
		}

		if tx.IsSystemTx == nil || !*tx.IsSystemTx {
			t.Error("Expected isSystemTx to be true")
		}
		if tx.DepositReceiptVersion == nil || *tx.DepositReceiptVersion != "0x1" {
			t.Error("Expected depositReceiptVersion to be '0x1'")
		}

		// Receipt with all L2 fee fields
		receipt := &JsonRpcReceipt{
			BlockHash:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			BlockNumber:         "0x1000",
			TransactionHash:     "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			TransactionIndex:    "0x0",
			From:                "0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001",
			To:                  "0x4200000000000000000000000000000000000015",
			GasUsed:             "0xb44c",
			CumulativeGasUsed:   "0xb44c",
			EffectiveGasPrice:   "0x0",
			LogsBloom:           "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			Status:              "0x1",
			Type:                "0x7e",
			L1GasPrice:          "0x3c2053c3",
			L1GasUsed:           "0x6da",
			L1Fee:               "0x0",
			L1BaseFeeScalar:     "0x8dd",
			L1BlobBaseFee:       "0x1",
			L1BlobBaseFeeScalar: "0x101c12",
			DepositNonce:        "0x211c31f",
			DepositReceiptVersion: "0x1",
			Logs:                []*JsonRpcLog{},
		}

		protoReceipt, err := receipt.ToProto()
		if err != nil {
			t.Fatalf("Failed to convert receipt: %v", err)
		}

		// Verify all L2 fee fields
		if protoReceipt.L1BlobBaseFee == nil || *protoReceipt.L1BlobBaseFee != "0x1" {
			t.Error("Expected L1BlobBaseFee to be '0x1'")
		}
		if protoReceipt.L1BlobBaseFeeScalar == nil || *protoReceipt.L1BlobBaseFeeScalar != 0x101c12 {
			t.Errorf("Expected L1BlobBaseFeeScalar to be 0x101c12")
		}
		if protoReceipt.DepositNonce == nil || *protoReceipt.DepositNonce != "0x211c31f" {
			t.Error("Expected DepositNonce to be '0x211c31f'")
		}
		if protoReceipt.DepositReceiptVersion == nil || *protoReceipt.DepositReceiptVersion != "0x1" {
			t.Error("Expected DepositReceiptVersion to be '0x1'")
		}
	})
}
