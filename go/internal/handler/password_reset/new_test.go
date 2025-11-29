package password_reset

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
)

// TestNew_PageMeta はパスワードリセット申請ページのPageMeta設定をテストします
func TestNew_PageMeta(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	queries := query.New(db)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	handler := NewHandler(cfg, userRepo, sessionManager, nil, nil, nil)

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
		})
	}
}
