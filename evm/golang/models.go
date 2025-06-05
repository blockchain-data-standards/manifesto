package domain

import (
	"math/big"
	"time"
)

type BlockHeader struct {
	BaseFeePerGas    string         `json:"base_fee_per_gas"`
	Difficulty       float64        `json:"difficulty"`
	ExtraData        []byte         `json:"extra_data"`
	GasLimit         uint64         `json:"gas_limit"`
	GasUsed          uint64         `json:"gas_used"`
	Hash             []byte         `json:"hash"`
	LogsBloom        []byte         `json:"logs_bloom"`
	Miner            []byte         `json:"miner"`
	Nonce            []byte         `json:"nonce"`
	Number           uint64         `json:"number"`
	ParentHash       []byte         `json:"parent_hash"`
	ReceiptsRoot     []byte         `json:"receipts_root"`
	Sha3Uncles       []byte         `json:"sha3_uncles"`
	Size             int64          `json:"size"`
	StateRoot        []byte         `json:"state_root"`
	Timestamp        time.Time      `json:"timestamp"`
	TotalDifficulty  float64        `json:"total_difficulty"`
	TransactionCount uint64         `json:"transaction_count"`
	Transactions     []*Transaction `json:"transactions"`
	TransactionsRoot []byte         `json:"transactions_root"`
	WithdrawalsRoot  []byte         `json:"withdrawals_root"`
}

type Transaction struct {
	BlockNumber      uint64  `json:"block_number"`
	From             []byte  `json:"from"`
	Gas              big.Int `json:"gas"`       // UInt128
	GasPrice         big.Int `json:"gas_price"` // UInt128
	Hash             []byte  `json:"hash"`
	Input            string  `json:"input"`
	Nonce            uint64  `json:"nonce"`
	To               []byte  `json:"to"`
	TransactionIndex uint64  `json:"transaction_index"`
	Type             int64   `json:"type"`
	Value            big.Int `json:"value"` // UInt256
}

type Receipt struct {
	BlockHash         []byte  `json:"block_hash"`
	BlockNumber       uint64  `json:"block_number"`
	ContractAddress   []byte  `json:"contract_address"`
	CumulativeGasUsed big.Int `json:"cumulative_gas_used"` // UInt128
	From              []byte  `json:"from"`
	GasUsed           big.Int `json:"gas_used"` // UInt128
	Logs              []*Log  `json:"logs"`
	Status            int32   `json:"status"`
	To                []byte  `json:"to"`
	TransactionHash   []byte  `json:"transaction_hash"`
	TransactionIndex  uint64  `json:"transaction_index"`
}

type Log struct {
	Address          []byte   `json:"address"`
	BlockHash        []byte   `json:"block_hash"`
	BlockNumber      uint64   `json:"block_number"`
	Data             []byte   `json:"data"`
	LogIndex         uint64   `json:"log_index"`
	Topics           [][]byte `json:"topics"`
	TransactionHash  []byte   `json:"transaction_hash"`
	TransactionIndex uint64   `json:"transaction_index"`
}
