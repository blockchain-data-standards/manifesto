package parse

import (
	"encoding/base64"
	"fmt"
)

// This file ports Agave's parse_bpf_loader.rs: both the v2 loader
// (solana-loader-v2-interface's LoaderInstruction) and the upgradeable v3
// loader (solana-loader-v3-interface 5.0.0's UpgradeableLoaderInstruction).
// Wire layout is bincode fixint LE: u32 discriminant, then fixed fields;
// Vec<u8> is a u64 LE length + raw bytes; usize serializes as u64. Trailing
// bytes are tolerated (see parseSystem).
//
// `bytes` renders exactly as Agave's BASE64_STANDARD.encode: the standard
// alphabet WITH padding — Go's base64.StdEncoding.

// bpfLoaderBytes decodes a bincode Vec<u8> program-data payload.
func bpfLoaderBytes(r *reader) ([]byte, error) {
	n, err := r.vecLen(1)
	if err != nil {
		return nil, err
	}
	return r.take(n)
}

// bpfLoaderOptAccount renders Agave's present-and-null optional trailing
// accounts: the key is always emitted, its value null when absent. Contrast
// initializeBuffer's authority, which is OMITTED when absent.
func bpfLoaderOptAccount(accounts []string, i int) any {
	if i < len(accounts) {
		return accounts[i]
	}
	return nil
}

// parseBpfLoader handles the non-upgradeable v2 loader.
func parseBpfLoader(data []byte, accounts []string) (typeInfo, error) {
	r := reader{buf: data}
	disc, err := r.u32()
	if err != nil {
		return typeInfo{}, err
	}
	switch disc {
	case 0: // Write { offset: u32, bytes: Vec<u8> }
		offset, err := r.u32()
		if err != nil {
			return typeInfo{}, err
		}
		bytes, err := bpfLoaderBytes(&r)
		if err := firstErr(err, checkNumAccounts(accounts, 1)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "write", Info: map[string]any{
			"offset":  offset,
			"bytes":   base64.StdEncoding.EncodeToString(bytes),
			"account": accounts[0],
		}}, nil
	case 1: // Finalize
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "finalize", Info: map[string]any{
			"account": accounts[0],
		}}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: bpf-loader discriminant %d", ErrNotParsable, disc)
}

// parseBpfUpgradeableLoader handles the v3 upgradeable loader — all ten
// UpgradeableLoaderInstruction variants of the pinned 5.0.0 interface.
func parseBpfUpgradeableLoader(data []byte, accounts []string) (typeInfo, error) {
	r := reader{buf: data}
	disc, err := r.u32()
	if err != nil {
		return typeInfo{}, err
	}
	switch disc {
	case 0: // InitializeBuffer
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{"account": accounts[0]}
		if len(accounts) > 1 { // authority OMITTED (not null) when absent
			info["authority"] = accounts[1]
		}
		return typeInfo{Type: "initializeBuffer", Info: info}, nil
	case 1: // Write { offset: u32, bytes: Vec<u8> }
		offset, err := r.u32()
		if err != nil {
			return typeInfo{}, err
		}
		bytes, err := bpfLoaderBytes(&r)
		if err := firstErr(err, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "write", Info: map[string]any{
			"offset":  offset,
			"bytes":   base64.StdEncoding.EncodeToString(bytes),
			"account": accounts[0], "authority": accounts[1],
		}}, nil
	case 2: // DeployWithMaxDataLen { max_data_len: usize (u64 on wire) }
		maxDataLen, err := r.u64()
		if err := firstErr(err, checkNumAccounts(accounts, 8)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "deployWithMaxDataLen", Info: map[string]any{
			"maxDataLen":         maxDataLen,
			"payerAccount":       accounts[0],
			"programDataAccount": accounts[1],
			"programAccount":     accounts[2],
			"bufferAccount":      accounts[3],
			"rentSysvar":         accounts[4],
			"clockSysvar":        accounts[5],
			"systemProgram":      accounts[6],
			"authority":          accounts[7],
		}}, nil
	case 3: // Upgrade
		if err := checkNumAccounts(accounts, 7); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "upgrade", Info: map[string]any{
			"programDataAccount": accounts[0],
			"programAccount":     accounts[1],
			"bufferAccount":      accounts[2],
			"spillAccount":       accounts[3],
			"rentSysvar":         accounts[4],
			"clockSysvar":        accounts[5],
			"authority":          accounts[6],
		}}, nil
	case 4: // SetAuthority
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "setAuthority", Info: map[string]any{
			"account": accounts[0], "authority": accounts[1],
			"newAuthority": bpfLoaderOptAccount(accounts, 2),
		}}, nil
	case 5: // Close
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "close", Info: map[string]any{
			"account": accounts[0], "recipient": accounts[1], "authority": accounts[2],
			"programAccount": bpfLoaderOptAccount(accounts, 3),
		}}, nil
	case 6: // ExtendProgram { additional_bytes: u32 }
		additionalBytes, err := r.u32()
		if err := firstErr(err, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "extendProgram", Info: map[string]any{
			"additionalBytes":    additionalBytes,
			"programDataAccount": accounts[0],
			"programAccount":     accounts[1],
			"systemProgram":      bpfLoaderOptAccount(accounts, 2),
			"payerAccount":       bpfLoaderOptAccount(accounts, 3),
		}}, nil
	case 7: // SetAuthorityChecked
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "setAuthorityChecked", Info: map[string]any{
			"account": accounts[0], "authority": accounts[1],
			"newAuthority": accounts[2],
		}}, nil
	case 8: // Migrate
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "migrate", Info: map[string]any{
			"programDataAccount": accounts[0],
			"programAccount":     accounts[1],
			"authority":          accounts[2],
		}}, nil
	case 9: // ExtendProgramChecked { additional_bytes: u32 }
		additionalBytes, err := r.u32()
		if err := firstErr(err, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "extendProgramChecked", Info: map[string]any{
			"additionalBytes":    additionalBytes,
			"programDataAccount": accounts[0],
			"programAccount":     accounts[1],
			"authority":          accounts[2],
			"systemProgram":      bpfLoaderOptAccount(accounts, 3),
			"payerAccount":       bpfLoaderOptAccount(accounts, 4),
		}}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: bpf-upgradeable-loader discriminant %d", ErrNotParsable, disc)
}
