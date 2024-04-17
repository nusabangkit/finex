package main

import (
	"github.com/nusabangkit/finex/config"
	"github.com/nusabangkit/finex/routes"
)

func main() {
	if err := config.InitializeConfig(); err != nil {
		config.Logger.Error(err.Error())
		return
	}

	r := routes.SetupRouter()
	// running
	r.Listen(":3000")
}
