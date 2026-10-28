package evm

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// NumberishString moved to util.go

type JsonRpcWithdrawal struct {
	Index          string `json:"index"`
	ValidatorIndex string `json:"validatorIndex"`
	Address        string `json:"address"`
	Amount         string `json:"amount"`
}

func (w *JsonRpcWithdrawal) ToProto() (*Withdrawal, error) {
	index, err := NumberishToUint64(w.Index)
	if err != nil {
		return nil, fmt.Errorf("failed to parse withdrawal index: %w", err)
	}

	validatorIndex, err := NumberishToUint64(w.ValidatorIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse validator index: %w", err)
	}

	address, err := HexToBytes(w.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse withdrawal address: %w", err)
	}

	amount, err := NumberishToUint64(w.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to parse withdrawal amount: %w", err)
	}

	return &Withdrawal{
		Index:          index,
		ValidatorIndex: validatorIndex,
		Address:        address,
		Amount:         amount,
	}, nil
}

type JsonRpcBlock struct {
	BaseFeePerGas         string               `json:"baseFeePerGas"`
	BlobGasUsed           string               `json:"blobGasUsed"`
	Difficulty            string               `json:"difficulty"`
	ExcessBlobGas         string               `json:"excessBlobGas"`
	ExtraData             string               `json:"extraData"`
	GasLimit              string               `json:"gasLimit"`
	GasUsed               string               `json:"gasUsed"`
	Hash                  string               `json:"hash"`
	LogsBloom             string               `json:"logsBloom"`
	Miner                 string               `json:"miner"`
	MixHash               string               `json:"mixHash"`
	Nonce                 string               `json:"nonce"`
	Number                string               `json:"number"`
	ParentBeaconBlockRoot string               `json:"parentBeaconBlockRoot"`
	ParentHash            string               `json:"parentHash"`
	ReceiptsRoot          string               `json:"receiptsRoot"`
	Sha3Uncles            string               `json:"sha3Uncles"`
	Size                  string               `json:"size"`
	StateRoot             string               `json:"stateRoot"`
	Timestamp             string               `json:"timestamp"`
	TotalDifficulty       string               `json:"totalDifficulty"`
	TransactionsRoot      string               `json:"transactionsRoot"`
	Uncles                []string             `json:"uncles"`
	WithdrawalsRoot       string               `json:"withdrawalsRoot"`
	RequestsHash          string               `json:"requestsHash"`
	L1BlockNumber         string               `json:"l1BlockNumber"`
	SendCount             string               `json:"sendCount"`
	SendRoot              string               `json:"sendRoot"`
	Epoch                 string               `json:"epoch"`
	Slot                  string               `json:"slot"`
	ProposerIndex         string               `json:"proposerIndex"`
	TransactionCount      string               `json:"transactionCount"`
	ProposerPublicKey     string               `json:"proposerPublicKey"`
	Withdrawals           []*JsonRpcWithdrawal `json:"withdrawals"`
	CanonicalRlp          string               `json:"canonicalRlp"`
	Transactions          []interface{}        `json:"transactions"`
	Timeboosted           bool                 `json:"timeboosted"`
}

func (b *JsonRpcBlock) ToProto() (*Block, error) {
	number, err := NumberishToUint64(b.Number)
	if err != nil {
		return nil, err
	}
	hash, err := HexToBytes(b.Hash)
	if err != nil {
		return nil, err
	}
	parentHash, err := HexToBytes(b.ParentHash)
	if err != nil {
		return nil, err
	}
	timestamp, err := NumberishToUint64(b.Timestamp)
	if err != nil {
		return nil, err
	}
	gasLimit, err := NumberishToUint64(b.GasLimit)
	if err != nil {
		return nil, err
	}
	gasUsed, err := NumberishToUint64(b.GasUsed)
	if err != nil {
		return nil, err
	}
	var size uint64
	if b.Size != "" {
		size, err = NumberishToUint64(b.Size)
		if err != nil {
			return nil, err
		}
	}
	logsBloom, err := HexToBytes(b.LogsBloom)
	if err != nil {
		return nil, err
	}
	transactionsRoot, err := HexToBytes(b.TransactionsRoot)
	if err != nil {
		return nil, err
	}
	stateRoot, err := HexToBytes(b.StateRoot)
	if err != nil {
		return nil, err
	}
	receiptsRoot, err := HexToBytes(b.ReceiptsRoot)
	if err != nil {
		return nil, err
	}
	sha3Uncles, err := HexToBytes(b.Sha3Uncles)
	if err != nil {
		return nil, err
	}
	miner, err := HexToBytes(b.Miner)
	if err != nil {
		return nil, err
	}
	extraData, err := HexToBytes(b.ExtraData)
	if err != nil {
		return nil, err
	}

	// Handle optional fields
	var nonce *uint64
	if b.Nonce != "" {
		n, err := NumberishToUint64(b.Nonce)
		if err != nil {
			return nil, err
		}
		nonce = &n
	}

	var mixHash []byte
	if b.MixHash != "" {
		mixHash, err = HexToBytes(b.MixHash)
		if err != nil {
			return nil, err
		}
	}

	var withdrawalsRoot []byte
	if b.WithdrawalsRoot != "" {
		withdrawalsRoot, err = HexToBytes(b.WithdrawalsRoot)
		if err != nil {
			return nil, err
		}
	}

	var requestsHash []byte
	if b.RequestsHash != "" {
		requestsHash, err = HexToBytes(b.RequestsHash)
		if err != nil {
			return nil, err
		}
	}

	var blobGasUsed *uint64
	if b.BlobGasUsed != "" {
		bgu, err := NumberishToUint64(b.BlobGasUsed)
		if err != nil {
			return nil, err
		}
		blobGasUsed = &bgu
	}

	var excessBlobGas *uint64
	if b.ExcessBlobGas != "" {
		ebg, err := NumberishToUint64(b.ExcessBlobGas)
		if err != nil {
			return nil, err
		}
		excessBlobGas = &ebg
	}

	var parentBeaconBlockRoot []byte
	if b.ParentBeaconBlockRoot != "" {
		parentBeaconBlockRoot, err = HexToBytes(b.ParentBeaconBlockRoot)
		if err != nil {
			return nil, err
		}
	}

	// Optional L2-specific fields
	var l1BlockNumber *uint64
	if b.L1BlockNumber != "" {
		n, err := NumberishToUint64(b.L1BlockNumber)
		if err != nil {
			return nil, err
		}
		l1BlockNumber = &n
	}

	var sendCount *uint64
	if b.SendCount != "" {
		sc, err := NumberishToUint64(b.SendCount)
		if err != nil {
			return nil, err
		}
		sendCount = &sc
	}

	var sendRoot []byte
	if b.SendRoot != "" {
		sendRoot, err = HexToBytes(b.SendRoot)
		if err != nil {
			return nil, err
		}
	}

	uncles := make([][]byte, 0, len(b.Uncles))
	for _, uncle := range b.Uncles {
		uncleBytes, err := HexToBytes(uncle)
		if err != nil {
			return nil, err
		}
		uncles = append(uncles, uncleBytes)
	}

	var epoch *uint64
	if b.Epoch != "" {
		e, err := NumberishToUint64(b.Epoch)
		if err != nil {
			return nil, err
		}
		epoch = &e
	}

	var slot *uint64
	if b.Slot != "" {
		sl, err := NumberishToUint64(b.Slot)
		if err != nil {
			return nil, err
		}
		slot = &sl
	}

	var proposerIndex *uint64
	if b.ProposerIndex != "" {
		pi, err := NumberishToUint64(b.ProposerIndex)
		if err != nil {
			return nil, err
		}
		proposerIndex = &pi
	}

	var transactionCount *uint32
	if b.TransactionCount != "" {
		tc, err := NumberishToUint32(b.TransactionCount)
		if err != nil {
			return nil, err
		}
		transactionCount = &tc
	}

	var proposerPublicKey *string
	if b.ProposerPublicKey != "" {
		pp := b.ProposerPublicKey
		proposerPublicKey = &pp
	}

	// Note: withdrawals array from JSON-RPC is not stored in BlockHeader
	// BlockHeader only contains withdrawalsRoot
	// The full withdrawals data would be in the Block message

	var canonicalRlp []byte
	if b.CanonicalRlp != "" {
		canonicalRlp, err = HexToBytes(b.CanonicalRlp)
		if err != nil {
			return nil, err
		}
	}

	header := &BlockHeader{
		Number:                number,
		Timestamp:             timestamp,
		GasLimit:              gasLimit,
		GasUsed:               gasUsed,
		Size:                  size,
		Hash:                  hash,
		ParentHash:            parentHash,
		StateRoot:             stateRoot,
		TransactionsRoot:      transactionsRoot,
		ReceiptsRoot:          receiptsRoot,
		Sha3Uncles:            sha3Uncles,
		Miner:                 miner,
		LogsBloom:             logsBloom,
		ExtraData:             extraData,
		Nonce:                 nonce,
		BlobGasUsed:           blobGasUsed,
		ExcessBlobGas:         excessBlobGas,
		MixHash:               mixHash,
		ParentBeaconBlockRoot: parentBeaconBlockRoot,
		WithdrawalsRoot:       withdrawalsRoot,
		RequestsHash:          requestsHash,
		L1BlockNumber:         l1BlockNumber,
		SendCount:             sendCount,
		SendRoot:              sendRoot,
		Epoch:                 epoch,
		Slot:                  slot,
		ProposerIndex:         proposerIndex,
		TransactionCount:      transactionCount,
		BaseFeePerGas:         &b.BaseFeePerGas,
		Difficulty:            &b.Difficulty,
		TotalDifficulty:       &b.TotalDifficulty,
		Uncles:                uncles,
		ProposerPublicKey:     proposerPublicKey,
		CanonicalRlp:          canonicalRlp,
	}

	hashes, txs, err := ParseJsonRpcTransactions(b.Transactions, header)
	if err != nil {
		return nil, err
	}

	withdrawals, err := ParseJsonRpcWithdrawals(b.Withdrawals)
	if err != nil {
		return nil, err
	}

	return &Block{
		Header:            header,
		FullTransactions:  txs,
		TransactionHashes: hashes,
		Withdrawals:       withdrawals,
	}, nil
}

func ParseJsonRpcTransactions(transactions []interface{}, header *BlockHeader) ([][]byte, []*Transaction, error) {
	hashes := make([][]byte, 0, len(transactions))
	txs := make([]*Transaction, 0, len(transactions))
	for _, tx := range transactions {
		switch v := tx.(type) {
		case string:
			// Transaction hash only
			if hashBytes, err := HexToBytes(v); err == nil {
				hashes = append(hashes, hashBytes)
			} else {
				return nil, nil, fmt.Errorf("failed to parse transaction hash: %w", err)
			}
		case map[string]interface{}:
			// Full transaction object
			tx, err := ParseJsonRpcTransaction(v, header)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse transaction: %w", err)
			}
			txs = append(txs, tx)
			hashes = append(hashes, tx.Hash)
		}
	}
	return hashes, txs, nil
}

type JsonRpcReceipt struct {
	BlockHash             string          `json:"blockHash"`
	BlockNumber           string          `json:"blockNumber"`
	BlockTimestamp        NumberishString `json:"blockTimestamp"`
	ContractAddress       string          `json:"contractAddress"`
	CumulativeGasUsed     string          `json:"cumulativeGasUsed"`
	EffectiveGasPrice     string          `json:"effectiveGasPrice"`
	From                  string          `json:"from"`
	GasUsed               string          `json:"gasUsed"`
	Logs                  []*JsonRpcLog   `json:"logs"`
	LogsBloom             string          `json:"logsBloom"`
	Root                  string          `json:"root"`
	Status                string          `json:"status"`
	To                    string          `json:"to"`
	TransactionHash       string          `json:"transactionHash"`
	TransactionIndex      string          `json:"transactionIndex"`
	Type                  string          `json:"type"`
	BlobGasUsed           string          `json:"blobGasUsed"`
	BlobGasPrice          string          `json:"blobGasPrice"`
	L1Fee                 string          `json:"l1Fee"`
	L1GasUsed             string          `json:"l1GasUsed"`
	L1GasPrice            string          `json:"l1GasPrice"`
	L1FeeScalar           string          `json:"l1FeeScalar"`
	L1BaseFeeScalar       string          `json:"l1BaseFeeScalar"`
	L1BlobBaseFee         string          `json:"l1BlobBaseFee"`
	L1BlobBaseFeeScalar    string          `json:"l1BlobBaseFeeScalar"`
	DaFootprintGasScalar   string          `json:"daFootprintGasScalar"`
	GasUsedForL1          string          `json:"gasUsedForL1"`
	L1BlockNumber         string          `json:"l1BlockNumber"`
	GatewayFee            string          `json:"gatewayFee"`
	DepositNonce          string          `json:"depositNonce"`
	DepositReceiptVersion string          `json:"depositReceiptVersion"`
	Timeboosted           *bool           `json:"timeboosted"`
}

func (r *JsonRpcReceipt) ToProto() (*Receipt, error) {
	blockNumber, err := NumberishToUint64(r.BlockNumber)
	if err != nil {
		return nil, err
	}
	transactionIndex, err := NumberishToUint32(r.TransactionIndex)
	if err != nil {
		return nil, err
	}
	gasUsed, err := NumberishToUint64(r.GasUsed)
	if err != nil {
		return nil, err
	}
	cumulativeGasUsed, err := NumberishToUint64(r.CumulativeGasUsed)
	if err != nil {
		return nil, err
	}
	logsBloom, err := HexToBytes(r.LogsBloom)
	if err != nil {
		return nil, err
	}
	blockHash, err := HexToBytes(r.BlockHash)
	if err != nil {
		return nil, err
	}
	transactionHash, err := HexToBytes(r.TransactionHash)
	if err != nil {
		return nil, err
	}

	// Parse from address
	from, err := HexToBytes(r.From)
	if err != nil {
		return nil, err
	}

	// Parse to address (optional - can be null for contract creation)
	var to []byte
	if r.To != "" && r.To != "0x" {
		to, err = HexToBytes(r.To)
		if err != nil {
			return nil, err
		}
	}

	// Handle optional fields
	var typ uint32
	if r.Type != "" {
		t, err := NumberishToUint32(r.Type)
		if err != nil {
			return nil, err
		}
		typ = t
	}

	var status *uint32
	if r.Status != "" {
		s, err := NumberishToUint32(r.Status)
		if err != nil {
			return nil, err
		}
		status = &s
	}

	var contractAddress []byte
	if r.ContractAddress != "" && r.ContractAddress != "0x" {
		contractAddress, err = HexToBytes(r.ContractAddress)
		if err != nil {
			return nil, err
		}
	}

	var root []byte
	if r.Root != "" {
		root, err = HexToBytes(r.Root)
		if err != nil {
			return nil, err
		}
	}

	logs := make([]*Log, 0, len(r.Logs))
	for _, log := range r.Logs {
		protoLog, err := log.ToProto()
		if err != nil {
			return nil, err
		}
		logs = append(logs, protoLog)
	}

	// Optional blob & L2 fee fields
	var blobGasUsed *uint64
	if r.BlobGasUsed != "" {
		bg, err := NumberishToUint64(r.BlobGasUsed)
		if err != nil {
			return nil, err
		}
		blobGasUsed = &bg
	}
	var gasUsedForL1 *uint64
	if r.GasUsedForL1 != "" {
		gu, err := NumberishToUint64(r.GasUsedForL1)
		if err != nil {
			return nil, err
		}
		gasUsedForL1 = &gu
	}
	var l1BlockNumber *uint64
	if r.L1BlockNumber != "" {
		bn, err := NumberishToUint64(r.L1BlockNumber)
		if err != nil {
			return nil, err
		}
		l1BlockNumber = &bn
	}

	// Scalars & decimal strings remain as-is (string pointers)
	var l1Fee *string
	if r.L1Fee != "" {
		l1Fee = &r.L1Fee
	}
	var l1GasUsed *string
	if r.L1GasUsed != "" {
		l1GasUsed = &r.L1GasUsed
	}
	var l1GasPrice *string
	if r.L1GasPrice != "" {
		l1GasPrice = &r.L1GasPrice
	}
	var l1FeeScalar *float64
	if r.L1FeeScalar != "" {
		f, err := strconv.ParseFloat(r.L1FeeScalar, 64)
		if err != nil {
			return nil, err
		}
		l1FeeScalar = &f
	}
	var l1BaseFeeScalar *uint64
	if r.L1BaseFeeScalar != "" {
		scl, err := NumberishToUint64(r.L1BaseFeeScalar)
		if err != nil {
			return nil, err
		}
		l1BaseFeeScalar = &scl
	}
	var l1BlobBaseFee *string
	if r.L1BlobBaseFee != "" {
		l1BlobBaseFee = &r.L1BlobBaseFee
	}
	var l1BlobBaseFeeScalar *uint64
	if r.L1BlobBaseFeeScalar != "" {
		scl, err := NumberishToUint64(r.L1BlobBaseFeeScalar)
		if err != nil {
			return nil, err
		}
		l1BlobBaseFeeScalar = &scl
	}
	var daFootprintGasScalar *uint64
	if r.DaFootprintGasScalar != "" {
		scl, err := NumberishToUint64(r.DaFootprintGasScalar)
		if err != nil {
			return nil, err
		}
		daFootprintGasScalar = &scl
	}
	var gatewayFee *string
	if r.GatewayFee != "" {
		gatewayFee = &r.GatewayFee
	}
	var depositNonce *string
	if r.DepositNonce != "" {
		depositNonce = &r.DepositNonce
	}
	var depositReceiptVersion *string
	if r.DepositReceiptVersion != "" {
		depositReceiptVersion = &r.DepositReceiptVersion
	}

	var blobGasPrice *string
	if r.BlobGasPrice != "" {
		bp := r.BlobGasPrice
		blobGasPrice = &bp
	}

	var blockTimestamp *uint64
	if r.BlockTimestamp != "" {
		if bt, err := r.BlockTimestamp.ToUint64(); err == nil {
			blockTimestamp = &bt
		}
	}

	var timeboosted *bool
	if r.Timeboosted != nil {
		timeboosted = r.Timeboosted
	}

	return &Receipt{
		TransactionHash:       transactionHash,
		BlockNumber:           blockNumber,
		BlockHash:             blockHash,
		TransactionIndex:      uint32(transactionIndex),
		Type:                  typ,
		From:                  from,
		To:                    to,
		Status:                status,
		GasUsed:               gasUsed,
		CumulativeGasUsed:     cumulativeGasUsed,
		EffectiveGasPrice:     r.EffectiveGasPrice,
		LogsBloom:             logsBloom,
		Logs:                  logs,
		ContractAddress:       contractAddress,
		Root:                  root,
		BlockTimestamp:        blockTimestamp,
		BlobGasUsed:           blobGasUsed,
		BlobGasPrice:          blobGasPrice,
		L1Fee:                 l1Fee,
		L1GasUsed:             l1GasUsed,
		L1GasPrice:            l1GasPrice,
		L1FeeScalar:           l1FeeScalar,
		L1BaseFeeScalar:       l1BaseFeeScalar,
		L1BlobBaseFee:         l1BlobBaseFee,
		L1BlobBaseFeeScalar:  l1BlobBaseFeeScalar,
		DaFootprintGasScalar: daFootprintGasScalar,
		GasUsedForL1:          gasUsedForL1,
		L1BlockNumber:         l1BlockNumber,
		GatewayFee:            gatewayFee,
		DepositNonce:          depositNonce,
		DepositReceiptVersion: depositReceiptVersion,
		Timeboosted:           timeboosted,
	}, nil
}

type JsonRpcLog struct {
	Address          string          `json:"address"`
	BlockHash        string          `json:"blockHash"`
	BlockNumber      string          `json:"blockNumber"`
	BlockTimestamp   NumberishString `json:"blockTimestamp"`
	Data             string          `json:"data"`
	LogIndex         string          `json:"logIndex"`
	Topics           []string        `json:"topics"`
	TransactionHash  string          `json:"transactionHash"`
	TransactionIndex string          `json:"transactionIndex"`
}

func (l *JsonRpcLog) ToProto() (*Log, error) {
	address, err := HexToBytes(l.Address)
	if err != nil {
		return nil, err
	}
	blockNumber, err := NumberishToUint64(l.BlockNumber)
	if err != nil {
		return nil, err
	}
	var blockTimestamp *uint64
	if l.BlockTimestamp != "" {
		u, err := l.BlockTimestamp.ToUint64()
		if err != nil {
			return nil, err
		}
		blockTimestamp = &u
	}
	data, err := HexToBytes(l.Data)
	if err != nil {
		return nil, err
	}
	logIndex, err := NumberishToUint32(l.LogIndex)
	if err != nil {
		return nil, err
	}
	transactionIndex, err := NumberishToUint32(l.TransactionIndex)
	if err != nil {
		return nil, err
	}
	transactionHash, err := HexToBytes(l.TransactionHash)
	if err != nil {
		return nil, err
	}
	blockHash, err := HexToBytes(l.BlockHash)
	if err != nil {
		return nil, err
	}
	topics := make([][]byte, 0, len(l.Topics))
	for _, topic := range l.Topics {
		topicBytes, err := HexToBytes(topic)
		if err != nil {
			return nil, err
		}
		topics = append(topics, topicBytes)
	}
	return &Log{
		Address:          address,
		BlockHash:        blockHash,
		BlockNumber:      blockNumber,
		BlockTimestamp:   blockTimestamp,
		Data:             data,
		LogIndex:         logIndex,
		Topics:           topics,
		TransactionHash:  transactionHash,
		TransactionIndex: transactionIndex,
	}, nil
}

// AppendLogsJsonRpc appends logs as the JSON-RPC array eth_getLogs returns.
func AppendLogsJsonRpc(b []byte, logs []*Log) []byte {
	w := jsonWriter{b: b}
	w.logs(logs)
	return w.b
}

func (w *jsonWriter) logs(logs []*Log) {
	w.open('[')
	for _, l := range logs {
		w.sep()
		w.log(l)
	}
	w.close(']')
}

func (w *jsonWriter) log(l *Log) {
	if l == nil {
		w.null()
		return
	}
	w.open('{')
	w.key("address")
	w.data(l.Address)
	w.key("blockHash")
	w.data(l.BlockHash)
	w.key("blockNumber")
	w.quantity(l.BlockNumber)
	if l.BlockTimestamp != nil {
		w.key("blockTimestamp")
		w.quantity(*l.BlockTimestamp)
	}
	w.key("data")
	w.data(l.Data)
	w.key("logIndex")
	w.quantity(uint64(l.LogIndex))
	w.key("removed")
	w.boolean(false)
	w.key("topics")
	w.dataList(l.Topics)
	w.key("transactionHash")
	w.data(l.TransactionHash)
	w.key("transactionIndex")
	w.quantity(uint64(l.TransactionIndex))
	w.close('}')
}

// SignatureEncoding selects how a transaction's signature r and s render in
// JSON-RPC output.
type SignatureEncoding uint8

const (
	// SignatureQuantity renders r and s as QUANTITY (no leading zeros), as the
	// execution-apis schema specifies and most nodes emit.
	SignatureQuantity SignatureEncoding = iota
	// SignatureFixedData32 renders r and s as 32-byte DATA, left-padded with
	// zeros, as Tron nodes (java-tron) emit.
	SignatureFixedData32
)

// SignatureEncodingForChain returns the encoding the nodes of chainId emit.
// Unknown chains, including 0, get SignatureQuantity.
func SignatureEncodingForChain(chainId uint64) SignatureEncoding {
	switch chainId {
	case 728126428, 2494104990, 3448148188: // Tron mainnet, Shasta, Nile
		return SignatureFixedData32
	}
	return SignatureQuantity
}

func (e SignatureEncoding) hex(b []byte) string {
	if e == SignatureFixedData32 {
		return BytesToHexFixed(b, 32)
	}
	return BytesToQuantityHex(b)
}

// AppendTransactionJsonRpc appends the JSON-RPC transaction object a node
// sends for eth_getTransactionByHash and full blocks, or null for nil.
// sig selects the r/s encoding; pass SignatureEncodingForChain(chainId).
//
// A BDS Transaction also carries its receipt's execution figures (gasUsed,
// effectiveGasPrice, blob gas, L2 fees) and optional fields a chain may not
// have. Nodes send those figures on the receipt only, and omit a field the
// transaction does not have rather than sending null, so this leaves them
// out: a value the record carries is never invented or nulled.
//
// The shape follows what providers send (2026-10-08 survey: 22 chains, 2-8
// providers each): an empty accessList only on the types that define one, no
// chainId on an OP deposit (0x7e) or a pre-EIP-155 legacy transaction.
func AppendTransactionJsonRpc(b []byte, tx *Transaction, sig SignatureEncoding) []byte {
	w := jsonWriter{b: b}
	w.transaction(tx, sig)
	return w.b
}

// accessListTypes are the transaction types that define an access list, so
// nodes send one even when it is empty: EIP-2930 (1), EIP-1559 (2), EIP-4844
// (3), EIP-7702 (4) and Celo's CIP-64 (0x7b).
var accessListTypes = map[uint32]bool{1: true, 2: true, 3: true, 4: true, 0x7b: true}

// sendsChainId says whether a node names tx's chainId: an OP deposit is not
// signed by its sender, and a legacy transaction names one only when its
// signature does (EIP-155: v = chainId*2 + 35 or 36, else 27 or 28).
func sendsChainId(tx *Transaction) bool {
	switch tx.Type {
	case 0x7e:
		return false
	case 0:
		if tx.V == nil {
			return true
		}
		v := tx.V
		for len(v) > 0 && v[0] == 0 {
			v = v[1:]
		}
		return len(v) > 1 || (len(v) == 1 && v[0] >= 35)
	}
	return true
}

func (w *jsonWriter) transaction(tx *Transaction, sig SignatureEncoding) {
	if tx == nil {
		w.null()
		return
	}
	w.open('{')
	if len(tx.AccessList) > 0 || accessListTypes[tx.Type] {
		w.key("accessList")
		w.open('[')
		for _, item := range tx.AccessList {
			if item == nil {
				continue
			}
			w.sep()
			w.open('{')
			w.key("address")
			w.data(item.Address)
			w.key("storageKeys")
			w.dataList(item.StorageKeys)
			w.close('}')
		}
		w.close(']')
	}
	if len(tx.AuthorizationList) > 0 {
		w.key("authorizationList")
		w.open('[')
		for _, auth := range tx.AuthorizationList {
			if auth == nil {
				continue
			}
			w.sep()
			w.open('{')
			w.key("address")
			w.data(auth.Address)
			if len(auth.Authority) > 0 {
				w.key("authority")
				w.data(auth.Authority)
			}
			w.key("chainId")
			w.quantity(auth.ChainId)
			w.key("nonce")
			w.quantity(auth.Nonce)
			w.key("r")
			w.signature(auth.R, sig)
			w.key("s")
			w.signature(auth.S, sig)
			w.key("yParity")
			w.quantity(uint64(auth.YParity))
			w.close('}')
		}
		w.close(']')
	}
	if len(tx.Beneficiary) > 0 {
		w.key("beneficiary")
		w.data(tx.Beneficiary)
	}
	if len(tx.BlobVersionedHashes) > 0 {
		w.key("blobVersionedHashes")
		w.dataList(tx.BlobVersionedHashes)
	}
	if tx.BlockHash != nil {
		w.key("blockHash")
		w.data(tx.BlockHash)
	}
	if tx.BlockNumber != nil {
		w.key("blockNumber")
		w.quantity(*tx.BlockNumber)
	}
	if tx.BlockTimestamp != nil {
		w.key("blockTimestamp")
		w.quantity(*tx.BlockTimestamp)
	}
	if tx.ChainId != nil && sendsChainId(tx) {
		w.key("chainId")
		w.quantity(*tx.ChainId)
	}
	w.quantityField("depositReceiptVersion", tx.DepositReceiptVersion)
	w.quantityField("depositValue", tx.DepositValue)
	if len(tx.FeeCurrency) > 0 {
		w.key("feeCurrency")
		w.data(tx.FeeCurrency)
	}
	w.key("from")
	w.data(tx.From)
	w.key("gas")
	w.quantity(tx.GasLimit)
	w.quantityField("gasPrice", tx.GasPrice)
	w.quantityField("gatewayFee", tx.GatewayFee)
	if len(tx.GatewayFeeRecipient) > 0 {
		w.key("gatewayFeeRecipient")
		w.data(tx.GatewayFeeRecipient)
	}
	w.key("hash")
	w.data(tx.Hash)
	w.key("input")
	w.data(tx.Input)
	if tx.IsSystemTx != nil {
		w.key("isSystemTx")
		w.boolean(*tx.IsSystemTx)
	}
	w.quantityField("l1BaseFee", tx.L1BaseFee)
	w.quantityField("maxFeePerBlobGas", tx.MaxFeePerBlobGas)
	w.quantityField("maxFeePerGas", tx.MaxFeePerGas)
	w.quantityField("maxPriorityFeePerGas", tx.MaxPriorityFeePerGas)
	w.quantityField("maxRefund", tx.MaxRefund)
	w.quantityField("maxSubmissionFee", tx.MaxSubmissionFee)
	w.quantityField("mint", tx.Mint)
	w.key("nonce")
	w.quantity(tx.Nonce)
	if tx.R != nil {
		w.key("r")
		w.signature(tx.R, sig)
	}
	if len(tx.RefundTo) > 0 {
		w.key("refundTo")
		w.data(tx.RefundTo)
	}
	if len(tx.RequestId) > 0 {
		w.key("requestId")
		w.data(tx.RequestId)
	}
	if len(tx.RetryData) > 0 {
		w.key("retryData")
		w.data(tx.RetryData)
	}
	if len(tx.RetryTo) > 0 {
		w.key("retryTo")
		w.data(tx.RetryTo)
	}
	w.quantityField("retryValue", tx.RetryValue)
	if tx.S != nil {
		w.key("s")
		w.signature(tx.S, sig)
	}
	if len(tx.SourceHash) > 0 {
		w.key("sourceHash")
		w.data(tx.SourceHash)
	}
	w.quantityField("submissionFeeRefund", tx.SubmissionFeeRefund)
	if len(tx.TicketId) > 0 {
		w.key("ticketId")
		w.data(tx.TicketId)
	}
	// "to" is null on a contract creation, and nodes send it that way.
	w.key("to")
	w.dataOrNull(tx.To)
	if tx.TransactionIndex != nil {
		w.key("transactionIndex")
		w.quantity(uint64(*tx.TransactionIndex))
	}
	w.key("type")
	w.quantity(uint64(tx.Type))
	if tx.V != nil {
		w.key("v")
		w.bytesQuantity(tx.V)
	}
	if tx.Value != "" {
		w.quantityField("value", &tx.Value)
	}
	if tx.YParity != nil {
		w.key("yParity")
		w.quantity(uint64(*tx.YParity))
	}
	w.close('}')
}

// AppendReceiptsJsonRpc appends receipts as the JSON-RPC array
// eth_getBlockReceipts returns. A value JSON cannot hold, a NaN or infinite
// l1FeeScalar, is an error; b then holds a partial result, returned so a
// caller can reuse its buffer.
func AppendReceiptsJsonRpc(b []byte, receipts []*Receipt) ([]byte, error) {
	w := jsonWriter{b: b}
	w.open('[')
	for _, r := range receipts {
		w.sep()
		w.receipt(r)
	}
	w.close(']')
	return w.b, w.err
}

// AppendReceiptJsonRpc appends the JSON-RPC receipt object, or null for nil.
// L2 fee fields appear only on chains that have them: a node elsewhere omits
// them rather than sending null. No blockTimestamp: no provider sends one on
// a receipt (it is on the receipt's logs instead). Errors as
// AppendReceiptsJsonRpc does.
func AppendReceiptJsonRpc(b []byte, r *Receipt) ([]byte, error) {
	w := jsonWriter{b: b}
	w.receipt(r)
	return w.b, w.err
}

func (w *jsonWriter) receipt(r *Receipt) {
	if r == nil {
		w.null()
		return
	}
	w.open('{')
	w.quantityField("blobGasPrice", r.BlobGasPrice)
	if r.BlobGasUsed != nil {
		w.key("blobGasUsed")
		w.quantity(*r.BlobGasUsed)
	}
	w.key("blockHash")
	w.data(r.BlockHash)
	w.key("blockNumber")
	w.quantity(r.BlockNumber)
	w.key("contractAddress")
	w.dataOrNull(r.ContractAddress)
	w.key("cumulativeGasUsed")
	w.quantity(r.CumulativeGasUsed)
	if r.DaFootprintGasScalar != nil {
		w.key("daFootprintGasScalar")
		w.quantity(*r.DaFootprintGasScalar)
	}
	w.quantityField("depositNonce", r.DepositNonce)
	w.quantityField("depositReceiptVersion", r.DepositReceiptVersion)
	if r.EffectiveGasPrice != "" {
		w.quantityField("effectiveGasPrice", &r.EffectiveGasPrice)
	}
	w.key("from")
	w.data(r.From)
	w.key("gasUsed")
	w.quantity(r.GasUsed)
	if r.GasUsedForL1 != nil {
		w.key("gasUsedForL1")
		w.quantity(*r.GasUsedForL1)
	}
	w.quantityField("gatewayFee", r.GatewayFee)
	if r.L1BaseFeeScalar != nil {
		w.key("l1BaseFeeScalar")
		w.quantity(*r.L1BaseFeeScalar)
	}
	w.quantityField("l1BlobBaseFee", r.L1BlobBaseFee)
	if r.L1BlobBaseFeeScalar != nil {
		w.key("l1BlobBaseFeeScalar")
		w.quantity(*r.L1BlobBaseFeeScalar)
	}
	if r.L1BlockNumber != nil {
		w.key("l1BlockNumber")
		w.quantity(*r.L1BlockNumber)
	}
	w.quantityField("l1Fee", r.L1Fee)
	w.floatField("l1FeeScalar", r.L1FeeScalar)
	w.quantityField("l1GasPrice", r.L1GasPrice)
	w.quantityField("l1GasUsed", r.L1GasUsed)
	w.key("logs")
	w.logs(r.Logs)
	w.key("logsBloom")
	w.data(r.LogsBloom)
	// Pre-Byzantium receipts carry a state root instead of a status.
	if len(r.Root) > 0 {
		w.key("root")
		w.data(r.Root)
	}
	if r.Status != nil {
		w.key("status")
		w.quantity(uint64(*r.Status))
	}
	if r.Timeboosted != nil {
		w.key("timeboosted")
		w.boolean(*r.Timeboosted)
	}
	// "to" is null on a contract creation.
	w.key("to")
	w.dataOrNull(r.To)
	w.key("transactionHash")
	w.data(r.TransactionHash)
	w.key("transactionIndex")
	w.quantity(uint64(r.TransactionIndex))
	w.key("type")
	w.quantity(uint64(r.Type))
	w.close('}')
}

func (w *jsonWriter) withdrawal(wd *Withdrawal) {
	if wd == nil {
		w.null()
		return
	}
	w.open('{')
	w.key("address")
	w.data(wd.Address)
	w.key("amount")
	w.quantity(wd.Amount)
	w.key("index")
	w.quantity(wd.Index)
	w.key("validatorIndex")
	w.quantity(wd.ValidatorIndex)
	w.close('}')
}

// AppendBlockJsonRpc appends the JSON-RPC block object for header, or null
// for nil. Its transactions are fullTxs when given, else txHashes.
// sig selects the r/s encoding of full transactions; pass SignatureEncodingForChain(chainId).
func AppendBlockJsonRpc(b []byte, header *BlockHeader, txHashes [][]byte, fullTxs []*Transaction, withdrawals []*Withdrawal, sig SignatureEncoding) []byte {
	w := jsonWriter{b: b}
	if header == nil {
		w.null()
		return w.b
	}
	w.open('{')
	w.quantityField("baseFeePerGas", header.BaseFeePerGas)
	if header.BlobGasUsed != nil {
		w.key("blobGasUsed")
		w.quantity(*header.BlobGasUsed)
	}
	if header.CanonicalRlp != nil {
		w.key("canonicalRlp")
		w.data(header.CanonicalRlp)
	}
	w.quantityField("difficulty", header.Difficulty)
	if header.Epoch != nil {
		w.key("epoch")
		w.quantity(*header.Epoch)
	}
	if header.ExcessBlobGas != nil {
		w.key("excessBlobGas")
		w.quantity(*header.ExcessBlobGas)
	}
	w.key("extraData")
	w.data(header.ExtraData)
	w.key("gasLimit")
	w.quantity(header.GasLimit)
	w.key("gasUsed")
	w.quantity(header.GasUsed)
	w.key("hash")
	w.data(header.Hash)
	if header.L1BlockNumber != nil {
		w.key("l1BlockNumber")
		w.quantity(*header.L1BlockNumber)
	}
	w.key("logsBloom")
	w.data(header.LogsBloom)
	w.key("miner")
	w.data(header.Miner)
	if header.MixHash != nil {
		w.key("mixHash")
		w.data(header.MixHash)
	}
	if header.Nonce != nil {
		// A block nonce is 8-byte DATA, not a QUANTITY.
		var digits [16]byte
		d := strconv.AppendUint(digits[:0], *header.Nonce, 16)
		w.key("nonce")
		w.b = append(w.b, '"', '0', 'x')
		w.b = append(w.b, "0000000000000000"[len(d):]...)
		w.b = append(w.b, d...)
		w.b = append(w.b, '"')
	}
	w.key("number")
	w.quantity(header.Number)
	if header.ParentBeaconBlockRoot != nil {
		w.key("parentBeaconBlockRoot")
		w.data(header.ParentBeaconBlockRoot)
	}
	w.key("parentHash")
	w.data(header.ParentHash)
	if header.ProposerIndex != nil {
		w.key("proposerIndex")
		w.quantity(*header.ProposerIndex)
	}
	if header.ProposerPublicKey != nil {
		w.key("proposerPublicKey")
		w.text(*header.ProposerPublicKey)
	}
	w.key("receiptsRoot")
	w.data(header.ReceiptsRoot)
	if header.RequestsHash != nil {
		w.key("requestsHash")
		w.data(header.RequestsHash)
	}
	if header.SendCount != nil {
		w.key("sendCount")
		w.quantity(*header.SendCount)
	}
	if header.SendRoot != nil {
		w.key("sendRoot")
		w.data(header.SendRoot)
	}
	w.key("sha3Uncles")
	w.data(header.Sha3Uncles)
	w.key("size")
	w.quantity(header.Size)
	if header.Slot != nil {
		w.key("slot")
		w.quantity(*header.Slot)
	}
	w.key("stateRoot")
	w.data(header.StateRoot)
	w.key("timestamp")
	w.quantity(header.Timestamp)
	w.quantityField("totalDifficulty", header.TotalDifficulty)
	if header.TransactionCount != nil {
		w.key("transactionCount")
		w.quantity(uint64(*header.TransactionCount))
	}
	w.key("transactions")
	if len(fullTxs) > 0 {
		w.open('[')
		for _, tx := range fullTxs {
			w.sep()
			w.transaction(tx, sig)
		}
		w.close(']')
	} else {
		w.dataList(txHashes)
	}
	w.key("transactionsRoot")
	w.data(header.TransactionsRoot)
	w.key("uncles")
	w.dataList(header.Uncles)
	// A header with withdrawalsRoot always carries a withdrawals list, which
	// proto decodes as nil when empty; JSON-RPC nodes return [] for it.
	if withdrawals != nil || header.WithdrawalsRoot != nil {
		w.key("withdrawals")
		w.open('[')
		for _, wd := range withdrawals {
			w.sep()
			w.withdrawal(wd)
		}
		w.close(']')
	}
	if header.WithdrawalsRoot != nil {
		w.key("withdrawalsRoot")
		w.data(header.WithdrawalsRoot)
	}
	w.close('}')
	return w.b
}

// ParseJsonRpcTransaction parses a JSON-RPC transaction into a proto Transaction.
// This handles all transaction types including legacy, EIP-1559, EIP-2930, EIP-4844, and EIP-7702.
func ParseJsonRpcTransaction(txMap map[string]interface{}, header *BlockHeader) (*Transaction, error) {
	// Helper function to get string from map
	getString := func(key string) string {
		if v, ok := txMap[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	// Helper function to get optional string from map
	getOptionalString := func(key string) *string {
		if v, ok := txMap[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return &s
			}
		}
		return nil
	}

	// Parse required fields
	hash, err := HexToBytes(getString("hash"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse hash: %w", err)
	}

	nonce, err := NumberishToUint64(getString("nonce"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse nonce: %w", err)
	}

	from, err := HexToBytes(getString("from"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse from: %w", err)
	}

	var to []byte
	toStr := getString("to")
	if toStr != "" && toStr != "0x" {
		to, err = HexToBytes(toStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse to: %w", err)
		}
	}

	value := getString("value")
	if value == "" {
		value = "0"
	}

	input, err := HexToBytes(getString("input"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	gasLimit, err := NumberishToUint64(getString("gas"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse gas: %w", err)
	}

	r, err := HexToBytes(getString("r"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse r: %w", err)
	}

	sSig, err := HexToBytes(getString("s"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse s: %w", err)
	}

	var v []byte
	vStr := getString("v")
	if vStr != "" {
		v, err = HexToBytes(vStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse v: %w", err)
		}
	}

	typ, err := NumberishToUint32(getString("type"))
	if err != nil {
		// Default to legacy type 0 if not specified
		typ = 0
	}

	var chainId *uint64
	chainIdStr := getString("chainId")
	if chainIdStr != "" {
		cid, err := NumberishToUint64(chainIdStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chainId: %w", err)
		}
		chainId = &cid
	}

	var yParity *uint32
	yParityStr := getString("yParity")
	if yParityStr != "" {
		yp, err := NumberishToUint32(yParityStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse yParity: %w", err)
		}
		yParity = &yp
	}

	// Parse access list
	var accessList []*AccessListItem
	if alRaw, ok := txMap["accessList"]; ok {
		if alArray, ok := alRaw.([]interface{}); ok {
			for _, alItem := range alArray {
				if alMap, ok := alItem.(map[string]interface{}); ok {
					// Get address from access list item map
					alAddr := ""
					if addrVal, ok := alMap["address"]; ok {
						if addrStr, ok := addrVal.(string); ok {
							alAddr = addrStr
						}
					}

					addr, err := HexToBytes(alAddr)
					if err != nil {
						return nil, fmt.Errorf("failed to parse access list address: %w", err)
					}

					var storageKeys [][]byte
					if skRaw, ok := alMap["storageKeys"]; ok {
						if skArray, ok := skRaw.([]interface{}); ok {
							for _, sk := range skArray {
								if skStr, ok := sk.(string); ok {
									skBytes, err := HexToBytes(skStr)
									if err != nil {
										return nil, fmt.Errorf("failed to parse storage key: %w", err)
									}
									storageKeys = append(storageKeys, skBytes)
								}
							}
						}
					}

					accessList = append(accessList, &AccessListItem{
						Address:     addr,
						StorageKeys: storageKeys,
					})
				}
			}
		}
	}

	// Parse authorization list (EIP-7702)
	var authorizationList []*AuthorizationListItem
	if authRaw, ok := txMap["authorizationList"]; ok {
		if authArray, ok := authRaw.([]interface{}); ok {
			for _, authItem := range authArray {
				if authMap, ok := authItem.(map[string]interface{}); ok {
					// Helper to get string from auth map
					getAuthString := func(key string) string {
						if v, ok := authMap[key]; ok {
							if s, ok := v.(string); ok {
								return s
							}
						}
						return ""
					}

					// Missing or malformed is an error (0 here means "valid on
					// any chain", so it must not stand in for "unknown"), but a
					// well-formed value too wide for 64 bits falls back to 0 —
					// deliberate, with the ceiling written down in models.proto.
					authChainIdStr := getAuthString("chainId")
					if authChainIdStr == "" {
						return nil, fmt.Errorf("authorization chainId is missing or not a string")
					}
					if err := ValidateUint256Quantity(authChainIdStr); err != nil {
						return nil, fmt.Errorf("failed to parse authorization chainId: %w", err)
					}
					authChainId, err := NumberishToUint64(authChainIdStr)
					if err != nil {
						authChainId = 0
					}

					authAddress, err := HexToBytes(getAuthString("address"))
					if err != nil {
						return nil, fmt.Errorf("failed to parse authorization address: %w", err)
					}

					authNonce, err := NumberishToUint64(getAuthString("nonce"))
					if err != nil {
						return nil, fmt.Errorf("failed to parse authorization nonce: %w", err)
					}

					authR, err := HexToBytes(getAuthString("r"))
					if err != nil {
						return nil, fmt.Errorf("failed to parse authorization r: %w", err)
					}

					authS, err := HexToBytes(getAuthString("s"))
					if err != nil {
						return nil, fmt.Errorf("failed to parse authorization s: %w", err)
					}

					authYParity, err := NumberishToUint32(getAuthString("yParity"))
					if err != nil {
						return nil, fmt.Errorf("failed to parse authorization yParity: %w", err)
					}

					// Optional authority field
					var authAuthority []byte
					if authAuthorityStr := getAuthString("authority"); authAuthorityStr != "" {
						if b, err := HexToBytes(authAuthorityStr); err == nil {
							authAuthority = b
						} else {
							return nil, fmt.Errorf("failed to parse authorization authority: %w", err)
						}
					}

					authorizationList = append(authorizationList, &AuthorizationListItem{
						ChainId:   authChainId,
						Address:   authAddress,
						Nonce:     authNonce,
						R:         authR,
						S:         authS,
						YParity:   authYParity,
						Authority: authAuthority,
					})
				}
			}
		}
	}

	// Parse block context from header if provided
	var blockNumber *uint64
	var blockHash []byte
	var blockTimestamp *uint64
	var transactionIndex *uint32

	if header != nil {
		blockNumber = &header.Number
		blockHash = header.Hash
		blockTimestamp = &header.Timestamp
	}

	// Override with explicit block info if present in transaction
	if blockNumStr := getString("blockNumber"); blockNumStr != "" {
		bn, err := NumberishToUint64(blockNumStr)
		if err == nil {
			blockNumber = &bn
		}
	}

	if blockHashStr := getString("blockHash"); blockHashStr != "" {
		if bh, err := HexToBytes(blockHashStr); err == nil {
			blockHash = bh
		}
	}

	if txIndexStr := getString("transactionIndex"); txIndexStr != "" {
		ti, err := NumberishToUint32(txIndexStr)
		if err == nil {
			transactionIndex = &ti
		}
	}

	// Parse blob fields
	var blobVersionedHashes [][]byte
	if bvhRaw, ok := txMap["blobVersionedHashes"]; ok {
		if bvhArray, ok := bvhRaw.([]interface{}); ok {
			for _, bvh := range bvhArray {
				if bvhStr, ok := bvh.(string); ok {
					bvhBytes, err := HexToBytes(bvhStr)
					if err != nil {
						return nil, fmt.Errorf("failed to parse blob versioned hash: %w", err)
					}
					blobVersionedHashes = append(blobVersionedHashes, bvhBytes)
				}
			}
		}
	}

	// Build transaction
	tx := &Transaction{
		Hash:                 hash,
		Nonce:                nonce,
		From:                 from,
		To:                   to,
		Value:                value,
		Input:                input,
		Type:                 typ,
		GasLimit:             gasLimit,
		GasPrice:             getOptionalString("gasPrice"),
		MaxFeePerGas:         getOptionalString("maxFeePerGas"),
		MaxPriorityFeePerGas: getOptionalString("maxPriorityFeePerGas"),
		R:                    r,
		S:                    sSig,
		V:                    v,
		YParity:              yParity,
		ChainId:              chainId,
		BlockNumber:          blockNumber,
		BlockHash:            blockHash,
		TransactionIndex:     transactionIndex,
		BlockTimestamp:       blockTimestamp,
		AccessList:           accessList,
		BlobVersionedHashes:  blobVersionedHashes,
		AuthorizationList:    authorizationList,
		MaxFeePerBlobGas:     getOptionalString("maxFeePerBlobGas"),
	}

	// Add L2 fee fields
	tx.L1Fee = getOptionalString("l1Fee")
	tx.L1GasPrice = getOptionalString("l1GasPrice")
	tx.L1GasUsed = getOptionalString("l1GasUsed")

	if l1FeeScalarStr := getString("l1FeeScalar"); l1FeeScalarStr != "" {
		scl, err := strconv.ParseFloat(l1FeeScalarStr, 64)
		if err == nil {
			tx.L1FeeScalar = &scl
		}
	}

	tx.L1BlobBaseFee = getOptionalString("l1BlobBaseFee")

	if l1BlobBaseFeeScalarStr := getString("l1BlobBaseFeeScalar"); l1BlobBaseFeeScalarStr != "" {
		bscl, err := NumberishToUint64(l1BlobBaseFeeScalarStr)
		if err == nil {
			tx.L1BlobBaseFeeScalar = &bscl
		}
	}

	// Add gateway fee fields
	tx.GatewayFee = getOptionalString("gatewayFee")

	if fcStr := getString("feeCurrency"); fcStr != "" {
		fc, err := HexToBytes(fcStr)
		if err == nil {
			tx.FeeCurrency = fc
		}
	}

	if gfrStr := getString("gatewayFeeRecipient"); gfrStr != "" {
		gfr, err := HexToBytes(gfrStr)
		if err == nil {
			tx.GatewayFeeRecipient = gfr
		}
	}

	// Add Arbitrum retryable ticket fields
	if benStr := getString("beneficiary"); benStr != "" {
		beneficiary, err := HexToBytes(benStr)
		if err == nil {
			tx.Beneficiary = beneficiary
		}
	}

	tx.DepositValue = getOptionalString("depositValue")
	tx.L1BaseFee = getOptionalString("l1BaseFee")
	tx.MaxSubmissionFee = getOptionalString("maxSubmissionFee")

	if refundToStr := getString("refundTo"); refundToStr != "" {
		refundTo, err := HexToBytes(refundToStr)
		if err == nil {
			tx.RefundTo = refundTo
		}
	}

	if requestIdStr := getString("requestId"); requestIdStr != "" {
		requestId, err := HexToBytes(requestIdStr)
		if err == nil {
			tx.RequestId = requestId
		}
	}

	if retryDataStr := getString("retryData"); retryDataStr != "" {
		retryData, err := HexToBytes(retryDataStr)
		if err == nil {
			tx.RetryData = retryData
		}
	}

	if retryToStr := getString("retryTo"); retryToStr != "" {
		retryTo, err := HexToBytes(retryToStr)
		if err == nil {
			tx.RetryTo = retryTo
		}
	}

	tx.RetryValue = getOptionalString("retryValue")
	tx.MaxRefund = getOptionalString("maxRefund")
	tx.SubmissionFeeRefund = getOptionalString("submissionFeeRefund")

	if ticketIdStr := getString("ticketId"); ticketIdStr != "" {
		ticketId, err := HexToBytes(ticketIdStr)
		if err == nil {
			tx.TicketId = ticketId
		}
	}

	// Parse execution result fields (only available for mined transactions)
	if gasUsedStr := getString("gasUsed"); gasUsedStr != "" {
		gasUsed, err := NumberishToUint64(gasUsedStr)
		if err == nil {
			tx.GasUsed = &gasUsed
		}
	}

	tx.EffectiveGasPrice = getOptionalString("effectiveGasPrice")

	// Receipt fields of a merged transaction+receipt object (MIP-16 shape).
	if statusStr := getString("status"); statusStr != "" {
		status, err := NumberishToUint32(statusStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse status: %w", err)
		}
		tx.Status = &status
	}
	if cumulativeStr := getString("cumulativeGasUsed"); cumulativeStr != "" {
		cumulative, err := NumberishToUint64(cumulativeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse cumulativeGasUsed: %w", err)
		}
		tx.CumulativeGasUsed = &cumulative
	}
	if addrStr := getString("contractAddress"); addrStr != "" && addrStr != "0x" {
		if tx.ContractAddress, err = HexToBytes(addrStr); err != nil {
			return nil, fmt.Errorf("failed to parse contractAddress: %w", err)
		}
	}
	if bloomStr := getString("logsBloom"); bloomStr != "" {
		if tx.LogsBloom, err = HexToBytes(bloomStr); err != nil {
			return nil, fmt.Errorf("failed to parse logsBloom: %w", err)
		}
	}

	// Parse blob fields
	if blobGasUsedStr := getString("blobGasUsed"); blobGasUsedStr != "" {
		blobGasUsed, err := NumberishToUint64(blobGasUsedStr)
		if err == nil {
			tx.BlobGasUsed = &blobGasUsed
		}
	}

	tx.BlobGasPrice = getOptionalString("blobGasPrice")

	// Parse Base-specific fields
	if isSystemTxRaw, ok := txMap["isSystemTx"]; ok {
		if isSystemTx, ok := isSystemTxRaw.(bool); ok {
			tx.IsSystemTx = &isSystemTx
		}
	}

	tx.DepositReceiptVersion = getOptionalString("depositReceiptVersion")

	// Additional L2/Base fields
	if sh := getString("sourceHash"); sh != "" {
		if b, err := HexToBytes(sh); err == nil {
			tx.SourceHash = b
		}
	}
	// Mint value (decimal or hex string)
	tx.Mint = getOptionalString("mint")

	return tx, nil
}

// MergeReceipt copies the execution results of r, the receipt of tx, onto tx:
// status, gasUsed, cumulativeGasUsed, effectiveGasPrice, contractAddress,
// logsBloom and the blob gas figures. MIP-16 transaction objects carry these
// receipt fields, so a server merges the receipt before it renders a query
// response. r's optional values that are absent leave tx unchanged.
func MergeReceipt(tx *Transaction, r *Receipt) {
	if tx == nil || r == nil {
		return
	}
	if r.Status != nil {
		status := *r.Status
		tx.Status = &status
	}
	gasUsed, cumulative := r.GasUsed, r.CumulativeGasUsed
	tx.GasUsed, tx.CumulativeGasUsed = &gasUsed, &cumulative
	if r.EffectiveGasPrice != "" {
		price := r.EffectiveGasPrice
		tx.EffectiveGasPrice = &price
	}
	tx.ContractAddress = r.ContractAddress
	tx.LogsBloom = r.LogsBloom
	if r.BlobGasUsed != nil {
		blobGasUsed := *r.BlobGasUsed
		tx.BlobGasUsed = &blobGasUsed
	}
	if r.BlobGasPrice != nil {
		blobGasPrice := *r.BlobGasPrice
		tx.BlobGasPrice = &blobGasPrice
	}
}

// ParseJsonRpcWithdrawals parses a list of JSON-RPC withdrawals into a list of proto withdrawals.
// This is useful when constructing a evm.Block with withdrawals.
func ParseJsonRpcWithdrawals(withdrawals []*JsonRpcWithdrawal) ([]*Withdrawal, error) {
	protoWithdrawals := make([]*Withdrawal, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		protoWithdrawal, err := withdrawal.ToProto()
		if err != nil {
			return nil, fmt.Errorf("failed to parse withdrawal: %w", err)
		}
		protoWithdrawals = append(protoWithdrawals, protoWithdrawal)
	}
	return protoWithdrawals, nil
}

// TraceFromParity converts one trace_block / trace_transaction frame. A
// suicide frame takes from, value and to from the action's address, balance
// and refundAddress.
//
// A flat Parity frame does not say whether a parent failed, so Reverted here
// reflects only the frame's own error. After converting the frames of a
// block or transaction, call PropagateParityReverted on all of them.
func TraceFromParity(raw map[string]interface{}, blockNumber uint64, blockHash []byte, blockTimestamp *uint64) (*Trace, error) {
	action, _ := raw["action"].(map[string]interface{})
	result, _ := raw["result"].(map[string]interface{})

	rawType := stringValue(raw["type"])
	traceType := traceTypeFromString(rawType)
	callType := traceCallTypeFromString(stringValue(action["callType"]))
	fromKey, valueKey := "from", "value"
	if traceType == TraceType_TRACE_SELFDESTRUCT {
		fromKey, valueKey = "address", "balance"
	}
	from, err := hexBytesOrEmpty(stringValue(action[fromKey]))
	if err != nil {
		return nil, err
	}

	var to []byte
	for _, candidate := range []string{
		stringValue(action["to"]),
		stringValue(result["address"]),
		stringValue(action["refundAddress"]),
		stringValue(action["author"]),
	} {
		if candidate == "" {
			continue
		}
		to, err = HexToBytes(candidate)
		if err != nil {
			return nil, err
		}
		break
	}

	value := normalizeNumberishString(stringValue(action[valueKey]))
	input, err := hexBytesOrEmpty(stringValue(action["input"]))
	if err != nil {
		return nil, err
	}
	output, err := hexBytesOrEmpty(stringValue(result["output"]))
	if err != nil {
		return nil, err
	}
	gas, _ := numberishUint64(stringValue(action["gas"]))
	gasUsed, _ := numberishUint64(stringValue(result["gasUsed"]))
	subtraces, _ := numberishUint32FromAny(raw["subtraces"])
	traceAddress := uint32SliceFromAny(raw["traceAddress"])
	transactionHash, err := hexBytesOrEmpty(stringValue(raw["transactionHash"]))
	if err != nil {
		return nil, err
	}
	transactionIndex, _ := numberishUint32FromAny(raw["transactionPosition"])
	if transactionIndex == 0 {
		transactionIndex, _ = numberishUint32FromAny(raw["transactionIndex"])
	}

	trace := &Trace{
		Type:             parityFrameType(rawType, action),
		TraceType:        traceType,
		CallType:         callType,
		From:             from,
		To:               to,
		Value:            value,
		Input:            input,
		Output:           output,
		Gas:              gas,
		GasUsed:          gasUsed,
		Subtraces:        subtraces,
		TraceAddress:     traceAddress,
		TransactionHash:  transactionHash,
		TransactionIndex: transactionIndex,
		BlockNumber:      blockNumber,
		BlockHash:        blockHash,
		BlockTimestamp:   blockTimestamp,
	}
	if errText := stringValue(raw["error"]); errText != "" {
		trace.Error = &errText
		trace.Reverted = true
	}

	return trace, nil
}

// PropagateParityReverted sets Reverted on every frame whose transaction has
// an errored frame at the same or a parent traceAddress: MIP-16 defines a
// frame as reverted when it or one of its parents failed. traces may hold the
// frames of several transactions and blocks, in any order. Frames from
// TraceFromGethDebug already carry this, and a second pass changes nothing.
func PropagateParityReverted(traces []*Trace) {
	for i, reverted := range revertedFrames(traces) {
		if reverted {
			traces[i].Reverted = true
		}
	}
}

// revertedFrames reports, for each frame, whether it or a parent frame of the
// same transaction has an error.
func revertedFrames(traces []*Trace) []bool {
	type txKey struct {
		blockNumber uint64
		hash        string
		index       uint32
	}
	keyOf := func(t *Trace) txKey {
		return txKey{t.BlockNumber, string(t.TransactionHash), t.TransactionIndex}
	}
	errored := make(map[txKey][][]uint32)
	for _, t := range traces {
		if t != nil && t.Error != nil {
			k := keyOf(t)
			errored[k] = append(errored[k], t.TraceAddress)
		}
	}
	out := make([]bool, len(traces))
	for i, t := range traces {
		if t == nil {
			continue
		}
		out[i] = t.Reverted
		for _, parent := range errored[keyOf(t)] {
			if out[i] {
				break
			}
			out[i] = isTraceAddressPrefix(parent, t.TraceAddress)
		}
	}
	return out
}

func isTraceAddressPrefix(prefix, address []uint32) bool {
	if len(prefix) > len(address) {
		return false
	}
	for i, v := range prefix {
		if address[i] != v {
			return false
		}
	}
	return true
}

// TraceFromGethDebug flattens one callTracer trace into frames in traceAddress
// order. raw is a debug_traceBlockByNumber / debug_traceBlockByHash item,
// {"txHash": ..., "result": {callTracer frame}}; its txHash and the
// transactionIndex the caller passes are set on every frame. To convert a
// debug_traceTransaction result, wrap it as {"txHash": hash, "result": result}.
// A frame is Reverted when it or one of its parents has an error.
func TraceFromGethDebug(raw map[string]interface{}, transactionIndex uint32, blockNumber uint64, blockHash []byte, blockTimestamp *uint64) ([]*Trace, error) {
	txHash, err := HexToBytes(stringValue(raw["txHash"]))
	if err != nil || len(txHash) != HashLength {
		return nil, fmt.Errorf("callTracer item needs a 32-byte txHash, got %q", stringValue(raw["txHash"]))
	}
	result, ok := raw["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("callTracer item for %s has no result frame", BytesToHex(txHash))
	}
	flattened := make([]*Trace, 0, 8)
	var walk func(map[string]interface{}, []uint32, bool) error
	walk = func(node map[string]interface{}, traceAddress []uint32, parentReverted bool) error {
		traceType := traceTypeFromString(stringValue(node["type"]))
		callType := traceCallTypeFromString(stringValue(node["type"]))
		from, err := hexBytesOrEmpty(stringValue(node["from"]))
		if err != nil {
			return err
		}
		to, err := hexBytesOrEmpty(stringValue(node["to"]))
		if err != nil {
			return err
		}
		input, err := hexBytesOrEmpty(stringValue(node["input"]))
		if err != nil {
			return err
		}
		output, err := hexBytesOrEmpty(stringValue(node["output"]))
		if err != nil {
			return err
		}
		gas, _ := numberishUint64(stringValue(node["gas"]))
		gasUsed, _ := numberishUint64(stringValue(node["gasUsed"]))
		value := normalizeNumberishString(stringValue(node["value"]))
		calls, _ := node["calls"].([]interface{})
		trace := &Trace{
			Type:             strings.ToUpper(stringValue(node["type"])),
			TraceType:        traceType,
			CallType:         callType,
			From:             from,
			To:               to,
			Value:            value,
			Input:            input,
			Output:           output,
			Gas:              gas,
			GasUsed:          gasUsed,
			Subtraces:        uint32(len(calls)),
			TraceAddress:     append([]uint32(nil), traceAddress...),
			TransactionHash:  txHash,
			TransactionIndex: transactionIndex,
			BlockNumber:      blockNumber,
			BlockHash:        blockHash,
			BlockTimestamp:   blockTimestamp,
		}
		if errText := stringValue(node["error"]); errText != "" {
			trace.Error = &errText
		}
		trace.Reverted = parentReverted || trace.Error != nil
		flattened = append(flattened, trace)
		for i, rawChild := range calls {
			child, ok := rawChild.(map[string]interface{})
			if !ok {
				continue
			}
			if err := walk(child, append(append([]uint32(nil), traceAddress...), uint32(i)), trace.Reverted); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(result, nil, false); err != nil {
		return nil, err
	}
	return flattened, nil
}

// NativeTransfersFromTraces returns the MIP-16 transfers of traces: every call
// frame whose value is above zero and whose type is not DELEGATECALL or
// CALLCODE, which move no value. Each transfer carries the full trace context.
// A transfer is reverted when its frame or a parent frame of the same
// transaction in traces failed, so Parity frames need no separate pass here.
// Reverted transfers are included; the server applies includeReverted.
func NativeTransfersFromTraces(traces []*Trace) []*NativeTransfer {
	transfers := make([]*NativeTransfer, 0, len(traces))
	reverted := revertedFrames(traces)
	value := new(big.Int)
	for i, trace := range traces {
		if trace == nil || trace.TraceType == TraceType_TRACE_REWARD {
			continue
		}
		typ := traceFrameType(trace)
		if typ == "DELEGATECALL" || typ == "CALLCODE" || trace.Value == "" {
			continue
		}
		if strings.HasPrefix(trace.Value, "0x") || strings.HasPrefix(trace.Value, "0X") {
			if _, ok := value.SetString(RemoveHexPrefix(trace.Value), 16); !ok {
				continue
			}
		} else if _, ok := value.SetString(trace.Value, 10); !ok {
			continue
		}
		if value.Sign() <= 0 {
			continue
		}
		transfers = append(transfers, &NativeTransfer{
			Type:             typ,
			From:             trace.From,
			To:               trace.To,
			Value:            trace.Value,
			Gas:              trace.Gas,
			GasUsed:          trace.GasUsed,
			Input:            trace.Input,
			Output:           trace.Output,
			Error:            trace.Error,
			Reverted:         reverted[i],
			TransactionHash:  trace.TransactionHash,
			TransactionIndex: trace.TransactionIndex,
			BlockNumber:      trace.BlockNumber,
			BlockHash:        trace.BlockHash,
			TraceAddress:     append([]uint32(nil), trace.TraceAddress...),
			BlockTimestamp:   trace.BlockTimestamp,
		})
	}
	return transfers
}

// traceFrameType returns the callTracer type of a frame (CALL, CALLCODE,
// DELEGATECALL, STATICCALL, CREATE, CREATE2 or SELFDESTRUCT). It is Type
// when set, else derived from TraceType and CallType; a reward is "".
func traceFrameType(t *Trace) string {
	if t.Type != "" {
		return t.Type
	}
	switch t.TraceType {
	case TraceType_TRACE_CREATE:
		return "CREATE"
	case TraceType_TRACE_SELFDESTRUCT:
		return "SELFDESTRUCT"
	case TraceType_TRACE_REWARD:
		return ""
	}
	switch t.CallType {
	case TraceCallType_TRACE_CALL_STATICCALL:
		return "STATICCALL"
	case TraceCallType_TRACE_CALL_DELEGATECALL:
		return "DELEGATECALL"
	case TraceCallType_TRACE_CALL_CALLCODE:
		return "CALLCODE"
	}
	return "CALL"
}

// parityFrameType maps a Parity trace type and action to the callTracer type.
func parityFrameType(traceType string, action map[string]interface{}) string {
	switch strings.ToLower(traceType) {
	case "call":
		if callType := stringValue(action["callType"]); callType != "" {
			return strings.ToUpper(callType)
		}
		return "CALL"
	case "create":
		if method := stringValue(action["creationMethod"]); method != "" {
			return strings.ToUpper(method)
		}
		return "CREATE"
	case "suicide", "selfdestruct":
		return "SELFDESTRUCT"
	}
	return ""
}

func stringValue(v interface{}) string {
	s, _ := v.(string)
	return s
}

func hexBytesOrEmpty(v string) ([]byte, error) {
	if v == "" || v == "0x" {
		return nil, nil
	}
	return HexToBytes(v)
}

func normalizeNumberishString(v string) string {
	if v == "" {
		return "0x0"
	}
	if normalized, err := DecimalStringToHex(v); err == nil {
		return normalized
	}
	return v
}

func numberishUint64(v string) (uint64, error) {
	if v == "" {
		return 0, nil
	}
	return NumberishToUint64(v)
}

func numberishUint32FromAny(v interface{}) (uint32, error) {
	switch t := v.(type) {
	case string:
		if t == "" {
			return 0, nil
		}
		return NumberishToUint32(t)
	case float64:
		return uint32(t), nil
	case int:
		return uint32(t), nil
	default:
		return 0, nil
	}
}

func uint32SliceFromAny(v interface{}) []uint32 {
	items, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]uint32, 0, len(items))
	for _, item := range items {
		if n, err := numberishUint32FromAny(item); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func traceTypeFromString(v string) TraceType {
	switch strings.ToLower(v) {
	case "call":
		return TraceType_TRACE_CALL
	case "create":
		return TraceType_TRACE_CREATE
	case "suicide", "selfdestruct":
		return TraceType_TRACE_SELFDESTRUCT
	case "reward":
		return TraceType_TRACE_REWARD
	default:
		return TraceType_TRACE_CALL
	}
}

func traceCallTypeFromString(v string) TraceCallType {
	switch strings.ToLower(v) {
	case "staticcall":
		return TraceCallType_TRACE_CALL_STATICCALL
	case "delegatecall":
		return TraceCallType_TRACE_CALL_DELEGATECALL
	case "callcode":
		return TraceCallType_TRACE_CALL_CALLCODE
	default:
		return TraceCallType_TRACE_CALL_CALL
	}
}
