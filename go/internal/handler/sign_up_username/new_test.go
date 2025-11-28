package sign_up_username

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/usecase"
)

func TestNew(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)
	cfg := &config.Config{
		Env:    "test",
		Domain: "annict-dev.page",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// テスト用Redisクライアントをセットアップ
	rdb := testutil.SetupTestRedis(t)
	completeSignUpUC := usecase.NewCompleteSignUpUsecase(db, queries, rdb)

	handler := NewHandler(cfg, sessionMgr, rdb, completeSignUpUC)

	tests := []struct {
		name           string
		token          string
		setupRedis     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "トークンなし",
			token:          "",
			expectedStatus: http.StatusSeeOther,
			expectedBody:   "",
		},
		{
			name:           "無効なトークン",
			token:          "invalid_token",
			setupRedis:     func() {},
			expectedStatus: http.StatusSeeOther,
			expectedBody:   "",
		},
		{
			name:  "有効なトークン",
			token: "valid_token",
			setupRedis: func() {
				rdb.Set(context.Background(), "sign_up_token:valid_token", "test@example.com", 0)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "ユーザー名を設定",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupRedis != nil {
				tt.setupRedis()
			}

			url := "/sign_up/username"
			if tt.token != "" {
				url = fmt.Sprintf("/sign_up/username?token=%s", tt.token)
			}

			req := httptest.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()

			handler.New(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("ステータスコードが一致しません: got %v want %v", rr.Code, tt.expectedStatus)
			}

			if tt.expectedBody != "" && rr.Code == http.StatusOK {
				body := rr.Body.String()
				if body == "" || len(body) < 100 {
					t.Errorf("レスポンスボディが空です")
				}
			}
		})
	}
}
