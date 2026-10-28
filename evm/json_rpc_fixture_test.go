package evm

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
)

// Real blocks as their nodes sent them, by eth_getBlockByNumber(n, true) and
// eth_getBlockReceipts(n). Base 53207078: legacy, 1559, 7702 and deposit
// transactions, OP fee and deposit receipt fields, daFootprintGasScalar.
// Ethereum 26292299: legacy, 2930, 1559, 4844 and 7702 transactions,
// withdrawals and blob gas fields.
var nodeFixtures = []string{"base-53207078", "ethereum-26292299"}

func readGzip(t testing.TB, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// wire sends m through protobuf encoding, as a BDS server's answer arrives:
// an empty bytes or repeated field comes back nil.
func wire[M proto.Message](t testing.TB, m M) M {
	t.Helper()
	b, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	out := m.ProtoReflect().New().Interface().(M)
	if err := proto.Unmarshal(b, out); err != nil {
		t.Fatal(err)
	}
	return out
}

type nodeFixture struct {
	receipts   *GetBlockReceiptsResponse
	fullBlock  *GetBlockResponse
	hashBlock  *GetBlockResponse
	blockLogs  []*Log
	signatures SignatureEncoding
}

// loadNodeFixture converts a node's JSON to the protobuf answers a BDS server
// gives for the same block. Full transactions carry their receipt's figures,
// as a server's stored records do.
func loadNodeFixture(t testing.TB, name string) nodeFixture {
	t.Helper()
	var jsonReceipts []*JsonRpcReceipt
	if err := json.Unmarshal(readGzip(t, filepath.Join("testdata", name+"-receipts.json.gz")), &jsonReceipts); err != nil {
		t.Fatal(err)
	}
	receipts := make([]*Receipt, len(jsonReceipts))
	var logs []*Log
	for i, jr := range jsonReceipts {
		r, err := jr.ToProto()
		if err != nil {
			t.Fatalf("receipt %d: %v", i, err)
		}
		receipts[i] = r
		logs = append(logs, r.Logs...)
	}

	var jsonBlock JsonRpcBlock
	if err := json.Unmarshal(readGzip(t, filepath.Join("testdata", name+"-block.json.gz")), &jsonBlock); err != nil {
		t.Fatal(err)
	}
	block, err := jsonBlock.ToProto()
	if err != nil {
		t.Fatal(err)
	}
	if len(block.FullTransactions) != len(receipts) {
		t.Fatalf("%d transactions, %d receipts", len(block.FullTransactions), len(receipts))
	}
	for i, tx := range block.FullTransactions {
		MergeReceipt(tx, receipts[i])
	}

	return nodeFixture{
		receipts: wire(t, &GetBlockReceiptsResponse{Receipts: receipts}),
		fullBlock: wire(t, &GetBlockResponse{
			Block:            block.Header,
			Transactions:     block.TransactionHashes,
			FullTransactions: block.FullTransactions,
			Withdrawals:      block.Withdrawals,
		}),
		hashBlock: wire(t, &GetBlockResponse{
			Block:        block.Header,
			Transactions: block.TransactionHashes,
			Withdrawals:  block.Withdrawals,
		}),
		blockLogs:  logs,
		signatures: SignatureQuantity,
	}
}

// syntheticFixture sets the fields no fixture block carries: Arbitrum, Celo
// and beacon fields, pre-Byzantium root, a pre-EIP-155 signature, a float
// l1FeeScalar, quantities beyond 64 bits, and a string that needs escaping.
func syntheticFixture() (*GetBlockResponse, []*Receipt) {
	u64 := func(v uint64) *uint64 { return &v }
	u32 := func(v uint32) *uint32 { return &v }
	str := func(v string) *string { return &v }
	f64 := func(v float64) *float64 { return &v }
	b := func(v bool) *bool { return &v }
	h := func(n int, fill byte) []byte { return bytes.Repeat([]byte{fill}, n) }

	header := &BlockHeader{
		Number: 7, Timestamp: 1700000000, GasLimit: 30000000, GasUsed: 21000, Size: 0x220,
		Hash: h(32, 1), ParentHash: h(32, 2), StateRoot: h(32, 3), TransactionsRoot: h(32, 4),
		ReceiptsRoot: h(32, 5), Sha3Uncles: h(32, 6), Miner: h(20, 7), LogsBloom: h(256, 0),
		ExtraData: []byte{}, Nonce: u64(0x42), BlobGasUsed: u64(0), ExcessBlobGas: u64(0x20000),
		L1BlockNumber: u64(19000000), Epoch: u64(3), Slot: u64(4), ProposerIndex: u64(5),
		SendCount: u64(6), TransactionCount: u32(4), MixHash: h(32, 8), ParentBeaconBlockRoot: h(32, 9),
		WithdrawalsRoot: h(32, 10), SendRoot: h(32, 11), BaseFeePerGas: str("7"),
		Difficulty: str("0"), TotalDifficulty: str("58750003716598352816469"),
		ProposerPublicKey: str("pk<&>\"\\\u2028é"), CanonicalRlp: []byte{0xf9, 0x02},
		Uncles: [][]byte{h(32, 12)}, RequestsHash: h(32, 13),
	}
	txs := []*Transaction{
		{ // pre-EIP-155 legacy contract creation: no chainId
			Hash: h(32, 0x21), Nonce: 0, From: h(20, 0x22), Value: "1000000000000000000000000",
			Input: []byte{0x60, 0x80}, Type: 0, GasLimit: 100000, GasPrice: str("20000000000"),
			R: h(32, 0xff), S: append([]byte{0}, h(31, 0x01)...), V: []byte{27}, ChainId: u64(1),
			BlockNumber: u64(7), BlockHash: h(32, 1), TransactionIndex: u32(0), BlockTimestamp: u64(1700000000),
			GasUsed: u64(53000), EffectiveGasPrice: str("20000000000"), L1Fee: str("not a number"),
			L1FeeScalar: f64(0.684), GatewayFee: str("5"), FeeCurrency: h(20, 0x23), GatewayFeeRecipient: h(20, 0x24),
		},
		{ // Arbitrum retryable, no access list type
			Hash: h(32, 0x31), Nonce: 1, From: h(20, 0x32), To: h(20, 0x33), Value: "0x0",
			Type: 0x69, GasLimit: 21000, MaxFeePerGas: str("0x5f5e100"), V: []byte{0, 0},
			Beneficiary: h(20, 0x34), DepositValue: str("10"), L1BaseFee: str("11"), MaxSubmissionFee: str("12"),
			RefundTo: h(20, 0x35), RequestId: h(32, 0x36), RetryData: []byte{1, 2}, RetryTo: h(20, 0x37),
			RetryValue: str("13"), MaxRefund: str("14"), SubmissionFeeRefund: str("15"), TicketId: h(32, 0x38),
			AccessList: []*AccessListItem{{Address: h(20, 0x39)}, nil},
		},
		{ // 7702 with an empty access list and an authority
			Hash: h(32, 0x41), Nonce: 2, From: h(20, 0x42), To: h(20, 0x43), Value: "",
			Type: 4, GasLimit: 60000, MaxFeePerGas: str("100"), MaxPriorityFeePerGas: str("1"),
			R: []byte{}, S: h(32, 0x44), V: []byte{1}, YParity: u32(1), ChainId: u64(8453),
			AuthorizationList: []*AuthorizationListItem{
				{ChainId: 0, Address: h(20, 0x45), Nonce: 9, R: nil, S: h(32, 0x46), YParity: 0},
				nil,
				{ChainId: 8453, Address: h(20, 0x47), Nonce: 10, R: h(31, 0x48), S: h(2, 0), YParity: 1, Authority: h(20, 0x49)},
			},
			MaxFeePerBlobGas: str("-1"), BlobVersionedHashes: [][]byte{h(32, 0x4a)},
		},
		{ // OP deposit
			Hash: h(32, 0x51), Nonce: 3, From: h(20, 0x52), To: h(20, 0x53), Value: "0",
			Type: 0x7e, GasLimit: 1000000, SourceHash: h(32, 0x54), Mint: str("0"),
			IsSystemTx: b(false), DepositReceiptVersion: str("1"), ChainId: u64(8453),
			R: []byte{}, S: []byte{}, V: []byte{},
		},
	}
	withdrawals := []*Withdrawal{{Index: 1, ValidatorIndex: 2, Address: h(20, 0x61), Amount: 3}, nil}
	receipts := []*Receipt{
		{
			TransactionHash: h(32, 0x21), BlockNumber: 7, BlockHash: h(32, 1), TransactionIndex: 0,
			Type: 0, From: h(20, 0x22), GasUsed: 53000, CumulativeGasUsed: 53000,
			EffectiveGasPrice: "20000000000", LogsBloom: h(256, 0), ContractAddress: h(20, 0x25),
			Root: h(32, 0x26), BlockTimestamp: u64(1700000000), BlobGasUsed: u64(0), BlobGasPrice: str("1"),
			Timeboosted: b(true), L1Fee: str("123456789012345678901234567890"), L1GasUsed: str("bad"),
			L1GasPrice: str("0x0"), L1FeeScalar: f64(1e-7), L1BaseFeeScalar: u64(1368), GasUsedForL1: u64(9),
			L1BlockNumber: u64(19000000), GatewayFee: str("2"), DepositNonce: str("0x10"),
			DepositReceiptVersion: str("1"), L1BlobBaseFee: str("1"), L1BlobBaseFeeScalar: u64(810949),
			DaFootprintGasScalar: u64(400),
			Logs: []*Log{{
				Address: h(20, 0x27), Topics: [][]byte{h(32, 0x28), {}}, Data: nil, BlockNumber: 7,
				BlockHash: h(32, 1), TransactionHash: h(32, 0x21), TransactionIndex: 0, LogIndex: 0,
				BlockTimestamp: u64(1700000000),
			}, nil},
		},
		nil,
		{
			TransactionHash: h(32, 0x41), BlockNumber: 7, BlockHash: h(32, 1), TransactionIndex: 2,
			Type: 4, From: h(20, 0x42), Status: u32(0), L1FeeScalar: f64(1.5e21),
		},
	}
	return &GetBlockResponse{Block: header, FullTransactions: txs, Withdrawals: withdrawals}, receipts
}

// renderings are the JSON-RPC answers a BDS-backed server sends.
func renderings(t testing.TB) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	for _, name := range nodeFixtures {
		f := loadNodeFixture(t, name)
		out[name+"-receipts"] = marshalOld(t, ReceiptsToJsonRpc(f.receipts.Receipts))
		out[name+"-block-full"] = marshalOld(t, BlockToJsonRpc(f.fullBlock.Block, f.fullBlock.Transactions, f.fullBlock.FullTransactions, f.fullBlock.Withdrawals, f.signatures))
		out[name+"-block-hashes"] = marshalOld(t, BlockToJsonRpc(f.hashBlock.Block, f.hashBlock.Transactions, f.hashBlock.FullTransactions, f.hashBlock.Withdrawals, f.signatures))
		out[name+"-logs"] = marshalOld(t, LogsToJsonRpc(f.blockLogs))
	}
	block, receipts := syntheticFixture()
	for _, sig := range []SignatureEncoding{SignatureQuantity, SignatureFixedData32} {
		suffix := map[SignatureEncoding]string{SignatureQuantity: "quantity", SignatureFixedData32: "fixed32"}[sig]
		out["synthetic-block-"+suffix] = marshalOld(t, BlockToJsonRpc(block.Block, nil, block.FullTransactions, block.Withdrawals, sig))
	}
	out["synthetic-receipts"] = marshalOld(t, ReceiptsToJsonRpc(receipts))
	out["synthetic-empty-block"] = marshalOld(t, BlockToJsonRpc(&BlockHeader{}, nil, nil, nil, SignatureQuantity))
	return out
}

func marshalOld(t testing.TB, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// The expected files were rendered by the map-based converters with
// encoding/json, which sorts object keys; the writers emit keys in that order.
func TestJsonRpcRenderingsMatchFixtures(t *testing.T) {
	for name, got := range renderings(t) {
		path := filepath.Join("testdata", name+".expected.json.gz")
		want := readGzip(t, path)
		var gotV, wantV interface{}
		if err := json.Unmarshal(got, &gotV); err != nil {
			t.Fatalf("%s: invalid JSON: %v", name, err)
		}
		if err := json.Unmarshal(want, &wantV); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotV, wantV) {
			t.Errorf("%s: parsed JSON differs from %s", name, path)
		}
		if !bytes.Equal(got, want) {
			i := 0
			for i < len(got) && i < len(want) && got[i] == want[i] {
				i++
			}
			t.Errorf("%s: bytes differ at %d: got …%s… want …%s…", name, i, snippet(got, i), snippet(want, i))
		}
	}
}

func snippet(b []byte, i int) string {
	lo, hi := max(0, i-60), min(len(b), i+60)
	return strings.ToValidUTF8(string(b[lo:hi]), "?")
}
