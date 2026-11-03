package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// HKDFで1つの鍵から用途ごとの鍵を導くときのinfo。
// 同じ鍵を暗号化とダイジェストの両方へそのまま使わないよう、用途で導く鍵を分ける。
const (
	totpSecretEncryptionInfo = "annict totp secret encryption"
	recoveryCodeDigestInfo   = "annict recovery code digest"
)

// ErrTOTPSecretCiphertextInvalidは、暗号化した秘密鍵を復号できないときに返すエラー。
// 鍵を取り違えた・値が壊れた・別のユーザーの値を持ち込まれたときに起きる。
var ErrTOTPSecretCiphertextInvalid = errors.New("TOTPの秘密鍵を復号できません")

// TwoFactorKeyは二要素認証の秘密を守る鍵。
// TOTPの秘密鍵の暗号化 (AES-256-GCM) とリカバリーコードのダイジェスト (HMAC-SHA-256) を受け持つ。
//
// どちらも環境変数の1つの鍵からHKDFで導いた別々の鍵を使う。
// 鍵を変えると、保存済みの秘密鍵もリカバリーコードも照合できなくなる。
type TwoFactorKey struct {
	aead      cipher.AEAD
	digestKey []byte
}

// NewTwoFactorKeyは環境変数の鍵からTwoFactorKeyを作る。
//
// 環境変数の値をそのままAESの鍵にせずHKDFを通すのは、「32バイト以上の任意の文字列」を受け付け、
// `openssl rand -base64 48` の出力をそのまま設定できるようにするため。
func NewTwoFactorKey(secret string) (*TwoFactorKey, error) {
	encryptionKey, err := hkdf.Key(sha256.New, []byte(secret), nil, totpSecretEncryptionInfo, 32)
	if err != nil {
		return nil, err
	}
	digestKey, err := hkdf.Key(sha256.New, []byte(secret), nil, recoveryCodeDigestInfo, 32)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &TwoFactorKey{aead: aead, digestKey: digestKey}, nil
}

// EncryptTOTPSecretはTOTPの秘密鍵を暗号化し、ランダムなnonceを先頭に付けて返す。
//
// associatedDataには秘密鍵の持ち主 (ユーザーID) を渡す。
// 暗号文を別のユーザーの行へ書き写されても、復号に失敗して使えないようにするため。
func (k *TwoFactorKey) EncryptTOTPSecret(secret string, associatedData []byte) ([]byte, error) {
	nonce := make([]byte, k.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return k.aead.Seal(nonce, nonce, []byte(secret), associatedData), nil
}

// DecryptTOTPSecretはEncryptTOTPSecretの暗号文を復号する。
// 暗号化したときと同じassociatedDataを渡す。復号できないときはErrTOTPSecretCiphertextInvalidを返す。
func (k *TwoFactorKey) DecryptTOTPSecret(ciphertext, associatedData []byte) (string, error) {
	nonceSize := k.aead.NonceSize()
	if len(ciphertext) < nonceSize+k.aead.Overhead() {
		return "", ErrTOTPSecretCiphertextInvalid
	}

	plaintext, err := k.aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], associatedData)
	if err != nil {
		return "", ErrTOTPSecretCiphertextInvalid
	}

	return string(plaintext), nil
}

// RecoveryCodeDigestはリカバリーコードの16進表記のHMAC-SHA-256のダイジェストを返す。
// 生成したコードも入力されたコードも、NormalizeRecoveryCodeを通してから渡す。
//
// 鍵を使うのは、コードのエントロピーが40ビットほどしか無いため。
// 鍵の無いダイジェストでは、データベースが漏れたときに総当たりで平文へ戻せてしまう。
// 決定的なダイジェストにするのは、保存した値をユーザーとの組で完全一致で引けるようにするため。
func (k *TwoFactorKey) RecoveryCodeDigest(normalizedCode string) string {
	mac := hmac.New(sha256.New, k.digestKey)
	mac.Write([]byte(normalizedCode))

	return hex.EncodeToString(mac.Sum(nil))
}
