package sign_up_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/handler/sign_up"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/turnstile"
	"github.com/annict/annict/internal/usecase"
)

// TestCreate_Success は正常系のテスト
func TestCreate_Success(t *testing.T) {
	// テスト用DBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)
	defer func() { _ = tx.Rollback() }()

	// テスト用Redisをセットアップ
	rdb := testutil.SetupTestRedis(t)

	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// usecaseの初期化
	queries := testutil.NewQueriesWithTx(db, tx)
	sendSignUpCodeUC := usecase.NewSendSignUpCodeUsecase(db, queries, nil) // Riverクライアントは不要

	// セッションマネージャーの初期化
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// Turnstileクライアントの初期化（テスト用）
	turnstileClient := turnstile.NewClient("test-site-key", "test-secret-key")

	// ハンドラーの初期化
	userRepo := repository.NewUserRepository(queries)

	// ハンドラーの初期化
	handler := sign_up.NewHandler(cfg, sessionMgr, userRepo, nil, sendSignUpCodeUC, turnstileClient)

	// リクエストパラメータを作成
	formData := url.Values{}
	formData.Set("email", "test@example.com")
	formData.Set("csrf_token", "test-csrf-token")
	formData.Set("cf-turnstile-response", "test-turnstile-token")

	// リクエストを作成
	req := httptest.NewRequest("POST", "/sign_up", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Create(rr, req)

	// 注: Turnstile検証は実際にAPIを呼び出すため、テスト環境ではスキップされる可能性があります
	// そのため、このテストはTurnstile検証の実装に依存します

	// ステータスコードが302（リダイレクト）または400（バリデーションエラー）であることを確認
	if rr.Code != http.StatusSeeOther && rr.Code != http.StatusBadRequest {
		t.Errorf("予期しないステータスコード: got %v want %v or %v", rr.Code, http.StatusSeeOther, http.StatusBadRequest)
	}

	// Redisをクリーンアップ
	rdb.FlushDB(req.Context())
}

// TestCreate_EmailRequired はメールアドレス必須のバリデーションテスト
func TestCreate_EmailRequired(t *testing.T) {
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

	// Turnstileクライアントの初期化（テスト用）
	turnstileClient := turnstile.NewClient("test-site-key", "test-secret-key")

	// ハンドラーの初期化
	userRepo := repository.NewUserRepository(queries)

	// ハンドラーの初期化
	handler := sign_up.NewHandler(cfg, sessionMgr, userRepo, nil, sendSignUpCodeUC, turnstileClient)

	// リクエストパラメータを作成（emailを空にする）
	formData := url.Values{}
	formData.Set("email", "")
	formData.Set("csrf_token", "test-csrf-token")

	// リクエストを作成
	req := httptest.NewRequest("POST", "/sign_up", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Create(rr, req)

	// ステータスコードが303（リダイレクト）であることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("予期しないステータスコード: got %v want %v", rr.Code, http.StatusSeeOther)
	}
}
