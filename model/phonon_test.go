package model

import (
	"fmt"
	"testing"
)

func TestPrintDenomination(t *testing.T) {
	d, err := NewDenomination(100000)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(d.Value())
	fmt.Println(d)
}
