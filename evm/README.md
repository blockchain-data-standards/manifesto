# EVM Blockchain Data Standards

This directory contains standardized Protocol Buffer (protobuf) schemas and data models for EVM-compatible blockchains. The protobuf definitions are compiled to provide type-safe data structures and gRPC services for querying blockchain data.

## Generated Code

The Go code is generated from the protobuf definitions:
- `models.proto` → `models.pb.go` - Core data structures (BlockHeader, Log)
- `query.proto` → `query.pb.go`, `query_grpc.pb.go` - gRPC query service definitions

## Models

### BlockHeader

A confirmed block on an EVM-compatible blockchain containing transactions and state changes

| Field | Type | Required | Description | Chains | Deprecated |
|-------|------|----------|-------------|---------|------------|
| `extraData` | `Bytes` | ✓ | Arbitrary data included by the block producer. Limited to 32 bytes in Ethereum mainnet. Often con... | * |  |
| `gasLimit` | `Uint64` | ✓ | Maximum total gas that can be consumed by all transactions in this block. Set by miners/validator... | * |  |
| `gasUsed` | `Uint64` | ✓ | Sum of gas actually consumed by all transactions in this block. Always less than or equal to gasL... | * |  |
| `hash` | `Bytes32` | ✓ | The Keccak-256 hash of the block header (parentHash, unclesHash, miner, stateRoot, transactionsRo... | * |  |
| `logsBloom` | `Bytes` | ✓ | 2048-bit (256 bytes) bloom filter of all log topics and addresses from all transactions in the bl... | * |  |
| `miner` | `Bytes20` | ✓ | The 20-byte Ethereum address that received the block reward and transaction fees. In PoW chains, ... | * |  |
| `number` | `Uint64` | ✓ | The block height - sequential position of this block in the blockchain starting from genesis (0).... | * |  |
| `parentHash` | `Bytes32` | ✓ | The Keccak-256 hash of the parent blocks header. This creates the blockchains linked structure wh... | * |  |
| `receiptsRoot` | `Bytes32` | ✓ | Root hash of the receipts trie containing transaction receipts (status, logs, gas used) for all t... | * |  |
| `sha3Uncles` | `Bytes32` | ✓ | Keccak-256 hash of the uncles (ommer blocks) list. Uncle blocks are valid blocks mined at the sam... | * |  |
| `size` | `Uint64` | ✓ | Total size of the block in bytes including header and all transactions. Used to enforce block siz... | * |  |
| `stateRoot` | `Bytes32` | ✓ | Root hash of the global state trie (Patricia Merkle Tree) after executing all transactions in thi... | * |  |
| `timestamp` | `Uint64` | ✓ | Unix timestamp (seconds since epoch) when this block was mined/produced. Set by the block produce... | * |  |
| `transactionsRoot` | `Bytes32` | ✓ | Root hash of the transactions trie containing all transactions included in this block. Enables Me... | * |  |
| `baseFeePerGas` | `Uint256?` | ✗ | Minimum fee per gas unit (in wei) required for transaction inclusion in this block. Introduced in... | ethereum, polygon, optimism, arbitrum, base |  |
| `blobGasUsed` | `Uint64?` | ✗ | Total blob gas consumed by blob-carrying transactions (EIP-4844) in this block. Each blob consume... | ethereum, optimism, base |  |
| `canonicalRlp` | `Bytes?` | ✗ | RLP (Recursive Length Prefix) encoded canonical block header. This is the exact bytes that when h... | * |  |
| `difficulty` | `Uint256?` | ✗ | PoW mining difficulty - how hard it was to find a valid nonce for this block. Adjusted every bloc... | * | ⚠️ |
| `epoch` | `Uint64?` | ✗ | Beacon chain epoch number (32 slots, ~6.4 minutes). Groups slots for validator duties, finality c... | ethereum |  |
| `excessBlobGas` | `Uint64?` | ✗ | Running total of blob gas consumed above the target (393216 gas, 3 blobs) from previous block. Us... | ethereum, optimism, base |  |
| `l1BlockNumber` | `Uint64?` | ✗ | The L1 (Ethereum mainnet) block number that this L2 block is anchored to. Used by L2 sequencers t... | optimism, arbitrum, base, zksync, polygon-zkevm |  |
| `mixHash` | `Bytes32?` | ✗ | PoW mining mix hash used with nonce to prove work was done. Part of Ethash algorithm preventing A... | * | ⚠️ |
| `nonce` | `Uint64?` | ✗ | PoW mining nonce - 64-bit value miners increment to find a valid block hash below the difficulty ... | * | ⚠️ |
| `parentBeaconBlockRoot` | `Bytes32?` | ✗ | Hash of the parent beacon chain block root. Introduced in EIP-4788 (Dencun) to enable trustless a... | ethereum |  |
| `proposerIndex` | `Uint64?` | ✗ | Beacon chain validator index that proposed this block. Identifies which of the hundreds of thousa... | ethereum |  |
| `proposerPublicKey` | `String?` | ✗ | BLS12-381 public key of the beacon chain validator that proposed this block. 48-byte hex-encoded ... | ethereum |  |
| `sendCount` | `Uint64?` | ✗ | Arbitrum-specific field tracking the number of L2-to-L1 messages sent in this block. Used for cro... | * |  |
| `sendRoot` | `Bytes32?` | ✗ | Arbitrum-specific merkle root of all L2-to-L1 messages sent up to and including this block. Enabl... | * |  |
| `slot` | `Uint64?` | ✗ | Beacon chain slot number (12 seconds). The primary time unit in PoS Ethereum. Each slot has one a... | ethereum |  |
| `totalDifficulty` | `Uint256?` | ✗ | Cumulative sum of all block difficulties from genesis to this block. Used to determine the canoni... | * | ⚠️ |
| `transactionCount` | `Uint32?` | ✗ | Number of transactions included in this block. Derived from transactions array length but stored ... | * |  |
| `uncles` | `Bytes32[]?` | ✗ | Array of uncle (ommer) block hashes included in this block. Uncle blocks are valid blocks mined a... | * |  |
| `withdrawals` | `Bytes?` | ✗ | Encoded withdrawal operations from beacon chain validators. Introduced in Shanghai upgrade. Conta... | ethereum |  |
| `withdrawalsRoot` | `Bytes32?` | ✗ | Root hash of the withdrawals trie containing validator withdrawals from the beacon chain. Introdu... | ethereum |  |

### Log

An event emitted by a smart contract during transaction execution on an EVM-compatible blockchain. Logs are the primary mechanism for smart contracts to communicate with external applications, enabling event-driven architectures and efficient querying of on-chain activity

| Field | Type | Required | Description | Chains | Deprecated |
|-------|------|----------|-------------|---------|------------|
| `address` | `Bytes20` | ✓ | The 20-byte address of the smart contract that emitted this log. This is the contract whose code... | * |  |
| `blockHash` | `Bytes32` | ✓ | The hash of the block containing this log. Provides a direct link to the block and enables verif... | * |  |
| `blockNumber` | `Uint64` | ✓ | The block number where this log was emitted. Used for querying logs within block ranges, calcula... | * |  |
| `data` | `Bytes` | ✓ | The non-indexed data of the log containing event parameters that are not indexed. While indexed ... | * |  |
| `logIndex` | `Uint32` | ✓ | The zero-based index position of this log within the block. Unique within a block and assigned s... | * |  |
| `topics` | `Bytes32[]` | ✓ | Array of indexed event parameters (max 4 topics). Topic[0] is the keccak256 hash of the event si... | * |  |
| `transactionHash` | `Bytes32` | ✓ | The hash of the transaction that emitted this log. Links the log to its originating transaction,... | * |  |
| `transactionIndex` | `Uint32` | ✓ | The zero-based index position of the transaction within its block. Combined with blockNumber/blo... | * |  |
| `blockTimestamp` | `Uint64?` | ✗ | Unix timestamp of the block containing this log. Denormalized from block data for query convenie... | * |  |

## Usage

### Go

```go
package main

import (
    "fmt"
    "encoding/hex"
    "github.com/blockchain-data-standards/manifesto/evm"
)

func main() {
    // Create a new block instance with protobuf-generated types
    block := &evm.BlockHeader{
        Number:     18000000,
        Hash:       evm.MustHexToBytes("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
        ParentHash: evm.MustHexToBytes("0x567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234"),
        Timestamp:  1699564800,
        Miner:      evm.MustHexToBytes("0x742d35Cc6634C0532925a3b844Bc9e7595f7BBDc"),
        GasLimit:   30000000,
        GasUsed:    15000000,
        BaseFeePerGas: evm.StringPtr("1000000000"), // 1 gwei in wei as string
    }
    
    // Check for EIP-1559 support
    if block.BaseFeePerGas != nil {
        fmt.Printf("Base fee: %s wei\n", *block.BaseFeePerGas)
    }
    
    // Work with L2-specific fields
    if block.L1BlockNumber != nil {
        fmt.Printf("L1 Block Number: %d\n", *block.L1BlockNumber)
    }
    
    // Create a new log instance
    log := &evm.Log{
        Address: evm.MustHexToAddress(evm.USDCAddress),
        Topics: [][]byte{
            evm.MustHexToTopic(evm.TransferEventSignature), // Transfer event
            evm.MustHexToBytes("0x0000000000000000000000001234567890123456789012345678901234567890"), // from
            evm.MustHexToBytes("0x0000000000000000000000009876543210987654321098765432109876543210"), // to
        },
        Data:             []byte{}, // amount would be encoded here for Transfer
        BlockNumber:      18000000,
        BlockHash:        block.Hash,
        TransactionHash:  evm.MustHexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
        TransactionIndex: 42,
        LogIndex:         123,
    }
    
    // Access log fields
    fmt.Printf("Log emitted by contract: %s\n", evm.BytesToHex(log.Address))
    fmt.Printf("Event signature: %s\n", evm.BytesToHex(log.Topics[0]))
}

### gRPC Query Service

The protobuf definitions also include a gRPC service for querying blockchain data:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "encoding/hex"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "github.com/blockchain-data-standards/manifesto/evm"
)

func main() {
    // Connect to gRPC server
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()
    
    // Create client
    client := evm.NewEVMQueryServiceClient(conn)
    ctx := context.Background()
    
    // Query blocks by range
    blocksResp, err := client.GetBlocksByRange(ctx, &evm.GetBlocksByRangeRequest{
        FromBlock:           18000000,
        ToBlock:             18000010,
        IncludeTransactions: true,
        Limit:               evm.Uint32Ptr(5),
    })
    if err != nil {
        log.Fatalf("Failed to get blocks: %v", err)
    }
    
    for _, blockWithTxs := range blocksResp.Blocks {
        fmt.Printf("Block #%d: %s\n", 
            blockWithTxs.Block.Number, 
            evm.BytesToHex(blockWithTxs.Block.Hash))
    }
    
    // Query logs with filters using the helper function
    logsReq, err := evm.BuildTransferLogQuery(18000000, 18000100, []string{
        evm.USDCAddress, // USDC
        evm.USDTAddress, // USDT
    })
    if err != nil {
        log.Fatalf("Failed to build query: %v", err)
    }
    
    logsResp, err := client.GetLogs(ctx, logsReq)
    if err != nil {
        log.Fatalf("Failed to get logs: %v", err)
    }
    
    fmt.Printf("Found %d Transfer events\n", len(logsResp.Logs))
    
    // Or build a custom query
    customLogsResp, err := client.GetLogs(ctx, &evm.GetLogsRequest{
        FromBlock: evm.Uint64Ptr(18000000),
        ToBlock:   evm.Uint64Ptr(18000100),
        Addresses: [][]byte{
            evm.MustHexToAddress(evm.WETHAddress),
        },
        Topics: []*evm.TopicFilter{
            evm.MustNewTopicFilter(
                evm.DepositEventSignature,
                evm.WithdrawalEventSignature,
            ),
        },
    })
    if err != nil {
        log.Fatalf("Failed to get logs: %v", err)
    }
    
    fmt.Printf("Found %d WETH Deposit/Withdrawal events\n", len(customLogsResp.Logs))
}
```

## See Also

- [Ethereum Yellow Paper](https://ethereum.github.io/yellowpaper/paper.pdf) - Formal specification
- [EIP-1559](https://eips.ethereum.org/EIPS/eip-1559) - Fee market changes
- [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844) - Proto-danksharding
- [The Merge](https://ethereum.org/en/upgrades/merge/) - PoW to PoS transition
