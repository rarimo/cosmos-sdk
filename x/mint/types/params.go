package types

import (
	"errors"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyMintDenom      = []byte("MintDenom")
	KeyBlocksPerMonth = []byte("BlocksPerMonth")
	KeyMonthReward    = []byte("MonthReward")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, blocksPerMonth uint64, monthReward sdk.Coin,
) Params {
	return Params{
		MintDenom:      mintDenom,
		BlocksPerMonth: blocksPerMonth,
		MonthReward:    monthReward,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:      sdk.DefaultBondDenom,
		BlocksPerMonth: uint64(60 * 60 * 24 * 30 / 5), // assuming 5 second block times
		MonthReward:    sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(223696209754194)),
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateBlocksPerMonth(p.BlocksPerMonth); err != nil {
		return err
	}
	if err := validateMonthReward(p.MonthReward); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyBlocksPerMonth, &p.BlocksPerMonth, validateBlocksPerMonth),
		paramtypes.NewParamSetPair(KeyMonthReward, &p.MonthReward, validateMonthReward),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateBlocksPerMonth(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}

func validateMonthReward(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !v.Amount.IsPositive() {
		return fmt.Errorf("month reward must be positive: %d", v)
	}

	return nil
}
