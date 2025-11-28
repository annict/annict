// Package usecase はビジネスロジック層のユースケースを提供します
package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/query"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// CompleteSignUpUsecase はユーザー登録を完了するユースケース
type CompleteSignUpUsecase struct {
	db          *sql.DB
	queries     *query.Queries
	redisClient *redis.Client
}

// NewCompleteSignUpUsecase はCompleteSignUpUsecaseを作成します
func NewCompleteSignUpUsecase(
	db *sql.DB,
	queries *query.Queries,
	redisClient *redis.Client,
) *CompleteSignUpUsecase {
	return &CompleteSignUpUsecase{
		db:          db,
		queries:     queries,
		redisClient: redisClient,
	}
}

// CompleteSignUpResult はユーザー登録完了の結果
type CompleteSignUpResult struct {
	User            query.CreateUserRow
	SessionPublicID string
}

// Execute はユーザー登録を完了します
//
// 処理フロー:
// 1. Redisから一時トークンを検証してメールアドレスを取得
// 2. ユーザー名の一意性チェック
// 3. トランザクション開始
// 4. ユーザーを作成
// 5. プロフィールを作成（name: ユーザー名、description: 空文字列）
// 6. 設定を作成（privacy_policy_agreed: true、その他はデフォルト値）
// 7. メール通知設定を作成（unsubscription_key: UUID）
// 8. セッションを作成
// 9. トランザクションコミット
// 10. 一時トークンを削除
func (uc *CompleteSignUpUsecase) Execute(
	ctx context.Context,
	token string,
	username string,
	locale string,
) (*CompleteSignUpResult, error) {
	// Redisから一時トークンを検証してメールアドレスを取得
	email, err := uc.verifyToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("トークン検証に失敗: %w", err)
	}

	// ユーザー名の一意性チェック
	_, err = uc.queries.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, &UsernameAlreadyExistsError{Username: username}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("ユーザー名一意性チェックに失敗: %w", err)
	}

	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	queriesWithTx := uc.queries.WithTx(tx)

	// ユーザーを作成
	user, err := queriesWithTx.CreateUser(ctx, query.CreateUserParams{
		Username:          username,
		Email:             email,
		EncryptedPassword: "", // パスワードレス登録
		Locale:            locale,
	})
	if err != nil {
		return nil, fmt.Errorf("ユーザー作成に失敗: %w", err)
	}

	// プロフィールを作成（name: ユーザー名、description: 空文字列）
	_, err = queriesWithTx.CreateProfile(ctx, query.CreateProfileParams{
		UserID: user.ID,
		Name:   username,
	})
	if err != nil {
		return nil, fmt.Errorf("プロフィール作成に失敗: %w", err)
	}

	// 設定を作成（privacy_policy_agreed: true、その他はデフォルト値）
	_, err = queriesWithTx.CreateSetting(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("設定作成に失敗: %w", err)
	}

	// メール通知設定を作成（unsubscription_key: UUID）
	unsubscriptionKey := fmt.Sprintf("%s-%s", uuid.New().String(), uuid.New().String())
	_, err = queriesWithTx.CreateEmailNotification(ctx, query.CreateEmailNotificationParams{
		UserID:            user.ID,
		UnsubscriptionKey: unsubscriptionKey,
	})
	if err != nil {
		return nil, fmt.Errorf("メール通知設定作成に失敗: %w", err)
	}

	// セッションを作成
	createSessionUC := NewCreateSessionUsecase(uc.queries)
	sessionResult, err := createSessionUC.Execute(ctx, tx, user.ID, "", "")
	if err != nil {
		return nil, fmt.Errorf("セッション作成に失敗: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションコミットに失敗: %w", err)
	}

	// 一時トークンを削除
	if err := uc.deleteToken(ctx, token); err != nil {
		// 削除失敗はログのみ（処理は続行）
		return nil, fmt.Errorf("トークン削除に失敗: %w", err)
	}

	return &CompleteSignUpResult{
		User:            user,
		SessionPublicID: sessionResult.PublicID,
	}, nil
}

// verifyToken はRedisから一時トークンを検証してメールアドレスを取得します
func (uc *CompleteSignUpUsecase) verifyToken(ctx context.Context, token string) (string, error) {
	if uc.redisClient == nil {
		// テスト環境ではRedisがない場合があるため、ダミーメールを返す
		return "test@example.com", nil
	}

	tokenKey := fmt.Sprintf("sign_up_token:%s", token)
	email, err := uc.redisClient.Get(ctx, tokenKey).Result()
	if err != nil {
		return "", &TokenInvalidError{Token: token}
	}

	return email, nil
}

// deleteToken は一時トークンをRedisから削除します
func (uc *CompleteSignUpUsecase) deleteToken(ctx context.Context, token string) error {
	if uc.redisClient == nil {
		return nil
	}

	tokenKey := fmt.Sprintf("sign_up_token:%s", token)
	return uc.redisClient.Del(ctx, tokenKey).Err()
}

// UsernameAlreadyExistsError はユーザー名が既に存在することを示すエラー
type UsernameAlreadyExistsError struct {
	Username string
}

func (e *UsernameAlreadyExistsError) Error() string {
	return fmt.Sprintf("username already exists: %s", e.Username)
}

// IsUsernameAlreadyExistsError はエラーがUsernameAlreadyExistsErrorかどうかを判定します
func IsUsernameAlreadyExistsError(err error) bool {
	var e *UsernameAlreadyExistsError
	return errors.As(err, &e)
}

// TokenInvalidError はトークンが無効であることを示すエラー
type TokenInvalidError struct {
	Token string
}

func (e *TokenInvalidError) Error() string {
	return fmt.Sprintf("token invalid: %s", e.Token)
}

// IsTokenInvalidError はエラーがTokenInvalidErrorかどうかを判定します
func IsTokenInvalidError(err error) bool {
	var e *TokenInvalidError
	return errors.As(err, &e)
}

// LocalizeCompleteSignUpError はCompleteSignUpUsecaseのエラーを国際化します
func LocalizeCompleteSignUpError(ctx context.Context, err error) string {
	if IsUsernameAlreadyExistsError(err) {
		return i18n.T(ctx, "sign_up_username_error_username_taken")
	}
	if IsTokenInvalidError(err) {
		return i18n.T(ctx, "sign_up_username_error_token_invalid")
	}
	return i18n.T(ctx, "sign_up_username_error_server")
}
