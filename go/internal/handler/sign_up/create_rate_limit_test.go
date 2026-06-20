package sign_up_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/handler/sign_up"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/turnstile"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

// TestCreate_RateLimiting_IP はIP単位のRate Limitingテスト
func TestCreate_RateLimiting_IP(t *testing.T) {
	t.Parallel()

	// テスト用DBとトランザクションをセットアップ
	db, tx := testutil.SetupTx(t)

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
	v := validator.NewSignUpCreateValidator()
	sendSignUpCodeUC := usecase.NewSendSignUpCodeUsecase(db, repository.NewSignUpCodeRepository(queries), repository.NewUserRepository(queries), nil, v)

	// セッションマネージャーの初期化
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// Turnstileクライアントの初期化（テスト用）
	turnstileClient := turnstile.NewClient("test-site-key", "test-secret-key")

	// Rate Limiterの初期化
	limiter := ratelimit.NewLimiter(rdb)

	// ハンドラーの初期化
	handler := sign_up.NewHandler(cfg, sessionMgr, testutil.NewTestFlashManager(), limiter, sendSignUpCodeUC, turnstileClient)

	// Build per-test unique values so this test does not collide with
	// other tests running in parallel against the shared Redis DB.
	//
	// [Ja] 共有 Redis DB に対する並列実行で他テストと衝突しないよう、
	// 本テスト固有のキー構成値を組み立てる。
	prefix := testutil.UniqueRateLimitPrefix(t)
	clientIP := prefix + "-ip"
	email := prefix + "@example.com"
	_ = limiter.Reset(ctx, "sign_up:ip:"+clientIP)
	_ = limiter.Reset(ctx, "sign_up:email:"+email)
	t.Cleanup(func() {
		_ = limiter.Reset(ctx, "sign_up:ip:"+clientIP)
		_ = limiter.Reset(ctx, "sign_up:email:"+email)
	})

	// 同一IPから6回アクセス（制限: 5回/時間）
	for i := 0; i < 6; i++ {
		// リクエストパラメータを作成
		formData := url.Values{}
		formData.Set("email", email)
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
			// 最初の5回は成功（303）または検証失敗（422）
			if rr.Code != http.StatusSeeOther && rr.Code != http.StatusUnprocessableEntity {
				t.Errorf("リクエスト %d: 予期しないステータスコード: got %v", i+1, rr.Code)
			}
		} else {
			// 6回目はRate Limitingで失敗（422でフォーム再描画）
			if rr.Code != http.StatusUnprocessableEntity {
				t.Errorf("リクエスト %d: Rate Limitingが発動していません: got %v want %v", i+1, rr.Code, http.StatusUnprocessableEntity)
			}
		}
	}
}

// TestCreate_RateLimiting_Email はメールアドレス単位のRate Limitingテスト
func TestCreate_RateLimiting_Email(t *testing.T) {
	t.Parallel()

	// テスト用DBとトランザクションをセットアップ
	db, tx := testutil.SetupTx(t)

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
	v := validator.NewSignUpCreateValidator()
	sendSignUpCodeUC := usecase.NewSendSignUpCodeUsecase(db, repository.NewSignUpCodeRepository(queries), repository.NewUserRepository(queries), nil, v)

	// セッションマネージャーの初期化
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// Turnstileクライアントの初期化（テスト用）
	turnstileClient := turnstile.NewClient("test-site-key", "test-secret-key")

	// Rate Limiterの初期化
	limiter := ratelimit.NewLimiter(rdb)

	// ハンドラーの初期化
	handler := sign_up.NewHandler(cfg, sessionMgr, testutil.NewTestFlashManager(), limiter, sendSignUpCodeUC, turnstileClient)

	// Build per-test unique values so this test does not collide with
	// other tests running in parallel against the shared Redis DB.
	//
	// [Ja] 共有 Redis DB に対する並列実行で他テストと衝突しないよう、
	// 本テスト固有のキー構成値を組み立てる。
	prefix := testutil.UniqueRateLimitPrefix(t)
	email := prefix + "@example.com"
	ipFor := func(i int) string { return prefix + "-ip" + strconv.Itoa(i) }

	resetKeys := func() {
		_ = limiter.Reset(ctx, "sign_up:email:"+email)
		for i := 0; i < 4; i++ {
			_ = limiter.Reset(ctx, "sign_up:ip:"+ipFor(i))
		}
	}
	resetKeys()
	t.Cleanup(resetKeys)

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
		req.RemoteAddr = ipFor(i) + ":12345"

		// レスポンスレコーダーを作成
		rr := httptest.NewRecorder()

		// ハンドラーを実行
		handler.Create(rr, req)

		if i < 3 {
			// 最初の3回は成功（303）または検証失敗（422）
			if rr.Code != http.StatusSeeOther && rr.Code != http.StatusUnprocessableEntity {
				t.Errorf("リクエスト %d: 予期しないステータスコード: got %v", i+1, rr.Code)
			}
		} else {
			// 4回目はRate Limitingで失敗（422でフォーム再描画）
			if rr.Code != http.StatusUnprocessableEntity {
				t.Errorf("リクエスト %d: Rate Limitingが発動していません: got %v want %v", i+1, rr.Code, http.StatusUnprocessableEntity)
			}
		}
	}
}
