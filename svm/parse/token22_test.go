package parse_test

// Token-2022 extension vectors: every direct arm (discriminants 25, 29, 31,
// 32, 35, 38) and every sub-instruction of every extension family (26, 27,
// 28, 30, 33, 34, 36, 37, 39, 40, 41, 42, 43, 44). Wire layouts hand-encoded
// from spl-token-2022 8.0.1, expected JSON hand-written from Agave's
// parse_token.rs extension arms. Base64 literals are hand-computed, so the
// expected side never routes through the implementation's encoder.
//
// Quirks pinned here, all verified against the Rust source:
//   - setTransferFee's multisig field is Agave's literal typo
//     "multisigtransferFeeConfigAuthority" (lowercase t).
//   - initializeMintCloseAuthority renders newAuthority PRESENT-and-null,
//     unlike initializeMint's freezeAuthority which is omitted.
//   - reallocate ALWAYS renders extensionTypes ([] when empty) while
//     getAccountDataSize omits the key when the list is empty.
//   - Two "optional-looking" wire fields are plain pods and therefore ALWAYS
//     render: pausable Initialize authority (base58, even all-zero) and CTF
//     InitializeConfig withdrawWithheldAuthorityElGamalPubkey (base64, even
//     all-zero).
//   - CT Deposit/Withdraw amounts are JSON NUMBERS, unlike the base-set
//     string amounts.
//   - CTF WithdrawFromAccounts labels its optional proof account
//     "proofAccount" (vs "recordAccount" in the FromMint arm) and gates it
//     on firstSourceIndex > 4; its pod decodes BEFORE the account count.
//   - TransferWithFee: wire order (eq, transferCV, feeSigma, feeCV, range)
//     differs from both the JSON key order and the proof-account labeling
//     order (eq, transferCV, feeCV, feeSigma, range).
//   - scaledUiAmount multipliers stringify through Rust's f64 Display
//     (shortest form: "1.5", "0.1", "2").
//   - Pod payloads refuse trailing bytes (decode_instruction_data wants the
//     exact length); the transferFee family's manual sub-unpack tolerates
//     them, as does the shared base set.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/blockchain-data-standards/manifesto/svm/parse"
)

func le16(v uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	return b
}

func i16le(v int16) []byte { return le16(uint16(v)) }

func f64le(v float64) []byte { return le64(math.Float64bits(v)) }

// rep is a run of n identical bytes — pod balances/ciphertexts/pubkeys.
func rep(n int, b byte) []byte { return bytes.Repeat([]byte{b}, n) }

// Hand-computed base64 literals for the pod payloads used below.
const (
	b64x86x32  = "hoaGhoaGhoaGhoaGhoaGhoaGhoaGhoaGhoaGhoaGhoY=" // 32 × 0x86
	b64x87x32  = "h4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4c=" // 32 × 0x87
	b64x88x36  = "iIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiI"
	b64x89x36  = "iYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJ"
	b64x8Ax36  = "ioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqK"
	b64x8Dx36  = "jY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2N"
	b64x8Fx32  = "j4+Pj4+Pj4+Pj4+Pj4+Pj4+Pj4+Pj4+Pj4+Pj4+Pj48=" // 32 × 0x8F
	b64x91x36  = "kZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGR"
	b64x92x32  = "kpKSkpKSkpKSkpKSkpKSkpKSkpKSkpKSkpKSkpKSkpI=" // 32 × 0x92
	b64x93x36  = "k5OTk5OTk5OTk5OTk5OTk5OTk5OTk5OTk5OTk5OTk5OTk5OT"
	b64zeros32 = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" // 32 × 0x00
	b58zeros32 = "11111111111111111111111111111111"             // base58 of 32 zeros
)

func TestToken22DirectVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			name:      "initializeMintCloseAuthority COption Some",
			programID: parse.Token2022ID,
			data:      cat([]byte{25, 1}, kb(0x81)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeMintCloseAuthority","info":{
				"mint":%q,"newAuthority":%q}}`, k(1), k(0x81)),
		},
		{
			name:      "initializeMintCloseAuthority COption None is present-null",
			programID: parse.Token2022ID,
			data:      []byte{25, 0},
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeMintCloseAuthority","info":{
				"mint":%q,"newAuthority":null}}`, k(1)),
		},
		{
			name:      "reallocate with extensions single owner",
			programID: parse.Token2022ID,
			data:      cat([]byte{29}, le16(1), le16(8)),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"reallocate","info":{
				"account":%q,"payer":%q,"systemProgram":%q,
				"extensionTypes":["transferFeeConfig","memoTransfer"],"owner":%q}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "reallocate empty extensionTypes stays present multisig owner",
			programID: parse.Token2022ID,
			data:      []byte{29},
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"reallocate","info":{
				"account":%q,"payer":%q,"systemProgram":%q,"extensionTypes":[],
				"multisigOwner":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4), k(5), k(6)),
		},
		{
			// The counterpart quirk: getAccountDataSize OMITS the key when
			// the extension list is empty...
			name:      "getAccountDataSize empty list omits extensionTypes",
			programID: parse.Token2022ID,
			data:      []byte{21},
			accounts:  accs(1),
			program:   "spl-token",
			parsed:    fmt.Sprintf(`{"type":"getAccountDataSize","info":{"mint":%q}}`, k(1)),
		},
		{
			// ...and renders it when non-empty.
			name:      "getAccountDataSize with extensions",
			programID: parse.Token2022ID,
			data:      cat([]byte{21}, le16(26), le16(27)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"getAccountDataSize","info":{
				"mint":%q,"extensionTypes":["pausable","pausableAccount"]}}`, k(1)),
		},
		{
			name:      "createNativeMint",
			programID: parse.Token2022ID,
			data:      []byte{31},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"createNativeMint","info":{
				"payer":%q,"nativeMint":%q,"systemProgram":%q}}`, k(1), k(2), k(3)),
		},
		{
			name:      "initializeNonTransferableMint",
			programID: parse.Token2022ID,
			data:      []byte{32},
			accounts:  accs(1),
			program:   "spl-token",
			parsed:    fmt.Sprintf(`{"type":"initializeNonTransferableMint","info":{"mint":%q}}`, k(1)),
		},
		{
			name:      "initializePermanentDelegate",
			programID: parse.Token2022ID,
			data:      cat([]byte{35}, kb(0x82)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializePermanentDelegate","info":{
				"mint":%q,"delegate":%q}}`, k(1), k(0x82)),
		},
		{
			name:      "withdrawExcessLamports single authority",
			programID: parse.Token2022ID,
			data:      []byte{38},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawExcessLamports","info":{
				"source":%q,"destination":%q,"authority":%q}}`, k(1), k(2), k(3)),
		},
		{
			name:      "withdrawExcessLamports multisig",
			programID: parse.Token2022ID,
			data:      []byte{38},
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawExcessLamports","info":{
				"source":%q,"destination":%q,"multisigAuthority":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
		{
			// Base-set trailing tolerance under the 2022 program id.
			name:      "base transfer trailing bytes tolerated",
			programID: parse.Token2022ID,
			data:      cat([]byte{3}, le64(42), []byte{0xAB, 0xCD}),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"transfer","info":{
				"source":%q,"destination":%q,"amount":"42","authority":%q}}`,
				k(1), k(2), k(3)),
		},
	})
}

func TestToken22TransferFeeVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			name:      "initializeTransferFeeConfig both authorities Some",
			programID: parse.Token2022ID,
			data: cat([]byte{26, 0}, []byte{1}, kb(0x83), []byte{1}, kb(0x84),
				le16(250), le64(1000000)),
			accounts: accs(1),
			program:  "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeTransferFeeConfig","info":{
				"mint":%q,"transferFeeConfigAuthority":%q,"withdrawWithheldAuthority":%q,
				"transferFeeBasisPoints":250,"maximumFee":1000000}}`,
				k(1), k(0x83), k(0x84)),
		},
		{
			name:      "initializeTransferFeeConfig both authorities None omit keys",
			programID: parse.Token2022ID,
			data:      cat([]byte{26, 0}, []byte{0, 0}, le16(0), le64(0)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeTransferFeeConfig","info":{
				"mint":%q,"transferFeeBasisPoints":0,"maximumFee":0}}`, k(1)),
		},
		{
			name:      "transferCheckedWithFee single authority",
			programID: parse.Token2022ID,
			data:      cat([]byte{26, 1}, le64(150000), []byte{2}, le64(1500)),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"transferCheckedWithFee","info":{
				"source":%q,"mint":%q,"destination":%q,
				"tokenAmount":{"uiAmount":1500,"decimals":2,"amount":"150000","uiAmountString":"1500"},
				"feeAmount":{"uiAmount":15,"decimals":2,"amount":"1500","uiAmountString":"15"},
				"authority":%q}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			name:      "transferCheckedWithFee multisig",
			programID: parse.Token2022ID,
			data:      cat([]byte{26, 1}, le64(150000), []byte{2}, le64(1500)),
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"transferCheckedWithFee","info":{
				"source":%q,"mint":%q,"destination":%q,
				"tokenAmount":{"uiAmount":1500,"decimals":2,"amount":"150000","uiAmountString":"1500"},
				"feeAmount":{"uiAmount":15,"decimals":2,"amount":"1500","uiAmountString":"15"},
				"multisigAuthority":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4), k(5), k(6)),
		},
		{
			name:      "withdrawWithheldTokensFromMint single",
			programID: parse.Token2022ID,
			data:      []byte{26, 2},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawWithheldTokensFromMint","info":{
				"mint":%q,"feeRecipient":%q,"withdrawWithheldAuthority":%q}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "withdrawWithheldTokensFromMint multisig",
			programID: parse.Token2022ID,
			data:      []byte{26, 2},
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawWithheldTokensFromMint","info":{
				"mint":%q,"feeRecipient":%q,
				"multisigWithdrawWithheldAuthority":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4), k(5)),
		},
		{
			name:      "withdrawWithheldTokensFromAccounts single",
			programID: parse.Token2022ID,
			data:      []byte{26, 3, 2},
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawWithheldTokensFromAccounts","info":{
				"mint":%q,"feeRecipient":%q,"sourceAccounts":[%q,%q],
				"withdrawWithheldAuthority":%q}}`,
				k(1), k(2), k(4), k(5), k(3)),
		},
		{
			name:      "withdrawWithheldTokensFromAccounts multisig",
			programID: parse.Token2022ID,
			data:      []byte{26, 3, 1},
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawWithheldTokensFromAccounts","info":{
				"mint":%q,"feeRecipient":%q,"sourceAccounts":[%q],
				"multisigWithdrawWithheldAuthority":%q,"signers":[%q]}}`,
				k(1), k(2), k(5), k(3), k(4)),
		},
		{
			name:      "harvestWithheldTokensToMint empty sources",
			programID: parse.Token2022ID,
			data:      []byte{26, 4},
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"harvestWithheldTokensToMint","info":{
				"mint":%q,"sourceAccounts":[]}}`, k(1)),
		},
		{
			name:      "harvestWithheldTokensToMint with sources",
			programID: parse.Token2022ID,
			data:      []byte{26, 4},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"harvestWithheldTokensToMint","info":{
				"mint":%q,"sourceAccounts":[%q,%q]}}`, k(1), k(2), k(3)),
		},
		{
			name:      "setTransferFee single",
			programID: parse.Token2022ID,
			data:      cat([]byte{26, 5}, le16(100), le64(5000)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"setTransferFee","info":{
				"mint":%q,"transferFeeBasisPoints":100,"maximumFee":5000,
				"transferFeeConfigAuthority":%q}}`, k(1), k(2)),
		},
		{
			// Agave's literal typo: lowercase-t multisigtransferFee...
			name:      "setTransferFee multisig typo field",
			programID: parse.Token2022ID,
			data:      cat([]byte{26, 5}, le16(100), le64(5000)),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"setTransferFee","info":{
				"mint":%q,"transferFeeBasisPoints":100,"maximumFee":5000,
				"multisigtransferFeeConfigAuthority":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			// The transferFee sub-unpack is manual prefix reads: trailing
			// bytes are TOLERATED here, unlike the pod families below.
			name:      "setTransferFee trailing bytes tolerated",
			programID: parse.Token2022ID,
			data:      cat([]byte{26, 5}, le16(100), le64(5000), []byte{7, 7}),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"setTransferFee","info":{
				"mint":%q,"transferFeeBasisPoints":100,"maximumFee":5000,
				"transferFeeConfigAuthority":%q}}`, k(1), k(2)),
		},
	})
}

func TestToken22ConfidentialTransferVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			name:      "initializeConfidentialTransferMint all set",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 0}, kb(0x85), []byte{1}, rep(32, 0x86)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeConfidentialTransferMint","info":{
				"mint":%q,"authority":%q,"autoApproveNewAccounts":true,
				"auditorElGamalPubkey":%q}}`, k(1), k(0x85), b64x86x32),
		},
		{
			// Zero authority OMITS the key; zero auditor renders NULL.
			name:      "initializeConfidentialTransferMint zeros split null vs omit",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 0}, rep(32, 0), []byte{0}, rep(32, 0)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeConfidentialTransferMint","info":{
				"mint":%q,"autoApproveNewAccounts":false,"auditorElGamalPubkey":null}}`, k(1)),
		},
		{
			name:      "updateConfidentialTransferMint",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 1}, []byte{1}, rep(32, 0x87)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateConfidentialTransferMint","info":{
				"mint":%q,"confidentialTransferMintAuthority":%q,
				"autoApproveNewAccounts":true,"auditorElGamalPubkey":%q}}`,
				k(1), k(2), b64x87x32),
		},
		{
			name:      "configureConfidentialTransferAccount context state",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 2}, rep(36, 0x88), le64(5), []byte{0}),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"configureConfidentialTransferAccount","info":{
				"account":%q,"mint":%q,"decryptableZeroBalance":%q,
				"maximumPendingBalanceCreditCounter":5,"proofInstructionOffset":0,
				"proofContextStateAccount":%q,"owner":%q}}`,
				k(1), k(2), b64x88x36, k(3), k(4)),
		},
		{
			// proofInstructionOffset is a SIGNED i8: 0xFF renders -1.
			name:      "configureConfidentialTransferAccount sysvar and record",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 2}, rep(36, 0x88), le64(5), []byte{0xFF}),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"configureConfidentialTransferAccount","info":{
				"account":%q,"mint":%q,"decryptableZeroBalance":%q,
				"maximumPendingBalanceCreditCounter":5,"proofInstructionOffset":-1,
				"instructionsSysvar":%q,"recordAccount":%q,"owner":%q}}`,
				k(1), k(2), b64x88x36, k(3), k(4), k(5)),
		},
		{
			name:      "configureConfidentialTransferAccount sysvar without record",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 2}, rep(36, 0x88), le64(5), []byte{1}),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"configureConfidentialTransferAccount","info":{
				"account":%q,"mint":%q,"decryptableZeroBalance":%q,
				"maximumPendingBalanceCreditCounter":5,"proofInstructionOffset":1,
				"instructionsSysvar":%q,"owner":%q}}`,
				k(1), k(2), b64x88x36, k(3), k(4)),
		},
		{
			// No pod payload: trailing bytes tolerated on this arm.
			name:      "approveConfidentialTransferAccount trailing tolerated",
			programID: parse.Token2022ID,
			data:      []byte{27, 3, 0xEE},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"approveConfidentialTransferAccount","info":{
				"account":%q,"mint":%q,"confidentialTransferAuditorAuthority":%q}}`,
				k(1), k(2), k(3)),
		},
		{
			name:      "emptyConfidentialTransferAccount context state",
			programID: parse.Token2022ID,
			data:      []byte{27, 4, 0},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"emptyConfidentialTransferAccount","info":{
				"account":%q,"proofInstructionOffset":0,
				"proofContextStateAccount":%q,"owner":%q}}`, k(1), k(2), k(3)),
		},
		{
			name:      "emptyConfidentialTransferAccount sysvar and record",
			programID: parse.Token2022ID,
			data:      []byte{27, 4, 0xFE},
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"emptyConfidentialTransferAccount","info":{
				"account":%q,"proofInstructionOffset":-2,
				"instructionsSysvar":%q,"recordAccount":%q,"owner":%q}}`,
				k(1), k(2), k(3), k(4)),
		},
		{
			// amount renders as a JSON NUMBER here, unlike base-set amounts.
			name:      "depositConfidentialTransfer amount is a number",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 5}, le64(42), []byte{6}),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"depositConfidentialTransfer","info":{
				"source":%q,"destination":%q,"mint":%q,"amount":42,"decimals":6,
				"owner":%q}}`, k(1), k(2), k(3), k(4)),
		},
		{
			name:      "depositConfidentialTransfer multisig",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 5}, le64(42), []byte{6}),
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"depositConfidentialTransfer","info":{
				"source":%q,"destination":%q,"mint":%q,"amount":42,"decimals":6,
				"multisigOwner":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4), k(5), k(6)),
		},
		{
			name:      "withdrawConfidentialTransfer mixed proof offsets",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 6}, le64(9), []byte{2}, rep(36, 0x89), []byte{1}, []byte{0}),
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawConfidentialTransfer","info":{
				"source":%q,"destination":%q,"mint":%q,"amount":9,"decimals":2,
				"newDecryptableAvailableBalance":%q,
				"equalityProofInstructionOffset":1,"rangeProofInstructionOffset":0,
				"instructionsSysvar":%q,"equalityProofRecordAccount":%q,"owner":%q}}`,
				k(1), k(2), k(3), b64x89x36, k(4), k(5), k(6)),
		},
		{
			name:      "withdrawConfidentialTransfer both context state",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 6}, le64(1), []byte{0}, rep(36, 0x89), []byte{0, 0}),
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawConfidentialTransfer","info":{
				"source":%q,"destination":%q,"mint":%q,"amount":1,"decimals":0,
				"newDecryptableAvailableBalance":%q,
				"equalityProofInstructionOffset":0,"rangeProofInstructionOffset":0,
				"equalityProofContextStateAccount":%q,"rangeProofContextStateAccount":%q,
				"owner":%q}}`,
				k(1), k(2), k(3), b64x89x36, k(4), k(5), k(6)),
		},
		{
			// The longest CT form used by the truncation sweep: 168-byte sub.
			name:      "confidentialTransfer full proof account walk",
			programID: parse.Token2022ID,
			data: cat([]byte{27, 7}, rep(36, 0x8A), rep(64, 0x8B), rep(64, 0x8C),
				[]byte{1, 2, 3}),
			accounts: accs(1, 2, 3, 4, 5, 6, 7, 8),
			program:  "spl-token",
			parsed: fmt.Sprintf(`{"type":"confidentialTransfer","info":{
				"source":%q,"mint":%q,"destination":%q,
				"newSourceDecryptableAvailableBalance":%q,
				"equalityProofInstructionOffset":1,
				"ciphertextValidityProofInstructionOffset":2,
				"rangeProofInstructionOffset":3,
				"instructionsSysvar":%q,"equalityProofRecordAccount":%q,
				"ciphertextValidityProofRecordAccount":%q,"rangeProofRecordAccount":%q,
				"owner":%q}}`,
				k(1), k(2), k(3), b64x8Ax36, k(4), k(5), k(6), k(7), k(8)),
		},
		{
			name:      "confidentialTransfer minimal accounts no proof walk",
			programID: parse.Token2022ID,
			data: cat([]byte{27, 7}, rep(36, 0x8A), rep(64, 0x8B), rep(64, 0x8C),
				[]byte{0, 0, 0}),
			accounts: accs(1, 2, 3, 4),
			program:  "spl-token",
			parsed: fmt.Sprintf(`{"type":"confidentialTransfer","info":{
				"source":%q,"mint":%q,"destination":%q,
				"newSourceDecryptableAvailableBalance":%q,
				"equalityProofInstructionOffset":0,
				"ciphertextValidityProofInstructionOffset":0,
				"rangeProofInstructionOffset":0,"owner":%q}}`,
				k(1), k(2), k(3), b64x8Ax36, k(4)),
		},
		{
			name:      "applyPendingConfidentialTransferBalance",
			programID: parse.Token2022ID,
			data:      cat([]byte{27, 8}, le64(3), rep(36, 0x8D)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"applyPendingConfidentialTransferBalance","info":{
				"account":%q,"expectedPendingBalanceCreditCounter":3,
				"newDecryptableAvailableBalance":%q,"owner":%q}}`,
				k(1), b64x8Dx36, k(2)),
		},
		{
			name:      "enableConfidentialTransferConfidentialCredits",
			programID: parse.Token2022ID,
			data:      []byte{27, 9},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"enableConfidentialTransferConfidentialCredits","info":{
				"account":%q,"owner":%q}}`, k(1), k(2)),
		},
		{
			name:      "disableConfidentialTransferConfidentialCredits multisig",
			programID: parse.Token2022ID,
			data:      []byte{27, 10},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"disableConfidentialTransferConfidentialCredits","info":{
				"account":%q,"multisigOwner":%q,"signers":[%q]}}`, k(1), k(2), k(3)),
		},
		{
			name:      "enableConfidentialTransferNonConfidentialCredits",
			programID: parse.Token2022ID,
			data:      []byte{27, 11},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"enableConfidentialTransferNonConfidentialCredits","info":{
				"account":%q,"owner":%q}}`, k(1), k(2)),
		},
		{
			name:      "disableConfidentialTransferNonConfidentialCredits",
			programID: parse.Token2022ID,
			data:      []byte{27, 12},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"disableConfidentialTransferNonConfidentialCredits","info":{
				"account":%q,"owner":%q}}`, k(1), k(2)),
		},
		{
			// Triple order pin: wire order eq/transferCV/feeSigma/feeCV/range
			// (offsets 1,2,3,4,5) vs labeling order eq/transferCV/feeCV/
			// feeSigma/range (accounts k5..k9).
			name:      "confidentialTransferWithFee three distinct orders",
			programID: parse.Token2022ID,
			data: cat([]byte{27, 13}, rep(36, 0x8A), rep(64, 0x8B), rep(64, 0x8C),
				[]byte{1, 2, 3, 4, 5}),
			accounts: accs(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
			program:  "spl-token",
			parsed: fmt.Sprintf(`{"type":"confidentialTransferWithFee","info":{
				"source":%q,"mint":%q,"destination":%q,
				"newSourceDecryptableAvailableBalance":%q,
				"equalityProofInstructionOffset":1,
				"transferAmountCiphertextValidityProofInstructionOffset":2,
				"feeSigmaProofInstructionOffset":3,
				"feeCiphertextValidityProofInstructionOffset":4,
				"rangeProofInstructionOffset":5,
				"instructionsSysvar":%q,
				"equalityProofRecordAccount":%q,
				"transferAmountCiphertextValidityProofRecordAccount":%q,
				"feeCiphertextValidityProofRecordAccount":%q,
				"feeSigmaProofRecordAccount":%q,
				"rangeProofRecordAccount":%q,
				"owner":%q}}`,
				k(1), k(2), k(3), b64x8Ax36, k(4), k(5), k(6), k(7), k(8), k(9), k(10)),
		},
		{
			name:      "configureConfidentialAccountWithRegistry",
			programID: parse.Token2022ID,
			data:      []byte{27, 14},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"configureConfidentialAccountWithRegistry","info":{
				"account":%q,"mint":%q,"registry":%q}}`, k(1), k(2), k(3)),
		},
	})
}

func TestToken22ConfidentialTransferFeeVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			name:      "initializeConfidentialTransferFeeConfig authority set",
			programID: parse.Token2022ID,
			data:      cat([]byte{37, 0}, kb(0x8E), rep(32, 0x8F)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeConfidentialTransferFeeConfig","info":{
				"mint":%q,"authority":%q,"withdrawWithheldAuthorityElGamalPubkey":%q}}`,
				k(1), k(0x8E), b64x8Fx32),
		},
		{
			// The ElGamal key is a PLAIN pod: all-zero still renders base64
			// (always Some via std's blanket From<T>), while the all-zero
			// OptionalNonZeroPubkey authority is omitted.
			name:      "initializeConfidentialTransferFeeConfig zeros keep elgamal",
			programID: parse.Token2022ID,
			data:      cat([]byte{37, 0}, rep(32, 0), rep(32, 0)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeConfidentialTransferFeeConfig","info":{
				"mint":%q,"withdrawWithheldAuthorityElGamalPubkey":%q}}`, k(1), b64zeros32),
		},
		{
			name:      "withdrawWithheldConfidentialTransferTokensFromMint context",
			programID: parse.Token2022ID,
			data:      cat([]byte{37, 1}, []byte{0}, rep(36, 0x88)),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawWithheldConfidentialTransferTokensFromMint","info":{
				"mint":%q,"feeRecipient":%q,"proofInstructionOffset":0,
				"newDecryptableAvailableBalance":%q,
				"proofContextStateAccount":%q,"withdrawWithheldAuthority":%q}}`,
				k(1), k(2), b64x88x36, k(3), k(4)),
		},
		{
			name:      "withdrawWithheldConfidentialTransferTokensFromMint record",
			programID: parse.Token2022ID,
			data:      cat([]byte{37, 1}, []byte{0xFE}, rep(36, 0x88)),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawWithheldConfidentialTransferTokensFromMint","info":{
				"mint":%q,"feeRecipient":%q,"proofInstructionOffset":-2,
				"newDecryptableAvailableBalance":%q,
				"instructionsSysvar":%q,"recordAccount":%q,"withdrawWithheldAuthority":%q}}`,
				k(1), k(2), b64x88x36, k(3), k(4), k(5)),
		},
		{
			name:      "withdrawWithheldConfidentialTransferTokensFromAccounts context",
			programID: parse.Token2022ID,
			data:      cat([]byte{37, 2}, []byte{1}, []byte{0}, rep(36, 0x89)),
			accounts:  accs(1, 2, 3, 4, 5),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawWithheldConfidentialTransferTokensFromAccounts","info":{
				"mint":%q,"feeRecipient":%q,"proofInstructionOffset":0,
				"newDecryptableAvailableBalance":%q,"proofContextStateAccount":%q,
				"sourceAccounts":[%q],"withdrawWithheldAuthority":%q}}`,
				k(1), k(2), b64x89x36, k(3), k(5), k(4)),
		},
		{
			// The optional proof account is labeled proofAccount HERE (vs
			// recordAccount in the FromMint arm) and gated on the pre-source
			// index, not the raw account count.
			name:      "withdrawWithheldConfidentialTransferTokensFromAccounts proofAccount label",
			programID: parse.Token2022ID,
			data:      cat([]byte{37, 2}, []byte{1}, []byte{1}, rep(36, 0x89)),
			accounts:  accs(1, 2, 3, 4, 5, 6),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"withdrawWithheldConfidentialTransferTokensFromAccounts","info":{
				"mint":%q,"feeRecipient":%q,"proofInstructionOffset":1,
				"newDecryptableAvailableBalance":%q,"instructionsSysvar":%q,
				"proofAccount":%q,"sourceAccounts":[%q],"withdrawWithheldAuthority":%q}}`,
				k(1), k(2), b64x89x36, k(3), k(4), k(6), k(5)),
		},
		{
			name:      "harvestWithheldConfidentialTransferTokensToMint empty",
			programID: parse.Token2022ID,
			data:      []byte{37, 3},
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"harvestWithheldConfidentialTransferTokensToMint","info":{
				"mint":%q,"sourceAccounts":[]}}`, k(1)),
		},
		{
			name:      "harvestWithheldConfidentialTransferTokensToMint sources",
			programID: parse.Token2022ID,
			data:      []byte{37, 3},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"harvestWithheldConfidentialTransferTokensToMint","info":{
				"mint":%q,"sourceAccounts":[%q]}}`, k(1), k(2)),
		},
		{
			name:      "enableConfidentialTransferFeeHarvestToMint",
			programID: parse.Token2022ID,
			data:      []byte{37, 4},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"enableConfidentialTransferFeeHarvestToMint","info":{
				"account":%q,"owner":%q}}`, k(1), k(2)),
		},
		{
			name:      "disableConfidentialTransferFeeHarvestToMint multisig",
			programID: parse.Token2022ID,
			data:      []byte{37, 5},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"disableConfidentialTransferFeeHarvestToMint","info":{
				"account":%q,"multisigOwner":%q,"signers":[%q]}}`, k(1), k(2), k(3)),
		},
	})
}

func TestToken22ConfidentialMintBurnVectors(t *testing.T) {
	runVectors(t, []vector{
		{
			// Plain pods: the all-zero ElGamal key still renders as base64.
			name:      "initializeConfidentialMintBurnMint zero supply key stays base64",
			programID: parse.Token2022ID,
			data:      cat([]byte{42, 0}, rep(32, 0), rep(36, 0x91)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeConfidentialMintBurnMint","info":{
				"mint":%q,"supplyElGamalPubkey":%q,"decryptableSupply":%q}}`,
				k(1), b64zeros32, b64x91x36),
		},
		{
			// Offset 0 is labeled proofAccount here — not ContextState.
			name:      "rotateConfidentialMintBurnSupplyElGamalPubkey proofAccount",
			programID: parse.Token2022ID,
			data:      cat([]byte{42, 1}, rep(32, 0x92), []byte{0}),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"rotateConfidentialMintBurnSupplyElGamalPubkey","info":{
				"mint":%q,"newSupplyElGamalPubkey":%q,"proofInstructionOffset":0,
				"proofAccount":%q,"owner":%q}}`, k(1), b64x92x32, k(2), k(3)),
		},
		{
			name:      "rotateConfidentialMintBurnSupplyElGamalPubkey sysvar",
			programID: parse.Token2022ID,
			data:      cat([]byte{42, 1}, rep(32, 0x92), []byte{1}),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"rotateConfidentialMintBurnSupplyElGamalPubkey","info":{
				"mint":%q,"newSupplyElGamalPubkey":%q,"proofInstructionOffset":1,
				"instructionsSysvar":%q,"owner":%q}}`, k(1), b64x92x32, k(2), k(3)),
		},
		{
			name:      "updateConfidentialMintBurnDecryptableSupply",
			programID: parse.Token2022ID,
			data:      cat([]byte{42, 2}, rep(36, 0x93)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateConfidentialMintBurnDecryptableSupply","info":{
				"mint":%q,"newDecryptableSupply":%q,"owner":%q}}`, k(1), b64x93x36, k(2)),
		},
		{
			name:      "updateConfidentialMintBurnDecryptableSupply multisig",
			programID: parse.Token2022ID,
			data:      cat([]byte{42, 2}, rep(36, 0x93)),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateConfidentialMintBurnDecryptableSupply","info":{
				"mint":%q,"newDecryptableSupply":%q,"multisigOwner":%q,"signers":[%q]}}`,
				k(1), b64x93x36, k(2), k(3)),
		},
		{
			name:      "confidentialMint minimal",
			programID: parse.Token2022ID,
			data: cat([]byte{42, 3}, rep(36, 0x91), rep(64, 0x8B), rep(64, 0x8C),
				[]byte{0, 0, 0}),
			accounts: accs(1, 2, 3),
			program:  "spl-token",
			parsed: fmt.Sprintf(`{"type":"confidentialMint","info":{
				"destination":%q,"mint":%q,"newDecryptableSupply":%q,
				"equalityProofInstructionOffset":0,
				"ciphertextValidityProofInstructionOffset":0,
				"rangeProofInstructionOffset":0,"owner":%q}}`,
				k(1), k(2), b64x91x36, k(3)),
		},
		{
			name:      "confidentialMint full proof walk",
			programID: parse.Token2022ID,
			data: cat([]byte{42, 3}, rep(36, 0x91), rep(64, 0x8B), rep(64, 0x8C),
				[]byte{1, 1, 1}),
			accounts: accs(1, 2, 3, 4, 5, 6, 7),
			program:  "spl-token",
			parsed: fmt.Sprintf(`{"type":"confidentialMint","info":{
				"destination":%q,"mint":%q,"newDecryptableSupply":%q,
				"equalityProofInstructionOffset":1,
				"ciphertextValidityProofInstructionOffset":1,
				"rangeProofInstructionOffset":1,
				"instructionsSysvar":%q,"equalityProofRecordAccount":%q,
				"ciphertextValidityProofRecordAccount":%q,"rangeProofRecordAccount":%q,
				"owner":%q}}`,
				k(1), k(2), b64x91x36, k(3), k(4), k(5), k(6), k(7)),
		},
		{
			// Burn renames the balance field newDecryptableAvailableBalance.
			name:      "confidentialBurn context state accounts",
			programID: parse.Token2022ID,
			data: cat([]byte{42, 4}, rep(36, 0x8D), rep(64, 0x8B), rep(64, 0x8C),
				[]byte{0, 0, 0}),
			accounts: accs(1, 2, 3, 4, 5, 6),
			program:  "spl-token",
			parsed: fmt.Sprintf(`{"type":"confidentialBurn","info":{
				"destination":%q,"mint":%q,"newDecryptableAvailableBalance":%q,
				"equalityProofInstructionOffset":0,
				"ciphertextValidityProofInstructionOffset":0,
				"rangeProofInstructionOffset":0,
				"equalityProofContextStateAccount":%q,
				"ciphertextValidityProofContextStateAccount":%q,
				"rangeProofContextStateAccount":%q,"owner":%q}}`,
				k(1), k(2), b64x8Dx36, k(3), k(4), k(5), k(6)),
		},
		{
			name:      "applyPendingBurn",
			programID: parse.Token2022ID,
			data:      []byte{42, 5},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"applyPendingBurn","info":{
				"mint":%q,"owner":%q}}`, k(1), k(2)),
		},
	})
}

func TestToken22SmallExtensionVectors(t *testing.T) {
	runVectors(t, []vector{
		// --- defaultAccountState (28) ---
		{
			name:      "initializeDefaultAccountState uninitialized",
			programID: parse.Token2022ID,
			data:      []byte{28, 0, 0},
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeDefaultAccountState","info":{
				"mint":%q,"accountState":"uninitialized"}}`, k(1)),
		},
		{
			name:      "initializeDefaultAccountState initialized",
			programID: parse.Token2022ID,
			data:      []byte{28, 0, 1},
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeDefaultAccountState","info":{
				"mint":%q,"accountState":"initialized"}}`, k(1)),
		},
		{
			name:      "updateDefaultAccountState frozen single",
			programID: parse.Token2022ID,
			data:      []byte{28, 1, 2},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateDefaultAccountState","info":{
				"mint":%q,"accountState":"frozen","freezeAuthority":%q}}`, k(1), k(2)),
		},
		{
			name:      "updateDefaultAccountState multisig",
			programID: parse.Token2022ID,
			data:      []byte{28, 1, 2},
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateDefaultAccountState","info":{
				"mint":%q,"accountState":"frozen",
				"multisigFreezeAuthority":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4)),
		},

		// --- memoTransfer (30) ---
		{
			name:      "enableRequiredMemoTransfers",
			programID: parse.Token2022ID,
			data:      []byte{30, 0},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"enableRequiredMemoTransfers","info":{
				"account":%q,"owner":%q}}`, k(1), k(2)),
		},
		{
			name:      "disableRequiredMemoTransfers multisig",
			programID: parse.Token2022ID,
			data:      []byte{30, 1},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"disableRequiredMemoTransfers","info":{
				"account":%q,"multisigOwner":%q,"signers":[%q]}}`, k(1), k(2), k(3)),
		},

		// --- cpiGuard (34) ---
		{
			name:      "enableCpiGuard",
			programID: parse.Token2022ID,
			data:      []byte{34, 0},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"enableCpiGuard","info":{
				"account":%q,"owner":%q}}`, k(1), k(2)),
		},
		{
			name:      "disableCpiGuard multisig",
			programID: parse.Token2022ID,
			data:      []byte{34, 1},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"disableCpiGuard","info":{
				"account":%q,"multisigOwner":%q,"signers":[%q]}}`, k(1), k(2), k(3)),
		},

		// --- interestBearingMint (33) ---
		{
			// rateAuthority is PRESENT-and-null when the pubkey is all-zero
			// (unlike the pointer extensions which omit); rate is a signed
			// i16 number.
			name:      "initializeInterestBearingConfig null authority negative rate",
			programID: parse.Token2022ID,
			data:      cat([]byte{33, 0}, rep(32, 0), i16le(-100)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeInterestBearingConfig","info":{
				"mint":%q,"rateAuthority":null,"rate":-100}}`, k(1)),
		},
		{
			name:      "initializeInterestBearingConfig authority set",
			programID: parse.Token2022ID,
			data:      cat([]byte{33, 0}, kb(0x94), i16le(500)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeInterestBearingConfig","info":{
				"mint":%q,"rateAuthority":%q,"rate":500}}`, k(1), k(0x94)),
		},
		{
			name:      "updateInterestBearingConfigRate single",
			programID: parse.Token2022ID,
			data:      cat([]byte{33, 1}, i16le(-1)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateInterestBearingConfigRate","info":{
				"mint":%q,"newRate":-1,"rateAuthority":%q}}`, k(1), k(2)),
		},
		{
			name:      "updateInterestBearingConfigRate multisig",
			programID: parse.Token2022ID,
			data:      cat([]byte{33, 1}, i16le(1)),
			accounts:  accs(1, 2, 3, 4),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateInterestBearingConfigRate","info":{
				"mint":%q,"newRate":1,"multisigRateAuthority":%q,"signers":[%q,%q]}}`,
				k(1), k(2), k(3), k(4)),
		},

		// --- pointer families (36/39/40/41): OptionalNonZeroPubkey fields
		// are OMITTED when zero, covering every present/absent combination
		// across the four structurally identical families. ---
		{
			name:      "initializeTransferHook both set",
			programID: parse.Token2022ID,
			data:      cat([]byte{36, 0}, kb(0x95), kb(0x96)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeTransferHook","info":{
				"mint":%q,"authority":%q,"programId":%q}}`, k(1), k(0x95), k(0x96)),
		},
		{
			name:      "updateTransferHook set single authority",
			programID: parse.Token2022ID,
			data:      cat([]byte{36, 1}, kb(0x97)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateTransferHook","info":{
				"mint":%q,"programId":%q,"authority":%q}}`, k(1), k(0x97), k(2)),
		},
		{
			name:      "initializeMetadataPointer both zero omit both",
			programID: parse.Token2022ID,
			data:      cat([]byte{39, 0}, rep(32, 0), rep(32, 0)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed:    fmt.Sprintf(`{"type":"initializeMetadataPointer","info":{"mint":%q}}`, k(1)),
		},
		{
			name:      "updateMetadataPointer zero address multisig",
			programID: parse.Token2022ID,
			data:      cat([]byte{39, 1}, rep(32, 0)),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateMetadataPointer","info":{
				"mint":%q,"multisigAuthority":%q,"signers":[%q]}}`, k(1), k(2), k(3)),
		},
		{
			name:      "initializeGroupPointer authority only",
			programID: parse.Token2022ID,
			data:      cat([]byte{40, 0}, kb(0x98), rep(32, 0)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeGroupPointer","info":{
				"mint":%q,"authority":%q}}`, k(1), k(0x98)),
		},
		{
			name:      "updateGroupPointer set",
			programID: parse.Token2022ID,
			data:      cat([]byte{40, 1}, kb(0x99)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateGroupPointer","info":{
				"mint":%q,"groupAddress":%q,"authority":%q}}`, k(1), k(0x99), k(2)),
		},
		{
			name:      "initializeGroupMemberPointer address only",
			programID: parse.Token2022ID,
			data:      cat([]byte{41, 0}, rep(32, 0), kb(0x9A)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeGroupMemberPointer","info":{
				"mint":%q,"memberAddress":%q}}`, k(1), k(0x9A)),
		},
		{
			name:      "updateGroupMemberPointer zero address single",
			programID: parse.Token2022ID,
			data:      cat([]byte{41, 1}, rep(32, 0)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateGroupMemberPointer","info":{
				"mint":%q,"authority":%q}}`, k(1), k(2)),
		},

		// --- scaledUiAmount (43): multiplier through Rust f64 Display. ---
		{
			name:      "initializeScaledUiAmountConfig null authority multiplier string",
			programID: parse.Token2022ID,
			data:      cat([]byte{43, 0}, rep(32, 0), f64le(1.5)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeScaledUiAmountConfig","info":{
				"mint":%q,"authority":null,"multiplier":"1.5"}}`, k(1)),
		},
		{
			name:      "initializeScaledUiAmountConfig authority set shortest decimal",
			programID: parse.Token2022ID,
			data:      cat([]byte{43, 0}, kb(0x9B), f64le(0.1)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializeScaledUiAmountConfig","info":{
				"mint":%q,"authority":%q,"multiplier":"0.1"}}`, k(1), k(0x9B)),
		},
		{
			// 2.0 prints "2"; the timestamp stays a NUMBER.
			name:      "updateMultiplier integral multiplier string",
			programID: parse.Token2022ID,
			data:      cat([]byte{43, 1}, f64le(2), i64le(1699999999)),
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateMultiplier","info":{
				"mint":%q,"newMultiplier":"2","newMultiplierTimestamp":1699999999,
				"authority":%q}}`, k(1), k(2)),
		},
		{
			name:      "updateMultiplier multisig",
			programID: parse.Token2022ID,
			data:      cat([]byte{43, 1}, f64le(2), i64le(-1)),
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"updateMultiplier","info":{
				"mint":%q,"newMultiplier":"2","newMultiplierTimestamp":-1,
				"multisigAuthority":%q,"signers":[%q]}}`, k(1), k(2), k(3)),
		},

		// --- pausable (44): plain-Pubkey authority always renders. ---
		{
			name:      "initializePausableConfig zero authority still base58",
			programID: parse.Token2022ID,
			data:      cat([]byte{44, 0}, rep(32, 0)),
			accounts:  accs(1),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"initializePausableConfig","info":{
				"mint":%q,"authority":%q}}`, k(1), b58zeros32),
		},
		{
			name:      "pause single",
			programID: parse.Token2022ID,
			data:      []byte{44, 1},
			accounts:  accs(1, 2),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"pause","info":{
				"mint":%q,"authority":%q}}`, k(1), k(2)),
		},
		{
			name:      "resume multisig",
			programID: parse.Token2022ID,
			data:      []byte{44, 2},
			accounts:  accs(1, 2, 3),
			program:   "spl-token",
			parsed: fmt.Sprintf(`{"type":"resume","info":{
				"mint":%q,"multisigAuthority":%q,"signers":[%q]}}`, k(1), k(2), k(3)),
		},
	})
}

// TestToken22NotParsable: pod exact-length strictness (trailing bytes
// REFUSED), sub-discriminant refusals, truncated pre-checks, and account
// shortfalls — everything past unpack stays final (no TLV fall-through).
func TestToken22NotParsable(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		accounts []string
	}{
		// Pod payloads refuse trailing bytes (decode_instruction_data).
		{"deposit pod trailing byte", cat([]byte{27, 5}, le64(1), []byte{0, 0xEE}), accs(1, 2, 3, 4)},
		{"interest update pod trailing byte", cat([]byte{33, 1}, i16le(1), []byte{9}), accs(1, 2)},
		{"transferHook update pod trailing byte", cat([]byte{36, 1}, kb(1), []byte{9}), accs(1, 2)},
		{"pausable init pod trailing byte", cat([]byte{44, 0}, rep(32, 0), []byte{9}), accs(1)},
		{"confidentialTransfer pod trailing byte",
			cat([]byte{27, 7}, rep(167, 1), []byte{9}), accs(1, 2, 3, 4)},
		{"defaultAccountState three payload bytes", []byte{28, 0, 1, 9}, accs(1)},

		// Truncated pods.
		{"confidentialTransfer pod one byte short", cat([]byte{27, 7}, rep(166, 1)), accs(1, 2, 3, 4)},
		{"ctf fromAccounts pod short", cat([]byte{37, 2}, rep(37, 1)), accs(1, 2, 3, 4, 5)},
		{"confidentialMint pod short", cat([]byte{42, 3}, rep(166, 1)), accs(1, 2, 3)},

		// Missing / unknown sub-discriminants.
		{"transferFee missing sub", []byte{26}, accs(1)},
		{"transferFee unknown sub 6", []byte{26, 6}, accs(1, 2)},
		{"confidentialTransfer missing sub", []byte{27}, accs(1)},
		{"confidentialTransfer unknown sub 15", []byte{27, 15}, accs(1, 2)},
		{"defaultAccountState missing sub", []byte{28, 0}, accs(1)},
		{"defaultAccountState unknown sub 2", []byte{28, 2, 1}, accs(1)},
		{"defaultAccountState unknown state 3", []byte{28, 0, 3}, accs(1)},
		{"memoTransfer missing sub", []byte{30}, accs(1, 2)},
		{"memoTransfer unknown sub 2", []byte{30, 2}, accs(1, 2)},
		{"interestBearing missing sub", []byte{33}, accs(1)},
		{"interestBearing unknown sub 2", []byte{33, 2}, accs(1, 2)},
		{"cpiGuard unknown sub 2", []byte{34, 2}, accs(1, 2)},
		{"transferHook missing sub", []byte{36}, accs(1)},
		{"transferHook unknown sub 2", []byte{36, 2}, accs(1, 2)},
		{"ctf missing sub", []byte{37}, accs(1)},
		{"ctf unknown sub 6", []byte{37, 6}, accs(1, 2)},
		{"metadataPointer missing sub", []byte{39}, accs(1)},
		{"groupPointer missing sub", []byte{40}, accs(1)},
		{"groupMemberPointer missing sub", []byte{41}, accs(1)},
		{"confidentialMintBurn missing sub", []byte{42}, accs(1)},
		{"confidentialMintBurn unknown sub 6", []byte{42, 6}, accs(1, 2)},
		{"scaledUiAmount missing sub", []byte{43}, accs(1)},
		{"scaledUiAmount unknown sub 2", []byte{43, 2}, accs(1, 2)},
		{"pausable missing sub", []byte{44}, accs(1)},
		{"pausable unknown sub 3", []byte{44, 3}, accs(1, 2)},

		// COption strictness in the transferFee manual unpack.
		{"initializeTransferFeeConfig COption tag 2",
			cat([]byte{26, 0}, []byte{2}), accs(1)},

		// Account shortfalls after a successful decode.
		{"deposit 3 accounts", cat([]byte{27, 5}, le64(1), []byte{0}), accs(1, 2, 3)},
		{"withdrawWithheldTokensFromAccounts n exceeds accounts", []byte{26, 3, 5}, accs(1, 2, 3, 4)},
		{"ctf fromAccounts n exceeds accounts",
			cat([]byte{37, 2}, []byte{3}, []byte{0}, rep(36, 1)), accs(1, 2, 3, 4)},
		{"reallocate 3 accounts", []byte{29}, accs(1, 2, 3)},
		{"createNativeMint 2 accounts", []byte{31}, accs(1, 2)},
		{"withdrawExcessLamports 2 accounts", []byte{38}, accs(1, 2)},

		// Unknown extension types at unpack level (falls to TLV, no match).
		{"reallocate unknown extension type", cat([]byte{29}, le16(28)), accs(1, 2, 3, 4)},
		{"reallocate odd extension bytes", []byte{29, 1}, accs(1, 2, 3, 4)},
		{"unknown discriminant 45", []byte{45}, accs(1)},
		{"unknown discriminant 45 with TLV-length payload", cat([]byte{45}, rep(15, 0)), accs(1)},
	}
	for _, tc := range cases {
		env, err := parse.Parse(parse.Token2022ID, tc.data, tc.accounts, nil)
		if err == nil {
			t.Errorf("%s: Parse() succeeded, want ErrNotParsable; parsed: %s", tc.name, env.Parsed)
			continue
		}
		if !errors.Is(err, parse.ErrNotParsable) {
			t.Errorf("%s: error %v does not wrap ErrNotParsable", tc.name, err)
		}
	}
}
