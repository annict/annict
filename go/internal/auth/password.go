// Package auth は認証機能を提供します
package auth

import (
	"context"
	"errors"

	"github.com/annict/annict/internal/i18n"
	"golang.org/x/crypto/bcrypt"
)

// CheckPassword bcryptでハッシュ化されたパスワードと平文パスワードを比較する
// Deviseのencrypted_passwordカラムとの互換性を保つ
func CheckPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

// HashPassword 平文パスワードをbcryptでハッシュ化する
// Deviseのencrypted_passwordカラムとの互換性を保つ
func HashPassword(plainPassword string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// ValidatePasswordStrength はパスワードの強度をチェックする
// NIST SP 800-63B-4 準拠:
// - 最小文字数: 8文字
// - 最大文字数: 128文字
// - 印字可能ASCII文字のみ許可（0x21～0x7E）
// - 文字種の複雑性要件は廃止（大文字・小文字・数字・記号の組み合わせ要求なし）
func ValidatePasswordStrength(ctx context.Context, password string) error {
	// 最小文字数チェック
	if len(password) < 8 {
		return errors.New(i18n.T(ctx, "password_strength_min_length"))
	}

	// 最大文字数チェック
	if len(password) > 128 {
		return errors.New(i18n.T(ctx, "password_strength_max_length"))
	}

	// 印字可能ASCII文字のみを許可（0x21～0x7E）
	for _, char := range password {
		if char < 0x21 || char > 0x7E {
			return errors.New(i18n.T(ctx, "password_strength_invalid_chars"))
		}
	}

	return nil
}
