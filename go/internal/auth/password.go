// Package auth は認証機能を提供します
package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost はbcryptのコスト値。テスト時はSetBcryptCostForTestで変更可能
var bcryptCost = bcrypt.DefaultCost

// SetBcryptCostForTest はテスト用にbcryptコストを変更する
// テスト以外からの呼び出しは想定していない
func SetBcryptCostForTest(cost int) {
	bcryptCost = cost
}

// CheckPassword bcryptでハッシュ化されたパスワードと平文パスワードを比較する
// Deviseのencrypted_passwordカラムとの互換性を保つ
func CheckPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

// HashPassword 平文パスワードをbcryptでハッシュ化する
// Deviseのencrypted_passwordカラムとの互換性を保つ
func HashPassword(plainPassword string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// パスワード強度の要件（NIST SP 800-63B-4 準拠）
const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
)

// パスワード強度のバリデーションエラー（sentinel error）
// auth パッケージは i18n に依存しないため、呼び出し元でエラー種別を判別して翻訳を解決する
var (
	ErrPasswordTooShort     = errors.New("password is too short")
	ErrPasswordTooLong      = errors.New("password is too long")
	ErrPasswordInvalidChars = errors.New("password contains invalid characters")
)

// ValidatePasswordStrength はパスワードの強度をチェックする
// NIST SP 800-63B-4 準拠:
// - 最小文字数: 8文字
// - 最大文字数: 128文字
// - 印字可能ASCII文字のみ許可（0x21～0x7E）
// - 文字種の複雑性要件は廃止（大文字・小文字・数字・記号の組み合わせ要求なし）
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}
	for _, char := range password {
		if char < 0x21 || char > 0x7E {
			return ErrPasswordInvalidChars
		}
	}
	return nil
}
