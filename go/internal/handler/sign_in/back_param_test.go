package sign_in

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/turnstile"
	"github.com/annict/annict/internal/usecase"
)

// setupTestSessionManager はテスト用のセッションマネージャーを作成します
func setupTestSessionManager(t *testing.T, queries *query.Queries) *session.Manager {
	t.Helper()
	cfg := &config.Config{}
	sessionRepo := repository.NewSessionRepository(queries)
	return session.NewManager(sessionRepo, cfg)
}

// TestNew_WithBackParameter はbackパラメータがテンプレートに渡されることを確認します
func TestNew_WithBackParameter(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テスト用の設定
	cfg := &config.Config{}

	// セッションマネージャー
	sessionMgr := setupTestSessionManager(t, queries)

	// ハンドラーを作成
	h := NewHandler(cfg, sessionMgr, nil, nil, nil)

	tests := []struct {
		name        string
		backURL     string
		wantInBody  string
		description string
	}{
		{
			name:        "backパラメータあり",
			backURL:     "/oauth/authorize?client_id=xxx",
			wantInBody:  `name="back" value="/oauth/authorize?client_id=xxx"`,
			description: "backパラメータがhiddenフィールドに含まれる",
		},
		{
			name:        "backパラメータなし",
			backURL:     "",
			wantInBody:  `name="back" value=""`,
			description: "backパラメータが空でもhiddenフィールドが存在する",
		},
		{
			name:        "日本語パスのbackパラメータ",
			backURL:     "/users/テスト",
			wantInBody:  `name="back" value="/users/テスト"`,
			description: "日本語パスもそのまま渡される",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// HTTPリクエストを作成
			targetURL := "/sign_in"
			if tt.backURL != "" {
				targetURL = "/sign_in?back=" + url.QueryEscape(tt.backURL)
			}
			req := httptest.NewRequest(http.MethodGet, targetURL, nil)
			rr := httptest.NewRecorder()

			// ハンドラーを実行
			h.New(rr, req)

			// レスポンスを確認
			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコード: got %d, want %d", rr.Code, http.StatusOK)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tt.wantInBody) {
				t.Errorf("%s: レスポンスに期待する文字列が含まれていません\nwant: %s", tt.description, tt.wantInBody)
			}
		})
	}
}

// TestCreate_BackParameterRedirect はbackパラメータがリダイレクトURLに含まれることを確認します
func TestCreate_BackParameterRedirect(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// パスワードありのユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("back_param_test_pw").
		WithEmail("back_param_test_pw@example.com").
		WithEncryptedPassword("encrypted_test_password"). // パスワードあり
		Build()

	// ユーザー情報を取得してemailを確認
	user, err := queries.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	// テスト用の設定
	cfg := &config.Config{}

	// セッションマネージャー
	sessionMgr := setupTestSessionManager(t, queries)

	// ユーザーリポジトリ
	userRepo := repository.NewUserRepository(queries)

	// Turnstileクライアント（テスト用：SecretKeyが空なので常に検証成功）
	turnstileClient := turnstile.NewClient("", "")

	// ハンドラーを作成
	h := NewHandler(cfg, sessionMgr, userRepo, nil, turnstileClient)

	tests := []struct {
		name             string
		backURL          string
		wantRedirectPath string
		description      string
	}{
		{
			name:             "backパラメータあり - パスワードログインへリダイレクト",
			backURL:          "/oauth/authorize?client_id=xxx",
			wantRedirectPath: "/sign_in/password?back=%2Foauth%2Fauthorize%3Fclient_id%3Dxxx",
			description:      "パスワードログインページへのリダイレクトにbackパラメータが含まれる",
		},
		{
			name:             "backパラメータなし",
			backURL:          "",
			wantRedirectPath: "/sign_in/password",
			description:      "backパラメータがない場合はシンプルなリダイレクト",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// フォームデータを作成
			formData := url.Values{}
			formData.Set("email", user.Email)
			formData.Set("csrf_token", "test-token")
			formData.Set("cf-turnstile-response", "test-response")
			if tt.backURL != "" {
				formData.Set("back", tt.backURL)
			}

			// HTTPリクエストを作成
			req := httptest.NewRequest(http.MethodPost, "/sign_in", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			// ハンドラーを実行
			h.Create(rr, req)

			// リダイレクトを確認
			if rr.Code != http.StatusSeeOther {
				t.Errorf("ステータスコード: got %d, want %d", rr.Code, http.StatusSeeOther)
			}

			location := rr.Header().Get("Location")
			if location != tt.wantRedirectPath {
				t.Errorf("%s: リダイレクト先が異なります\ngot: %s\nwant: %s", tt.description, location, tt.wantRedirectPath)
			}
		})
	}
}

// TestCreate_BackParameterRedirectToCode はパスワードなしユーザーの場合にbackパラメータがコード入力ページへのリダイレクトに含まれることを確認します
func TestCreate_BackParameterRedirectToCode(t *testing.T) {
	// パスワードなしユーザーの場合、トランザクションをコミットするため並列実行を無効化

	// テスト用DBをセットアップ
	db, tx := testutil.SetupTestDB(t)

	// ユニークなメールアドレスとユーザー名を生成
	uniqueID := time.Now().UnixNano()
	testEmail := fmt.Sprintf("back_param_test_code_%d@example.com", uniqueID)
	testUsername := fmt.Sprintf("back_param_test_code_%d", uniqueID)

	// パスワードなしのユーザーを作成
	testutil.NewUserBuilder(t, tx).
		WithUsername(testUsername).
		WithEmail(testEmail).
		WithEncryptedPassword(""). // パスワードなし
		Build()

	// パスワードなしユーザーの場合、ユースケースが新しいトランザクションを開始するため、
	// ユーザーをDBに永続化する必要があります
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗しました: %v", err)
	}

	// 新しいトランザクションを開始
	newTx, err := db.Begin()
	if err != nil {
		t.Fatalf("新しいトランザクションの開始に失敗しました: %v", err)
	}
	t.Cleanup(func() {
		_ = newTx.Rollback()
	})

	// sqlcリポジトリを作成
	queries := query.New(db).WithTx(newTx)

	// テスト用の設定
	cfg := &config.Config{}

	// セッションマネージャー
	sessionMgr := setupTestSessionManager(t, queries)

	// ユーザーリポジトリ
	userRepo := repository.NewUserRepository(queries)

	// Turnstileクライアント（テスト用：SecretKeyが空なので常に検証成功）
	turnstileClient := turnstile.NewClient("", "")

	// SendSignInCodeユースケース
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil) // メール送信はnilでOK（テスト用）

	// ハンドラーを作成
	h := NewHandler(cfg, sessionMgr, userRepo, sendSignInCodeUC, turnstileClient)

	// フォームデータを作成
	backURL := "/oauth/authorize?client_id=xxx"
	formData := url.Values{}
	formData.Set("email", testEmail)
	formData.Set("csrf_token", "test-token")
	formData.Set("cf-turnstile-response", "test-response")
	formData.Set("back", backURL)

	// HTTPリクエストを作成
	req := httptest.NewRequest(http.MethodPost, "/sign_in", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	h.Create(rr, req)

	// リダイレクトを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコード: got %d, want %d", rr.Code, http.StatusSeeOther)
	}

	wantRedirectPath := "/sign_in/code?back=%2Foauth%2Fauthorize%3Fclient_id%3Dxxx"
	location := rr.Header().Get("Location")
	if location != wantRedirectPath {
		t.Errorf("リダイレクト先が異なります\ngot: %s\nwant: %s", location, wantRedirectPath)
	}
}

// TestCreate_ValidationErrorPreservesBackParameter はバリデーションエラー時にbackパラメータが保持されることを確認します
func TestCreate_ValidationErrorPreservesBackParameter(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テスト用の設定
	cfg := &config.Config{}

	// セッションマネージャー
	sessionMgr := setupTestSessionManager(t, queries)

	// Turnstileクライアント（テスト用：SecretKeyが空なので常に検証成功）
	turnstileClient := turnstile.NewClient("", "")

	// ハンドラーを作成
	h := NewHandler(cfg, sessionMgr, nil, nil, turnstileClient)

	// フォームデータを作成（メールアドレスが空）
	backURL := "/oauth/authorize?client_id=xxx"
	formData := url.Values{}
	formData.Set("email", "") // 空のメールアドレス（バリデーションエラー）
	formData.Set("csrf_token", "test-token")
	formData.Set("cf-turnstile-response", "test-response")
	formData.Set("back", backURL)

	// HTTPリクエストを作成
	req := httptest.NewRequest(http.MethodPost, "/sign_in", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	h.Create(rr, req)

	// リダイレクトを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコード: got %d, want %d", rr.Code, http.StatusSeeOther)
	}

	// backパラメータが保持されていることを確認
	wantRedirectPath := "/sign_in?back=%2Foauth%2Fauthorize%3Fclient_id%3Dxxx"
	location := rr.Header().Get("Location")
	if location != wantRedirectPath {
		t.Errorf("リダイレクト先が異なります\ngot: %s\nwant: %s", location, wantRedirectPath)
	}
}
