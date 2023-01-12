package model

import (
	"errors"
)

var ErrUnsupportedCurrency = errors.New("unsupported currency type")
var ErrEmptyProposal = errors.New("there were no phonon(s) to validate")
var ErrBalanceTooLow = errors.New("phonon balance is less than what is stated in denomination value")
var ErrBalanceTooHigh = errors.New("phonon balance is greater than what is stated in denomination value")
var ErrInvalidChainID = errors.New("eth chainID unsupported")
var ErrInvalidEthAddress = errors.New("invalid ethereum address")

type AssetValidationResult struct {
	P     *Phonon
	Valid bool
	Err   error
}

type AssetValidator interface {
	Validate(proposal []*Phonon) ([]*AssetValidationResult, error)
}
