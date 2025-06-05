# Common Artifacts

This directory contains common specifications and utilities for all blockchains.

## Catalog Specifications

[catalog.spec.json](./catalog.spec.json) contains a meta JSON-Schema definition file that is used to define downstream blockchain data models.

The main purpose of this schema is to unify how we describe different models in different blockchains, common types (such as Uint256), blockchain-specific information (such as supported chains, required node type, etc), in a machine-readable manner so that we can develop automation for library generation, data validation, etc.

## Schema Structure

A catalog file must include the following top-level fields:

* **version** - Semantic version of the catalog (e.g., "1.0.0")
* **blockchain** - Architecture type: "evm", "solana", "cosmos", or "bitcoin"
* **models** - Map of entity definitions (e.g., BlockHeader, Transaction, Log)
* **types** (optional) - Custom type definitions
* **providers** (optional) - Data provider definitions

## Defining Blockchain Models

Models represent blockchain data entities. Each model should be named in PascalCase and contain:

### Required Fields

- **description**: Human-readable description of the entity
- **fields**: Map of field names (camelCase) to field definitions

### Field Definition Structure

Each field must include:

- **type**: Data type (see supported types below)
- **description**: Clear description of what this field represents

Optional field properties:

- **required**: Whether the field is mandatory (default: false)
- **nullable**: Whether the field can be null (default: true)
- **chains**: Array of chain identifiers where this field applies (default: ["*"])
- **nodeType**: Array of node types that provide this field (default: ["full", "archive"])
- **engines**: Array of specific engines/providers that generate this field (default: ["*"])
- **examples**: Array of example values
- **validation**: Object containing validation rules (pattern, min, max, minLength, maxLength)
- **deprecated**: Boolean indicating if the field is deprecated
- **deprecationNote**: Explanation if deprecated is true
- **libraryHints**: Language-specific type overrides (go, typescript, rust, proto)

### Supported Data Types

- **Bytes**, **Bytes20**, **Bytes32**, **Bytes64**: Fixed or variable-length byte arrays
- **Uint256**, **Uint64**, **Uint32**, **Uint8**: Unsigned integers
- **Int256**, **Int64**, **Int32**: Signed integers
- **Boolean**: True/false values
- **String**: Text data
- **Timestamp**: Unix timestamps or ISO 8601 dates
- **BigDecimal**: High-precision decimal numbers
- **Array**: List of items (requires `arrayType`)
- **Object**: Nested structure (requires `objectSchema`)

### Example Model Definition

```json
{
  "BlockHeader": {
    "description": "Represents a block in the blockchain",
    "fields": {
      "number": {
        "type": "Uint64",
        "description": "BlockHeader number/height",
        "required": true,
        "nullable": false,
        "examples": [12345678, 0]
      },
      "hash": {
        "type": "Bytes32",
        "description": "BlockHeader hash",
        "required": true,
        "examples": ["0x1234...abcd"]
      },
      "timestamp": {
        "type": "Timestamp",
        "description": "BlockHeader timestamp",
        "required": true
      },
      "transactions": {
        "type": "Array",
        "arrayType": "String",
        "description": "List of transaction hashes",
        "chains": ["ethereum", "polygon"],
        "nodeType": ["full", "archive"]
      },
      "gasUsed": {
        "type": "Uint256",
        "description": "Total gas used in the block",
        "chains": ["*"],
        "validation": {
          "min": 0,
          "max": 30000000
        }
      }
    }
  }
}
```

## Defining Blockchain Data Providers

Providers represent data sources that implement the models. Each provider should have a lowercase kebab-case slug as its key.

### Required Provider Fields

- **title**: Human-readable name of the provider
- **description**: Detailed description of capabilities
- **support**: Object defining which chains/models/fields are supported

### Optional Provider Fields

- **siteUrl**: Official website URL
- **githubUrl**: GitHub repository URL (for open-source providers)
- **logoUrl**: URL to provider's logo
- **features**: Array of special capabilities
- **apiType**: API type ("rest", "graphql", "websocket", "grpc")
- **pricingModel**: Pricing structure ("free", "freemium", "paid", "open-source")

### Support Structure

The `support` field uses a flexible nested structure to express different levels of granularity:

#### Level 1: Full Support

```json
{
  "support": {
    "ethereum": "*"  // Supports all models and fields for Ethereum
  }
}
```

#### Level 2: Model-Specific Support

```json
{
  "support": {
    "arbitrum": {
      "models": ["BlockHeader", "Transaction", "Log"]  // Only these models
    }
  }
}
```

#### Level 3: Field-Specific Support

```json
{
  "support": {
    "optimism": {
      "models": {
        "BlockHeader": "*",  // All fields of BlockHeader
        "Transaction": ["hash", "from", "to", "value"]  // Only these fields
      }
    }
  }
}
```

### Complete Provider Example

```json
{
  "providers": {
    "infura": {
      "title": "Infura",
      "description": "Enterprise-grade Ethereum API provider",
      "siteUrl": "https://infura.io",
      "githubUrl": "https://github.com/INFURA",
      "logoUrl": "https://infura.io/logo.png",
      "support": {
        "ethereum": "*",
        "polygon": "*",
        "arbitrum": {
          "models": ["BlockHeader", "Transaction", "Log"]
        }
      },
      "features": ["websocket-support", "archive-data"],
      "apiType": "rest",
      "pricingModel": "freemium"
    }
  }
}
```

## Contributing Guide

### Adding a New Model

1. Choose an appropriate name in PascalCase (e.g., `TokenTransfer`)
2. Add it to the `models` section of the appropriate blockchain catalog
3. Include a clear `description`
4. Define all relevant `fields` with accurate types and descriptions
5. Specify chain-specific fields using the `chains` array
6. Add validation rules where appropriate
7. Include realistic examples for complex fields

### Adding a New Provider

1. Choose a unique slug in kebab-case (e.g., `alchemy-api`)
2. Add it to the `providers` section
3. Fill in all required fields (`title`, `description`, `support`)
4. Be specific about support levels - avoid over-claiming capabilities
5. Include relevant features that distinguish this provider
6. Add URLs for documentation and resources

### Best Practices

1. **Be Specific**: Use precise types and clear descriptions
2. **Consider Compatibility**: Think about how fields map across different chains
3. **Document Limitations**: Use `chains`, `nodeType`, and `engines` to indicate availability
4. **Provide Examples**: Include realistic example values, especially for complex types
5. **Validate Consistently**: Use validation rules to ensure data quality
6. **Version Carefully**: Update the catalog version when making breaking changes

### Testing Your Contributions

1. Validate your JSON against the schema
2. Ensure all required fields are present
3. Check that examples match the specified types
4. Verify that chain-specific fields make sense
5. Test that provider support claims are accurate

### Submitting Changes

1. Fork the repository
2. Create a feature branch for your changes
3. Update the relevant catalog files
4. Include examples demonstrating your additions
5. Submit a pull request with a clear description
6. Be prepared to iterate based on feedback 