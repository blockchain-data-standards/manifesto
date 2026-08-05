package parse_test

// Address-lookup-table and loader vectors: all 5 ProgramInstruction variants
// (solana-address-lookup-table-interface 2.2.2), both LoaderInstruction
// variants (solana-loader-v2-interface 2.2.1), and all 10
// UpgradeableLoaderInstruction variants (solana-loader-v3-interface 5.0.0).
// Expected JSON hand-written from parse_address_lookup_table.rs and
// parse_bpf_loader.rs.
//
// Quirks pinned here, all verified against the Rust source:
//   - Loader `bytes` render as BASE64_STANDARD — standard alphabet WITH
//     padding. The vectors use payload lengths that force '=' padding, so a
//     RawStdEncoding port would fail.
//   - extendLookupTable adds payerAccount AND systemProgram only when the
//     instruction carries >= 4 accounts: both keys or neither, never one;
//     newAddresses renders [] when empty.
//   - v2 finalize checks 2 accounts but renders only account[0].
//   - Two optionality renderings side by side in the v3 loader:
//     initializeBuffer's authority is OMITTED when absent, while
//     setAuthority.newAuthority, close.programAccount and the
//     extendProgram(Checked) systemProgram/payerAccount are PRESENT-and-null.
//   - extendProgramChecked's authority is mandatory at index 2 with
//     systemProgram/payerAccount shifted to 3/4 (unlike extendProgram).

import (
	"errors"
	"fmt"
	"testing"

	"github.com/blockchain-data-standards/manifesto/svm/parse"
)

func TestAddressLookupTableVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			name:      "createLookupTable",
			programID: parse.LookupTableID,
			data:      cat(le32(0), le64(12345), []byte{254}),
			accounts:  accs(1, 2, 3, 4),
			program:   "address-lookup-table",
			parsed: fmt.Sprintf(`{"type":"createLookupTable","info":{
				"lookupTableAccount":%q,"lookupTableAuthority":%q,
				"payerAccount":%q,"systemProgram":%q,"recentSlot":12345,"bumpSeed":254}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "freezeLookupTable",
			programID: parse.LookupTableID,
			data:      le32(1),
			accounts:  accs(1, 2),
			program:   "address-lookup-table",
			parsed: fmt.Sprintf(`{"type":"freezeLookupTable","info":{
				"lookupTableAccount":%q,"lookupTableAuthority":%q}}`, k(1), k(2)),
		},
		{
			// bincode trailing-bytes tolerance, ALT family.
			name:      "freezeLookupTable trailing bytes tolerated",
			programID: parse.LookupTableID,
			data:      cat(le32(1), []byte{9, 9, 9}),
			accounts:  accs(1, 2),
			program:   "address-lookup-table",
			parsed: fmt.Sprintf(`{"type":"freezeLookupTable","info":{
				"lookupTableAccount":%q,"lookupTableAuthority":%q}}`, k(1), k(2)),
		},
		{
			name:      "extendLookupTable with payer and systemProgram",
			programID: parse.LookupTableID,
			data:      cat(le32(2), le64(2), kb(0x71), kb(0x72)),
			accounts:  accs(1, 2, 3, 4),
			program:   "address-lookup-table",
			parsed: fmt.Sprintf(`{"type":"extendLookupTable","info":{
				"lookupTableAccount":%q,"lookupTableAuthority":%q,
				"newAddresses":[%q,%q],"payerAccount":%q,"systemProgram":%q}}`,
				k(1), k(2), k(0x71), k(0x72), k(3), k(4)),
		},
		{
			name:      "extendLookupTable empty addresses two accounts",
			programID: parse.LookupTableID,
			data:      cat(le32(2), le64(0)),
			accounts:  accs(1, 2),
			program:   "address-lookup-table",
			parsed: fmt.Sprintf(`{"type":"extendLookupTable","info":{
				"lookupTableAccount":%q,"lookupTableAuthority":%q,"newAddresses":[]}}`,
				k(1), k(2)),
		},
		{
			// Three accounts is NOT enough for the payer/system pair: both
			// keys stay absent (a single combined >= 4 condition).
			name:      "extendLookupTable three accounts adds neither payer nor system",
			programID: parse.LookupTableID,
			data:      cat(le32(2), le64(1), kb(0x73)),
			accounts:  accs(1, 2, 3),
			program:   "address-lookup-table",
			parsed: fmt.Sprintf(`{"type":"extendLookupTable","info":{
				"lookupTableAccount":%q,"lookupTableAuthority":%q,"newAddresses":[%q]}}`,
				k(1), k(2), k(0x73)),
		},
		{
			name:      "deactivateLookupTable",
			programID: parse.LookupTableID,
			data:      le32(3),
			accounts:  accs(1, 2),
			program:   "address-lookup-table",
			parsed: fmt.Sprintf(`{"type":"deactivateLookupTable","info":{
				"lookupTableAccount":%q,"lookupTableAuthority":%q}}`, k(1), k(2)),
		},
		{
			name:      "closeLookupTable",
			programID: parse.LookupTableID,
			data:      le32(4),
			accounts:  accs(1, 2, 3),
			program:   "address-lookup-table",
			parsed: fmt.Sprintf(`{"type":"closeLookupTable","info":{
				"lookupTableAccount":%q,"lookupTableAuthority":%q,"recipient":%q}}`,
				k(1), k(2), k(3)),
		},
	})
}

func TestBpfLoaderVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			// 5 payload bytes force base64 padding: "AQIDBAU=".
			name:      "v2 write base64 std padded",
			programID: parse.BpfLoaderID,
			data:      cat(le32(0), le32(4096), le64(5), []byte{1, 2, 3, 4, 5}),
			accounts:  accs(1),
			program:   "bpf-loader",
			parsed: fmt.Sprintf(`{"type":"write","info":{
				"offset":4096,"bytes":"AQIDBAU=","account":%q}}`, k(1)),
		},
		{
			// Requires 2 accounts but renders only account[0].
			name:      "v2 finalize renders only the program account",
			programID: parse.BpfLoaderID,
			data:      le32(1),
			accounts:  accs(1, 2),
			program:   "bpf-loader",
			parsed:    fmt.Sprintf(`{"type":"finalize","info":{"account":%q}}`, k(1)),
		},
	})
}

func TestBpfUpgradeableLoaderVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			name:      "initializeBuffer omits authority when absent",
			programID: parse.BpfUpgradeableLoaderID,
			data:      le32(0),
			accounts:  accs(1),
			program:   "bpf-upgradeable-loader",
			parsed:    fmt.Sprintf(`{"type":"initializeBuffer","info":{"account":%q}}`, k(1)),
		},
		{
			name:      "initializeBuffer with authority",
			programID: parse.BpfUpgradeableLoaderID,
			data:      le32(0),
			accounts:  accs(1, 2),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"initializeBuffer","info":{
				"account":%q,"authority":%q}}`, k(1), k(2)),
		},
		{
			// 4 payload bytes force padding: "3q2+7w==".
			name:      "v3 write base64 std padded",
			programID: parse.BpfUpgradeableLoaderID,
			data:      cat(le32(1), le32(8), le64(4), []byte{0xDE, 0xAD, 0xBE, 0xEF}),
			accounts:  accs(1, 2),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"write","info":{
				"offset":8,"bytes":"3q2+7w==","account":%q,"authority":%q}}`, k(1), k(2)),
		},
		{
			name:      "deployWithMaxDataLen",
			programID: parse.BpfUpgradeableLoaderID,
			data:      cat(le32(2), le64(999999)),
			accounts:  accs(1, 2, 3, 4, 5, 6, 7, 8),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"deployWithMaxDataLen","info":{
				"maxDataLen":999999,"payerAccount":%q,"programDataAccount":%q,
				"programAccount":%q,"bufferAccount":%q,"rentSysvar":%q,
				"clockSysvar":%q,"systemProgram":%q,"authority":%q}}`,
				k(1), k(2), k(3), k(4), k(5), k(6), k(7), k(8)),
		},
		{
			name:      "upgrade",
			programID: parse.BpfUpgradeableLoaderID,
			data:      le32(3),
			accounts:  accs(1, 2, 3, 4, 5, 6, 7),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"upgrade","info":{
				"programDataAccount":%q,"programAccount":%q,"bufferAccount":%q,
				"spillAccount":%q,"rentSysvar":%q,"clockSysvar":%q,"authority":%q}}`,
				k(1), k(2), k(3), k(4), k(5), k(6), k(7)),
		},
		{
			name:      "setAuthority newAuthority present-and-null",
			programID: parse.BpfUpgradeableLoaderID,
			data:      le32(4),
			accounts:  accs(1, 2),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"setAuthority","info":{
				"account":%q,"authority":%q,"newAuthority":null}}`, k(1), k(2)),
		},
		{
			name:      "setAuthority with newAuthority",
			programID: parse.BpfUpgradeableLoaderID,
			data:      le32(4),
			accounts:  accs(1, 2, 3),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"setAuthority","info":{
				"account":%q,"authority":%q,"newAuthority":%q}}`, k(1), k(2), k(3)),
		},
		{
			// bincode trailing-bytes tolerance, loader family.
			name:      "setAuthority trailing bytes tolerated",
			programID: parse.BpfUpgradeableLoaderID,
			data:      cat(le32(4), []byte{0xEE}),
			accounts:  accs(1, 2, 3),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"setAuthority","info":{
				"account":%q,"authority":%q,"newAuthority":%q}}`, k(1), k(2), k(3)),
		},
		{
			name:      "close programAccount present-and-null",
			programID: parse.BpfUpgradeableLoaderID,
			data:      le32(5),
			accounts:  accs(1, 2, 3),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"close","info":{
				"account":%q,"recipient":%q,"authority":%q,"programAccount":null}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "close with programAccount",
			programID: parse.BpfUpgradeableLoaderID,
			data:      le32(5),
			accounts:  accs(1, 2, 3, 4),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"close","info":{
				"account":%q,"recipient":%q,"authority":%q,"programAccount":%q}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "extendProgram both optionals null",
			programID: parse.BpfUpgradeableLoaderID,
			data:      cat(le32(6), le32(1024)),
			accounts:  accs(1, 2),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"extendProgram","info":{
				"additionalBytes":1024,"programDataAccount":%q,"programAccount":%q,
				"systemProgram":null,"payerAccount":null}}`, k(1), k(2)),
		},
		{
			name:      "extendProgram three accounts fills only systemProgram",
			programID: parse.BpfUpgradeableLoaderID,
			data:      cat(le32(6), le32(1024)),
			accounts:  accs(1, 2, 3),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"extendProgram","info":{
				"additionalBytes":1024,"programDataAccount":%q,"programAccount":%q,
				"systemProgram":%q,"payerAccount":null}}`, k(1), k(2), k(3)),
		},
		{
			name:      "extendProgram four accounts fills both",
			programID: parse.BpfUpgradeableLoaderID,
			data:      cat(le32(6), le32(1024)),
			accounts:  accs(1, 2, 3, 4),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"extendProgram","info":{
				"additionalBytes":1024,"programDataAccount":%q,"programAccount":%q,
				"systemProgram":%q,"payerAccount":%q}}`, k(1), k(2), k(3), k(4)),
		},
		{
			name:      "setAuthorityChecked",
			programID: parse.BpfUpgradeableLoaderID,
			data:      le32(7),
			accounts:  accs(1, 2, 3),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"setAuthorityChecked","info":{
				"account":%q,"authority":%q,"newAuthority":%q}}`, k(1), k(2), k(3)),
		},
		{
			name:      "migrate",
			programID: parse.BpfUpgradeableLoaderID,
			data:      le32(8),
			accounts:  accs(1, 2, 3),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"migrate","info":{
				"programDataAccount":%q,"programAccount":%q,"authority":%q}}`,
				k(1), k(2), k(3)),
		},
		{
			// authority is mandatory at index 2; system/payer shift to 3/4.
			name:      "extendProgramChecked minimal has null system and payer",
			programID: parse.BpfUpgradeableLoaderID,
			data:      cat(le32(9), le32(2048)),
			accounts:  accs(1, 2, 3),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"extendProgramChecked","info":{
				"additionalBytes":2048,"programDataAccount":%q,"programAccount":%q,
				"authority":%q,"systemProgram":null,"payerAccount":null}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "extendProgramChecked five accounts fills both",
			programID: parse.BpfUpgradeableLoaderID,
			data:      cat(le32(9), le32(2048)),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "bpf-upgradeable-loader",
			parsed: fmt.Sprintf(`{"type":"extendProgramChecked","info":{
				"additionalBytes":2048,"programDataAccount":%q,"programAccount":%q,
				"authority":%q,"systemProgram":%q,"payerAccount":%q}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
	})
}

func TestAltAndLoadersNotParsable(t *testing.T) {
	cases := []struct {
		name      string
		programID string
		data      []byte
		accounts  []string
	}{
		{"alt unknown discriminant 5", parse.LookupTableID, le32(5), accs(1, 2)},
		{"alt create missing bump", parse.LookupTableID, cat(le32(0), le64(1)), accs(1, 2, 3, 4)},
		{"alt create 3 accounts", parse.LookupTableID, cat(le32(0), le64(1), []byte{0}), accs(1, 2, 3)},
		{"alt extend vec length exceeds data", parse.LookupTableID, cat(le32(2), le64(1<<40)), accs(1, 2)},
		{"alt extend truncated pubkey", parse.LookupTableID, cat(le32(2), le64(1), kb(1)[:31]), accs(1, 2)},
		{"alt close 2 accounts", parse.LookupTableID, le32(4), accs(1, 2)},
		{"alt empty accounts", parse.LookupTableID, le32(1), nil},

		{"v2 unknown discriminant 2", parse.BpfLoaderID, le32(2), accs(1)},
		{"v2 write declared 5 bytes got 3", parse.BpfLoaderID,
			cat(le32(0), le32(0), le64(5), []byte{1, 2, 3}), accs(1)},
		{"v2 finalize 1 account", parse.BpfLoaderID, le32(1), accs(1)},

		{"v3 unknown discriminant 10", parse.BpfUpgradeableLoaderID, le32(10), accs(1, 2, 3)},
		{"v3 write truncated offset", parse.BpfUpgradeableLoaderID, cat(le32(1), []byte{1, 2}), accs(1, 2)},
		{"v3 write declared 4 bytes got 2", parse.BpfUpgradeableLoaderID,
			cat(le32(1), le32(0), le64(4), []byte{1, 2}), accs(1, 2)},
		{"v3 deploy 7 accounts", parse.BpfUpgradeableLoaderID,
			cat(le32(2), le64(1)), accs(1, 2, 3, 4, 5, 6, 7)},
		{"v3 extendProgramChecked 2 accounts", parse.BpfUpgradeableLoaderID,
			cat(le32(9), le32(1)), accs(1, 2)},
		{"v3 extendProgram truncated u32", parse.BpfUpgradeableLoaderID,
			cat(le32(6), []byte{1}), accs(1, 2)},
	}
	for _, tc := range cases {
		env, err := parse.Parse(tc.programID, tc.data, tc.accounts, nil)
		if err == nil {
			t.Errorf("%s: Parse() succeeded, want ErrNotParsable; parsed: %s", tc.name, env.Parsed)
			continue
		}
		if !errors.Is(err, parse.ErrNotParsable) {
			t.Errorf("%s: error %v does not wrap ErrNotParsable", tc.name, err)
		}
	}
}
