package model

import (
	"errors"
)

var ErrUnsupportedCurrency = errors.New("unsupported currency type")
var ErrEmptyProposal = errors.New("there were no phonon(s) to validate")

type AssetValidationResult struct {
	P     *Phonon
	Valid bool
	Err   error
}

type AssetValidator interface {
	Validate(proposal []*Phonon) ([]*AssetValidationResult, error)
}
