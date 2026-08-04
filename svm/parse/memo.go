package parse

import (
	"fmt"
	"unicode/utf8"
)

// parseMemo ports Agave's parse_memo: the parsed form is a BARE JSON string —
// no {type, info} envelope — and invalid utf8 refuses (falls back) rather
// than lossily re-encoding.
func parseMemo(data []byte) (any, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("%w: memo is not valid utf8", ErrNotParsable)
	}
	return string(data), nil
}
