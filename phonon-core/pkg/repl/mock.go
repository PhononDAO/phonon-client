package repl

import (
	"fmt"

	"github.com/GridPlus/phonon-core/pkg/backend/mock"
	"github.com/abiosoft/ishell/v2"
)

func addMock(_ *ishell.Context) {
	c, err := mock.NewMockCard(true, false)
	terminal.NewSession(c)
	if err != nil {
		fmt.Println("Error generating mock phonon card")
	}

}
