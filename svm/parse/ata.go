package parse

import "fmt"

// parseAssociatedToken ports Agave's parse_associated_token.rs. The wire form
// is a borsh unit-enum with one quirk Agave preserves: EMPTY data is the
// original Create (pre-idempotent deployments emitted no discriminant), and
// exactly one byte selects create/createIdempotent/recoverNested. Anything
// longer is malformed (borsh requires exact consumption).
func parseAssociatedToken(data []byte, accounts []string) (typeInfo, error) {
	kind := byte(0)
	if len(data) == 1 {
		kind = data[0]
	} else if len(data) > 1 {
		return typeInfo{}, fmt.Errorf("%w: associated-token instruction data %d bytes", ErrNotParsable, len(data))
	}
	switch kind {
	case 0, 1: // Create / CreateIdempotent
		if err := checkNumAccounts(accounts, 6); err != nil {
			return typeInfo{}, err
		}
		t := "create"
		if kind == 1 {
			t = "createIdempotent"
		}
		// The legacy 7-account form (trailing rent sysvar) resolves to the
		// same six named fields; the sysvar is simply not rendered.
		return typeInfo{Type: t, Info: map[string]any{
			"source":        accounts[0],
			"account":       accounts[1],
			"wallet":        accounts[2],
			"mint":          accounts[3],
			"systemProgram": accounts[4],
			"tokenProgram":  accounts[5],
		}}, nil
	case 2: // RecoverNested
		if err := checkNumAccounts(accounts, 7); err != nil {
			return typeInfo{}, err
		}
		return typeInfo{Type: "recoverNested", Info: map[string]any{
			"nestedSource": accounts[0],
			"nestedMint":   accounts[1],
			"destination":  accounts[2],
			"nestedOwner":  accounts[3],
			"ownerMint":    accounts[4],
			"wallet":       accounts[5],
			"tokenProgram": accounts[6],
		}}, nil
	}
	return typeInfo{}, fmt.Errorf("%w: associated-token discriminant %d", ErrNotParsable, kind)
}
