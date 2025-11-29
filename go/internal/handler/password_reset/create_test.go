package password_reset

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/ratelimit"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/usecase"
)

// mockTurnstileClient はテスト用のモック Turnstile クライアントです
type mockTurnstileClient struct {
	shouldSucceed bool
}

func (m *mockTurnstileClient) Verify(ctx context.Context, token string) (bool, error) {
	return m.shouldSucceed, nil
}

// TestCreate_RateLimiting_IP はIPアドレス単位のRate Limitingをテストします
func TestCreate_RateLimiting_IP(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	rdb := testutil.SetupTestRedis(t)
	queries := query.New(db).WithTx(tx)
	limiter := ratelimit.NewLimiter(rdb)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)
	// Rate Limitingのテストのため、明示的に有効化
	cfg.DisableRateLimit = false

	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, queries, nil)
	handler := NewHandler(cfg, userRepo, sessionManager, limiter, mockClient, createPasswordResetTokenUC)

	for i := 0; i < 5; i++ {
		form := url.Values{}
		form.Add("email", "test"+string(rune(i+48))+"@example.com")
		form.Add("cf-turnstile-response", "valid-token")

		req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("attempt %d: expected status OK, got %d", i+1, rr.Code)
		}
	}

	form := url.Values{}
	form.Add("email", "test5@example.com")
	form.Add("cf-turnstile-response", "valid-token")

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("6th attempt should be rate limited by IP, got status %d", rr.Code)
	}

	form = url.Values{}
	form.Add("email", "test6@example.com")
	form.Add("cf-turnstile-response", "valid-token")

	req = httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "192.168.1.2:12345"
	rr = httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("request from different IP should be allowed, got status %d", rr.Code)
	}
}

// TestCreate_RateLimiting_Email はメールアドレス単位のRate Limitingをテストします
func TestCreate_RateLimiting_Email(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	rdb := testutil.SetupTestRedis(t)
	queries := query.New(db).WithTx(tx)
	limiter := ratelimit.NewLimiter(rdb)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)
	// Rate Limitingのテストのため、明示的に有効化
	cfg.DisableRateLimit = false

	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, queries, nil)
	handler := NewHandler(cfg, userRepo, sessionManager, limiter, mockClient, createPasswordResetTokenUC)

	for i := 0; i < 3; i++ {
		form := url.Values{}
		form.Add("email", "ratelimit@example.com")
		form.Add("cf-turnstile-response", "valid-token")

		req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "192.168.1." + string(rune(i+49)) + ":12345"
		rr := httptest.NewRecorder()

		testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("attempt %d: expected status OK, got %d", i+1, rr.Code)
		}
	}

	form := url.Values{}
	form.Add("email", "ratelimit@example.com")
	form.Add("cf-turnstile-response", "valid-token")

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("4th attempt should be rate limited by email, got status %d", rr.Code)
	}

	form = url.Values{}
	form.Add("email", "different@example.com")
	form.Add("cf-turnstile-response", "valid-token")

	req = httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "192.168.1.100:12345"
	rr = httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("request with different email should be allowed, got status %d", rr.Code)
	}
}

// TestPasswordResetSentPage_UXMessages はメール送信完了ページのUXメッセージをテストします
func TestPasswordResetSentPage_UXMessages(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	encryptedPassword, _ := auth.HashPassword("Password123!")
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("ux_sent_test_user").
		WithEmail("ux_sent_test@example.com").
		WithEncryptedPassword(encryptedPassword).
		Build()

	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
	})

	queries := query.New(db)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, queries, nil)
	handler := NewHandler(cfg, userRepo, sessionManager, nil, mockClient, createPasswordResetTokenUC)

	tests := []struct {
		name     string
		locale   string
		expected []string
	}{
		{
			name:   "日本語メッセージ",
			locale: "ja",
			expected: []string{
				"メールを確認してください",
				"パスワードリセット用のリンクを送信しました",
				"メールが届かない場合",
				"迷惑メールフォルダを確認してください",
				`<a href="/password/reset"`,
			},
		},
		{
			name:   "英語メッセージ",
			locale: "en",
			expected: []string{
				"Please check your email",
				"We have sent you a password reset link",
				"If you don&#39;t receive the email",
				"Please check your spam folder",
				`<a href="/password/reset"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("email", "ux_sent_test@example.com")
			form.Add("cf-turnstile-response", "valid-token")

			req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept-Language", tt.locale)

			ctx := i18n.SetLocale(req.Context(), tt.locale)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusOK)
			}

			body := rr.Body.String()
			for _, exp := range tt.expected {
				if !strings.Contains(body, exp) {
					t.Errorf("期待されるメッセージが見つかりません: %q", exp)
				}
			}
		})
	}
}

// TestPasswordResetFlow_Integration はパスワードリセット申請のフローをテストします
func TestPasswordResetFlow_Integration(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	oldPassword := "OldPassword123!"
	encryptedPassword, err := auth.HashPassword(oldPassword)
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("integration_test_user").
		WithEmail("integration@example.com").
		WithEncryptedPassword(encryptedPassword).
		Build()

	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
	})

	queries := query.New(db)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, queries, nil)
	handler := NewHandler(cfg, userRepo, sessionManager, nil, mockClient, createPasswordResetTokenUC)

	ctx := context.Background()

	form := url.Values{}
	form.Add("email", "integration@example.com")
	form.Add("cf-turnstile-response", "valid-token")

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("パスワードリセット申請が失敗しました: status=%d", rr.Code)
	}

	tokens, err := queries.GetPasswordResetTokensByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}

	if len(tokens) == 0 {
		t.Fatal("トークンがデータベースに保存されていません")
	}

	t.Logf("統合テスト: トークンが正常に作成されました (user_id=%d, token_count=%d)", userID, len(tokens))
}

// TestCreate_TurnstileVerification_Success はTurnstile検証が成功した場合のテストです
func TestCreate_TurnstileVerification_Success(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, queries, nil)
	handler := NewHandler(cfg, userRepo, sessionManager, nil, mockClient, createPasswordResetTokenUC)

	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("cf-turnstile-response", "valid-token")

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Turnstile検証が成功したはずですが、ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusOK)
	}
}

// TestCreate_TurnstileVerification_Failed はTurnstile検証が失敗した場合のテストです
func TestCreate_TurnstileVerification_Failed(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// モック Turnstile クライアント（常に失敗）
	mockClient := &mockTurnstileClient{shouldSucceed: false}
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, queries, nil)
	handler := NewHandler(cfg, userRepo, sessionManager, nil, mockClient, createPasswordResetTokenUC)

	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("cf-turnstile-response", "invalid-token")

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Turnstile検証が失敗したはずですが、ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/password/reset" {
		t.Errorf("Turnstile検証失敗時のリダイレクト先が正しくありません: got=%s, want=%s", location, "/password/reset")
	}
}

// TestCreate_TurnstileVerification_MissingToken はTurnstileトークンが欠落している場合のテストです
func TestCreate_TurnstileVerification_MissingToken(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// モック Turnstile クライアント（常に失敗）
	mockClient := &mockTurnstileClient{shouldSucceed: false}
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, queries, nil)
	handler := NewHandler(cfg, userRepo, sessionManager, nil, mockClient, createPasswordResetTokenUC)

	form := url.Values{}
	form.Add("email", "test@example.com")
	// cf-turnstile-response を含めない

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Turnstileトークンが欠落している場合、ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/password/reset" {
		t.Errorf("Turnstileトークン欠落時のリダイレクト先が正しくありません: got=%s, want=%s", location, "/password/reset")
	}
}
