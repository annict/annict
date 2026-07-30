package password_reset

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

// TestNew_PageMeta はパスワードリセット申請ページのPageMeta設定をテストします
func TestNew_PageMeta(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	queries := query.New(db)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// The site key is overridden because config.Load can clear it in test and development
	// when ANNICT_TURNSTILE_DISABLE=true. The positive assertions below must not depend on
	// the caller's environment.
	//
	// [Ja] config.Load は test / dev で ANNICT_TURNSTILE_DISABLE=true のときサイトキーを
	// 空にしうるため、上書きする。以下の存在検証を実行環境に依存させないため。
	cfg.TurnstileSiteKey = "1x00000000000000000000AA"

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)

	handler := NewHandler(cfg, sessionManager, nil, nil, nil)

	tests := []struct {
		name           string
		acceptLanguage string
		expectedTitle  string
	}{
		{
			name:           "日本語タイトル",
			acceptLanguage: "ja",
			expectedTitle:  "パスワードリセット | Annict",
		},
		{
			name:           "英語タイトル",
			acceptLanguage: "en",
			expectedTitle:  "Password Reset | Annict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/password/reset", nil)
			req.Header.Set("Accept-Language", tt.acceptLanguage)
			rr := httptest.NewRecorder()

			testutil.ApplyI18nMiddleware(t, handler.New)(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("パスワードリセット申請フォームの表示が失敗しました: status=%d", rr.Code)
			}

			body := rr.Body.String()

			if !strings.Contains(body, "<title>"+tt.expectedTitle+"</title>") {
				t.Errorf("期待されるタイトルが見つかりません: %q\nレスポンス: %s", tt.expectedTitle, body)
			}

			expectedOGTitle := `<meta property="og:title" content="` + tt.expectedTitle + `">`
			if !strings.Contains(body, expectedOGTitle) {
				t.Errorf("期待されるog:titleが見つかりません: %q", expectedOGTitle)
			}

			expectedPreconnect := `<link rel="preconnect" href="https://challenges.cloudflare.com">`
			if !strings.Contains(body, expectedPreconnect) {
				t.Errorf("期待されるpreconnectが見つかりません: %q", expectedPreconnect)
			}
		})
	}
}
