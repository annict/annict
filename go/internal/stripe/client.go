// Package stripe はStripe APIとの連携機能を提供します
package stripe

import (
	"github.com/stripe/stripe-go/v84"
)

// Config はStripe関連の設定を保持します
type Config struct {
	SecretKey      string // Stripe Secret Key (sk_xxx)
	PublishableKey string // Stripe Publishable Key (pk_xxx)
	WebhookSecret  string // Webhook署名検証用シークレット (whsec_xxx)
	PriceMonthlyID string // 月額プランの価格ID (price_xxx)
	PriceYearlyID  string // 年額プランの価格ID (price_xxx)
}

// NewClient はStripe APIクライアントを作成します
func NewClient(secretKey string) *stripe.Client {
	return stripe.NewClient(secretKey)
}
