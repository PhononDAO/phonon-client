package card

import (
	"errors"
	"fmt"

	"github.com/GridPlus/keycard-go/io"
	"github.com/ebfe/scard"
	log "github.com/sirupsen/logrus"
)

// type Safecard keycard.CommandSet

func Connect() (*PhononCommandSet, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	readers, err := ctx.ListReaders()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for i, reader := range readers {
		fmt.Printf("[%d] %s\n", i, reader)
	}

	if len(readers) > 0 {
		card, err := ctx.Connect(readers[0], scard.ShareShared, scard.ProtocolAny)
		if err != nil {
			log.Error(err)
		}
		// defer card.Disconnect(scard.ResetCard)

		fmt.Println("Card status:")
		status, err := card.Status()
		if err != nil {
			log.Error(err)
		}

		fmt.Printf("\treader: %s\n\tstate: %x\n\tactive protocol: %x\n\tatr: % x\n",
			status.Reader, status.State, status.ActiveProtocol, status.Atr)

		// c.c = io.NewNormalChannel(card)
		// //Set card context
		// c.ctx = ctx
		// c.card = card
		return NewPhononCommandSet(io.NewNormalChannel(card)), nil
	}
	return nil, errors.New("no card reader found")
}
