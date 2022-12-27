package usb

import (
	"errors"

	"github.com/ebfe/scard"
	log "github.com/sirupsen/logrus"
)

var ErrReaderNotFound = errors.New("card reader not found")

func ConnectAllUSBReaders() (cards []*scard.Card, err error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, err
	}
	readers, err := ctx.ListReaders()
	if err != nil {
		return nil, err
	}
	log.Debugf("readers: %v", readers)
	if len(readers) == 0 {
		return nil, ErrReaderNotFound
	}
	for _, reader := range readers {
		c, err := ctx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
		if err == nil {
			cards = append(cards, c)
		} else {
			log.Debugf("unable to connect to card on reader %v: %v\n", reader, err)
		}
	}
	return cards, nil
}

func ConnectUSBReader(i int) (*scard.Card, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, err
	}
	readers, err := ctx.ListReaders()
	if err != nil {
		return nil, err
	}
	log.Debugf("readers: %v", readers)
	if len(readers) < (i + 1) {
		return nil, ErrReaderNotFound
	}
	card, err := ctx.Connect(readers[i], scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return nil, err
	}
	return card, nil
}
