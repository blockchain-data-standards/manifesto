package evm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// jsonWriter appends JSON to b. more says whether a value already precedes
// the next one at the current depth, which then needs a comma.
//
// The Append*JsonRpc writers write object members in byte order of their
// keys, the order encoding/json gives a map, so their output is stable.
type jsonWriter struct {
	b    []byte
	more bool
	err  error // the first value JSON cannot hold
}

func (w *jsonWriter) sep() {
	if w.more {
		w.b = append(w.b, ',')
	}
	w.more = true
}

// key starts an object member; callers write members in key order.
func (w *jsonWriter) key(k string) {
	w.sep()
	w.b = append(w.b, '"')
	w.b = append(w.b, k...)
	w.b = append(w.b, '"', ':')
}

func (w *jsonWriter) open(c byte) {
	w.b = append(w.b, c)
	w.more = false
}

func (w *jsonWriter) close(c byte) {
	w.b = append(w.b, c)
	w.more = true
}

func (w *jsonWriter) null() { w.b = append(w.b, "null"...) }

func (w *jsonWriter) boolean(v bool) { w.b = strconv.AppendBool(w.b, v) }

// data writes DATA: 0x and two hex digits per byte, "0x" for none.
func (w *jsonWriter) data(v []byte) {
	w.b = append(w.b, '"', '0', 'x')
	w.b = hex.AppendEncode(w.b, v)
	w.b = append(w.b, '"')
}

// dataOrNull writes v as DATA, or null when empty.
func (w *jsonWriter) dataOrNull(v []byte) {
	if len(v) == 0 {
		w.null()
		return
	}
	w.data(v)
}

func (w *jsonWriter) dataList(vs [][]byte) {
	w.open('[')
	for _, v := range vs {
		w.sep()
		w.data(v)
	}
	w.close(']')
}

// quantity writes a QUANTITY: 0x and hex digits without leading zeros.
func (w *jsonWriter) quantity(v uint64) {
	w.b = append(w.b, '"', '0', 'x')
	w.b = strconv.AppendUint(w.b, v, 16)
	w.b = append(w.b, '"')
}

// bytesQuantity writes big-endian v as a QUANTITY, 0x0 for none.
func (w *jsonWriter) bytesQuantity(v []byte) {
	for len(v) > 0 && v[0] == 0 {
		v = v[1:]
	}
	if len(v) == 0 {
		w.b = append(w.b, `"0x0"`...)
		return
	}
	w.b = append(w.b, '"', '0', 'x')
	if v[0] < 0x10 {
		w.b = append(w.b, "0123456789abcdef"[v[0]])
		v = v[1:]
	}
	w.b = hex.AppendEncode(w.b, v)
	w.b = append(w.b, '"')
}

// fixed32 writes v as 32-byte DATA, left-padded with zeros; nil is "0x".
func (w *jsonWriter) fixed32(v []byte) {
	if v == nil {
		w.b = append(w.b, `"0x"`...)
		return
	}
	w.b = append(w.b, '"', '0', 'x')
	for i := len(v); i < 32; i++ {
		w.b = append(w.b, '0', '0')
	}
	w.b = hex.AppendEncode(w.b, v)
	w.b = append(w.b, '"')
}

func (w *jsonWriter) signature(v []byte, sig SignatureEncoding) {
	if sig == SignatureFixedData32 {
		w.fixed32(v)
		return
	}
	w.bytesQuantity(v)
}

// quantityField writes member k with *s, a decimal or 0x-hex number, as a
// QUANTITY (DecimalStringToHex); nil or not a number writes nothing.
func (w *jsonWriter) quantityField(k string, s *string) {
	if s == nil {
		return
	}
	n, more := len(w.b), w.more
	w.key(k)
	var v uint64
	var err error
	if len(*s) > 2 && (*s)[0] == '0' && ((*s)[1] == 'x' || (*s)[1] == 'X') {
		v, err = strconv.ParseUint((*s)[2:], 16, 64)
	} else {
		v, err = strconv.ParseUint(*s, 10, 64)
	}
	if err == nil {
		w.quantity(v)
		return
	}
	// Beyond 64 bits, or a form only big.Int parses.
	q, err := DecimalStringToHex(*s)
	if err != nil {
		w.b, w.more = w.b[:n], more
		return
	}
	w.b = append(w.b, '"')
	w.b = append(w.b, q...)
	w.b = append(w.b, '"')
}

// floatTextField writes member k with v as a JSON string: the shortest
// decimal that parses back to v, without an exponent ("0.684", "1"), as
// OP-stack nodes send l1FeeScalar. A NaN or infinity sets w.err.
func (w *jsonWriter) floatTextField(k string, v *float64) {
	if v == nil {
		return
	}
	if math.IsNaN(*v) || math.IsInf(*v, 0) {
		if w.err == nil {
			w.err = fmt.Errorf("%s is %v, which no node sends", k, *v)
		}
		return
	}
	w.key(k)
	w.b = append(w.b, '"')
	w.b = strconv.AppendFloat(w.b, *v, 'f', -1, 64)
	w.b = append(w.b, '"')
}

// text writes v as a JSON string, escaped as encoding/json escapes it.
func (w *jsonWriter) text(v string) {
	q, _ := json.Marshal(v) // a string always marshals
	w.b = append(w.b, q...)
}
