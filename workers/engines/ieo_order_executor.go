package engines

import (
	"encoding/json"

	"github.com/nusabangkit/finex/config"
	"github.com/nusabangkit/finex/models"
)

type IEOOrderExecutorWorker struct {
}

func NewIEOOrderExecutorWorker() *IEOOrderExecutorWorker {
	return &IEOOrderExecutorWorker{}
}

func (w *IEOOrderExecutorWorker) Process(payload []byte) error {
	var payload_ieo_order_message *models.IEOOrderJSON
	if err := json.Unmarshal(payload, &payload_ieo_order_message); err != nil {
		return err
	}

	var ieo_order *models.IEOOrder
	config.DataBase.First(&ieo_order, "id = ?", payload_ieo_order_message.ID)

	if err := ieo_order.Strike(); err != nil {
		return err
	}

	return nil
}
