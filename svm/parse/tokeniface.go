package parse

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"unicode/utf8"

	"github.com/blockchain-data-standards/manifesto/svm"
)

// This file ports Agave's TLV-interface fallback chain from parse_token.rs
// (solana-transaction-status 2.3.13, lines 706-724): data that failed
// TokenInstruction::unpack is tried as a TokenGroupInstruction
// (spl-token-group-interface 0.6.0), then as a TokenMetadataInstruction
// (spl-token-metadata-interface 0.7.0), else it is not parsable.
//
// Both interfaces address instructions by an 8-byte SplDiscriminate
// discriminator: the first 8 bytes of sha256(seed), computed at init below
// from the exact #[discriminator_hash_input] seed strings in the crates.
//
// Payload strictness differs per interface and both REJECT trailing bytes
// (unlike the bincode programs):
//   - token-group payloads are Pod structs decoded via bytemuck
//     try_from_bytes, which demands the payload length equal the struct size
//     exactly. InitializeMember is zero-sized: its payload must be empty.
//   - token-metadata payloads are borsh via try_from_slice, which errors
//     with "Not all bytes read" on leftovers; bool and Option tags are
//     strictly 0/1.
// A payload decode failure means unpack itself failed, so the chain falls
// through. Once an unpack succeeds, an account-count error in the parse arm
// is final — no further fallback.

func tokenIfaceDiscriminator(seed string) [8]byte {
	sum := sha256.Sum256([]byte(seed))
	var d [8]byte
	copy(d[:], sum[:8])
	return d
}

var (
	tgDiscInitializeGroup    = tokenIfaceDiscriminator("spl_token_group_interface:initialize_token_group")
	tgDiscUpdateGroupMaxSize = tokenIfaceDiscriminator("spl_token_group_interface:update_group_max_size")
	tgDiscUpdateAuthority    = tokenIfaceDiscriminator("spl_token_group_interface:update_authority")
	tgDiscInitializeMember   = tokenIfaceDiscriminator("spl_token_group_interface:initialize_member")

	tmDiscInitialize      = tokenIfaceDiscriminator("spl_token_metadata_interface:initialize_account")
	tmDiscUpdateField     = tokenIfaceDiscriminator("spl_token_metadata_interface:updating_field")
	tmDiscRemoveKey       = tokenIfaceDiscriminator("spl_token_metadata_interface:remove_key_ix")
	tmDiscUpdateAuthority = tokenIfaceDiscriminator("spl_token_metadata_interface:update_the_authority")
	tmDiscEmit            = tokenIfaceDiscriminator("spl_token_metadata_interface:emitter")
)

func parseTokenInterface(data []byte, accounts []string) (typeInfo, error) {
	if parsed, ok, err := parseTokenGroupInstruction(data, accounts); ok {
		return parsed, err
	}
	if parsed, ok, err := parseTokenMetadataInstruction(data, accounts); ok {
		return parsed, err
	}
	return typeInfo{}, fmt.Errorf("%w: not a token group or token metadata instruction", ErrNotParsable)
}

// parseTokenGroupInstruction mirrors TokenGroupInstruction::unpack plus
// parse_token_group.rs. ok reports whether unpack succeeded; when false the
// caller falls through to the next interface, when true any err is final.
func parseTokenGroupInstruction(data []byte, accounts []string) (typeInfo, bool, error) {
	if len(data) < 8 {
		return typeInfo{}, false, nil
	}
	var disc [8]byte
	copy(disc[:], data)
	rest := data[8:]
	switch disc {
	case tgDiscInitializeGroup:
		// Pod{update_authority: OptionalNonZeroPubkey, max_size: PodU64}:
		// exactly 40 bytes.
		if len(rest) != 40 {
			return typeInfo{}, false, nil
		}
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, true, err
		}
		return typeInfo{
			Type: "initializeTokenGroup",
			Info: map[string]any{
				"group":           accounts[0],
				"maxSize":         binary.LittleEndian.Uint64(rest[32:40]),
				"mint":            accounts[1],
				"mintAuthority":   accounts[2],
				"updateAuthority": tokenIfaceOptionalPubkey(rest[:32]),
			},
		}, true, nil
	case tgDiscUpdateGroupMaxSize:
		// Pod{max_size: PodU64}: exactly 8 bytes.
		if len(rest) != 8 {
			return typeInfo{}, false, nil
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, true, err
		}
		return typeInfo{
			Type: "updateTokenGroupMaxSize",
			Info: map[string]any{
				"group":           accounts[0],
				"maxSize":         binary.LittleEndian.Uint64(rest),
				"updateAuthority": accounts[1],
			},
		}, true, nil
	case tgDiscUpdateAuthority:
		// Pod{new_authority: OptionalNonZeroPubkey}: exactly 32 bytes.
		if len(rest) != 32 {
			return typeInfo{}, false, nil
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, true, err
		}
		return typeInfo{
			Type: "updateTokenGroupAuthority",
			Info: map[string]any{
				"group":           accounts[0],
				"updateAuthority": accounts[1],
				"newAuthority":    tokenIfaceOptionalPubkey(rest),
			},
		}, true, nil
	case tgDiscInitializeMember:
		// InitializeMember is a zero-sized Pod: payload must be empty.
		if len(rest) != 0 {
			return typeInfo{}, false, nil
		}
		if err := checkNumAccounts(accounts, 5); err != nil {
			return typeInfo{}, true, err
		}
		return typeInfo{
			Type: "initializeTokenGroupMember",
			Info: map[string]any{
				"member":               accounts[0],
				"memberMint":           accounts[1],
				"memberMintAuthority":  accounts[2],
				"group":                accounts[3],
				"groupUpdateAuthority": accounts[4],
			},
		}, true, nil
	}
	return typeInfo{}, false, nil
}

// parseTokenMetadataInstruction mirrors TokenMetadataInstruction::unpack
// plus parse_token_metadata.rs, with the same ok/err contract as
// parseTokenGroupInstruction.
func parseTokenMetadataInstruction(data []byte, accounts []string) (typeInfo, bool, error) {
	if len(data) < 8 {
		return typeInfo{}, false, nil
	}
	var disc [8]byte
	copy(disc[:], data)
	rest := data[8:]
	switch disc {
	case tmDiscInitialize:
		name, r, ok := tokenIfaceBorshString(rest)
		if !ok {
			return typeInfo{}, false, nil
		}
		symbol, r, ok := tokenIfaceBorshString(r)
		if !ok {
			return typeInfo{}, false, nil
		}
		uri, r, ok := tokenIfaceBorshString(r)
		if !ok || len(r) != 0 {
			return typeInfo{}, false, nil
		}
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, true, err
		}
		return typeInfo{
			Type: "initializeTokenMetadata",
			Info: map[string]any{
				"metadata":        accounts[0],
				"updateAuthority": accounts[1],
				"mint":            accounts[2],
				"mintAuthority":   accounts[3],
				"name":            name,
				"symbol":          symbol,
				"uri":             uri,
			},
		}, true, nil
	case tmDiscUpdateField:
		// Field is a borsh enum: 0 Name, 1 Symbol, 2 Uri, 3 Key(String).
		if len(rest) < 1 {
			return typeInfo{}, false, nil
		}
		var field string
		r := rest[1:]
		switch rest[0] {
		case 0:
			field = "name"
		case 1:
			field = "symbol"
		case 2:
			field = "uri"
		case 3:
			var ok bool
			field, r, ok = tokenIfaceBorshString(r)
			if !ok {
				return typeInfo{}, false, nil
			}
		default:
			return typeInfo{}, false, nil
		}
		value, r, ok := tokenIfaceBorshString(r)
		if !ok || len(r) != 0 {
			return typeInfo{}, false, nil
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, true, err
		}
		return typeInfo{
			Type: "updateTokenMetadataField",
			Info: map[string]any{
				"metadata":        accounts[0],
				"updateAuthority": accounts[1],
				"field":           field,
				"value":           value,
			},
		}, true, nil
	case tmDiscRemoveKey:
		// Struct field order puts idempotent BEFORE key on the wire.
		if len(rest) < 1 || rest[0] > 1 {
			return typeInfo{}, false, nil
		}
		idempotent := rest[0] == 1
		key, r, ok := tokenIfaceBorshString(rest[1:])
		if !ok || len(r) != 0 {
			return typeInfo{}, false, nil
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, true, err
		}
		return typeInfo{
			Type: "removeTokenMetadataKey",
			Info: map[string]any{
				"metadata":        accounts[0],
				"updateAuthority": accounts[1],
				"key":             key,
				"idempotent":      idempotent,
			},
		}, true, nil
	case tmDiscUpdateAuthority:
		// Borsh OptionalNonZeroPubkey is the raw 32 bytes: exactly 32.
		if len(rest) != 32 {
			return typeInfo{}, false, nil
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, true, err
		}
		return typeInfo{
			Type: "updateTokenMetadataAuthority",
			Info: map[string]any{
				"metadata":        accounts[0],
				"updateAuthority": accounts[1],
				"newAuthority":    tokenIfaceOptionalPubkey(rest),
			},
		}, true, nil
	case tmDiscEmit:
		start, r, ok := tokenIfaceBorshOptionU64(rest)
		if !ok {
			return typeInfo{}, false, nil
		}
		end, r, ok := tokenIfaceBorshOptionU64(r)
		if !ok || len(r) != 0 {
			return typeInfo{}, false, nil
		}
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, true, err
		}
		info := map[string]any{"metadata": accounts[0]}
		// Unlike the null-rendering authority fields, the Emit arm OMITS
		// start/end when None (conditional map.insert in the Rust).
		if start != nil {
			info["start"] = *start
		}
		if end != nil {
			info["end"] = *end
		}
		return typeInfo{Type: "emitTokenMetadata", Info: info}, true, nil
	}
	return typeInfo{}, false, nil
}

// tokenIfaceOptionalPubkey renders spl-pod's OptionalNonZeroPubkey the way
// the parse arms do: Option::<Pubkey>::from + map(to_string), i.e. a base58
// string, or null when all 32 bytes are zero. The key stays present.
func tokenIfaceOptionalPubkey(b []byte) any {
	for _, c := range b {
		if c != 0 {
			return svm.Base58Encode(b)
		}
	}
	return nil
}

// tokenIfaceBorshString reads a borsh string: u32 LE length + bytes, which
// borsh requires to be valid utf8.
func tokenIfaceBorshString(b []byte) (string, []byte, bool) {
	if len(b) < 4 {
		return "", nil, false
	}
	n := binary.LittleEndian.Uint32(b)
	b = b[4:]
	if uint64(len(b)) < uint64(n) {
		return "", nil, false
	}
	if !utf8.Valid(b[:n]) {
		return "", nil, false
	}
	return string(b[:n]), b[n:], true
}

// tokenIfaceBorshOptionU64 reads a borsh Option<u64>: a u8 tag that is
// strictly 0 or 1, then 8 LE bytes when 1.
func tokenIfaceBorshOptionU64(b []byte) (*uint64, []byte, bool) {
	if len(b) < 1 {
		return nil, nil, false
	}
	switch b[0] {
	case 0:
		return nil, b[1:], true
	case 1:
		if len(b) < 9 {
			return nil, nil, false
		}
		v := binary.LittleEndian.Uint64(b[1:9])
		return &v, b[9:], true
	}
	return nil, nil, false
}
