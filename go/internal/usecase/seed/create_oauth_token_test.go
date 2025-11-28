package seed

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/annict/annict/internal/testutil"
)

// TestCreateOAuthTokenUsecase_Execute は Execute メソッドのテスト
func TestCreateOAuthTokenUsecase_Execute(t *testing.T) {
	// テストDBをセットアップ
	db, _ := testutil.SetupTestDB(t)

	// Usecaseを作成
	uc := NewCreateOAuthTokenUsecase(db, nil)

	// テストケース
	tests := []struct {
		name           string
		setupUsers     func(t *testing.T, tx *sql.Tx) []int64 // テストユーザーを作成する関数
		params         CreateOAuthTokenParams
		wantTokenCount int
		wantErr        bool
		validateFunc   func(t *testing.T, tx *sql.Tx, result *CreateOAuthTokenResult) // カスタム検証関数
	}{
		{
			name: "正常系: アプリケーション1件 + トークン3件を作成",
			setupUsers: func(t *testing.T, tx *sql.Tx) []int64 {
				return createTestUsersForOAuth(t, tx, 3)
			},
			params: CreateOAuthTokenParams{
				ApplicationName: "Test App",
				ApplicationUID:  "test_app_uid_123",
				RedirectURI:     "https://example.com/callback",
				Scopes:          "",
				TokenCount:      3,
			},
			wantTokenCount: 3,
			wantErr:        false,
			validateFunc: func(t *testing.T, tx *sql.Tx, result *CreateOAuthTokenResult) {
				// アプリケーションが作成されたことを確認
				assertOAuthApplicationExists(t, tx, result.ApplicationID, "Test App", "test_app_uid_123")
				// トークンが作成されたことを確認
				for _, tokenID := range result.TokenIDs {
					assertOAuthAccessTokenExists(t, tx, tokenID, result.ApplicationID)
				}
			},
		},
		{
			name: "正常系: デフォルト値でアプリケーション作成",
			setupUsers: func(t *testing.T, tx *sql.Tx) []int64 {
				return createTestUsersForOAuth(t, tx, 2)
			},
			params: CreateOAuthTokenParams{
				ApplicationName: "", // デフォルト値を使用
				ApplicationUID:  "", // デフォルト値（ランダム生成）
				RedirectURI:     "", // デフォルト値（urn:ietf:wg:oauth:2.0:oob）
				Scopes:          "",
				TokenCount:      2,
			},
			wantTokenCount: 2,
			wantErr:        false,
			validateFunc: func(t *testing.T, tx *sql.Tx, result *CreateOAuthTokenResult) {
				// アプリケーションがデフォルト名で作成されたことを確認
				var appName string
				err := tx.QueryRow("SELECT name FROM oauth_applications WHERE id = $1", result.ApplicationID).Scan(&appName)
				if err != nil {
					t.Fatalf("Failed to get application name: %v", err)
				}
				if appName != "Test Application" {
					t.Errorf("Application name = %q, want %q", appName, "Test Application")
				}
			},
		},
		{
			name: "正常系: 大量のトークンを作成（150件）",
			setupUsers: func(t *testing.T, tx *sql.Tx) []int64 {
				return createTestUsersForOAuth(t, tx, 150)
			},
			params: CreateOAuthTokenParams{
				ApplicationName: "Batch Test App",
				ApplicationUID:  "batch_test_uid",
				RedirectURI:     "https://example.com/batch",
				Scopes:          "read write",
				TokenCount:      150,
			},
			wantTokenCount: 150,
			wantErr:        false,
			validateFunc: func(t *testing.T, tx *sql.Tx, result *CreateOAuthTokenResult) {
				// トークン数を確認
				var tokenCount int
				err := tx.QueryRow("SELECT COUNT(*) FROM oauth_access_tokens WHERE application_id = $1", result.ApplicationID).Scan(&tokenCount)
				if err != nil {
					t.Fatalf("Failed to count tokens: %v", err)
				}
				if tokenCount != 150 {
					t.Errorf("Token count = %d, want %d", tokenCount, 150)
				}
			},
		},
		{
			name: "正常系: トークン1件のみ作成",
			setupUsers: func(t *testing.T, tx *sql.Tx) []int64 {
				return createTestUsersForOAuth(t, tx, 1)
			},
			params: CreateOAuthTokenParams{
				ApplicationName: "Single Token App",
				ApplicationUID:  "single_token_uid",
				RedirectURI:     "https://example.com/single",
				Scopes:          "",
				TokenCount:      1,
			},
			wantTokenCount: 1,
			wantErr:        false,
			validateFunc: func(t *testing.T, tx *sql.Tx, result *CreateOAuthTokenResult) {
				assertOAuthApplicationExists(t, tx, result.ApplicationID, "Single Token App", "single_token_uid")
				assertOAuthAccessTokenExists(t, tx, result.TokenIDs[0], result.ApplicationID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各サブテストで新しいトランザクションを作成
			_, tx := testutil.SetupTestDB(t)
			defer tx.Rollback()

			ctx := context.Background()

			// テストユーザーを作成
			userIDs := tt.setupUsers(t, tx)

			// パラメータにユーザーIDを設定
			params := tt.params
			params.UserIDs = userIDs

			// ExecuteWithTxを実行（テスト用トランザクションを使用）
			result, err := uc.ExecuteWithTx(ctx, tx, params, nil)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return // エラーケースの場合は以降の検証をスキップ
			}

			// 結果のチェック
			if result.ApplicationID == 0 {
				t.Error("Execute() returned ApplicationID = 0, want non-zero")
			}

			// トークン数のチェック
			if len(result.TokenIDs) != tt.wantTokenCount {
				t.Errorf("Execute() returned %d tokens, want %d", len(result.TokenIDs), tt.wantTokenCount)
			}

			// すべてのトークンIDが非ゼロであることを確認
			for i, tokenID := range result.TokenIDs {
				if tokenID == 0 {
					t.Errorf("Execute() returned TokenIDs[%d] = 0, want non-zero", i)
				}
			}

			// カスタム検証関数を実行
			if tt.validateFunc != nil {
				tt.validateFunc(t, tx, result)
			}
		})
	}
}

// createTestUsersForOAuth テスト用のユーザーを作成するヘルパー関数（OAuth専用）
func createTestUsersForOAuth(t *testing.T, tx *sql.Tx, count int) []int64 {
	t.Helper()
	userIDs := make([]int64, count)
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("oauth_test_user_%s_%d", t.Name(), i)
		email := fmt.Sprintf("oauth_test_%s_%d@example.com", t.Name(), i)
		userID := testutil.NewUserBuilder(t, tx).
			WithUsername(username).
			WithEmail(email).
			Build()
		userIDs[i] = userID
	}
	return userIDs
}

// assertOAuthApplicationExists OAuth アプリケーションがDBに存在することを検証するヘルパー関数
func assertOAuthApplicationExists(t *testing.T, tx *sql.Tx, applicationID int64, expectedName, expectedUID string) {
	t.Helper()

	query := `SELECT name, uid FROM oauth_applications WHERE id = $1`
	var name, uid string
	err := tx.QueryRow(query, applicationID).Scan(&name, &uid)
	if err != nil {
		t.Fatalf("Failed to get OAuth application (id=%d): %v", applicationID, err)
	}

	if name != expectedName {
		t.Errorf("Application name = %q, want %q", name, expectedName)
	}
	if uid != expectedUID {
		t.Errorf("Application UID = %q, want %q", uid, expectedUID)
	}
}

// assertOAuthAccessTokenExists OAuth アクセストークンがDBに存在することを検証するヘルパー関数
func assertOAuthAccessTokenExists(t *testing.T, tx *sql.Tx, tokenID int64, expectedApplicationID int64) {
	t.Helper()

	query := `SELECT application_id, token FROM oauth_access_tokens WHERE id = $1`
	var applicationID int64
	var token string
	err := tx.QueryRow(query, tokenID).Scan(&applicationID, &token)
	if err != nil {
		t.Fatalf("Failed to get OAuth access token (id=%d): %v", tokenID, err)
	}

	if applicationID != expectedApplicationID {
		t.Errorf("Token application_id = %d, want %d", applicationID, expectedApplicationID)
	}
	if token == "" {
		t.Error("Token is empty, want non-empty")
	}
}
