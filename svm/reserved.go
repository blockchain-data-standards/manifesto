package svm

// BpfUpgradeableLoaderID gates program-id demotion: a key called as a program
// stays writable only when this loader is among the message's keys (closing
// over upgradeable programs that write to themselves during upgrades).
const BpfUpgradeableLoaderID = "BPFLoaderUpgradeab1e11111111111111111111111"

// reservedAccountKeys is ReservedAccountKeys::new_all_activated().active from
// agave-reserved-account-keys 2.3.13 — the set the getBlock encode path uses
// for BOTH legacy and v0 messages (solana-transaction-status lib.rs:554,783).
// new_all_activated includes feature-pending entries (secp256r1) on purpose;
// so does Agave's renderer regardless of cluster feature state.
var reservedAccountKeys = map[string]bool{
	"AddressLookupTab1e1111111111111111111111111": true,
	"BPFLoader2111111111111111111111111111111111": true,
	"BPFLoader1111111111111111111111111111111111": true,
	BpfUpgradeableLoaderID:                        true,
	"ComputeBudget111111111111111111111111111111": true,
	"Config1111111111111111111111111111111111111": true,
	"Ed25519SigVerify111111111111111111111111111": true,
	"Feature111111111111111111111111111111111111": true,
	"LoaderV411111111111111111111111111111111111": true,
	"KeccakSecp256k11111111111111111111111111111": true,
	"Secp256r1SigVerify1111111111111111111111111": true,
	"StakeConfig11111111111111111111111111111111": true,
	"Stake11111111111111111111111111111111111111": true,
	"11111111111111111111111111111111":            true,
	"Vote111111111111111111111111111111111111111": true,
	"ZkE1Gama1Proof11111111111111111111111111111": true,
	"ZkTokenProof1111111111111111111111111111111": true,
	"SysvarC1ock11111111111111111111111111111111": true,
	"SysvarEpochRewards1111111111111111111111111": true,
	"SysvarEpochSchedu1e111111111111111111111111": true,
	"SysvarFees111111111111111111111111111111111": true,
	"Sysvar1nstructions1111111111111111111111111": true,
	"SysvarLastRestartS1ot1111111111111111111111": true,
	"SysvarRecentB1ockHashes11111111111111111111": true,
	"SysvarRent111111111111111111111111111111111": true,
	"SysvarRewards111111111111111111111111111111": true,
	"SysvarS1otHashes111111111111111111111111111": true,
	"SysvarS1otHistory11111111111111111111111111": true,
	"SysvarStakeHistory1111111111111111111111111": true,
	"NativeLoader1111111111111111111111111111111": true,
	"Sysvar1111111111111111111111111111111111111": true,
}
