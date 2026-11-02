package ratelimit

import (
	"context"
	"fmt"
	"time"
)

const (
	// SignInTwoFactorUserLimitは、ログインの二要素認証を1人のユーザーが窓の間に試せる回数。
	// 6桁のコードとリカバリーコードの総当たりを防ぐため、コードが合った試行も含めて数える。
	SignInTwoFactorUserLimit = 5
	// SignInTwoFactorIPLimitは、ログインの二要素認証を1つのIPアドレスから窓の間に試せる回数。
	// 多くのユーザーへ少しずつ試す総当たりを防ぐ。
	SignInTwoFactorIPLimit = 30
	// SignInTwoFactorWindowは、ログインの二要素認証の回数を数える窓の長さ。
	SignInTwoFactorWindow = 15 * time.Minute
)

// SignInTwoFactorResultは、ログインの二要素認証の回数制限の判定の結果。
type SignInTwoFactorResult struct {
	// Allowedは試行を受け付けるかどうか
	Allowed bool
	// Scopeは上限を超えた単位 ("ip" か "user")。Allowedがtrueのときは空
	Scope string
}

// CheckSignInTwoFactorは、ログインの二要素認証の試行をIPアドレスとユーザーの単位で数え、上限を超えていないかを判定する。
//
// 認証アプリのコードの画面とリカバリーコードの画面で同じキーを使い、2つの画面を行き来しても回数が分かれないようにする。
func (l *Limiter) CheckSignInTwoFactor(ctx context.Context, ipAddress string, userID string) (*SignInTwoFactorResult, error) {
	checks := []struct {
		scope string
		input CheckInput
	}{
		{scope: "ip", input: CheckInput{Key: "sign_in:two_factor:ip:" + ipAddress, Limit: SignInTwoFactorIPLimit, Window: SignInTwoFactorWindow}},
		{scope: "user", input: CheckInput{Key: "sign_in:two_factor:user:" + userID, Limit: SignInTwoFactorUserLimit, Window: SignInTwoFactorWindow}},
	}

	for _, check := range checks {
		result, err := l.Check(ctx, check.input)
		if err != nil {
			return nil, fmt.Errorf("ログインの二要素認証の回数制限の判定に失敗 (%s): %w", check.scope, err)
		}
		if !result.Allowed {
			return &SignInTwoFactorResult{Allowed: false, Scope: check.scope}, nil
		}
	}

	return &SignInTwoFactorResult{Allowed: true}, nil
}
