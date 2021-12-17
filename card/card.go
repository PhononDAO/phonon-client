package card

import (
	"github.com/GridPlus/keycard-go/io"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/usb"
)

func Connect(readerIndex int) (*PhononCommandSet, error) {
	scard, err := usb.ConnectUSBReader(0)
	if err != nil {
		return nil, err
	}
	cs := NewPhononCommandSet(io.NewNormalChannel(scard))
	return cs, nil
}

/*QuickSecureConnection is a convenience function which establishes a connection to the card attached
to the readerIndex given and immediately attempts to open a secure channel with it.
Equivalent to running SELECT, PAIR, OPEN_SECURE_CHANNEL.
Does not handle the details of uninitialized cards*/
func QuickSecureConnection(readerIndex int, isStatic bool) (cs model.PhononCard, err error) {
	baseCS, err := Connect(readerIndex)
	if err != nil {
		return nil, err
	}
	if isStatic {
		cs = NewStaticPhononCommandSet(baseCS)
	} else {
		cs = baseCS
	}
	err = cs.OpenSecureConnection()
	if err != nil {
		return nil, err
	}
	return cs, nil
}
