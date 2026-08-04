package parse_test

// Vote program vectors: every VoteInstruction variant (discriminants 0..=15),
// wire layouts hand-encoded from solana-vote-interface 2.2.6, expected JSON
// hand-written from parse_vote.rs (solana-transaction-status 2.3.13).
//
// Render quirks pinned here, all verified against the Rust source:
//   - authorityType is the RAW serde unit enum: "Voter" / "Withdrawer"
//     (capitalized; VoteAuthorize has no rename attribute).
//   - Lockout renders SNAKE_CASE keys: "slot" and "confirmation_count"
//     (state/mod.rs — no serde rename on the struct).
//   - Six all-lowercase type strings: updatevotestate, updatevotestateswitch,
//     compactupdatevotestate, compactupdatevotestateswitch, towersync,
//     towersyncswitch.
//   - initialize renders "node" from accounts[3]; the decoded
//     VoteInit.node_pubkey is parsed but NEVER rendered.
//   - Compact forms: root is a raw u64 where u64::MAX means null; lockouts
//     are {varint offset, u8 cc} deltas accumulated from the root (or 0 when
//     the root is None); TowerSync carries a trailing blockId hash.
//   - root/timestamp keys are always present, null when None.

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"github.com/blockchain-data-standards/manifesto/svm/parse"

	"fmt"
)

// shortU16 is solana-short-vec's canonical compact-u16 encoding.
func shortU16(v uint16) []byte {
	var out []byte
	for {
		b := byte(v & 0x7F)
		v >>= 7
		if v == 0 {
			return append(out, b)
		}
		out = append(out, b|0x80)
	}
}

// varint64 is solana-serde-varint's canonical u64 encoding.
func varint64(v uint64) []byte {
	var out []byte
	for {
		b := byte(v & 0x7F)
		v >>= 7
		if v == 0 {
			return append(out, b)
		}
		out = append(out, b|0x80)
	}
}

// i64le encodes a bincode i64 (two's complement LE).
func i64le(v int64) []byte { return le64(uint64(v)) }

// optSomeI64/optSomeU64/optNone are bincode Option<T> encodings.
func optSomeI64(v int64) []byte  { return cat([]byte{1}, i64le(v)) }
func optSomeU64(v uint64) []byte { return cat([]byte{1}, le64(v)) }

var optNone = []byte{0}

func TestVoteVectors(t *testing.T) {
	le64max := le64(math.MaxUint64)
	runVectors(t, []vector{
		{
			name:      "initialize renders node from accounts not from VoteInit",
			programID: parse.VoteProgramID,
			data:      cat(le32(0), kb(0xAA), kb(0xBB), kb(0xCC), []byte{7}),
			accounts:  accs(1, 2, 3, 4),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"initialize","info":{
				"voteAccount":%q,"rentSysvar":%q,"clockSysvar":%q,"node":%q,
				"authorizedVoter":%q,"authorizedWithdrawer":%q,"commission":7}}`,
				k(1), k(2), k(3), k(4), k(0xBB), k(0xCC)),
		},
		{
			name:      "authorize Voter capitalized",
			programID: parse.VoteProgramID,
			data:      cat(le32(1), kb(0xDD), le32(0)),
			accounts:  accs(1, 2, 3),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"authorize","info":{
				"voteAccount":%q,"clockSysvar":%q,"authority":%q,
				"newAuthority":%q,"authorityType":"Voter"}}`,
				k(1), k(2), k(3), k(0xDD)),
		},
		{
			name:      "authorize Withdrawer capitalized",
			programID: parse.VoteProgramID,
			data:      cat(le32(1), kb(0xDD), le32(1)),
			accounts:  accs(1, 2, 3),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"authorize","info":{
				"voteAccount":%q,"clockSysvar":%q,"authority":%q,
				"newAuthority":%q,"authorityType":"Withdrawer"}}`,
				k(1), k(2), k(3), k(0xDD)),
		},
		{
			name:      "vote with slots and timestamp",
			programID: parse.VoteProgramID,
			data:      cat(le32(2), le64(2), le64(100), le64(101), kb(0x42), optSomeI64(1717)),
			accounts:  accs(1, 2, 3, 4),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"vote","info":{
				"voteAccount":%q,"slotHashesSysvar":%q,"clockSysvar":%q,"voteAuthority":%q,
				"vote":{"slots":[100,101],"hash":%q,"timestamp":1717}}}`,
				k(1), k(2), k(3), k(4), k(0x42)),
		},
		{
			name:      "vote empty slots null timestamp",
			programID: parse.VoteProgramID,
			data:      cat(le32(2), le64(0), kb(0x42), optNone),
			accounts:  accs(1, 2, 3, 4),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"vote","info":{
				"voteAccount":%q,"slotHashesSysvar":%q,"clockSysvar":%q,"voteAuthority":%q,
				"vote":{"slots":[],"hash":%q,"timestamp":null}}}`,
				k(1), k(2), k(3), k(4), k(0x42)),
		},
		{
			name:      "withdraw",
			programID: parse.VoteProgramID,
			data:      cat(le32(3), le64(5000)),
			accounts:  accs(1, 2, 3),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"withdraw","info":{
				"voteAccount":%q,"destination":%q,"withdrawAuthority":%q,"lamports":5000}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "updateValidatorIdentity",
			programID: parse.VoteProgramID,
			data:      le32(4),
			accounts:  accs(1, 2, 3),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"updateValidatorIdentity","info":{
				"voteAccount":%q,"newValidatorIdentity":%q,"withdrawAuthority":%q}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "updateCommission",
			programID: parse.VoteProgramID,
			data:      cat(le32(5), []byte{99}),
			accounts:  accs(1, 2),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"updateCommission","info":{
				"voteAccount":%q,"withdrawAuthority":%q,"commission":99}}`, k(1), k(2)),
		},
		{
			// bincode trailing-bytes tolerance, vote family.
			name:      "updateCommission trailing bytes tolerated",
			programID: parse.VoteProgramID,
			data:      cat(le32(5), []byte{99}, []byte{0xDE, 0xAD, 0xBE}),
			accounts:  accs(1, 2),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"updateCommission","info":{
				"voteAccount":%q,"withdrawAuthority":%q,"commission":99}}`, k(1), k(2)),
		},
		{
			name:      "voteSwitch",
			programID: parse.VoteProgramID,
			data:      cat(le32(6), le64(1), le64(7), kb(0x42), optNone, kb(0x43)),
			accounts:  accs(1, 2, 3, 4),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"voteSwitch","info":{
				"voteAccount":%q,"slotHashesSysvar":%q,"clockSysvar":%q,"voteAuthority":%q,
				"vote":{"slots":[7],"hash":%q,"timestamp":null},"hash":%q}}`,
				k(1), k(2), k(3), k(4), k(0x42), k(0x43)),
		},
		{
			name:      "authorizeChecked newAuthority from accounts",
			programID: parse.VoteProgramID,
			data:      cat(le32(7), le32(1)),
			accounts:  accs(1, 2, 3, 4),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"authorizeChecked","info":{
				"voteAccount":%q,"clockSysvar":%q,"authority":%q,
				"newAuthority":%q,"authorityType":"Withdrawer"}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "updatevotestate lowercase type snake confirmation_count",
			programID: parse.VoteProgramID,
			data: cat(le32(8), le64(2), le64(10), le32(31), le64(11), le32(30),
				optSomeU64(9), kb(0x44), optSomeI64(1234)),
			accounts: accs(1, 2),
			program:  "vote",
			parsed: fmt.Sprintf(`{"type":"updatevotestate","info":{
				"voteAccount":%q,"voteAuthority":%q,
				"voteStateUpdate":{
					"lockouts":[{"slot":10,"confirmation_count":31},{"slot":11,"confirmation_count":30}],
					"root":9,"hash":%q,"timestamp":1234}}}`,
				k(1), k(2), k(0x44)),
		},
		{
			name:      "updatevotestateswitch empty lockouts null root and timestamp",
			programID: parse.VoteProgramID,
			data:      cat(le32(9), le64(0), optNone, kb(0x45), optNone, kb(0x46)),
			accounts:  accs(1, 2),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"updatevotestateswitch","info":{
				"voteAccount":%q,"voteAuthority":%q,
				"voteStateUpdate":{"lockouts":[],"root":null,"hash":%q,"timestamp":null},
				"hash":%q}}`,
				k(1), k(2), k(0x45), k(0x46)),
		},
		{
			name:      "authorizeWithSeed newAuthority from args",
			programID: parse.VoteProgramID,
			data:      cat(le32(10), le32(1), kb(0x51), bincodeStr("seed🌱"), kb(0x52)),
			accounts:  accs(1, 2, 3),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"authorizeWithSeed","info":{
				"voteAccount":%q,"clockSysvar":%q,"authorityBaseKey":%q,
				"authorityOwner":%q,"authoritySeed":"seed🌱","newAuthority":%q,
				"authorityType":"Withdrawer"}}`,
				k(1), k(2), k(3), k(0x51), k(0x52)),
		},
		{
			name:      "authorizeCheckedWithSeed newAuthority from accounts",
			programID: parse.VoteProgramID,
			data:      cat(le32(11), le32(0), kb(0x53), bincodeStr("s")),
			accounts:  accs(1, 2, 3, 4),
			program:   "vote",
			parsed: fmt.Sprintf(`{"type":"authorizeCheckedWithSeed","info":{
				"voteAccount":%q,"clockSysvar":%q,"authorityBaseKey":%q,
				"authorityOwner":%q,"authoritySeed":"s","newAuthority":%q,
				"authorityType":"Voter"}}`,
				k(1), k(2), k(3), k(0x53), k(4)),
		},
		{
			// Accumulator starts at the root; a two-byte varint offset (300)
			// pins the multi-group decoding.
			name:      "compactupdatevotestate accumulates lockouts from root",
			programID: parse.VoteProgramID,
			data: cat(le32(12), le64(1000), shortU16(2),
				varint64(300), []byte{5}, varint64(1), []byte{4},
				kb(0x47), optSomeI64(999)),
			accounts: accs(1, 2),
			program:  "vote",
			parsed: fmt.Sprintf(`{"type":"compactupdatevotestate","info":{
				"voteAccount":%q,"voteAuthority":%q,
				"voteStateUpdate":{
					"lockouts":[{"slot":1300,"confirmation_count":5},{"slot":1301,"confirmation_count":4}],
					"root":1000,"hash":%q,"timestamp":999}}}`,
				k(1), k(2), k(0x47)),
		},
		{
			// u64::MAX root means null AND the accumulator starts at zero.
			name:      "compactupdatevotestateswitch root MAX is null accumulator from zero",
			programID: parse.VoteProgramID,
			data: cat(le32(13), le64max, shortU16(1), varint64(7), []byte{2},
				kb(0x48), optNone, kb(0x49)),
			accounts: accs(1, 2),
			program:  "vote",
			parsed: fmt.Sprintf(`{"type":"compactupdatevotestateswitch","info":{
				"voteAccount":%q,"voteAuthority":%q,
				"voteStateUpdate":{
					"lockouts":[{"slot":7,"confirmation_count":2}],
					"root":null,"hash":%q,"timestamp":null},
				"hash":%q}}`,
				k(1), k(2), k(0x48), k(0x49)),
		},
		{
			name:      "towersync carries blockId",
			programID: parse.VoteProgramID,
			data: cat(le32(14), le64(50), shortU16(1), varint64(10), []byte{3},
				kb(0x4A), optSomeI64(777), kb(0x4B)),
			accounts: accs(1, 2),
			program:  "vote",
			parsed: fmt.Sprintf(`{"type":"towersync","info":{
				"voteAccount":%q,"voteAuthority":%q,
				"towerSync":{
					"lockouts":[{"slot":60,"confirmation_count":3}],
					"root":50,"hash":%q,"timestamp":777,"blockId":%q}}}`,
				k(1), k(2), k(0x4A), k(0x4B)),
		},
		{
			name:      "towersyncswitch empty tower",
			programID: parse.VoteProgramID,
			data: cat(le32(15), le64max, shortU16(0),
				kb(0x4C), optNone, kb(0x4D), kb(0x4E)),
			accounts: accs(1, 2),
			program:  "vote",
			parsed: fmt.Sprintf(`{"type":"towersyncswitch","info":{
				"voteAccount":%q,"voteAuthority":%q,
				"towerSync":{"lockouts":[],"root":null,"hash":%q,"timestamp":null,"blockId":%q},
				"hash":%q}}`,
				k(1), k(2), k(0x4C), k(0x4D), k(0x4E)),
		},
	})
}

// TestVoteNotParsable: decode strictness for the vote wire — the compact
// short_vec/varint refusal semantics come straight from solana-short-vec and
// solana-serde-varint (aliases, continuation abuse, truncation).
func TestVoteNotParsable(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		accounts []string
	}{
		{"empty accounts refused before any arm", le32(4), nil},
		{"unknown discriminant 16", le32(16), accs(1, 2)},
		{"truncated discriminant", []byte{2, 0}, accs(1, 2)},
		{"authorize variant 2", cat(le32(1), kb(1), le32(2)), accs(1, 2, 3)},
		{"vote Option timestamp tag 2", cat(le32(2), le64(0), kb(1), []byte{2}), accs(1, 2, 3, 4)},
		{"initialize 3 accounts", cat(le32(0), kb(1), kb(2), kb(3), []byte{0}), accs(1, 2, 3)},
		{"vote truncated hash", cat(le32(2), le64(1), le64(5), kb(1)[:31]), accs(1, 2, 3, 4)},
		{"vote slot vec length exceeds data", cat(le32(2), le64(1<<40)), accs(1, 2, 3, 4)},
		{"authorizeWithSeed non-utf8 seed",
			cat(le32(10), le32(0), kb(1), le64(2), []byte{0xFF, 0xFE}, kb(2)), accs(1, 2, 3)},

		// short_vec strictness (solana-short-vec): aliases and continuation.
		{"short_vec alias 0x80 0x00", cat(le32(12), le64(0), []byte{0x80, 0x00}), accs(1, 2)},
		{"short_vec third byte continues", cat(le32(12), le64(0), []byte{0xFF, 0xFF, 0x80}), accs(1, 2)},
		{"short_vec overflow past u16", cat(le32(12), le64(0), []byte{0xFF, 0xFF, 0x04}), accs(1, 2)},

		// varint strictness (solana-serde-varint): trailing zero, truncation.
		{"varint trailing zero", cat(le32(12), le64(0), shortU16(1), []byte{0x80, 0x00}), accs(1, 2)},
		{"varint truncated last byte",
			cat(le32(12), le64(0), shortU16(1),
				append(bytes.Repeat([]byte{0x80}, 9), 0x02)), accs(1, 2)},

		// Compact-form structural refusals.
		{"lockout count exceeds remaining bytes",
			cat(le32(12), le64(0), shortU16(200), []byte{1, 1}), accs(1, 2)},
		{"lockout offset overflows accumulator",
			cat(le32(12), le64(math.MaxUint64-1), shortU16(1), varint64(3), []byte{1}), accs(1, 2)},
		{"towersync truncated blockId",
			cat(le32(14), le64(0), shortU16(0), kb(1), optNone, kb(2)[:31]), accs(1, 2)},
		{"towersync 1 account",
			cat(le32(14), le64(0), shortU16(0), kb(1), optNone, kb(2)), accs(1)},
	}
	for _, tc := range cases {
		env, err := parse.Parse(parse.VoteProgramID, tc.data, tc.accounts, nil)
		if err == nil {
			t.Errorf("%s: Parse() succeeded, want ErrNotParsable; parsed: %s", tc.name, env.Parsed)
			continue
		}
		if !errors.Is(err, parse.ErrNotParsable) {
			t.Errorf("%s: error %v does not wrap ErrNotParsable", tc.name, err)
		}
	}
}
