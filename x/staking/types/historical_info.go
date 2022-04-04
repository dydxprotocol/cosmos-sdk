package types

import (
	"bytes"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHistoricalInfo will create a historical information struct from header and valset
// it will first sort valset before inclusion into historical info
func NewHistoricalInfo(height int64, time time.Time, apphash, nextValidatorsHash []byte, valSet Validators, powerReduction sdk.Int) HistoricalInfo {
	// Must sort in the same way that tendermint does
	sort.SliceStable(valSet, func(i, j int) bool {
		return ValidatorsByVotingPower(valSet).Less(i, j, powerReduction)
	})

	return HistoricalInfo{
		Height:             height, // Not really needed, as it is also the key
		Time:               time,
		AppHash:            apphash,
		NextValidatorsHash: nextValidatorsHash,
		Valset:             valSet,
	}
}

// MustUnmarshalHistoricalInfo wll unmarshal historical info and panic on error
func MustUnmarshalHistoricalInfo(cdc codec.BinaryCodec, value []byte) HistoricalInfo {
	hi, err := UnmarshalHistoricalInfo(cdc, value)
	if err != nil {
		panic(err)
	}

	return hi
}

// UnmarshalHistoricalInfo will unmarshal historical info and return any error
func UnmarshalHistoricalInfo(cdc codec.BinaryCodec, value []byte) (hi HistoricalInfo, err error) {
	err = cdc.Unmarshal(value, &hi)
	return hi, err
}

// ValidateBasic will ensure HistoricalInfo is not nil and sorted
func ValidateBasic(hi HistoricalInfo) error {
	if len(hi.Valset) == 0 {
		return sdkerrors.Wrap(ErrInvalidHistoricalInfo, "validator set is empty")
	}

	if !sort.IsSorted(Validators(hi.Valset)) {
		return sdkerrors.Wrap(ErrInvalidHistoricalInfo, "validator set is not sorted by address")
	}

	return nil
}

// Equal checks if receiver is equal to the parameter
func (hi *HistoricalInfo) Equal(hi2 *HistoricalInfo) bool {
	if &hi.Height != &hi2.Height {
		return false
	}
	if !hi.Time.Equal(hi2.Time) {
		return false
	}
	if bytes.Compare(hi.AppHash, hi2.AppHash) != 0 {
		return false
	}
	if bytes.Compare(hi.NextValidatorsHash, hi2.NextValidatorsHash) != 0 {
		return false
	}
	if len(hi.Valset) != len(hi2.Valset) {
		return false
	}
	for i := range hi.Valset {
		if !hi.Valset[i].Equal(&hi2.Valset[i]) {
			return false
		}
	}
	return true
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (hi HistoricalInfo) UnpackInterfaces(c codectypes.AnyUnpacker) error {
	for i := range hi.Valset {
		if err := hi.Valset[i].UnpackInterfaces(c); err != nil {
			return err
		}
	}
	return nil
}
