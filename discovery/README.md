# Blockchain Data Standards - Discovery

The Discovery standards allows blockchain data providers to expose their capabilities, enabling clients to discover which chains, transports, operations, models, and fields a provider supports.

## Overview

The Discovery service and schemas provide a standardized way for providers to declare:

- Which blockchain networks they support
- What data models (BlockHeader, Log, Transaction, etc.) are available for each network
- Which specific fields within those models they provide
- What query capabilities they offer (bulk operations, real-time updates, etc.)
- What transport protocols are available (REST, GraphQL, WebSocket, gRPC)

## Service

The Discovery service enables blockchain data providers to advertise their capabilities in a standardized format. Providers have two options for exposing their discovery information:

1. **Static JSON**: Providers can host a static JSON file conforming to the Discovery JSON Schema at the well-known endpoint: `DOMAIN/.well-known/bds.json`
2. **gRPC Service**: Providers can expose a dynamic gRPC service that implements the Discovery protocol for real-time capability queries

The well-known endpoint `/.well-known/bds.json` ensures clients can automatically discover provider capabilities without prior configuration.

## Schema

The Discovery schema defines the structure for describing provider capabilities:

- **Provider Information**: Basic metadata about the data provider (name, version, contact)
- **Supported Networks**: List of blockchain networks with their unique identifiers
- **Available Models**: Data models supported for each network (BlockHeader, Log, Transaction, Receipt, etc.)
- **Field Availability**: Specific fields within each model that the provider supports
- **Transport Protocols**: Available methods for accessing the data (REST, GraphQL, WebSocket, gRPC)
- **Query Capabilities**: Supported operations like bulk queries, real-time subscriptions, and historical data access

For detailed schema definitions and implementation:

- JSON Schema: See [schema.json](./schema.json)
- gRPC Protocol: See [discovery.proto](./discovery.proto)

## Example

Checkout [example.json](./example.json) to see how a provider can declare their supported chains and data models.

### Network Identification

Networks are uniquely identified using a UUID format: `ARCH:CHAIN_ID:GENESIS_SHORT_HASH`

- `ARCH`: Architecture type (evm, solana, bitcoin, cosmos)
- `CHAINID`: Chain-specific identifier
- `GENESIS_SHORT_HASH`: First 6 characters of genesis block hash in hex format (prevents chainId conflicts)

Examples:

- `evm:1:0xd4e567` - Ethereum Mainnet
- `evm:43114:0x31ced5` - Avalanche C-Chain
- `sol:101:0x452969` - Solana Mainnet
