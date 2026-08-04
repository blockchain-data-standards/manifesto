package parse

// Port of Agave's parse_vote.rs (solana-transaction-status 2.3.13), wire
// layouts from solana-vote-interface 2.2.6. Bincode fixint LE, u32 enum
// discriminant, trailing bytes tolerated.
//
// Render quirks carried over verbatim from the Rust source:
//   - Lockout serializes with its Rust field names — "slot" and
//     "confirmation_count" (snake_case; the struct has no serde rename).
//   - VoteAuthorize is a plain serde unit enum: "Voter" / "Withdrawer".
//   - The updateVoteState/compactUpdateVoteState/towerSync families use
//     all-lowercase instruction_type strings ("updatevotestate", ...).
//   - InitializeAccount renders "node" from accounts[3]; the decoded
//     VoteInit.node_pubkey is never rendered (but must still parse).
//   - CompactUpdateVoteState/TowerSync are ALWAYS compact on the wire
//     (serde(with) applies to deserialization): root is a raw u64 where
//     u64::MAX means None, lockouts are a short_vec of
//     {varint offset, u8 confirmation_count} rebuilt into absolute slots.

import (
	"fmt"
	"math"
)

func parseVote(data []byte, accounts []string) (typeInfo, error) {
	r := &reader{buf: data}
	disc, err := r.u32()
	if err != nil {
		return typeInfo{}, err
	}
	switch disc {
	case 0: // InitializeAccount(VoteInit)
		_, err1 := r.pubkey() // node_pubkey: decoded, not rendered
		voter, err2 := r.pubkey()
		withdrawer, err3 := r.pubkey()
		commission, err4 := r.u8()
		if err := firstErr(err1, err2, err3, err4, checkNumAccounts(accounts, 4)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initialize", Info: map[string]any{
			"voteAccount": accounts[0], "rentSysvar": accounts[1],
			"clockSysvar": accounts[2], "node": accounts[3],
			"authorizedVoter": voter, "authorizedWithdrawer": withdrawer,
			"commission": commission,
		}}, nil
	case 1: // Authorize(Pubkey, VoteAuthorize)
		newAuthority, err1 := r.pubkey()
		authorityType, err2 := voteAuthorize(r)
		if err := firstErr(err1, err2, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "authorize", Info: map[string]any{
			"voteAccount": accounts[0], "clockSysvar": accounts[1],
			"authority": accounts[2], "newAuthority": newAuthority,
			"authorityType": authorityType,
		}}, nil
	case 2: // Vote(Vote)
		vote, err1 := voteData(r)
		if err := firstErr(err1, checkNumAccounts(accounts, 4)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "vote", Info: map[string]any{
			"voteAccount": accounts[0], "slotHashesSysvar": accounts[1],
			"clockSysvar": accounts[2], "voteAuthority": accounts[3],
			"vote": vote,
		}}, nil
	case 3: // Withdraw(u64)
		lamports, err1 := r.u64()
		if err := firstErr(err1, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "withdraw", Info: map[string]any{
			"voteAccount": accounts[0], "destination": accounts[1],
			"withdrawAuthority": accounts[2], "lamports": lamports,
		}}, nil
	case 4: // UpdateValidatorIdentity
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "updateValidatorIdentity", Info: map[string]any{
			"voteAccount": accounts[0], "newValidatorIdentity": accounts[1],
			"withdrawAuthority": accounts[2],
		}}, nil
	case 5: // UpdateCommission(u8)
		commission, err1 := r.u8()
		if err := firstErr(err1, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "updateCommission", Info: map[string]any{
			"voteAccount": accounts[0], "withdrawAuthority": accounts[1],
			"commission": commission,
		}}, nil
	case 6: // VoteSwitch(Vote, Hash)
		vote, err1 := voteData(r)
		hash, err2 := r.pubkey()
		if err := firstErr(err1, err2, checkNumAccounts(accounts, 4)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "voteSwitch", Info: map[string]any{
			"voteAccount": accounts[0], "slotHashesSysvar": accounts[1],
			"clockSysvar": accounts[2], "voteAuthority": accounts[3],
			"vote": vote, "hash": hash,
		}}, nil
	case 7: // AuthorizeChecked(VoteAuthorize)
		authorityType, err1 := voteAuthorize(r)
		if err := firstErr(err1, checkNumAccounts(accounts, 4)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "authorizeChecked", Info: map[string]any{
			"voteAccount": accounts[0], "clockSysvar": accounts[1],
			"authority": accounts[2], "newAuthority": accounts[3],
			"authorityType": authorityType,
		}}, nil
	case 8: // UpdateVoteState(VoteStateUpdate)
		vsu, err1 := voteStateUpdate(r)
		if err := firstErr(err1, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "updatevotestate", Info: map[string]any{
			"voteAccount": accounts[0], "voteAuthority": accounts[1],
			"voteStateUpdate": vsu,
		}}, nil
	case 9: // UpdateVoteStateSwitch(VoteStateUpdate, Hash)
		vsu, err1 := voteStateUpdate(r)
		hash, err2 := r.pubkey()
		if err := firstErr(err1, err2, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "updatevotestateswitch", Info: map[string]any{
			"voteAccount": accounts[0], "voteAuthority": accounts[1],
			"voteStateUpdate": vsu, "hash": hash,
		}}, nil
	case 10: // AuthorizeWithSeed(VoteAuthorizeWithSeedArgs)
		authorityType, err1 := voteAuthorize(r)
		owner, err2 := r.pubkey()
		seed, err3 := r.str()
		newAuthority, err4 := r.pubkey()
		if err := firstErr(err1, err2, err3, err4, checkNumAccounts(accounts, 3)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "authorizeWithSeed", Info: map[string]any{
			"voteAccount": accounts[0], "clockSysvar": accounts[1],
			"authorityBaseKey": accounts[2], "authorityOwner": owner,
			"authoritySeed": seed, "newAuthority": newAuthority,
			"authorityType": authorityType,
		}}, nil
	case 11: // AuthorizeCheckedWithSeed(VoteAuthorizeCheckedWithSeedArgs)
		authorityType, err1 := voteAuthorize(r)
		owner, err2 := r.pubkey()
		seed, err3 := r.str()
		if err := firstErr(err1, err2, err3, checkNumAccounts(accounts, 4)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "authorizeCheckedWithSeed", Info: map[string]any{
			"voteAccount": accounts[0], "clockSysvar": accounts[1],
			"authorityBaseKey": accounts[2], "authorityOwner": owner,
			"authoritySeed": seed, "newAuthority": accounts[3],
			"authorityType": authorityType,
		}}, nil
	case 12: // CompactUpdateVoteState(VoteStateUpdate)
		vsu, err1 := voteCompactStateUpdate(r)
		if err := firstErr(err1, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "compactupdatevotestate", Info: map[string]any{
			"voteAccount": accounts[0], "voteAuthority": accounts[1],
			"voteStateUpdate": vsu,
		}}, nil
	case 13: // CompactUpdateVoteStateSwitch(VoteStateUpdate, Hash)
		vsu, err1 := voteCompactStateUpdate(r)
		hash, err2 := r.pubkey()
		if err := firstErr(err1, err2, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "compactupdatevotestateswitch", Info: map[string]any{
			"voteAccount": accounts[0], "voteAuthority": accounts[1],
			"voteStateUpdate": vsu, "hash": hash,
		}}, nil
	case 14: // TowerSync(TowerSync)
		ts, err1 := voteTowerSyncData(r)
		if err := firstErr(err1, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "towersync", Info: map[string]any{
			"voteAccount": accounts[0], "voteAuthority": accounts[1],
			"towerSync": ts,
		}}, nil
	case 15: // TowerSyncSwitch(TowerSync, Hash)
		ts, err1 := voteTowerSyncData(r)
		hash, err2 := r.pubkey()
		if err := firstErr(err1, err2, checkNumAccounts(accounts, 2)); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "towersyncswitch", Info: map[string]any{
			"voteAccount": accounts[0], "voteAuthority": accounts[1],
			"towerSync": ts, "hash": hash,
		}}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: vote discriminant %d", ErrNotParsable, disc)
}

// voteAuthorize is VoteAuthorize's bincode form (u32 variant index) rendered
// the way serde_json renders the unit variant: "Voter" / "Withdrawer".
func voteAuthorize(r *reader) (string, error) {
	v, err := r.u32()
	if err != nil {
		return "", err
	}
	switch v {
	case 0:
		return "Voter", nil
	case 1:
		return "Withdrawer", nil
	}
	return "", fmt.Errorf("%w: vote authorize variant %d", ErrNotParsable, v)
}

// voteData is Vote{slots: Vec<u64>, hash: Hash, timestamp: Option<i64>},
// rendered as parse_vote.rs does (timestamp key present, null when None).
func voteData(r *reader) (map[string]any, error) {
	slots, err1 := r.vecU64()
	hash, err2 := r.pubkey()
	timestamp, err3 := r.optI64()
	if err := firstErr(err1, err2, err3); err != nil {
		return nil, err
	}
	return map[string]any{"slots": slots, "hash": hash, "timestamp": timestamp}, nil
}

// voteLockouts is bincode VecDeque<Lockout>: u64 length + n × (slot u64,
// confirmation_count u32). Lockout has no serde rename, so the JSON keys
// stay snake_case.
func voteLockouts(r *reader) ([]any, error) {
	n, err := r.vecLen(12)
	if err != nil {
		return nil, err
	}
	out := make([]any, n)
	for i := range out {
		slot, err1 := r.u64()
		cc, err2 := r.u32()
		if err := firstErr(err1, err2); err != nil {
			return nil, err
		}
		out[i] = map[string]any{"slot": slot, "confirmation_count": cc}
	}
	return out, nil
}

// voteStateUpdate is the NON-compact VoteStateUpdate wire form
// (UpdateVoteState / UpdateVoteStateSwitch): lockouts, Option<u64> root,
// hash, Option<i64> timestamp. root and timestamp render null when None.
func voteStateUpdate(r *reader) (map[string]any, error) {
	lockouts, err1 := voteLockouts(r)
	root, err2 := r.optU64()
	hash, err3 := r.pubkey()
	timestamp, err4 := r.optI64()
	if err := firstErr(err1, err2, err3, err4); err != nil {
		return nil, err
	}
	return map[string]any{
		"lockouts": lockouts, "root": root, "hash": hash, "timestamp": timestamp,
	}, nil
}

// voteCompactLockouts reads the compact prefix shared by
// serde_compact_vote_state_update and serde_tower_sync: root as a raw u64
// (u64::MAX = None) followed by a short_vec of {varint offset, u8
// confirmation_count}, rebuilt into absolute-slot lockouts exactly like the
// Rust deserializer (accumulator starts at root.unwrap_or_default(),
// checked_add on every offset).
func voteCompactLockouts(r *reader) ([]any, *uint64, error) {
	root, err := r.u64()
	if err != nil {
		return nil, nil, err
	}
	var rootOpt *uint64
	slot := uint64(0)
	if root != math.MaxUint64 {
		rootOpt = &root
		slot = root
	}
	n, err := voteShortU16(r)
	if err != nil {
		return nil, nil, err
	}
	if int(n) > (len(r.buf)-r.pos)/2 { // each offset entry is ≥ 2 bytes
		return nil, nil, fmt.Errorf("%w: lockout count %d exceeds remaining bytes", ErrNotParsable, n)
	}
	out := make([]any, n)
	for i := range out {
		offset, err1 := voteVarintU64(r)
		cc, err2 := r.u8()
		if err := firstErr(err1, err2); err != nil {
			return nil, nil, err
		}
		if offset > math.MaxUint64-slot {
			return nil, nil, fmt.Errorf("%w: invalid lockout offset", ErrNotParsable)
		}
		slot += offset
		out[i] = map[string]any{"slot": slot, "confirmation_count": uint32(cc)}
	}
	return out, rootOpt, nil
}

// voteCompactStateUpdate is CompactVoteStateUpdate on the wire, rendered in
// the same shape as the non-compact form.
func voteCompactStateUpdate(r *reader) (map[string]any, error) {
	lockouts, root, err := voteCompactLockouts(r)
	if err != nil {
		return nil, err
	}
	hash, err1 := r.pubkey()
	timestamp, err2 := r.optI64()
	if err := firstErr(err1, err2); err != nil {
		return nil, err
	}
	return map[string]any{
		"lockouts": lockouts, "root": root, "hash": hash, "timestamp": timestamp,
	}, nil
}

// voteTowerSyncData is CompactTowerSync: the compact state-update layout plus
// a trailing block_id hash, rendered with Agave's "blockId" key.
func voteTowerSyncData(r *reader) (map[string]any, error) {
	m, err := voteCompactStateUpdate(r)
	if err != nil {
		return nil, err
	}
	blockID, err := r.pubkey()
	if err != nil {
		return nil, err
	}
	m["blockId"] = blockID
	return m, nil
}

// voteShortU16 is solana-short-vec's compact-u16: 1-3 bytes of 7-bit LE
// groups. Strict like the crate: a zero continuation byte is an alias
// encoding (refused), byte three may not continue, and overflow past u16
// refuses.
func voteShortU16(r *reader) (uint16, error) {
	var val uint32
	for nth := range 3 {
		b, err := r.u8()
		if err != nil {
			return 0, err
		}
		if b == 0 && nth != 0 {
			return 0, fmt.Errorf("%w: short_vec alias encoding", ErrNotParsable)
		}
		done := b&0x80 == 0
		if nth == 2 && !done {
			return 0, fmt.Errorf("%w: short_vec byte three continues", ErrNotParsable)
		}
		val |= uint32(b&0x7F) << (7 * nth)
		if val > math.MaxUint16 {
			return 0, fmt.Errorf("%w: short_vec overflow %d", ErrNotParsable, val)
		}
		if done {
			return uint16(val), nil
		}
	}
	return 0, fmt.Errorf("%w: short_vec unterminated", ErrNotParsable) // unreachable
}

// voteVarintU64 is solana-serde-varint's u64: 7-bit LE groups, at most ten
// bytes. Strict like the crate: the last byte may not have been truncated by
// the shift, and a zero last byte is only legal as the sole byte of zero.
func voteVarintU64(r *reader) (uint64, error) {
	var out uint64
	for shift := uint(0); shift < 64; shift += 7 {
		b, err := r.u8()
		if err != nil {
			return 0, err
		}
		out |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			if byte(out>>shift) != b {
				return 0, fmt.Errorf("%w: varint last byte truncated", ErrNotParsable)
			}
			if b == 0 && (shift != 0 || out != 0) {
				return 0, fmt.Errorf("%w: varint trailing zeros", ErrNotParsable)
			}
			return out, nil
		}
	}
	return 0, fmt.Errorf("%w: varint shift overflow", ErrNotParsable)
}
