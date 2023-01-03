package repl

import (
	"fmt"

	"github.com/GridPlus/phonon-core/pkg/backend/mock"
	"github.com/GridPlus/phonon-core/pkg/orchestrator"
	"github.com/abiosoft/ishell/v2"
)

func addMock(ctx *ishell.Context) {
	c, err := mock.NewMockCard(true, false)
	terminal := orchestrator.NewPhononTerminal()
	sess, err := orchestrator.NewSession(c)
	if err != nil {
		ctx.Err(fmt.Errorf("unable to generate session from newly created mock: %s", err.Error))
		return
	}
	terminal.AddSession(sess)
	if err != nil {
		fmt.Println("Error generating mock phonon card")
	}

}
