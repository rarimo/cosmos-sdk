package keeper

import (
	"fmt"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Keeper defines the governance module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set gov specific params
	paramSpace types.ParamSubspace

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper

	// The reference to the DelegationSet and ValidatorSet to get information about validators and delegators
	sk types.StakingKeeper

	// GovHooks
	hooks types.GovHooks

	// The (unexposed) keys used to access the stores from the Context.
	storeKey storetypes.StoreKey

	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// Legacy Proposal router
	legacyRouter v1beta1.Router

	// Msg server router
	router *baseapp.MsgServiceRouter

	config types.Config
}

// NewKeeper returns a governance keeper. It handles:
// - submitting governance proposals
// - depositing funds into proposals, and activating upon sufficient funds being deposited
// - users voting on proposals, with weight proportional to stake in the system
// - and tallying the result of the vote.
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace types.ParamSubspace,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, sk types.StakingKeeper,
	legacyRouter v1beta1.Router, router *baseapp.MsgServiceRouter,
	config types.Config,
) Keeper {
	// ensure governance module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// It is vital to seal the governance proposal router here as to not allow
	// further handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	legacyRouter.Seal()

	// If MaxMetadataLen not set by app developer, set to default value.
	if config.MaxMetadataLen == 0 {
		config.MaxMetadataLen = types.DefaultConfig().MaxMetadataLen
	}

	return Keeper{
		storeKey:     key,
		paramSpace:   paramSpace,
		authKeeper:   authKeeper,
		bankKeeper:   bankKeeper,
		sk:           sk,
		cdc:          cdc,
		legacyRouter: legacyRouter,
		router:       router,
		config:       config,
	}
}

// SetHooks sets the hooks for governance
func (keeper *Keeper) SetHooks(gh types.GovHooks) *Keeper {
	if keeper.hooks != nil {
		panic("cannot set governance hooks twice")
	}

	keeper.hooks = gh

	return keeper
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Router returns the gov keeper's router
func (keeper Keeper) Router() *baseapp.MsgServiceRouter {
	return keeper.router
}

// LegacyRouter returns the gov keeper's legacy router
func (keeper Keeper) LegacyRouter() v1beta1.Router {
	return keeper.legacyRouter
}

// GetGovernanceAccount returns the governance ModuleAccount
func (keeper Keeper) GetGovernanceAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return keeper.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// ProposalQueues

// InsertActiveProposalQueue inserts a ProposalID into the active proposal queue at endBlock
func (keeper Keeper) InsertActiveProposalQueue(ctx sdk.Context, proposalID uint64, endBlock uint64) {
	store := ctx.KVStore(keeper.storeKey)
	bz := types.GetProposalIDBytes(proposalID)
	store.Set(types.ActiveProposalQueueKey(proposalID, endBlock), bz)
}

// RemoveFromActiveProposalQueue removes a proposalID from the Active Proposal Queue
func (keeper Keeper) RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, endBlock uint64) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(types.ActiveProposalQueueKey(proposalID, endBlock))
}

// InsertInactiveProposalQueue Inserts a ProposalID into the inactive proposal queue at endBlock
func (keeper Keeper) InsertInactiveProposalQueue(ctx sdk.Context, proposalID uint64, endBlock uint64) {
	store := ctx.KVStore(keeper.storeKey)
	bz := types.GetProposalIDBytes(proposalID)
	store.Set(types.InactiveProposalQueueKey(proposalID, endBlock), bz)
}

// RemoveFromInactiveProposalQueue removes a proposalID from the Inactive Proposal Queue
func (keeper Keeper) RemoveFromInactiveProposalQueue(ctx sdk.Context, proposalID uint64, endBlock uint64) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(types.InactiveProposalQueueKey(proposalID, endBlock))
}

// Iterators

// IterateActiveProposalsQueue iterates over the proposals in the active proposal queue
// and performs a callback function
func (keeper Keeper) IterateActiveProposalsQueue(ctx sdk.Context, endBlock uint64, cb func(proposal v1.Proposal) (stop bool)) {
	iterator := keeper.ActiveProposalQueueIterator(ctx, endBlock)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposalID, _ := types.SplitActiveProposalQueueKey(iterator.Key())
		proposal, found := keeper.GetProposal(ctx, proposalID)
		if !found {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}

		if cb(proposal) {
			break
		}
	}
}

// IterateInactiveProposalsQueue iterates over the proposals in the inactive proposal queue
// and performs a callback function
func (keeper Keeper) IterateInactiveProposalsQueue(ctx sdk.Context, endBlock uint64, cb func(proposal v1.Proposal) (stop bool)) {
	iterator := keeper.InactiveProposalQueueIterator(ctx, endBlock)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposalID, _ := types.SplitInactiveProposalQueueKey(iterator.Key())
		proposal, found := keeper.GetProposal(ctx, proposalID)
		if !found {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}

		if cb(proposal) {
			break
		}
	}
}

// ActiveProposalQueueIterator returns an sdk.Iterator for all the proposals in the Active Queue that expire by endBlock
func (keeper Keeper) ActiveProposalQueueIterator(ctx sdk.Context, endBlock uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.ActiveProposalByTimeKey(endBlock))
}

// InactiveProposalQueueIterator returns an sdk.Iterator for all the proposals in the Inactive Queue that expire by endBlock
func (keeper Keeper) InactiveProposalQueueIterator(ctx sdk.Context, endBlock uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.ActiveProposalByTimeKey(endBlock))
}

// assertMetadataLength returns an error if given metadata length
// is greater than a pre-defined maxMetadataLen.
func (k Keeper) assertMetadataLength(metadata string) error {
	if metadata != "" && uint64(len(metadata)) > k.config.MaxMetadataLen {
		return types.ErrMetadataTooLong.Wrapf("got metadata with length %d", len(metadata))
	}
	return nil
}
