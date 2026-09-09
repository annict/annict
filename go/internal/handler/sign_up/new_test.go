package sign_up_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/handler/sign_up"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/turnstile"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

// TestNew は新規登録フォーム表示のテスト
func TestNew(t *testing.T) {
	t.Parallel()

	// テスト用DBとトランザクションをセットアップ
	db, tx := testutil.SetupTx(t)

	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// The site key is overridden because config.Load can clear it in test and development
	// when ANNICT_TURNSTILE_DISABLE=true. The positive assertions below must not depend on
	// the caller's environment.
	//
	// [Ja] config.Load は test / dev で ANNICT_TURNSTILE_DISABLE=true のときサイトキーを
	// 空にしうるため、上書きする。以下の存在検証を実行環境に依存させないため。
	cfg.TurnstileSiteKey = "1x00000000000000000000AA"

	// usecaseの初期化
	queries := testutil.NewQueriesWithTx(db, tx)
	v := validator.NewSignUpCreateValidator()
	sendSignUpCodeUC := usecase.NewSendSignUpCodeUsecase(db, repository.NewSignUpCodeRepository(queries), repository.NewUserRepository(queries), nil, v)

	// セッションマネージャーの初期化
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// Turnstileクライアントの初期化（テスト用）
	turnstileClient := turnstile.NewClient("test-site-key", "test-secret-key")

	// ハンドラーの初期化
	handler := sign_up.NewHandler(cfg, sessionMgr, testutil.NewTestFlashManager(), nil, sendSignUpCodeUC, turnstileClient)

	// リクエストを作成
	req := httptest.NewRequest("GET", "/sign_up", nil)

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.New(rr, req)

	// ステータスコードが200であることを確認
	if rr.Code != http.StatusOK {
		t.Errorf("予期しないステータスコード: got %v want %v", rr.Code, http.StatusOK)
	}

	// Content-Typeがtext/htmlであることを確認
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("予期しないContent-Type: got %v want %v", contentType, "text/html; charset=utf-8")
	}

	body := rr.Body.String()
	expectedStrings := []string{
		`class="cf-turnstile"`,
		`<link rel="preconnect" href="https://challenges.cloudflare.com">`,
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスボディに期待される文字列が含まれていない: %s", expected)
		}
	}
}
