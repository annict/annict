package auth_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/annict/annict/go/internal/auth"
)

// testTOTPSecretはテストで使うbase32の秘密鍵。
const testTOTPSecret = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

// totpCodeAtはtの時点のコードを、認証アプリと同じパラメーターで作る。
func totpCodeAt(t *testing.T, at time.Time) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(testTOTPSecret, at, totp.ValidateOpts{
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("GenerateCodeCustom()のエラー = %v", err)
	}

	return code
}

// TestGenerateTOTPSecretは、秘密鍵が160ビットのbase32で、呼び出しごとに異なる値になることを検証する。
func TestGenerateTOTPSecret(t *testing.T) {
	t.Parallel()

	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret()のエラー = %v", err)
	}
	// 20バイトをパディングの無いbase32にすると32文字になる。
	if len(secret) != 32 {
		t.Errorf("秘密鍵の長さ = %d、期待値 = 32", len(secret))
	}

	// 生成した秘密鍵でコードを作れる。
	if _, err := totp.GenerateCode(secret, time.Now()); err != nil {
		t.Errorf("生成した秘密鍵でコードを作れない: %v", err)
	}

	other, err := auth.GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret()のエラー = %v", err)
	}
	if secret == other {
		t.Error("2回の呼び出しが同じ秘密鍵を返した")
	}
}

// TestBuildOTPAuthURLは、otpauth URIに発行者・アカウント名・秘密鍵・認証アプリの既定のパラメーターが載ることを検証する。
func TestBuildOTPAuthURL(t *testing.T) {
	t.Parallel()

	raw, err := auth.BuildOTPAuthURL(testTOTPSecret, "alice")
	if err != nil {
		t.Fatalf("BuildOTPAuthURL()のエラー = %v", err)
	}

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("URIの解析のエラー = %v", err)
	}
	if u.Scheme != "otpauth" || u.Host != "totp" || u.Path != "/Annict:alice" {
		t.Errorf("URI = %q、otpauth://totp/Annict:alice を期待", raw)
	}

	want := map[string]string{
		"secret":    testTOTPSecret,
		"issuer":    "Annict",
		"algorithm": "SHA1",
		"digits":    "6",
		"period":    "30",
	}
	for name, value := range want {
		if got := u.Query().Get(name); got != value {
			t.Errorf("%s = %q、期待値 = %q", name, got, value)
		}
	}
}

// TestBuildOTPAuthURL_InvalidSecretは、base32として読めない秘密鍵にエラーを返すことを検証する。
func TestBuildOTPAuthURL_InvalidSecret(t *testing.T) {
	t.Parallel()

	if _, err := auth.BuildOTPAuthURL("not base32!", "alice"); err == nil {
		t.Error("エラーを期待したが、nilだった")
	}
}

// TestMatchTOTPCodeは、前後1ステップまでのコードを受け付けて一致したステップを返し、
// それより離れたコードや形の違う入力を拒むことを検証する。
func TestMatchTOTPCode(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 25, 12, 0, 15, 0, time.UTC)
	current := auth.TOTPStep(now)

	tests := []struct {
		name     string
		secret   string
		code     string
		wantStep int64
		wantOK   bool
	}{
		{name: "今のステップのコード", secret: testTOTPSecret, code: totpCodeAt(t, now), wantStep: current, wantOK: true},
		{name: "1つ前のステップのコード", secret: testTOTPSecret, code: totpCodeAt(t, now.Add(-30*time.Second)), wantStep: current - 1, wantOK: true},
		{name: "1つ後のステップのコード", secret: testTOTPSecret, code: totpCodeAt(t, now.Add(30*time.Second)), wantStep: current + 1, wantOK: true},
		{name: "2つ前のステップのコード", secret: testTOTPSecret, code: totpCodeAt(t, now.Add(-60*time.Second))},
		{name: "2つ後のステップのコード", secret: testTOTPSecret, code: totpCodeAt(t, now.Add(60*time.Second))},
		{name: "5桁のコード", secret: testTOTPSecret, code: totpCodeAt(t, now)[:5]},
		{name: "数字でないコード", secret: testTOTPSecret, code: "abcdef"},
		{name: "空のコード", secret: testTOTPSecret, code: ""},
		{name: "読めない秘密鍵", secret: "not base32!", code: totpCodeAt(t, now)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			step, ok := auth.MatchTOTPCode(tt.secret, tt.code, now)
			if ok != tt.wantOK || step != tt.wantStep {
				t.Errorf("MatchTOTPCode() = (%d, %t)、期待値 = (%d, %t)", step, ok, tt.wantStep, tt.wantOK)
			}
		})
	}
}

// TestTOTPStepは、タイムステップが30秒ごとに1つ進むことを検証する。
func TestTOTPStep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		at   time.Time
		want int64
	}{
		{at: time.Unix(0, 0), want: 0},
		{at: time.Unix(29, 0), want: 0},
		{at: time.Unix(30, 0), want: 1},
		{at: time.Unix(59, 999), want: 1},
	}

	for _, tt := range tests {
		if got := auth.TOTPStep(tt.at); got != tt.want {
			t.Errorf("TOTPStep(%v) = %d、期待値 = %d", tt.at.Unix(), got, tt.want)
		}
	}
}

// TestNormalizeTOTPCodeは、貼り付けで混じった空白とハイフンを取り除き、全角の数字を半角にすることを検証する。
func TestNormalizeTOTPCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "そのまま", input: "123456", want: "123456"},
		{name: "前後と間の空白", input: " 123 456\n", want: "123456"},
		{name: "ハイフン", input: "123-456", want: "123456"},
		{name: "全角の数字と空白", input: "１２３　４５６", want: "123456"},
		{name: "日本語入力で打ったハイフン (長音符)", input: "１２３ー４５６", want: "123456"},
		{name: "貼り付けで混じるハイフン類", input: "1‐2‑3–4—5−6", want: "123456"},
		{name: "数字以外は残す", input: "12a456", want: "12a456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := auth.NormalizeTOTPCode(tt.input); got != tt.want {
				t.Errorf("NormalizeTOTPCode(%q) = %q、期待値 = %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsTOTPCodeFormatは、6桁の数字だけを認証アプリのコードの形として受け付けることを検証する。
// 1つの入力欄でリカバリーコードやパスワードと見分けるため、数字以外や桁の違う値を拒む。
func TestIsTOTPCodeFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want bool
	}{
		{name: "6桁の数字", code: "012345", want: true},
		{name: "5桁の数字", code: "01234", want: false},
		{name: "7桁の数字", code: "0123456", want: false},
		{name: "英字を含む", code: "01234a", want: false},
		{name: "リカバリーコードの形", code: "abcd2345", want: false},
		{name: "全角の数字 (正規化前)", code: "０１２３４５", want: false},
		{name: "空", code: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := auth.IsTOTPCodeFormat(tt.code); got != tt.want {
				t.Errorf("IsTOTPCodeFormat(%q) = %t、期待値 = %t", tt.code, got, tt.want)
			}
		})
	}
}
