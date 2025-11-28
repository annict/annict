package seed

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
)

// TestCreateUserUsecase_ExecuteBatch はExecuteBatchメソッドのテスト
func TestCreateUserUsecase_ExecuteBatch(t *testing.T) {
	// テストDBをセットアップ（トランザクションは各サブテストで作成）
	db, _ := testutil.SetupTestDB(t)

	// Usecaseを作成
	queries := query.New(db)
	uc := NewCreateUserUsecase(db, queries)

	// ユニークなIDを生成（テスト間の衝突を避ける）
	uniqueID := fmt.Sprintf("%d", time.Now().UnixNano())

	// テストケース
	tests := []struct {
		name      string
		users     []CreateUserParams
		wantCount int
		wantErr   bool
	}{
		{
			name: "正常系: 3人のユーザーを作成",
			users: []CreateUserParams{
				{
					Username: fmt.Sprintf("test_user_1_%s", uniqueID),
					Email:    fmt.Sprintf("test1_%s@example.com", uniqueID),
					Password: "password123",
					Locale:   "ja",
				},
				{
					Username: fmt.Sprintf("test_user_2_%s", uniqueID),
					Email:    fmt.Sprintf("test2_%s@example.com", uniqueID),
					Password: "password456",
					Locale:   "en",
				},
				{
					Username: fmt.Sprintf("test_user_3_%s", uniqueID),
					Email:    fmt.Sprintf("test3_%s@example.com", uniqueID),
					Password: "password789",
					Locale:   "ja",
				},
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "正常系: 1人のユーザーを作成",
			users: []CreateUserParams{
				{
					Username: fmt.Sprintf("single_user_%s", uniqueID),
					Email:    fmt.Sprintf("single_%s@example.com", uniqueID),
					Password: "password_single",
					Locale:   "ja",
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "正常系: ロケール未指定（デフォルトja）",
			users: []CreateUserParams{
				{
					Username: fmt.Sprintf("no_locale_user_%s", uniqueID),
					Email:    fmt.Sprintf("nolocale_%s@example.com", uniqueID),
					Password: "password_no_locale",
					Locale:   "", // 空文字の場合はデフォルト"ja"
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各サブテストで新しいトランザクションを作成
			_, tx := testutil.SetupTestDB(t)
			defer tx.Rollback()

			ctx := context.Background()

			// ExecuteBatchWithTxを実行（テスト用トランザクションを使用）
			results, err := uc.ExecuteBatchWithTx(ctx, tx, tt.users, nil)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 結果の数をチェック
			if len(results) != tt.wantCount {
				t.Errorf("ExecuteBatch() got %d results, want %d", len(results), tt.wantCount)
				return
			}

			// 各ユーザーが正しく作成されたか検証
			for i, result := range results {
				// UserIDとProfileIDが取得できていることを確認
				if result.UserID == 0 {
					t.Errorf("result[%d].UserID is 0", i)
				}
				if result.ProfileID == 0 {
					t.Errorf("result[%d].ProfileID is 0", i)
				}

				// usersテーブルにレコードが作成されたか確認
				var username, email, locale string
				err := tx.QueryRow("SELECT username, email, locale FROM users WHERE id = $1", result.UserID).
					Scan(&username, &email, &locale)
				if err != nil {
					t.Errorf("ユーザーレコードの取得に失敗: %v", err)
					continue
				}

				// ユーザー名とメールアドレスが正しいか確認
				if username != tt.users[i].Username {
					t.Errorf("username = %v, want %v", username, tt.users[i].Username)
				}
				if email != tt.users[i].Email {
					t.Errorf("email = %v, want %v", email, tt.users[i].Email)
				}

				// ロケールが正しいか確認（空文字の場合は"ja"がデフォルト）
				expectedLocale := tt.users[i].Locale
				if expectedLocale == "" {
					expectedLocale = "ja"
				}
				if locale != expectedLocale {
					t.Errorf("locale = %v, want %v", locale, expectedLocale)
				}

				// profilesテーブルにレコードが作成されたか確認
				var profileUserID int64
				err = tx.QueryRow("SELECT user_id FROM profiles WHERE id = $1", result.ProfileID).
					Scan(&profileUserID)
				if err != nil {
					t.Errorf("プロフィールレコードの取得に失敗: %v", err)
					continue
				}

				// プロフィールのuser_idが正しいか確認
				if profileUserID != result.UserID {
					t.Errorf("profile.user_id = %v, want %v", profileUserID, result.UserID)
				}

				// settingsテーブルにレコードが作成されたか確認
				var settingUserID int64
				var privacyPolicyAgreed bool
				err = tx.QueryRow("SELECT user_id, privacy_policy_agreed FROM settings WHERE user_id = $1", result.UserID).
					Scan(&settingUserID, &privacyPolicyAgreed)
				if err != nil {
					t.Errorf("設定レコードの取得に失敗: %v", err)
					continue
				}

				// 設定のuser_idが正しいか確認
				if settingUserID != result.UserID {
					t.Errorf("setting.user_id = %v, want %v", settingUserID, result.UserID)
				}

				// privacy_policy_agreedがtrueであることを確認
				if !privacyPolicyAgreed {
					t.Errorf("privacy_policy_agreed = %v, want true", privacyPolicyAgreed)
				}

				// email_notificationsテーブルにレコードが作成されたか確認
				var emailNotificationUserID int64
				var unsubscriptionKey string
				err = tx.QueryRow("SELECT user_id, unsubscription_key FROM email_notifications WHERE user_id = $1", result.UserID).
					Scan(&emailNotificationUserID, &unsubscriptionKey)
				if err != nil {
					t.Errorf("メール通知設定レコードの取得に失敗: %v", err)
					continue
				}

				// メール通知設定のuser_idが正しいか確認
				if emailNotificationUserID != result.UserID {
					t.Errorf("email_notification.user_id = %v, want %v", emailNotificationUserID, result.UserID)
				}

				// unsubscription_keyが空でないことを確認
				if unsubscriptionKey == "" {
					t.Error("unsubscription_key should not be empty")
				}
			}
		})
	}
}

// TestCreateUserUsecase_PasswordHashing はパスワードハッシュ化のテスト
func TestCreateUserUsecase_PasswordHashing(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)
	defer tx.Rollback()

	// Usecaseを作成
	queries := query.New(db)
	uc := NewCreateUserUsecase(db, queries)

	ctx := context.Background()

	// ユーザーを作成
	users := []CreateUserParams{
		{
			Username: "password_test_user",
			Email:    "password@example.com",
			Password: "my_secret_password",
			Locale:   "ja",
		},
	}

	results, err := uc.ExecuteBatchWithTx(ctx, tx, users, nil)
	if err != nil {
		t.Fatalf("ExecuteBatch() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("ExecuteBatch() returned %d results, want 1", len(results))
	}

	// DBからencrypted_passwordを取得
	var encryptedPassword string
	err = tx.QueryRow("SELECT encrypted_password FROM users WHERE id = $1", results[0].UserID).
		Scan(&encryptedPassword)
	if err != nil {
		t.Fatalf("encrypted_passwordの取得に失敗: %v", err)
	}

	// bcryptでハッシュ化されていることを確認
	if encryptedPassword == "" {
		t.Error("encrypted_passwordが空")
	}

	// 平文パスワードと一致しないことを確認
	if encryptedPassword == "my_secret_password" {
		t.Error("パスワードが平文で保存されている")
	}

	// bcryptでパスワードが検証できることを確認
	err = auth.CheckPassword(encryptedPassword, "my_secret_password")
	if err != nil {
		t.Errorf("パスワード検証に失敗: %v", err)
	}

	// 間違ったパスワードで検証が失敗することを確認
	err = auth.CheckPassword(encryptedPassword, "wrong_password")
	if err == nil {
		t.Error("間違ったパスワードで検証が成功してしまった")
	}
}

// TestCreateUserUsecase_LargeBatch は大量のユーザー作成のテスト
func TestCreateUserUsecase_LargeBatch(t *testing.T) {
	// -short フラグが指定されている場合はスキップ（CI用）
	if testing.Short() {
		t.Skip("長時間テストのため -short フラグでスキップします")
	}

	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)
	defer tx.Rollback()

	// Usecaseを作成
	queries := query.New(db)
	uc := NewCreateUserUsecase(db, queries)

	ctx := context.Background()

	// 2500人のユーザーを作成（バッチサイズ1000を超えるケース）
	userCount := 2500
	users := make([]CreateUserParams, userCount)
	for i := 0; i < userCount; i++ {
		users[i] = CreateUserParams{
			Username: fmt.Sprintf("bulk_user_%d", i+1),
			Email:    fmt.Sprintf("bulk_%d@example.com", i+1),
			Password: "password",
			Locale:   "ja",
		}
	}

	// ExecuteBatchWithTxを実行（テスト用トランザクションを使用）
	results, err := uc.ExecuteBatchWithTx(ctx, tx, users, nil)
	if err != nil {
		t.Fatalf("ExecuteBatch() error = %v", err)
	}

	// 結果の数をチェック
	if len(results) != userCount {
		t.Errorf("ExecuteBatch() got %d results, want %d", len(results), userCount)
	}

	// 最初と最後のユーザーを検証
	if len(results) > 0 {
		if results[0].UserID == 0 {
			t.Error("results[0].UserID is 0")
		}
		if results[len(results)-1].UserID == 0 {
			t.Error("results[last].UserID is 0")
		}
	}
}
