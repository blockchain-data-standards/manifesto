package svm

import "fmt"

// Base58 (Bitcoin alphabet) is Solana's canonical text encoding for pubkeys,
// signatures, blockhashes and — under `encoding: "json"` — instruction data.
//
// It is implemented here rather than pulled in as a dependency: the module
// otherwise depends only on grpc/protobuf, and the algorithm is small and
// fully specified. Keeping it in-tree also keeps the JSON-RPC mapping helpers
// dependency-free for consumers that vendor this package.
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var base58Decode = func() [256]int8 {
	var t [256]int8
	for i := range t {
		t[i] = -1
	}
	for i, c := range []byte(base58Alphabet) {
		t[c] = int8(i)
	}
	return t
}()

// Base58Encode renders bytes in base58. Each leading zero byte becomes a
// literal '1', which is what makes the 32-zero-byte system program id encode
// as 32 ones rather than the empty string.
func Base58Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	zeros := 0
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}

	// log(256)/log(58) ~= 1.365; 138/100 is the standard safe over-allocation.
	size := (len(input)-zeros)*138/100 + 1
	buf := make([]byte, size)

	high := size - 1
	for i := zeros; i < len(input); i++ {
		carry := int(input[i])
		j := size - 1
		for ; j > high || carry != 0; j-- {
			carry += 256 * int(buf[j])
			buf[j] = byte(carry % 58)
			carry /= 58
		}
		high = j
	}

	idx := 0
	for idx < size && buf[idx] == 0 {
		idx++
	}

	out := make([]byte, zeros+(size-idx))
	for i := range zeros {
		out[i] = base58Alphabet[0]
	}
	for i, v := range buf[idx:] {
		out[zeros+i] = base58Alphabet[v]
	}
	return string(out)
}

// Base58Decode is the inverse of Base58Encode. It errors on any character
// outside the alphabet rather than skipping it, so a malformed pubkey surfaces
// at the boundary instead of silently decoding to the wrong bytes.
func Base58Decode(input string) ([]byte, error) {
	if input == "" {
		return []byte{}, nil
	}

	zeros := 0
	for zeros < len(input) && input[zeros] == base58Alphabet[0] {
		zeros++
	}

	// log(58)/log(256) ~= 0.733; 733/1000 is the standard safe over-allocation.
	size := (len(input)-zeros)*733/1000 + 1
	buf := make([]byte, size)

	high := size - 1
	for i := zeros; i < len(input); i++ {
		d := base58Decode[input[i]]
		if d < 0 {
			return nil, fmt.Errorf("invalid base58 character %q at index %d", input[i], i)
		}
		carry := int(d)
		j := size - 1
		for ; j > high || carry != 0; j-- {
			carry += 58 * int(buf[j])
			buf[j] = byte(carry % 256)
			carry /= 256
		}
		high = j
	}

	idx := 0
	for idx < size && buf[idx] == 0 {
		idx++
	}

	out := make([]byte, zeros+(size-idx))
	copy(out[zeros:], buf[idx:])
	return out, nil
}
