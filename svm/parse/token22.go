package parse

// Token-2022 extension instructions: TokenInstruction discriminants 25..=44,
// ported from Agave's parse_token.rs extension arms (solana-transaction-status
// 2.3.13) with wire layouts from spl-token-2022 8.0.1.
//
// Error taxonomy (load-bearing, see errTokenUnpack): fields decoded inside
// TokenInstruction::unpack in Rust (disc 25 COption, disc 29 extension-type
// list, disc 35 pubkey, unknown discriminant) wrap errTokenUnpack — Agave
// falls through to the TokenGroup/TokenMetadata TLV parsers on those. Every
// failure past unpack (sub-discriminant decode, pod length mismatch, account
// counts) wraps plain ErrNotParsable and must NOT fall through.

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/blockchain-data-standards/manifesto/svm"
)

// parseTokenExtension owns every TokenInstruction discriminant > 24.
// data is the full instruction data with data[0] == disc.
func parseTokenExtension(disc byte, data []byte, accounts []string) (typeInfo, error) {
	rest := data[1:]
	switch disc {
	case 25: // InitializeMintCloseAuthority — COption decodes at unpack level.
		closeAuthority, _, err := tokenPubkeyOption(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		// Agave renders newAuthority present-null (map_coption_pubkey), unlike
		// initializeMint's freezeAuthority which is omitted when absent.
		info := map[string]any{"mint": accounts[0], "newAuthority": any(nil)}
		if closeAuthority != nil {
			info["newAuthority"] = *closeAuthority
		}
		return typeInfo{Type: "initializeMintCloseAuthority", Info: info}, nil

	case 26: // TransferFeeExtension — no unpack-level fields, no length pre-check.
		return parseTransferFeeExtension(rest, accounts)

	case 27: // ConfidentialTransferExtension
		return parseConfidentialTransferExtension(rest, accounts)

	case 28: // DefaultAccountStateExtension — Agave pre-checks data.len() <= 2.
		if len(data) <= 2 {
			return typeInfo{}, fmt.Errorf("%w: truncated defaultAccountState", ErrNotParsable)
		}
		return parseDefaultAccountStateExtension(rest, accounts)

	case 29: // Reallocate — extension-type list decodes at unpack level.
		if len(rest)%2 != 0 {
			return typeInfo{}, fmt.Errorf("%w: odd extension-type bytes", errTokenUnpack)
		}
		extensions := []string{} // always rendered, [] when empty
		for i := 0; i < len(rest); i += 2 {
			name, ok := uiExtensionTypeName(binary.LittleEndian.Uint16(rest[i:]))
			if !ok {
				return typeInfo{}, fmt.Errorf("%w: extension type %d", errTokenUnpack, binary.LittleEndian.Uint16(rest[i:]))
			}
			extensions = append(extensions, name)
		}
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"account": accounts[0], "payer": accounts[1],
			"systemProgram": accounts[2], "extensionTypes": extensions,
		}
		parseSigners(info, 3, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "reallocate", Info: info}, nil

	case 30: // MemoTransferExtension — Agave pre-checks data.len() < 2.
		if len(data) < 2 {
			return typeInfo{}, fmt.Errorf("%w: truncated memoTransfer", ErrNotParsable)
		}
		return parseEnableDisableExtension(rest, accounts, "RequiredMemoTransfers")

	case 31: // CreateNativeMint
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "createNativeMint", Info: map[string]any{
			"payer": accounts[0], "nativeMint": accounts[1], "systemProgram": accounts[2],
		}}, nil

	case 32: // InitializeNonTransferableMint
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializeNonTransferableMint", Info: map[string]any{
			"mint": accounts[0],
		}}, nil

	case 33: // InterestBearingMintExtension
		if len(data) < 2 {
			return typeInfo{}, fmt.Errorf("%w: truncated interestBearingMint", ErrNotParsable)
		}
		return parseInterestBearingMintExtension(rest, accounts)

	case 34: // CpiGuardExtension
		if len(data) < 2 {
			return typeInfo{}, fmt.Errorf("%w: truncated cpiGuard", ErrNotParsable)
		}
		return parseEnableDisableExtension(rest, accounts, "CpiGuard")

	case 35: // InitializePermanentDelegate — pubkey decodes at unpack level.
		delegate, _, err := tokenPubkey(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializePermanentDelegate", Info: map[string]any{
			"mint": accounts[0], "delegate": delegate,
		}}, nil

	case 36: // TransferHookExtension
		if len(data) < 2 {
			return typeInfo{}, fmt.Errorf("%w: truncated transferHook", ErrNotParsable)
		}
		return parsePointerExtension(rest, accounts, "initializeTransferHook", "updateTransferHook", "programId")

	case 37: // ConfidentialTransferFeeExtension
		if len(data) < 2 {
			return typeInfo{}, fmt.Errorf("%w: truncated confidentialTransferFee", ErrNotParsable)
		}
		return parseConfidentialTransferFeeExtension(rest, accounts)

	case 38: // WithdrawExcessLamports
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{"source": accounts[0], "destination": accounts[1]}
		parseSigners(info, 2, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: "withdrawExcessLamports", Info: info}, nil

	case 39: // MetadataPointerExtension
		if len(data) < 2 {
			return typeInfo{}, fmt.Errorf("%w: truncated metadataPointer", ErrNotParsable)
		}
		return parsePointerExtension(rest, accounts, "initializeMetadataPointer", "updateMetadataPointer", "metadataAddress")

	case 40: // GroupPointerExtension
		if len(data) < 2 {
			return typeInfo{}, fmt.Errorf("%w: truncated groupPointer", ErrNotParsable)
		}
		return parsePointerExtension(rest, accounts, "initializeGroupPointer", "updateGroupPointer", "groupAddress")

	case 41: // GroupMemberPointerExtension
		if len(data) < 2 {
			return typeInfo{}, fmt.Errorf("%w: truncated groupMemberPointer", ErrNotParsable)
		}
		return parsePointerExtension(rest, accounts, "initializeGroupMemberPointer", "updateGroupMemberPointer", "memberAddress")

	case 42: // ConfidentialMintBurnExtension — no length pre-check in Agave.
		return parseConfidentialMintBurnExtension(rest, accounts)

	case 43: // ScaledUiAmountExtension
		return parseScaledUiAmountExtension(rest, accounts)

	case 44: // PausableExtension
		return parsePausableExtension(rest, accounts)
	}
	// Unknown discriminant: TokenInstruction::unpack fails in Agave.
	return typeInfo{}, fmt.Errorf("%w: token discriminant %d", errTokenUnpack, disc)
}

// parseTransferFeeExtension is TransferFeeInstruction (disc 26 sub-set).
// The Rust sub-unpack is manual prefix reads: trailing bytes are tolerated.
func parseTransferFeeExtension(sub []byte, accounts []string) (typeInfo, error) {
	if len(sub) < 1 {
		return typeInfo{}, fmt.Errorf("%w: missing transferFee sub-instruction", ErrNotParsable)
	}
	body := sub[1:]
	switch sub[0] {
	case 0: // InitializeTransferFeeConfig
		configAuthority, body, err := t22COptionPubkey(body)
		if err != nil {
			return typeInfo{}, err
		}
		withdrawAuthority, body, err := t22COptionPubkey(body)
		if err != nil {
			return typeInfo{}, err
		}
		basisPoints, body, err := t22U16(body)
		if err != nil {
			return typeInfo{}, err
		}
		maximumFee, _, err := t22U64(body)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint":                   accounts[0],
			"transferFeeBasisPoints": basisPoints,
			"maximumFee":             maximumFee,
		}
		// Both authorities are omitted-when-absent COptions.
		if configAuthority != nil {
			info["transferFeeConfigAuthority"] = *configAuthority
		}
		if withdrawAuthority != nil {
			info["withdrawWithheldAuthority"] = *withdrawAuthority
		}
		return typeInfo{Type: "initializeTransferFeeConfig", Info: info}, nil

	case 1: // TransferCheckedWithFee
		amount, body, err := t22U64(body)
		if err != nil {
			return typeInfo{}, err
		}
		if len(body) < 1 {
			return typeInfo{}, fmt.Errorf("%w: truncated decimals", ErrNotParsable)
		}
		decimals := body[0]
		fee, _, err := t22U64(body[1:])
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"source":      accounts[0],
			"mint":        accounts[1],
			"destination": accounts[2],
			"tokenAmount": tokenAmount(amount, decimals),
			"feeAmount":   tokenAmount(fee, decimals),
		}
		parseSigners(info, 3, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: "transferCheckedWithFee", Info: info}, nil

	case 2: // WithdrawWithheldTokensFromMint
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{"mint": accounts[0], "feeRecipient": accounts[1]}
		parseSigners(info, 2, accounts, "withdrawWithheldAuthority", "multisigWithdrawWithheldAuthority")
		return typeInfo{Type: "withdrawWithheldTokensFromMint", Info: info}, nil

	case 3: // WithdrawWithheldTokensFromAccounts
		if len(body) < 1 {
			return typeInfo{}, fmt.Errorf("%w: truncated numTokenAccounts", ErrNotParsable)
		}
		n := int(body[0])
		if err := checkNumAccounts(accounts, 3+n); err != nil {
			return typeInfo{}, err
		}
		firstSource := len(accounts) - n
		sources := []string{}
		sources = append(sources, accounts[firstSource:]...)
		info := map[string]any{
			"mint": accounts[0], "feeRecipient": accounts[1], "sourceAccounts": sources,
		}
		parseSigners(info, 2, accounts[:firstSource], "withdrawWithheldAuthority", "multisigWithdrawWithheldAuthority")
		return typeInfo{Type: "withdrawWithheldTokensFromAccounts", Info: info}, nil

	case 4: // HarvestWithheldTokensToMint
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		sources := []string{}
		sources = append(sources, accounts[1:]...)
		return typeInfo{Type: "harvestWithheldTokensToMint", Info: map[string]any{
			"mint": accounts[0], "sourceAccounts": sources,
		}}, nil

	case 5: // SetTransferFee
		basisPoints, body, err := t22U16(body)
		if err != nil {
			return typeInfo{}, err
		}
		maximumFee, _, err := t22U64(body)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint":                   accounts[0],
			"transferFeeBasisPoints": basisPoints,
			"maximumFee":             maximumFee,
		}
		// Agave's multisig field really is lowercase-t "multisigtransferFee...".
		parseSigners(info, 1, accounts, "transferFeeConfigAuthority", "multisigtransferFeeConfigAuthority")
		return typeInfo{Type: "setTransferFee", Info: info}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: transferFee sub-instruction %d", ErrNotParsable, sub[0])
}

// parseConfidentialTransferExtension is ConfidentialTransferInstruction
// (disc 27 sub-set). Pod payloads require the exact encoded length.
func parseConfidentialTransferExtension(sub []byte, accounts []string) (typeInfo, error) {
	if len(sub) < 1 {
		return typeInfo{}, fmt.Errorf("%w: missing confidentialTransfer sub-instruction", ErrNotParsable)
	}
	switch sub[0] {
	case 0: // InitializeMint
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 65) // authority(32) auto_approve(1) auditor(32)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint":                   accounts[0],
			"autoApproveNewAccounts": body[32] != 0,
			"auditorElGamalPubkey":   t22OptionalBase64(body[33:65]),
		}
		if authority := t22OptionalPubkey(body[0:32]); authority != nil {
			info["authority"] = *authority
		}
		return typeInfo{Type: "initializeConfidentialTransferMint", Info: info}, nil

	case 1: // UpdateMint
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 33) // auto_approve(1) auditor(32)
		if err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "updateConfidentialTransferMint", Info: map[string]any{
			"mint":                              accounts[0],
			"confidentialTransferMintAuthority": accounts[1],
			"autoApproveNewAccounts":            body[0] != 0,
			"auditorElGamalPubkey":              t22OptionalBase64(body[1:33]),
		}}, nil

	case 2: // ConfigureAccount
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 45) // zero_balance(36) max_counter(8) proof_offset(1)
		if err != nil {
			return typeInfo{}, err
		}
		proofOffset := int8(body[44])
		info := map[string]any{
			"account":                            accounts[0],
			"mint":                               accounts[1],
			"decryptableZeroBalance":             t22Base64(body[0:36]),
			"maximumPendingBalanceCreditCounter": binary.LittleEndian.Uint64(body[36:44]),
			"proofInstructionOffset":             proofOffset,
		}
		offset := 3
		if proofOffset == 0 {
			info["proofContextStateAccount"] = accounts[2]
		} else {
			info["instructionsSysvar"] = accounts[2]
			// Assume the extra account is a proof account, not a multisig
			// signer — same best-effort guess Agave makes.
			if len(accounts) > 4 {
				info["recordAccount"] = accounts[3]
				offset = 4
			}
		}
		parseSigners(info, offset, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "configureConfidentialTransferAccount", Info: info}, nil

	case 3: // ApproveAccount — no pod payload, trailing bytes tolerated.
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "approveConfidentialTransferAccount", Info: map[string]any{
			"account": accounts[0], "mint": accounts[1],
			"confidentialTransferAuditorAuthority": accounts[2],
		}}, nil

	case 4: // EmptyAccount
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 1) // proof_offset(1)
		if err != nil {
			return typeInfo{}, err
		}
		proofOffset := int8(body[0])
		info := map[string]any{
			"account":                accounts[0],
			"proofInstructionOffset": proofOffset,
		}
		offset := 2
		if proofOffset == 0 {
			info["proofContextStateAccount"] = accounts[1]
		} else {
			info["instructionsSysvar"] = accounts[1]
			if len(accounts) > 3 {
				info["recordAccount"] = accounts[2]
				offset = 3
			}
		}
		parseSigners(info, offset, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "emptyConfidentialTransferAccount", Info: info}, nil

	case 5: // Deposit
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 9) // amount(8) decimals(1)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"source":      accounts[0],
			"destination": accounts[1],
			"mint":        accounts[2],
			// json! renders the u64 as a number here, unlike base-set amounts.
			"amount":   binary.LittleEndian.Uint64(body[0:8]),
			"decimals": body[8],
		}
		parseSigners(info, 3, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "depositConfidentialTransfer", Info: info}, nil

	case 6: // Withdraw
		if err := checkNumAccounts(accounts, 5); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 47) // amount(8) decimals(1) balance(36) eq(1) range(1)
		if err != nil {
			return typeInfo{}, err
		}
		equality, rangeProof := int8(body[45]), int8(body[46])
		info := map[string]any{
			"source":                         accounts[0],
			"destination":                    accounts[1],
			"mint":                           accounts[2],
			"amount":                         binary.LittleEndian.Uint64(body[0:8]),
			"decimals":                       body[8],
			"newDecryptableAvailableBalance": t22Base64(body[9:45]),
			"equalityProofInstructionOffset": equality,
			"rangeProofInstructionOffset":    rangeProof,
		}
		offset := t22ProofAccounts(info, accounts, 3, []t22Proof{
			{"equalityProof", equality}, {"rangeProof", rangeProof},
		})
		parseSigners(info, offset, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "withdrawConfidentialTransfer", Info: info}, nil

	case 7: // Transfer
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		// balance(36) ciphertext_lo(64) ciphertext_hi(64) eq(1) cv(1) range(1);
		// the auditor ciphertexts are not rendered.
		body, err := t22PodData(sub, 167)
		if err != nil {
			return typeInfo{}, err
		}
		equality, validity, rangeProof := int8(body[164]), int8(body[165]), int8(body[166])
		info := map[string]any{
			"source":                               accounts[0],
			"mint":                                 accounts[1],
			"destination":                          accounts[2],
			"newSourceDecryptableAvailableBalance": t22Base64(body[0:36]),
			"equalityProofInstructionOffset":       equality,
			"ciphertextValidityProofInstructionOffset": validity,
			"rangeProofInstructionOffset":              rangeProof,
		}
		offset := t22ProofAccounts(info, accounts, 3, []t22Proof{
			{"equalityProof", equality}, {"ciphertextValidityProof", validity}, {"rangeProof", rangeProof},
		})
		parseSigners(info, offset, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "confidentialTransfer", Info: info}, nil

	case 8: // ApplyPendingBalance
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 44) // counter(8) balance(36)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"account":                             accounts[0],
			"newDecryptableAvailableBalance":      t22Base64(body[8:44]),
			"expectedPendingBalanceCreditCounter": binary.LittleEndian.Uint64(body[0:8]),
		}
		parseSigners(info, 1, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "applyPendingConfidentialTransferBalance", Info: info}, nil

	case 9, 10, 11, 12: // Enable/Disable (Non)ConfidentialCredits
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		types := map[byte]string{
			9:  "enableConfidentialTransferConfidentialCredits",
			10: "disableConfidentialTransferConfidentialCredits",
			11: "enableConfidentialTransferNonConfidentialCredits",
			12: "disableConfidentialTransferNonConfidentialCredits",
		}
		info := map[string]any{"account": accounts[0]}
		parseSigners(info, 1, accounts, "owner", "multisigOwner")
		return typeInfo{Type: types[sub[0]], Info: info}, nil

	case 13: // TransferWithFee
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		// balance(36) lo(64) hi(64) eq(1) transfer_cv(1) fee_sigma(1) fee_cv(1) range(1)
		body, err := t22PodData(sub, 169)
		if err != nil {
			return typeInfo{}, err
		}
		equality := int8(body[164])
		transferValidity := int8(body[165])
		feeSigma := int8(body[166])
		feeValidity := int8(body[167])
		rangeProof := int8(body[168])
		info := map[string]any{
			"source":                               accounts[0],
			"mint":                                 accounts[1],
			"destination":                          accounts[2],
			"newSourceDecryptableAvailableBalance": t22Base64(body[0:36]),
			"equalityProofInstructionOffset":       equality,
			"transferAmountCiphertextValidityProofInstructionOffset": transferValidity,
			"feeCiphertextValidityProofInstructionOffset":            feeValidity,
			"feeSigmaProofInstructionOffset":                         feeSigma,
			"rangeProofInstructionOffset":                            rangeProof,
		}
		offset := t22ProofAccounts(info, accounts, 3, []t22Proof{
			{"equalityProof", equality},
			{"transferAmountCiphertextValidityProof", transferValidity},
			{"feeCiphertextValidityProof", feeValidity},
			{"feeSigmaProof", feeSigma},
			{"rangeProof", rangeProof},
		})
		parseSigners(info, offset, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "confidentialTransferWithFee", Info: info}, nil

	case 14: // ConfigureAccountWithRegistry
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "configureConfidentialAccountWithRegistry", Info: map[string]any{
			"account": accounts[0], "mint": accounts[1], "registry": accounts[2],
		}}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: confidentialTransfer sub-instruction %d", ErrNotParsable, sub[0])
}

// parseConfidentialTransferFeeExtension is ConfidentialTransferFeeInstruction
// (disc 37 sub-set).
func parseConfidentialTransferFeeExtension(sub []byte, accounts []string) (typeInfo, error) {
	if len(sub) < 1 {
		return typeInfo{}, fmt.Errorf("%w: missing confidentialTransferFee sub-instruction", ErrNotParsable)
	}
	switch sub[0] {
	case 0: // InitializeConfidentialTransferFeeConfig
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 64) // authority(32) withdraw_withheld_elgamal(32)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint": accounts[0],
			// The wire field is a plain PodElGamalPubkey (not OptionalNonZero):
			// Agave's Option round-trip is std's blanket From<T>, so this is
			// always Some — always a base64 string, even for all-zero bytes.
			"withdrawWithheldAuthorityElGamalPubkey": t22Base64(body[32:64]),
		}
		if authority := t22OptionalPubkey(body[0:32]); authority != nil {
			info["authority"] = *authority
		}
		return typeInfo{Type: "initializeConfidentialTransferFeeConfig", Info: info}, nil

	case 1: // WithdrawWithheldTokensFromMint
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 37) // proof_offset(1) balance(36)
		if err != nil {
			return typeInfo{}, err
		}
		proofOffset := int8(body[0])
		info := map[string]any{
			"mint":                           accounts[0],
			"feeRecipient":                   accounts[1],
			"proofInstructionOffset":         proofOffset,
			"newDecryptableAvailableBalance": t22Base64(body[1:37]),
		}
		offset := 3
		if proofOffset == 0 {
			info["proofContextStateAccount"] = accounts[2]
		} else {
			info["instructionsSysvar"] = accounts[2]
			if len(accounts) > 4 {
				info["recordAccount"] = accounts[3]
				offset = 4
			}
		}
		parseSigners(info, offset, accounts, "withdrawWithheldAuthority", "multisigWithdrawWithheldAuthority")
		return typeInfo{Type: "withdrawWithheldConfidentialTransferTokensFromMint", Info: info}, nil

	case 2: // WithdrawWithheldTokensFromAccounts — pod decode precedes the count check.
		body, err := t22PodData(sub, 38) // num_accounts(1) proof_offset(1) balance(36)
		if err != nil {
			return typeInfo{}, err
		}
		n := int(body[0])
		if err := checkNumAccounts(accounts, 4+n); err != nil {
			return typeInfo{}, err
		}
		proofOffset := int8(body[1])
		info := map[string]any{
			"mint":                           accounts[0],
			"feeRecipient":                   accounts[1],
			"proofInstructionOffset":         proofOffset,
			"newDecryptableAvailableBalance": t22Base64(body[2:38]),
		}
		firstSource := len(accounts) - n
		offset := 3
		if proofOffset == 0 {
			info["proofContextStateAccount"] = accounts[2]
		} else {
			info["instructionsSysvar"] = accounts[2]
			// Note: labeled proofAccount here, unlike the FromMint arm's
			// recordAccount, and gated on the pre-source count.
			if firstSource > 4 {
				info["proofAccount"] = accounts[3]
				offset = 4
			}
		}
		sources := []string{}
		sources = append(sources, accounts[firstSource:]...)
		info["sourceAccounts"] = sources
		parseSigners(info, offset, accounts[:firstSource], "withdrawWithheldAuthority", "multisigWithdrawWithheldAuthority")
		return typeInfo{Type: "withdrawWithheldConfidentialTransferTokensFromAccounts", Info: info}, nil

	case 3: // HarvestWithheldTokensToMint
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		sources := []string{}
		sources = append(sources, accounts[1:]...)
		return typeInfo{Type: "harvestWithheldConfidentialTransferTokensToMint", Info: map[string]any{
			"mint": accounts[0], "sourceAccounts": sources,
		}}, nil

	case 4, 5: // Enable/DisableHarvestToMint
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		name := "enableConfidentialTransferFeeHarvestToMint"
		if sub[0] == 5 {
			name = "disableConfidentialTransferFeeHarvestToMint"
		}
		info := map[string]any{"account": accounts[0]}
		parseSigners(info, 1, accounts, "owner", "multisigOwner")
		return typeInfo{Type: name, Info: info}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: confidentialTransferFee sub-instruction %d", ErrNotParsable, sub[0])
}

// parseConfidentialMintBurnExtension is ConfidentialMintBurnInstruction
// (disc 42 sub-set).
func parseConfidentialMintBurnExtension(sub []byte, accounts []string) (typeInfo, error) {
	if len(sub) < 1 {
		return typeInfo{}, fmt.Errorf("%w: missing confidentialMintBurn sub-instruction", ErrNotParsable)
	}
	switch sub[0] {
	case 0: // InitializeMint
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 68) // supply_elgamal(32) decryptable_supply(36)
		if err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializeConfidentialMintBurnMint", Info: map[string]any{
			"mint": accounts[0],
			// Plain (non-optional) pods: always base64 strings.
			"supplyElGamalPubkey": t22Base64(body[0:32]),
			"decryptableSupply":   t22Base64(body[32:68]),
		}}, nil

	case 1: // RotateSupplyElGamalPubkey
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 33) // new_supply_elgamal(32) proof_offset(1)
		if err != nil {
			return typeInfo{}, err
		}
		proofOffset := int8(body[32])
		info := map[string]any{
			"mint":                   accounts[0],
			"newSupplyElGamalPubkey": t22Base64(body[0:32]),
			"proofInstructionOffset": proofOffset,
		}
		if proofOffset == 0 {
			info["proofAccount"] = accounts[1]
		} else {
			info["instructionsSysvar"] = accounts[1]
		}
		parseSigners(info, 2, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "rotateConfidentialMintBurnSupplyElGamalPubkey", Info: info}, nil

	case 2: // UpdateDecryptableSupply
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 36) // new_decryptable_supply(36)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint":                 accounts[0],
			"newDecryptableSupply": t22Base64(body[0:36]),
		}
		parseSigners(info, 1, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "updateConfidentialMintBurnDecryptableSupply", Info: info}, nil

	case 3, 4: // Mint / Burn — identical layout, different balance field name.
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		// supply_or_balance(36) lo(64) hi(64) eq(1) cv(1) range(1)
		body, err := t22PodData(sub, 167)
		if err != nil {
			return typeInfo{}, err
		}
		equality, validity, rangeProof := int8(body[164]), int8(body[165]), int8(body[166])
		balanceField, name := "newDecryptableSupply", "confidentialMint"
		if sub[0] == 4 {
			balanceField, name = "newDecryptableAvailableBalance", "confidentialBurn"
		}
		info := map[string]any{
			"destination":                    accounts[0],
			"mint":                           accounts[1],
			balanceField:                     t22Base64(body[0:36]),
			"equalityProofInstructionOffset": equality,
			"ciphertextValidityProofInstructionOffset": validity,
			"rangeProofInstructionOffset":              rangeProof,
		}
		offset := t22ProofAccounts(info, accounts, 2, []t22Proof{
			{"equalityProof", equality}, {"ciphertextValidityProof", validity}, {"rangeProof", rangeProof},
		})
		parseSigners(info, offset, accounts, "owner", "multisigOwner")
		return typeInfo{Type: name, Info: info}, nil

	case 5: // ApplyPendingBurn
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{"mint": accounts[0]}
		parseSigners(info, 1, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "applyPendingBurn", Info: info}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: confidentialMintBurn sub-instruction %d", ErrNotParsable, sub[0])
}

// parseDefaultAccountStateExtension is DefaultAccountStateInstruction
// (disc 28 sub-set). decode_instruction demands exactly two bytes.
func parseDefaultAccountStateExtension(sub []byte, accounts []string) (typeInfo, error) {
	if len(sub) != 2 {
		return typeInfo{}, fmt.Errorf("%w: defaultAccountState wants 2 bytes, got %d", ErrNotParsable, len(sub))
	}
	var state string
	switch sub[1] {
	case 0:
		state = "uninitialized"
	case 1:
		state = "initialized"
	case 2:
		state = "frozen"
	default:
		return typeInfo{}, fmt.Errorf("%w: account state %d", ErrNotParsable, sub[1])
	}
	switch sub[0] {
	case 0: // Initialize
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializeDefaultAccountState", Info: map[string]any{
			"mint": accounts[0], "accountState": state,
		}}, nil
	case 1: // Update
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{"mint": accounts[0], "accountState": state}
		parseSigners(info, 1, accounts, "freezeAuthority", "multisigFreezeAuthority")
		return typeInfo{Type: "updateDefaultAccountState", Info: info}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: defaultAccountState sub-instruction %d", ErrNotParsable, sub[0])
}

// parseEnableDisableExtension covers the twin MemoTransfer (30) and CpiGuard
// (34) sub-sets: Enable=0/Disable=1, account + owner signers, no payload.
// Agave checks accounts before the sub-discriminant; trailing bytes tolerated.
func parseEnableDisableExtension(sub []byte, accounts []string, suffix string) (typeInfo, error) {
	if err := checkNumAccounts(accounts, 2); err != nil {
		return typeInfo{}, err
	}
	var prefix string
	switch sub[0] {
	case 0:
		prefix = "enable"
	case 1:
		prefix = "disable"
	default:
		return typeInfo{}, fmt.Errorf("%w: %s sub-instruction %d", ErrNotParsable, suffix, sub[0])
	}
	info := map[string]any{"account": accounts[0]}
	parseSigners(info, 1, accounts, "owner", "multisigOwner")
	return typeInfo{Type: prefix + suffix, Info: info}, nil
}

// parseInterestBearingMintExtension is InterestBearingMintInstruction
// (disc 33 sub-set).
func parseInterestBearingMintExtension(sub []byte, accounts []string) (typeInfo, error) {
	switch sub[0] {
	case 0: // Initialize
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 34) // rate_authority(32) rate(2, i16 LE)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint": accounts[0],
			// rateAuthority is present-null (OptionalNonZeroPubkey in a json!
			// literal), unlike the pointer extensions which omit the key.
			"rateAuthority": any(nil),
			"rate":          int16(binary.LittleEndian.Uint16(body[32:34])),
		}
		if authority := t22OptionalPubkey(body[0:32]); authority != nil {
			info["rateAuthority"] = *authority
		}
		return typeInfo{Type: "initializeInterestBearingConfig", Info: info}, nil

	case 1: // UpdateRate
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 2)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint":    accounts[0],
			"newRate": int16(binary.LittleEndian.Uint16(body[0:2])),
		}
		parseSigners(info, 1, accounts, "rateAuthority", "multisigRateAuthority")
		return typeInfo{Type: "updateInterestBearingConfigRate", Info: info}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: interestBearingMint sub-instruction %d", ErrNotParsable, sub[0])
}

// parsePointerExtension covers the four structurally identical pointer
// sub-sets — TransferHook (36), MetadataPointer (39), GroupPointer (40),
// GroupMemberPointer (41): Initialize = authority(32) + address(32), Update =
// address(32), both OptionalNonZeroPubkey rendered omit-when-zero.
func parsePointerExtension(sub []byte, accounts []string, initType, updateType, addrField string) (typeInfo, error) {
	switch sub[0] {
	case 0: // Initialize
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 64)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{"mint": accounts[0]}
		if authority := t22OptionalPubkey(body[0:32]); authority != nil {
			info["authority"] = *authority
		}
		if addr := t22OptionalPubkey(body[32:64]); addr != nil {
			info[addrField] = *addr
		}
		return typeInfo{Type: initType, Info: info}, nil

	case 1: // Update
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 32)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{"mint": accounts[0]}
		if addr := t22OptionalPubkey(body[0:32]); addr != nil {
			info[addrField] = *addr
		}
		parseSigners(info, 1, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: updateType, Info: info}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: pointer sub-instruction %d", ErrNotParsable, sub[0])
}

// parseScaledUiAmountExtension is ScaledUiAmountMintInstruction (disc 43
// sub-set).
func parseScaledUiAmountExtension(sub []byte, accounts []string) (typeInfo, error) {
	if len(sub) < 1 {
		return typeInfo{}, fmt.Errorf("%w: missing scaledUiAmount sub-instruction", ErrNotParsable)
	}
	switch sub[0] {
	case 0: // Initialize
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 40) // authority(32) multiplier(8, f64 LE)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint":      accounts[0],
			"authority": any(nil), // present-null OptionalNonZeroPubkey
			// Agave stringifies the multiplier through Rust's f64 Display.
			"multiplier": t22F64String(body[32:40]),
		}
		if authority := t22OptionalPubkey(body[0:32]); authority != nil {
			info["authority"] = *authority
		}
		return typeInfo{Type: "initializeScaledUiAmountConfig", Info: info}, nil

	case 1: // UpdateMultiplier
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 16) // multiplier(8) effective_timestamp(8, i64 LE)
		if err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint":                   accounts[0],
			"newMultiplier":          t22F64String(body[0:8]),
			"newMultiplierTimestamp": int64(binary.LittleEndian.Uint64(body[8:16])),
		}
		parseSigners(info, 1, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: "updateMultiplier", Info: info}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: scaledUiAmount sub-instruction %d", ErrNotParsable, sub[0])
}

// parsePausableExtension is PausableInstruction (disc 44 sub-set).
func parsePausableExtension(sub []byte, accounts []string) (typeInfo, error) {
	if len(sub) < 1 {
		return typeInfo{}, fmt.Errorf("%w: missing pausable sub-instruction", ErrNotParsable)
	}
	switch sub[0] {
	case 0: // Initialize
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		body, err := t22PodData(sub, 32)
		if err != nil {
			return typeInfo{}, err
		}
		// The wire field is a plain Pubkey (not OptionalNonZero): Agave's
		// Option round-trip is the blanket From<T>, so the authority is
		// always a base58 string — an all-zero key renders as the default
		// pubkey, never null.
		return typeInfo{Type: "initializePausableConfig", Info: map[string]any{
			"mint":      accounts[0],
			"authority": svm.Base58Encode(body[0:32]),
		}}, nil

	case 1, 2: // Pause / Resume
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		name := "pause"
		if sub[0] == 2 {
			name = "resume"
		}
		info := map[string]any{"mint": accounts[0]}
		parseSigners(info, 1, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: name, Info: info}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: pausable sub-instruction %d", ErrNotParsable, sub[0])
}

// t22Proof labels one proof-location account in the shared confidential
// withdraw/transfer/mint/burn account layout.
type t22Proof struct {
	label  string
	offset int8
}

// t22ProofAccounts ports Agave's repeated proof-account walk: when any proof
// offset is nonzero the next account is the instructions sysvar, then each
// proof consumes one account labeled either <label>ContextStateAccount
// (offset 0) or <label>RecordAccount — always leaving at least one trailing
// account for the owner. Returns the signer offset for parse_signers.
func t22ProofAccounts(info map[string]any, accounts []string, offset int, proofs []t22Proof) int {
	anyNonZero := false
	for _, p := range proofs {
		if p.offset != 0 {
			anyNonZero = true
		}
	}
	if offset < len(accounts)-1 && anyNonZero {
		info["instructionsSysvar"] = accounts[offset]
		offset++
	}
	// Assume extra accounts are proof accounts, not multisig signers — the
	// same best-effort guess Agave documents.
	for _, p := range proofs {
		if offset < len(accounts)-1 {
			if p.offset == 0 {
				info[p.label+"ContextStateAccount"] = accounts[offset]
			} else {
				info[p.label+"RecordAccount"] = accounts[offset]
			}
			offset++
		}
	}
	return offset
}

// t22PodData is spl-token-2022's decode_instruction_data: sub must be exactly
// one sub-discriminant byte plus the pod payload — trailing bytes refused.
func t22PodData(sub []byte, size int) ([]byte, error) {
	if len(sub) != size+1 {
		return nil, fmt.Errorf("%w: sub-instruction wants %d data bytes, got %d", ErrNotParsable, size, len(sub)-1)
	}
	return sub[1:], nil
}

// t22OptionalPubkey is spl-pod's OptionalNonZeroPubkey: all-zero means None.
func t22OptionalPubkey(b []byte) *string {
	if t22AllZero(b) {
		return nil
	}
	s := svm.Base58Encode(b)
	return &s
}

// t22OptionalBase64 is OptionalNonZeroElGamalPubkey rendered the way the
// confidential arms do: JSON null when all-zero, base64 otherwise.
func t22OptionalBase64(b []byte) any {
	if t22AllZero(b) {
		return nil
	}
	return t22Base64(b)
}

func t22AllZero(b []byte) bool {
	for _, c := range b {
		if c != 0 {
			return false
		}
	}
	return true
}

// t22Base64 renders zk-sdk pod bytes the way their Display impls do:
// standard base64 with padding.
func t22Base64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// t22F64String is Rust's f64 Display: shortest round-tripping decimal, never
// scientific notation, with Rust's spellings for the specials.
func t22F64String(b []byte) string {
	f := math.Float64frombits(binary.LittleEndian.Uint64(b))
	if math.IsInf(f, 1) {
		return "inf"
	}
	if math.IsInf(f, -1) {
		return "-inf"
	}
	return strconv.FormatFloat(f, 'f', -1, 64) // NaN formats as "NaN", like Rust
}

// t22U16/t22U64/t22COptionPubkey are the arm-level (plain ErrNotParsable)
// twins of the unpack-level helpers in token.go: TransferFeeInstruction's
// manual sub-unpack happens inside a parse arm in Agave, so its failures must
// not trigger the TLV interface fall-through.
func t22U16(b []byte) (uint16, []byte, error) {
	if len(b) < 2 {
		return 0, nil, fmt.Errorf("%w: truncated u16", ErrNotParsable)
	}
	return binary.LittleEndian.Uint16(b[:2]), b[2:], nil
}

func t22U64(b []byte) (uint64, []byte, error) {
	if len(b) < 8 {
		return 0, nil, fmt.Errorf("%w: truncated u64", ErrNotParsable)
	}
	return binary.LittleEndian.Uint64(b[:8]), b[8:], nil
}

func t22COptionPubkey(b []byte) (*string, []byte, error) {
	if len(b) < 1 {
		return nil, nil, fmt.Errorf("%w: truncated COption tag", ErrNotParsable)
	}
	switch b[0] {
	case 0:
		return nil, b[1:], nil
	case 1:
		if len(b) < 33 {
			return nil, nil, fmt.Errorf("%w: truncated pubkey", ErrNotParsable)
		}
		s := svm.Base58Encode(b[1:33])
		return &s, b[33:], nil
	}
	return nil, nil, fmt.Errorf("%w: COption tag %d", ErrNotParsable, b[0])
}

// uiExtensionTypeName maps spl-token-2022's ExtensionType (u16) to Agave's
// UiExtensionType camelCase serde name. The cfg(test)-only padding variants
// near u16::MAX do not exist in release Agave and stay unknown here.
func uiExtensionTypeName(t uint16) (string, bool) {
	names := []string{
		"uninitialized",                 // 0
		"transferFeeConfig",             // 1
		"transferFeeAmount",             // 2
		"mintCloseAuthority",            // 3
		"confidentialTransferMint",      // 4
		"confidentialTransferAccount",   // 5
		"defaultAccountState",           // 6
		"immutableOwner",                // 7
		"memoTransfer",                  // 8
		"nonTransferable",               // 9
		"interestBearingConfig",         // 10
		"cpiGuard",                      // 11
		"permanentDelegate",             // 12
		"nonTransferableAccount",        // 13
		"transferHook",                  // 14
		"transferHookAccount",           // 15
		"confidentialTransferFeeConfig", // 16
		"confidentialTransferFeeAmount", // 17
		"metadataPointer",               // 18
		"tokenMetadata",                 // 19
		"groupPointer",                  // 20
		"tokenGroup",                    // 21
		"groupMemberPointer",            // 22
		"tokenGroupMember",              // 23
		"confidentialMintBurn",          // 24
		"scaledUiAmount",                // 25
		"pausable",                      // 26
		"pausableAccount",               // 27
	}
	if int(t) < len(names) {
		return names[t], true
	}
	return "", false
}
