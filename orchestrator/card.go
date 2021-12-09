package orchestrator

import (
	"github.com/GridPlus/keycard-go/io"
	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/usb"
)

func Connect(readerIndex int) (*card.PhononCommandSet, error) {
	scard, err := usb.ConnectUSBReader(0)
	if err != nil {
		return nil, err
	}
	cs := card.NewPhononCommandSet(io.NewNormalChannel(scard))
	return cs, nil
}

/*QuickSecureConnection is a convenienc function which establishes a connection to the card attached
to the readerIndex given and
immediately attempts to open a secure channel with it
Does not handle the details of uninitialized cards*/
func QuickSecureConnection(readerIndex int) (cs *card.PhononCommandSet, err error) {
	cs, err = Connect(0)
	if err != nil {
		return nil, err
	}
	err = cs.OpenSecureConnection()
	if err != nil {
		return nil, err
	}
	return cs, nil
}
