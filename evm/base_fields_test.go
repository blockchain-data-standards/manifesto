package evm

import (
	"testing"
)

func TestBaseChainFields(t *testing.T) {
	// Test Transaction with Base-specific fields
	t.Run("Transaction with isSystemTx and depositReceiptVersion", func(t *testing.T) {
		txMap := map[string]interface{}{
			"hash":                  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"nonce":                 "0x1",
			"from":                  "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb7",
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

		// Test conversion back to JSON-RPC
		jsonRpc := TransactionToJsonRpc(tx)
		if isSystemTx, ok := jsonRpc["isSystemTx"].(bool); !ok || !isSystemTx {
			t.Error("Expected isSystemTx in JSON-RPC output to be true")
		}
		if depositReceiptVersion, ok := jsonRpc["depositReceiptVersion"]; !ok || depositReceiptVersion != "0x1" {
			t.Errorf("Expected depositReceiptVersion in JSON-RPC output to be '0x1', got '%v'", depositReceiptVersion)
		}
	})

	// Test Receipt with l1BlobBaseFeeScalar
	t.Run("Receipt with l1BlobBaseFeeScalar", func(t *testing.T) {
		jsonReceipt := &JsonRpcReceipt{
			BlockHash:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			BlockNumber:         "0x1",
			TransactionHash:     "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			TransactionIndex:    "0x0",
			From:                "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb7",
			To:                  "0x4200000000000000000000000000000000000015",
			GasUsed:             "0x5208",
			CumulativeGasUsed:   "0x5208",
			EffectiveGasPrice:   "0x0",
			LogsBloom:           "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			Status:              "0x1",
			Type:                "0x7e",
			L1BlobBaseFee:       "0x1",
			L1BlobBaseFeeScalar: "0x101c12",
			L1BaseFeeScalar:     "0x8dd",
			Logs:                []*JsonRpcLog{},
		}

		protoReceipt, err := jsonReceipt.ToProto()
		if err != nil {
			t.Fatalf("Failed to convert receipt to proto: %v", err)
		}

		if protoReceipt.L1BlobBaseFeeScalar == nil {
			t.Error("Expected L1BlobBaseFeeScalar to be set")
		} else if *protoReceipt.L1BlobBaseFeeScalar != 0x101c12 {
			t.Errorf("Expected L1BlobBaseFeeScalar to be 0x101c12, got 0x%x", *protoReceipt.L1BlobBaseFeeScalar)
		}

		// Test conversion back to JSON-RPC
		jsonRpcMap := ReceiptToJsonRpc(protoReceipt)
		if l1BlobBaseFeeScalar, ok := jsonRpcMap["l1BlobBaseFeeScalar"]; !ok || l1BlobBaseFeeScalar != "0x101c12" {
			t.Errorf("Expected l1BlobBaseFeeScalar in JSON-RPC output to be '0x101c12', got '%v'", l1BlobBaseFeeScalar)
		}
	})

	// Test BlockHeader with requestsHash
	t.Run("BlockHeader with requestsHash", func(t *testing.T) {
		block := &JsonRpcBlock{
			Number:           "0x1",
			Hash:             "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ParentHash:       "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Timestamp:        "0x1",
			GasLimit:         "0x1000",
			GasUsed:          "0x100",
			Size:             "0x100",
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
			t.Fatalf("Failed to convert block to proto: %v", err)
		}

		if protoBlock.Header.RequestsHash == nil {
			t.Error("Expected RequestsHash to be set")
		} else {
			expectedHash := "0x7685abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456"
			actualHash := BytesToHex(protoBlock.Header.RequestsHash)
			if actualHash != expectedHash {
				t.Errorf("Expected RequestsHash to be '%s', got '%s'", expectedHash, actualHash)
			}
		}

		// Test conversion back to JSON-RPC
		jsonRpcMap := BlockToJsonRpc(protoBlock.Header, nil, nil, nil)
		if requestsHash, ok := jsonRpcMap["requestsHash"]; !ok || requestsHash != "0x7685abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456" {
			t.Errorf("Expected requestsHash in JSON-RPC output, got '%v'", requestsHash)
		}
	})
}
