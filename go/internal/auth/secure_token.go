package auth

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateSecureTokenは安全なランダムトークンを生成する
// 24バイトのランダムデータをBase64 URL-safeエンコードした32文字の文字列を返す
func GenerateSecureToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
