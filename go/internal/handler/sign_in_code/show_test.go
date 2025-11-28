package sign_in_code

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/usecase"
)

// TestShow GET /sign_in/codeのテスト（正常系）
func TestShow(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:     ".example.com",
		SessionSecure:    "false", // テスト環境ではfalse
		DisableRateLimit: true,
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ユースケースを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// リクエストを作成
	req := httptest.NewRequest("GET", "/sign_in/code", nil)
	rr := httptest.NewRecorder()

	// セッションにメールアドレスを設定
	testEmail := "test@example.com"
	ctx := context.Background()
	if err := sessionMgr.SetValue(ctx, rr, req, "sign_in_email", testEmail); err != nil {
		t.Fatalf("セッション値の設定に失敗しました: %v", err)
	}

	// セッションクッキーを取得してリクエストに設定
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// レスポンスレコーダーをリセット
	rr = httptest.NewRecorder()

	// I18nミドルウェアを適用（テストでもlocaleを設定）
	testutil.ApplyI18nMiddleware(t, handler.Show)(rr, req)

	// ステータスコードを確認
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusOK)
	}

	// Content-Typeを確認
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Typeが正しくない: got %v", contentType)
	}

	// レスポンスボディに期待される文字列が含まれているか確認
	body := rr.Body.String()
	expectedStrings := []string{
		"6桁のコードを入力",                    // ページタイトル
		"6桁のコード",                       // フォームラベル
		"ログイン",                         // ボタンテキスト
		"コードが届かない場合",                   // 再送信案内
		"コードを再送信",                      // 再送信ボタン
		"戻る",                           // 戻るリンク
		"csrf_token",                   // CSRFトークン
		`action="/sign_in/code"`,       // フォームアクション
		`pattern="[0-9]{6}"`,           // コード入力パターン
		`inputmode="numeric"`,          // 数字入力モード
		`autocomplete="one-time-code"`, // ワンタイムコード
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスボディに期待される文字列が含まれていない: %s", expected)
		}
	}
}

// TestShow_NoEmailInSession セッションにメールアドレスがない場合のテスト
func TestShow_NoEmailInSession(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:     ".example.com",
		SessionSecure:    "false", // テスト環境ではfalse
		DisableRateLimit: true,
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ユースケースを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// リクエストを作成（セッションにメールアドレスを設定しない）
	req := httptest.NewRequest("GET", "/sign_in/code", nil)
	rr := httptest.NewRecorder()

	// I18nミドルウェアを適用（テストでもlocaleを設定）
	testutil.ApplyI18nMiddleware(t, handler.Show)(rr, req)

	// ステータスコードを確認（/sign_in にリダイレクトされる）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先を確認
	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("リダイレクト先が正しくない: got %v want /sign_in", location)
	}
}
