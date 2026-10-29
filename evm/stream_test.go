package evm

import (
	"bytes"
	"encoding/hex"
	"testing"

	"google.golang.org/protobuf/proto"
)

// streamBlocksResponsePR26 is a header-only StreamBlocksResponse marshalled
// (deterministic) by the manifesto#26 code, where header was a plain field 1.
const streamBlocksResponsePR26 = "0a7e08c0de810a1080e9f6b806188087a70e20cec2f1053220aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa3a20bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb6214cccccccccccccccccccccccccccccccccccccccc7800e2010a30783362396163613030"

// A #26-era client and the oneof StreamBlocksResponse share one wire format
// for header messages, in both directions.
func TestStreamBlocksResponseHeaderWireCompatible(t *testing.T) {
	old, err := hex.DecodeString(streamBlocksResponsePR26)
	if err != nil {
		t.Fatal(err)
	}

	var got StreamBlocksResponse
	if err := proto.Unmarshal(old, &got); err != nil {
		t.Fatalf("unmarshal #26 bytes: %v", err)
	}
	h := got.GetHeader()
	if h == nil {
		t.Fatalf("header not set: %v", &got)
	}
	if h.Number != 21_000_000 || h.Timestamp != 1_730_000_000 || h.GasUsed != 12_345_678 ||
		!bytes.Equal(h.Hash, bytes.Repeat([]byte{0xaa}, 32)) ||
		!bytes.Equal(h.ParentHash, bytes.Repeat([]byte{0xbb}, 32)) ||
		h.GetBaseFeePerGas() != "0x3b9aca00" || h.Nonce == nil {
		t.Fatalf("header fields: %v", h)
	}
	if len(got.ProtoReflect().GetUnknown()) != 0 {
		t.Fatalf("unknown fields left: %x", got.ProtoReflect().GetUnknown())
	}

	again, err := proto.MarshalOptions{Deterministic: true}.Marshal(&StreamBlocksResponse{
		Message: &StreamBlocksResponse_Header{Header: h},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, old) {
		t.Fatalf("re-encoded bytes differ from #26:\n got %x\nwant %x", again, old)
	}
}
