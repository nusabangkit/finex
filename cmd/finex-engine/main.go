package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/nusabangkit/finex/config"
	"github.com/nusabangkit/finex/workers/engines"
	"github.com/nusabangkit/pkg/services"
)

func CreateWorker(id string) engines.Worker {
	switch id {
	case "order_processor":
		return engines.NewOrderProcessorWorker()
	case "trade_executor":
		return engines.NewTradeExecutorWorker()
	case "ieo_order_processor":
		return engines.NewIEOOrderProcessorWorker()
	case "ieo_order_executor":
		return engines.NewIEOOrderExecutorWorker()
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
	id := ARVG[0]
	consumer, err := services.NewKafkaConsumer(strings.Split(os.Getenv("KAFKA_URL"), ","), uuid.NewString(), []string{id})
	if err != nil {
		panic(err)
	}

	fmt.Println("Start finex-engine: " + id)
	worker := CreateWorker(id)

	defer consumer.Close()

	for {
		records, err := consumer.Poll()
		if err != nil {
			config.Logger.Fatalf("Failed to poll consumer %v", err)
		}

		for _, record := range records {
			if record.Topic != id {
				continue
			}

			config.Logger.Debugf("Recevie message from topic: %s payload: %s", record.Topic, string(record.Value))
			err := worker.Process(record.Value)

			if err != nil {
				config.Logger.Errorf("Worker error: %v", err.Error())
			}

			consumer.CommitRecords(*record)
		}
	}
}
