package testutil

import (
	"testing"
	"time"

	"github.com/annict/annict/go/internal/auth"
)

// TestTOTPCodeは、作ったコードをアプリの照合がその時点のステップのコードとして受け付けることを検証する。
func TestTOTPCode(t *testing.T) {
	t.Parallel()

	const secret = "AAAAAAAAAAAAAAAAAAAAAAAAAAABELKN"
	at := time.Date(2026, 11, 2, 12, 0, 29, 0, time.UTC)
	code := TOTPCode(t, secret, at)

	step, ok := auth.MatchTOTPCode(secret, code, at)
	if !ok || step != auth.TOTPStep(at) {
		t.Errorf("auth.MatchTOTPCode(%q) = (%d, %t)、期待値 = (%d, true)", code, step, ok, auth.TOTPStep(at))
	}
}

// TestNonMatchingTOTPCodeは、次のステップが000000になる秘密鍵でも不一致用のコードを選び、
// 照合までに時刻が1ステップ進んだり戻ったりしても受け付けられないことを検証する。
func TestNonMatchingTOTPCode(t *testing.T) {
	t.Parallel()

	// この秘密鍵は、atのコードが711077、次のステップが000000となる再現用の値。
	const secret = "AAAAAAAAAAAAAAAAAAAAAAAAAAABELKN"
	at := time.Date(2026, 11, 2, 12, 0, 29, 0, time.UTC)
	code := NonMatchingTOTPCode(t, secret, at)
	if !auth.IsTOTPCodeFormat(code) {
		t.Fatalf("不一致用のコード = %q、6桁の数字を期待", code)
	}

	for _, now := range []time.Time{at.Add(-30 * time.Second), at, at.Add(2 * time.Second)} {
		if _, ok := auth.MatchTOTPCode(secret, code, now); ok {
			t.Errorf("%vで不一致用のコード%qを受け付けた", now, code)
		}
	}
}
