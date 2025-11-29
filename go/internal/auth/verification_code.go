package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// GenerateVerificationCode は6桁の確認コードを生成します
// 新規登録とログインの両方で共通して使用します
func GenerateVerificationCode() (string, error) {
	// 100000～999999の範囲の乱数を生成
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	code := n.Int64() + 100000
	return fmt.Sprintf("%06d", code), nil
}

// HashCode はコードをbcryptでハッシュ化します
// データベースに保存する際は、このハッシュ化された値を使用します
// 新規登録とログインの両方で共通して使用します
func HashCode(code string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyCode はコードとハッシュを比較して一致を検証します
// 新規登録とログインの両方で共通して使用します
func VerifyCode(code, hashedCode string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(code))
	return err == nil
}
