package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// TestCreateCheckoutSessionUsecase_Execute_PropagatesNonNotFoundError verifies that
// the duplicate-subscription check propagates a non-not-found retrieval error instead
// of swallowing it and continuing the checkout. It is a regression test for the
// previous `if err == nil` guard that ignored every GetByID error.
//
// [Ja] 重複サブスクリプションチェックが、not-found 以外の取得エラーを握り潰して checkout を
// 続行せず、そのまま伝播させることをテストする。全 GetByID エラーを無視していた以前の
// `if err == nil` ガードに対する回帰テスト。
func TestCreateCheckoutSessionUsecase_Execute_PropagatesNonNotFoundError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutil.GetTestDB()

	// Begin and immediately roll back a transaction so that subsequent queries fail
	// with sql.ErrTxDone (a non-not-found error) rather than returning sql.ErrNoRows.
	// This simulates a genuine retrieval error without depending on connection state.
	//
	// [Ja] トランザクションを開始して即座にロールバックし、以降のクエリが sql.ErrNoRows
	// ではなく sql.ErrTxDone (not-found 以外のエラー) で失敗するようにする。接続状態に
	// 依存せず、本物の取得エラーを再現できる。
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("トランザクションの開始に失敗: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("トランザクションのロールバックに失敗: %v", err)
	}

	queries := query.New(db).WithTx(tx)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	v := validator.NewSupportersCheckoutCreateValidator()

	// nil Stripe client: the retrieval error must surface before any Stripe call.
	// [Ja] Stripe クライアントは nil。取得エラーは Stripe 呼び出しより前に表面化するはず。
	uc := NewCreateCheckoutSessionUsecase(cfg, stripeSubscriberRepo, stripeCfg, nil, v)

	subscriberID := model.StripeSubscriberID(99999)
	user := &model.User{
		ID:                 model.UserID(1),
		StripeSubscriberID: &subscriberID,
	}

	_, err = uc.Execute(ctx, CreateCheckoutSessionInput{
		User:   user,
		Plan:   "monthly",
		Locale: "ja",
	})
	if err == nil {
		t.Fatal("not-found 以外の取得エラー時はエラーが伝播するべきですが、nil が返りました")
	}

	// The propagated error must wrap the underlying retrieval error, not a misleading
	// downstream error (e.g. missing Stripe client).
	//
	// [Ja] 伝播するエラーは下流の誤解を招くエラー (例: Stripe クライアント未設定) ではなく、
	// 元の取得エラーをラップしているべき。
	if !errors.Is(err, sql.ErrTxDone) {
		t.Errorf("取得エラーがそのまま伝播するべきですが、別のエラーが返りました: %v", err)
	}

	// It must not be treated as a duplicate (Conflict): not-found / errors are distinct
	// from an active subscription.
	//
	// [Ja] 重複 (Conflict) として扱われてはならない。not-found / エラーはアクティブな
	// サブスクリプションとは区別される。
	if appErr := model.AsAppError(err); appErr != nil && appErr.Code == model.AppErrCodeConflict {
		t.Errorf("取得エラーを重複 (Conflict) として扱うべきではありません: %v", err)
	}
}

func TestCreateCheckoutSessionUsecase_Execute(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	stripeCfgWithPrices := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	v := validator.NewSupportersCheckoutCreateValidator()

	// A user with no linked subscriber skips the duplicate-subscription lookup, so
	// the duplicate-check repository is never consulted and can be nil for every
	// case except the Conflict one (which seeds an active subscriber in the DB).
	//
	// [Ja] サブスクライバー未紐付けのユーザーは重複サブスクリプションの取得をスキップ
	// するため、重複チェック用 Repository は参照されず、DB にアクティブな subscriber を
	// 仕込む Conflict ケース以外では nil でよい。
	unlinkedUser := &model.User{ID: model.UserID(1)}

	t.Run("正常系: monthly / yearly で CheckoutURL を返し metadata と locale が設定される", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Locale is "ja" only when the input is "ja"; anything else falls back to "en".
		//
		// [Ja] locale は input が "ja" のときだけ "ja"。それ以外は "en" にフォールバックする。
		tests := []struct {
			name        string
			plan        string
			inputLocale string
			wantPriceID string
			wantLocale  string
		}{
			{name: "monthly / locale ja", plan: "monthly", inputLocale: "ja", wantPriceID: "price_monthly_test", wantLocale: "ja"},
			{name: "yearly / locale en", plan: "yearly", inputLocale: "en", wantPriceID: "price_yearly_test", wantLocale: "en"},
			{name: "locale が ja 以外なら en にフォールバック", plan: "monthly", inputLocale: "fr", wantPriceID: "price_monthly_test", wantLocale: "en"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				creator := &fakeCheckoutSessionCreator{url: "https://checkout.stripe.com/c/session_test"}
				uc := NewCreateCheckoutSessionUsecase(cfg, nil, stripeCfgWithPrices, creator, v)

				output, err := uc.Execute(ctx, CreateCheckoutSessionInput{
					User:   unlinkedUser,
					Plan:   tt.plan,
					Locale: tt.inputLocale,
				})
				if err != nil {
					t.Fatalf("予期しないエラー: %v", err)
				}
				if output.CheckoutURL != creator.url {
					t.Errorf("CheckoutURL: got %q, want %q", output.CheckoutURL, creator.url)
				}
				if !creator.called {
					t.Fatal("CreateCheckoutSession が呼ばれていません")
				}
				if creator.gotParams.PriceID != tt.wantPriceID {
					t.Errorf("PriceID: got %q, want %q", creator.gotParams.PriceID, tt.wantPriceID)
				}
				if creator.gotParams.UserID != unlinkedUser.ID.String() {
					t.Errorf("metadata user_id: got %q, want %q", creator.gotParams.UserID, unlinkedUser.ID.String())
				}
				if creator.gotParams.Locale != tt.wantLocale {
					t.Errorf("Locale: got %q, want %q", creator.gotParams.Locale, tt.wantLocale)
				}

				// Fix the expected URLs as literals (not recomputed from cfg.AppURL())
				// so the assertion also covers the base URL composition (scheme +
				// domain), not just the path/query suffix. cfg.Domain is "test.annict.com"
				// and cfg.AppURL() returns "https://" + Domain.
				//
				// [Ja] 期待値は cfg.AppURL() から再計算せずリテラルで固定し、パス・クエリの
				// サフィックスだけでなくベース URL の組み立て (スキーム + ドメイン) まで
				// 検証する。cfg.Domain は "test.annict.com" で、cfg.AppURL() は
				// "https://" + Domain を返す。
				const (
					wantSuccessURL = "https://test.annict.com/supporters?success=true"
					wantCancelURL  = "https://test.annict.com/supporters?canceled=true"
				)
				if creator.gotParams.SuccessURL != wantSuccessURL {
					t.Errorf("SuccessURL: got %q, want %q", creator.gotParams.SuccessURL, wantSuccessURL)
				}
				if creator.gotParams.CancelURL != wantCancelURL {
					t.Errorf("CancelURL: got %q, want %q", creator.gotParams.CancelURL, wantCancelURL)
				}
			})
		}
	})

	t.Run("異常系: 無効なプランは ValidationError を返し Stripe を呼ばない", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		creator := &fakeCheckoutSessionCreator{url: "https://checkout.stripe.com/c/session_test"}
		uc := NewCreateCheckoutSessionUsecase(cfg, nil, stripeCfgWithPrices, creator, v)

		_, err := uc.Execute(ctx, CreateCheckoutSessionInput{
			User:   unlinkedUser,
			Plan:   "invalid",
			Locale: "ja",
		})
		if ve := model.AsValidationError(err); ve == nil {
			t.Fatalf("ValidationError が期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("バリデーション失敗時に Stripe を呼んではいけません")
		}
	})

	t.Run("異常系: アクティブなサブスクリプションが既にある場合は Conflict を返し Stripe を呼ばない", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// The duplicate check reads via GetByID, and Execute opens no transaction of
		// its own, so SetupTx (auto-rollback) can seed the active subscriber the
		// repository reads back. This differs from the 2-1 UseCase, which needs
		// GetTestDB because it opens an internal transaction (Tx isolation).
		//
		// [Ja] 重複チェックは GetByID で読み、Execute は自前のトランザクションを開かない
		// ため、SetupTx (自動ロールバック) で仕込んだアクティブな subscriber を Repository が
		// 読み戻せる。内部トランザクションを開くために GetTestDB が必要だった 2-1 の UseCase
		// (Tx 隔離) とは異なる。
		db, tx := testutil.SetupTx(t)
		repo := repository.NewStripeSubscriberRepository(query.New(db).WithTx(tx))

		seeded, err := repo.Create(ctx, query.CreateStripeSubscriberParams{
			StripeCustomerID:         "cus_conflict_" + randomString(8),
			StripeSubscriptionID:     "sub_conflict_" + randomString(8),
			StripePriceID:            "price_monthly_test",
			StripeStatus:             string(model.StripeSubscriptionStatusActive),
			StripeCurrentPeriodStart: time.Now(),
			StripeCurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		})
		if err != nil {
			t.Fatalf("シード用 StripeSubscriber の作成に失敗: %v", err)
		}

		seededID := seeded.ID
		user := &model.User{ID: model.UserID(1), StripeSubscriberID: &seededID}

		creator := &fakeCheckoutSessionCreator{url: "https://checkout.stripe.com/c/session_test"}
		uc := NewCreateCheckoutSessionUsecase(cfg, repo, stripeCfgWithPrices, creator, v)

		_, err = uc.Execute(ctx, CreateCheckoutSessionInput{
			User:   user,
			Plan:   "monthly",
			Locale: "ja",
		})
		appErr := model.AsAppError(err)
		if appErr == nil || appErr.Code != model.AppErrCodeConflict {
			t.Fatalf("Conflict の AppError が期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("重複時に Stripe を呼んではいけません")
		}
	})

	t.Run("異常系: 価格IDが未設定なら Stripe を呼ばずエラーを返す", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Empty price IDs: the plan is valid but no price is configured, so the
		// UseCase must fail with a plain (system) error before reaching Stripe.
		//
		// [Ja] 価格 ID が空: プラン自体は有効だが価格が未設定のため、UseCase は Stripe に
		// 到達する前に素の (システム) エラーで失敗するべき。
		emptyStripeCfg := &annictStripe.Config{}
		creator := &fakeCheckoutSessionCreator{url: "https://checkout.stripe.com/c/session_test"}
		uc := NewCreateCheckoutSessionUsecase(cfg, nil, emptyStripeCfg, creator, v)

		_, err := uc.Execute(ctx, CreateCheckoutSessionInput{
			User:   unlinkedUser,
			Plan:   "monthly",
			Locale: "ja",
		})
		if err == nil {
			t.Fatal("価格ID未設定時はエラーが期待されますが、nil が返りました")
		}
		if ve := model.AsValidationError(err); ve != nil {
			t.Errorf("価格ID未設定は入力エラーではなくシステムエラーであるべきです: %v", err)
		}
		if appErr := model.AsAppError(err); appErr != nil {
			t.Errorf("価格ID未設定は AppError ではなくシステムエラーであるべきです: %v", err)
		}
		if creator.called {
			t.Error("価格ID未設定時に Stripe を呼んではいけません")
		}
	})

	t.Run("異常系: Stripe API エラーは握り潰さず伝播する", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Sentinel error so we can assert propagation via errors.Is even though the
		// UseCase wraps it with fmt.Errorf.
		//
		// [Ja] UseCase が fmt.Errorf でラップしても errors.Is で伝播を検証できるよう
		// sentinel error を使う。
		errStripeAPI := errors.New("stripe api unavailable")
		creator := &fakeCheckoutSessionCreator{err: errStripeAPI}
		uc := NewCreateCheckoutSessionUsecase(cfg, nil, stripeCfgWithPrices, creator, v)

		_, err := uc.Execute(ctx, CreateCheckoutSessionInput{
			User:   unlinkedUser,
			Plan:   "monthly",
			Locale: "ja",
		})
		if !errors.Is(err, errStripeAPI) {
			t.Fatalf("Stripe API エラーが伝播していません: %v", err)
		}
		if !creator.called {
			t.Error("Stripe API エラーの検証では CreateCheckoutSession が呼ばれているべきです")
		}
	})
}
