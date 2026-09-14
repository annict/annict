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

// mockTurnstileClientはテスト用のモックTurnstileクライアントです
type mockTurnstileClient struct {
	shouldSucceed bool
}

func (m *mockTurnstileClient) Verify(ctx context.Context, token string) (bool, error) {
	return m.shouldSucceed, nil
}

// TestCreate_RateLimiting_IPはIPアドレス単位のRate Limitingをテストします
func TestCreate_RateLimiting_IP(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	rdb := testutil.SetupTestRedis(t)
	queries := query.New(db).WithTx(tx)
	limiter := ratelimit.NewLimiter(rdb)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みエラー = %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	// Rate Limitingのテストのため、明示的に有効化
	cfg.DisableRateLimit = false

	// モックTurnstileクライアント (常に成功)
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, limiter, mockClient, createPasswordResetTokenUC)

	// 共有Redis DBに対する並列実行で他テストと衝突しないよう、
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
			t.Errorf("%d回目のステータスコード = %d、期待値 = OK", i+1, rr.Code)
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
		t.Errorf("6回目の試行のステータスコード = %d、期待値 = 422 (IPによるレート制限)", rr.Code)
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
		t.Errorf("別のIPからのリクエストのステータスコード = %d、期待値 = OK", rr.Code)
	}
}

// TestCreate_RateLimiting_Emailはメールアドレス単位のRate Limitingをテストします
func TestCreate_RateLimiting_Email(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	rdb := testutil.SetupTestRedis(t)
	queries := query.New(db).WithTx(tx)
	limiter := ratelimit.NewLimiter(rdb)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みエラー = %v", err)
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	// Rate Limitingのテストのため、明示的に有効化
	cfg.DisableRateLimit = false

	// モックTurnstileクライアント (常に成功)
	mockClient := &mockTurnstileClient{shouldSucceed: true}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, limiter, mockClient, createPasswordResetTokenUC)

	// 共有Redis DBに対する並列実行で他テストと衝突しないよう、
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
			t.Errorf("%d回目のステータスコード = %d、期待値 = OK", i+1, rr.Code)
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
		t.Errorf("4回目の試行のステータスコード = %d、期待値 = 422 (メールによるレート制限)", rr.Code)
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
		t.Errorf("別のメールでのリクエストのステータスコード = %d、期待値 = OK", rr.Code)
	}
}

// TestPasswordResetSentPage_UXMessagesはメール送信完了ページのUXメッセージをテストします
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
	// モックTurnstileクライアント (常に成功)
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
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
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

// TestPasswordResetFlow_Integrationはパスワードリセット申請のフローをテストします
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
	// モックTurnstileクライアント (常に成功)
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

// TestCreate_TurnstileVerification_SuccessはTurnstile検証が成功した場合のテストです
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
	// モックTurnstileクライアント (常に成功)
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
		t.Errorf("Turnstile検証が成功した場合のステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
}

// TestCreate_TurnstileVerification_FailedはTurnstile検証が失敗した場合のテストです
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
	// モックTurnstileクライアント (常に失敗)
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
		t.Errorf("Turnstile検証が失敗した場合のステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

// TestCreate_TurnstileVerification_MissingTokenはTurnstileトークンが欠落している場合のテストです
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
	// モックTurnstileクライアント (常に失敗)
	mockClient := &mockTurnstileClient{shouldSucceed: false}
	v := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)
	handler := NewHandler(cfg, sessionManager, nil, mockClient, createPasswordResetTokenUC)

	form := url.Values{}
	form.Add("email", "test@example.com")
	// cf-turnstile-responseを含めない

	req := httptest.NewRequest("POST", "/password/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("Turnstileトークンが欠落している場合のステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
}
