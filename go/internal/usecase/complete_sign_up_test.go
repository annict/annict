package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/annict/annict/internal/testutil"
)

func TestCompleteSignUpUsecase_Execute(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	rdb := testutil.SetupTestRedis(t)
	ctx := context.Background()

	queries := testutil.NewQueriesWithTx(db, tx)

	tests := []struct {
		name      string
		token     string
		username  string
		locale    string
		setupFunc func(t *testing.T) string // トークンを返す
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:     "正常系",
			username: "testuser",
			locale:   "ja",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				// 一時トークンをRedisに保存
				token := "valid-token-123"
				tokenKey := fmt.Sprintf("sign_up_token:%s", token)
				err := rdb.Set(ctx, tokenKey, "test@example.com", 15*time.Minute).Err()
				if err != nil {
					t.Fatalf("failed to set token in redis: %v", err)
				}
				return token
			},
			wantErr: false,
		},
		{
			name:     "無効なトークン",
			token:    "invalid-token",
			username: "testuser2",
			locale:   "ja",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				return "invalid-token"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				if !IsTokenInvalidError(err) {
					t.Errorf("expected TokenInvalidError, got %v", err)
				}
			},
		},
		{
			name:     "ユーザー名が既に存在",
			username: "existinguser",
			locale:   "ja",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				// 既存ユーザーを作成
				_ = testutil.NewUserBuilder(t, tx).
					WithUsername("existinguser").
					WithEmail("existing@example.com").
					Build()

				// 一時トークンをRedisに保存
				token := "valid-token-456"
				tokenKey := fmt.Sprintf("sign_up_token:%s", token)
				err := rdb.Set(ctx, tokenKey, "new@example.com", 15*time.Minute).Err()
				if err != nil {
					t.Fatalf("failed to set token in redis: %v", err)
				}
				return token
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				if !IsUsernameAlreadyExistsError(err) {
					t.Errorf("expected UsernameAlreadyExistsError, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// セットアップ関数を実行してトークンを取得
			token := tt.setupFunc(t)
			if tt.token == "" {
				tt.token = token
			}

			// ユースケースを実行
			uc := NewCompleteSignUpUsecase(db, queries, rdb)
			result, err := uc.Execute(ctx, tt.token, tt.username, tt.locale)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
				return
			}

			// 正常系の検証
			if result == nil {
				t.Fatal("result should not be nil")
			}

			// ユーザー情報の検証
			if result.User.Username != tt.username {
				t.Errorf("User.Username = %v, want %v", result.User.Username, tt.username)
			}

			if result.User.Locale != tt.locale {
				t.Errorf("User.Locale = %v, want %v", result.User.Locale, tt.locale)
			}

			// セッションIDの検証
			if result.SessionPublicID == "" {
				t.Error("SessionPublicID should not be empty")
			}

			// プロフィールが作成されているか確認
			profile, err := queries.GetProfileByUserID(ctx, result.User.ID)
			if err != nil {
				t.Fatalf("failed to get profile: %v", err)
			}
			if profile.Name != tt.username {
				t.Errorf("Profile.Name = %v, want %v", profile.Name, tt.username)
			}
			if profile.Description != "" {
				t.Errorf("Profile.Description = %v, want empty string", profile.Description)
			}

			// 設定が作成されているか確認
			setting, err := queries.GetSettingByUserID(ctx, result.User.ID)
			if err != nil {
				t.Fatalf("failed to get setting: %v", err)
			}
			if !setting.PrivacyPolicyAgreed {
				t.Error("Setting.PrivacyPolicyAgreed should be true")
			}

			// メール通知設定が作成されているか確認
			emailNotification, err := queries.GetEmailNotificationByUserID(ctx, result.User.ID)
			if err != nil {
				t.Fatalf("failed to get email notification: %v", err)
			}
			if emailNotification.UnsubscriptionKey == "" {
				t.Error("EmailNotification.UnsubscriptionKey should not be empty")
			}

			// トークンが削除されているか確認
			tokenKey := fmt.Sprintf("sign_up_token:%s", tt.token)
			_, err = rdb.Get(ctx, tokenKey).Result()
			if err == nil {
				t.Error("token should be deleted from redis")
			}
		})
	}
}

func TestCompleteSignUpUsecase_Execute_Integration(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	rdb := testutil.SetupTestRedis(t)
	ctx := context.Background()

	queries := testutil.NewQueriesWithTx(db, tx)

	// 一時トークンをRedisに保存
	token := "valid-token-integration"
	tokenKey := fmt.Sprintf("sign_up_token:%s", token)
	err := rdb.Set(ctx, tokenKey, "noredis@example.com", 15*time.Minute).Err()
	if err != nil {
		t.Fatalf("failed to set token in redis: %v", err)
	}

	// ユースケースを作成（Redisあり）
	uc := NewCompleteSignUpUsecase(db, queries, rdb)

	// ユーザー登録を実行
	result, err := uc.Execute(ctx, token, "testuser_noredis", "ja")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// ユーザー情報の検証
	if result.User.Username != "testuser_noredis" {
		t.Errorf("User.Username = %v, want testuser_noredis", result.User.Username)
	}

	if result.User.Email != "noredis@example.com" {
		t.Errorf("User.Email = %v, want noredis@example.com", result.User.Email)
	}

	// プロフィールが作成されているか確認
	profile, err := queries.GetProfileByUserID(ctx, result.User.ID)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}
	if profile.Name != "testuser_noredis" {
		t.Errorf("Profile.Name = %v, want testuser_noredis", profile.Name)
	}

	// 設定が作成されているか確認
	setting, err := queries.GetSettingByUserID(ctx, result.User.ID)
	if err != nil {
		t.Fatalf("failed to get setting: %v", err)
	}
	if !setting.PrivacyPolicyAgreed {
		t.Error("Setting.PrivacyPolicyAgreed should be true")
	}

	// メール通知設定が作成されているか確認
	emailNotification, err := queries.GetEmailNotificationByUserID(ctx, result.User.ID)
	if err != nil {
		t.Fatalf("failed to get email notification: %v", err)
	}
	if emailNotification.UnsubscriptionKey == "" {
		t.Error("EmailNotification.UnsubscriptionKey should not be empty")
	}
}
