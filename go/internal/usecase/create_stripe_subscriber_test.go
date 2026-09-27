package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	annictstripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/testutil"
)

func TestParseUserIDFromMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		metadata  map[string]string
		wantID    model.UserID
		wantError bool
		errorType string
	}{
		{
			name:      "user_idが存在しない場合はエラー",
			metadata:  map[string]string{},
			wantID:    0,
			wantError: true,
			errorType: "MetadataUserIDMissingError",
		},
		{
			name:      "user_idが空の場合はエラー",
			metadata:  map[string]string{"user_id": ""},
			wantID:    0,
			wantError: true,
			errorType: "MetadataUserIDInvalidError",
		},
		{
			name:      "user_idが数値でない場合はエラー",
			metadata:  map[string]string{"user_id": "abc"},
			wantID:    0,
			wantError: true,
			errorType: "MetadataUserIDInvalidError",
		},
		{
			name:      "user_idが正しい場合は成功",
			metadata:  map[string]string{"user_id": "12345"},
			wantID:    12345,
			wantError: false,
		},
		{
			name:      "user_idが0でも成功",
			metadata:  map[string]string{"user_id": "0"},
			wantID:    0,
			wantError: false,
		},
		{
			name:      "user_idが負の値でも成功",
			metadata:  map[string]string{"user_id": "-1"},
			wantID:    -1,
			wantError: false,
		},
		{
			name:      "他のキーが含まれていても成功",
			metadata:  map[string]string{"user_id": "999", "plan": "monthly"},
			wantID:    999,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userID, err := ParseUserIDFromMetadata(tt.metadata)

			if tt.wantError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、nilが返されました")
					return
				}

				switch tt.errorType {
				case "MetadataUserIDMissingError":
					if !IsMetadataUserIDMissingError(err) {
						t.Errorf("MetadataUserIDMissingErrorが期待されましたが、%vが返されました", err)
					}
				case "MetadataUserIDInvalidError":
					if !IsMetadataUserIDInvalidError(err) {
						t.Errorf("MetadataUserIDInvalidErrorが期待されましたが、%vが返されました", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("予期しないエラー: %v", err)
				return
			}

			if userID != tt.wantID {
				t.Errorf("userID = %d、期待値 = %d", userID, tt.wantID)
			}
		})
	}
}

func TestInvalidSubscriptionStatusError(t *testing.T) {
	t.Parallel()

	err := &InvalidSubscriptionStatusError{Status: "unknown"}

	// Error()メソッドのテスト
	expected := "invalid subscription status: unknown"
	if err.Error() != expected {
		t.Errorf("Error() = %q、期待値 = %q", err.Error(), expected)
	}

	// IsInvalidSubscriptionStatusErrorのテスト
	if !IsInvalidSubscriptionStatusError(err) {
		t.Error("IsInvalidSubscriptionStatusError = false、期待値 = true")
	}
}

func TestMetadataUserIDMissingError(t *testing.T) {
	t.Parallel()

	err := &MetadataUserIDMissingError{}

	// Error()メソッドのテスト
	expected := "user_id is missing from metadata"
	if err.Error() != expected {
		t.Errorf("Error() = %q、期待値 = %q", err.Error(), expected)
	}

	// IsMetadataUserIDMissingErrorのテスト
	if !IsMetadataUserIDMissingError(err) {
		t.Error("IsMetadataUserIDMissingError = false、期待値 = true")
	}
}

func TestMetadataUserIDInvalidError(t *testing.T) {
	t.Parallel()

	err := &MetadataUserIDInvalidError{Value: "abc"}

	// Error()メソッドのテスト
	expected := "invalid user_id in metadata: abc"
	if err.Error() != expected {
		t.Errorf("Error() = %q、期待値 = %q", err.Error(), expected)
	}

	// IsMetadataUserIDInvalidErrorのテスト
	if !IsMetadataUserIDInvalidError(err) {
		t.Error("IsMetadataUserIDInvalidError = false、期待値 = true")
	}
}

// insertStripeTestUserはStripeサブスクライバーUseCaseテスト用に、
// コミット済みで未紐付けのユーザーを1件作成してIDを返す。
// CreateStripeSubscriberUsecaseは内部で自前のトランザクションを開くため、
// テストデータは外側のトランザクションに閉じ込めずDBへコミットしておく必要が
// ある (テストガイドのGetTestDBの解説を参照)。
func insertStripeTestUser(t *testing.T, db *sql.DB) model.UserID {
	t.Helper()
	return insertStripeTestUserLinkedTo(t, db, sql.NullInt64{})
}

// insertStripeTestUserLinkedToはstripe_subscriber_idをsubscriberIDに設定した
// コミット済みユーザーを作成してIDを返す (未紐付けにするにはゼロ値のNullInt64を渡す)。
// insertStripeTestUserの背後にある共通のINSERTであり、仕込んだsubscriberに紐付け済みの
// ユーザーを必要とする (紐付け解除を検証する) delete webhookテストから直接使う。
func insertStripeTestUserLinkedTo(t *testing.T, db *sql.DB, subscriberID sql.NullInt64) model.UserID {
	t.Helper()

	suffix := randomString(8)
	var userID int64
	err := db.QueryRow(`
		INSERT INTO users (
			username, email, role, locale,
			created_at, updated_at,
			encrypted_password, sign_in_count,
			time_zone, allowed_locales, stripe_subscriber_id
		) VALUES (
			$1, $2, 0, 'ja',
			NOW(), NOW(),
			'encrypted_test_password', 0,
			'Asia/Tokyo', ARRAY[]::varchar[], $3
		) RETURNING id
	`, "tcs"+suffix, "test_create_sub_"+suffix+"@example.com", subscriberID).Scan(&userID)
	if err != nil {
		t.Fatalf("ユーザーの作成に失敗: %v", err)
	}
	return model.UserID(userID)
}

// validSubscriptionは単一のアクティブなアイテムを持つドメイン形の
// Subscriptionを返す。fake SubscriptionRetrieverの正常系の戻り値として使う。
func validSubscription() *annictstripe.Subscription {
	now := time.Now()
	return &annictstripe.Subscription{
		Status: string(model.StripeSubscriptionStatusActive),
		Items: []annictstripe.SubscriptionItem{
			{
				PriceID:            "price_monthly",
				CurrentPeriodStart: now,
				CurrentPeriodEnd:   now.AddDate(0, 1, 0),
			},
		},
	}
}

// timesCloseはgotとwantが1秒以内に収まるかを返す。
// PostgreSQLを往復したタイムスタンプはマイクロ秒未満の精度を失う (丸めも入りうる)
// ため、永続化後の時刻は厳密一致ではなく許容差付きで比較する。
func timesClose(got, want time.Time) bool {
	const tolerance = time.Second
	diff := got.Sub(want)
	return diff >= -tolerance && diff <= tolerance
}

func TestCreateStripeSubscriberUsecase_Execute(t *testing.T) {
	t.Parallel()

	// UseCaseは自前でトランザクションを開くため、外側のロールバック用
	// トランザクションではなく共有DBを直接使い、テストデータをコミットする
	// (GetTestDB)。
	db := testutil.GetTestDB()
	queries := query.New(db)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	t.Run("正常系: サブスクリプション取得・StripeSubscriber作成・ユーザー紐付けが行われる", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		userID := insertStripeTestUser(t, db)
		customerID := "cus_create_" + randomString(8)
		subscriptionID := "sub_create_" + randomString(8)

		sub := validSubscription()
		uc := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, &fakeSubscriptionRetriever{
			subscription: sub,
		})

		result, err := uc.Execute(ctx, CreateStripeSubscriberInput{
			StripeCustomerID:     customerID,
			StripeSubscriptionID: subscriptionID,
			UserID:               userID,
		})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}

		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID))
			_, _ = db.Exec("DELETE FROM stripe_subscribers WHERE id = $1", int64(result.StripeSubscriber.ID))
		})

		if result.StripeSubscriber.StripeCustomerID != customerID {
			t.Errorf("StripeCustomerID = %s、期待値 = %s", result.StripeSubscriber.StripeCustomerID, customerID)
		}
		if result.StripeSubscriber.StripeSubscriptionID != subscriptionID {
			t.Errorf("StripeSubscriptionID = %s、期待値 = %s", result.StripeSubscriber.StripeSubscriptionID, subscriptionID)
		}
		if result.StripeSubscriber.StripePriceID != "price_monthly" {
			t.Errorf("StripePriceID = %s、期待値 = %s", result.StripeSubscriber.StripePriceID, "price_monthly")
		}
		if result.StripeSubscriber.StripeStatus != string(model.StripeSubscriptionStatusActive) {
			t.Errorf("StripeStatus = %s、期待値 = %s", result.StripeSubscriber.StripeStatus, model.StripeSubscriptionStatusActive)
		}

		// サブスクリプションアイテムの請求期間が永続化レコードへマッピングされる
		// ことを確認する。値はDBを往復する (マイクロ秒精度・丸めが入りうる) ため、
		// 許容差を設けて比較する。
		wantPeriodStart := sub.Items[0].CurrentPeriodStart
		wantPeriodEnd := sub.Items[0].CurrentPeriodEnd
		if !timesClose(result.StripeSubscriber.StripeCurrentPeriodStart, wantPeriodStart) {
			t.Errorf("StripeCurrentPeriodStart = %v、期待値 = %v前後", result.StripeSubscriber.StripeCurrentPeriodStart, wantPeriodStart)
		}
		if !timesClose(result.StripeSubscriber.StripeCurrentPeriodEnd, wantPeriodEnd) {
			t.Errorf("StripeCurrentPeriodEnd = %v、期待値 = %v前後", result.StripeSubscriber.StripeCurrentPeriodEnd, wantPeriodEnd)
		}

		persisted, err := stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
		if err != nil {
			t.Fatalf("StripeSubscriber取得エラー: %v", err)
		}
		if persisted == nil {
			t.Fatal("StripeSubscriberが永続化されていません")
		}

		linked, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("ユーザー取得エラー: %v", err)
		}
		if !linked.StripeSubscriberID.Valid {
			t.Fatal("ユーザーにStripeSubscriberが紐付けられていません")
		}
		if linked.StripeSubscriberID.Int64 != int64(result.StripeSubscriber.ID) {
			t.Errorf("紐付けられたStripeSubscriberID = %d、期待値 = %d", linked.StripeSubscriberID.Int64, int64(result.StripeSubscriber.ID))
		}
	})

	t.Run("異常系: 無効なサブスクリプションステータスはエラー", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		userID := insertStripeTestUser(t, db)
		t.Cleanup(func() { _, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID)) })

		subscriptionID := "sub_invalid_status_" + randomString(8)
		uc := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, &fakeSubscriptionRetriever{
			subscription: &annictstripe.Subscription{
				Status: "invalid_status",
				Items: []annictstripe.SubscriptionItem{
					{PriceID: "price_monthly", CurrentPeriodStart: time.Now(), CurrentPeriodEnd: time.Now().AddDate(0, 1, 0)},
				},
			},
		})

		_, err := uc.Execute(ctx, CreateStripeSubscriberInput{
			StripeCustomerID:     "cus_invalid_status_" + randomString(8),
			StripeSubscriptionID: subscriptionID,
			UserID:               userID,
		})
		if !IsInvalidSubscriptionStatusError(err) {
			t.Fatalf("InvalidSubscriptionStatusErrorが期待されましたが、別のエラーが返されました: %v", err)
		}

		// ステータス検証はトランザクション開始前に行われるため、何も作成されない。
		got, err := stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
		if err != nil {
			t.Fatalf("StripeSubscriber取得エラー: %v", err)
		}
		if got != nil {
			t.Error("無効ステータスなのにStripeSubscriberが作成されている")
		}
	})

	t.Run("異常系: アイテムが空のサブスクリプションはエラー", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		userID := insertStripeTestUser(t, db)
		t.Cleanup(func() { _, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID)) })

		subscriptionID := "sub_empty_items_" + randomString(8)
		uc := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, &fakeSubscriptionRetriever{
			subscription: &annictstripe.Subscription{
				Status: string(model.StripeSubscriptionStatusActive),
				Items:  nil,
			},
		})

		_, err := uc.Execute(ctx, CreateStripeSubscriberInput{
			StripeCustomerID:     "cus_empty_items_" + randomString(8),
			StripeSubscriptionID: subscriptionID,
			UserID:               userID,
		})
		if err == nil {
			t.Fatal("エラーが期待されましたが、nilが返されました")
		}

		got, err := stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
		if err != nil {
			t.Fatalf("StripeSubscriber取得エラー: %v", err)
		}
		if got != nil {
			t.Error("アイテムが空なのにStripeSubscriberが作成されている")
		}
	})

	t.Run("異常系: Stripe APIエラーは握り潰さず伝播する", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		userID := insertStripeTestUser(t, db)
		t.Cleanup(func() { _, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID)) })

		// UseCaseがfmt.Errorfでラップしてもerrors.Isで検出できるよう、
		// Stripe APIの失敗を握り潰さず伝播することをsentinel errorで検証する。
		errStripeAPI := errors.New("stripe api unavailable")
		uc := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, &fakeSubscriptionRetriever{
			err: errStripeAPI,
		})

		_, err := uc.Execute(ctx, CreateStripeSubscriberInput{
			StripeCustomerID:     "cus_api_err_" + randomString(8),
			StripeSubscriptionID: "sub_api_err_" + randomString(8),
			UserID:               userID,
		})
		if !errors.Is(err, errStripeAPI) {
			t.Fatalf("Stripe APIエラーが伝播していません: %v", err)
		}
	})

	t.Run("異常系: 永続化に失敗するとトランザクションがロールバックされユーザー紐付けが残らない", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		userID := insertStripeTestUser(t, db)

		// subscriptionIDを既に持つsubscriberをコミット済みで用意し、UseCaseの
		// Createがstripe_subscription_idの一意インデックスに当たってBeginTx後に
		// トランザクションが中断するようにする。紐付けのみを失敗させることはできない:
		// UpdateUserStripeSubscriberIDは :execで0行はエラーにならず、紐付けのUPDATE
		// は作成直後のsubscriberへのFKを常に満たすため、重複キーによるCreate失敗が
		// トランザクション内で決定的に失敗させて部分的な状態が残らないことを検証する唯一の
		// 手段となる。
		subscriptionID := "sub_rollback_" + randomString(8)
		seeded, err := stripeSubscriberRepo.Create(ctx, query.CreateStripeSubscriberParams{
			StripeCustomerID:         "cus_seed_" + randomString(8),
			StripeSubscriptionID:     subscriptionID,
			StripePriceID:            "price_monthly",
			StripeStatus:             "active",
			StripeCurrentPeriodStart: time.Now(),
			StripeCurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		})
		if err != nil {
			t.Fatalf("シード用StripeSubscriberの作成に失敗: %v", err)
		}
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID))
			_, _ = db.Exec("DELETE FROM stripe_subscribers WHERE id = $1", int64(seeded.ID))
		})

		uc := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, &fakeSubscriptionRetriever{
			subscription: validSubscription(),
		})

		_, err = uc.Execute(ctx, CreateStripeSubscriberInput{
			StripeCustomerID:     "cus_rollback_" + randomString(8),
			StripeSubscriptionID: subscriptionID,
			UserID:               userID,
		})
		if err == nil {
			t.Fatal("エラーが期待されましたが、nilが返されました")
		}

		linked, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("ユーザー取得エラー: %v", err)
		}
		if linked.StripeSubscriberID.Valid {
			t.Errorf("ロールバックされず紐付けが残っている: stripe_subscriber_id=%d", linked.StripeSubscriberID.Int64)
		}
	})
}
