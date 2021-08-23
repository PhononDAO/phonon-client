package card

import (
	"errors"

	"github.com/GridPlus/keycard-go/io"
	"github.com/ebfe/scard"
	log "github.com/sirupsen/logrus"
)

var ErrReaderNotFound = errors.New("card reader not found")

//Connects to the first card reader listed by default
func Connect() (*PhononCommandSet, error) {
	return ConnectWithReaderIndex(0)
}


func ConnectWithContext(ctx *scard.Context, index int)(*PhononCommandSet, error){
	readers, err := ctx.ListReaders()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	for i, reader := range readers {
		log.Debugf("[%d] %s\n", i, reader)
	}

	if len(readers) > index {
		card, err := ctx.Connect(readers[index], scard.ShareShared, scard.ProtocolAny)
		if err != nil {
			log.Error(err)
		}
		// defer card.Disconnect(scard.ResetCard)

		log.Debug("Card status:")
		status, err := card.Status()
		if err != nil {
			log.Error(err)
		}

		log.Debugf("\treader: %s\n\tstate: %x\n\tactive protocol: %x\n\tatr: % x\n",
			status.Reader, status.State, status.ActiveProtocol, status.Atr)
		cs :=  NewPhononCommandSet(io.NewNormalChannel(card))
		_, _, _, err = cs.Select()
		return cs, err
	}
	return nil, ErrReaderNotFound
}

func ConnectWithReaderIndex(index int) (*PhononCommandSet, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return ConnectWithContext(ctx, index)
}

//Connects and Opens a Secure Connection with a card
func OpenSecureConnection() (*PhononCommandSet, error) {
	cs, err := Connect()
	if err != nil {
		log.Error("could not connect to card: ", err)
	}
	_, _, _, err = cs.Select()
	if err != nil {
		log.Error("could not select phonon applet: ", err)
		return nil, err
	}
	err = cs.Pair()
	if err != nil {
		log.Error("could not pair: ", err)
		return nil, err
	}
	err = cs.OpenSecureChannel()
	if err != nil {
		log.Error("could not open secure channel: ", err)
		return nil, err
	}
	return cs, nil
}

//Connects to a card and checks it's initialization status
//If uninitialized, opens a normal channel
//If initialized, opens a secure channel
//Uses default reader index 0
func OpenBestConnection() (cs *PhononCommandSet, initalized bool, err error) {
	return OpenBestConnectionWithReaderIndex(0)
}

//Connects to a card and checks it's initialization status
//If uninitialized, opens a normal channel
//If initialized, opens a secure channel
func OpenBestConnectionWithReaderIndex(index int) (cs *PhononCommandSet, initalized bool, err error) {
	cs, err = ConnectWithReaderIndex(index)
	if err != nil {
		return nil, false, err
	}
	_, _, initialized, err := cs.Select()
	if !initialized {
		return cs, false, err
	}
	err = cs.Pair()
	if err != nil {
		return nil, false, err
	}
	err = cs.OpenSecureChannel()
	if err != nil {
		return nil, false, err
	}
	return cs, initialized, nil
}
