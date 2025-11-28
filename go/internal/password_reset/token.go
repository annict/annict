// Package password_reset はパスワードリセット機能を提供します
package password_reset

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// GenerateToken は暗号学的に安全なリセットトークンを生成します
func GenerateToken() (string, error) {
	// 32バイト（256ビット）のランダムデータを生成
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// URLセーフなBase64エンコード（パディングなし）
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken はトークンをSHA-256でハッシュ化します
// データベースに保存する際は、このハッシュ化された値を使用します
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
