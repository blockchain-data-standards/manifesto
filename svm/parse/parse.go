// Package parse renders Solana instructions into Agave's `jsonParsed` form.
//
// This is a mechanical Go port of Agave's own parsers — solana-transaction-status
// 2.3.13, the exact code a validator runs to answer `encoding: "jsonParsed"` —
// for the COMPLETE registry that version parses:
//
//	system                        all 13 instruction types
//	spl-token                     the shared base set (discriminants 0..=24)
//	                              plus every token-2022 extension family
//	                              (transfer fee, confidential transfer/fee/
//	                              mint-burn, default account state, memo
//	                              transfer, interest bearing, CPI guard,
//	                              permanent delegate, transfer hook, metadata/
//	                              group/group-member pointers, scaled ui
//	                              amount, pausable, reallocate, native mint,
//	                              non-transferable, withdraw excess lamports)
//	                              and the TokenMetadata/TokenGroup TLV
//	                              interface instructions, under BOTH deployed
//	                              token program ids
//	spl-associated-token-account  create, createIdempotent, recoverNested
//	spl-memo                      v1 and v3
//	vote                          every VoteInstruction variant
//	stake                         every StakeInstruction variant
//	address-lookup-table          all five ProgramInstruction variants
//	bpf-loader                    write, finalize
//	bpf-upgradeable-loader        every UpgradeableLoaderInstruction variant
//
// Programs outside Agave's registry (compute-budget, arbitrary user
// programs) return ErrNotParsable. That is NOT a gap in the output contract:
// the caller renders those instructions in Agave's partiallyDecoded form —
// exactly what a real node emits for them. The same discipline covers a
// FUTURE Agave adding instruction types this pin predates: unknown
// discriminants degrade to the fallback; they never corrupt.
//
// THE PORT RULE, for every future edit: shapes are copied from the pinned
// Agave source, never inferred from documentation or memory. Key-presence
// semantics are load-bearing and differ per field — `freezeAuthority` is
// OMITTED when absent, `newAuthority` is PRESENT-and-null, `extensionTypes`
// is omitted when empty. The differential test against a live node's output
// (see the golden fixture) is the drift tripwire: when Agave adds an
// instruction type, that test is what turns "silent divergence" into a red
// build. Fallback discipline: on ANY anomaly — short data, unknown
// discriminant, missing accounts, invalid utf8 — return an error and let the
// caller fall back. Never emit a partial or guessed object.
package parse

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Program ids, base58. The two memo ids and the two token ids are distinct
// deployed programs that share a parser, exactly as in Agave's registry.
const (
	SystemProgramID = "11111111111111111111111111111111"
	TokenProgramID  = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	Token2022ID     = "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"
	AssociatedID    = "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL"
	MemoV1ID        = "Memo1UhkJRfHyvLMcVucJwxXeuD728EqVDDwQDxFMNo"
	MemoV3ID        = "MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr"
	VoteProgramID   = "Vote111111111111111111111111111111111111111"
	StakeProgramID  = "Stake11111111111111111111111111111111111111"
	// LookupTableID is the address-lookup-table program.
	LookupTableID = "AddressLookupTab1e1111111111111111111111111"
	// BpfLoaderID is loader v2 — the only non-upgradeable loader in Agave's
	// registry (the deprecated v1 loader is NOT parsed by Agave).
	BpfLoaderID            = "BPFLoader2111111111111111111111111111111111"
	BpfUpgradeableLoaderID = "BPFLoaderUpgradeab1e11111111111111111111111"
)

// ErrNotParsable reports that this package has no parser for the program, or
// that the instruction's bytes do not decode as any instruction the parser
// knows. The caller's correct response is Agave's partiallyDecoded rendering.
var ErrNotParsable = errors.New("svm/parse: instruction not parsable")

// Instruction is Agave's ParsedInstruction: the object a node places at a
// jsonParsed instruction site. `Parsed` is `{"type": ..., "info": {...}}` for
// every program except memo, whose parsed form is a bare JSON string.
type Instruction struct {
	Program     string          `json:"program"`
	ProgramID   string          `json:"programId"`
	Parsed      json.RawMessage `json:"parsed"`
	StackHeight *uint32         `json:"stackHeight"`
}

// typeInfo mirrors Agave's ParsedInstructionEnum: `info` is omitted when the
// instruction carries none (Agave: skip_serializing_if Value::is_null).
type typeInfo struct {
	Type string         `json:"type"`
	Info map[string]any `json:"info,omitempty"`
}

// Parse renders one instruction, or ErrNotParsable when the caller must fall
// back to partiallyDecoded.
//
// `accounts` is the instruction's account list ALREADY resolved to base58
// pubkeys, in instruction order — i.e. accounts[i] is the key at the
// instruction's i-th account index, resolved against the merged
// (static ++ loadedWritable ++ loadedReadonly) key list.
func Parse(programID string, data []byte, accounts []string, stackHeight *uint32) (*Instruction, error) {
	var (
		program string
		parsed  any
		err     error
	)
	// Every Agave parser except memo opens with a key-mismatch guard over
	// instruction.accounts.iter().max(); an EMPTY account list is None and
	// errors before any arm runs — even for arms that need no accounts
	// (stake getMinimumDelegation). Memo has no guard: it never reads
	// accounts. Index range itself is the caller's concern (accounts arrive
	// pre-resolved), so emptiness is the only reachable half of the guard.
	switch programID {
	case MemoV1ID, MemoV3ID:
	default:
		if len(accounts) == 0 {
			return nil, fmt.Errorf("%w: no accounts", ErrNotParsable)
		}
	}
	switch programID {
	case SystemProgramID:
		program = "system"
		parsed, err = parseSystem(data, accounts)
	case TokenProgramID, Token2022ID:
		program = "spl-token"
		parsed, err = parseToken(data, accounts)
		// Agave tries the TokenGroup then TokenMetadata TLV interfaces only
		// when TokenInstruction::unpack itself failed — never when a decoded
		// arm errs (e.g. on account counts).
		if err != nil && errors.Is(err, errTokenUnpack) {
			parsed, err = parseTokenInterface(data, accounts)
		}
	case AssociatedID:
		program = "spl-associated-token-account"
		parsed, err = parseAssociatedToken(data, accounts)
	case MemoV1ID, MemoV3ID:
		program = "spl-memo"
		parsed, err = parseMemo(data)
	case VoteProgramID:
		program = "vote"
		parsed, err = parseVote(data, accounts)
	case StakeProgramID:
		program = "stake"
		parsed, err = parseStake(data, accounts)
	case LookupTableID:
		program = "address-lookup-table"
		parsed, err = parseAddressLookupTable(data, accounts)
	case BpfLoaderID:
		program = "bpf-loader"
		parsed, err = parseBpfLoader(data, accounts)
	case BpfUpgradeableLoaderID:
		program = "bpf-upgradeable-loader"
		parsed, err = parseBpfUpgradeableLoader(data, accounts)
	default:
		return nil, fmt.Errorf("%w: program %s", ErrNotParsable, programID)
	}
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(parsed)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotParsable, err)
	}
	return &Instruction{
		Program:     program,
		ProgramID:   programID,
		Parsed:      raw,
		StackHeight: stackHeight,
	}, nil
}

// checkNumAccounts is Agave's check_num_accounts: a minimum, not an exact
// count — several arms read optional trailing signer accounts beyond it.
func checkNumAccounts(accounts []string, num int) error {
	if len(accounts) < num {
		return fmt.Errorf("%w: expected at least %d accounts, got %d", ErrNotParsable, num, len(accounts))
	}
	return nil
}
