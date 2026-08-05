package parse_test

// Layer 3: AttachToBlock/AttachToTransaction semantics on a hand-built block —
// Parsed set for parsable sites, nil for unparsable ones, indexes resolved
// against static ++ loadedWritable ++ loadedReadonly, malformed indexes
// degrade without panicking, and re-attach clears stale values.

import (
	"encoding/json"
	"testing"

	"github.com/blockchain-data-standards/manifesto/svm"
	"github.com/blockchain-data-standards/manifesto/svm/parse"
)

func mustB58(t *testing.T, s string) []byte {
	t.Helper()
	b, err := svm.Base58Decode(s)
	if err != nil {
		t.Fatalf("base58 %q: %v", s, err)
	}
	return b
}

func TestAttachSemantics(t *testing.T) {
	payer, dest, extra := kb(0x11), kb(0x22), kb(0x33)
	sysProg := mustB58(t, parse.SystemProgramID)
	memoProg := mustB58(t, parse.MemoV3ID)

	// Merged key list the attacher must resolve against:
	//   0 payer, 1 dest, 2 system (static)
	//   3 extra           (loaded writable)
	//   4 memo program    (loaded readonly)
	transferIx := &svm.CompiledInstruction{
		ProgramIdIndex: 2,
		Accounts:       []byte{0, 1},
		Data:           cat(le32(2), le64(99)),
	}
	unknownProgIx := &svm.CompiledInstruction{
		ProgramIdIndex: 0, // payer is no program this package knows
		Accounts:       []byte{1},
		Data:           []byte{1, 2, 3},
	}
	badAccountIx := &svm.CompiledInstruction{
		ProgramIdIndex: 2,
		Accounts:       []byte{0, 200}, // 200 is outside the merged list
		Data:           cat(le32(2), le64(1)),
		Parsed:         []byte(`{"stale":true}`), // must be cleared, not kept
	}
	badProgramIdxIx := &svm.CompiledInstruction{
		ProgramIdIndex: 250, // outside the merged list
		Data:           cat(le32(2), le64(1)),
	}
	memoIx := &svm.CompiledInstruction{
		ProgramIdIndex: 4, // resolves through loadedReadonly
		Data:           []byte("hi"),
		StackHeight:    u32p(2),
	}

	block := &svm.ConfirmedBlock{
		Transactions: []*svm.ConfirmedTransaction{{
			Transaction: &svm.Transaction{
				Message: &svm.Message{
					AccountKeys:  [][]byte{payer, dest, sysProg},
					Instructions: []*svm.CompiledInstruction{transferIx, unknownProgIx, badAccountIx, badProgramIdxIx},
				},
			},
			Meta: &svm.TransactionStatusMeta{
				LoadedWritableAddresses: [][]byte{extra},
				LoadedReadonlyAddresses: [][]byte{memoProg},
				InnerInstructions: []*svm.InnerInstructions{
					{Index: 0, Instructions: []*svm.CompiledInstruction{memoIx}},
				},
			},
		}},
	}

	parse.AttachToBlock(block)

	// The parsable top-level instruction gets exactly Parse's envelope.
	if transferIx.Parsed == nil {
		t.Fatal("transfer instruction: Parsed is nil, want envelope")
	}
	env, err := parse.Parse(parse.SystemProgramID, transferIx.Data,
		[]string{svm.Base58Encode(payer), svm.Base58Encode(dest)}, nil)
	if err != nil {
		t.Fatalf("reference Parse: %v", err)
	}
	wantRaw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal reference envelope: %v", err)
	}
	if path, diff, ok := firstDiff("", decodeJSON(t, wantRaw), decodeJSON(t, transferIx.Parsed)); !ok {
		t.Errorf("attached envelope mismatch at %q: %s\nwant: %s\ngot:  %s",
			path, diff, wantRaw, transferIx.Parsed)
	}
	// Top-level: stackHeight key present and explicitly null.
	obj := decodeJSON(t, transferIx.Parsed).(map[string]any)
	if v, present := obj["stackHeight"]; !present || v != nil {
		t.Errorf("top-level stackHeight = %v (present=%v), want explicit null", v, present)
	}

	// Unparsable sites: nil Parsed, no panic — including the stale value.
	if unknownProgIx.Parsed != nil {
		t.Errorf("unknown-program instruction: Parsed = %s, want nil", unknownProgIx.Parsed)
	}
	if badAccountIx.Parsed != nil {
		t.Errorf("out-of-range account index: Parsed = %s, want nil (stale value cleared)", badAccountIx.Parsed)
	}
	if badProgramIdxIx.Parsed != nil {
		t.Errorf("out-of-range programIdIndex: Parsed = %s, want nil", badProgramIdxIx.Parsed)
	}

	// The inner memo resolves via loadedReadonly and keeps its stackHeight.
	if memoIx.Parsed == nil {
		t.Fatal("inner memo instruction: Parsed is nil, want envelope")
	}
	inner := decodeJSON(t, memoIx.Parsed).(map[string]any)
	if inner["program"] != "spl-memo" || inner["programId"] != parse.MemoV3ID {
		t.Errorf("inner envelope program/programId = %v/%v", inner["program"], inner["programId"])
	}
	if inner["parsed"] != "hi" {
		t.Errorf("memo parsed = %v, want the bare string \"hi\"", inner["parsed"])
	}
	if n, ok := inner["stackHeight"].(json.Number); !ok || n.String() != "2" {
		t.Errorf("inner stackHeight = %v, want 2", inner["stackHeight"])
	}

	// Idempotency: mutate and re-attach. The formerly-parsable instruction
	// must LOSE its envelope (stale value cleared), and a formerly-unparsable
	// one must gain its own.
	transferIx.Data = le32(99) // unknown system discriminant now
	unknownProgIx.ProgramIdIndex = 2
	unknownProgIx.Accounts = []byte{0, 1}
	unknownProgIx.Data = cat(le32(2), le64(7))
	parse.AttachToBlock(block)

	if transferIx.Parsed != nil {
		t.Errorf("re-attach kept a stale envelope: %s", transferIx.Parsed)
	}
	if unknownProgIx.Parsed == nil {
		t.Error("re-attach did not parse the now-valid instruction")
	}
	if memoIx.Parsed == nil {
		t.Error("re-attach dropped the still-valid memo envelope")
	}
}

// Nil and hollow shapes must be no-ops, never panics.
func TestAttachNilSafety(t *testing.T) {
	parse.AttachToBlock(nil)
	parse.AttachToBlock(&svm.ConfirmedBlock{Transactions: []*svm.ConfirmedTransaction{nil}})
	parse.AttachToTransaction(nil)
	parse.AttachToTransaction(&svm.ConfirmedTransaction{})
	parse.AttachToTransaction(&svm.ConfirmedTransaction{Transaction: &svm.Transaction{}})
	parse.AttachToTransaction(&svm.ConfirmedTransaction{
		Transaction: &svm.Transaction{Message: &svm.Message{
			Instructions: []*svm.CompiledInstruction{nil},
		}},
		Meta: &svm.TransactionStatusMeta{
			InnerInstructions: []*svm.InnerInstructions{nil},
		},
	})
}
