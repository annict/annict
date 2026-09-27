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

// 重複サブスクリプションチェックが、not-found以外の取得エラーを握り潰してcheckoutを
// 続行せず、そのまま伝播させることをテストする。全GetByIDエラーを無視していた以前の
// `if err == nil` ガードに対する回帰テスト。
func TestCreateCheckoutSessionUsecase_Execute_PropagatesNonNotFoundError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutil.GetTestDB()

	// トランザクションを開始して即座にロールバックし、以降のクエリがsql.ErrNoRows
	// ではなくsql.ErrTxDone (not-found以外のエラー) で失敗するようにする。接続状態に
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

	// Stripeクライアントはnil。取得エラーはStripe呼び出しより前に表面化するはず。
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
		t.Fatal("not-found以外の取得エラー時はエラーが伝播するべきですが、nilが返りました")
	}

	// 伝播するエラーは下流の誤解を招くエラー (例: Stripeクライアント未設定) ではなく、
	// 元の取得エラーをラップしているべき。
	if !errors.Is(err, sql.ErrTxDone) {
		t.Errorf("取得エラーがそのまま伝播するべきですが、別のエラーが返りました: %v", err)
	}

	// 重複 (Conflict) として扱われてはならない。not-found / エラーはアクティブな
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

	// サブスクライバー未紐付けのユーザーは重複サブスクリプションの取得をスキップ
	// するため、重複チェック用Repositoryは参照されず、DBにアクティブなsubscriberを
	// 仕込むConflictケース以外ではnilでよい。
	unlinkedUser := &model.User{ID: model.UserID(1)}

	t.Run("正常系: monthly / yearlyでCheckoutURLを返しmetadataとlocaleが設定される", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// localeはinputが "ja" のときだけ "ja"。それ以外は "en" にフォールバックする。
		tests := []struct {
			name        string
			plan        string
			inputLocale string
			wantPriceID string
			wantLocale  string
		}{
			{name: "monthlyかつlocaleがja", plan: "monthly", inputLocale: "ja", wantPriceID: "price_monthly_test", wantLocale: "ja"},
			{name: "yearlyかつlocaleがen", plan: "yearly", inputLocale: "en", wantPriceID: "price_yearly_test", wantLocale: "en"},
			{name: "localeがja以外ならenにフォールバック", plan: "monthly", inputLocale: "fr", wantPriceID: "price_monthly_test", wantLocale: "en"},
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
					t.Errorf("CheckoutURL = %q、期待値 = %q", output.CheckoutURL, creator.url)
				}
				if !creator.called {
					t.Fatal("CreateCheckoutSessionが呼ばれていません")
				}
				if creator.gotParams.PriceID != tt.wantPriceID {
					t.Errorf("PriceID = %q、期待値 = %q", creator.gotParams.PriceID, tt.wantPriceID)
				}
				if creator.gotParams.UserID != unlinkedUser.ID.String() {
					t.Errorf("metadata user_id = %q、期待値 = %q", creator.gotParams.UserID, unlinkedUser.ID.String())
				}
				if creator.gotParams.Locale != tt.wantLocale {
					t.Errorf("Locale = %q、期待値 = %q", creator.gotParams.Locale, tt.wantLocale)
				}

				// 期待値はcfg.AppURL() から再計算せずリテラルで固定し、パス・クエリの
				// サフィックスだけでなくベースURLの組み立て (スキーム + ドメイン) まで
				// 検証する。cfg.Domainは "test.annict.com" で、cfg.AppURL() は
				// "https://" + Domainを返す。
				const (
					wantSuccessURL = "https://test.annict.com/supporters?success=true"
					wantCancelURL  = "https://test.annict.com/supporters?canceled=true"
				)
				if creator.gotParams.SuccessURL != wantSuccessURL {
					t.Errorf("SuccessURL = %q、期待値 = %q", creator.gotParams.SuccessURL, wantSuccessURL)
				}
				if creator.gotParams.CancelURL != wantCancelURL {
					t.Errorf("CancelURL = %q、期待値 = %q", creator.gotParams.CancelURL, wantCancelURL)
				}
			})
		}
	})

	t.Run("異常系: 無効なプランはValidationErrorを返しStripeを呼ばない", func(t *testing.T) {
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
			t.Fatalf("ValidationErrorが期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("バリデーション失敗時にStripeを呼んではいけません")
		}
	})

	t.Run("異常系: アクティブなサブスクリプションが既にある場合はConflictを返しStripeを呼ばない", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 重複チェックはGetByIDで読み、Executeは自前のトランザクションを開かない
		// ため、SetupTx (自動ロールバック) で仕込んだアクティブなsubscriberをRepositoryが
		// 読み戻せる。内部トランザクションを開くためにGetTestDBが必要だった2-1のUseCase
		// (Tx隔離) とは異なる。
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
			t.Fatalf("シード用StripeSubscriberの作成に失敗: %v", err)
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
			t.Fatalf("ConflictのAppErrorが期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("重複時にStripeを呼んではいけません")
		}
	})

	t.Run("異常系: 価格IDが未設定ならStripeを呼ばずエラーを返す", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 価格IDが空: プラン自体は有効だが価格が未設定のため、UseCaseはStripeに
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
			t.Fatal("価格ID未設定時はエラーが期待されますが、nilが返りました")
		}
		if ve := model.AsValidationError(err); ve != nil {
			t.Errorf("価格ID未設定は入力エラーではなくシステムエラーであるべきです: %v", err)
		}
		if appErr := model.AsAppError(err); appErr != nil {
			t.Errorf("価格ID未設定はAppErrorではなくシステムエラーであるべきです: %v", err)
		}
		if creator.called {
			t.Error("価格ID未設定時にStripeを呼んではいけません")
		}
	})

	t.Run("異常系: Stripe APIエラーは握り潰さず伝播する", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// UseCaseがfmt.Errorfでラップしてもerrors.Isで伝播を検証できるよう
		// sentinel errorを使う。
		errStripeAPI := errors.New("stripe api unavailable")
		creator := &fakeCheckoutSessionCreator{err: errStripeAPI}
		uc := NewCreateCheckoutSessionUsecase(cfg, nil, stripeCfgWithPrices, creator, v)

		_, err := uc.Execute(ctx, CreateCheckoutSessionInput{
			User:   unlinkedUser,
			Plan:   "monthly",
			Locale: "ja",
		})
		if !errors.Is(err, errStripeAPI) {
			t.Fatalf("Stripe APIエラーが伝播していません: %v", err)
		}
		if !creator.called {
			t.Error("Stripe APIエラーの検証ではCreateCheckoutSessionが呼ばれているべきです")
		}
	})
}
