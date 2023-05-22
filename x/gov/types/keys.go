package types

import (
	"encoding/binary"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gov"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	lenUint64 = 8
)

// Keys for governance store
// Items are stored with the following key: values
//
// - 0x00<proposalID_Bytes>: Proposal
//
// - 0x01<endTime_Bytes><proposalID_Bytes>: activeProposalID
//
// - 0x02<endTime_Bytes><proposalID_Bytes>: inactiveProposalID
//
// - 0x03: nextProposalID
//
// - 0x10<proposalID_Bytes><depositorAddrLen (1 Byte)><depositorAddr_Bytes>: Deposit
//
// - 0x20<proposalID_Bytes><voterAddrLen (1 Byte)><voterAddr_Bytes>: Voter
var (
	ProposalsKeyPrefix          = []byte{0x00}
	ActiveProposalQueuePrefix   = []byte{0x01}
	InactiveProposalQueuePrefix = []byte{0x02}
	ProposalIDKey               = []byte{0x03}

	DepositsKeyPrefix = []byte{0x10}

	VotesKeyPrefix = []byte{0x20}
)

// GetProposalIDBytes returns the byte representation of the proposalID
func GetProposalIDBytes(proposalID uint64) (proposalIDBz []byte) {
	proposalIDBz = make([]byte, 8)
	binary.BigEndian.PutUint64(proposalIDBz, proposalID)
	return
}

// GetProposalIDFromBytes returns proposalID in uint64 format from a byte array
func GetProposalIDFromBytes(bz []byte) (proposalID uint64) {
	return binary.BigEndian.Uint64(bz)
}

// ProposalKey gets a specific proposal from the store
func ProposalKey(proposalID uint64) []byte {
	return append(ProposalsKeyPrefix, GetProposalIDBytes(proposalID)...)
}

// ActiveProposalByTimeKey gets the active proposal queue key by endTime
func ActiveProposalByTimeKey(endBlock uint64) []byte {
	return append(ActiveProposalQueuePrefix, getBlockKey(endBlock)...)
}

// ActiveProposalQueueKey returns the key for a proposalID in the activeProposalQueue
func ActiveProposalQueueKey(proposalID uint64, endBlock uint64) []byte {
	return append(ActiveProposalByTimeKey(endBlock), GetProposalIDBytes(proposalID)...)
}

// InactiveProposalByTimeKey gets the inactive proposal queue key by endTime
func InactiveProposalByTimeKey(endBlock uint64) []byte {
	return append(InactiveProposalQueuePrefix, getBlockKey(endBlock)...)
}

// InactiveProposalQueueKey returns the key for a proposalID in the inactiveProposalQueue
func InactiveProposalQueueKey(proposalID uint64, endBlock uint64) []byte {
	return append(InactiveProposalByTimeKey(endBlock), GetProposalIDBytes(proposalID)...)
}

// DepositsKey gets the first part of the deposits key based on the proposalID
func DepositsKey(proposalID uint64) []byte {
	return append(DepositsKeyPrefix, GetProposalIDBytes(proposalID)...)
}

// DepositKey key of a specific deposit from the store
func DepositKey(proposalID uint64, depositorAddr sdk.AccAddress) []byte {
	return append(DepositsKey(proposalID), address.MustLengthPrefix(depositorAddr.Bytes())...)
}

// VotesKey gets the first part of the votes key based on the proposalID
func VotesKey(proposalID uint64) []byte {
	return append(VotesKeyPrefix, GetProposalIDBytes(proposalID)...)
}

// VoteKey key of a specific vote from the store
func VoteKey(proposalID uint64, voterAddr sdk.AccAddress) []byte {
	return append(VotesKey(proposalID), address.MustLengthPrefix(voterAddr.Bytes())...)
}

// Split keys function; used for iterators

// SplitProposalKey split the proposal key and returns the proposal id
func SplitProposalKey(key []byte) (proposalID uint64) {
	kv.AssertKeyLength(key[1:], 8)

	return GetProposalIDFromBytes(key[1:])
}

// SplitActiveProposalQueueKey split the active proposal key and returns the proposal id and endTime
func SplitActiveProposalQueueKey(key []byte) (proposalID uint64, endBlock uint64) {
	return splitKeyWithTime(key)
}

// SplitInactiveProposalQueueKey split the inactive proposal key and returns the proposal id and endTime
func SplitInactiveProposalQueueKey(key []byte) (proposalID uint64, endBlock uint64) {
	return splitKeyWithTime(key)
}

// SplitKeyDeposit split the deposits key and returns the proposal id and depositor address
func SplitKeyDeposit(key []byte) (proposalID uint64, depositorAddr sdk.AccAddress) {
	return splitKeyWithAddress(key)
}

// SplitKeyVote split the votes key and returns the proposal id and voter address
func SplitKeyVote(key []byte) (proposalID uint64, voterAddr sdk.AccAddress) {
	return splitKeyWithAddress(key)
}

// private functions

func splitKeyWithTime(key []byte) (proposalID uint64, endBlock uint64) {
	kv.AssertKeyLength(key[1:], 8+lenUint64)

	endBlock = parseBlockKey(key[1 : 1+lenUint64])
	proposalID = GetProposalIDFromBytes(key[1+lenUint64:])
	return
}

func splitKeyWithAddress(key []byte) (proposalID uint64, addr sdk.AccAddress) {
	// Both Vote and Deposit store keys are of format:
	// <prefix (1 Byte)><proposalID (8 bytes)><addrLen (1 Byte)><addr_Bytes>
	kv.AssertKeyAtLeastLength(key, 10)
	proposalID = GetProposalIDFromBytes(key[1:9])
	kv.AssertKeyAtLeastLength(key, 11)
	addr = sdk.AccAddress(key[10:])
	return
}

func getBlockKey(i uint64) []byte {
	b := make([]byte, lenUint64)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

func parseBlockKey(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
