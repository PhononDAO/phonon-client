package smartcard

import (
	"github.com/GridPlus/keycard-go/io"
	"github.com/GridPlus/phonon-core/pkg/backend/smartcard/usb"
	"github.com/GridPlus/phonon-core/pkg/model"
	"github.com/sirupsen/logrus"
)

func Connect(readerIndex int, cert []byte, logger logrus.Logger) (*PhononCommandSet, error) {
	scard, err := usb.ConnectUSBReader(readerIndex)
	if err != nil {
		return nil, err
	}
	cs := NewPhononCommandSet(io.NewNormalChannel(scard), cert, logger)
	return cs, nil
}

/*
QuickSecureConnection is a convenience function which establishes a connection to the card attached
to the readerIndex given and immediately attempts to open a secure channel with it.
Equivalent to running SELECT, PAIR, OPEN_SECURE_CHANNEL.
Does not handle the details of uninitialized cards
*/
func QuickSecureConnection(readerIndex int, cert []byte, logger logrus.Logger) (cs model.PhononCard, err error) {
	baseCS, err := Connect(readerIndex, cert, logger)
	if err != nil {
		return nil, err
	}
	cs = baseCS
	err = cs.OpenSecureConnection()
	if err != nil {
		return nil, err
	}
	return cs, nil
}
