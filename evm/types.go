package evm

// Common type aliases for better readability
type (
	// Address represents a 20-byte Ethereum address
	Address []byte
	
	// Hash represents a 32-byte hash (block hash, transaction hash, etc.)
	Hash []byte
	
	// Topic represents a 32-byte log topic
	Topic []byte
	
	// Topics represents an array of log topics
	Topics []Topic
)

// Common event signatures
const (
	// Transfer event signature: Transfer(address,address,uint256)
	TransferEventSignature = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	
	// Approval event signature: Approval(address,address,uint256)
	ApprovalEventSignature = "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"
)

// Well-known addresses
const (
	// Zero address
	ZeroAddress = "0x0000000000000000000000000000000000000000"
)

// Chain-specific constants
const (
	// AddressLength is the length of an Ethereum address in bytes
	AddressLength = 20
	
	// HashLength is the length of a hash in bytes
	HashLength = 32
	
	// TopicLength is the length of a log topic in bytes
	TopicLength = 32
	
	// BloomLength is the length of a bloom filter in bytes
	BloomLength = 256
	
	// MaxTopics is the maximum number of topics in a log (1 event signature + 3 indexed params)
	MaxTopics = 4
)

// NewAddress creates an Address from a byte slice, ensuring it's the correct length
func NewAddress(b []byte) Address {
	if len(b) == AddressLength {
		return Address(b)
	}
	// Pad or truncate to correct length
	addr := make([]byte, AddressLength)
	copy(addr[AddressLength-len(b):], b)
	return Address(addr)
}

// NewHash creates a Hash from a byte slice, ensuring it's the correct length
func NewHash(b []byte) Hash {
	if len(b) == HashLength {
		return Hash(b)
	}
	// Pad or truncate to correct length
	hash := make([]byte, HashLength)
	copy(hash[HashLength-len(b):], b)
	return Hash(hash)
}

// NewTopic creates a Topic from a byte slice, ensuring it's the correct length
func NewTopic(b []byte) Topic {
	if len(b) == TopicLength {
		return Topic(b)
	}
	// Pad or truncate to correct length
	topic := make([]byte, TopicLength)
	copy(topic[TopicLength-len(b):], b)
	return Topic(topic)
}
