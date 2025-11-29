package auth

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/internal/i18n"
)

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "有効なパスワード: 8文字",
			password: "abcd1234",
			wantErr:  false,
		},
		{
			name:     "有効なパスワード: 記号を含む",
			password: "pass@word123!",
			wantErr:  false,
		},
		{
			name:     "有効なパスワード: 小文字のみ（文字種要件なし）",
			password: "abcdefgh",
			wantErr:  false,
		},
		{
			name:     "有効なパスワード: 数字のみ（文字種要件なし）",
			password: "12345678",
			wantErr:  false,
		},
		{
			name:     "有効なパスワード: 128文字（最大長）",
			password: strings.Repeat("a", 128),
			wantErr:  false,
		},
		{
			name:     "有効なパスワード: 印字可能ASCII文字全種類",
			password: "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
			wantErr:  false,
		},
		{
			name:     "無効: 7文字（最小長未満）",
			password: "abcd123",
			wantErr:  true,
			errMsg:   "8文字以上",
		},
		{
			name:     "無効: 129文字（最大長超過）",
			password: strings.Repeat("a", 129),
			wantErr:  true,
			errMsg:   "128文字以内",
		},
		{
			name:     "無効: 空文字",
			password: "",
			wantErr:  true,
			errMsg:   "8文字以上",
		},
		{
			name:     "無効: スペースを含む",
			password: "pass word123",
			wantErr:  true,
			errMsg:   "半角英数記号のみ",
		},
		{
			name:     "無効: タブ文字を含む",
			password: "pass\tword123",
			wantErr:  true,
			errMsg:   "半角英数記号のみ",
		},
		{
			name:     "無効: 改行を含む",
			password: "pass\nword123",
			wantErr:  true,
			errMsg:   "半角英数記号のみ",
		},
		{
			name:     "無効: Unicode文字を含む",
			password: "パスワード12345",
			wantErr:  true,
			errMsg:   "半角英数記号のみ",
		},
		{
			name:     "無効: 絵文字を含む",
			password: "password😀",
			wantErr:  true,
			errMsg:   "半角英数記号のみ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// デフォルトロケール（日本語）でテスト
			ctx := context.Background()
			err := ValidatePasswordStrength(ctx, tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePasswordStrength() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePasswordStrength() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// TestValidatePasswordStrength_Japanese は日本語ロケールでのエラーメッセージをテストします
func TestValidatePasswordStrength_Japanese(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		expectedMsg string
	}{
		{
			name:        "最小長エラー（日本語）",
			password:    "short",
			wantErr:     true,
			expectedMsg: "パスワードは8文字以上である必要があります",
		},
		{
			name:        "最大長エラー（日本語）",
			password:    strings.Repeat("a", 129),
			wantErr:     true,
			expectedMsg: "パスワードは128文字以内である必要があります",
		},
		{
			name:        "無効な文字エラー（日本語）",
			password:    "password with space",
			wantErr:     true,
			expectedMsg: "パスワードは半角英数記号のみ使用できます",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 日本語ロケールを設定
			ctx := i18n.SetLocale(context.Background(), "ja")
			err := ValidatePasswordStrength(ctx, tt.password)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("ValidatePasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err == nil {
				t.Errorf("ValidatePasswordStrength() error = nil, wantErr %v", tt.wantErr)
				return
			}

			if err.Error() != tt.expectedMsg {
				t.Errorf("ValidatePasswordStrength() error = %v, want %v", err.Error(), tt.expectedMsg)
			}
		})
	}
}

// TestValidatePasswordStrength_English は英語ロケールでのエラーメッセージをテストします
func TestValidatePasswordStrength_English(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		expectedMsg string
	}{
		{
			name:        "最小長エラー（英語）",
			password:    "short",
			wantErr:     true,
			expectedMsg: "Password must be at least 8 characters long",
		},
		{
			name:        "最大長エラー（英語）",
			password:    strings.Repeat("a", 129),
			wantErr:     true,
			expectedMsg: "Password must be no more than 128 characters long",
		},
		{
			name:        "無効な文字エラー（英語）",
			password:    "password with space",
			wantErr:     true,
			expectedMsg: "Password can only use alphanumeric characters and symbols",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 英語ロケールを設定
			ctx := i18n.SetLocale(context.Background(), "en")
			err := ValidatePasswordStrength(ctx, tt.password)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("ValidatePasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err == nil {
				t.Errorf("ValidatePasswordStrength() error = nil, wantErr %v", tt.wantErr)
				return
			}

			if err.Error() != tt.expectedMsg {
				t.Errorf("ValidatePasswordStrength() error = %v, want %v", err.Error(), tt.expectedMsg)
			}
		})
	}
}
