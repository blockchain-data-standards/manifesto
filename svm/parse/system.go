package parse

import (
	"encoding/binary"
	"fmt"
	"unicode/utf8"

	"github.com/blockchain-data-standards/manifesto/svm"
)

// parseSystem ports Agave's parse_system.rs. Wire layout is what
// bincode::deserialize reads: fixint LE — u32 discriminant, then fixed
// fields; strings are u64 length + utf8; pubkeys are raw 32 bytes. TRAILING
// BYTES ARE TOLERATED: bincode::deserialize is fixint + allow_trailing_bytes
// (bincode 1.3.3 lib.rs:177), so a payload longer than the decoded
// instruction still parses. This applies to every bincode program here
// (system, vote, stake, address-lookup-table, both loaders).
//
// Instruction-type strings are Agave's exactly — note the nonce family drops
// the "Account" suffix ("advanceNonce", "withdrawFromNonce", …), which no
// amount of guessing from the enum names would produce.
func parseSystem(data []byte, accounts []string) (typeInfo, error) {
	r := reader{buf: data}
	disc, err := r.u32()
	if err != nil {
		return typeInfo{}, err
	}
	switch disc {
	case 0: // CreateAccount
		lamports, err1 := r.u64()
		space, err2 := r.u64()
		owner, err3 := r.pubkey()
		if err := firstErr(err1, err2, err3, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "createAccount", Info: map[string]any{
			"source": accounts[0], "newAccount": accounts[1],
			"lamports": lamports, "space": space, "owner": owner,
		}}, nil
	case 1: // Assign
		owner, err1 := r.pubkey()
		if err := firstErr(err1, checkNumAccounts(accounts, 1)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "assign", Info: map[string]any{
			"account": accounts[0], "owner": owner,
		}}, nil
	case 2: // Transfer
		lamports, err1 := r.u64()
		if err := firstErr(err1, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "transfer", Info: map[string]any{
			"source": accounts[0], "destination": accounts[1], "lamports": lamports,
		}}, nil
	case 3: // CreateAccountWithSeed
		base, err1 := r.pubkey()
		seed, err2 := r.str()
		lamports, err3 := r.u64()
		space, err4 := r.u64()
		owner, err5 := r.pubkey()
		if err := firstErr(err1, err2, err3, err4, err5, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "createAccountWithSeed", Info: map[string]any{
			"source": accounts[0], "newAccount": accounts[1], "base": base,
			"seed": seed, "lamports": lamports, "space": space, "owner": owner,
		}}, nil
	case 4: // AdvanceNonceAccount
		if err := firstErr(checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "advanceNonce", Info: map[string]any{
			"nonceAccount": accounts[0], "recentBlockhashesSysvar": accounts[1],
			"nonceAuthority": accounts[2],
		}}, nil
	case 5: // WithdrawNonceAccount
		lamports, err1 := r.u64()
		if err := firstErr(err1, checkNumAccounts(accounts, 5)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "withdrawFromNonce", Info: map[string]any{
			"nonceAccount": accounts[0], "destination": accounts[1],
			"recentBlockhashesSysvar": accounts[2], "rentSysvar": accounts[3],
			"nonceAuthority": accounts[4], "lamports": lamports,
		}}, nil
	case 6: // InitializeNonceAccount
		authority, err1 := r.pubkey()
		if err := firstErr(err1, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializeNonce", Info: map[string]any{
			"nonceAccount": accounts[0], "recentBlockhashesSysvar": accounts[1],
			"rentSysvar": accounts[2], "nonceAuthority": authority,
		}}, nil
	case 7: // AuthorizeNonceAccount
		authority, err1 := r.pubkey()
		if err := firstErr(err1, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "authorizeNonce", Info: map[string]any{
			"nonceAccount": accounts[0], "nonceAuthority": accounts[1],
			"newAuthorized": authority,
		}}, nil
	case 8: // Allocate
		space, err1 := r.u64()
		if err := firstErr(err1, checkNumAccounts(accounts, 1)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "allocate", Info: map[string]any{
			"account": accounts[0], "space": space,
		}}, nil
	case 9: // AllocateWithSeed
		base, err1 := r.pubkey()
		seed, err2 := r.str()
		space, err3 := r.u64()
		owner, err4 := r.pubkey()
		if err := firstErr(err1, err2, err3, err4, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "allocateWithSeed", Info: map[string]any{
			"account": accounts[0], "base": base, "seed": seed,
			"space": space, "owner": owner,
		}}, nil
	case 10: // AssignWithSeed
		base, err1 := r.pubkey()
		seed, err2 := r.str()
		owner, err3 := r.pubkey()
		if err := firstErr(err1, err2, err3, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "assignWithSeed", Info: map[string]any{
			"account": accounts[0], "base": base, "seed": seed, "owner": owner,
		}}, nil
	case 11: // TransferWithSeed
		lamports, err1 := r.u64()
		seed, err2 := r.str()
		owner, err3 := r.pubkey()
		if err := firstErr(err1, err2, err3, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "transferWithSeed", Info: map[string]any{
			"source": accounts[0], "sourceBase": accounts[1], "destination": accounts[2],
			"lamports": lamports, "sourceSeed": seed, "sourceOwner": owner,
		}}, nil
	case 12: // UpgradeNonceAccount
		if err := firstErr(checkNumAccounts(accounts, 1)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "upgradeNonce", Info: map[string]any{
			"nonceAccount": accounts[0],
		}}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: system discriminant %d", ErrNotParsable, disc)
}

// reader decodes the bincode subset the system program uses. Every method
// error is ErrNotParsable-wrapped so the caller's fallback rule holds.
type reader struct {
	buf []byte
	pos int
}

func (r *reader) take(n int) ([]byte, error) {
	if r.pos+n > len(r.buf) {
		return nil, fmt.Errorf("%w: truncated at byte %d", ErrNotParsable, r.pos)
	}
	b := r.buf[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *reader) u32() (uint32, error) {
	b, err := r.take(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (r *reader) u64() (uint64, error) {
	b, err := r.take(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func (r *reader) pubkey() (string, error) {
	b, err := r.take(32)
	if err != nil {
		return "", err
	}
	return svm.Base58Encode(b), nil
}

// str is a bincode string: u64 LE length + utf8 bytes. bincode enforces utf8;
// so do we — a non-utf8 seed falls back rather than emitting mojibake.
func (r *reader) str() (string, error) {
	n, err := r.u64()
	if err != nil {
		return "", err
	}
	if n > uint64(len(r.buf)-r.pos) {
		return "", fmt.Errorf("%w: string length %d exceeds remaining bytes", ErrNotParsable, n)
	}
	b, err := r.take(int(n))
	if err != nil {
		return "", err
	}
	if !utf8.Valid(b) {
		return "", fmt.Errorf("%w: string is not valid utf8", ErrNotParsable)
	}
	return string(b), nil
}

func (r *reader) u8() (byte, error) {
	b, err := r.take(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *reader) i64() (int64, error) {
	v, err := r.u64()
	return int64(v), err
}

// optTag is bincode's Option<T>: u8 0 (None) / 1 (Some); anything else refuses.
func (r *reader) optTag() (bool, error) {
	b, err := r.u8()
	if err != nil {
		return false, err
	}
	switch b {
	case 0:
		return false, nil
	case 1:
		return true, nil
	}
	return false, fmt.Errorf("%w: option tag %d", ErrNotParsable, b)
}

func (r *reader) optU64() (*uint64, error) {
	some, err := r.optTag()
	if err != nil || !some {
		return nil, err
	}
	v, err := r.u64()
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *reader) optI64() (*int64, error) {
	some, err := r.optTag()
	if err != nil || !some {
		return nil, err
	}
	v, err := r.i64()
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *reader) optPubkey() (*string, error) {
	some, err := r.optTag()
	if err != nil || !some {
		return nil, err
	}
	v, err := r.pubkey()
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// vecLen is bincode's Vec length prefix (u64 LE), bounded by the bytes that
// could possibly remain (elemMin ≥ 1) so hostile lengths refuse instead of
// allocating.
func (r *reader) vecLen(elemMin int) (int, error) {
	n, err := r.u64()
	if err != nil {
		return 0, err
	}
	if n > uint64((len(r.buf)-r.pos)/elemMin) {
		return 0, fmt.Errorf("%w: vec length %d exceeds remaining bytes", ErrNotParsable, n)
	}
	return int(n), nil
}

func (r *reader) vecU64() ([]uint64, error) {
	n, err := r.vecLen(8)
	if err != nil {
		return nil, err
	}
	out := make([]uint64, n)
	for i := range out {
		if out[i], err = r.u64(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
