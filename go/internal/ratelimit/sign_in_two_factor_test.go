package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/testutil"
)

// TestLimiter_CheckSignInTwoFactor_Userは、1人のユーザーの試行を15分で5回まで受け付け、
// 6回目からユーザーの単位で拒むことを検証する。
func TestLimiter_CheckSignInTwoFactor_User(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	limiter := newTestLimiter(t, &now)
	prefix := testutil.UniqueRateLimitPrefix(t)

	for i := range SignInTwoFactorUserLimit {
		result, err := limiter.CheckSignInTwoFactor(context.Background(), prefix+"-ip", prefix+"-user")
		if err != nil || !result.Allowed {
			t.Fatalf("%d回目のCheckSignInTwoFactor() = (%+v, %v)、受け付けることを期待", i+1, result, err)
		}
	}

	result, err := limiter.CheckSignInTwoFactor(context.Background(), prefix+"-ip", prefix+"-user")
	if err != nil {
		t.Fatalf("CheckSignInTwoFactor()のエラー = %v", err)
	}
	if result.Allowed || result.Scope != "user" {
		t.Errorf("上限を超えた試行の結果 = %+v、期待値 = ユーザーの単位で拒む", result)
	}

	// 窓が変われば、また試せる。
	now = now.Add(SignInTwoFactorWindow)
	if result, err := limiter.CheckSignInTwoFactor(context.Background(), prefix+"-ip", prefix+"-user"); err != nil || !result.Allowed {
		t.Errorf("次の窓のCheckSignInTwoFactor() = (%+v, %v)、受け付けることを期待", result, err)
	}
}

// TestLimiter_CheckSignInTwoFactor_IPは、1つのIPアドレスからの試行を、ユーザーが違っても15分で30回まで受け付け、
// 31回目からIPアドレスの単位で拒むことを検証する。
func TestLimiter_CheckSignInTwoFactor_IP(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	limiter := newTestLimiter(t, &now)
	prefix := testutil.UniqueRateLimitPrefix(t)

	for i := range SignInTwoFactorIPLimit {
		result, err := limiter.CheckSignInTwoFactor(context.Background(), prefix+"-ip", prefix+"-user-"+string(rune('a'+i%26))+string(rune('a'+i/26)))
		if err != nil || !result.Allowed {
			t.Fatalf("%d回目のCheckSignInTwoFactor() = (%+v, %v)、受け付けることを期待", i+1, result, err)
		}
	}

	result, err := limiter.CheckSignInTwoFactor(context.Background(), prefix+"-ip", prefix+"-user-new")
	if err != nil {
		t.Fatalf("CheckSignInTwoFactor()のエラー = %v", err)
	}
	if result.Allowed || result.Scope != "ip" {
		t.Errorf("上限を超えた試行の結果 = %+v、期待値 = IPアドレスの単位で拒む", result)
	}
}
