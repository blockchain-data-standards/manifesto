package evm

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"
)

// The writers against the map converters they replaced (oracle*, in
// json_rpc_oracle_test.go) marshalled by encoding/json: the same bytes, or an
// error from both. go test -fuzz=FuzzReceiptsJsonRpc and -fuzz=FuzzBlockJsonRpc.

// gen turns fuzz input into field values, each drawn from boundary pools.
type gen struct{ data []byte }

func (g *gen) next() byte {
	if len(g.data) == 0 {
		return 0
	}
	b := g.data[0]
	g.data = g.data[1:]
	return b
}

func (g *gen) u64() uint64 {
	switch g.next() % 8 {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return 34 // a legacy v just under EIP-155
	case 3:
		return 35
	case 4:
		return math.MaxUint64
	case 5:
		return 1 << 32
	default:
		var v uint64
		for range 8 {
			v = v<<8 | uint64(g.next())
		}
		return v
	}
}

func (g *gen) u32() uint32 { return uint32(g.u64()) }

func (g *gen) optU64() *uint64 {
	if g.next()%2 == 0 {
		return nil
	}
	v := g.u64()
	return &v
}

func (g *gen) optU32() *uint32 {
	if g.next()%2 == 0 {
		return nil
	}
	v := g.u32()
	return &v
}

func (g *gen) bytesV() []byte {
	switch g.next() % 8 {
	case 0:
		return nil
	case 1:
		return []byte{}
	case 2:
		return []byte{g.next()}
	case 3: // leading zero bytes
		b := make([]byte, 1+int(g.next()%40))
		for i := len(b) / 2; i < len(b); i++ {
			b[i] = g.next()
		}
		return b
	case 4:
		b := make([]byte, 32)
		for i := range b {
			b[i] = g.next()
		}
		return b
	case 5: // large, up to 64 KiB
		b := make([]byte, int(g.next())<<8|int(g.next()))
		for i := range b {
			b[i] = byte(i)
		}
		return b
	default:
		b := make([]byte, int(g.next()%70))
		for i := range b {
			b[i] = g.next()
		}
		return b
	}
}

func (g *gen) list() [][]byte {
	switch g.next() % 4 {
	case 0:
		return nil
	case 1:
		return [][]byte{}
	}
	n := int(g.next() % 5)
	out := make([][]byte, n)
	for i := range out {
		out[i] = g.bytesV()
	}
	return out
}

var numberStrings = []string{
	"", "0", "00", "0x", "0x0", "0x0001", "0X1F", "0xABCdef", "-1", "-0", "+5", "0x+5", " 12", "12 ",
	"18446744073709551615", "18446744073709551616", "0xffffffffffffffff", "0x10000000000000000",
	"115792089237316195423570985008687907853269984665640564039457584007913129639935",
	"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	"abc", "1e3", "0x-1", "1_000", "0xg", "007",
}

func (g *gen) numStr() string {
	i := int(g.next())
	if i < len(numberStrings) {
		return numberStrings[i]
	}
	n := 1 + i%30
	b := make([]byte, n)
	for j := range b {
		b[j] = '0' + g.next()%10
	}
	return string(b)
}

func (g *gen) optStr() *string {
	if g.next()%2 == 0 {
		return nil
	}
	s := g.numStr()
	return &s
}

var floats = []float64{0, math.Copysign(0, -1), 1e-7, 1e-6, 0.684, 1, 1.5e21, 1e21, 1e20, 123456789.123,
	math.MaxFloat64, 5e-324, -2.5, math.NaN(), math.Inf(1), math.Inf(-1)}

func (g *gen) optF64() *float64 {
	if g.next()%2 == 0 {
		return nil
	}
	v := floats[int(g.next())%len(floats)]
	return &v
}

func (g *gen) optBool() *bool {
	switch g.next() % 3 {
	case 0:
		return nil
	case 1:
		v := true
		return &v
	}
	v := false
	return &v
}

func (g *gen) log() *Log {
	if g.next()%8 == 0 {
		return nil
	}
	return &Log{
		Address: g.bytesV(), Topics: g.list(), Data: g.bytesV(), BlockNumber: g.u64(),
		BlockHash: g.bytesV(), TransactionHash: g.bytesV(), TransactionIndex: g.u32(),
		LogIndex: g.u32(), BlockTimestamp: g.optU64(),
	}
}

func (g *gen) receipt() *Receipt {
	if g.next()%8 == 0 {
		return nil
	}
	r := &Receipt{
		TransactionHash: g.bytesV(), BlockNumber: g.u64(), BlockHash: g.bytesV(),
		TransactionIndex: g.u32(), Type: g.u32(), From: g.bytesV(), To: g.bytesV(), Status: g.optU32(),
		GasUsed: g.u64(), CumulativeGasUsed: g.u64(), EffectiveGasPrice: g.numStr(), LogsBloom: g.bytesV(),
		ContractAddress: g.bytesV(), Root: g.bytesV(), BlockTimestamp: g.optU64(), BlobGasUsed: g.optU64(),
		BlobGasPrice: g.optStr(), Timeboosted: g.optBool(), L1Fee: g.optStr(), L1GasUsed: g.optStr(),
		L1GasPrice: g.optStr(), L1FeeScalar: g.optF64(), L1BaseFeeScalar: g.optU64(), GasUsedForL1: g.optU64(),
		L1BlockNumber: g.optU64(), GatewayFee: g.optStr(), DepositNonce: g.optStr(),
		DepositReceiptVersion: g.optStr(), L1BlobBaseFee: g.optStr(), L1BlobBaseFeeScalar: g.optU64(),
		DaFootprintGasScalar: g.optU64(),
	}
	if n := int(g.next() % 5); n > 0 {
		r.Logs = make([]*Log, n)
		for i := range r.Logs {
			r.Logs[i] = g.log()
		}
	} else if n == 0 && g.next()%2 == 0 {
		r.Logs = []*Log{}
	}
	return r
}

var txTypes = []uint32{0, 1, 2, 3, 4, 0x7e, 0x7b, 0x69, 0x6a, 0x71, 5, math.MaxUint32}

func (g *gen) transaction() *Transaction {
	if g.next()%8 == 0 {
		return nil
	}
	tx := &Transaction{
		Hash: g.bytesV(), Nonce: g.u64(), From: g.bytesV(), To: g.bytesV(), Value: g.numStr(),
		Input: g.bytesV(), Type: txTypes[int(g.next())%len(txTypes)], GasLimit: g.u64(),
		GasPrice: g.optStr(), MaxFeePerGas: g.optStr(), MaxPriorityFeePerGas: g.optStr(),
		GasUsed: g.optU64(), EffectiveGasPrice: g.optStr(), R: g.bytesV(), S: g.bytesV(), V: g.bytesV(),
		YParity: g.optU32(), ChainId: g.optU64(), BlockNumber: g.optU64(), BlockHash: g.bytesV(),
		TransactionIndex: g.optU32(), BlockTimestamp: g.optU64(), MaxFeePerBlobGas: g.optStr(),
		BlobVersionedHashes: g.list(), BlobGasUsed: g.optU64(), BlobGasPrice: g.optStr(),
		L1Fee: g.optStr(), L1GasPrice: g.optStr(), L1GasUsed: g.optStr(), L1FeeScalar: g.optF64(),
		L1BlobBaseFee: g.optStr(), L1BlobBaseFeeScalar: g.optU64(), GatewayFee: g.optStr(),
		FeeCurrency: g.bytesV(), GatewayFeeRecipient: g.bytesV(), Beneficiary: g.bytesV(),
		DepositValue: g.optStr(), L1BaseFee: g.optStr(), MaxSubmissionFee: g.optStr(), RefundTo: g.bytesV(),
		RequestId: g.bytesV(), RetryData: g.bytesV(), RetryTo: g.bytesV(), RetryValue: g.optStr(),
		MaxRefund: g.optStr(), SubmissionFeeRefund: g.optStr(), TicketId: g.bytesV(),
		IsSystemTx: g.optBool(), DepositReceiptVersion: g.optStr(), SourceHash: g.bytesV(), Mint: g.optStr(),
	}
	if n := int(g.next() % 4); n > 0 {
		tx.AccessList = make([]*AccessListItem, n)
		for i := range tx.AccessList {
			if g.next()%5 != 0 {
				tx.AccessList[i] = &AccessListItem{Address: g.bytesV(), StorageKeys: g.list()}
			}
		}
	}
	if n := int(g.next() % 4); n > 0 {
		tx.AuthorizationList = make([]*AuthorizationListItem, n)
		for i := range tx.AuthorizationList {
			if g.next()%5 != 0 {
				tx.AuthorizationList[i] = &AuthorizationListItem{
					ChainId: g.u64(), Address: g.bytesV(), Nonce: g.u64(), R: g.bytesV(), S: g.bytesV(),
					YParity: g.u32(), Authority: g.bytesV(),
				}
			}
		}
	}
	return tx
}

func (g *gen) header() *BlockHeader {
	if g.next()%10 == 0 {
		return nil
	}
	h := &BlockHeader{
		Number: g.u64(), Timestamp: g.u64(), GasLimit: g.u64(), GasUsed: g.u64(), Size: g.u64(),
		Hash: g.bytesV(), ParentHash: g.bytesV(), StateRoot: g.bytesV(), TransactionsRoot: g.bytesV(),
		ReceiptsRoot: g.bytesV(), Sha3Uncles: g.bytesV(), Miner: g.bytesV(), LogsBloom: g.bytesV(),
		ExtraData: g.bytesV(), Nonce: g.optU64(), BlobGasUsed: g.optU64(), ExcessBlobGas: g.optU64(),
		L1BlockNumber: g.optU64(), Epoch: g.optU64(), Slot: g.optU64(), ProposerIndex: g.optU64(),
		SendCount: g.optU64(), TransactionCount: g.optU32(), MixHash: g.bytesV(),
		ParentBeaconBlockRoot: g.bytesV(), WithdrawalsRoot: g.bytesV(), SendRoot: g.bytesV(),
		BaseFeePerGas: g.optStr(), Difficulty: g.optStr(), TotalDifficulty: g.optStr(),
		CanonicalRlp: g.bytesV(), Uncles: g.list(), RequestsHash: g.bytesV(),
	}
	if g.next()%2 == 0 {
		k := string(g.bytesV()) // arbitrary bytes, invalid UTF-8 included
		h.ProposerPublicKey = &k
	}
	return h
}

// dst starts some fuzz runs with a prefix, as a pooled caller's buffer does.
func (g *gen) dst() []byte {
	if g.next()%2 == 0 {
		return nil
	}
	b := make([]byte, 3, 3+int(g.next()))
	copy(b, "pre")
	return b
}

func expectSame(t *testing.T, dst, got []byte, gotErr error, old interface{}) {
	t.Helper()
	want, wantErr := json.Marshal(old)
	if (gotErr != nil) != (wantErr != nil) {
		t.Fatalf("error: writer %v, oracle %v", gotErr, wantErr)
	}
	if gotErr != nil {
		return
	}
	if !bytes.Equal(got[:len(dst)], dst) {
		t.Fatalf("writer overwrote its prefix: %q", got[:len(dst)])
	}
	got = got[len(dst):]
	if !json.Valid(got) {
		t.Fatalf("invalid JSON: %s", got)
	}
	if !bytes.Equal(got, want) {
		i := 0
		for i < len(got) && i < len(want) && got[i] == want[i] {
			i++
		}
		t.Fatalf("differs at %d:\n got …%s…\nwant …%s…", i, snippet(got, i), snippet(want, i))
	}
}

func FuzzReceiptsJsonRpc(f *testing.F) {
	f.Add([]byte{})
	f.Add(bytes.Repeat([]byte{0xff}, 256))
	f.Add(bytes.Repeat([]byte{1}, 256))
	f.Add([]byte{1, 3, 7, 13, 2, 1, 1, 0, 1, 1, 13})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := &gen{data: data}
		dst := g.dst()
		var rs []*Receipt
		if n := int(g.next() % 4); n > 0 {
			rs = make([]*Receipt, n)
			for i := range rs {
				rs[i] = g.receipt()
			}
		}
		got, err := AppendReceiptsJsonRpc(dst, rs)
		expectSame(t, dst, got, err, oracleReceiptsToJsonRpc(rs))
		var r *Receipt
		if len(rs) > 0 {
			r = rs[0]
		}
		got, err = AppendReceiptJsonRpc(dst, r)
		expectSame(t, dst, got, err, oracleReceiptToJsonRpc(r))
		var logs []*Log
		if r != nil {
			logs = r.Logs
		}
		expectSame(t, dst, AppendLogsJsonRpc(dst, logs), nil, oracleLogsToJsonRpc(logs))
	})
}

func FuzzBlockJsonRpc(f *testing.F) {
	f.Add([]byte{})
	f.Add(bytes.Repeat([]byte{0xff}, 512))
	f.Add(bytes.Repeat([]byte{1}, 512))
	f.Add(bytes.Repeat([]byte{3, 7}, 256))
	f.Fuzz(func(t *testing.T, data []byte) {
		g := &gen{data: data}
		dst := g.dst()
		sig := SignatureQuantity
		if g.next()%2 == 1 {
			sig = SignatureFixedData32
		}
		header := g.header()
		hashes := g.list()
		var txs []*Transaction
		if n := int(g.next() % 4); n > 0 {
			txs = make([]*Transaction, n)
			for i := range txs {
				txs[i] = g.transaction()
			}
		}
		var withdrawals []*Withdrawal
		switch g.next() % 3 {
		case 1:
			withdrawals = []*Withdrawal{}
		case 2:
			withdrawals = []*Withdrawal{{Index: g.u64(), ValidatorIndex: g.u64(), Address: g.bytesV(), Amount: g.u64()}, nil}
		}
		got := AppendBlockJsonRpc(dst, header, hashes, txs, withdrawals, sig)
		expectSame(t, dst, got, nil, oracleBlockToJsonRpc(header, hashes, txs, withdrawals, sig))
		var tx *Transaction
		if len(txs) > 0 {
			tx = txs[0]
		}
		expectSame(t, dst, AppendTransactionJsonRpc(dst, tx, sig), nil, oracleTransactionToJsonRpc(tx, sig))
	})
}
