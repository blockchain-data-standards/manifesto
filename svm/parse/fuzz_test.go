package parse_test

// Layer 4: adversarial nets.
//
//   - Prefix truncation sweeps: every strict prefix of a representative wire
//     per family must refuse cleanly (nil envelope + ErrNotParsable) and the
//     full wire must parse — no prefix may panic or emit a partial object.
//     The representatives are MINIMAL encodings (no trailing junk), so even
//     under the bincode/token trailing-bytes tolerance no strict prefix can
//     form a complete instruction: erroring until the exact full length is
//     the correct expectation for these specific wires.
//   - A randomized no-panic net over every registry program id with a
//     deterministic PCG stream: Parse must never panic and must always
//     return envelope-XOR-error, the error always wrapping ErrNotParsable.
//   - AttachToBlock over the golden block under random byte flips: never
//     panics, and every instruction ends with Parsed either nil (exactly
//     when Parse refuses) or Parse's own envelope.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/blockchain-data-standards/manifesto/svm"
	"github.com/blockchain-data-standards/manifesto/svm/parse"
)

func TestPrefixTruncationSweeps(t *testing.T) {
	sweeps := []struct {
		name      string
		programID string
		wire      []byte
		accounts  []string
	}{
		{
			name:      "system createAccountWithSeed",
			programID: parse.SystemProgramID,
			wire:      cat(le32(3), kb(7), bincodeStr("seed"), le64(1), le64(2), kb(8)),
			accounts:  accs(1, 2),
		},
		{
			name:      "vote towersync",
			programID: parse.VoteProgramID,
			wire: cat(le32(14), le64(50), shortU16(1), varint64(10), []byte{3},
				kb(0x4A), optSomeI64(777), kb(0x4B)),
			accounts: accs(1, 2),
		},
		{
			name:      "stake authorizeWithSeed",
			programID: parse.StakeProgramID,
			wire:      cat(le32(8), kb(0x66), le32(1), bincodeStr("marinade"), kb(0x67)),
			accounts:  accs(1, 2, 3, 4),
		},
		{
			name:      "alt extendLookupTable",
			programID: parse.LookupTableID,
			wire:      cat(le32(2), le64(2), kb(0x71), kb(0x72)),
			accounts:  accs(1, 2, 3, 4),
		},
		{
			name:      "loader v3 write",
			programID: parse.BpfUpgradeableLoaderID,
			wire:      cat(le32(1), le32(8), le64(4), []byte{0xDE, 0xAD, 0xBE, 0xEF}),
			accounts:  accs(1, 2),
		},
		{
			name:      "token transferChecked",
			programID: parse.TokenProgramID,
			wire:      cat([]byte{12}, le64(5), []byte{2}),
			accounts:  accs(1, 2, 3, 4),
		},
		{
			// The longest confidential-transfer form: 168-byte sub payload.
			name:      "token22 confidentialTransfer 168B",
			programID: parse.Token2022ID,
			wire: cat([]byte{27, 7}, rep(36, 0x8A), rep(64, 0x8B), rep(64, 0x8C),
				[]byte{1, 2, 3}),
			accounts: accs(1, 2, 3, 4, 5, 6, 7, 8),
		},
		{
			name:      "tlv initializeTokenMetadata",
			programID: parse.Token2022ID,
			wire:      cat(tlvInitMeta, borshStr("N"), borshStr("S"), borshStr("U")),
			accounts:  accs(1, 2, 3, 4),
		},
	}
	for _, sw := range sweeps {
		full, err := parse.Parse(sw.programID, sw.wire, sw.accounts, nil)
		if err != nil {
			t.Errorf("%s: full wire failed to parse: %v", sw.name, err)
			continue
		}
		for i := range sw.wire {
			env, err := parse.Parse(sw.programID, sw.wire[:i], sw.accounts, nil)
			if err == nil {
				t.Errorf("%s: prefix %d/%d parsed: %s (full: %s)",
					sw.name, i, len(sw.wire), env.Parsed, full.Parsed)
				continue
			}
			if env != nil {
				t.Errorf("%s: prefix %d: non-nil envelope alongside error", sw.name, i)
			}
			if !errors.Is(err, parse.ErrNotParsable) {
				t.Errorf("%s: prefix %d: error %v does not wrap ErrNotParsable", sw.name, i, err)
			}
		}
	}
}

// registryIDs is every program id Parse dispatches on.
var registryIDs = []string{
	parse.SystemProgramID, parse.TokenProgramID, parse.Token2022ID,
	parse.AssociatedID, parse.MemoV1ID, parse.MemoV3ID,
	parse.VoteProgramID, parse.StakeProgramID, parse.LookupTableID,
	parse.BpfLoaderID, parse.BpfUpgradeableLoaderID,
}

func TestRandomizedNoPanic(t *testing.T) {
	rng := rand.New(rand.NewPCG(0xC0FFEE, 0xBD5EED)) // deterministic
	keyPool := make([]string, 16)
	for i := range keyPool {
		keyPool[i] = k(byte(i + 1))
	}
	const iterations = 50_000
	for i := range iterations {
		programID := registryIDs[rng.IntN(len(registryIDs))]
		data := make([]byte, rng.IntN(65))
		for j := range data {
			data[j] = byte(rng.IntN(256))
		}
		// Half the stream is discriminant-biased so deep arms get hit:
		// bincode programs take a small u32, token programs a small u8.
		if rng.IntN(2) == 0 {
			if len(data) >= 4 {
				copy(data, le32(uint32(rng.IntN(20))))
			}
			if len(data) >= 1 && (programID == parse.TokenProgramID || programID == parse.Token2022ID) {
				data[0] = byte(rng.IntN(48))
			}
		}
		accounts := make([]string, rng.IntN(13))
		for j := range accounts {
			accounts[j] = keyPool[rng.IntN(len(keyPool))]
		}

		env, err := parse.Parse(programID, data, accounts, nil)
		if (env == nil) == (err == nil) {
			t.Fatalf("iter %d: envelope XOR error violated (env=%v err=%v)", i, env, err)
		}
		if err != nil && !errors.Is(err, parse.ErrNotParsable) {
			t.Fatalf("iter %d: error %v does not wrap ErrNotParsable", i, err)
		}
		if err == nil && !json.Valid(env.Parsed) {
			t.Fatalf("iter %d: envelope carries invalid JSON: %q", i, env.Parsed)
		}
	}
}

// blockFromFixture rebuilds an svm.ConfirmedBlock from the golden fixture's
// plain-JSON side, so attach can run against real mainnet shapes.
func blockFromFixture(t *testing.T, g *fxFixture) *svm.ConfirmedBlock {
	t.Helper()
	decodeKeys := func(keys []string) [][]byte {
		out := make([][]byte, len(keys))
		for i, s := range keys {
			out[i] = mustB58(t, s)
		}
		return out
	}
	toIx := func(ji *fxIx) *svm.CompiledInstruction {
		accounts := make([]byte, len(ji.Accounts))
		for i, idx := range ji.Accounts {
			if idx > 255 {
				t.Fatalf("account index %d exceeds a byte", idx)
			}
			accounts[i] = byte(idx)
		}
		data, err := svm.Base58Decode(ji.Data)
		if err != nil {
			t.Fatalf("fixture data %q: %v", ji.Data, err)
		}
		return &svm.CompiledInstruction{
			ProgramIdIndex: ji.ProgramIdIndex,
			Accounts:       accounts,
			Data:           data,
			StackHeight:    ji.StackHeight,
		}
	}

	block := &svm.ConfirmedBlock{}
	for ti := range g.Json.Transactions {
		jt := &g.Json.Transactions[ti]
		msg := &svm.Message{AccountKeys: decodeKeys(jt.Transaction.Message.AccountKeys)}
		for i := range jt.Transaction.Message.Instructions {
			msg.Instructions = append(msg.Instructions, toIx(&jt.Transaction.Message.Instructions[i]))
		}
		tx := &svm.ConfirmedTransaction{Transaction: &svm.Transaction{Message: msg}}
		if jt.Meta != nil {
			meta := &svm.TransactionStatusMeta{}
			if jt.Meta.LoadedAddresses != nil {
				meta.LoadedWritableAddresses = decodeKeys(jt.Meta.LoadedAddresses.Writable)
				meta.LoadedReadonlyAddresses = decodeKeys(jt.Meta.LoadedAddresses.Readonly)
			}
			for gi := range jt.Meta.InnerInstructions {
				grp := &jt.Meta.InnerInstructions[gi]
				inner := &svm.InnerInstructions{Index: grp.Index}
				for i := range grp.Instructions {
					inner.Instructions = append(inner.Instructions, toIx(&grp.Instructions[i]))
				}
				meta.InnerInstructions = append(meta.InnerInstructions, inner)
			}
			tx.Meta = meta
		}
		block.Transactions = append(block.Transactions, tx)
	}
	return block
}

// checkAttachInvariant re-derives every instruction's parse result and
// demands AttachToBlock left exactly that: Parse's envelope where it
// succeeds, nil where it refuses (or where an index is out of range).
func checkAttachInvariant(t *testing.T, block *svm.ConfirmedBlock, round int) {
	t.Helper()
	for ti, tx := range block.Transactions {
		var keys []string
		for _, raw := range tx.Transaction.Message.AccountKeys {
			keys = append(keys, svm.Base58Encode(raw))
		}
		if tx.Meta != nil {
			for _, raw := range tx.Meta.LoadedWritableAddresses {
				keys = append(keys, svm.Base58Encode(raw))
			}
			for _, raw := range tx.Meta.LoadedReadonlyAddresses {
				keys = append(keys, svm.Base58Encode(raw))
			}
		}
		check := func(where string, ci *svm.CompiledInstruction) {
			var want []byte
			if int(ci.ProgramIdIndex) < len(keys) {
				accounts := make([]string, 0, len(ci.Accounts))
				ok := true
				for _, idx := range ci.Accounts {
					if int(idx) >= len(keys) {
						ok = false
						break
					}
					accounts = append(accounts, keys[idx])
				}
				if ok {
					if env, err := parse.Parse(keys[ci.ProgramIdIndex], ci.Data, accounts, ci.StackHeight); err == nil {
						raw, merr := json.Marshal(env)
						if merr != nil {
							t.Fatalf("round %d tx %d %s: marshal: %v", round, ti, where, merr)
						}
						want = raw
					}
				}
			}
			if !bytes.Equal(ci.Parsed, want) {
				t.Errorf("round %d tx %d %s: Parsed = %s, want %s",
					round, ti, where, ci.Parsed, want)
			}
		}
		for i, ci := range tx.Transaction.Message.Instructions {
			check(fmt.Sprintf("top[%d]", i), ci)
		}
		if tx.Meta != nil {
			for _, inner := range tx.Meta.InnerInstructions {
				for i, ci := range inner.Instructions {
					check(fmt.Sprintf("inner[%d]", i), ci)
				}
			}
		}
	}
}

func TestAttachToBlockMutatedGolden(t *testing.T) {
	g := loadFixture(t)
	block := blockFromFixture(t, g)

	// Baseline: the untouched block attaches cleanly and matches Parse.
	parse.AttachToBlock(block)
	checkAttachInvariant(t, block, -1)

	// Flatten every instruction for mutation targeting.
	var all []*svm.CompiledInstruction
	for _, tx := range block.Transactions {
		all = append(all, tx.Transaction.Message.Instructions...)
		if tx.Meta != nil {
			for _, inner := range tx.Meta.InnerInstructions {
				all = append(all, inner.Instructions...)
			}
		}
	}

	rng := rand.New(rand.NewPCG(42, 4242)) // deterministic
	const rounds = 40
	for round := range rounds {
		// A burst of random byte flips (and occasional truncations) in
		// instruction data — never in the key lists, mirroring the failure
		// mode attach must degrade on: corrupt instruction payloads.
		for range 8 {
			ci := all[rng.IntN(len(all))]
			if len(ci.Data) == 0 {
				continue
			}
			if rng.IntN(8) == 0 {
				ci.Data = ci.Data[:rng.IntN(len(ci.Data))]
				continue
			}
			ci.Data[rng.IntN(len(ci.Data))] ^= byte(1 + rng.IntN(255))
		}
		parse.AttachToBlock(block) // must never panic
		checkAttachInvariant(t, block, round)
	}
}
