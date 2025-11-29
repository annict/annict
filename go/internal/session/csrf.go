package session

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateCSRFToken はRails互換のCSRFトークンを生成する
// Rails版: SecureRandom.base64(32)
// 32バイトのランダムデータを生成し、Base64エンコードして返す
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
