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

// seedSubscriberUserは指定したステータスのStripeSubscriberを仕込み、それに
// 紐付くユーザー・読み戻せるtxバインドのRepository・仕込んだ顧客IDを返す。
// CreatePortalSessionUsecase.Executeは自前のトランザクションを開かず (GetByIDで
// 読むだけ) のため、SetupTxの自動ロールバック用トランザクションに仕込んだ
// subscriberをUseCaseから参照できる。2-2のConflictケースと同じ仕込み方。
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
		t.Fatalf("シード用StripeSubscriberの作成に失敗: %v", err)
	}

	seededID := seeded.ID
	user := &model.User{ID: model.UserID(1), StripeSubscriberID: &seededID}
	return user, repo, customerID
}

func TestCreatePortalSessionUsecase_Execute(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}

	t.Run("正常系: PortalURLを返しcustomer / return URL / localeが設定される", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// localeはinputが "ja" のときだけ "ja"。それ以外は "en" にフォールバックする。
		tests := []struct {
			name        string
			inputLocale string
			wantLocale  string
		}{
			{name: "localeがjaならjaのまま", inputLocale: "ja", wantLocale: "ja"},
			{name: "localeがenならenのまま", inputLocale: "en", wantLocale: "en"},
			{name: "localeがja以外ならenにフォールバック", inputLocale: "fr", wantLocale: "en"},
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
					t.Errorf("PortalURL = %q、期待値 = %q", output.PortalURL, creator.url)
				}
				if !creator.called {
					t.Fatal("CreatePortalSessionが呼ばれていません")
				}
				if creator.gotParams.CustomerID != customerID {
					t.Errorf("CustomerID = %q、期待値 = %q", creator.gotParams.CustomerID, customerID)
				}
				if creator.gotParams.Locale != tt.wantLocale {
					t.Errorf("Locale = %q、期待値 = %q", creator.gotParams.Locale, tt.wantLocale)
				}

				// 期待値はcfg.AppURL() から再計算せずリテラルで固定し、"/supporters"
				// サフィックスだけでなくベースURLの組み立て (スキーム + ドメイン) まで
				// 検証する。cfg.Domainは "test.annict.com" で、cfg.AppURL() は
				// "https://" + Domainを返す。
				const wantReturnURL = "https://test.annict.com/supporters"
				if creator.gotParams.ReturnURL != wantReturnURL {
					t.Errorf("ReturnURL = %q、期待値 = %q", creator.gotParams.ReturnURL, wantReturnURL)
				}
			})
		}
	})

	t.Run("異常系: 非サポーター (StripeSubscriberIDがnil) はNotStripeSubscriberErrorを返しStripeを呼ばない", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// サブスクライバー未紐付けのユーザーはRepositoryを参照する前に短絡する
		// ため、このケースではrepoはnilでよい。
		creator := &fakePortalSessionCreator{url: "https://billing.stripe.com/p/session_test"}
		uc := NewCreatePortalSessionUsecase(cfg, nil, creator)

		_, err := uc.Execute(ctx, CreatePortalSessionInput{
			User:   &model.User{ID: model.UserID(1)},
			Locale: "ja",
		})
		if !IsNotStripeSubscriberError(err) {
			t.Fatalf("NotStripeSubscriberErrorが期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("非サポーター時にStripeを呼んではいけません")
		}
	})

	t.Run("異常系: 参照先サブスクライバーが存在しない (GetByIDがnil) はNotStripeSubscriberErrorを返しStripeを呼ばない", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// ユーザーが既に存在しないサブスクライバーIDを参照しており、GetByIDは
		// (nil, nil) を返す。UseCaseはnilのsubscriberを参照外しせず、これを非
		// サポーターとして扱ってStripe呼び出し前に短絡するべき。GetByIDが実際に
		// クエリして0件と分かるよう、何も仕込まないロールバック前の有効なtxを使う。
		// これは上のStripeSubscriberID == nilケースとは別経路の (nil, nil) not-found
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
			t.Fatalf("NotStripeSubscriberErrorが期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("参照先サブスクライバーが存在しないときStripeを呼んではいけません")
		}
	})

	t.Run("異常系: 非アクティブなサブスクリプションはNotStripeSubscriberErrorを返しStripeを呼ばない", func(t *testing.T) {
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
			t.Fatalf("NotStripeSubscriberErrorが期待されましたが、別のエラーが返りました: %v", err)
		}
		if creator.called {
			t.Error("非アクティブ時にStripeを呼んではいけません")
		}
	})

	t.Run("異常系: Stripe APIエラーは握り潰さず伝播する", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		user, repo, _ := seedSubscriberUser(ctx, t, model.StripeSubscriptionStatusActive)

		// UseCaseがfmt.Errorfでラップしてもerrors.Isで伝播を検証できるよう
		// sentinel errorを使う。
		errStripeAPI := errors.New("stripe api unavailable")
		creator := &fakePortalSessionCreator{err: errStripeAPI}
		uc := NewCreatePortalSessionUsecase(cfg, repo, creator)

		_, err := uc.Execute(ctx, CreatePortalSessionInput{
			User:   user,
			Locale: "ja",
		})
		if !errors.Is(err, errStripeAPI) {
			t.Fatalf("Stripe APIエラーが伝播していません: %v", err)
		}
		if !creator.called {
			t.Error("Stripe APIエラーの検証ではCreatePortalSessionが呼ばれているべきです")
		}
	})
}

// GetByIDが返すnot-found以外の取得エラーが、NotStripeSubscriberErrorや
// 「Stripeクライアント未設定」という誤解を招くエラーに化けず、ラップされてそのまま
// 伝播することをテストする。portal経路でも取得エラーのカバレッジを2-2のcheckout
// 回帰テスト (TestCreateCheckoutSessionUsecase_Execute_PropagatesNonNotFoundError)
// と揃えるため、同じパターンに沿わせている。
func TestCreatePortalSessionUsecase_Execute_PropagatesNonNotFoundError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutil.GetTestDB()

	// トランザクションを開始して即座にロールバックし、GetByIDがsql.ErrNoRows
	// ではなくsql.ErrTxDone (not-found以外のエラー) で失敗するようにする。接続状態に
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

	// portal creatorはnil。取得エラーはStripeクライアントのチェックより前に
	// 表面化するはずで、伝播するエラーは「Stripeクライアント未設定」ではなく取得
	// エラーになる。
	uc := NewCreatePortalSessionUsecase(cfg, repo, nil)

	subscriberID := model.StripeSubscriberID(99999)
	user := &model.User{ID: model.UserID(1), StripeSubscriberID: &subscriberID}

	_, err = uc.Execute(ctx, CreatePortalSessionInput{
		User:   user,
		Locale: "ja",
	})
	if err == nil {
		t.Fatal("not-found以外の取得エラー時はエラーが伝播するべきですが、nilが返りました")
	}

	// 伝播するエラーは下流の誤解を招くエラー (例: Stripeクライアント未設定) では
	// なく、元の取得エラーをラップしているべき。
	if !errors.Is(err, sql.ErrTxDone) {
		t.Errorf("取得エラーがそのまま伝播するべきですが、別のエラーが返りました: %v", err)
	}

	// 非サポーターとして誤分類されてはならない。取得エラーは (nil, nil) の
	// not-foundとは区別される。
	if IsNotStripeSubscriberError(err) {
		t.Errorf("取得エラーを非サポーター (NotStripeSubscriberError) として扱うべきではありません: %v", err)
	}
}
