package types

import (
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeMintTokens = "mint_tokens"
)

func init() {
	govv1beta1.RegisterProposalType(ProposalTypeMintTokens)
	//gov.RegisterProposalTypeCodec(&MintTokensProposal{}, "rarimocore/MintTokensProposal") // we register it in codec.go/RegisterLegacyAminoCodec
}

// Implements Proposal Interface
var _ govv1beta1.Content = &MintTokensProposal{}

func (m *MintTokensProposal) ProposalRoute() string { return RouterKey }
func (m *MintTokensProposal) ProposalType() string  { return ProposalTypeMintTokens }

func (m *MintTokensProposal) ValidateBasic() error {
	return govv1beta1.ValidateAbstract(m)
}
