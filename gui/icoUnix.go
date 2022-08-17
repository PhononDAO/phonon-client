//go:build !windows
// +build !windows

package gui

import (
	_ "embed"
)

//go:embed icons/phonon.png
var phononLogo []byte

//go:embed icons/x.png
var xIcon []byte
