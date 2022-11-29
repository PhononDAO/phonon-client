package card

import (
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/tlv"
	log "github.com/sirupsen/logrus"
)

type MockPhonon struct {
	model.Phonon
	PrivateKey    []byte
	deleted       bool
	markedForSend bool
}

func (phonon *MockPhonon) EncodePhononForTransferProposal() (tlv.TLV, error) {
	curveTypeTLV, err := tlv.NewTLV(TagCurveType, []byte{byte(phonon.CurveType)})
	if err != nil {
		return tlv.TLV{}, err
	}
	pubKeyTLV, err := tlv.NewTLV(TagPhononPubKey, phonon.PubKey.Bytes())
	if err != nil {
		return tlv.TLV{}, err
	}
	//Encode internal phonon
	phononTLV, err := TLVEncodePhononDescriptor(&phonon.Phonon)
	if err != nil {
		log.Error("mock could not encode inner phonon: ", phonon.Phonon)
		return tlv.TLV{}, err
	}
	data := append(curveTypeTLV.Encode(), phononTLV...)
	data = append(data, pubKeyTLV.Encode()...)
	phononDescriptionTLV, err := tlv.NewTLV(TagPhononProposalDescriptor, data)
	if err != nil {
		log.Error("mock could not encode phonon description: ", err)
		return tlv.TLV{}, err
	}
	return phononDescriptionTLV, nil
}

func (phonon *MockPhonon) EncodeForFinalSend() (tlv.TLV, error) {
	privKeyTLV, err := tlv.NewTLV(TagPhononPrivKey, phonon.PrivateKey)
	if err != nil {
		log.Error("could not encode mockPhonon privKey: ", err)
		return tlv.TLV{}, err
	}
	//Also include CurveType
	curveTypeTLV, err := tlv.NewTLV(TagCurveType, []byte{byte(phonon.CurveType)})
	if err != nil {
		return tlv.TLV{}, err
	}
	//Encode internal phonon
	phononTLV, err := TLVEncodePhononDescriptor(&phonon.Phonon)
	if err != nil {
		log.Error("mock could not encode inner phonon: ", phonon.Phonon)
		return tlv.TLV{}, err
	}
	data := append(privKeyTLV.Encode(), curveTypeTLV.Encode()...)
	data = append(data, phononTLV...)
	phononDescriptionTLV, err := tlv.NewTLV(TagPhononPrivateDescription, data)
	if err != nil {
		log.Error("mock could not encode phonon description: ", err)
		return tlv.TLV{}, err
	}

	return phononDescriptionTLV, nil
}
