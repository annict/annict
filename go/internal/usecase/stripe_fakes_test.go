package usecase

import (
	"context"

	annictstripe "github.com/annict/annict/go/internal/stripe"
)

// This file provides test doubles for the caller-side Stripe interfaces
// (SubscriptionRetriever / CheckoutSessionCreator / PortalSessionCreator) so
// that UseCase behaviour around Stripe can be tested without calling the real
// Stripe API. The behavioural tests that exercise these fakes are added in the
// later phases; here they only provide the seam and a compile-time check that
// each fake satisfies its interface.
//
// [Ja] このファイルは呼び出し側の Stripe interface (SubscriptionRetriever /
// CheckoutSessionCreator / PortalSessionCreator) のテストダブルを提供し、実際の
// Stripe API を呼ばずに Stripe 周りの UseCase をテストできるようにする。これらの
// fake を使う振る舞いテストは後続フェーズで追加する。ここでは seam の提供と、各
// fake が interface を満たすことのコンパイル時チェックのみを行う。

// fakeSubscriptionRetriever is a test double for SubscriptionRetriever.
// [Ja] fakeSubscriptionRetriever は SubscriptionRetriever のテストダブル。
type fakeSubscriptionRetriever struct {
	subscription *annictstripe.Subscription
	err          error
}

func (f *fakeSubscriptionRetriever) RetrieveSubscription(ctx context.Context, subscriptionID string) (*annictstripe.Subscription, error) {
	return f.subscription, f.err
}

// fakeCheckoutSessionCreator is a test double for CheckoutSessionCreator.
// [Ja] fakeCheckoutSessionCreator は CheckoutSessionCreator のテストダブル。
type fakeCheckoutSessionCreator struct {
	url string
	err error
}

func (f *fakeCheckoutSessionCreator) CreateCheckoutSession(ctx context.Context, params annictstripe.CheckoutSessionParams) (string, error) {
	return f.url, f.err
}

// fakePortalSessionCreator is a test double for PortalSessionCreator.
// [Ja] fakePortalSessionCreator は PortalSessionCreator のテストダブル。
type fakePortalSessionCreator struct {
	url string
	err error
}

func (f *fakePortalSessionCreator) CreatePortalSession(ctx context.Context, params annictstripe.PortalSessionParams) (string, error) {
	return f.url, f.err
}

// Compile-time checks that each fake satisfies its caller-side interface.
// [Ja] 各 fake が呼び出し側 interface を満たすことのコンパイル時チェック。
var (
	_ SubscriptionRetriever  = (*fakeSubscriptionRetriever)(nil)
	_ CheckoutSessionCreator = (*fakeCheckoutSessionCreator)(nil)
	_ PortalSessionCreator   = (*fakePortalSessionCreator)(nil)
)
