package parse

import "fmt"

// parseStake ports Agave's parse_stake.rs (solana-transaction-status 2.3.13)
// against solana-stake-interface 1.2.1. Wire layout is bincode fixint LE:
// u32 discriminant, then fixed fields; Option<T> is a u8 tag 0/1; String is
// u64 length + utf8; pubkeys are raw 32 bytes. Trailing bytes are tolerated
// (see parseSystem's header).
//
// Rendering quirks copied verbatim from the Rust arms:
//   - StakeAuthorize serializes with NO serde rename: the JSON values are
//     capitalized "Staker" / "Withdrawer".
//   - setLockup/setLockupChecked build the "lockup" object with keys inserted
//     only for Some(_) fields — None fields are OMITTED, not null. An
//     all-None args renders "lockup": {}.
//   - Several arms append optional trailing accounts (custodian, clockSysvar)
//     keyed on the instruction's account count.
//   - getMinimumDelegation has no account check of its own and renders a null
//     info (Info nil here — omitted by typeInfo's omitempty).
func parseStake(data []byte, accounts []string) (typeInfo, error) {
	r := reader{buf: data}
	disc, err := r.u32()
	if err != nil {
		return typeInfo{}, err
	}
	switch disc {
	case 0: // Initialize(Authorized, Lockup)
		staker, err1 := r.pubkey()
		withdrawer, err2 := r.pubkey()
		unixTimestamp, err3 := r.i64()
		epoch, err4 := r.u64()
		custodian, err5 := r.pubkey()
		if err := firstErr(err1, err2, err3, err4, err5, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initialize", Info: map[string]any{
			"stakeAccount": accounts[0], "rentSysvar": accounts[1],
			"authorized": map[string]any{
				"staker": staker, "withdrawer": withdrawer,
			},
			"lockup": map[string]any{
				"unixTimestamp": unixTimestamp, "epoch": epoch, "custodian": custodian,
			},
		}}, nil
	case 1: // Authorize(Pubkey, StakeAuthorize)
		newAuthorized, err1 := r.pubkey()
		authorityType, err2 := stakeAuthorize(&r)
		if err := firstErr(err1, err2, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"stakeAccount": accounts[0], "clockSysvar": accounts[1],
			"authority": accounts[2], "newAuthority": newAuthorized,
			"authorityType": authorityType,
		}
		if len(accounts) >= 4 {
			info["custodian"] = accounts[3]
		}
		return typeInfo{Type: "authorize", Info: info}, nil
	case 2: // DelegateStake
		if err := checkNumAccounts(accounts, 6); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "delegate", Info: map[string]any{
			"stakeAccount": accounts[0], "voteAccount": accounts[1],
			"clockSysvar": accounts[2], "stakeHistorySysvar": accounts[3],
			"stakeConfigAccount": accounts[4], "stakeAuthority": accounts[5],
		}}, nil
	case 3: // Split(u64)
		lamports, err1 := r.u64()
		if err := firstErr(err1, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "split", Info: map[string]any{
			"stakeAccount": accounts[0], "newSplitAccount": accounts[1],
			"stakeAuthority": accounts[2], "lamports": lamports,
		}}, nil
	case 4: // Withdraw(u64)
		lamports, err1 := r.u64()
		if err := firstErr(err1, checkNumAccounts(accounts, 5)); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"stakeAccount": accounts[0], "destination": accounts[1],
			"clockSysvar": accounts[2], "stakeHistorySysvar": accounts[3],
			"withdrawAuthority": accounts[4], "lamports": lamports,
		}
		if len(accounts) >= 6 {
			info["custodian"] = accounts[5]
		}
		return typeInfo{Type: "withdraw", Info: info}, nil
	case 5: // Deactivate
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "deactivate", Info: map[string]any{
			"stakeAccount": accounts[0], "clockSysvar": accounts[1],
			"stakeAuthority": accounts[2],
		}}, nil
	case 6: // SetLockup(LockupArgs)
		unixTimestamp, err1 := r.optI64()
		epoch, err2 := r.optU64()
		custodian, err3 := r.optPubkey()
		if err := firstErr(err1, err2, err3, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		lockup := map[string]any{}
		if unixTimestamp != nil {
			lockup["unixTimestamp"] = *unixTimestamp
		}
		if epoch != nil {
			lockup["epoch"] = *epoch
		}
		if custodian != nil {
			lockup["custodian"] = *custodian
		}
		return typeInfo{Type: "setLockup", Info: map[string]any{
			"stakeAccount": accounts[0], "custodian": accounts[1], "lockup": lockup,
		}}, nil
	case 7: // Merge
		if err := checkNumAccounts(accounts, 5); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "merge", Info: map[string]any{
			"destination": accounts[0], "source": accounts[1],
			"clockSysvar": accounts[2], "stakeHistorySysvar": accounts[3],
			"stakeAuthority": accounts[4],
		}}, nil
	case 8: // AuthorizeWithSeed(AuthorizeWithSeedArgs)
		newAuthorized, err1 := r.pubkey()
		authorityType, err2 := stakeAuthorize(&r)
		authoritySeed, err3 := r.str()
		authorityOwner, err4 := r.pubkey()
		if err := firstErr(err1, err2, err3, err4, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"stakeAccount": accounts[0], "authorityBase": accounts[1],
			"newAuthorized": newAuthorized, "authorityType": authorityType,
			"authoritySeed": authoritySeed, "authorityOwner": authorityOwner,
		}
		if len(accounts) >= 3 {
			info["clockSysvar"] = accounts[2]
		}
		if len(accounts) >= 4 {
			info["custodian"] = accounts[3]
		}
		return typeInfo{Type: "authorizeWithSeed", Info: info}, nil
	case 9: // InitializeChecked
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializeChecked", Info: map[string]any{
			"stakeAccount": accounts[0], "rentSysvar": accounts[1],
			"staker": accounts[2], "withdrawer": accounts[3],
		}}, nil
	case 10: // AuthorizeChecked(StakeAuthorize)
		authorityType, err1 := stakeAuthorize(&r)
		if err := firstErr(err1, checkNumAccounts(accounts, 4)); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"stakeAccount": accounts[0], "clockSysvar": accounts[1],
			"authority": accounts[2], "newAuthority": accounts[3],
			"authorityType": authorityType,
		}
		if len(accounts) >= 5 {
			info["custodian"] = accounts[4]
		}
		return typeInfo{Type: "authorizeChecked", Info: info}, nil
	case 11: // AuthorizeCheckedWithSeed(AuthorizeCheckedWithSeedArgs)
		authorityType, err1 := stakeAuthorize(&r)
		authoritySeed, err2 := r.str()
		authorityOwner, err3 := r.pubkey()
		if err := firstErr(err1, err2, err3, checkNumAccounts(accounts, 4)); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"stakeAccount": accounts[0], "authorityBase": accounts[1],
			"clockSysvar": accounts[2], "newAuthorized": accounts[3],
			"authorityType": authorityType, "authoritySeed": authoritySeed,
			"authorityOwner": authorityOwner,
		}
		if len(accounts) >= 5 {
			info["custodian"] = accounts[4]
		}
		return typeInfo{Type: "authorizeCheckedWithSeed", Info: info}, nil
	case 12: // SetLockupChecked(LockupCheckedArgs)
		unixTimestamp, err1 := r.optI64()
		epoch, err2 := r.optU64()
		if err := firstErr(err1, err2, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		lockup := map[string]any{}
		if unixTimestamp != nil {
			lockup["unixTimestamp"] = *unixTimestamp
		}
		if epoch != nil {
			lockup["epoch"] = *epoch
		}
		if len(accounts) >= 3 {
			lockup["custodian"] = accounts[2]
		}
		return typeInfo{Type: "setLockupChecked", Info: map[string]any{
			"stakeAccount": accounts[0], "custodian": accounts[1], "lockup": lockup,
		}}, nil
	case 13: // GetMinimumDelegation — no accounts, null info (Rust: Value::default()).
		return typeInfo{Type: "getMinimumDelegation"}, nil
	case 14: // DeactivateDelinquent
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "deactivateDelinquent", Info: map[string]any{
			"stakeAccount": accounts[0], "voteAccount": accounts[1],
			"referenceVoteAccount": accounts[2],
		}}, nil
	case 15: // Redelegate (deprecated in 2.1.0 but still in the enum and parsed)
		if err := checkNumAccounts(accounts, 5); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "redelegate", Info: map[string]any{
			"stakeAccount": accounts[0], "newStakeAccount": accounts[1],
			"voteAccount": accounts[2], "stakeConfigAccount": accounts[3],
			"stakeAuthority": accounts[4],
		}}, nil
	case 16: // MoveStake(u64)
		lamports, err1 := r.u64()
		if err := firstErr(err1, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "moveStake", Info: map[string]any{
			"source": accounts[0], "destination": accounts[1],
			"stakeAuthority": accounts[2], "lamports": lamports,
		}}, nil
	case 17: // MoveLamports(u64)
		lamports, err1 := r.u64()
		if err := firstErr(err1, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "moveLamports", Info: map[string]any{
			"source": accounts[0], "destination": accounts[1],
			"stakeAuthority": accounts[2], "lamports": lamports,
		}}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: stake discriminant %d", ErrNotParsable, disc)
}

// stakeAuthorize decodes StakeAuthorize (bincode u32 enum discriminant) and
// renders it as serde does with no rename attribute: the capitalized variant
// name — "Staker" / "Withdrawer".
func stakeAuthorize(r *reader) (string, error) {
	disc, err := r.u32()
	if err != nil {
		return "", err
	}
	switch disc {
	case 0:
		return "Staker", nil
	case 1:
		return "Withdrawer", nil
	}
	return "", fmt.Errorf("%w: stake authorize discriminant %d", ErrNotParsable, disc)
}
