package usecase

import (
	"fmt"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/model"
)

// twoFactorSecretAssociatedDataはTOTPの秘密鍵の暗号文を持ち主に結び付ける付加データ (ユーザーIDの10進表記) を返す。
// 暗号化と復号で同じ値を渡さないと復号できないため、どのUseCaseもこの関数で作る。
func twoFactorSecretAssociatedData(userID model.UserID) []byte {
	return []byte(userID.String())
}

// decryptTwoFactorSecretはユーザーの二要素認証の設定から、平文のTOTPの秘密鍵を取り出す。
func decryptTwoFactorSecret(key *auth.TwoFactorKey, twoFactorAuth *model.UserTwoFactorAuth) (string, error) {
	secret, err := key.DecryptTOTPSecret(twoFactorAuth.SecretCiphertext, twoFactorSecretAssociatedData(twoFactorAuth.UserID))
	if err != nil {
		return "", fmt.Errorf("TOTPの秘密鍵の復号に失敗: %w", err)
	}

	return secret, nil
}

// TwoFactorAuthSetupは、認証アプリへの登録の途中で画面に出す秘密鍵。
type TwoFactorAuthSetup struct {
	// Secretはbase32の秘密鍵。認証アプリへ手で入力してもらう。
	Secret string
	// OTPAuthURLは秘密鍵を認証アプリへ登録するotpauth URI。QRコードとリンクにする。
	OTPAuthURL string
}

// newTwoFactorAuthSetupは秘密鍵から、画面に出すTwoFactorAuthSetupを作る。
// 認証アプリにはユーザー名を表示し、どのアカウントのコードかを見分けられるようにする。
func newTwoFactorAuthSetup(secret string, user *model.User) (*TwoFactorAuthSetup, error) {
	url, err := auth.BuildOTPAuthURL(secret, user.Username)
	if err != nil {
		return nil, fmt.Errorf("otpauth URIの作成に失敗: %w", err)
	}

	return &TwoFactorAuthSetup{Secret: secret, OTPAuthURL: url}, nil
}
