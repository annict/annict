package password_reset

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	rdb := testutil.SetupTestRedis(t)
	queries := query.New(db).WithTx(tx)
	limiter := ratelimit.NewLimiter(rdb)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	// Rate Limitingのテストのため、明示的に有効化
	cfg.DisableRateLimit = false

	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, limiter, mockClient, createPasswordResetTokenUC)

	// Build per-test unique values so this test does not collide with
	// other tests running in parallel against the shared Redis DB.
	//
	// [Ja] 共有 Redis DB に対する並列実行で他テストと衝突しないよう、
	// 本テスト固有のキー構成値を組み立てる。
	prefix := testutil.UniqueRateLimitPrefix(t)
	primaryIP := prefix + "-ip1"
	altIP := prefix + "-ip2"
	emailFor := func(i int) string { return fmt.Sprintf("%s-%d@example.com", prefix, i) }

	ctx := context.Background()
	resetKeys := func() {
		_ = limiter.Reset(ctx, "password_reset:ip:"+primaryIP)
		_ = limiter.Reset(ctx, "password_reset:ip:"+altIP)
		for i := 0; i < 7; i++ {
			_ = limiter.Reset(ctx, "password_reset:email:"+emailFor(i))
		}
	}
	resetKeys()
	t.Cleanup(resetKeys)

	for i := 0; i < 5; i++ {
		form := url.Values{}
		form.Add("email", emailFor(i))
		form.Add("cf-turnstile-response", "valid-token")

		req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = primaryIP + ":12345"
		rr := httptest.NewRecorder()

		testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("attempt %d: expected status OK, got %d", i+1, rr.Code)
		}
	}

	form := url.Values{}
	form.Add("email", emailFor(5))
	form.Add("cf-turnstile-response", "valid-token")

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = primaryIP + ":12345"
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("6th attempt should be rate limited by IP (422), got status %d", rr.Code)
	}

	form = url.Values{}
	form.Add("email", emailFor(6))
	form.Add("cf-turnstile-response", "valid-token")

	req = httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = altIP + ":12345"
	rr = httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("request from different IP should be allowed, got status %d", rr.Code)
	}
}

// TestCreate_RateLimiting_Email はメールアドレス単位のRate Limitingをテストします
func TestCreate_RateLimiting_Email(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	rdb := testutil.SetupTestRedis(t)
	queries := query.New(db).WithTx(tx)
	limiter := ratelimit.NewLimiter(rdb)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	// Rate Limitingのテストのため、明示的に有効化
	cfg.DisableRateLimit = false

	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, limiter, mockClient, createPasswordResetTokenUC)

	// Build per-test unique values so this test does not collide with
	// other tests running in parallel against the shared Redis DB.
	//
	// [Ja] 共有 Redis DB に対する並列実行で他テストと衝突しないよう、
	// 本テスト固有のキー構成値を組み立てる。
	prefix := testutil.UniqueRateLimitPrefix(t)
	primaryEmail := prefix + "@example.com"
	altEmail := prefix + "-alt@example.com"
	ipFor := func(i int) string { return prefix + "-ip" + strconv.Itoa(i) }
	altIP := prefix + "-ip-alt"

	ctx := context.Background()
	resetKeys := func() {
		_ = limiter.Reset(ctx, "password_reset:email:"+primaryEmail)
		_ = limiter.Reset(ctx, "password_reset:email:"+altEmail)
		for i := 0; i < 3; i++ {
			_ = limiter.Reset(ctx, "password_reset:ip:"+ipFor(i))
		}
		_ = limiter.Reset(ctx, "password_reset:ip:"+altIP)
	}
	resetKeys()
	t.Cleanup(resetKeys)

	for i := 0; i < 3; i++ {
		form := url.Values{}
		form.Add("email", primaryEmail)
		form.Add("cf-turnstile-response", "valid-token")

		req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = ipFor(i) + ":12345"
		rr := httptest.NewRecorder()

		testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("attempt %d: expected status OK, got %d", i+1, rr.Code)
		}
	}

	form := url.Values{}
	form.Add("email", primaryEmail)
	form.Add("cf-turnstile-response", "valid-token")

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = altIP + ":12345"
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("4th attempt should be rate limited by email (422), got status %d", rr.Code)
	}

	form = url.Values{}
	form.Add("email", altEmail)
	form.Add("cf-turnstile-response", "valid-token")

	req = httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = altIP + ":12345"
	rr = httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("request with different email should be allowed, got status %d", rr.Code)
	}
}

// TestPasswordResetSentPage_UXMessages はメール送信完了ページのUXメッセージをテストします
func TestPasswordResetSentPage_UXMessages(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

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
	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, nil, mockClient, createPasswordResetTokenUC)

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
	t.Parallel()

	db, tx := testutil.SetupTx(t)

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
	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, nil, mockClient, createPasswordResetTokenUC)

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

	tokens, err := repository.NewPasswordResetTokenRepository(queries).GetByUserID(ctx, userID)
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	// モック Turnstile クライアント（常に成功）
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, nil, mockClient, createPasswordResetTokenUC)

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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	// モック Turnstile クライアント（常に失敗）
	mockClient := &mockTurnstileClient{shouldSucceed: false}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, nil, mockClient, createPasswordResetTokenUC)

	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("cf-turnstile-response", "invalid-token")

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("Turnstile検証が失敗したはずですが、ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusUnprocessableEntity)
	}
}

// TestCreate_TurnstileVerification_MissingToken はTurnstileトークンが欠落している場合のテストです
func TestCreate_TurnstileVerification_MissingToken(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	// モック Turnstile クライアント（常に失敗）
	mockClient := &mockTurnstileClient{shouldSucceed: false}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, nil, mockClient, createPasswordResetTokenUC)

	form := url.Values{}
	form.Add("email", "test@example.com")
	// cf-turnstile-response を含めない

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("Turnstileトークンが欠落している場合、ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusUnprocessableEntity)
	}
}
