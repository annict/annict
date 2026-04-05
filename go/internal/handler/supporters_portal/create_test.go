package supporters_portal

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

// setupTestHandler はテスト用のハンドラーをセットアップします
func setupTestHandler(t *testing.T, tx *sql.Tx, db *sql.DB) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)

	// テスト用には nil を渡す（テストでは Stripe API を呼び出さない）
	createPortalSessionUC := usecase.NewCreatePortalSessionUsecase(cfg, stripeSubscriberRepo, nil)

	return NewHandler(sessionManager, createPortalSessionUC)
}

// createUserWithStripeSubscriber はStripeサブスクライバーを持つユーザーを作成します
func createUserWithStripeSubscriber(t *testing.T, tx *sql.Tx, stripeStatus string) (int64, int64) {
	t.Helper()

	userID := testutil.NewUserBuilder(t, tx).Build()
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus(stripeStatus).
		Build()

	// ユーザーにStripeサブスクライバーIDを関連付け
	_, err := tx.Exec(`UPDATE users SET stripe_subscriber_id = $1 WHERE id = $2`, subscriberID, userID)
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの関連付けに失敗しました: %v", err)
	}

	return userID, subscriberID
}

// createUserWithGumroadSubscriber はGumroadサブスクライバーを持つユーザーを作成します
func createUserWithGumroadSubscriber(t *testing.T, tx *sql.Tx, ended bool) (int64, int64) {
	t.Helper()

	userID := testutil.NewUserBuilder(t, tx).Build()

	builder := testutil.NewGumroadSubscriberBuilder(t, tx)
	if ended {
		// 過去の日時を設定（終了済み）
		builder = builder.WithGumroadEndedAt(time.Now().AddDate(-1, 0, 0))
	}
	subscriberID := builder.Build()

	// ユーザーにGumroadサブスクライバーIDを関連付け
	_, err := tx.Exec(`UPDATE users SET gumroad_subscriber_id = $1 WHERE id = $2`, subscriberID, userID)
	if err != nil {
		t.Fatalf("Gumroadサブスクライバーの関連付けに失敗しました: %v", err)
	}

	return userID, subscriberID
}

// getUserByID はユーザーIDからユーザー情報を取得します（テスト用）
func getUserByID(t *testing.T, tx *sql.Tx, userID int64) *model.User {
	t.Helper()

	var user model.User
	err := tx.QueryRow(`
		SELECT id, username, email, role, encrypted_password, locale,
			   stripe_subscriber_id, gumroad_subscriber_id,
			   created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
		&user.EncryptedPassword, &user.Locale,
		&user.StripeSubscriberID, &user.GumroadSubscriberID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗しました: %v", err)
	}

	return &user
}

// TestCreate_NotLoggedIn は未ログインユーザーがアクセスした場合のテスト
func TestCreate_NotLoggedIn(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// ログインページへリダイレクトされることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/sign_in")
	}
}

// TestCreate_NotSupporter は非サポーターユーザーがアクセスした場合のテスト
func TestCreate_NotSupporter(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// サブスクリプションを持たないユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// サポーターページへリダイレクトされることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestCreate_CanceledSubscription はキャンセル済みサブスクリプションの場合のテスト
func TestCreate_CanceledSubscription(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// canceledステータスのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "canceled")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// 非アクティブなのでサポーターページへリダイレクト
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestCreate_UnpaidSubscription はunpaidステータスの場合のテスト
func TestCreate_UnpaidSubscription(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// unpaidステータスのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "unpaid")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// 非アクティブなのでサポーターページへリダイレクト
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestCreate_ActiveSubscription_StripeAPIError はアクティブなサブスクリプションでStripe API呼び出しがエラーの場合のテスト
// 注: テスト環境ではStripe APIが設定されていないため、エラーが発生してリダイレクトされる
func TestCreate_ActiveSubscription_StripeAPIError(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// アクティブなStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "active")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// Stripe APIがテスト環境で設定されていないためエラーになり、リダイレクトされる
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestCreate_PastDueSubscription_StripeAPIError は支払い遅延中のサブスクリプションの場合のテスト
// past_dueはアクティブとして扱われるため、Stripe APIを呼び出す
func TestCreate_PastDueSubscription_StripeAPIError(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// past_dueステータスのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "past_due")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// past_dueはアクティブなのでStripe APIを呼び出すが、テスト環境ではエラーになる
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestCreate_GumroadSubscriber はGumroadサポーターがアクセスした場合のテスト
func TestCreate_GumroadSubscriber(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// アクティブなGumroadサポーターを作成（Stripeサブスクリプションなし）
	userID, _ := createUserWithGumroadSubscriber(t, tx, false)
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// Stripeサポーターではないのでサポーターページへリダイレクト
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}
