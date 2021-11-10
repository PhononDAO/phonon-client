package repl

import (
	"fmt"

	"github.com/abiosoft/ishell/v2"
)

func addMock(c *ishell.Context){
	err := t.GenerateMock()
	if err != nil{
		fmt.Println("Error generating mock phonon card")
	}

}
