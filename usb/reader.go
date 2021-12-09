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
		if err != nil {
			return nil, err
		}

		cards = append(cards, c)
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

// func ConnectAll() (sessions []*card.Session, err error) {
// 	ctx, err := scard.EstablishContext()
// 	if err != nil {
// 		return nil, err
// 	}
// 	readers, err := ctx.ListReaders()
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Debugf("readers: %v", readers)
// 	if len(readers) == 0 {
// 		return nil, ErrReaderNotFound
// 	}
// 	for _, reader := range readers {
// 		c, err := ctx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
// 		if err != nil {
// 			return nil, err
// 		}
// 		maxDataRate, _ := c.GetAttrib(scard.AttrMaxDataRate)
// 		status, err := c.Status()

// 		log.Debug("card protocol set: ", c.ActiveProtocol())
// 		session, err := card.NewSession(card.NewPhononCommandSet(io.NewNormalChannel(c)))
// 		if err != nil {
// 			return nil, err
// 		}
// 		sessions = append(sessions, session)
// 	}
// 	return sessions, nil
// }

//Connects to the first card reader listed by default
// func Connect() (*card.PhononCommandSet, error) {
// 	return ConnectWithReaderIndex(0)
// }

// func ConnectWithContext(ctx *scard.Context, index int) (*card.PhononCommandSet, error) {
// 	readers, err := ctx.ListReaders()
// 	if err != nil {
// 		log.Error(err)
// 		return nil, err
// 	}

// 	for i, reader := range readers {
// 		log.Debugf("[%d] %s\n", i, reader)
// 	}

// 	if len(readers) > index {
// 		c, err := ctx.Connect(readers[index], scard.ShareShared, scard.ProtocolAny)
// 		if err != nil {
// 			log.Error(err)
// 			return nil, err
// 		}
// 		// defer card.Disconnect(scard.ResetCard)

// 		log.Debug("Card status:")
// 		status, err := c.Status()
// 		if err != nil {
// 			log.Error(err)
// 			return nil, err
// 		}

// 		log.Debugf("\treader: %s\n\tstate: %x\n\tactive protocol: %x\n\tatr: % x\n",
// 			status.Reader, status.State, status.ActiveProtocol, status.Atr)
// 		cs := card.NewPhononCommandSet(io.NewNormalChannel(c))
// 		return cs, nil
// 	}
// 	return nil, ErrReaderNotFound
// }

// func ConnectWithReaderIndex(index int) (*card.PhononCommandSet, error) {
// 	ctx, err := scard.EstablishContext()
// 	if err != nil {
// 		log.Error(err)
// 		return nil, err
// 	}
// 	return ConnectWithContext(ctx, index)
// }

// //Connects and Opens a Secure Connection with a card
// func OpenSecureConnection() (*card.PhononCommandSet, error) {
// 	cs, err := Connect()
// 	if err != nil {
// 		log.Error("could not connect to card: ", err)
// 	}
// 	_, _, _, err = cs.Select()
// 	if err != nil {
// 		log.Error("could not select phonon applet: ", err)
// 		return nil, err
// 	}
// 	_, err = cs.Pair()
// 	if err != nil {
// 		log.Error("could not pair: ", err)
// 		return nil, err
// 	}
// 	err = cs.OpenSecureChannel()
// 	if err != nil {
// 		log.Error("could not open secure channel: ", err)
// 		return nil, err
// 	}
// 	return cs, nil
// }

//Connects to a card and checks it's initialization status
//If uninitialized, opens a normal channel
//If initialized, opens a secure channel
//Uses default reader index 0
// func OpenBestConnection() (cs *card.PhononCommandSet, initalized bool, err error) {
// 	return OpenBestConnectionWithReaderIndex(0)
// }

// //Connects to a card and checks it's initialization status
// //If uninitialized, opens a normal channel
// //If initialized, opens a secure channel
// func OpenBestConnectionWithReaderIndex(index int) (cs *card.PhononCommandSet, initalized bool, err error) {
// 	cs, err = ConnectWithReaderIndex(index)
// 	if err != nil {
// 		return nil, false, err
// 	}
// 	_, _, initialized, err := cs.Select()
// 	if !initialized {
// 		return cs, false, err
// 	}
// 	_, err = cs.Pair()
// 	if err != nil {
// 		return nil, false, err
// 	}
// 	err = cs.OpenSecureChannel()
// 	if err != nil {
// 		return nil, false, err
// 	}
// 	return cs, initialized, nil
// }
