package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/model"
)

// TestTOTPEncryptionKeyはテストでauth.TwoFactorKeyを作るときに使う鍵 (32バイト)。
const TestTOTPEncryptionKey = "test-totp-encryption-key-32bytes"

// NewTestTwoFactorKeyはTestTOTPEncryptionKeyから作ったauth.TwoFactorKeyを返す。
func NewTestTwoFactorKey(t *testing.T) *auth.TwoFactorKey {
	t.Helper()

	key, err := auth.NewTwoFactorKey(TestTOTPEncryptionKey)
	if err != nil {
		t.Fatalf("テスト用のTwoFactorKeyの作成に失敗しました: %v", err)
	}
	return key
}

// UserTwoFactorAuthBuilderはuser_two_factor_authsのテストデータのビルダー。
// 既定では、復号できない仮の暗号文を持つ、認証アプリへの登録の途中の設定を作る。
type UserTwoFactorAuthBuilder struct {
	t                *testing.T
	tx               *sql.Tx
	userID           model.UserID
	secretCiphertext []byte
	lastUsedStep     int64
	enabledAt        *time.Time
}

// NewUserTwoFactorAuthBuilderはuserIDのユーザーのUserTwoFactorAuthBuilderを作成する。
func NewUserTwoFactorAuthBuilder(t *testing.T, tx *sql.Tx, userID model.UserID) *UserTwoFactorAuthBuilder {
	return &UserTwoFactorAuthBuilder{
		t:                t,
		tx:               tx,
		userID:           userID,
		secretCiphertext: []byte("test-secret-ciphertext"),
	}
}

// WithSecretCiphertextは暗号化したTOTPの秘密鍵を設定する。
// コードを照合するテストは、NewTestTwoFactorKeyで実際に暗号化した値を渡す。
func (b *UserTwoFactorAuthBuilder) WithSecretCiphertext(secretCiphertext []byte) *UserTwoFactorAuthBuilder {
	b.secretCiphertext = secretCiphertext
	return b
}

// WithLastUsedStepは最後に受け付けたコードのタイムステップを設定する。
func (b *UserTwoFactorAuthBuilder) WithLastUsedStep(step int64) *UserTwoFactorAuthBuilder {
	b.lastUsedStep = step
	return b
}

// WithEnabledAtは二要素認証を有効にした日時を設定する。設定すると有効な設定になる。
func (b *UserTwoFactorAuthBuilder) WithEnabledAt(enabledAt time.Time) *UserTwoFactorAuthBuilder {
	b.enabledAt = &enabledAt
	return b
}

// Buildは二要素認証の設定をデータベースに作成し、IDを返す。
func (b *UserTwoFactorAuthBuilder) Build() model.UserTwoFactorAuthID {
	b.t.Helper()

	var id int64
	err := b.tx.QueryRow(
		`INSERT INTO user_two_factor_auths (user_id, secret_ciphertext, last_used_step, enabled_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		int64(b.userID), b.secretCiphertext, b.lastUsedStep, b.enabledAt,
	).Scan(&id)
	if err != nil {
		b.t.Fatalf("テスト用の二要素認証の設定の作成に失敗しました: %v", err)
	}

	return model.UserTwoFactorAuthID(id)
}

// EnableTwoFactorAuthForTestは、userIDのユーザーの二要素認証を有効にした設定を作り、認証アプリに登録したことにする秘密鍵を返す。
// 秘密鍵はNewTestTwoFactorKeyの鍵で暗号化し、付加データにはUseCaseと同じくユーザーIDの10進表記を渡す。
// UseCaseのテスト (CreateSignInTwoFactorUsecase) がこの関数で作った設定を照合し、付加データの作り方が食い違わないことを確かめている。
func EnableTwoFactorAuthForTest(t *testing.T, tx *sql.Tx, userID model.UserID) string {
	t.Helper()

	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("TOTPの秘密鍵の生成に失敗しました: %v", err)
	}
	ciphertext, err := NewTestTwoFactorKey(t).EncryptTOTPSecret(secret, []byte(userID.String()))
	if err != nil {
		t.Fatalf("TOTPの秘密鍵の暗号化に失敗しました: %v", err)
	}
	NewUserTwoFactorAuthBuilder(t, tx, userID).WithSecretCiphertext(ciphertext).WithEnabledAt(time.Now()).Build()

	return secret
}
