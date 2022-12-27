package smartcard

import (
	log "github.com/sirupsen/logrus"
)

/*
Defines a custom logrus log formatter in order to print APDUs in a format which the eclipse Javacard debugger can utilize
and allow the program to turn output on and off based on log level
*/

type APDUDebugFormatter struct{}

func (f *APDUDebugFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}
