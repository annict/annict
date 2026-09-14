package usecase

import (
	"context"

	annictstripe "github.com/annict/annict/go/internal/stripe"
)

// このファイルは呼び出し側のStripe interface (SubscriptionRetriever /
// CheckoutSessionCreator / PortalSessionCreator) のテストダブルを提供し、実際の
// Stripe APIを呼ばずにStripe周りのUseCaseをテストできるようにする。これらの
// fakeを使う振る舞いテストは後続フェーズで追加する。ここではseamの提供と、各
// fakeがinterfaceを満たすことのコンパイル時チェックのみを行う。

// fakeSubscriptionRetrieverはSubscriptionRetrieverのテストダブル。
type fakeSubscriptionRetriever struct {
	subscription *annictstripe.Subscription
	err          error
}

func (f *fakeSubscriptionRetriever) RetrieveSubscription(ctx context.Context, subscriptionID string) (*annictstripe.Subscription, error) {
	return f.subscription, f.err
}

// fakeCheckoutSessionCreatorはCheckoutSessionCreatorのテストダブル。
// 呼び出しの有無と受け取ったparamsを記録し、metadata / localeの受け渡しや、
// バリデーション・価格エラーがStripe呼び出し前に短絡することをテストで検証できる。
type fakeCheckoutSessionCreator struct {
	url string
	err error

	called    bool
	gotParams annictstripe.CheckoutSessionParams
}

func (f *fakeCheckoutSessionCreator) CreateCheckoutSession(ctx context.Context, params annictstripe.CheckoutSessionParams) (string, error) {
	f.called = true
	f.gotParams = params
	return f.url, f.err
}

// fakePortalSessionCreatorはPortalSessionCreatorのテストダブル。
// 呼び出しの有無と受け取ったparamsを記録し、customer / return URL / localeの
// 受け渡しや、非サポーター・非アクティブ時にStripe呼び出し前に短絡することを
// テストで検証できる。
type fakePortalSessionCreator struct {
	url string
	err error

	called    bool
	gotParams annictstripe.PortalSessionParams
}

func (f *fakePortalSessionCreator) CreatePortalSession(ctx context.Context, params annictstripe.PortalSessionParams) (string, error) {
	f.called = true
	f.gotParams = params
	return f.url, f.err
}

// 各fakeが呼び出し側interfaceを満たすことのコンパイル時チェック。
var (
	_ SubscriptionRetriever  = (*fakeSubscriptionRetriever)(nil)
	_ CheckoutSessionCreator = (*fakeCheckoutSessionCreator)(nil)
	_ PortalSessionCreator   = (*fakePortalSessionCreator)(nil)
)
