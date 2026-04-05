package supporters

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

// setupTestHandler はテスト用のハンドラーをセットアップします
func setupTestHandler(t *testing.T, tx *sql.Tx, db *sql.DB) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{
		Env: "test",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	imageHelper := image.NewHelper(cfg)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)

	// テスト用のStripe設定（テストではStripe APIを呼び出さないため空でOK）
	stripeCfg := &annictStripe.Config{}

	getSupporterStatusUC := usecase.NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)

	// テスト用には nil を渡す（テストでは Stripe API を呼び出さない）
	return NewHandler(cfg, sessionManager, imageHelper, getSupporterStatusUC, stripeCfg, nil)
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

// TestShow_NotLoggedIn は未ログインユーザーの場合のテスト
func TestShow_NotLoggedIn(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	req := httptest.NewRequest("GET", "/supporters", nil)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	// ステータスコードを確認
	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// ログイン促進セクションが表示されることを確認
	expectedContents := []string{
		"サポーター", // ページタイトル
		"ログイン",  // ログインボタン
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}

	// Content-Typeを確認
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("wrong content-type: got %v want %v", ct, "text/html; charset=utf-8")
	}
}

// TestShow_LoggedIn_NotSupporter はログイン済み・非サポーターユーザーの場合のテスト
func TestShow_LoggedIn_NotSupporter(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// ユーザーを作成（サブスクリプションなし）
	userID := testutil.NewUserBuilder(t, tx).Build()
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// Checkout セクションが表示されることを確認（非サポーター向け）
	expectedContents := []string{
		"サポーター",                           // ページタイトル
		"¥290",                            // 月額プラン
		"¥2,900",                          // 年額プラン
		"action=\"/supporters/checkout\"", // Checkoutフォーム
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}
}

// TestShow_StripeSupporter はStripeサポーターの場合のテスト
func TestShow_StripeSupporter(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// アクティブなStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "active")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// Stripeサポーターセクションが表示されることを確認
	expectedContents := []string{
		"サポーター",                         // ページタイトル
		"action=\"/supporters/portal\"", // 管理ポータルフォーム
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}

	// 非サポーター向けのCheckoutセクションが表示されないことを確認
	if strings.Contains(body, "action=\"/supporters/checkout\"") {
		t.Error("response contains checkout form when it shouldn't for Stripe supporter")
	}
}

// TestShow_GumroadSupporter はGumroadサポーターの場合のテスト
func TestShow_GumroadSupporter(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// アクティブなGumroadサポーターを作成
	userID, _ := createUserWithGumroadSubscriber(t, tx, false)
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// Gumroadサポーターセクションが表示されることを確認
	// （移行案内メッセージが含まれる緑のセクション）
	if !strings.Contains(body, "bg-green-50") {
		t.Error("response doesn't contain Gumroad supporter section (green background)")
	}

	// 非サポーター向けのCheckoutセクションが表示されないことを確認
	if strings.Contains(body, "action=\"/supporters/checkout\"") {
		t.Error("response contains checkout form when it shouldn't for Gumroad supporter")
	}
}

// TestShow_BothActive は両方アクティブな場合のテスト
func TestShow_BothActive(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// StripeとGumroad両方のサブスクリプションを持つユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()

	stripeSubscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()
	gumroadSubscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()

	// 両方を関連付け
	_, err := tx.Exec(`UPDATE users SET stripe_subscriber_id = $1, gumroad_subscriber_id = $2 WHERE id = $3`,
		stripeSubscriberID, gumroadSubscriberID, userID)
	if err != nil {
		t.Fatalf("サブスクライバーの関連付けに失敗しました: %v", err)
	}

	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// 両方アクティブの場合、Stripeセクションが優先表示される
	expectedContents := []string{
		"action=\"/supporters/portal\"", // Stripe管理ポータル
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}
}

// TestShow_SuccessQueryParam はsuccessクエリパラメータがある場合のテスト
func TestShow_SuccessQueryParam(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	req := httptest.NewRequest("GET", "/supporters?success=true", nil)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// 成功メッセージが表示されることを確認
	if !strings.Contains(body, "bg-success") {
		t.Error("response doesn't contain success message (success background)")
	}
}

// TestShow_CanceledQueryParam はcanceledクエリパラメータがある場合のテスト
func TestShow_CanceledQueryParam(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	req := httptest.NewRequest("GET", "/supporters?canceled=true", nil)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// キャンセルメッセージが表示されることを確認
	if !strings.Contains(body, "bg-warning") {
		t.Error("response doesn't contain canceled message (warning background)")
	}
}

// TestShow_InactiveStripeSubscription は非アクティブなStripeサブスクリプションの場合のテスト
func TestShow_InactiveStripeSubscription(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// キャンセル済みのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "canceled")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// 非アクティブなので、Checkoutセクションが表示される
	if !strings.Contains(body, "action=\"/supporters/checkout\"") {
		t.Error("response doesn't contain checkout form for inactive subscription")
	}
}

// TestShow_EndedGumroadSubscription は終了したGumroadサブスクリプションの場合のテスト
func TestShow_EndedGumroadSubscription(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupTestHandler(t, tx, db)

	// 終了したGumroadサポーターを作成
	userID, _ := createUserWithGumroadSubscriber(t, tx, true)
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// 終了済みなので、Checkoutセクションが表示される
	if !strings.Contains(body, "action=\"/supporters/checkout\"") {
		t.Error("response doesn't contain checkout form for ended subscription")
	}
}
