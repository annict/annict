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

// Init はStripe APIクライアントを初期化します
// Stripe SDKはグローバルにAPIキーを設定する方式を採用しているため、
// この関数を呼び出すことでパッケージ全体でStripe APIが利用可能になります
func Init(cfg *Config) {
	stripe.Key = cfg.SecretKey
}
