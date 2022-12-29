package validator

import (
	"errors"

	"github.com/GridPlus/phonon-client/chain"
	"github.com/GridPlus/phonon-client/model"
)

var ErrUnsupportedCurrency = errors.New("unsupported currency type")
var ErrEmptyProposal = errors.New("there were no phonon(s) to validate")

type AssetValidationResult struct {
	P     *model.Phonon
	Valid bool
	Err   error
}

type AssetValidator interface {
	Validate(proposal []*model.Phonon) ([]*AssetValidationResult, error)
}

type Validator struct {
	ethValidator   *chain.EthChainService
	assetValidator AssetValidator
}

func NewValidator(assetValidator AssetValidator) *Validator {
	return &Validator{assetValidator: assetValidator}
}

func (v *Validator) Validate(proposal []*model.Phonon) ([]*AssetValidationResult, error) {
	result := []*AssetValidationResult{}

	if len(proposal) == 0 {
		return nil, ErrEmptyProposal
	}

	for _, p := range proposal {
		switch p.CurrencyType {
		case 2:
			result, err := v.ethValidator.Validate(p)
			if err != nil {
				return nil, err
			}

			result = append(result, result...)
		default:
			return nil, ErrUnsupportedCurrency
		}
	}

	return result, nil
}
