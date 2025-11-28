package middleware_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/middleware"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
)

// generatePrivateID はテスト用のprivate ID生成関数（Repositoryの実装と同じロジック）
func generatePrivateID(publicID string) string {
	hash := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("2::%s", hex.EncodeToString(hash[:]))
}

func TestRequireAuth_UpdatesSessionUpdatedAt(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("test@example.com").
		Build()

	// セッションを作成
	queries := query.New(db).WithTx(tx)
	publicID := testutil.NewSessionBuilder(t, tx).
		WithSessionID("test-session-id-12345").
		WithUserID(userID).
		Build()

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)

	// セッションの初期updated_atを取得
	privateID := generatePrivateID(publicID)
	initialSession, err := queries.GetSessionByID(context.Background(), privateID)
	if err != nil {
		t.Fatalf("初期セッションの取得に失敗しました: %v", err)
	}

	// 少し待ってからTouchSessionを実行
	time.Sleep(100 * time.Millisecond)

	// 認証ミドルウェアを作成
	authMW := middleware.NewAuthMiddleware(sessionManager, sessionRepo)

	// テスト用のハンドラー（認証が必要）
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// リクエストを作成（セッションクッキー付き）
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  session.SessionKey,
		Value: publicID,
	})

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// Middlewareミドルウェアを適用してユーザー情報をコンテキストに設定し、その後RequireAuthを適用
	handler := authMW.Middleware(authMW.RequireAuth(testHandler))
	handler.ServeHTTP(rr, req)

	// ステータスコードが200であることを確認
	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// セッションのupdated_atが更新されていることを確認
	updatedSession, err := queries.GetSessionByID(context.Background(), privateID)
	if err != nil {
		t.Fatalf("セッションの取得に失敗しました: %v", err)
	}

	// updated_atが初期値より新しいことを確認
	if !updatedSession.UpdatedAt.After(initialSession.UpdatedAt) {
		t.Errorf("セッションのupdated_atが更新されていません: initial=%v, updated=%v",
			initialSession.UpdatedAt, updatedSession.UpdatedAt)
	}
}

func TestRequireAuth_RedirectsWhenNotAuthenticated(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)

	// セッションマネージャーとリポジトリを作成
	queries := query.New(db).WithTx(tx)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)

	// 認証ミドルウェアを作成
	authMW := middleware.NewAuthMiddleware(sessionManager, sessionRepo)

	// テスト用のハンドラー（認証が必要）
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// リクエストを作成（未認証）
	req := httptest.NewRequest("GET", "/test", nil)

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// RequireAuthミドルウェアを適用
	authMW.RequireAuth(testHandler).ServeHTTP(rr, req)

	// ステータスコードが303（See Other）であることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先が/sign_in?back=<元のURL>であることを確認
	location := rr.Header().Get("Location")
	expectedLocation := "/sign_in?back=%2Ftest"
	if location != expectedLocation {
		t.Errorf("wrong redirect location: got %v want %v", location, expectedLocation)
	}
}

func TestRequireAuth_RedirectsWhenSessionIDNotFound(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("test@example.com").
		Build()

	// セッションマネージャーとリポジトリを作成
	queries := query.New(db).WithTx(tx)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)

	// 認証ミドルウェアを作成
	authMW := middleware.NewAuthMiddleware(sessionManager, sessionRepo)

	// テスト用のハンドラー（認証が必要）
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// リクエストを作成（セッションクッキーなし、ただしコンテキストにユーザー情報あり）
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, &query.GetUserByIDRow{
		ID:    userID,
		Email: "test@example.com",
	}))

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// RequireAuthミドルウェアを適用
	authMW.RequireAuth(testHandler).ServeHTTP(rr, req)

	// ステータスコードが303（See Other）であることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先が/sign_in?back=<元のURL>であることを確認
	location := rr.Header().Get("Location")
	expectedLocation := "/sign_in?back=%2Ftest"
	if location != expectedLocation {
		t.Errorf("wrong redirect location: got %v want %v", location, expectedLocation)
	}
}

func TestRequireAuth_RedirectsWithBackParam(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)

	// セッションマネージャーとリポジトリを作成
	queries := query.New(db).WithTx(tx)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)

	// 認証ミドルウェアを作成
	authMW := middleware.NewAuthMiddleware(sessionManager, sessionRepo)

	// テスト用のハンドラー（認証が必要）
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name             string
		requestURL       string
		expectedLocation string
	}{
		{
			name:             "シンプルなパス",
			requestURL:       "/oauth/authorize",
			expectedLocation: "/sign_in?back=%2Foauth%2Fauthorize",
		},
		{
			name:             "クエリパラメータ付きのパス",
			requestURL:       "/oauth/authorize?client_id=xxx&redirect_uri=http://example.com",
			expectedLocation: "/sign_in?back=%2Foauth%2Fauthorize%3Fclient_id%3Dxxx%26redirect_uri%3Dhttp%3A%2F%2Fexample.com",
		},
		{
			name:             "ルートパス",
			requestURL:       "/",
			expectedLocation: "/sign_in?back=%2F",
		},
		{
			name:             "複数のクエリパラメータ",
			requestURL:       "/works?season=2024-spring&sort=popular",
			expectedLocation: "/sign_in?back=%2Fworks%3Fseason%3D2024-spring%26sort%3Dpopular",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.requestURL, nil)
			rr := httptest.NewRecorder()

			authMW.RequireAuth(testHandler).ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Errorf("ステータスコードが異なります: got %v, want %v", rr.Code, http.StatusSeeOther)
			}

			location := rr.Header().Get("Location")
			if location != tt.expectedLocation {
				t.Errorf("リダイレクト先が異なります:\n  got:  %v\n  want: %v", location, tt.expectedLocation)
			}
		})
	}
}
