package evm

type JsonRpcBlock struct {
	BaseFeePerGas    string `json:"baseFeePerGas"`
	ExtraData        string `json:"extraData"`
	GasLimit         string `json:"gasLimit"`
	GasUsed          string `json:"gasUsed"`
	Hash             string `json:"hash"`
	LogsBloom        string `json:"logsBloom"`
	Number           string `json:"number"`
	ParentHash       string `json:"parentHash"`
	ReceiptsRoot     string `json:"receiptsRoot"`
	StateRoot        string `json:"stateRoot"`
	Timestamp        string `json:"timestamp"`
	TransactionsRoot string `json:"transactionsRoot"`
}

func (b *JsonRpcBlock) ToProto() (*BlockHeader, error) {
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
	extraData, err := HexToBytes(b.ExtraData)
	if err != nil {
		return nil, err
	}
	return &BlockHeader{
		BaseFeePerGas:    &b.BaseFeePerGas,
		ExtraData:        extraData,
		GasLimit:         gasLimit,
		GasUsed:          gasUsed,
		Hash:             hash,
		LogsBloom:        logsBloom,
		Number:           number,
		ParentHash:       parentHash,
		ReceiptsRoot:     receiptsRoot,
		StateRoot:        stateRoot,
		Timestamp:        timestamp,
		TransactionsRoot: transactionsRoot,
	}, nil
}

type JsonRpcReceipt struct {
	BlockHash        string        `json:"blockHash"`
	BlockNumber      string        `json:"blockNumber"`
	From             string        `json:"from"`
	GasUsed          string        `json:"gasUsed"`
	Logs             []*JsonRpcLog `json:"logs"`
	LogsBloom        string        `json:"logsBloom"`
	To               string        `json:"to"`
	TransactionHash  string        `json:"transactionHash"`
	TransactionIndex string        `json:"transactionIndex"`
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
	logs := make([]*Log, 0, len(r.Logs))
	for _, log := range r.Logs {
		protoLog, err := log.ToProto()
		if err != nil {
			return nil, err
		}
		logs = append(logs, protoLog)
	}
	return &Receipt{
		BlockNumber:      blockNumber,
		TransactionIndex: uint32(transactionIndex),
		GasUsed:          gasUsed,
		LogsBloom:        logsBloom,
		BlockHash:        blockHash,
		TransactionHash:  transactionHash,
		Logs:             logs,
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
