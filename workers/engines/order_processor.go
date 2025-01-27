package engines

import (
	"encoding/json"

	"github.com/nusabangkit/finex/config"
	"github.com/nusabangkit/finex/models"
	"github.com/nusabangkit/pkg"
)

type OrderProcessorPayloadMessage struct {
	Action pkg.PayloadAction `json:"action"`
	ID     int64             `json:"id"`
}

type OrderProcessorWorker struct {
}

func NewOrderProcessorWorker() *OrderProcessorWorker {
	kclass := &OrderProcessorWorker{}

	var orders []models.Order
	config.DataBase.Where("state = ?", models.StatePending).Find(&orders)
	for _, order := range orders {
		if err := models.SubmitOrder(order.ID); err != nil {
			config.Logger.Errorf("Error: %s", err.Error())
			break
		}
	}

	return kclass
}

func (w OrderProcessorWorker) Process(payload []byte) error {
	var order_processor_payload OrderProcessorPayloadMessage
	err := json.Unmarshal(payload, &order_processor_payload)
	if err != nil {
		return err
	}

	id := order_processor_payload.ID

	switch order_processor_payload.Action {
	case pkg.ActionSubmit:
		err = models.SubmitOrder(id)
	case pkg.ActionCancel:
		err = models.CancelOrder(id)
	}

	if err != nil {
		return err
	}

	return nil
}
