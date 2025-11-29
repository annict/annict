package sign_up_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/handler/sign_up"
	"github.com/annict/annict/internal/ratelimit"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/turnstile"
	"github.com/annict/annict/internal/usecase"
)

// TestCreate_RateLimiting_IP はIP単位のRate Limitingテスト
func TestCreate_RateLimiting_IP(t *testing.T) {
	// テスト用DBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)
	defer func() { _ = tx.Rollback() }()

	// テスト用Redisをセットアップ
	rdb := testutil.SetupTestRedis(t)
	ctx := context.Background()

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

	// Turnstileクライアントの初期化（テスト用）
	turnstileClient := turnstile.NewClient("test-site-key", "test-secret-key")

	// Rate Limiterの初期化
	limiter := ratelimit.NewLimiter(rdb)

	// ハンドラーの初期化
	userRepo := repository.NewUserRepository(queries)
	handler := sign_up.NewHandler(cfg, sessionMgr, userRepo, limiter, sendSignUpCodeUC, turnstileClient)

	// 同一IPから6回アクセス（制限: 5回/時間）
	clientIP := "203.0.113.1"
	for i := 0; i < 6; i++ {
		// リクエストパラメータを作成
		formData := url.Values{}
		formData.Set("email", "test@example.com")
		formData.Set("terms_agreed", "true")
		formData.Set("csrf_token", "test-csrf-token")
		formData.Set("cf-turnstile-response", "test-turnstile-token")

		// リクエストを作成
		req := httptest.NewRequest("POST", "/sign_up", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = clientIP + ":12345"

		// レスポンスレコーダーを作成
		rr := httptest.NewRecorder()

		// ハンドラーを実行
		handler.Create(rr, req)

		if i < 5 {
			// 最初の5回は成功（302または400）
			if rr.Code != http.StatusSeeOther && rr.Code != http.StatusBadRequest {
				t.Errorf("リクエスト %d: 予期しないステータスコード: got %v", i+1, rr.Code)
			}
		} else {
			// 6回目はRate Limitingで失敗（303リダイレクト）
			if rr.Code != http.StatusSeeOther {
				t.Errorf("リクエスト %d: Rate Limitingが発動していません: got %v want %v", i+1, rr.Code, http.StatusSeeOther)
			}
		}
	}

	// Redisをクリーンアップ
	rdb.FlushDB(ctx)
}

// TestCreate_RateLimiting_Email はメールアドレス単位のRate Limitingテスト
func TestCreate_RateLimiting_Email(t *testing.T) {
	// テスト用DBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)
	defer func() { _ = tx.Rollback() }()

	// テスト用Redisをセットアップ
	rdb := testutil.SetupTestRedis(t)
	ctx := context.Background()

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

	// Turnstileクライアントの初期化（テスト用）
	turnstileClient := turnstile.NewClient("test-site-key", "test-secret-key")

	// Rate Limiterの初期化
	limiter := ratelimit.NewLimiter(rdb)

	// ハンドラーの初期化
	userRepo := repository.NewUserRepository(queries)
	handler := sign_up.NewHandler(cfg, sessionMgr, userRepo, limiter, sendSignUpCodeUC, turnstileClient)

	// 同一メールアドレスで4回アクセス（制限: 3回/時間）
	email := "test@example.com"
	for i := 0; i < 4; i++ {
		// リクエストパラメータを作成
		formData := url.Values{}
		formData.Set("email", email)
		formData.Set("terms_agreed", "true")
		formData.Set("csrf_token", "test-csrf-token")
		formData.Set("cf-turnstile-response", "test-turnstile-token")

		// リクエストを作成（異なるIPアドレスから）
		req := httptest.NewRequest("POST", "/sign_up", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "203.0.113." + string(rune(1+i)) + ":12345"

		// レスポンスレコーダーを作成
		rr := httptest.NewRecorder()

		// ハンドラーを実行
		handler.Create(rr, req)

		if i < 3 {
			// 最初の3回は成功（302または400）
			if rr.Code != http.StatusSeeOther && rr.Code != http.StatusBadRequest {
				t.Errorf("リクエスト %d: 予期しないステータスコード: got %v", i+1, rr.Code)
			}
		} else {
			// 4回目はRate Limitingで失敗（303リダイレクト）
			if rr.Code != http.StatusSeeOther {
				t.Errorf("リクエスト %d: Rate Limitingが発動していません: got %v want %v", i+1, rr.Code, http.StatusSeeOther)
			}
		}
	}

	// Redisをクリーンアップ
	rdb.FlushDB(ctx)
}
