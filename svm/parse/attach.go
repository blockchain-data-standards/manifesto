package parse

import (
	"encoding/json"

	"github.com/blockchain-data-standards/manifesto/svm"
)

// AttachToBlock walks every instruction of every transaction — top-level and
// inner — and populates CompiledInstruction.Parsed with the serialized
// ParsedInstruction envelope for each instruction this package can render.
// Instructions it cannot render are left untouched: the JSON layer already
// renders those partiallyDecoded, so attach failures degrade, never corrupt.
//
// Account indexes resolve against the merged key list — static keys, then
// loaded writable, then loaded readonly — the same order the runtime (and
// Agave's jsonParsed renderer) uses for v0 transactions. An index outside
// the merged list marks the instruction unparsable; it does not panic.
//
// Idempotent: re-attaching overwrites prior Parsed values.
func AttachToBlock(block *svm.ConfirmedBlock) {
	if block == nil {
		return
	}
	for _, tx := range block.Transactions {
		AttachToTransaction(tx)
	}
}

// AttachToTransaction attaches parsed forms within a single transaction.
func AttachToTransaction(tx *svm.ConfirmedTransaction) {
	if tx == nil || tx.Transaction == nil || tx.Transaction.Message == nil {
		return
	}
	msg := tx.Transaction.Message

	// Merged key list, base58-encoded once per transaction.
	n := len(msg.AccountKeys)
	if tx.Meta != nil {
		n += len(tx.Meta.LoadedWritableAddresses) + len(tx.Meta.LoadedReadonlyAddresses)
	}
	keys := make([]string, 0, n)
	for _, k := range msg.AccountKeys {
		keys = append(keys, svm.Base58Encode(k))
	}
	if tx.Meta != nil {
		for _, k := range tx.Meta.LoadedWritableAddresses {
			keys = append(keys, svm.Base58Encode(k))
		}
		for _, k := range tx.Meta.LoadedReadonlyAddresses {
			keys = append(keys, svm.Base58Encode(k))
		}
	}

	for _, ci := range msg.Instructions {
		attachOne(ci, keys)
	}
	if tx.Meta != nil {
		for _, inner := range tx.Meta.InnerInstructions {
			if inner == nil {
				continue
			}
			for _, ci := range inner.Instructions {
				attachOne(ci, keys)
			}
		}
	}
}

func attachOne(ci *svm.CompiledInstruction, keys []string) {
	if ci == nil {
		return
	}
	ci.Parsed = nil // idempotent: never leave a stale value behind
	if int(ci.ProgramIdIndex) >= len(keys) {
		return
	}
	programID := keys[ci.ProgramIdIndex]

	accounts := make([]string, len(ci.Accounts))
	for i, idx := range ci.Accounts {
		if int(idx) >= len(keys) {
			return // malformed index: leave unparsed, renderer falls back
		}
		accounts[i] = keys[idx]
	}

	parsed, err := Parse(programID, ci.Data, accounts, ci.StackHeight)
	if err != nil {
		return
	}
	raw, err := json.Marshal(parsed)
	if err != nil {
		return
	}
	ci.Parsed = raw
}
