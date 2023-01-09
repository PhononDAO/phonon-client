package chain

import (
	"github.com/GridPlus/phonon-client/model"
)

type MultiAssetValidator struct {
}

func NewMultiAssetValidator() *MultiAssetValidator {
	return &MultiAssetValidator{}
}

func (m *MultiAssetValidator) Validate(proposal []*model.Phonon) (result []*model.AssetValidationResult, err error) {
	if len(proposal) == 0 {
		return nil, model.ErrEmptyProposal
	}

	ethChainSrv, err := NewEthChainService()
	if err != nil {
		return nil, err
	}

	for _, p := range proposal {
		switch p.CurrencyType {
		case model.Ethereum:
			ethResult, err := ethChainSrv.Validate([]*model.Phonon{p})
			if err != nil {
				result = append(result, &model.AssetValidationResult{
					P:   p,
					Err: err,
				})
			}

			result = append(result, ethResult...)
		default:
			result = append(result, &model.AssetValidationResult{
				P:   p,
				Err: ErrUnsupportedCurrency,
			})
		}
	}

	return result, nil
}
