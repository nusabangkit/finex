package models

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/nusabangkit/finex/config"
)

type Liability struct {
	ID            int64           `json:"id"`
	Code          int32           `json:"code"`
	CurrencyID    string          `json:"currency_id"`
	MemberID      int64           `json:"member_id"`
	ReferenceType string          `json:"reference_type"`
	ReferenceID   int64           `json:"reference_id"`
	Debit         decimal.Decimal `json:"debit"`
	Credit        decimal.Decimal `json:"credit"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func GetOperationsCode(currency *Currency, kind string) int32 {
	var operations_account OperationsAccount
	config.DataBase.Where("type = ? AND kind = ? AND currency_type = ?", TypeLiability, kind, currency.Type).Find(&operations_account)

	return operations_account.Code
}

func LiabilityCredit(amount decimal.Decimal, currency *Currency, reference Reference, kind string, member_id int64) {
	code := GetOperationsCode(currency, kind)

	liability := Liability{
		Code:          code,
		CurrencyID:    currency.ID,
		ReferenceType: reference.Type,
		ReferenceID:   reference.ID,
		Credit:        amount,
		MemberID:      member_id,
	}

	config.DataBase.Create(&liability)
}

func LiabilityDebit(amount decimal.Decimal, currency *Currency, reference Reference, kind string, member_id int64) {
	code := GetOperationsCode(currency, kind)

	liability := Liability{
		Code:          code,
		CurrencyID:    currency.ID,
		ReferenceType: reference.Type,
		ReferenceID:   reference.ID,
		Debit:         amount,
		MemberID:      member_id,
	}

	config.DataBase.Create(&liability)
}

func LiabilityTranfer(amount decimal.Decimal, currency *Currency, reference Reference, from_kind, to_kind string, member_id int64) {
	LiabilityCredit(amount, currency, reference, from_kind, member_id)
	LiabilityDebit(amount, currency, reference, to_kind, member_id)
}
