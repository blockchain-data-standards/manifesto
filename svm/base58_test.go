package svm

import (
	"bytes"
	"testing"
)

func TestBase58EncodeKnownVectors(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want string
	}{
		{"empty", []byte{}, ""},
		{"single zero byte becomes 1", []byte{0}, "1"},
		{"leading zeros are preserved as ones", []byte{0, 0, 0}, "111"},
		{"ascii", []byte("hello world"), "StV1DL6CwTryKyV"},
		{"leading zero then last alphabet char", []byte{0x00, 0x39}, "1z"},
		{"leading zero then multi-digit", []byte{0x00, 0x3c}, "123"},
		{"single byte", []byte{57}, "z"},
		// The all-zero 32-byte pubkey is the Solana system program id; it must
		// render as 32 ones, not the empty string. This is the case a naive
		// bignum implementation gets wrong.
		{"system program id", make([]byte, 32), "11111111111111111111111111111111"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Base58Encode(tc.in); got != tc.want {
				t.Fatalf("Base58Encode(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBase58RoundTrip(t *testing.T) {
	cases := [][]byte{
		{},
		{0},
		{0, 0, 1, 2, 3},
		[]byte("hello world"),
		make([]byte, 32),
		bytes.Repeat([]byte{0xff}, 32),
		bytes.Repeat([]byte{0xab}, 64),
	}

	for _, in := range cases {
		enc := Base58Encode(in)
		out, err := Base58Decode(enc)
		if err != nil {
			t.Fatalf("Base58Decode(%q) errored: %v", enc, err)
		}
		if !bytes.Equal(in, out) {
			t.Fatalf("round trip mismatch for %v: got %v via %q", in, out, enc)
		}
	}
}

func TestBase58DecodeRejectsInvalidCharacters(t *testing.T) {
	// '0', 'O', 'I' and 'l' are excluded from the Bitcoin alphabet precisely
	// because they are visually ambiguous; decoding must reject rather than
	// silently skip them.
	for _, in := range []string{"0", "O", "I", "l", "abc!def"} {
		if _, err := Base58Decode(in); err == nil {
			t.Fatalf("Base58Decode(%q) succeeded, want error", in)
		}
	}
}
