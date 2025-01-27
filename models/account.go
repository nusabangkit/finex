package models

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/nusabangkit/finex/config"
	"github.com/nusabangkit/finex/types"
	"gorm.io/gorm"
)

var Zero float64 = 0

type Account struct {
	MemberID   int64             `json:"member_id"`
	CurrencyID string            `json:"currency_id"`
	Balance    decimal.Decimal   `json:"balance" gorm:"default:0" validate:"ValidateBalance"`
	Locked     decimal.Decimal   `json:"locked" gorm:"default:0" validate:"ValidateLocked"`
	Type       types.AccountType `json:"type" gorm:"default:spot"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

func (a Account) ValidateBalance(Balance decimal.Decimal) bool {
	return Balance.GreaterThanOrEqual(decimal.Zero)
}

func (a Account) ValidateLocked(Locked decimal.Decimal) bool {
	return Locked.GreaterThanOrEqual(decimal.Zero)
}

func (a *Account) Currency() *Currency {
	var currency *Currency

	config.DataBase.First(&currency, "id = ?", a.CurrencyID)

	return currency
}

func (a *Account) Member() *Member {
	var member *Member

	config.DataBase.First(&member, "id = ?", a.MemberID)

	return member
}

func (a *Account) TriggerEvent() {
	member := a.Member()

	config.RangoClient.EnqueueEvent("private", member.UID, "balance", a.ToJSON())
}

func (a *Account) PlusFunds(tx *gorm.DB, amount decimal.Decimal) error {
	if !amount.IsPositive() {
		return fmt.Errorf("cannot add funds (member id: %d, currency id: %s, amount: %s, balance: %s)", a.MemberID, a.CurrencyID, amount.String(), a.Balance.String())
	}

	tx = tx.Model(a).Where("currency_id = ? AND member_id = ?", a.CurrencyID, a.MemberID).Updates(Account{Balance: a.Balance.Add(amount)})
	a.TriggerEvent()
	return tx.Error
}

func (a *Account) PlusLockedFunds(tx *gorm.DB, amount decimal.Decimal) error {
	if !amount.IsPositive() {
		return fmt.Errorf("cannot add funds (member id: %d, currency id: %s, amount: %s, locked: %s)", a.MemberID, a.CurrencyID, amount.String(), a.Locked.String())
	}

	tx = tx.Model(a).Where("currency_id = ? AND member_id = ?", a.CurrencyID, a.MemberID).Updates(Account{Locked: a.Locked.Add(amount)})
	a.TriggerEvent()
	return tx.Error
}

func (a *Account) SubFunds(tx *gorm.DB, amount decimal.Decimal) error {
	if !amount.IsPositive() || amount.GreaterThan(a.Balance) {
		return fmt.Errorf("cannot subtract funds (member id: %d, currency id: %s, amount: %s, balance: %s)", a.MemberID, a.CurrencyID, amount.String(), a.Balance.String())
	}

	tx = tx.Model(a).Where("currency_id = ? AND member_id = ?", a.CurrencyID, a.MemberID).Updates(Account{Balance: a.Balance.Sub(amount)})
	a.TriggerEvent()
	return tx.Error
}

func (a *Account) LockFunds(tx *gorm.DB, amount decimal.Decimal) error {
	if !amount.IsPositive() || amount.GreaterThan(a.Balance) {
		return fmt.Errorf("cannot lock funds (member id: %d, currency id: %s, amount: %s, balance: %s, locked: %s)", a.MemberID, a.CurrencyID, amount.String(), a.Balance.String(), a.Locked.String())
	}

	tx = tx.Model(a).Where("currency_id = ? AND member_id = ?", a.CurrencyID, a.MemberID).Updates(Account{Balance: a.Balance.Sub(amount), Locked: a.Locked.Add(amount)})
	a.TriggerEvent()
	return tx.Error
}

func (a *Account) UnlockFunds(tx *gorm.DB, amount decimal.Decimal) error {
	if !amount.IsPositive() || amount.GreaterThan(a.Locked) {
		return fmt.Errorf("cannot unlock funds (member id: %d, currency id: %s, amount: %s, balance: %s, locked: %s)", a.MemberID, a.CurrencyID, amount.String(), a.Balance.String(), a.Locked.String())
	}

	tx = tx.Model(a).Where("currency_id = ? AND member_id = ?", a.CurrencyID, a.MemberID).Updates(Account{Balance: a.Balance.Add(amount), Locked: a.Locked.Sub(amount)})
	a.TriggerEvent()
	return tx.Error
}

func (a *Account) UnlockAndSubFunds(tx *gorm.DB, amount decimal.Decimal) error {
	if !amount.IsPositive() || amount.GreaterThan(a.Locked) {
		return fmt.Errorf("cannot unlock and sub funds (member id: %d, currency id: %s, amount: %s, balance: %s, locked: %s)", a.MemberID, a.CurrencyID, amount.String(), a.Balance.String(), a.Locked.String())
	}

	tx = tx.Model(a).Where("currency_id = ? AND member_id = ?", a.CurrencyID, a.MemberID).Updates(Account{Locked: a.Locked.Sub(amount)})
	a.TriggerEvent()
	return tx.Error
}

func (a *Account) Amount() decimal.Decimal {
	return a.Balance.Add(a.Locked)
}

type AccountJSON struct {
	Currency string          `json:"currency"`
	Balance  decimal.Decimal `json:"balance"`
	Locked   decimal.Decimal `json:"locked"`
}

func (a *Account) ToJSON() AccountJSON {
	return AccountJSON{
		Currency: a.CurrencyID,
		Balance:  a.Balance,
		Locked:   a.Locked,
	}
}
