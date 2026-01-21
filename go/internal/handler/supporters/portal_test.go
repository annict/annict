package supporters

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/testutil"
)

// setupPortalTestHandler はテスト用のハンドラーをセットアップします
func setupPortalTestHandler(t *testing.T, tx *sql.Tx, db *sql.DB) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	imageHelper := image.NewHelper(cfg)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)

	// テスト用のStripe設定（テストではStripe APIを呼び出さないため空でOK）
	stripeCfg := &annictStripe.Config{}

	// テスト用には nil を渡す（テストでは Stripe API を呼び出さない）
	return NewHandler(cfg, sessionManager, imageHelper, stripeSubscriberRepo, gumroadSubscriberRepo, stripeCfg, nil)
}

// TestPortal_NotLoggedIn は未ログインユーザーがアクセスした場合のテスト
func TestPortal_NotLoggedIn(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupPortalTestHandler(t, tx, db)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	rr := httptest.NewRecorder()

	handler.Portal(rr, req)

	// ログインページへリダイレクトされることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/sign_in")
	}
}

// TestPortal_NotSupporter は非サポーターユーザーがアクセスした場合のテスト
func TestPortal_NotSupporter(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupPortalTestHandler(t, tx, db)

	// サブスクリプションを持たないユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Portal(rr, req)

	// サポーターページへリダイレクトされることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestPortal_CanceledSubscription はキャンセル済みサブスクリプションの場合のテスト
func TestPortal_CanceledSubscription(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupPortalTestHandler(t, tx, db)

	// canceledステータスのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "canceled")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Portal(rr, req)

	// 非アクティブなのでサポーターページへリダイレクト
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestPortal_UnpaidSubscription はunpaidステータスの場合のテスト
func TestPortal_UnpaidSubscription(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupPortalTestHandler(t, tx, db)

	// unpaidステータスのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "unpaid")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Portal(rr, req)

	// 非アクティブなのでサポーターページへリダイレクト
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestPortal_ActiveSubscription_StripeAPIError はアクティブなサブスクリプションでStripe API呼び出しがエラーの場合のテスト
// 注: テスト環境ではStripe APIが設定されていないため、エラーが発生してリダイレクトされる
func TestPortal_ActiveSubscription_StripeAPIError(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupPortalTestHandler(t, tx, db)

	// アクティブなStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "active")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Portal(rr, req)

	// Stripe APIがテスト環境で設定されていないためエラーになり、リダイレクトされる
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestPortal_PastDueSubscription_StripeAPIError は支払い遅延中のサブスクリプションの場合のテスト
// past_dueはアクティブとして扱われるため、Stripe APIを呼び出す
func TestPortal_PastDueSubscription_StripeAPIError(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupPortalTestHandler(t, tx, db)

	// past_dueステータスのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "past_due")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Portal(rr, req)

	// past_dueはアクティブなのでStripe APIを呼び出すが、テスト環境ではエラーになる
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestPortal_GumroadSubscriber はGumroadサポーターがアクセスした場合のテスト
func TestPortal_GumroadSubscriber(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	handler := setupPortalTestHandler(t, tx, db)

	// アクティブなGumroadサポーターを作成（Stripeサブスクリプションなし）
	userID, _ := createUserWithGumroadSubscriber(t, tx, false)
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("POST", "/supporters/portal", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Portal(rr, req)

	// Stripeサポーターではないのでサポーターページへリダイレクト
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}
