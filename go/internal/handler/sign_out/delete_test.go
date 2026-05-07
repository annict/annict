package sign_out

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

// setupSession はログイン状態をシミュレートするためのセッションを作成し、セッションCookieを返します
func setupSession(t *testing.T, sessionMgr *session.Manager, userID model.UserID) *http.Cookie {
	t.Helper()

	// ダミーリクエストとレスポンスを作成
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	ctx := req.Context()

	// ログインセッションを作成（userID付き）
	if err := sessionMgr.CreateSession(ctx, rr, req, userID); err != nil {
		t.Fatalf("セッションの作成に失敗: %v", err)
	}

	// 作成されたセッションCookieを取得
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == session.SessionKey {
			return cookie
		}
	}

	t.Fatal("セッションCookieが作成されませんでした")
	return nil
}

// TestDelete_WithSession セッションがある場合のログアウトテスト
func TestDelete_WithSession(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("signout_test_user").
		WithEmail("signout_test@example.com").
		Build()

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	handler := NewHandler(sessionMgr)

	// ログインセッションを作成
	sessionCookie := setupSession(t, sessionMgr, userID)

	// DELETEリクエストを作成
	req := httptest.NewRequest("DELETE", "/sign_out", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Delete(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusFound {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusFound)
	}

	// リダイレクト先を確認
	location := rr.Header().Get("Location")
	if location != "/" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/")
	}

	// セッションCookieが削除されているか確認（MaxAge=-1）
	cookies := rr.Result().Cookies()
	var newSessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == session.SessionKey {
			newSessionCookie = cookie
			break
		}
	}

	if newSessionCookie == nil {
		t.Error("セッションCookie削除のレスポンスがありません")
	} else {
		if newSessionCookie.MaxAge != -1 {
			t.Errorf("セッションCookieのMaxAgeが正しくない: got %v want %v", newSessionCookie.MaxAge, -1)
		}
	}

	// DBからセッションが削除されていることを確認
	sessionData, err := sessionMgr.GetSession(context.Background(), sessionCookie.Value)
	if err != nil {
		t.Fatalf("セッション取得エラー: %v", err)
	}
	if sessionData != nil {
		t.Error("DBからセッションが削除されていません")
	}
}

// TestDelete_WithoutSession セッションがない場合のログアウトテスト
func TestDelete_WithoutSession(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	handler := NewHandler(sessionMgr)

	// セッションなしでDELETEリクエストを作成
	req := httptest.NewRequest("DELETE", "/sign_out", nil)
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Delete(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusFound {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusFound)
	}

	// リダイレクト先を確認
	location := rr.Header().Get("Location")
	if location != "/" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/")
	}
}

// TestDelete_POSTMethod HTMLフォームからのPOSTリクエストでも動作することをテスト
func TestDelete_POSTMethod(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("signout_post_user").
		WithEmail("signout_post@example.com").
		Build()

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	handler := NewHandler(sessionMgr)

	// ログインセッションを作成
	sessionCookie := setupSession(t, sessionMgr, userID)

	// POSTリクエストを作成（Method Override経由でDELETEに変換される前の状態）
	req := httptest.NewRequest("POST", "/sign_out", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	// ハンドラーを実行（直接呼び出し）
	handler.Delete(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusFound {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusFound)
	}

	// リダイレクト先を確認
	location := rr.Header().Get("Location")
	if location != "/" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/")
	}
}
