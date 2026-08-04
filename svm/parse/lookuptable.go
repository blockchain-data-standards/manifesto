package parse

import "fmt"

// parseAddressLookupTable ports Agave's parse_address_lookup_table.rs. Wire
// layout is solana-address-lookup-table-interface's ProgramInstruction under
// bincode fixint LE: u32 discriminant, then fixed fields; Vec<Pubkey> is a
// u64 LE length + raw 32-byte keys. Trailing bytes are tolerated (see
// parseSystem).
//
// Quirk copied from the arm: extendLookupTable emits payerAccount AND
// systemProgram only when the instruction carries >= 4 accounts — both keys
// or neither, never one.
func parseAddressLookupTable(data []byte, accounts []string) (typeInfo, error) {
	r := reader{buf: data}
	disc, err := r.u32()
	if err != nil {
		return typeInfo{}, err
	}
	switch disc {
	case 0: // CreateLookupTable { recent_slot: u64, bump_seed: u8 }
		recentSlot, err1 := r.u64()
		bumpSeed, err2 := r.u8()
		if err := firstErr(err1, err2, checkNumAccounts(accounts, 4)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "createLookupTable", Info: map[string]any{
			"lookupTableAccount": accounts[0], "lookupTableAuthority": accounts[1],
			"payerAccount": accounts[2], "systemProgram": accounts[3],
			"recentSlot": recentSlot, "bumpSeed": bumpSeed,
		}}, nil
	case 1: // FreezeLookupTable
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "freezeLookupTable", Info: map[string]any{
			"lookupTableAccount": accounts[0], "lookupTableAuthority": accounts[1],
		}}, nil
	case 2: // ExtendLookupTable { new_addresses: Vec<Pubkey> }
		n, err := r.vecLen(32)
		if err != nil {
			return typeInfo{}, err
		}
		newAddresses := make([]string, n)
		for i := range newAddresses {
			if newAddresses[i], err = r.pubkey(); err != nil {
				return typeInfo{}, err
			}
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"lookupTableAccount": accounts[0], "lookupTableAuthority": accounts[1],
			"newAddresses": newAddresses,
		}
		if len(accounts) >= 4 {
			info["payerAccount"] = accounts[2]
			info["systemProgram"] = accounts[3]
		}
		return typeInfo{Type: "extendLookupTable", Info: info}, nil
	case 3: // DeactivateLookupTable
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "deactivateLookupTable", Info: map[string]any{
			"lookupTableAccount": accounts[0], "lookupTableAuthority": accounts[1],
		}}, nil
	case 4: // CloseLookupTable
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "closeLookupTable", Info: map[string]any{
			"lookupTableAccount": accounts[0], "lookupTableAuthority": accounts[1],
			"recipient": accounts[2],
		}}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: address-lookup-table discriminant %d", ErrNotParsable, disc)
}
