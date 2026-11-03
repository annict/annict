package auth_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"testing"

	"github.com/annict/annict/go/internal/auth"
)

// newTwoFactorKeyはテスト用の鍵でTwoFactorKeyを作る。
func newTwoFactorKey(t *testing.T, secret string) *auth.TwoFactorKey {
	t.Helper()

	key, err := auth.NewTwoFactorKey(secret)
	if err != nil {
		t.Fatalf("NewTwoFactorKey()のエラー = %v", err)
	}

	return key
}

// TestTwoFactorKey_EncryptTOTPSecretは、暗号化した秘密鍵が平文を含まず、同じ鍵と付加データで復号できることを検証する。
func TestTwoFactorKey_EncryptTOTPSecret(t *testing.T) {
	t.Parallel()

	key := newTwoFactorKey(t, "test-totp-encryption-key-0123456789")
	userID := []byte("user-1")

	ciphertext, err := key.EncryptTOTPSecret("JBSWY3DPEHPK3PXP", userID)
	if err != nil {
		t.Fatalf("EncryptTOTPSecret()のエラー = %v", err)
	}
	if bytes.Contains(ciphertext, []byte("JBSWY3DPEHPK3PXP")) {
		t.Error("暗号文に平文の秘密鍵が含まれている")
	}

	// nonceが毎回変わるため、同じ秘密鍵でも暗号文は一致しない。
	other, err := key.EncryptTOTPSecret("JBSWY3DPEHPK3PXP", userID)
	if err != nil {
		t.Fatalf("2回目のEncryptTOTPSecret()のエラー = %v", err)
	}
	if bytes.Equal(ciphertext, other) {
		t.Error("同じ秘密鍵の2回の暗号化が同じ暗号文を返した")
	}

	got, err := key.DecryptTOTPSecret(ciphertext, userID)
	if err != nil {
		t.Fatalf("DecryptTOTPSecret()のエラー = %v", err)
	}
	if got != "JBSWY3DPEHPK3PXP" {
		t.Errorf("復号した秘密鍵 = %q、期待値 = %q", got, "JBSWY3DPEHPK3PXP")
	}
}

// TestTwoFactorKey_DecryptTOTPSecret_Invalidは、鍵・付加データ・暗号文が食い違うときに復号を拒むことを検証する。
func TestTwoFactorKey_DecryptTOTPSecret_Invalid(t *testing.T) {
	t.Parallel()

	key := newTwoFactorKey(t, "test-totp-encryption-key-0123456789")
	ciphertext, err := key.EncryptTOTPSecret("JBSWY3DPEHPK3PXP", []byte("user-1"))
	if err != nil {
		t.Fatalf("EncryptTOTPSecret()のエラー = %v", err)
	}

	tampered := bytes.Clone(ciphertext)
	tampered[len(tampered)-1] ^= 0x01

	tests := []struct {
		name           string
		key            *auth.TwoFactorKey
		ciphertext     []byte
		associatedData []byte
	}{
		{name: "別の鍵", key: newTwoFactorKey(t, "other-totp-encryption-key-0123456789"), ciphertext: ciphertext, associatedData: []byte("user-1")},
		{name: "別のユーザーの付加データ", key: key, ciphertext: ciphertext, associatedData: []byte("user-2")},
		{name: "書き換えた暗号文", key: key, ciphertext: tampered, associatedData: []byte("user-1")},
		{name: "nonceより短い暗号文", key: key, ciphertext: ciphertext[:4], associatedData: []byte("user-1")},
		{name: "空の暗号文", key: key, ciphertext: nil, associatedData: []byte("user-1")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.key.DecryptTOTPSecret(tt.ciphertext, tt.associatedData)
			if !errors.Is(err, auth.ErrTOTPSecretCiphertextInvalid) {
				t.Errorf("DecryptTOTPSecret()のエラー = %v、期待値 = %v", err, auth.ErrTOTPSecretCiphertextInvalid)
			}
		})
	}
}

// TestTwoFactorKey_RecoveryCodeDigestは、ダイジェストが同じ鍵とコードで一致し、鍵やコードが違えば一致しないことを検証する。
// 保存したダイジェストを完全一致で引けることと、鍵が無いと作れないことが、この性質に依存している。
func TestTwoFactorKey_RecoveryCodeDigest(t *testing.T) {
	t.Parallel()

	key := newTwoFactorKey(t, "test-totp-encryption-key-0123456789")
	digest := key.RecoveryCodeDigest("abcd2345")

	if got := key.RecoveryCodeDigest("abcd2345"); got != digest {
		t.Errorf("同じコードのダイジェスト = %q、期待値 = %q", got, digest)
	}
	if got := key.RecoveryCodeDigest("abcd2346"); got == digest {
		t.Error("異なるコードが同じダイジェストになった")
	}
	if got := newTwoFactorKey(t, "other-totp-encryption-key-0123456789").RecoveryCodeDigest("abcd2345"); got == digest {
		t.Error("異なる鍵で同じダイジェストになった")
	}
	// 鍵の無いSHA-256のダイジェストとは一致しない。
	plainDigest := sha256.Sum256([]byte("abcd2345"))
	if digest == hex.EncodeToString(plainDigest[:]) {
		t.Error("ダイジェストが鍵の無いSHA-256と同じ値になっている")
	}

	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(digest) {
		t.Errorf("ダイジェスト = %q、64文字の16進表記を期待", digest)
	}
}
