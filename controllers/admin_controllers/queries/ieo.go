package queries

import (
	"github.com/shopspring/decimal"
	"github.com/nusabangkit/finex/types"
)

type IEOPayload struct {
	ID                  int64             `json:"id"`
	CurrencyID          string            `json:"currency_id"`
	Description         string            `json:"description"`
	MainPaymentCurrency string            `json:"main_payment_currency"`
	Price               decimal.Decimal   `json:"price"`
	OriginQuantity      decimal.Decimal   `json:"origin_quantity"`
	LimitPerUser        decimal.Decimal   `json:"limit_per_user"`
	MinAmount           decimal.Decimal   `json:"min_amount"`
	State               types.MarketState `json:"state"`
	StartTime           int64             `json:"start_time"`
	BannerUrl           string            `json:"banner_url"`
	EndTime             int64             `json:"end_time"`
	Data                string            `json:"data"`
}
