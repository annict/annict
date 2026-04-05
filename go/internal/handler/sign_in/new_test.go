package sign_in

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/turnstile"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

// TestNew GET /sign_inのテスト
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:       ".example.com",
		SessionSecure:      "false",                               // テスト環境ではfalse
		TurnstileSiteKey:   "1x00000000000000000000AA",            // テスト用Site Key
		TurnstileSecretKey: "1x0000000000000000000000000000000AA", // テスト用Secret Key
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// ログインコード送信ユースケースを作成（Dispatcher は nil でメール送信をスキップ）
	v := validator.NewCreateSignInValidator()
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, repository.NewSignInCodeRepository(queries), repository.NewUserRepository(queries), nil, v)

	// Turnstile クライアントを作成（テスト環境用: 空のSecretKeyで検証をスキップ）
	turnstileClient := turnstile.NewClient("", "")

	handler := NewHandler(cfg, sessionMgr, sendSignInCodeUC, turnstileClient)

	// リクエストを作成
	req := httptest.NewRequest("GET", "/sign_in", nil)
	rr := httptest.NewRecorder()

	// I18nミドルウェアを適用（テストでもlocaleを設定）
	testutil.ApplyI18nMiddleware(t, handler.New)(rr, req)

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
		"Annictにログイン",       // ページタイトル
		"メールアドレス",           // フォームラベル
		"次へ",                // ボタンテキスト
		"新規登録",              // サインアップリンク
		"csrf_token",        // CSRFトークン
		`action="/sign_in"`, // フォームアクション
		"https://challenges.cloudflare.com/turnstile/v0/api.js", // Turnstile JavaScript
		`class="cf-turnstile"`,                    // Turnstile ウィジェット
		`data-sitekey="1x00000000000000000000AA"`, // Turnstile Site Key
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスボディに期待される文字列が含まれていない: %s", expected)
		}
	}
}
