//go:build windows
// +build windows

package gui

import (
	_ "embed"
)

//go:embed icons/phonon.ico
var phononLogo []byte

//go:embed icons/x.ico
var xIcon []byte
