package pay

import (
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"time"
)

func InitStripe() {
	stripe.Key = viper.GetString("stripe.key")
}
func StripeTrade(subject, orderID, amount string) (string, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return "", err
	}
	successURL := viper.GetString("stripe.successUrl")
	expire := time.Now().Add(30 * time.Minute).Unix()

	params := &stripe.CheckoutSessionParams{
		ExpiresAt: &expire,
		Mode:      stripe.String(string(stripe.CheckoutSessionModePayment)),
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
