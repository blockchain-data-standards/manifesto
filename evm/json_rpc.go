package evm

type JsonRpcBlock struct {
	BaseFeePerGas         string   `json:"baseFeePerGas"`
	BlobGasUsed           string   `json:"blobGasUsed"`
	Difficulty            string   `json:"difficulty"`
	ExcessBlobGas         string   `json:"excessBlobGas"`
	ExtraData             string   `json:"extraData"`
	GasLimit              string   `json:"gasLimit"`
	GasUsed               string   `json:"gasUsed"`
	Hash                  string   `json:"hash"`
	LogsBloom             string   `json:"logsBloom"`
	Miner                 string   `json:"miner"`
	MixHash               string   `json:"mixHash"`
	Nonce                 string   `json:"nonce"`
	Number                string   `json:"number"`
	ParentBeaconBlockRoot string   `json:"parentBeaconBlockRoot"`
	ParentHash            string   `json:"parentHash"`
	ReceiptsRoot          string   `json:"receiptsRoot"`
	Sha3Uncles            string   `json:"sha3Uncles"`
	Size                  string   `json:"size"`
	StateRoot             string   `json:"stateRoot"`
	Timestamp             string   `json:"timestamp"`
	TotalDifficulty       string   `json:"totalDifficulty"`
	TransactionsRoot      string   `json:"transactionsRoot"`
	Uncles                []string `json:"uncles"`
	WithdrawalsRoot       string   `json:"withdrawalsRoot"`
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
	size, err := NumberishToUint64(b.Size)
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

	// Convert uncles array
	uncles := make([][]byte, 0, len(b.Uncles))
	for _, uncle := range b.Uncles {
		uncleBytes, err := HexToBytes(uncle)
		if err != nil {
			return nil, err
		}
		uncles = append(uncles, uncleBytes)
	}

	// Build the BlockHeader
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
		BaseFeePerGas:         &b.BaseFeePerGas,
		Difficulty:            &b.Difficulty,
		TotalDifficulty:       &b.TotalDifficulty,
		Uncles:                uncles,
	}

	return header, nil
}

type JsonRpcReceipt struct {
	BlockHash         string        `json:"blockHash"`
	BlockNumber       string        `json:"blockNumber"`
	ContractAddress   string        `json:"contractAddress"`
	CumulativeGasUsed string        `json:"cumulativeGasUsed"`
	EffectiveGasPrice string        `json:"effectiveGasPrice"`
	From              string        `json:"from"`
	GasUsed           string        `json:"gasUsed"`
	Logs              []*JsonRpcLog `json:"logs"`
	LogsBloom         string        `json:"logsBloom"`
	Root              string        `json:"root"`
	Status            string        `json:"status"`
	To                string        `json:"to"`
	TransactionHash   string        `json:"transactionHash"`
	TransactionIndex  string        `json:"transactionIndex"`
	Type              string        `json:"type"`
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
	
	return &Receipt{
		TransactionHash:   transactionHash,
		BlockNumber:       blockNumber,
		BlockHash:         blockHash,
		TransactionIndex:  uint32(transactionIndex),
		Type:              typ,
		Status:            status,
		GasUsed:           gasUsed,
		CumulativeGasUsed: cumulativeGasUsed,
		EffectiveGasPrice: r.EffectiveGasPrice,
		LogsBloom:         logsBloom,
		Logs:              logs,
		ContractAddress:   contractAddress,
		Root:              root,
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
