package session

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/model"
)

const (
	// signInTwoFactorKeyは、第1の要素 (パスワードかメールのコード) を確かめたユーザーと、
	// 認証アプリのコードを受け付ける期限を持つセッションの値のキー。値は「ユーザーID:期限のUnix時刻」の形。
	//
	// sign_in_user_idはメールアドレスを入力した時点 (照合の前) で書かれるため使わず、
	// 照合のあとにだけ書くこのキーで、二要素認証の画面がコードを照合する相手を決める。
	signInTwoFactorKey = "sign_in_two_factor"

	// SignInTwoFactorTTLは、第1の要素を確かめてから認証アプリのコードを受け付ける時間。
	SignInTwoFactorTTL = 10 * time.Minute
)

// StartSignInTwoFactorは、第1の要素を確かめたユーザーをログイン途中の状態にする。
//
// メールアドレスの入力からの一時的な値 (sign_in_email・sign_in_user_id) を消し、
// ユーザーIDと期限 (nowからSignInTwoFactorTTL後) を1つの値として書く。
// 1つの値にするのは、セッションが無いときにSetValueを重ねて呼ぶと別々のセッションができるため。
func (m *Manager) StartSignInTwoFactor(ctx context.Context, w http.ResponseWriter, r *http.Request, userID model.UserID, now time.Time) error {
	for _, key := range []string{"sign_in_email", "sign_in_user_id"} {
		if err := m.DeleteValue(ctx, r, key); err != nil {
			return fmt.Errorf("一時的なセッションの値 (%s) の削除に失敗しました: %w", key, err)
		}
	}

	value := userID.String() + ":" + strconv.FormatInt(now.Add(SignInTwoFactorTTL).Unix(), 10)
	if err := m.SetValue(ctx, w, r, signInTwoFactorKey, value); err != nil {
		return fmt.Errorf("ログイン途中の状態の保存に失敗しました: %w", err)
	}
	return nil
}

// SignInTwoFactorUserIDは、ログイン途中のユーザーIDを返す。
// 状態が無い・壊れている・期限 (now) を過ぎたときはokをfalseにする。
func (m *Manager) SignInTwoFactorUserID(ctx context.Context, r *http.Request, now time.Time) (userID model.UserID, ok bool, err error) {
	value, err := m.GetValue(ctx, r, signInTwoFactorKey)
	if err != nil {
		return 0, false, fmt.Errorf("ログイン途中の状態の取得に失敗しました: %w", err)
	}

	id, expiresAt, found := strings.Cut(value, ":")
	if !found {
		return 0, false, nil
	}
	parsedID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedID <= 0 {
		return 0, false, nil
	}
	expiresAtUnix, err := strconv.ParseInt(expiresAt, 10, 64)
	if err != nil || !now.Before(time.Unix(expiresAtUnix, 0)) {
		return 0, false, nil
	}

	return model.UserID(parsedID), true, nil
}

// ClearSignInTwoFactorは、ログイン途中の状態を消す。
// ログインを終えたときと、やり直させるときに呼ぶ。
func (m *Manager) ClearSignInTwoFactor(ctx context.Context, r *http.Request) error {
	if err := m.DeleteValue(ctx, r, signInTwoFactorKey); err != nil {
		return fmt.Errorf("ログイン途中の状態の削除に失敗しました: %w", err)
	}
	return nil
}
