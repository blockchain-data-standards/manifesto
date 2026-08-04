package parse

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/blockchain-data-standards/manifesto/svm"
)

// errTokenUnpack marks failures that correspond to TokenInstruction::unpack
// returning Err in Rust — as opposed to failures inside a successfully
// unpacked arm (account counts). The distinction is load-bearing: Agave only
// falls through to the TokenGroup/TokenMetadata TLV interface parsers when
// unpack itself fails. It wraps ErrNotParsable so callers' fallback rule is
// unchanged.
var errTokenUnpack = fmt.Errorf("%w: does not decode as a token instruction", ErrNotParsable)

// parseToken ports Agave's parse_token.rs for the BASE instruction set,
// discriminants 0..=24 — the layouts shared verbatim by spl-token and
// spl-token-2022. Token-2022 extension discriminants (25+) and the
// group/metadata interface instructions are phase 2: they return
// ErrNotParsable and render partiallyDecoded, which is also what any
// pre-extension validator renders for them.
func parseToken(data []byte, accounts []string) (typeInfo, error) {
	if len(data) == 0 {
		return typeInfo{}, fmt.Errorf("%w: empty data", errTokenUnpack)
	}
	disc, rest := data[0], data[1:]
	switch disc {
	case 0, 20: // InitializeMint / InitializeMint2
		if len(rest) < 1 {
			return typeInfo{}, fmt.Errorf("%w: truncated initializeMint", errTokenUnpack)
		}
		decimals := rest[0]
		mintAuthority, rest, err := tokenPubkey(rest[1:])
		if err != nil {
			return typeInfo{}, err
		}
		freezeAuthority, _, err := tokenPubkeyOption(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if disc == 0 {
			if err := checkNumAccounts(accounts, 2); err != nil {
				return typeInfo{}, err
			}
			info := map[string]any{
				"mint": accounts[0], "decimals": decimals,
				"mintAuthority": mintAuthority, "rentSysvar": accounts[1],
			}
			// Agave omits the key entirely when absent — not present-null.
			if freezeAuthority != nil {
				info["freezeAuthority"] = *freezeAuthority
			}
			return typeInfo{Type: "initializeMint", Info: info}, nil
		}
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint": accounts[0], "decimals": decimals, "mintAuthority": mintAuthority,
		}
		if freezeAuthority != nil {
			info["freezeAuthority"] = *freezeAuthority
		}
		return typeInfo{Type: "initializeMint2", Info: info}, nil

	case 1: // InitializeAccount
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializeAccount", Info: map[string]any{
			"account": accounts[0], "mint": accounts[1],
			"owner": accounts[2], "rentSysvar": accounts[3],
		}}, nil

	case 2, 19: // InitializeMultisig / InitializeMultisig2
		if len(rest) < 1 {
			return typeInfo{}, fmt.Errorf("%w: truncated initializeMultisig", errTokenUnpack)
		}
		m := rest[0]
		if disc == 2 {
			if err := checkNumAccounts(accounts, 3); err != nil {
				return typeInfo{}, err
			}
			return typeInfo{Type: "initializeMultisig", Info: map[string]any{
				"multisig": accounts[0], "rentSysvar": accounts[1],
				"signers": accounts[2:], "m": m,
			}}, nil
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializeMultisig2", Info: map[string]any{
			"multisig": accounts[0], "signers": accounts[1:], "m": m,
		}}, nil

	case 3: // Transfer (deprecated but ubiquitous)
		amount, err := tokenU64(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"source": accounts[0], "destination": accounts[1], "amount": fmt.Sprintf("%d", amount),
		}
		parseSigners(info, 2, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: "transfer", Info: info}, nil

	case 4: // Approve
		amount, err := tokenU64(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"source": accounts[0], "delegate": accounts[1], "amount": fmt.Sprintf("%d", amount),
		}
		parseSigners(info, 2, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "approve", Info: info}, nil

	case 5: // Revoke
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{"source": accounts[0]}
		parseSigners(info, 1, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "revoke", Info: info}, nil

	case 6: // SetAuthority
		if len(rest) < 1 {
			return typeInfo{}, fmt.Errorf("%w: truncated setAuthority", errTokenUnpack)
		}
		authorityType, ok := authorityTypeName(rest[0])
		if !ok {
			return typeInfo{}, fmt.Errorf("%w: authority type %d", errTokenUnpack, rest[0])
		}
		newAuthority, _, err := tokenPubkeyOption(rest[1:])
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		// Which key the first account names depends on the authority class.
		owned := "mint"
		if rest[0] == 2 || rest[0] == 3 { // AccountOwner | CloseAccount
			owned = "account"
		}
		info := map[string]any{
			owned:           accounts[0],
			"authorityType": authorityType,
			// Present-and-null when unset — unlike freezeAuthority above.
			"newAuthority": newAuthority,
		}
		parseSigners(info, 1, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: "setAuthority", Info: info}, nil

	case 7: // MintTo
		amount, err := tokenU64(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint": accounts[0], "account": accounts[1], "amount": fmt.Sprintf("%d", amount),
		}
		parseSigners(info, 2, accounts, "mintAuthority", "multisigMintAuthority")
		return typeInfo{Type: "mintTo", Info: info}, nil

	case 8: // Burn
		amount, err := tokenU64(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"account": accounts[0], "mint": accounts[1], "amount": fmt.Sprintf("%d", amount),
		}
		parseSigners(info, 2, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: "burn", Info: info}, nil

	case 9: // CloseAccount
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"account": accounts[0], "destination": accounts[1],
		}
		parseSigners(info, 2, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "closeAccount", Info: info}, nil

	case 10, 11: // FreezeAccount / ThawAccount
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"account": accounts[0], "mint": accounts[1],
		}
		parseSigners(info, 2, accounts, "freezeAuthority", "multisigFreezeAuthority")
		t := "freezeAccount"
		if disc == 11 {
			t = "thawAccount"
		}
		return typeInfo{Type: t, Info: info}, nil

	case 12: // TransferChecked
		amount, decimals, err := tokenU64Decimals(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"source": accounts[0], "mint": accounts[1], "destination": accounts[2],
			"tokenAmount": tokenAmount(amount, decimals),
		}
		parseSigners(info, 3, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: "transferChecked", Info: info}, nil

	case 13: // ApproveChecked
		amount, decimals, err := tokenU64Decimals(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 4); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"source": accounts[0], "mint": accounts[1], "delegate": accounts[2],
			"tokenAmount": tokenAmount(amount, decimals),
		}
		parseSigners(info, 3, accounts, "owner", "multisigOwner")
		return typeInfo{Type: "approveChecked", Info: info}, nil

	case 14: // MintToChecked
		amount, decimals, err := tokenU64Decimals(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"mint": accounts[0], "account": accounts[1],
			"tokenAmount": tokenAmount(amount, decimals),
		}
		parseSigners(info, 2, accounts, "mintAuthority", "multisigMintAuthority")
		return typeInfo{Type: "mintToChecked", Info: info}, nil

	case 15: // BurnChecked
		amount, decimals, err := tokenU64Decimals(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 3); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{
			"account": accounts[0], "mint": accounts[1],
			"tokenAmount": tokenAmount(amount, decimals),
		}
		parseSigners(info, 2, accounts, "authority", "multisigAuthority")
		return typeInfo{Type: "burnChecked", Info: info}, nil

	case 16, 18: // InitializeAccount2 / InitializeAccount3
		owner, _, err := tokenPubkey(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if disc == 16 {
			if err := checkNumAccounts(accounts, 3); err != nil {
				return typeInfo{}, err
			}
			return typeInfo{Type: "initializeAccount2", Info: map[string]any{
				"account": accounts[0], "mint": accounts[1],
				"owner": owner, "rentSysvar": accounts[2],
			}}, nil
		}
		if err := checkNumAccounts(accounts, 2); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializeAccount3", Info: map[string]any{
			"account": accounts[0], "mint": accounts[1], "owner": owner,
		}}, nil

	case 17: // SyncNative
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "syncNative", Info: map[string]any{
			"account": accounts[0],
		}}, nil

	case 21: // GetAccountDataSize
		// The extension list decodes at unpack level in Agave: u16 LE chunks,
		// each a known ExtensionType — a trailing odd byte or unknown value
		// fails unpack. The key is omitted when the list is empty (classic
		// program instructions always are).
		if len(rest)%2 != 0 {
			return typeInfo{}, fmt.Errorf("%w: odd extension-type bytes", errTokenUnpack)
		}
		var extensions []string
		for i := 0; i < len(rest); i += 2 {
			name, ok := uiExtensionTypeName(binary.LittleEndian.Uint16(rest[i:]))
			if !ok {
				return typeInfo{}, fmt.Errorf("%w: extension type %d", errTokenUnpack, binary.LittleEndian.Uint16(rest[i:]))
			}
			extensions = append(extensions, name)
		}
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		info := map[string]any{"mint": accounts[0]}
		if len(extensions) > 0 {
			info["extensionTypes"] = extensions
		}
		return typeInfo{Type: "getAccountDataSize", Info: info}, nil

	case 22: // InitializeImmutableOwner
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "initializeImmutableOwner", Info: map[string]any{
			"account": accounts[0],
		}}, nil

	case 23: // AmountToUiAmount
		amount, err := tokenU64(rest)
		if err != nil {
			return typeInfo{}, err
		}
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "amountToUiAmount", Info: map[string]any{
			"mint": accounts[0], "amount": fmt.Sprintf("%d", amount),
		}}, nil

	case 24: // UiAmountToAmount
		if !utf8.Valid(rest) {
			return typeInfo{}, fmt.Errorf("%w: uiAmount is not utf8", errTokenUnpack)
		}
		if err := checkNumAccounts(accounts, 1); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "uiAmountToAmount", Info: map[string]any{
			"mint": accounts[0], "uiAmount": string(rest),
		}}, nil
	}
	// Discriminants past the shared base set: token-2022 extension families.
	return parseTokenExtension(disc, data, accounts)
}

// parseSigners is Agave's multisig rule: with accounts beyond the authority
// position, the authority is a multisig and the remainder are its signers;
// otherwise the single-owner field is used. Positions, not signature flags.
func parseSigners(info map[string]any, lastNonsignerIndex int, accounts []string, ownerField, multisigField string) {
	if len(accounts) > lastNonsignerIndex+1 {
		info[multisigField] = accounts[lastNonsignerIndex]
		info["signers"] = accounts[lastNonsignerIndex+1:]
	} else {
		info[ownerField] = accounts[lastNonsignerIndex]
	}
}

// tokenAmount is token_amount_to_ui_amount_v3's plain-decimals path.
//
// DELIBERATELY different from the meta token-balance rule: instruction
// tokenAmounts render uiAmount 0 for a zero amount (Some(0.0) in Agave),
// while balances render null — two different Agave helpers, both pinned by
// the differential tests. Do not "unify" them.
func tokenAmount(amount uint64, decimals uint8) map[string]any {
	return map[string]any{
		"uiAmount":       float64(amount) / pow10(decimals),
		"decimals":       decimals,
		"amount":         fmt.Sprintf("%d", amount),
		"uiAmountString": trimmedDecimalString(amount, decimals),
	}
}

func pow10(d uint8) float64 {
	v := 1.0
	for i := uint8(0); i < d; i++ {
		v *= 10
	}
	return v
}

// trimmedDecimalString is Agave's real_number_string_trimmed: exact decimal
// via string math (never through a float), trailing zeros then a trailing
// point trimmed.
func trimmedDecimalString(amount uint64, decimals uint8) string {
	s := fmt.Sprintf("%d", amount)
	if decimals == 0 {
		return s
	}
	d := int(decimals)
	for len(s) <= d {
		s = "0" + s
	}
	intPart, fracPart := s[:len(s)-d], s[len(s)-d:]
	fracPart = strings.TrimRight(fracPart, "0")
	if fracPart == "" {
		return intPart
	}
	return intPart + "." + fracPart
}

// authorityTypeName maps the wire byte to Agave's UiAuthorityType camelCase
// name. The full table is cheap and layout-free, so phase 1 carries all of
// it even though only 0..=3 occur on the classic program.
func authorityTypeName(b byte) (string, bool) {
	names := []string{
		"mintTokens", "freezeAccount", "accountOwner", "closeAccount",
		"transferFeeConfig", "withheldWithdraw", "closeMint", "interestRate",
		"permanentDelegate", "confidentialTransferMint", "transferHookProgramId",
		"confidentialTransferFeeConfig", "metadataPointer", "groupPointer",
		"groupMemberPointer", "scaledUiAmount", "pause",
	}
	if int(b) < len(names) {
		return names[int(b)], true
	}
	return "", false
}

func tokenU64(b []byte) (uint64, error) {
	if len(b) < 8 {
		return 0, fmt.Errorf("%w: truncated u64", errTokenUnpack)
	}
	return binary.LittleEndian.Uint64(b[:8]), nil
}

func tokenU64Decimals(b []byte) (uint64, uint8, error) {
	amount, err := tokenU64(b)
	if err != nil {
		return 0, 0, err
	}
	if len(b) < 9 {
		return 0, 0, fmt.Errorf("%w: truncated decimals", errTokenUnpack)
	}
	return amount, b[8], nil
}

func tokenPubkey(b []byte) (string, []byte, error) {
	if len(b) < 32 {
		return "", nil, fmt.Errorf("%w: truncated pubkey", errTokenUnpack)
	}
	return svm.Base58Encode(b[:32]), b[32:], nil
}

// tokenPubkeyOption is the token program's COption<Pubkey>: a one-byte tag,
// then 32 bytes when the tag is 1. Any other tag is malformed.
func tokenPubkeyOption(b []byte) (*string, []byte, error) {
	if len(b) < 1 {
		return nil, nil, fmt.Errorf("%w: truncated COption tag", errTokenUnpack)
	}
	switch b[0] {
	case 0:
		return nil, b[1:], nil
	case 1:
		s, rest, err := tokenPubkey(b[1:])
		if err != nil {
			return nil, nil, err
		}
		return &s, rest, nil
	}
	return nil, nil, fmt.Errorf("%w: COption tag %d", errTokenUnpack, b[0])
}
