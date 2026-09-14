package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

func TestCompleteSignUpUsecase_Execute(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	rdb := testutil.SetupTestRedis(t)
	ctx := context.Background()

	queries := testutil.NewQueriesWithTx(db, tx)
	userRepo := repository.NewUserRepository(queries).WithTx(tx)
	profileRepo := repository.NewProfileRepository(queries).WithTx(tx)
	settingRepo := repository.NewSettingRepository(queries).WithTx(tx)
	emailNotificationRepo := repository.NewEmailNotificationRepository(queries).WithTx(tx)

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
					t.Fatalf("Redisへのトークン保存エラー = %v", err)
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
				ve := model.AsValidationError(err)
				if ve == nil {
					t.Errorf("エラーの型 = %v、期待値 = *model.ValidationError", err)
					return
				}
				if !ve.HasFieldError("token") {
					t.Errorf("tokenフィールドのエラー = %+v、期待値 = バリデーションエラー", ve)
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
					t.Fatalf("Redisへのトークン保存エラー = %v", err)
				}
				return token
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				ve := model.AsValidationError(err)
				if ve == nil {
					t.Errorf("エラーの型 = %v、期待値 = *model.ValidationError", err)
					return
				}
				if !ve.HasFieldError("username") {
					t.Errorf("usernameフィールドのエラー = %+v、期待値 = バリデーションエラー", ve)
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
			v := validator.NewSignUpUsernameCreateValidator()
			uc := NewCompleteSignUpUsecase(db, userRepo, profileRepo, settingRepo, emailNotificationRepo, repository.NewSessionRepository(queries), rdb, v)
			result, err := uc.Execute(ctx, CompleteSignUpInput{
				Token:    tt.token,
				Username: tt.username,
				Locale:   tt.locale,
			})

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute()のエラー = %v、期待値 = %v", err, tt.wantErr)
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
				t.Fatal("resultがnilだった")
			}

			// ユーザー情報の検証
			if result.User.Username != tt.username {
				t.Errorf("User.Username = %v、期待値 = %v", result.User.Username, tt.username)
			}

			if result.User.Locale != tt.locale {
				t.Errorf("User.Locale = %v、期待値 = %v", result.User.Locale, tt.locale)
			}

			// セッションIDの検証
			if result.SessionPublicID == "" {
				t.Error("SessionPublicIDが空だった")
			}

			// プロフィールが作成されているか確認
			profile, err := queries.GetProfileByUserID(ctx, int64(result.User.ID))
			if err != nil {
				t.Fatalf("プロフィールの取得エラー = %v", err)
			}
			if profile.Name != tt.username {
				t.Errorf("Profile.Name = %v、期待値 = %v", profile.Name, tt.username)
			}
			if profile.Description != "" {
				t.Errorf("Profile.Description = %v、期待値 = 空文字列", profile.Description)
			}

			// 設定が作成されているか確認
			setting, err := queries.GetSettingByUserID(ctx, int64(result.User.ID))
			if err != nil {
				t.Fatalf("設定の取得エラー = %v", err)
			}
			if !setting.PrivacyPolicyAgreed {
				t.Error("Setting.PrivacyPolicyAgreed = false、期待値 = true")
			}

			// メール通知設定が作成されているか確認
			emailNotification, err := queries.GetEmailNotificationByUserID(ctx, int64(result.User.ID))
			if err != nil {
				t.Fatalf("メール通知設定の取得エラー = %v", err)
			}
			if emailNotification.UnsubscriptionKey == "" {
				t.Error("EmailNotification.UnsubscriptionKeyが空だった")
			}

			// トークンが削除されているか確認
			tokenKey := fmt.Sprintf("sign_up_token:%s", tt.token)
			_, err = rdb.Get(ctx, tokenKey).Result()
			if err == nil {
				t.Error("Redisからトークンが削除されなかった")
			}
		})
	}
}

func TestCompleteSignUpUsecase_Execute_Integration(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	rdb := testutil.SetupTestRedis(t)
	ctx := context.Background()

	queries := testutil.NewQueriesWithTx(db, tx)
	userRepo := repository.NewUserRepository(queries).WithTx(tx)
	profileRepo := repository.NewProfileRepository(queries).WithTx(tx)
	settingRepo := repository.NewSettingRepository(queries).WithTx(tx)
	emailNotificationRepo := repository.NewEmailNotificationRepository(queries).WithTx(tx)

	// 一時トークンをRedisに保存
	token := "valid-token-integration"
	tokenKey := fmt.Sprintf("sign_up_token:%s", token)
	err := rdb.Set(ctx, tokenKey, "noredis@example.com", 15*time.Minute).Err()
	if err != nil {
		t.Fatalf("Redisへのトークン保存エラー = %v", err)
	}

	// ユースケースを作成 (Redisあり)
	v := validator.NewSignUpUsernameCreateValidator()
	uc := NewCompleteSignUpUsecase(db, userRepo, profileRepo, settingRepo, emailNotificationRepo, repository.NewSessionRepository(queries), rdb, v)

	// ユーザー登録を実行
	result, err := uc.Execute(ctx, CompleteSignUpInput{
		Token:    token,
		Username: "testuser_noredis",
		Locale:   "ja",
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	// ユーザー情報の検証
	if result.User.Username != "testuser_noredis" {
		t.Errorf("User.Username = %v、期待値 = testuser_noredis", result.User.Username)
	}

	if result.User.Email != "noredis@example.com" {
		t.Errorf("User.Email = %v、期待値 = noredis@example.com", result.User.Email)
	}

	// プロフィールが作成されているか確認
	profile, err := queries.GetProfileByUserID(ctx, int64(result.User.ID))
	if err != nil {
		t.Fatalf("プロフィールの取得エラー = %v", err)
	}
	if profile.Name != "testuser_noredis" {
		t.Errorf("Profile.Name = %v、期待値 = testuser_noredis", profile.Name)
	}

	// 設定が作成されているか確認
	setting, err := queries.GetSettingByUserID(ctx, int64(result.User.ID))
	if err != nil {
		t.Fatalf("設定の取得エラー = %v", err)
	}
	if !setting.PrivacyPolicyAgreed {
		t.Error("Setting.PrivacyPolicyAgreed = false、期待値 = true")
	}

	// メール通知設定が作成されているか確認
	emailNotification, err := queries.GetEmailNotificationByUserID(ctx, int64(result.User.ID))
	if err != nil {
		t.Fatalf("メール通知設定の取得エラー = %v", err)
	}
	if emailNotification.UnsubscriptionKey == "" {
		t.Error("EmailNotification.UnsubscriptionKeyが空だった")
	}
}
