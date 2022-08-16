package main

import (
	"github.com/GridPlus/phonon-client/remote/v1/server"
)

func main() {
	server.StartServer("443", "./cert", "./key")
}
