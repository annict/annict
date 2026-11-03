package auth_test

import (
	"regexp"
	"testing"

	"github.com/annict/annict/go/internal/auth"
)

// TestGenerateRecoveryCodesは、10個のコードが見間違えやすい文字を含まない `xxxx-xxxx` の形で、互いに異なることを検証する。
func TestGenerateRecoveryCodes(t *testing.T) {
	t.Parallel()

	codes, err := auth.GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes()のエラー = %v", err)
	}
	if len(codes) != auth.RecoveryCodeCount {
		t.Fatalf("コードの数 = %d、期待値 = %d", len(codes), auth.RecoveryCodeCount)
	}

	shape := regexp.MustCompile(`^[a-z2-9]{4}-[a-z2-9]{4}$`)
	confusable := regexp.MustCompile(`[01oli]`)
	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		if !shape.MatchString(code) {
			t.Errorf("コード = %q、xxxx-xxxx の形の小文字と数字を期待", code)
		}
		if confusable.MatchString(code) {
			t.Errorf("コード = %q に見間違えやすい文字が含まれている", code)
		}
		if seen[code] {
			t.Errorf("コード %q が重複している", code)
		}
		seen[code] = true
	}
}

// TestNormalizeRecoveryCodeは、大文字小文字・ハイフン・空白・全角の違いを吸収して同じ形にすることを検証する。
// 生成したコードと入力したコードのダイジェストが一致することが、この正規化に依存している。
func TestNormalizeRecoveryCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "生成した形", input: "abcd-2345"},
		{name: "ハイフン無し", input: "abcd2345"},
		{name: "大文字", input: "ABCD-2345"},
		{name: "空白区切り", input: " abcd 2345 "},
		{name: "全角", input: "ＡＢＣＤ－２３４５"},
		{name: "長音符の区切り", input: "abcdー2345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := auth.NormalizeRecoveryCode(tt.input); got != "abcd2345" {
				t.Errorf("NormalizeRecoveryCode(%q) = %q、期待値 = %q", tt.input, got, "abcd2345")
			}
		})
	}
}

// TestIsRecoveryCodeFormatは、発行したコードを正規化すると必ず受け付け、
// 使わない文字や長さの違うコードを拒むことを検証する。
func TestIsRecoveryCodeFormat(t *testing.T) {
	t.Parallel()

	codes, err := auth.GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes()のエラー = %v", err)
	}
	for _, code := range codes {
		if normalized := auth.NormalizeRecoveryCode(code); !auth.IsRecoveryCodeFormat(normalized) {
			t.Errorf("発行したコード %q (正規化後 %q) を受け付けない", code, normalized)
		}
	}

	for _, input := range []string{"", "abcd234", "abcd23456", "abcd234o", "abcd2340", "abcd234l", "abcd2341", "abcd234i", "abcd-234", "ABCD2345"} {
		if auth.IsRecoveryCodeFormat(input) {
			t.Errorf("IsRecoveryCodeFormat(%q) = true、falseを期待", input)
		}
	}
}
