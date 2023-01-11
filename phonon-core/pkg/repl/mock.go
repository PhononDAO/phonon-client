package repl

import (
	"fmt"

	"github.com/GridPlus/phonon-core/pkg/backend/mock"
	"github.com/GridPlus/phonon-core/pkg/orchestrator"
	"github.com/abiosoft/ishell/v2"
)

func addMock(ctx *ishell.Context) {
	c, err := mock.NewMockCard(true, false)
	if err != nil {
		ctx.Err(fmt.Errorf("unable to create mock %s", err.Error()))
	}
	terminal := orchestrator.NewPhononTerminal()
	sess, err := orchestrator.NewSession(c)
	if err != nil {
		ctx.Err(fmt.Errorf("unable to generate session from newly created mock: %s", err.Error()))
		return
	}
	terminal.AddSession(sess)

}
