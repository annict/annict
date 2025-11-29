package sign_up_username

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/usecase"
)

func TestCreate(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)
	cfg := &config.Config{
		Env:           "test",
		Domain:        "annict-dev.page",
		SessionSecure: "false",
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
		username       string
		setupRedis     func()
		expectedStatus int
		checkUser      bool
	}{
		{
			name:           "トークンなし",
			token:          "",
			username:       "testuser",
			expectedStatus: http.StatusSeeOther,
			checkUser:      false,
		},
		{
			name:     "ユーザー名なし",
			token:    "valid_token",
			username: "",
			setupRedis: func() {
				rdb.Set(context.Background(), "sign_up_token:valid_token", "test@example.com", 0)
			},
			expectedStatus: http.StatusSeeOther,
			checkUser:      false,
		},
		{
			name:     "ユーザー名が長すぎる",
			token:    "valid_token2",
			username: "verylongusernamemorethan20chars",
			setupRedis: func() {
				rdb.Set(context.Background(), "sign_up_token:valid_token2", "test2@example.com", 0)
			},
			expectedStatus: http.StatusSeeOther,
			checkUser:      false,
		},
		{
			name:     "正常系",
			token:    "valid_token3",
			username: "testuser3",
			setupRedis: func() {
				rdb.Set(context.Background(), "sign_up_token:valid_token3", "test3@example.com", 0)
			},
			expectedStatus: http.StatusSeeOther,
			checkUser:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupRedis != nil {
				tt.setupRedis()
			}

			form := url.Values{}
			form.Add("token", tt.token)
			form.Add("username", tt.username)

			req := httptest.NewRequest("POST", "/sign_up/username", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("ステータスコードが一致しません: got %v want %v", rr.Code, tt.expectedStatus)
			}

			if tt.checkUser && rr.Code == http.StatusSeeOther {
				// ユーザーが作成されたか確認
				user, err := queries.GetUserByUsername(context.Background(), tt.username)
				if err != nil {
					if err == sql.ErrNoRows {
						t.Errorf("ユーザーが作成されていません: username=%s", tt.username)
					} else {
						t.Errorf("ユーザー取得エラー: %v", err)
					}
				} else {
					if user.Username != tt.username {
						t.Errorf("ユーザー名が一致しません: got %v want %v", user.Username, tt.username)
					}
				}

				// セッションCookieが設定されているか確認
				cookies := rr.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if cookie.Name == session.SessionKey {
						found = true
						if cookie.Value == "" {
							t.Error("セッションCookieの値が空です")
						}
					}
				}
				if !found {
					t.Error("セッションCookieが設定されていません")
				}
			}
		})
	}
}

func TestCreate_UsernameTaken(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)
	cfg := &config.Config{
		Env:           "test",
		Domain:        "annict-dev.page",
		SessionSecure: "false",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// テスト用Redisクライアントをセットアップ
	rdb := testutil.SetupTestRedis(t)
	completeSignUpUC := usecase.NewCompleteSignUpUsecase(db, queries, rdb)

	handler := NewHandler(cfg, sessionMgr, rdb, completeSignUpUC)

	// 既存ユーザーを作成
	existingUser := testutil.NewUserBuilder(t, tx).
		WithUsername("existinguser").
		WithEmail("existing@example.com").
		Build()

	if existingUser == 0 {
		t.Fatal("既存ユーザーの作成に失敗しました")
	}

	// Redisにトークンを設定
	token := "valid_token_duplicate"
	email := "newuser@example.com"
	rdb.Set(context.Background(), fmt.Sprintf("sign_up_token:%s", token), email, 0)

	// 既存のユーザー名で登録を試みる
	form := url.Values{}
	form.Add("token", token)
	form.Add("username", "existinguser")

	req := httptest.NewRequest("POST", "/sign_up/username", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// リダイレクトされるべき
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが一致しません: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// 新しいユーザーは作成されていないはず
	_, err := queries.GetUserByEmail(context.Background(), email)
	if err == nil {
		t.Error("重複したユーザー名でユーザーが作成されてしまいました")
	} else if err != sql.ErrNoRows {
		t.Errorf("予期しないエラー: %v", err)
	}
}
