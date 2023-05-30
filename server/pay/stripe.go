package pay

import (
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

func InitStripe() {
	stripe.Key = viper.GetString("stripe.key")
}
func StripeTrade(subject, orderID, amount string) (string, error) {
	successURL := viper.GetString("stripe.successUrl")
	amt, _ := decimal.NewFromString(amount)
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(orderID),
						Description: stripe.String(subject),
					},
					UnitAmount: stripe.Int64(amt.Mul(decimal.NewFromInt(100)).BigInt().Int64()),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
	}

	s, err := session.New(params)
	if err != nil {
		return "", err
	}
	return s.URL, nil
}
