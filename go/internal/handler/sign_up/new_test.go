package sign_up_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/handler/sign_up"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/turnstile"
	"github.com/annict/annict/internal/usecase"
)

// TestNew は新規登録フォーム表示のテスト
func TestNew(t *testing.T) {
	// テスト用DBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)
	defer func() { _ = tx.Rollback() }()

	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// usecaseの初期化
	queries := testutil.NewQueriesWithTx(db, tx)
	sendSignUpCodeUC := usecase.NewSendSignUpCodeUsecase(db, queries, nil)

	// セッションマネージャーの初期化
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// UserRepositoryの初期化
	userRepo := repository.NewUserRepository(queries)

	// Turnstileクライアントの初期化（テスト用）
	turnstileClient := turnstile.NewClient("test-site-key", "test-secret-key")

	// ハンドラーの初期化
	handler := sign_up.NewHandler(cfg, sessionMgr, userRepo, nil, sendSignUpCodeUC, turnstileClient)

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
}
