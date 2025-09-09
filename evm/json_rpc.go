package evm

import (
	"fmt"
	"strconv"
)

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
	BlockHash             string        `json:"blockHash"`
	BlockNumber           string        `json:"blockNumber"`
	BlockTimestamp        string        `json:"blockTimestamp"`
	ContractAddress       string        `json:"contractAddress"`
	CumulativeGasUsed     string        `json:"cumulativeGasUsed"`
	EffectiveGasPrice     string        `json:"effectiveGasPrice"`
	From                  string        `json:"from"`
	GasUsed               string        `json:"gasUsed"`
	Logs                  []*JsonRpcLog `json:"logs"`
	LogsBloom             string        `json:"logsBloom"`
	Root                  string        `json:"root"`
	Status                string        `json:"status"`
	To                    string        `json:"to"`
	TransactionHash       string        `json:"transactionHash"`
	TransactionIndex      string        `json:"transactionIndex"`
	Type                  string        `json:"type"`
	BlobGasUsed           string        `json:"blobGasUsed"`
	BlobGasPrice          string        `json:"blobGasPrice"`
	L1Fee                 string        `json:"l1Fee"`
	L1GasUsed             string        `json:"l1GasUsed"`
	L1GasPrice            string        `json:"l1GasPrice"`
	L1FeeScalar           string        `json:"l1FeeScalar"`
	L1BaseFeeScalar       string        `json:"l1BaseFeeScalar"`
	L1BlobBaseFee         string        `json:"l1BlobBaseFee"`
	L1BlobBaseFeeScalar   string        `json:"l1BlobBaseFeeScalar"`
	GasUsedForL1          string        `json:"gasUsedForL1"`
	L1BlockNumber         string        `json:"l1BlockNumber"`
	GatewayFee            string        `json:"gatewayFee"`
	DepositNonce          string        `json:"depositNonce"`
	DepositReceiptVersion string        `json:"depositReceiptVersion"`
	Timeboosted           *bool         `json:"timeboosted"`
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
		bt, err := NumberishToUint64(r.BlockTimestamp)
		if err == nil {
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
		L1BlobBaseFeeScalar:   l1BlobBaseFeeScalar,
		GasUsedForL1:          gasUsedForL1,
		L1BlockNumber:         l1BlockNumber,
		GatewayFee:            gatewayFee,
		DepositNonce:          depositNonce,
		DepositReceiptVersion: depositReceiptVersion,
		Timeboosted:           timeboosted,
	}, nil
}

type JsonRpcLog struct {
	Address          string   `json:"address"`
	BlockHash        string   `json:"blockHash"`
	BlockNumber      string   `json:"blockNumber"`
	BlockTimestamp   string   `json:"blockTimestamp"`
	Data             string   `json:"data"`
	LogIndex         string   `json:"logIndex"`
	Topics           []string `json:"topics"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
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
		u, err := NumberishToUint64(l.BlockTimestamp)
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

// LogToJsonRpc serialises a *Log into a JSON-RPC compatible map[string]interface{}.
func LogToJsonRpc(l *Log) map[string]interface{} {
	if l == nil {
		return nil
	}
	topics := make([]string, len(l.Topics))
	for i, t := range l.Topics {
		topics[i] = BytesToHex(t)
	}

	result := map[string]interface{}{
		"address":          BytesToHex(l.Address),
		"topics":           topics,
		"data":             BytesToHex(l.Data),
		"blockNumber":      fmt.Sprintf("0x%x", l.BlockNumber),
		"transactionHash":  BytesToHex(l.TransactionHash),
		"transactionIndex": fmt.Sprintf("0x%x", l.TransactionIndex),
		"blockHash":        BytesToHex(l.BlockHash),
		"logIndex":         fmt.Sprintf("0x%x", l.LogIndex),
		"removed":          false,
	}

	// Include optional block timestamp when present
	if l.BlockTimestamp != nil {
		result["blockTimestamp"] = fmt.Sprintf("0x%x", *l.BlockTimestamp)
	}

	return result
}

// LogsToJsonRpc converts a slice of *Log to []interface{} for JSON-RPC responses.
func LogsToJsonRpc(logs []*Log) []interface{} {
	out := make([]interface{}, len(logs))
	for i, l := range logs {
		out[i] = LogToJsonRpc(l)
	}
	return out
}

// TransactionToJsonRpc converts a *Transaction into JSON-RPC representation.
func TransactionToJsonRpc(tx *Transaction) map[string]interface{} {
	if tx == nil {
		return nil
	}

	o := map[string]interface{}{
		"hash":  BytesToHex(tx.Hash),
		"nonce": fmt.Sprintf("0x%x", tx.Nonce),
		"from":  BytesToHex(tx.From),
		"gas":   fmt.Sprintf("0x%x", tx.GasLimit),
		"input": BytesToHex(tx.Input),
		"type":  fmt.Sprintf("0x%x", tx.Type),
	}

	if len(tx.To) > 0 {
		o["to"] = BytesToHex(tx.To)
	} else {
		o["to"] = nil
	}

	if tx.BlockHash != nil {
		o["blockHash"] = BytesToHex(tx.BlockHash)
	}
	if tx.BlockNumber != nil {
		o["blockNumber"] = fmt.Sprintf("0x%x", *tx.BlockNumber)
	}
	if tx.TransactionIndex != nil {
		o["transactionIndex"] = fmt.Sprintf("0x%x", *tx.TransactionIndex)
	}
	if tx.BlockTimestamp != nil {
		o["blockTimestamp"] = fmt.Sprintf("0x%x", *tx.BlockTimestamp)
	}

	// Value (decimal -> hex)
	if tx.Value != "" {
		if hex, err := DecimalStringToHex(tx.Value); err == nil {
			o["value"] = hex
		}
	}

	// Fee related fields – pointers to decimal strings
	if tx.GasPrice != nil {
		if hex, err := DecimalStringToHex(*tx.GasPrice); err == nil {
			o["gasPrice"] = hex
		}
	}
	if tx.MaxFeePerGas != nil {
		if hex, err := DecimalStringToHex(*tx.MaxFeePerGas); err == nil {
			o["maxFeePerGas"] = hex
		}
	}
	if tx.MaxPriorityFeePerGas != nil {
		if hex, err := DecimalStringToHex(*tx.MaxPriorityFeePerGas); err == nil {
			o["maxPriorityFeePerGas"] = hex
		}
	}

	// Execution result fields (only available for mined transactions)
	if tx.GasUsed != nil {
		o["gasUsed"] = fmt.Sprintf("0x%x", *tx.GasUsed)
	}
	if tx.EffectiveGasPrice != nil {
		if hex, err := DecimalStringToHex(*tx.EffectiveGasPrice); err == nil {
			o["effectiveGasPrice"] = hex
		}
	}

	if tx.R != nil {
		// r is 32-byte DATA; keep leading zeros
		o["r"] = BytesToHexFixed(tx.R, 32)
	}
	if tx.S != nil {
		// s is 32-byte DATA; keep leading zeros
		o["s"] = BytesToHexFixed(tx.S, 32)
	}
	if tx.V != nil {
		// v is QUANTITY
		o["v"] = BytesToQuantityHex(tx.V)
	}

	// Chain ID (optional)
	if tx.ChainId != nil {
		o["chainId"] = fmt.Sprintf("0x%x", *tx.ChainId)
	} else {
		o["chainId"] = nil
	}

	// yParity (optional)
	if tx.YParity != nil {
		o["yParity"] = fmt.Sprintf("0x%x", *tx.YParity)
	} else {
		o["yParity"] = nil
	}

	// Access list (EIP-2930)
	accessList := make([]interface{}, 0, len(tx.AccessList))
	for _, item := range tx.AccessList {
		if item == nil {
			continue
		}
		obj := map[string]interface{}{
			"address": BytesToHex(item.Address),
		}
		if len(item.StorageKeys) > 0 {
			keys := make([]string, len(item.StorageKeys))
			for i, k := range item.StorageKeys {
				keys[i] = BytesToHex(k)
			}
			obj["storageKeys"] = keys
		} else {
			obj["storageKeys"] = []interface{}{}
		}
		accessList = append(accessList, obj)
	}
	o["accessList"] = accessList

	// EIP-4844 fields
	if tx.MaxFeePerBlobGas != nil {
		if hex, err := DecimalStringToHex(*tx.MaxFeePerBlobGas); err == nil {
			o["maxFeePerBlobGas"] = hex
		}
	}
	if len(tx.BlobVersionedHashes) > 0 {
		hashes := make([]string, len(tx.BlobVersionedHashes))
		for i, h := range tx.BlobVersionedHashes {
			hashes[i] = BytesToHex(h)
		}
		o["blobVersionedHashes"] = hashes
	}
	if tx.BlobGasUsed != nil {
		o["blobGasUsed"] = fmt.Sprintf("0x%x", *tx.BlobGasUsed)
	}
	if tx.BlobGasPrice != nil {
		if hex, err := DecimalStringToHex(*tx.BlobGasPrice); err == nil {
			o["blobGasPrice"] = hex
		}
	}

	// EIP-7702 authorization list
	if len(tx.AuthorizationList) > 0 {
		authList := make([]interface{}, 0, len(tx.AuthorizationList))
		for _, auth := range tx.AuthorizationList {
			if auth == nil {
				continue
			}
			authItem := map[string]interface{}{
				"chainId": fmt.Sprintf("0x%x", auth.ChainId),
				"address": BytesToHex(auth.Address),
				"nonce":   fmt.Sprintf("0x%x", auth.Nonce),
				"r":       BytesToHexFixed(auth.R, 32),
				"s":       BytesToHexFixed(auth.S, 32),
				"yParity": fmt.Sprintf("0x%x", auth.YParity),
			}
			// Optional authority (bytes) – include when present
			if len(auth.Authority) > 0 {
				authItem["authority"] = BytesToHex(auth.Authority)
			}
			authList = append(authList, authItem)
		}
		o["authorizationList"] = authList
	}

	// L2 fee breakdown & miscellaneous fields
	if tx.L1Fee != nil {
		if hex, err := DecimalStringToHex(*tx.L1Fee); err == nil {
			o["l1Fee"] = hex
		} else {
			o["l1Fee"] = nil
		}
	} else {
		o["l1Fee"] = nil
	}
	if tx.L1GasUsed != nil {
		if hex, err := DecimalStringToHex(*tx.L1GasUsed); err == nil {
			o["l1GasUsed"] = hex
		}
	}
	if tx.L1GasPrice != nil {
		if hex, err := DecimalStringToHex(*tx.L1GasPrice); err == nil {
			o["l1GasPrice"] = hex
		}
	}
	if tx.L1FeeScalar != nil {
		o["l1FeeScalar"] = *tx.L1FeeScalar
	}
	if tx.L1BlobBaseFee != nil {
		if hex, err := DecimalStringToHex(*tx.L1BlobBaseFee); err == nil {
			o["l1BlobBaseFee"] = hex
		}
	}
	if tx.L1BlobBaseFeeScalar != nil {
		o["l1BlobBaseFeeScalar"] = fmt.Sprintf("0x%x", *tx.L1BlobBaseFeeScalar)
	}
	if tx.GatewayFee != nil {
		if hex, err := DecimalStringToHex(*tx.GatewayFee); err == nil {
			o["gatewayFee"] = hex
		}
	}
	if len(tx.FeeCurrency) > 0 {
		o["feeCurrency"] = BytesToHex(tx.FeeCurrency)
	}
	if len(tx.GatewayFeeRecipient) > 0 {
		o["gatewayFeeRecipient"] = BytesToHex(tx.GatewayFeeRecipient)
	}

	// Arbitrum retryable ticket fields
	if len(tx.Beneficiary) > 0 {
		o["beneficiary"] = BytesToHex(tx.Beneficiary)
	}
	if tx.DepositValue != nil {
		if hex, err := DecimalStringToHex(*tx.DepositValue); err == nil {
			o["depositValue"] = hex
		}
	}
	if tx.L1BaseFee != nil {
		if hex, err := DecimalStringToHex(*tx.L1BaseFee); err == nil {
			o["l1BaseFee"] = hex
		}
	}
	if tx.MaxSubmissionFee != nil {
		if hex, err := DecimalStringToHex(*tx.MaxSubmissionFee); err == nil {
			o["maxSubmissionFee"] = hex
		}
	}
	if len(tx.RefundTo) > 0 {
		o["refundTo"] = BytesToHex(tx.RefundTo)
	}
	if len(tx.RequestId) > 0 {
		o["requestId"] = BytesToHex(tx.RequestId)
	}
	if len(tx.RetryData) > 0 {
		o["retryData"] = BytesToHex(tx.RetryData)
	}
	if len(tx.RetryTo) > 0 {
		o["retryTo"] = BytesToHex(tx.RetryTo)
	}
	if tx.RetryValue != nil {
		if hex, err := DecimalStringToHex(*tx.RetryValue); err == nil {
			o["retryValue"] = hex
		}
	}
	if tx.MaxRefund != nil {
		if hex, err := DecimalStringToHex(*tx.MaxRefund); err == nil {
			o["maxRefund"] = hex
		}
	}
	if tx.SubmissionFeeRefund != nil {
		if hex, err := DecimalStringToHex(*tx.SubmissionFeeRefund); err == nil {
			o["submissionFeeRefund"] = hex
		}
	}
	if len(tx.TicketId) > 0 {
		o["ticketId"] = BytesToHex(tx.TicketId)
	}

	// Base-specific fields
	if tx.IsSystemTx != nil {
		o["isSystemTx"] = *tx.IsSystemTx
	}
	if tx.DepositReceiptVersion != nil {
		if hex, err := DecimalStringToHex(*tx.DepositReceiptVersion); err == nil {
			o["depositReceiptVersion"] = hex
		}
	}

	return o
}

// ReceiptsToJsonRpc converts a slice of *Receipt to []interface{} for JSON-RPC responses.
func ReceiptsToJsonRpc(receipts []*Receipt) []interface{} {
	out := make([]interface{}, len(receipts))
	for i, r := range receipts {
		out[i] = ReceiptToJsonRpc(r)
	}
	return out
}

// ReceiptToJsonRpc converts a *Receipt into JSON-RPC representation.
func ReceiptToJsonRpc(r *Receipt) map[string]interface{} {
	if r == nil {
		return nil
	}

	out := map[string]interface{}{
		"transactionHash":   BytesToHex(r.TransactionHash),
		"transactionIndex":  fmt.Sprintf("0x%x", r.TransactionIndex),
		"blockHash":         BytesToHex(r.BlockHash),
		"blockNumber":       fmt.Sprintf("0x%x", r.BlockNumber),
		"from":              BytesToHex(r.From),
		"cumulativeGasUsed": fmt.Sprintf("0x%x", r.CumulativeGasUsed),
		"gasUsed":           fmt.Sprintf("0x%x", r.GasUsed),
		"logsBloom":         BytesToHex(r.LogsBloom),
		"logs":              LogsToJsonRpc(r.Logs),
		"contractAddress":   nil,
	}

	// Add "to" field - can be null for contract creation
	if len(r.To) > 0 {
		out["to"] = BytesToHex(r.To)
	} else {
		out["to"] = nil
	}

	if r.Status != nil {
		out["status"] = fmt.Sprintf("0x%x", *r.Status)
	}
	if len(r.ContractAddress) > 0 {
		out["contractAddress"] = BytesToHex(r.ContractAddress)
	}
	out["type"] = fmt.Sprintf("0x%x", r.Type)

	if r.EffectiveGasPrice != "" {
		if hex, err := DecimalStringToHex(r.EffectiveGasPrice); err == nil {
			out["effectiveGasPrice"] = hex
		}
	}

	// Add root field for pre-Byzantium compatibility
	if len(r.Root) > 0 {
		out["root"] = BytesToHex(r.Root)
	}

	// Add blockTimestamp
	if r.BlockTimestamp != nil {
		out["blockTimestamp"] = fmt.Sprintf("0x%x", *r.BlockTimestamp)
	}

	if r.GasUsedForL1 != nil {
		out["gasUsedForL1"] = fmt.Sprintf("0x%x", *r.GasUsedForL1)
	}

	if r.L1BlockNumber != nil {
		out["l1BlockNumber"] = fmt.Sprintf("0x%x", *r.L1BlockNumber)
	}

	// Additional L2 fee breakdown fields
	if r.L1Fee != nil {
		if hex, err := DecimalStringToHex(*r.L1Fee); err == nil {
			out["l1Fee"] = hex
		}
	} else {
		out["l1Fee"] = nil
	}
	if r.L1GasUsed != nil {
		if hex, err := DecimalStringToHex(*r.L1GasUsed); err == nil {
			out["l1GasUsed"] = hex
		}
	} else {
		out["l1GasUsed"] = nil
	}
	if r.L1GasPrice != nil {
		if hex, err := DecimalStringToHex(*r.L1GasPrice); err == nil {
			out["l1GasPrice"] = hex
		}
	} else {
		out["l1GasPrice"] = nil
	}
	if r.GatewayFee != nil {
		if hex, err := DecimalStringToHex(*r.GatewayFee); err == nil {
			out["gatewayFee"] = hex
		}
	}

	if r.BlobGasUsed != nil {
		out["blobGasUsed"] = fmt.Sprintf("0x%x", *r.BlobGasUsed)
	}
	if r.BlobGasPrice != nil {
		if hex, err := DecimalStringToHex(*r.BlobGasPrice); err == nil {
			out["blobGasPrice"] = hex
		}
	}
	if r.L1FeeScalar != nil {
		out["l1FeeScalar"] = *r.L1FeeScalar
	}
	if r.L1BaseFeeScalar != nil {
		out["l1BaseFeeScalar"] = fmt.Sprintf("0x%x", *r.L1BaseFeeScalar)
	}
	if r.L1BlobBaseFee != nil {
		if hex, err := DecimalStringToHex(*r.L1BlobBaseFee); err == nil {
			out["l1BlobBaseFee"] = hex
		}
	}
	if r.L1BlobBaseFeeScalar != nil {
		out["l1BlobBaseFeeScalar"] = fmt.Sprintf("0x%x", *r.L1BlobBaseFeeScalar)
	}
	if r.DepositNonce != nil {
		if hex, err := DecimalStringToHex(*r.DepositNonce); err == nil {
			out["depositNonce"] = hex
		}
	}
	if r.DepositReceiptVersion != nil {
		if hex, err := DecimalStringToHex(*r.DepositReceiptVersion); err == nil {
			out["depositReceiptVersion"] = hex
		}
	}
	if r.Timeboosted != nil {
		out["timeboosted"] = *r.Timeboosted
	}

	return out
}

// WithdrawalToJsonRpc converts a *Withdrawal into JSON-RPC representation.
func WithdrawalToJsonRpc(w *Withdrawal) map[string]interface{} {
	if w == nil {
		return nil
	}

	return map[string]interface{}{
		"index":          fmt.Sprintf("0x%x", w.Index),
		"validatorIndex": fmt.Sprintf("0x%x", w.ValidatorIndex),
		"address":        BytesToHex(w.Address),
		"amount":         fmt.Sprintf("0x%x", w.Amount),
	}
}

// WithdrawalsToJsonRpc converts a slice of *Withdrawal to []interface{} for JSON-RPC responses.
func WithdrawalsToJsonRpc(withdrawals []*Withdrawal) []interface{} {
	out := make([]interface{}, len(withdrawals))
	for i, w := range withdrawals {
		out[i] = WithdrawalToJsonRpc(w)
	}
	return out
}

// BlockToJsonRpc converts an EVM BlockHeader plus optional transactions to JSON-RPC format.
// If fullTxs is supplied it will be used; otherwise transaction hashes are used.
// If withdrawals are supplied, they will be included in the response.
func BlockToJsonRpc(header *BlockHeader, txHashes [][]byte, fullTxs []*Transaction, withdrawals []*Withdrawal) map[string]interface{} {
	if header == nil {
		return nil
	}

	res := map[string]interface{}{
		"number":           fmt.Sprintf("0x%x", header.Number),
		"hash":             BytesToHex(header.Hash),
		"parentHash":       BytesToHex(header.ParentHash),
		"sha3Uncles":       BytesToHex(header.Sha3Uncles),
		"logsBloom":        BytesToHex(header.LogsBloom),
		"transactionsRoot": BytesToHex(header.TransactionsRoot),
		"stateRoot":        BytesToHex(header.StateRoot),
		"receiptsRoot":     BytesToHex(header.ReceiptsRoot),
		"miner":            BytesToHex(header.Miner),
		"extraData":        BytesToHex(header.ExtraData),
		"size":             fmt.Sprintf("0x%x", header.Size),
		"gasLimit":         fmt.Sprintf("0x%x", header.GasLimit),
		"gasUsed":          fmt.Sprintf("0x%x", header.GasUsed),
		"timestamp":        fmt.Sprintf("0x%x", header.Timestamp),
	}

	if header.Nonce != nil {
		res["nonce"] = fmt.Sprintf("0x%016x", *header.Nonce)
	}
	if header.BaseFeePerGas != nil {
		if hex, err := DecimalStringToHex(*header.BaseFeePerGas); err == nil {
			res["baseFeePerGas"] = hex
		}
	}
	if header.Difficulty != nil {
		if hex, err := DecimalStringToHex(*header.Difficulty); err == nil {
			res["difficulty"] = hex
		}
	}
	if header.TotalDifficulty != nil {
		if hex, err := DecimalStringToHex(*header.TotalDifficulty); err == nil {
			res["totalDifficulty"] = hex
		}
	}
	if header.MixHash != nil {
		res["mixHash"] = BytesToHex(header.MixHash)
	}
	if header.WithdrawalsRoot != nil {
		res["withdrawalsRoot"] = BytesToHex(header.WithdrawalsRoot)
	}
	if header.RequestsHash != nil {
		res["requestsHash"] = BytesToHex(header.RequestsHash)
	}
	if header.BlobGasUsed != nil {
		res["blobGasUsed"] = fmt.Sprintf("0x%x", *header.BlobGasUsed)
	}
	if header.ExcessBlobGas != nil {
		res["excessBlobGas"] = fmt.Sprintf("0x%x", *header.ExcessBlobGas)
	}
	if header.ParentBeaconBlockRoot != nil {
		res["parentBeaconBlockRoot"] = BytesToHex(header.ParentBeaconBlockRoot)
	}
	if header.L1BlockNumber != nil {
		res["l1BlockNumber"] = fmt.Sprintf("0x%x", *header.L1BlockNumber)
	}
	if header.SendCount != nil {
		res["sendCount"] = fmt.Sprintf("0x%x", *header.SendCount)
	}
	if header.SendRoot != nil {
		res["sendRoot"] = BytesToHex(header.SendRoot)
	}
	if header.Epoch != nil {
		res["epoch"] = fmt.Sprintf("0x%x", *header.Epoch)
	}
	if header.Slot != nil {
		res["slot"] = fmt.Sprintf("0x%x", *header.Slot)
	}
	if header.ProposerIndex != nil {
		res["proposerIndex"] = fmt.Sprintf("0x%x", *header.ProposerIndex)
	}
	if header.TransactionCount != nil {
		res["transactionCount"] = fmt.Sprintf("0x%x", *header.TransactionCount)
	}
	if header.ProposerPublicKey != nil {
		res["proposerPublicKey"] = *header.ProposerPublicKey
	}
	// Output withdrawals array if provided
	if withdrawals != nil {
		res["withdrawals"] = WithdrawalsToJsonRpc(withdrawals)
	}
	if header.CanonicalRlp != nil {
		res["canonicalRlp"] = BytesToHex(header.CanonicalRlp)
	}

	// Uncles
	if len(header.Uncles) > 0 {
		uncles := make([]string, len(header.Uncles))
		for i, u := range header.Uncles {
			uncles[i] = BytesToHex(u)
		}
		res["uncles"] = uncles
	} else {
		res["uncles"] = []interface{}{}
	}

	// Transactions
	switch {
	case len(fullTxs) > 0:
		txs := make([]interface{}, len(fullTxs))
		for i, t := range fullTxs {
			txs[i] = TransactionToJsonRpc(t)
		}
		res["transactions"] = txs
	case len(txHashes) > 0:
		hashes := make([]string, len(txHashes))
		for i, h := range txHashes {
			hashes[i] = BytesToHex(h)
		}
		res["transactions"] = hashes
	default:
		res["transactions"] = []interface{}{}
	}

	return res
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

					authChainId, err := NumberishToUint64(getAuthString("chainId"))
					if err != nil {
						return nil, fmt.Errorf("failed to parse authorization chainId: %w", err)
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
						ChainId: authChainId,
						Address: authAddress,
						Nonce:   authNonce,
						R:       authR,
						S:       authS,
						YParity: authYParity,
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

	return tx, nil
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
