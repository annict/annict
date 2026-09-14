package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestGetSupporterStatusUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("正常系: サブスクリプションなしのユーザー", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
		gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)
		uc := NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)

		userID := testutil.NewUserBuilder(t, tx).Build()
		user := getUserByIDForTest(t, tx, userID)

		result, err := uc.Execute(context.Background(), GetSupporterStatusInput{User: user})
		if err != nil {
			t.Fatalf("想定外のエラー = %v", err)
		}

		if result.IsStripeActive {
			t.Error("IsStripeActive = true、期待値 = false")
		}
		if result.IsGumroadActive {
			t.Error("IsGumroadActive = true、期待値 = false")
		}
		if result.StripeSubscriber != nil {
			t.Error("StripeSubscriberがnilでなかった")
		}
		if result.GumroadSubscriber != nil {
			t.Error("GumroadSubscriberがnilでなかった")
		}
	})

	t.Run("正常系: アクティブなStripeサブスクリプション", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
		gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)
		uc := NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)

		userID := testutil.NewUserBuilder(t, tx).Build()
		subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
			WithStripeStatus("active").
			Build()
		_, err := tx.Exec(`UPDATE users SET stripe_subscriber_id = $1 WHERE id = $2`, int64(subscriberID), int64(userID))
		if err != nil {
			t.Fatalf("Stripeサブスクライバーの関連付けに失敗: %v", err)
		}

		user := getUserByIDForTest(t, tx, userID)

		result, err := uc.Execute(context.Background(), GetSupporterStatusInput{User: user})
		if err != nil {
			t.Fatalf("想定外のエラー = %v", err)
		}

		if !result.IsStripeActive {
			t.Error("IsStripeActive = false、期待値 = true")
		}
		if result.StripeSubscriber == nil {
			t.Error("StripeSubscriberがnilだった")
		}
	})

	t.Run("正常系: アクティブなGumroadサブスクリプション", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
		gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)
		uc := NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)

		userID := testutil.NewUserBuilder(t, tx).Build()
		subscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()
		_, err := tx.Exec(`UPDATE users SET gumroad_subscriber_id = $1 WHERE id = $2`, int64(subscriberID), int64(userID))
		if err != nil {
			t.Fatalf("Gumroadサブスクライバーの関連付けに失敗: %v", err)
		}

		user := getUserByIDForTest(t, tx, userID)

		result, err := uc.Execute(context.Background(), GetSupporterStatusInput{User: user})
		if err != nil {
			t.Fatalf("想定外のエラー = %v", err)
		}

		if !result.IsGumroadActive {
			t.Error("IsGumroadActive = false、期待値 = true")
		}
		if result.GumroadSubscriber == nil {
			t.Error("GumroadSubscriberがnilだった")
		}
	})

	t.Run("正常系: 両方アクティブ", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
		gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)
		uc := NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)

		userID := testutil.NewUserBuilder(t, tx).Build()
		stripeSubscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
			WithStripeStatus("active").
			Build()
		gumroadSubscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()
		_, err := tx.Exec(`UPDATE users SET stripe_subscriber_id = $1, gumroad_subscriber_id = $2 WHERE id = $3`,
			int64(stripeSubscriberID), int64(gumroadSubscriberID), int64(userID))
		if err != nil {
			t.Fatalf("サブスクライバーの関連付けに失敗: %v", err)
		}

		user := getUserByIDForTest(t, tx, userID)

		result, err := uc.Execute(context.Background(), GetSupporterStatusInput{User: user})
		if err != nil {
			t.Fatalf("想定外のエラー = %v", err)
		}

		if !result.IsStripeActive {
			t.Error("IsStripeActive = false、期待値 = true")
		}
		if !result.IsGumroadActive {
			t.Error("IsGumroadActive = false、期待値 = true")
		}
		if result.StripeSubscriber == nil {
			t.Error("StripeSubscriberがnilだった")
		}
		if result.GumroadSubscriber == nil {
			t.Error("GumroadSubscriberがnilだった")
		}
	})

	t.Run("正常系: キャンセル済みStripeは非アクティブ", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
		gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)
		uc := NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)

		userID := testutil.NewUserBuilder(t, tx).Build()
		subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
			WithStripeStatus("canceled").
			Build()
		_, err := tx.Exec(`UPDATE users SET stripe_subscriber_id = $1 WHERE id = $2`, int64(subscriberID), int64(userID))
		if err != nil {
			t.Fatalf("Stripeサブスクライバーの関連付けに失敗: %v", err)
		}

		user := getUserByIDForTest(t, tx, userID)

		result, err := uc.Execute(context.Background(), GetSupporterStatusInput{User: user})
		if err != nil {
			t.Fatalf("想定外のエラー = %v", err)
		}

		if result.IsStripeActive {
			t.Error("解約済みのサブスクリプションでIsStripeActive = true、期待値 = false")
		}
		if result.StripeSubscriber != nil {
			t.Error("解約済みのサブスクリプションでStripeSubscriberがnilでなかった")
		}
	})

	t.Run("正常系: 終了済みGumroadは非アクティブ", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
		gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)
		uc := NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)

		userID := testutil.NewUserBuilder(t, tx).Build()
		subscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).
			WithGumroadEndedAt(time.Now().AddDate(-1, 0, 0)).
			Build()
		_, err := tx.Exec(`UPDATE users SET gumroad_subscriber_id = $1 WHERE id = $2`, int64(subscriberID), int64(userID))
		if err != nil {
			t.Fatalf("Gumroadサブスクライバーの関連付けに失敗: %v", err)
		}

		user := getUserByIDForTest(t, tx, userID)

		result, err := uc.Execute(context.Background(), GetSupporterStatusInput{User: user})
		if err != nil {
			t.Fatalf("想定外のエラー = %v", err)
		}

		if result.IsGumroadActive {
			t.Error("終了済みのサブスクリプションでIsGumroadActive = true、期待値 = false")
		}
		if result.GumroadSubscriber != nil {
			t.Error("終了済みのサブスクリプションでGumroadSubscriberがnilでなかった")
		}
	})

	t.Run("異常系: nilユーザーはエラー", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
		gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)
		uc := NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)

		_, err := uc.Execute(context.Background(), GetSupporterStatusInput{User: nil})
		if err == nil {
			t.Error("ユーザーがnilのときにエラーを期待したが、nilだった")
		}
	})
}

// getUserByIDForTestはユーザーIDからユーザー情報を取得します (テスト用)
func getUserByIDForTest(t *testing.T, tx *sql.Tx, userID model.UserID) *model.User {
	t.Helper()

	var user model.User
	var stripeSubID, gumroadSubID sql.NullInt64
	err := tx.QueryRow(`
		SELECT id, username, email, role, encrypted_password, locale,
			   stripe_subscriber_id, gumroad_subscriber_id,
			   created_at, updated_at
		FROM users WHERE id = $1
	`, int64(userID)).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
		&user.EncryptedPassword, &user.Locale,
		&stripeSubID, &gumroadSubID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗しました: %v", err)
	}
	if stripeSubID.Valid {
		id := model.StripeSubscriberID(stripeSubID.Int64)
		user.StripeSubscriberID = &id
	}
	if gumroadSubID.Valid {
		id := model.GumroadSubscriberID(gumroadSubID.Int64)
		user.GumroadSubscriberID = &id
	}

	return &user
}
