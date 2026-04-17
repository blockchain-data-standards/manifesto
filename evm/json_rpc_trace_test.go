package evm

import "testing"

func TestTraceFromParityAndNativeTransfers(t *testing.T) {
	blockTs := uint64(42)
	trace, err := TraceFromParity(map[string]interface{}{
		"type": "call",
		"action": map[string]interface{}{
			"from":     "0x0000000000000000000000000000000000000001",
			"to":       "0x0000000000000000000000000000000000000002",
			"value":    "0x10",
			"gas":      "0x5208",
			"input":    "0x12345678",
			"callType": "call",
		},
		"result": map[string]interface{}{
			"gasUsed": "0x5208",
			"output":  "0x",
		},
		"subtraces":         float64(0),
		"traceAddress":      []interface{}{},
		"transactionHash":   "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"transactionIndex":  float64(1),
	}, 100, MustHexToBytes("0xbb"), &blockTs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trace.TraceType != TraceType_TRACE_CALL {
		t.Fatalf("unexpected trace type: %v", trace.TraceType)
	}

	transfers := NativeTransfersFromTraces([]*Trace{trace})
	if len(transfers) != 1 {
		t.Fatalf("unexpected transfer count: %d", len(transfers))
	}
	if transfers[0].Value != "0x10" {
		t.Fatalf("unexpected transfer value: %s", transfers[0].Value)
	}
}

func TestTraceFromGethDebugFlattensNestedCalls(t *testing.T) {
	traces, err := TraceFromGethDebug(map[string]interface{}{
		"type":  "CALL",
		"from":  "0x0000000000000000000000000000000000000001",
		"to":    "0x0000000000000000000000000000000000000002",
		"value": "0x0",
		"gas":   "0x1",
		"gasUsed": "0x1",
		"calls": []interface{}{
			map[string]interface{}{
				"type":  "STATICCALL",
				"from":  "0x0000000000000000000000000000000000000002",
				"to":    "0x0000000000000000000000000000000000000003",
				"value": "0x0",
				"gas":   "0x1",
				"gasUsed": "0x1",
			},
		},
	}, 1, MustHexToBytes("0xaa"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(traces) != 2 {
		t.Fatalf("unexpected trace count: %d", len(traces))
	}
	if len(traces[1].TraceAddress) != 1 || traces[1].TraceAddress[0] != 0 {
		t.Fatalf("unexpected child trace address: %+v", traces[1].TraceAddress)
	}
}
