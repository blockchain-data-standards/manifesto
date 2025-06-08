package evm

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// Hex conversion utilities

// HexToBytes converts a hex string to bytes, removing the 0x prefix if present
func HexToBytes(s string) ([]byte, error) {
	s = RemoveHexPrefix(s)
	return hex.DecodeString(s)
}

// MustHexToBytes converts a hex string to bytes, panicking on error
// Useful for hardcoded values in tests or known-good constants
func MustHexToBytes(s string) []byte {
	b, err := HexToBytes(s)
	if err != nil {
		panic(fmt.Sprintf("invalid hex string: %s", s))
	}
	return b
}

// HexToAddress converts a hex string to an Address
func HexToAddress(s string) (Address, error) {
	b, err := HexToBytes(s)
	if err != nil {
		return nil, err
	}
	return NewAddress(b), nil
}

// MustHexToAddress converts a hex string to an Address, panicking on error
func MustHexToAddress(s string) Address {
	addr, err := HexToAddress(s)
	if err != nil {
		panic(fmt.Sprintf("invalid address hex: %s", s))
	}
	return addr
}


func HexToUint32(hex string) (uint32, error) {
	if len(hex) < 2 || hex[:2] != "0x" {
		return 0, fmt.Errorf("invalid hex string: %s", hex)
	}
	var result uint32
	_, err := fmt.Sscanf(hex, "0x%x", &result)
	return result, err
}

func HexToUint64(hex string) (uint64, error) {
	if len(hex) < 2 || hex[:2] != "0x" {
		return 0, fmt.Errorf("invalid hex string: %s", hex)
	}
	var result uint64
	_, err := fmt.Sscanf(hex, "0x%x", &result)
	return result, err
}

func MustHexToUint64(hex string) uint64 {
	result, err := HexToUint64(hex)
	if err != nil {
		panic(fmt.Sprintf("invalid hex string: %s", hex))
	}
	return result
}

func MustHexToUint32(hex string) uint32 {
	result, err := HexToUint32(hex)
	if err != nil {
		panic(fmt.Sprintf("invalid hex string: %s", hex))
	}
	return result
}

func NumberishToUint64(s string) (uint64, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return HexToUint64(s)
	}
	return strconv.ParseUint(s, 10, 64)
}

func NumberishToUint32(s string) (uint32, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return HexToUint32(s)
	}
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(u), nil
}

// HexToHash converts a hex string to a Hash
func HexToHash(s string) (Hash, error) {
	b, err := HexToBytes(s)
	if err != nil {
		return nil, err
	}
	return NewHash(b), nil
}

// MustHexToHash converts a hex string to a Hash, panicking on error
func MustHexToHash(s string) Hash {
	h, err := HexToHash(s)
	if err != nil {
		panic(fmt.Sprintf("invalid hash hex: %s", s))
	}
	return h
}

// HexToTopic converts a hex string to a Topic
func HexToTopic(s string) (Topic, error) {
	b, err := HexToBytes(s)
	if err != nil {
		return nil, err
	}
	return NewTopic(b), nil
}

// MustHexToTopic converts a hex string to a Topic, panicking on error
func MustHexToTopic(s string) Topic {
	t, err := HexToTopic(s)
	if err != nil {
		panic(fmt.Sprintf("invalid topic hex: %s", s))
	}
	return t
}

// BytesToHex converts bytes to a hex string with 0x prefix
func BytesToHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

// RemoveHexPrefix removes the 0x prefix from a hex string if present
func RemoveHexPrefix(s string) string {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return s[2:]
	}
	return s
}

// AddHexPrefix adds the 0x prefix to a hex string if not present
func AddHexPrefix(s string) string {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return s
	}
	return "0x" + s
}

// Pointer helper functions

// Uint32Ptr returns a pointer to the given uint32 value
func Uint32Ptr(v uint32) *uint32 {
	return &v
}

// Uint64Ptr returns a pointer to the given uint64 value
func Uint64Ptr(v uint64) *uint64 {
	return &v
}

// StringPtr returns a pointer to the given string value
func StringPtr(s string) *string {
	return &s
}

// BoolPtr returns a pointer to the given bool value
func BoolPtr(b bool) *bool {
	return &b
}

// BytesPtr returns a pointer to the given byte slice
func BytesPtr(b []byte) *[]byte {
	return &b
}

// Topic filter helpers

// NewTopicFilter creates a TopicFilter with the given values
func NewTopicFilter(values ...string) (*TopicFilter, error) {
	filter := &TopicFilter{
		Values: make([][]byte, 0, len(values)),
	}
	
	for _, v := range values {
		b, err := HexToBytes(v)
		if err != nil {
			return nil, fmt.Errorf("invalid topic value %s: %w", v, err)
		}
		filter.Values = append(filter.Values, b)
	}
	
	return filter, nil
}

// MustNewTopicFilter creates a TopicFilter with the given values, panicking on error
func MustNewTopicFilter(values ...string) *TopicFilter {
	filter, err := NewTopicFilter(values...)
	if err != nil {
		panic(err)
	}
	return filter
}

// Common query builders

// BuildTransferLogQuery builds a GetLogsRequest for ERC20/ERC721 Transfer events
func BuildTransferLogQuery(fromBlock, toBlock uint64, contractAddresses []string) (*GetLogsRequest, error) {
	req := &GetLogsRequest{
		FromBlock: &fromBlock,
		ToBlock:   &toBlock,
		Addresses: make([][]byte, 0, len(contractAddresses)),
	}
	
	// Add contract addresses
	for _, addr := range contractAddresses {
		b, err := HexToBytes(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid contract address %s: %w", addr, err)
		}
		req.Addresses = append(req.Addresses, b)
	}
	
	// Add Transfer event signature as first topic
	transferTopic, err := NewTopicFilter(TransferEventSignature)
	if err != nil {
		return nil, err
	}
	req.Topics = []*TopicFilter{transferTopic}
	
	return req, nil
}

// IsZeroAddress checks if an address is the zero address
func IsZeroAddress(addr Address) bool {
	for _, b := range addr {
		if b != 0 {
			return false
		}
	}
	return len(addr) == AddressLength
}

// IsZeroHash checks if a hash is the zero hash
func IsZeroHash(h Hash) bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return len(h) == HashLength
}
