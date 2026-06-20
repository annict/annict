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
	"github.com/annict/annict/go/internal/testutil"
)

// seedSubscriberUser seeds a StripeSubscriber with the given status and returns a
// user linked to it, the tx-bound repository that can read it back, and the
// seeded customer ID. CreatePortalSessionUsecase.Execute opens no transaction of
// its own (it only reads via GetByID), so the subscriber held in SetupTx's
// auto-rollback transaction is visible to the UseCase, the same way the 2-2
// Conflict case seeds a subscriber.
//
// [Ja] seedSubscriberUser は指定したステータスの StripeSubscriber を仕込み、それに
// 紐付くユーザー・読み戻せる tx バインドの Repository・仕込んだ顧客 ID を返す。
// CreatePortalSessionUsecase.Execute は自前のトランザクションを開かず (GetByID で
// 読むだけ) のため、SetupTx の自動ロールバック用トランザクションに仕込んだ
// subscriber を UseCase から参照できる。2-2 の Conflict ケースと同じ仕込み方。
func seedSubscriberUser(ctx context.Context, t *testing.T, status model.StripeSubscriptionStatus) (*model.User, *repository.StripeSubscriberRepository, string) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewStripeSubscriberRepository(query.New(db).WithTx(tx))

	customerID := "cus_portal_" + randomString(8)
	seeded, err := repo.Create(ctx, query.CreateStripeSubscriberParams{
		StripeCustomerID:         customerID,
		StripeSubscriptionID:     "sub_portal_" + randomString(8),
		StripePriceID:            "price_monthly_test",
		StripeStatus:             string(status),
		StripeCurrentPeriodStart: time.Now(),
		StripeCurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
	})
	if err != nil {
		t.Fatalf("シード用 StripeSubscriber の作成に失敗: %v", err)
	}

	seededID := seeded.ID
	user := &model.User{ID: model.UserID(1), StripeSubscriberID: &seededID}
	return user, repo, customerID
}

func TestCreatePortalSessionUsecase_Execute(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}

	t.Run("正常系: PortalURL を返し customer / return URL / locale が設定される", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Locale is "ja" only when the input is "ja"; anything else falls back to "en".
		//
		// [Ja] locale は input が "ja" のときだけ "ja"。それ以外は "en" にフォールバックする。
		tests := []struct {
			name        string
			inputLocale string
			wantLocale  string
		}{
			{name: "locale ja", inputLocale: "ja", wantLocale: "ja"},
			{name: "locale en", inputLocale: "en", wantLocale: "en"},
			{name: "locale が ja 以外なら en にフォールバック", inputLocale: "fr", wantLocale: "en"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				user, repo, customerID := seedSubscriberUser(ctx, t, model.StripeSubscriptionStatusActive)

				creator := &fakePortalSessionCreator{url: "https://billing.stripe.com/p/session_test"}
				uc := NewCreatePortalSessionUsecase(cfg, repo, creator)

				output, err := uc.Execute(ctx, CreatePortalSessionInput{
					User:   user,
					Locale: tt.inputLocale,
				})
				if err != nil {
					t.Fatalf("予期しないエラー: %v", err)
				}
				if output.PortalURL != creator.url {
					t.Errorf("PortalURL: got %q, want %q", output.PortalURL, creator.url)
				}
				if !creator.called {
					t.Fatal("CreatePortalSession が呼ばれていません")
				}
				if creator.gotParams.CustomerID != customerID {
					t.Errorf("CustomerID: got %q, want %q", creator.gotParams.CustomerID, customerID)
				}
				if creator.gotParams.Locale != tt.wantLocale {
					t.Errorf("Locale: got %q, want %q", creator.gotParams.Locale, tt.wantLocale)
				}

				// Fix the expected return URL as a literal (not recomputed from
				// cfg.AppURL()) so the assertion also covers the base URL composition
				// (scheme + domain), not just the "/supporters" suffix. cfg.Domain is
				// "test.annict.com" and cfg.AppURL() returns "https://" + Domain.
				//
				// [Ja] 期待値は cfg.AppURL() から再計算せずリテラルで固定し、"/supporters"
				// サフィックスだけでなくベース URL の組み立て (スキーム + ドメイン) まで
				// 検証する。cfg.Domain は "test.annict.com" で、cfg.AppURL() は
				// "https://" + Domain を返す。
				const wantReturnURL = "https://test.annict.com/supporters"
				if creator.gotParams.ReturnURL != wantReturnURL {
					t.Errorf("ReturnURL: got %q, want %q", creator.gotParams.ReturnURL, wantReturnURL)
				}
			})
		}
	})

	t.Run("異常系: 非サポーター (StripeSubscriberID が nil) は NotStripeSubscriberError を返し Stripe を呼ばない", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// A user with no linked subscriber short-circuits before the repository is
		// consulted, so the repo can be nil for this case.
		//
		// [Ja] サブスクライバー未紐付けのユーザーは Repository を参照する前に短絡する
		// ため、このケースでは repo は nil でよい。
		creator := &fakePortalSessionCreator{url: "https://billing.stripe.com/p/session_test"}
		uc := NewCreatePortalSessionUsecase(cfg, nil, creator)

		_, err := uc.Execute(ctx, CreatePortalSessionInput{
			User:   &model.User{ID: model.UserID(1)},
			Locale: "ja",
		})
		if !IsNotStripeSubscriberError(err) {
			t.Fatalf("NotStripeSubscriberError が期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("非サポーター時に Stripe を呼んではいけません")
		}
	})

	t.Run("異常系: 参照先サブスクライバーが存在しない (GetByID が nil) は NotStripeSubscriberError を返し Stripe を呼ばない", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// The user references a subscriber ID that no longer exists, so GetByID
		// returns (nil, nil). The UseCase must treat this as a non-supporter and
		// short-circuit before calling Stripe, rather than dereferencing the nil
		// subscriber. A live (non-rolled-back) tx with nothing seeded is used so
		// GetByID actually queries and finds no row; this exercises the (nil, nil)
		// not-found branch, distinct from the StripeSubscriberID == nil case above.
		//
		// [Ja] ユーザーが既に存在しないサブスクライバー ID を参照しており、GetByID は
		// (nil, nil) を返す。UseCase は nil の subscriber を参照外しせず、これを非
		// サポーターとして扱って Stripe 呼び出し前に短絡するべき。GetByID が実際に
		// クエリして 0 件と分かるよう、何も仕込まないロールバック前の有効な tx を使う。
		// これは上の StripeSubscriberID == nil ケースとは別経路の (nil, nil) not-found
		// 分岐を検証する。
		db, tx := testutil.SetupTx(t)
		repo := repository.NewStripeSubscriberRepository(query.New(db).WithTx(tx))

		missingID := model.StripeSubscriberID(99999)
		user := &model.User{ID: model.UserID(1), StripeSubscriberID: &missingID}

		creator := &fakePortalSessionCreator{url: "https://billing.stripe.com/p/session_test"}
		uc := NewCreatePortalSessionUsecase(cfg, repo, creator)

		_, err := uc.Execute(ctx, CreatePortalSessionInput{
			User:   user,
			Locale: "ja",
		})
		if !IsNotStripeSubscriberError(err) {
			t.Fatalf("NotStripeSubscriberError が期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("参照先サブスクライバーが存在しないとき Stripe を呼んではいけません")
		}
	})

	t.Run("異常系: 非アクティブなサブスクリプションは NotStripeSubscriberError を返し Stripe を呼ばない", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		user, repo, _ := seedSubscriberUser(ctx, t, model.StripeSubscriptionStatusCanceled)

		creator := &fakePortalSessionCreator{url: "https://billing.stripe.com/p/session_test"}
		uc := NewCreatePortalSessionUsecase(cfg, repo, creator)

		_, err := uc.Execute(ctx, CreatePortalSessionInput{
			User:   user,
			Locale: "ja",
		})
		if !IsNotStripeSubscriberError(err) {
			t.Fatalf("NotStripeSubscriberError が期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("非アクティブ時に Stripe を呼んではいけません")
		}
	})

	t.Run("異常系: Stripe API エラーは握り潰さず伝播する", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		user, repo, _ := seedSubscriberUser(ctx, t, model.StripeSubscriptionStatusActive)

		// Sentinel error so we can assert propagation via errors.Is even though the
		// UseCase wraps it with fmt.Errorf.
		//
		// [Ja] UseCase が fmt.Errorf でラップしても errors.Is で伝播を検証できるよう
		// sentinel error を使う。
		errStripeAPI := errors.New("stripe api unavailable")
		creator := &fakePortalSessionCreator{err: errStripeAPI}
		uc := NewCreatePortalSessionUsecase(cfg, repo, creator)

		_, err := uc.Execute(ctx, CreatePortalSessionInput{
			User:   user,
			Locale: "ja",
		})
		if !errors.Is(err, errStripeAPI) {
			t.Fatalf("Stripe API エラーが伝播していません: %v", err)
		}
		if !creator.called {
			t.Error("Stripe API エラーの検証では CreatePortalSession が呼ばれているべきです")
		}
	})
}

// TestCreatePortalSessionUsecase_Execute_PropagatesNonNotFoundError verifies that
// a non-not-found retrieval error from GetByID is wrapped and propagated, rather
// than being swallowed into a NotStripeSubscriberError or a misleading "Stripe
// client not configured" error. It mirrors the 2-2 checkout regression test
// (TestCreateCheckoutSessionUsecase_Execute_PropagatesNonNotFoundError) so the
// portal path keeps the same retrieval-error coverage.
//
// [Ja] GetByID が返す not-found 以外の取得エラーが、NotStripeSubscriberError や
// 「Stripe クライアント未設定」という誤解を招くエラーに化けず、ラップされてそのまま
// 伝播することをテストする。portal 経路でも取得エラーのカバレッジを 2-2 の checkout
// 回帰テスト (TestCreateCheckoutSessionUsecase_Execute_PropagatesNonNotFoundError)
// と揃えるため、同じパターンに沿わせている。
func TestCreatePortalSessionUsecase_Execute_PropagatesNonNotFoundError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutil.GetTestDB()

	// Begin and immediately roll back a transaction so that GetByID fails with
	// sql.ErrTxDone (a non-not-found error) instead of returning sql.ErrNoRows.
	// This reproduces a genuine retrieval error without depending on connection
	// state.
	//
	// [Ja] トランザクションを開始して即座にロールバックし、GetByID が sql.ErrNoRows
	// ではなく sql.ErrTxDone (not-found 以外のエラー) で失敗するようにする。接続状態に
	// 依存せず、本物の取得エラーを再現できる。
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("トランザクションの開始に失敗: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("トランザクションのロールバックに失敗: %v", err)
	}

	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	repo := repository.NewStripeSubscriberRepository(query.New(db).WithTx(tx))

	// nil portal creator: the retrieval error must surface before the Stripe-client
	// check, so the propagated error is the retrieval error rather than the "Stripe
	// client not configured" error.
	//
	// [Ja] portal creator は nil。取得エラーは Stripe クライアントのチェックより前に
	// 表面化するはずで、伝播するエラーは「Stripe クライアント未設定」ではなく取得
	// エラーになる。
	uc := NewCreatePortalSessionUsecase(cfg, repo, nil)

	subscriberID := model.StripeSubscriberID(99999)
	user := &model.User{ID: model.UserID(1), StripeSubscriberID: &subscriberID}

	_, err = uc.Execute(ctx, CreatePortalSessionInput{
		User:   user,
		Locale: "ja",
	})
	if err == nil {
		t.Fatal("not-found 以外の取得エラー時はエラーが伝播するべきですが、nil が返りました")
	}

	// The propagated error must wrap the underlying retrieval error, not a
	// misleading downstream error (e.g. missing Stripe client).
	//
	// [Ja] 伝播するエラーは下流の誤解を招くエラー (例: Stripe クライアント未設定) では
	// なく、元の取得エラーをラップしているべき。
	if !errors.Is(err, sql.ErrTxDone) {
		t.Errorf("取得エラーがそのまま伝播するべきですが、別のエラーが返りました: %v", err)
	}

	// It must not be misclassified as a non-supporter: a retrieval error is distinct
	// from a (nil, nil) not-found result.
	//
	// [Ja] 非サポーターとして誤分類されてはならない。取得エラーは (nil, nil) の
	// not-found とは区別される。
	if IsNotStripeSubscriberError(err) {
		t.Errorf("取得エラーを非サポーター (NotStripeSubscriberError) として扱うべきではありません: %v", err)
	}
}
