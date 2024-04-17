package main

import (
	"fmt"
	"os"

	"github.com/nusabangkit/finex/config"
	"github.com/nusabangkit/finex/workers/daemons"
)

func CreateWorker(id string) daemons.Worker {
	switch id {
	case "cron_job":
		return daemons.NewCronJob()
	default:
		return nil
	}
}

func main() {
	if err := config.InitializeConfig(); err != nil {
		fmt.Println(err.Error())
		return
	}

	ARVG := os.Args[1:]

	for _, id := range ARVG {
		fmt.Println("Start finex-daemon: " + id)
		worker := CreateWorker(id)

		worker.Start()
	}
}
