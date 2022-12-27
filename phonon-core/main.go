package main

import (
	"github.com/GridPlus/phonon-core/pkg/backend/mock"
	"github.com/GridPlus/phonon-core/pkg/orchestrator"
)

func main() {
	// parse arguments
	// parse configuration

	// start repl
	phonTerm := orchestrator.NewPhononTerminal()
	mockBackend := mock.NewMockBackend()
	phonTerm.AddBackend(mockBackend)
	phonTerm.RefreshSessions()
}
