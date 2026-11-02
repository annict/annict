package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// totpPeriodSecondsは認証アプリがコードを切り替える間隔 (秒)。
const totpPeriodSeconds = 30

// TOTPCodeは、秘密鍵のatの時点のコードを、認証アプリと同じパラメーター (SHA1・6桁・30秒) で作る。
func TOTPCode(t *testing.T, secret string, at time.Time) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(secret, at, totp.ValidateOpts{
		Period:    totpPeriodSeconds,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("TOTPのコードの作成に失敗しました: %v", err)
	}
	return code
}

// NonMatchingTOTPCodeは、atの前後の照合窓内のコードと一致しない6桁の数字を返す。
// 前後2ステップを除外し、コードを用意してから照合するまでに30秒の境界を1度またいでも不一致を保つ。
func NonMatchingTOTPCode(t *testing.T, secret string, at time.Time) string {
	t.Helper()

	excluded := make(map[string]bool, 5)
	for offset := -2; offset <= 2; offset++ {
		excluded[TOTPCode(t, secret, at.Add(time.Duration(offset)*totpPeriodSeconds*time.Second))] = true
	}

	// 除外する値は高々5つのため、先頭の6候補以内で必ず選べる。
	for candidate := 0; ; candidate++ {
		code := fmt.Sprintf("%06d", candidate)
		if !excluded[code] {
			return code
		}
	}
}
