package parse_test

// Stake program vectors: every StakeInstruction variant (discriminants
// 0..=17), wire layouts hand-encoded from solana-stake-interface 1.2.1,
// expected JSON hand-written from parse_stake.rs.
//
// Quirks pinned here, all verified against the Rust source:
//   - authorityType is the RAW serde unit enum: "Staker" / "Withdrawer"
//     (capitalized; StakeAuthorize has no rename attribute, state.rs:272).
//   - setLockup/setLockupChecked insert lockup keys only for Some(_) fields:
//     None fields are OMITTED (never null); all-None renders "lockup": {}.
//   - setLockupChecked renders TWO custodians: the top-level "custodian" is
//     accounts[1] (the signing authority), while the optional accounts[2]
//     lands INSIDE the lockup object as lockup.custodian
//     (parse_stake.rs:247-268).
//   - authorizeWithSeed's clockSysvar is itself optional (accounts[2] when
//     present) — the only arm where the clock is not mandatory.
//   - getMinimumDelegation renders a null info (the key is omitted) and has
//     no account check of its own; only Parse's empty-accounts guard applies.
//   - merge names accounts[0]/[1] destination/source; moveStake and
//     moveLamports name them source/destination.

import (
	"errors"
	"fmt"
	"testing"

	"github.com/blockchain-data-standards/manifesto/svm/parse"
)

func TestStakeVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			name:      "initialize with negative unixTimestamp",
			programID: parse.StakeProgramID,
			data:      cat(le32(0), kb(0x61), kb(0x62), i64le(-5), le64(20), kb(0x63)),
			accounts:  accs(1, 2),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"initialize","info":{
				"stakeAccount":%q,"rentSysvar":%q,
				"authorized":{"staker":%q,"withdrawer":%q},
				"lockup":{"unixTimestamp":-5,"epoch":20,"custodian":%q}}}`,
				k(1), k(2), k(0x61), k(0x62), k(0x63)),
		},
		{
			name:      "authorize Staker capitalized no custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(1), kb(0x64), le32(0)),
			accounts:  accs(1, 2, 3),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"authorize","info":{
				"stakeAccount":%q,"clockSysvar":%q,"authority":%q,
				"newAuthority":%q,"authorityType":"Staker"}}`,
				k(1), k(2), k(3), k(0x64)),
		},
		{
			name:      "authorize Withdrawer with custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(1), kb(0x64), le32(1)),
			accounts:  accs(1, 2, 3, 4),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"authorize","info":{
				"stakeAccount":%q,"clockSysvar":%q,"authority":%q,
				"newAuthority":%q,"authorityType":"Withdrawer","custodian":%q}}`,
				k(1), k(2), k(3), k(0x64), k(4)),
		},
		{
			name:      "delegate",
			programID: parse.StakeProgramID,
			data:      le32(2),
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"delegate","info":{
				"stakeAccount":%q,"voteAccount":%q,"clockSysvar":%q,
				"stakeHistorySysvar":%q,"stakeConfigAccount":%q,"stakeAuthority":%q}}`,
				k(1), k(2), k(3), k(4), k(5), k(6)),
		},
		{
			name:      "split",
			programID: parse.StakeProgramID,
			data:      cat(le32(3), le64(777)),
			accounts:  accs(1, 2, 3),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"split","info":{
				"stakeAccount":%q,"newSplitAccount":%q,"stakeAuthority":%q,"lamports":777}}`,
				k(1), k(2), k(3)),
		},
		{
			// bincode trailing-bytes tolerance, stake family.
			name:      "split trailing bytes tolerated",
			programID: parse.StakeProgramID,
			data:      cat(le32(3), le64(777), []byte{1, 2, 3, 4}),
			accounts:  accs(1, 2, 3),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"split","info":{
				"stakeAccount":%q,"newSplitAccount":%q,"stakeAuthority":%q,"lamports":777}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "withdraw no custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(4), le64(888)),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"withdraw","info":{
				"stakeAccount":%q,"destination":%q,"clockSysvar":%q,
				"stakeHistorySysvar":%q,"withdrawAuthority":%q,"lamports":888}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
		{
			name:      "withdraw with custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(4), le64(888)),
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"withdraw","info":{
				"stakeAccount":%q,"destination":%q,"clockSysvar":%q,
				"stakeHistorySysvar":%q,"withdrawAuthority":%q,"lamports":888,
				"custodian":%q}}`,
				k(1), k(2), k(3), k(4), k(5), k(6)),
		},
		{
			name:      "deactivate",
			programID: parse.StakeProgramID,
			data:      le32(5),
			accounts:  accs(1, 2, 3),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"deactivate","info":{
				"stakeAccount":%q,"clockSysvar":%q,"stakeAuthority":%q}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "setLockup all None renders empty lockup object",
			programID: parse.StakeProgramID,
			data:      cat(le32(6), optNone, optNone, optNone),
			accounts:  accs(1, 2),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"setLockup","info":{
				"stakeAccount":%q,"custodian":%q,"lockup":{}}}`, k(1), k(2)),
		},
		{
			name:      "setLockup only epoch omits the None keys",
			programID: parse.StakeProgramID,
			data:      cat(le32(6), optNone, optSomeU64(3), optNone),
			accounts:  accs(1, 2),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"setLockup","info":{
				"stakeAccount":%q,"custodian":%q,"lockup":{"epoch":3}}}`, k(1), k(2)),
		},
		{
			name:      "setLockup all Some",
			programID: parse.StakeProgramID,
			data:      cat(le32(6), optSomeI64(1234), optSomeU64(9), []byte{1}, kb(0x65)),
			accounts:  accs(1, 2),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"setLockup","info":{
				"stakeAccount":%q,"custodian":%q,
				"lockup":{"unixTimestamp":1234,"epoch":9,"custodian":%q}}}`,
				k(1), k(2), k(0x65)),
		},
		{
			name:      "merge destination then source",
			programID: parse.StakeProgramID,
			data:      le32(7),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"merge","info":{
				"destination":%q,"source":%q,"clockSysvar":%q,
				"stakeHistorySysvar":%q,"stakeAuthority":%q}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
		{
			name:      "authorizeWithSeed minimal has neither clockSysvar nor custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(8), kb(0x66), le32(1), bincodeStr("marinade"), kb(0x67)),
			accounts:  accs(1, 2),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"authorizeWithSeed","info":{
				"stakeAccount":%q,"authorityBase":%q,"newAuthorized":%q,
				"authorityType":"Withdrawer","authoritySeed":"marinade","authorityOwner":%q}}`,
				k(1), k(2), k(0x66), k(0x67)),
		},
		{
			name:      "authorizeWithSeed three accounts adds only clockSysvar",
			programID: parse.StakeProgramID,
			data:      cat(le32(8), kb(0x66), le32(0), bincodeStr("s"), kb(0x67)),
			accounts:  accs(1, 2, 3),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"authorizeWithSeed","info":{
				"stakeAccount":%q,"authorityBase":%q,"newAuthorized":%q,
				"authorityType":"Staker","authoritySeed":"s","authorityOwner":%q,
				"clockSysvar":%q}}`,
				k(1), k(2), k(0x66), k(0x67), k(3)),
		},
		{
			name:      "authorizeWithSeed four accounts adds custodian too",
			programID: parse.StakeProgramID,
			data:      cat(le32(8), kb(0x66), le32(0), bincodeStr("s"), kb(0x67)),
			accounts:  accs(1, 2, 3, 4),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"authorizeWithSeed","info":{
				"stakeAccount":%q,"authorityBase":%q,"newAuthorized":%q,
				"authorityType":"Staker","authoritySeed":"s","authorityOwner":%q,
				"clockSysvar":%q,"custodian":%q}}`,
				k(1), k(2), k(0x66), k(0x67), k(3), k(4)),
		},
		{
			name:      "initializeChecked",
			programID: parse.StakeProgramID,
			data:      le32(9),
			accounts:  accs(1, 2, 3, 4),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"initializeChecked","info":{
				"stakeAccount":%q,"rentSysvar":%q,"staker":%q,"withdrawer":%q}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "authorizeChecked no custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(10), le32(0)),
			accounts:  accs(1, 2, 3, 4),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"authorizeChecked","info":{
				"stakeAccount":%q,"clockSysvar":%q,"authority":%q,
				"newAuthority":%q,"authorityType":"Staker"}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "authorizeChecked with custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(10), le32(1)),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"authorizeChecked","info":{
				"stakeAccount":%q,"clockSysvar":%q,"authority":%q,
				"newAuthority":%q,"authorityType":"Withdrawer","custodian":%q}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
		{
			name:      "authorizeCheckedWithSeed no custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(11), le32(1), bincodeStr("x"), kb(0x68)),
			accounts:  accs(1, 2, 3, 4),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"authorizeCheckedWithSeed","info":{
				"stakeAccount":%q,"authorityBase":%q,"clockSysvar":%q,"newAuthorized":%q,
				"authorityType":"Withdrawer","authoritySeed":"x","authorityOwner":%q}}`,
				k(1), k(2), k(3), k(4), k(0x68)),
		},
		{
			name:      "authorizeCheckedWithSeed with custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(11), le32(0), bincodeStr("x"), kb(0x68)),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"authorizeCheckedWithSeed","info":{
				"stakeAccount":%q,"authorityBase":%q,"clockSysvar":%q,"newAuthorized":%q,
				"authorityType":"Staker","authoritySeed":"x","authorityOwner":%q,
				"custodian":%q}}`,
				k(1), k(2), k(3), k(4), k(0x68), k(5)),
		},
		{
			name:      "setLockupChecked all None two accounts",
			programID: parse.StakeProgramID,
			data:      cat(le32(12), optNone, optNone),
			accounts:  accs(1, 2),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"setLockupChecked","info":{
				"stakeAccount":%q,"custodian":%q,"lockup":{}}}`, k(1), k(2)),
		},
		{
			// The dual-custodian quirk: top-level custodian is the SIGNING
			// authority accounts[1]; lockup.custodian is the optional
			// accounts[2]. Distinct keys prove neither leaks into the other.
			name:      "setLockupChecked dual custodian",
			programID: parse.StakeProgramID,
			data:      cat(le32(12), optSomeI64(55), optSomeU64(66)),
			accounts:  accs(1, 2, 3),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"setLockupChecked","info":{
				"stakeAccount":%q,"custodian":%q,
				"lockup":{"unixTimestamp":55,"epoch":66,"custodian":%q}}}`,
				k(1), k(2), k(3)),
		},
		{
			// Null info: the "info" key must be entirely absent.
			name:      "getMinimumDelegation omits info",
			programID: parse.StakeProgramID,
			data:      le32(13),
			accounts:  accs(1),
			program:   "stake",
			parsed:    `{"type":"getMinimumDelegation"}`,
		},
		{
			name:      "deactivateDelinquent",
			programID: parse.StakeProgramID,
			data:      le32(14),
			accounts:  accs(1, 2, 3),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"deactivateDelinquent","info":{
				"stakeAccount":%q,"voteAccount":%q,"referenceVoteAccount":%q}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "redelegate still parses though deprecated",
			programID: parse.StakeProgramID,
			data:      le32(15),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"redelegate","info":{
				"stakeAccount":%q,"newStakeAccount":%q,"voteAccount":%q,
				"stakeConfigAccount":%q,"stakeAuthority":%q}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
		{
			name:      "moveStake source then destination",
			programID: parse.StakeProgramID,
			data:      cat(le32(16), le64(11)),
			accounts:  accs(1, 2, 3),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"moveStake","info":{
				"source":%q,"destination":%q,"stakeAuthority":%q,"lamports":11}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "moveLamports",
			programID: parse.StakeProgramID,
			data:      cat(le32(17), le64(22)),
			accounts:  accs(1, 2, 3),
			program:   "stake",
			parsed: fmt.Sprintf(`{"type":"moveLamports","info":{
				"source":%q,"destination":%q,"stakeAuthority":%q,"lamports":22}}`,
				k(1), k(2), k(3)),
		},
	})
}

func TestStakeNotParsable(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		accounts []string
	}{
		// Even the no-account arm refuses an EMPTY account list: Agave's
		// key-mismatch guard errors on empty accounts before any arm runs.
		{"getMinimumDelegation empty accounts", le32(13), nil},
		{"unknown discriminant 18", le32(18), accs(1, 2)},
		{"authorize variant 2", cat(le32(1), kb(1), le32(2)), accs(1, 2, 3)},
		{"setLockup Option tag 2", cat(le32(6), []byte{2}), accs(1, 2)},
		{"setLockupChecked Option tag 2", cat(le32(12), optNone, []byte{2}), accs(1, 2)},
		{"initialize truncated custodian",
			cat(le32(0), kb(1), kb(2), i64le(0), le64(0), kb(3)[:31]), accs(1, 2)},
		{"delegate 5 accounts", le32(2), accs(1, 2, 3, 4, 5)},
		{"authorizeWithSeed non-utf8 seed",
			cat(le32(8), kb(1), le32(0), le64(2), []byte{0xFF, 0xFE}, kb(2)), accs(1, 2)},
		{"split truncated lamports", cat(le32(3), le64(1)[:7]), accs(1, 2, 3)},
	}
	for _, tc := range cases {
		env, err := parse.Parse(parse.StakeProgramID, tc.data, tc.accounts, nil)
		if err == nil {
			t.Errorf("%s: Parse() succeeded, want ErrNotParsable; parsed: %s", tc.name, env.Parsed)
			continue
		}
		if !errors.Is(err, parse.ErrNotParsable) {
			t.Errorf("%s: error %v does not wrap ErrNotParsable", tc.name, err)
		}
	}
}
