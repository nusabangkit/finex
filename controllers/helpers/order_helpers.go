package helpers

import (
	"github.com/gookit/validate"
	"github.com/shopspring/decimal"

	"github.com/nusabangkit/pkg"
	GrpcEngine "github.com/nusabangkit/pkg/Grpc/engine"
	GrpcSymbol "github.com/nusabangkit/pkg/Grpc/symbol"
	GrpcUtils "github.com/nusabangkit/pkg/Grpc/utils"
	clientEngine "github.com/nusabangkit/pkg/client/engine"

	"github.com/nusabangkit/finex/config"
	"github.com/nusabangkit/finex/models"
	"github.com/nusabangkit/finex/types"
)

type CreateOrderParams struct {
	Market    string              `json:"market" form:"market" validate:"required"`
	Side      types.OrderSide     `json:"side" form:"side" validate:"required|VaildateSide"`
	OrdType   types.OrderType     `json:"ord_type" form:"ord_type" validate:"VaildateOrdType"`
	Price     decimal.NullDecimal `json:"price" form:"price" validate:"VaildatePrice"`
	StopPrice decimal.NullDecimal `json:"stop_price" form:"stop_price" validate:"VaildateStopPrice"`
	Quantity  decimal.NullDecimal `json:"quantity" form:"quantity"`
	Volume    decimal.NullDecimal `json:"volume" form:"volume"`
}

func (p CreateOrderParams) Messages() map[string]string {
	invalid_message := "market.order.invalid_{field}"

	return validate.MS{
		"required":          invalid_message,
		"VaildateSide":      invalid_message,
		"VaildatePrice":     "market.order.non_positive_price",
		"VaildateStopPrice": "market.order.non_positive_stop_price",
		"VaildateVolume":    "market.order.non_positive_volume",
	}
}

func (p CreateOrderParams) VaildatePrice(Price decimal.NullDecimal) bool {
	if Price.Valid {
		return Price.Decimal.IsPositive()
	}

	return true
}

func (p CreateOrderParams) VaildateStopPrice(StopPrice decimal.NullDecimal) bool {
	if StopPrice.Valid {
		return StopPrice.Decimal.IsPositive()
	}

	return true
}

func (p CreateOrderParams) VaildateOrdType(OrdType types.OrderType) bool {
	if OrdType == types.TypeMarket && (p.Price.Valid || p.StopPrice.Valid) {
		return false
	} else if OrdType == types.TypeLimit && !p.Price.Valid {
		return false
	}

	return true
}

func (p CreateOrderParams) VaildateVolume(Volume decimal.Decimal) bool {
	return Volume.IsPositive()
}

func (p CreateOrderParams) VaildateSide(val types.OrderSide) bool {
	return p.Side == types.SideBuy || p.Side == types.SideSell
}

func (p CreateOrderParams) GetMarket() models.Market {
	var market models.Market

	config.DataBase.First(&market, "symbol = ?", p.Market)

	return market
}

func (p CreateOrderParams) BuildOrder(member *models.Member, err_src *Errors) *models.Order {
	var order_side models.OrderSide
	market := p.GetMarket()

	if len(p.OrdType) == 0 {
		p.OrdType = types.TypeLimit
	}

	if p.Side == types.SideBuy {
		order_side = models.SideBuy
	} else {
		order_side = models.SideSell
	}

	if len(p.OrdType) == 0 {
		p.OrdType = types.TypeLimit
	}

	trading_fee := models.TradingFeeFor(member.Group, "spot", market.Symbol)
	var quantity decimal.Decimal
	var locked decimal.Decimal

	if p.OrdType == types.TypeMarket {
		var side pkg.OrderSide

		if p.Side == types.SideBuy {
			side = pkg.SideBuy
		} else {
			side = pkg.SideSell
		}

		matching_client := clientEngine.NewMatchingClient()
		defer matching_client.Close()

		symbol := market.GetSymbol()

		calc_market_order_response, err := matching_client.CalcMarketOrder(&GrpcEngine.CalcMarketOrderRequest{
			Symbol: &GrpcSymbol.Symbol{BaseCurrency: symbol.BaseCurrency, QuoteCurrency: symbol.QuoteCurrency},
			Side:   string(side),
			Quantity: &GrpcUtils.Decimal{
				Val: p.Quantity.Decimal.CoefficientInt64(),
				Exp: p.Quantity.Decimal.Exponent(),
			},
			// Volume: &engineGrpc.Decimal{
			// 	Val: p.Volume.Decimal.CoefficientInt64(),
			// 	Exp: p.Volume.Decimal.Exponent(),
			// },
		})
		if err != nil {
			err_src.Errors = append(err_src.Errors, "market.order.insufficient_market_liquidity")

			return nil
		}

		quantity = calc_market_order_response.Quantity.ToDecimal()
		locked = calc_market_order_response.Locked.ToDecimal()

		if quantity.IsZero() || locked.IsZero() {
			err_src.Errors = append(err_src.Errors, "market.order.insufficient_market_liquidity")
		}
	} else {
		quantity = p.Quantity.Decimal
		if p.Side == types.SideBuy {
			locked = p.Price.Decimal.Mul(p.Quantity.Decimal)
		} else {
			locked = p.Quantity.Decimal
		}
	}

	if err_src.Size() > 0 {
		return nil
	}

	order := &models.Order{
		MemberID:     member.ID,
		Ask:          market.BaseUnit,
		Bid:          market.QuoteUnit,
		MarketID:     market.Symbol,
		MarketType:   types.AccountTypeSpot,
		OrdType:      p.OrdType,
		State:        models.StatePending,
		Type:         order_side,
		Price:        p.Price,
		StopPrice:    p.StopPrice,
		Volume:       quantity,
		MakerFee:     trading_fee.Maker,
		TakerFee:     trading_fee.Taker,
		OriginVolume: quantity,
		Locked:       locked,
		OriginLocked: locked,
	}

	Vaildate(order, err_src)

	return order
}

func (p CreateOrderParams) CreateOrder(member *models.Member, err_src *Errors) (order *models.Order) {
	order = p.BuildOrder(member, err_src)

	if len(err_src.Errors) > 0 {
		return
	}

	if err := config.DataBase.Create(&order).Error; err != nil {
		err_src.Errors = append(err_src.Errors, "market.order.invalid_volume_or_price")

		return nil
	}
	if err := order.Submit(); err != nil {
		err_src.Errors = append(err_src.Errors, err.Error())

		return nil
	}

	return order
}
