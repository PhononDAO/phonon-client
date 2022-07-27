package repl

import (
	"fmt"

	"github.com/abiosoft/ishell/v2"
)

func addMock(_ *ishell.Context) {
	_, err := t.GenerateMock()
	if err != nil {
		fmt.Println("Error generating mock phonon card")
	}

}
