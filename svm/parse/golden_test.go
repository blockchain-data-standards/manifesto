package parse_test

// The golden differential: walk the captured mainnet block's `json` and
// `jsonParsed` encodings in lockstep and require that parse.Parse reproduces,
// byte-for-semantic-byte, EVERY instruction the node itself rendered in
// parsed form — the package now ports Agave's complete jsonParsed registry,
// so a parsed site under any program is a fidelity obligation. Where the
// node fell back to partiallyDecoded, Parse must refuse: for a registry
// program that means the node's own parser rejected the bytes (malformed or
// foreign-loader data), and succeeding where Agave refused is divergence;
// for a non-registry program (ComputeBudget, arbitrary user programs) there
// is no parser at all. This is the drift tripwire named in the package doc:
// when Agave's renderer and this port disagree, this test is what turns
// silent divergence into a red build.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/blockchain-data-standards/manifesto/svm"
	"github.com/blockchain-data-standards/manifesto/svm/parse"
)

// registryProgramIDs is Agave's COMPLETE jsonParsed registry, by program id.
// A partiallyDecoded site under one of these ids means the node's own parser
// refused the instruction — Parse must refuse it too.
var registryProgramIDs = map[string]bool{
	parse.SystemProgramID:        true,
	parse.TokenProgramID:         true,
	parse.Token2022ID:            true,
	parse.AssociatedID:           true,
	parse.MemoV1ID:               true,
	parse.MemoV3ID:               true,
	parse.VoteProgramID:          true,
	parse.StakeProgramID:         true,
	parse.LookupTableID:          true,
	parse.BpfLoaderID:            true,
	parse.BpfUpgradeableLoaderID: true,
}

// ---------------------------------------------------------------------------
// Fixture shapes — just enough of each side to walk instructions in lockstep.
// ---------------------------------------------------------------------------

type fxFixture struct {
	Json struct {
		Transactions []fxJsonTx `json:"transactions"`
	} `json:"json"`
	JsonParsed struct {
		Transactions []fxParsedTx `json:"transactions"`
	} `json:"jsonParsed"`
}

type fxIx struct {
	ProgramIdIndex uint32   `json:"programIdIndex"`
	Accounts       []uint32 `json:"accounts"`
	Data           string   `json:"data"`
	StackHeight    *uint32  `json:"stackHeight"`
}

type fxJsonTx struct {
	Transaction struct {
		Message struct {
			AccountKeys  []string `json:"accountKeys"`
			Instructions []fxIx   `json:"instructions"`
		} `json:"message"`
	} `json:"transaction"`
	Meta *struct {
		InnerInstructions []struct {
			Index        uint32 `json:"index"`
			Instructions []fxIx `json:"instructions"`
		} `json:"innerInstructions"`
		LoadedAddresses *struct {
			Writable []string `json:"writable"`
			Readonly []string `json:"readonly"`
		} `json:"loadedAddresses"`
	} `json:"meta"`
}

type fxParsedTx struct {
	Transaction struct {
		Message struct {
			Instructions []json.RawMessage `json:"instructions"`
		} `json:"message"`
	} `json:"transaction"`
	Meta *struct {
		InnerInstructions []struct {
			Index        uint32            `json:"index"`
			Instructions []json.RawMessage `json:"instructions"`
		} `json:"innerInstructions"`
	} `json:"meta"`
}

func loadFixture(t *testing.T) *fxFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "testdata", "parsed_golden.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	var g fxFixture
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatalf("decoding fixture: %v", err)
	}
	if len(g.Json.Transactions) == 0 || len(g.Json.Transactions) != len(g.JsonParsed.Transactions) {
		t.Fatalf("fixture has %d json vs %d jsonParsed transactions",
			len(g.Json.Transactions), len(g.JsonParsed.Transactions))
	}
	return &g
}

// TestGoldenDifferential is layer 1: every site the node parsed — under ANY
// program — must deep-equal Parse's envelope; every partiallyDecoded site
// must return ErrNotParsable, whether its program is in the registry (the
// node's own parser refused those bytes) or outside it (no parser exists).
func TestGoldenDifferential(t *testing.T) {
	g := loadFixture(t)

	parsedSites, registryFallbacks, foreignFallbacks := 0, 0, 0
	parsedByProgram := map[string]int{}

	for ti := range g.Json.Transactions {
		jt, pt := &g.Json.Transactions[ti], &g.JsonParsed.Transactions[ti]

		// Merged key list exactly as the runtime resolves v0 indexes:
		// static keys, then loaded writable, then loaded readonly.
		merged := append([]string{}, jt.Transaction.Message.AccountKeys...)
		if jt.Meta != nil && jt.Meta.LoadedAddresses != nil {
			merged = append(merged, jt.Meta.LoadedAddresses.Writable...)
			merged = append(merged, jt.Meta.LoadedAddresses.Readonly...)
		}

		checkSite := func(where string, ji *fxIx, nodeRaw json.RawMessage) {
			node, ok := decodeJSON(t, nodeRaw).(map[string]any)
			if !ok {
				t.Fatalf("tx %d %s: jsonParsed instruction is not an object", ti, where)
			}

			if int(ji.ProgramIdIndex) >= len(merged) {
				t.Fatalf("tx %d %s: programIdIndex %d out of range", ti, where, ji.ProgramIdIndex)
			}
			programID := merged[ji.ProgramIdIndex]
			accounts := make([]string, len(ji.Accounts))
			for i, idx := range ji.Accounts {
				if int(idx) >= len(merged) {
					t.Fatalf("tx %d %s: account index %d out of range", ti, where, idx)
				}
				accounts[i] = merged[idx]
			}
			data, err := svm.Base58Decode(ji.Data)
			if err != nil {
				t.Fatalf("tx %d %s: data %q: %v", ti, where, ji.Data, err)
			}

			program, _ := node["program"].(string)
			_, nodeParsed := node["parsed"]

			if nodeParsed {
				parsedSites++
				parsedByProgram[program]++
				env, err := parse.Parse(programID, data, accounts, ji.StackHeight)
				if err != nil {
					t.Errorf("tx %d %s: node parsed as %s but Parse failed: %v\nnode: %s",
						ti, where, program, err, nodeRaw)
					return
				}
				raw, err := json.Marshal(env)
				if err != nil {
					t.Errorf("tx %d %s: marshaling envelope: %v", ti, where, err)
					return
				}
				got := decodeJSON(t, raw).(map[string]any)
				// The envelope always carries stackHeight (null when unset);
				// an older node may omit the key. Missing == null.
				if _, ok := node["stackHeight"]; !ok {
					node["stackHeight"] = nil
				}
				if path, diff, ok := firstDiff("", any(node), any(got)); !ok {
					t.Errorf("tx %d %s (%s): envelope mismatch at %q: %s\nnode: %s\nport: %s",
						ti, where, program, path, diff, nodeRaw, raw)
				}
				return
			}

			// A partiallyDecoded site: the node refused to parse it. If the
			// program is in the registry, its own parser rejected the bytes
			// (malformed data, foreign loader, bad account shape) — parsing
			// it here would emit an object the node never emitted. If it is
			// not, there is no parser at all. Either way: ErrNotParsable.
			if registryProgramIDs[programID] {
				registryFallbacks++
			} else {
				foreignFallbacks++
			}
			env, err := parse.Parse(programID, data, accounts, ji.StackHeight)
			if err == nil {
				t.Errorf("tx %d %s: Parse succeeded where the node fell back (program %s): %s",
					ti, where, programID, env.Parsed)
			} else if !errors.Is(err, parse.ErrNotParsable) {
				t.Errorf("tx %d %s: error does not wrap ErrNotParsable: %v", ti, where, err)
			}
		}

		jTop := jt.Transaction.Message.Instructions
		pTop := pt.Transaction.Message.Instructions
		if len(jTop) != len(pTop) {
			t.Fatalf("tx %d: %d json vs %d jsonParsed top-level instructions", ti, len(jTop), len(pTop))
		}
		for k := range jTop {
			checkSite(fmt.Sprintf("top[%d]", k), &jTop[k], pTop[k])
		}

		if jt.Meta == nil || pt.Meta == nil {
			if (jt.Meta == nil) != (pt.Meta == nil) {
				t.Fatalf("tx %d: meta present on only one fixture side", ti)
			}
			continue
		}
		if len(jt.Meta.InnerInstructions) != len(pt.Meta.InnerInstructions) {
			t.Fatalf("tx %d: inner group count mismatch between fixture sides", ti)
		}
		for gi := range jt.Meta.InnerInstructions {
			jg, pg := &jt.Meta.InnerInstructions[gi], &pt.Meta.InnerInstructions[gi]
			if jg.Index != pg.Index || len(jg.Instructions) != len(pg.Instructions) {
				t.Fatalf("tx %d inner group %d: shape mismatch between fixture sides", ti, gi)
			}
			for k := range jg.Instructions {
				checkSite(fmt.Sprintf("inner[%d][%d]", jg.Index, k), &jg.Instructions[k], pg.Instructions[k])
			}
		}
	}

	// A silently-empty walk must fail: the captured block carries 16 parsed
	// sites (7 system, 5 spl-token, 3 spl-associated-token-account, 1 vote).
	if parsedSites < 16 {
		t.Errorf("walked only %d parsed sites, want >= 16 — fixture or walk is broken", parsedSites)
	}
	if parsedByProgram["vote"] == 0 {
		t.Errorf("no parsed vote site walked — fixture or walk is broken")
	}
	if foreignFallbacks == 0 {
		t.Errorf("walked no foreign fallback sites — fixture or walk is broken")
	}
	// This capture happens to carry none, but the branch above stays armed:
	// a registry program the node rendered partiallyDecoded must refuse.
	_ = registryFallbacks
}

// ---------------------------------------------------------------------------
// Decoded-JSON comparison: key order and 0 vs 0.0 must not count as drift.
// ---------------------------------------------------------------------------

// decodeJSON parses with UseNumber so integer literals compare exactly (u64
// magnitudes never round through float64).
func decodeJSON(t *testing.T, b []byte) any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("decoding for comparison: %v\ninput: %s", err, b)
	}
	return v
}

// firstDiff reports the JSON pointer of the first difference between two
// decoded values, or ok=true when deep-equal. Both-integral numbers compare
// by digits; anything with a fraction or exponent compares by float64 value,
// so 0 vs 0.0 and 1.50 vs 1.5 are equal while 1 vs 2 is not.
func firstDiff(path string, want, got any) (string, string, bool) {
	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			return path, fmt.Sprintf("want object, got %s", describeJSON(got)), false
		}
		keySet := make(map[string]struct{}, len(w)+len(g))
		for k := range w {
			keySet[k] = struct{}{}
		}
		for k := range g {
			keySet[k] = struct{}{}
		}
		keys := make([]string, 0, len(keySet))
		for k := range keySet {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			wv, inW := w[k]
			gv, inG := g[k]
			if !inW {
				return path + "/" + k, fmt.Sprintf("unexpected key, got %s", describeJSON(gv)), false
			}
			if !inG {
				return path + "/" + k, fmt.Sprintf("missing key, want %s", describeJSON(wv)), false
			}
			if p, d, ok := firstDiff(path+"/"+k, wv, gv); !ok {
				return p, d, false
			}
		}
		return "", "", true
	case []any:
		g, ok := got.([]any)
		if !ok {
			return path, fmt.Sprintf("want array, got %s", describeJSON(got)), false
		}
		for i := 0; i < len(w) && i < len(g); i++ {
			if p, d, ok := firstDiff(path+"/"+strconv.Itoa(i), w[i], g[i]); !ok {
				return p, d, false
			}
		}
		if len(w) != len(g) {
			return path, fmt.Sprintf("array length %d, want %d", len(g), len(w)), false
		}
		return "", "", true
	case json.Number:
		g, ok := got.(json.Number)
		if !ok || !numbersEqual(w, g) {
			return path, fmt.Sprintf("want %s, got %s", w, describeJSON(got)), false
		}
		return "", "", true
	default:
		if want != got {
			return path, fmt.Sprintf("want %s, got %s", describeJSON(want), describeJSON(got)), false
		}
		return "", "", true
	}
}

func numbersEqual(a, b json.Number) bool {
	if a.String() == b.String() {
		return true
	}
	if isIntegralLiteral(a.String()) && isIntegralLiteral(b.String()) {
		return false // both exact integers with different digits
	}
	af, aerr := a.Float64()
	bf, berr := b.Float64()
	return aerr == nil && berr == nil && af == bf
}

func isIntegralLiteral(s string) bool {
	return !bytes.ContainsAny([]byte(s), ".eE")
}

func describeJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%#v", v)
	}
	return string(b)
}
