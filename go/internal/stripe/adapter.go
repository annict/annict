package stripe

import (
	"context"
	"database/sql"
	"errors"
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
// 締め出す。1つのadapterがUseCaseごとに定義された小さなinterface
// (SubscriptionRetriever / CheckoutSessionCreator / PortalSessionCreatorなど) を
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

// CancelSubscriptionResultはCancelSubscriptionがどの状態で終えたかを表す。
// 退会時の解約は再試行されるため、呼び出し側は解約済み・存在しない場合も
// 成功として扱える。
type CancelSubscriptionResult int

const (
	// CancelSubscriptionResultCanceledは今回の呼び出しで解約したことを表す。
	CancelSubscriptionResultCanceled CancelSubscriptionResult = iota + 1
	// CancelSubscriptionResultAlreadyEndedは呼び出し前に解約済み (または
	// 支払いが完了せず失効済み) だったことを表す。
	CancelSubscriptionResultAlreadyEnded
	// CancelSubscriptionResultNotFoundはStripe上にサブスクリプションが
	// 存在しなかったことを表す。
	CancelSubscriptionResultNotFound
)

// CancelSubscriptionはサブスクリプションを即時解約する。残りの期間の日割り
// 返金と、未請求分の即時請求は行わない。
// 終了済みのサブスクリプションに解約を送るとStripeはエラーを返すため、
// 先に取得して状態を確かめる。
func (a *Adapter) CancelSubscription(ctx context.Context, subscriptionID string) (CancelSubscriptionResult, error) {
	sub, err := a.client.V1Subscriptions.Retrieve(ctx, subscriptionID, nil)
	if err != nil {
		if isResourceMissing(err) {
			return CancelSubscriptionResultNotFound, nil
		}
		return 0, err
	}
	if isSubscriptionEnded(sub.Status) {
		return CancelSubscriptionResultAlreadyEnded, nil
	}

	params := &stripego.SubscriptionCancelParams{
		InvoiceNow: stripego.Bool(false),
		Prorate:    stripego.Bool(false),
	}
	if _, err := a.client.V1Subscriptions.Cancel(ctx, subscriptionID, params); err != nil {
		// 取得から解約までの間に削除された場合も、存在しないものとして扱う。
		if isResourceMissing(err) {
			return CancelSubscriptionResultNotFound, nil
		}
		return 0, err
	}
	return CancelSubscriptionResultCanceled, nil
}

// isSubscriptionEndedはサブスクリプションがこれ以上課金されない終了状態かを返す。
func isSubscriptionEnded(status stripego.SubscriptionStatus) bool {
	return status == stripego.SubscriptionStatusCanceled ||
		status == stripego.SubscriptionStatusIncompleteExpired
}

// isResourceMissingはStripe APIのエラーが対象のリソースが存在しないことを
// 示すかを返す。
func isResourceMissing(err error) bool {
	var stripeErr *stripego.Error
	return errors.As(err, &stripeErr) && stripeErr.Code == stripego.ErrorCodeResourceMissing
}
