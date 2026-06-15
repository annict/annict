package stripe

import (
	"context"
	"database/sql"
	"time"

	stripego "github.com/stripe/stripe-go/v84"
)

// Subscription is the domain-shaped view of a Stripe subscription that the
// adapter exposes to UseCases. It carries only the fields the UseCases need so
// that stripe-go types never leak into UseCase signatures.
//
// [Ja] Subscription は adapter が UseCase に公開する Stripe サブスクリプションの
// ドメイン形ビュー。UseCase が必要とするフィールドのみを持ち、stripe-go の型が
// UseCase シグネチャに漏れないようにする。
type Subscription struct {
	Status     string
	Items      []SubscriptionItem
	CancelAt   sql.NullTime
	CanceledAt sql.NullTime
}

// SubscriptionItem is the domain-shaped view of a single subscription item.
// [Ja] SubscriptionItem はサブスクリプションアイテム 1 件のドメイン形ビュー。
type SubscriptionItem struct {
	PriceID            string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
}

// CheckoutSessionParams holds the inputs needed to create a Checkout session.
// The adapter builds the stripe-go params from these domain-shaped values.
//
// [Ja] CheckoutSessionParams は Checkout セッション作成に必要な入力を保持する。
// adapter はこのドメイン形の値から stripe-go のパラメータを組み立てる。
type CheckoutSessionParams struct {
	PriceID    string
	SuccessURL string
	CancelURL  string
	// UserID is the value placed in metadata's user_id.
	// [Ja] UserID は metadata の user_id に入れる値。
	UserID string
	// Locale is "ja" or "en".
	// [Ja] Locale は "ja" または "en"。
	Locale string
}

// PortalSessionParams holds the inputs needed to create a Billing Portal session.
// [Ja] PortalSessionParams は Billing Portal セッション作成に必要な入力を保持する。
type PortalSessionParams struct {
	CustomerID string
	ReturnURL  string
	// Locale is "ja" or "en".
	// [Ja] Locale は "ja" または "en"。
	Locale string
}

// Adapter wraps the stripe-go client and adapts each Stripe operation to the
// caller-side (UseCase) interfaces, keeping stripe-go types out of UseCase
// signatures. A single adapter implements the three small interfaces
// (SubscriptionRetriever / CheckoutSessionCreator / PortalSessionCreator)
// defined per UseCase; each UseCase still depends only on its own interface.
//
// [Ja] Adapter は stripe-go クライアントをラップし、各 Stripe 操作を呼び出し側
// (UseCase) の interface に適合させて、stripe-go の型を UseCase シグネチャから
// 締め出す。1 つの adapter が UseCase ごとに定義された 3 つの小さな interface
// (SubscriptionRetriever / CheckoutSessionCreator / PortalSessionCreator) を
// 実装するが、各 UseCase は自分の interface のみに依存する。
type Adapter struct {
	client *stripego.Client
}

// NewAdapter creates an Adapter wrapping the given stripe-go client.
// [Ja] NewAdapter は渡された stripe-go クライアントをラップする Adapter を作成する。
func NewAdapter(client *stripego.Client) *Adapter {
	return &Adapter{client: client}
}

// RetrieveSubscription fetches a subscription from Stripe and converts it to the
// domain-shaped Subscription.
//
// [Ja] RetrieveSubscription は Stripe からサブスクリプションを取得し、ドメイン形の
// Subscription に変換する。
func (a *Adapter) RetrieveSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	sub, err := a.client.V1Subscriptions.Retrieve(ctx, subscriptionID, nil)
	if err != nil {
		return nil, err
	}
	return newSubscriptionFromStripe(sub), nil
}

// newSubscriptionFromStripe converts a stripe-go subscription into the
// domain-shaped Subscription. It is split out from RetrieveSubscription so that
// the conversion (Unix-time conversion, item mapping) can be unit-tested without
// a live Stripe client.
//
// [Ja] newSubscriptionFromStripe は stripe-go のサブスクリプションをドメイン形の
// Subscription に変換する。RetrieveSubscription から切り出すことで、変換処理
// (Unix 時刻変換・アイテムのマッピング) を実 Stripe クライアント無しに単体テスト
// できるようにする。
func newSubscriptionFromStripe(sub *stripego.Subscription) *Subscription {
	items := make([]SubscriptionItem, 0, len(sub.Items.Data))
	for _, item := range sub.Items.Data {
		items = append(items, SubscriptionItem{
			PriceID:            item.Price.ID,
			CurrentPeriodStart: time.Unix(item.CurrentPeriodStart, 0),
			CurrentPeriodEnd:   time.Unix(item.CurrentPeriodEnd, 0),
		})
	}

	return &Subscription{
		Status:     string(sub.Status),
		Items:      items,
		CancelAt:   NullTimeFromUnix(sub.CancelAt),
		CanceledAt: NullTimeFromUnix(sub.CanceledAt),
	}
}

// CreateCheckoutSession creates a Stripe Checkout session and returns its URL.
// [Ja] CreateCheckoutSession は Stripe Checkout セッションを作成し、その URL を返す。
func (a *Adapter) CreateCheckoutSession(ctx context.Context, params CheckoutSessionParams) (string, error) {
	stripeParams := &stripego.CheckoutSessionCreateParams{
		Mode: stripego.String(string(stripego.CheckoutSessionModeSubscription)),
		LineItems: []*stripego.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripego.String(params.PriceID),
				Quantity: stripego.Int64(1),
			},
		},
		SuccessURL: stripego.String(params.SuccessURL),
		CancelURL:  stripego.String(params.CancelURL),
		Metadata: map[string]string{
			"user_id": params.UserID,
		},
		Locale: stripego.String(params.Locale),
	}

	session, err := a.client.V1CheckoutSessions.Create(ctx, stripeParams)
	if err != nil {
		return "", err
	}
	return session.URL, nil
}

// CreatePortalSession creates a Stripe Billing Portal session and returns its URL.
// [Ja] CreatePortalSession は Stripe Billing Portal セッションを作成し、その URL を返す。
func (a *Adapter) CreatePortalSession(ctx context.Context, params PortalSessionParams) (string, error) {
	stripeParams := &stripego.BillingPortalSessionCreateParams{
		Customer:  stripego.String(params.CustomerID),
		ReturnURL: stripego.String(params.ReturnURL),
		Locale:    stripego.String(params.Locale),
	}

	session, err := a.client.V1BillingPortalSessions.Create(ctx, stripeParams)
	if err != nil {
		return "", err
	}
	return session.URL, nil
}
