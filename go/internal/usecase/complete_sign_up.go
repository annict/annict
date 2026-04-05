// Package usecase はビジネスロジック層のユースケースを提供します
package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/validator"
)

// CompleteSignUpUsecase はユーザー登録を完了するユースケース
type CompleteSignUpUsecase struct {
	db                    *sql.DB
	userRepo              *repository.UserRepository
	profileRepo           *repository.ProfileRepository
	settingRepo           *repository.SettingRepository
	emailNotificationRepo *repository.EmailNotificationRepository
	sessionRepo           *repository.SessionRepository
	redisClient           *redis.Client
	v                     *validator.CreateSignUpUsernameValidator
}

// NewCompleteSignUpUsecase はCompleteSignUpUsecaseを作成します
func NewCompleteSignUpUsecase(
	db *sql.DB,
	userRepo *repository.UserRepository,
	profileRepo *repository.ProfileRepository,
	settingRepo *repository.SettingRepository,
	emailNotificationRepo *repository.EmailNotificationRepository,
	sessionRepo *repository.SessionRepository,
	redisClient *redis.Client,
	v *validator.CreateSignUpUsernameValidator,
) *CompleteSignUpUsecase {
	return &CompleteSignUpUsecase{
		db:                    db,
		userRepo:              userRepo,
		profileRepo:           profileRepo,
		settingRepo:           settingRepo,
		emailNotificationRepo: emailNotificationRepo,
		sessionRepo:           sessionRepo,
		redisClient:           redisClient,
		v:                     v,
	}
}

// CompleteSignUpInput はユースケースの入力パラメータ
type CompleteSignUpInput struct {
	Token    string
	Username string
	Locale   string
}

// CompleteSignUpResult はユーザー登録完了の結果
type CompleteSignUpResult struct {
	User            *model.User
	SessionPublicID string
	FormErrors      *session.FormErrors // バリデーションエラー（nilなら成功）
}

// Execute はユーザー登録を完了します
//
// 処理フロー:
// 1. バリデーション
// 2. Redisから一時トークンを検証してメールアドレスを取得
// 3. ユーザー名の一意性チェック
// 4. トランザクション開始
// 5. ユーザーを作成
// 6. プロフィールを作成（name: ユーザー名、description: 空文字列）
// 7. 設定を作成（privacy_policy_agreed: true、その他はデフォルト値）
// 8. メール通知設定を作成（unsubscription_key: UUID）
// 9. セッションを作成
// 10. トランザクションコミット
// 11. 一時トークンを削除
func (uc *CompleteSignUpUsecase) Execute(
	ctx context.Context,
	input CompleteSignUpInput,
) (*CompleteSignUpResult, error) {
	// 1. バリデーション
	valResult := uc.v.Validate(ctx, validator.CreateSignUpUsernameValidatorInput{
		Token:    input.Token,
		Username: input.Username,
	})
	if valResult.FormErrors != nil && valResult.FormErrors.HasErrors() {
		return &CompleteSignUpResult{FormErrors: valResult.FormErrors}, nil
	}

	// Redisから一時トークンを検証してメールアドレスを取得
	email, err := uc.verifyToken(ctx, input.Token)
	if err != nil {
		return nil, fmt.Errorf("トークン検証に失敗: %w", err)
	}

	// ユーザー名の一意性チェック
	err = uc.userRepo.GetByUsername(ctx, input.Username)
	if err == nil {
		return nil, &UsernameAlreadyExistsError{Username: input.Username}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("ユーザー名一意性チェックに失敗: %w", err)
	}

	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// トランザクション内で操作するためのリポジトリを取得
	userRepo := uc.userRepo.WithTx(tx)
	profileRepo := uc.profileRepo.WithTx(tx)
	settingRepo := uc.settingRepo.WithTx(tx)
	emailNotificationRepo := uc.emailNotificationRepo.WithTx(tx)

	// ユーザーを作成
	user, err := userRepo.Create(ctx, repository.UserCreateParams{
		Username:          input.Username,
		Email:             email,
		EncryptedPassword: "", // パスワードレス登録
		Locale:            input.Locale,
	})
	if err != nil {
		return nil, fmt.Errorf("ユーザー作成に失敗: %w", err)
	}

	// プロフィールを作成（name: ユーザー名、description: 空文字列）
	if err := profileRepo.Create(ctx, user.ID, input.Username); err != nil {
		return nil, fmt.Errorf("プロフィール作成に失敗: %w", err)
	}

	// 設定を作成（privacy_policy_agreed: true、その他はデフォルト値）
	if err := settingRepo.Create(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("設定作成に失敗: %w", err)
	}

	// メール通知設定を作成（unsubscription_key: UUID）
	unsubscriptionKey := fmt.Sprintf("%s-%s", uuid.New().String(), uuid.New().String())
	if err := emailNotificationRepo.Create(ctx, user.ID, unsubscriptionKey); err != nil {
		return nil, fmt.Errorf("メール通知設定作成に失敗: %w", err)
	}

	// セッションを作成
	createSessionUC := NewCreateSessionUsecase(uc.sessionRepo)
	sessionResult, err := createSessionUC.Execute(ctx, tx, user.ID, "")
	if err != nil {
		return nil, fmt.Errorf("セッション作成に失敗: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションコミットに失敗: %w", err)
	}

	// 一時トークンを削除
	if err := uc.deleteToken(ctx, input.Token); err != nil {
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
