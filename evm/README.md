# EVM Blockchain Data Standards

This directory contains standardized Protocol Buffer (protobuf) schemas and data models for EVM-compatible blockchains. The protobuf definitions are compiled to provide type-safe data structures and gRPC services for querying blockchain data.

## Generated Code

The Go code is generated from the protobuf definitions:

- `models.proto` → `models.pb.go` - Core data structures (BlockHeader, Log)
- `query.proto` → `query.pb.go`, `query_grpc.pb.go` - gRPC query service definitions

## Models

### BlockHeader

A confirmed block on an EVM-compatible blockchain containing transactions and state changes.

- [json_rpc.go -> JsonRpcBlock](./json_rpc.go#L44)
- [json_rpc.go -> JsonRpcBlock.ToProto()](./json_rpc.go#L84)
- [json_rpc.go -> BlockToJsonRpc()](./json_rpc.go#L1163)

### Transaction

Represents a transaction on an EVM-compatible blockchain.

- [json_rpc.go -> ParseJsonRpcTransaction()](./json_rpc.go#L1287)
- [json_rpc.go -> TransactionToJsonRpc()](./json_rpc.go#L741)

### Log

An event emitted by a smart contract during transaction execution on an EVM-compatible blockchain. Logs are the primary mechanism for smart contracts to communicate with external applications, enabling event-driven architectures and efficient querying of on-chain activity

- [json_rpc.go -> JsonRpcLog](./json_rpc.go#L632)
- [json_rpc.go -> JsonRpcLog.ToProto()](./json_rpc.go#L644)
- [json_rpc.go -> LogToJsonRpc()](./json_rpc.go#L702)

### Receipt

Represents the result of executing a transaction on an EVM blockchain.

- [json_rpc.go -> JsonRpcReceipt](./json_rpc.go#L375)
- [json_rpc.go -> JsonRpcReceipt.ToProto()](./json_rpc.go#L409)
- [json_rpc.go -> ReceiptToJsonRpc()](./json_rpc.go#L1012)

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
