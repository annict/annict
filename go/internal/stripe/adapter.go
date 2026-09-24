package stripe

import (
	"context"
	"database/sql"
	"time"

	stripego "github.com/stripe/stripe-go/v84"
)

// SubscriptionはadapterがUseCaseに公開するStripeサブスクリプションの
// ドメイン形ビュー。UseCaseが必要とするフィールドのみを持ち、stripe-goの型が
// UseCaseシグネチャに漏れないようにする。
type Subscription struct {
	Status     string
	Items      []SubscriptionItem
	CancelAt   sql.NullTime
	CanceledAt sql.NullTime
}

// SubscriptionItemはサブスクリプションアイテム1件のドメイン形ビュー。
type SubscriptionItem struct {
	PriceID            string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
}

// CheckoutSessionParamsはCheckoutセッション作成に必要な入力を保持する。
// adapterはこのドメイン形の値からstripe-goのパラメータを組み立てる。
type CheckoutSessionParams struct {
	PriceID    string
	SuccessURL string
	CancelURL  string
	// UserIDはmetadataのuser_idに入れる値。
	UserID string
	// Localeは "ja" または "en"。
	Locale string
}

// PortalSessionParamsはBilling Portalセッション作成に必要な入力を保持する。
type PortalSessionParams struct {
	CustomerID string
	ReturnURL  string
	// Localeは "ja" または "en"。
	Locale string
}

// Adapterはstripe-goクライアントをラップし、各Stripe操作を呼び出し側
// (UseCase) のinterfaceに適合させて、stripe-goの型をUseCaseシグネチャから
// 締め出す。1つのadapterがUseCaseごとに定義された3つの小さなinterface
// (SubscriptionRetriever / CheckoutSessionCreator / PortalSessionCreator) を
// 実装するが、各UseCaseは自分のinterfaceのみに依存する。
type Adapter struct {
	client *stripego.Client
}

// NewAdapterは渡されたstripe-goクライアントをラップするAdapterを作成する。
func NewAdapter(client *stripego.Client) *Adapter {
	return &Adapter{client: client}
}

// RetrieveSubscriptionはStripeからサブスクリプションを取得し、ドメイン形の
// Subscriptionに変換する。
func (a *Adapter) RetrieveSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	sub, err := a.client.V1Subscriptions.Retrieve(ctx, subscriptionID, nil)
	if err != nil {
		return nil, err
	}
	return newSubscriptionFromStripe(sub), nil
}

// newSubscriptionFromStripeはstripe-goのサブスクリプションをドメイン形の
// Subscriptionに変換する。RetrieveSubscriptionから切り出すことで、変換処理
// (Unix時刻変換・アイテムのマッピング) を実Stripeクライアント無しに単体テスト
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

// CreateCheckoutSessionはStripe Checkoutセッションを作成し、そのURLを返す。
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

// CreatePortalSessionはStripe Billing Portalセッションを作成し、そのURLを返す。
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
