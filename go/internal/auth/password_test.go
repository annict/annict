package auth

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSetBcryptCostForTest(t *testing.T) {
	// テスト前のデフォルト値を保存して、テスト後に復元
	originalCost := bcryptCost
	t.Cleanup(func() {
		bcryptCost = originalCost
	})

	t.Run("MinCostに変更してHashPasswordが動作する", func(t *testing.T) {
		SetBcryptCostForTest(bcrypt.MinCost)

		hashed, err := HashPassword("testpassword")
		if err != nil {
			t.Fatalf("HashPassword() error = %v", err)
		}

		// ハッシュ化されたパスワードが検証できることを確認
		if err := CheckPassword(hashed, "testpassword"); err != nil {
			t.Errorf("CheckPassword() error = %v", err)
		}

		// bcryptのコストがMinCostであることを確認
		cost, err := bcrypt.Cost([]byte(hashed))
		if err != nil {
			t.Fatalf("bcrypt.Cost() error = %v", err)
		}
		if cost != bcrypt.MinCost {
			t.Errorf("bcrypt cost = %d, want %d", cost, bcrypt.MinCost)
		}
	})

	t.Run("DefaultCostに戻してHashPasswordが動作する", func(t *testing.T) {
		SetBcryptCostForTest(bcrypt.DefaultCost)

		hashed, err := HashPassword("testpassword")
		if err != nil {
			t.Fatalf("HashPassword() error = %v", err)
		}

		cost, err := bcrypt.Cost([]byte(hashed))
		if err != nil {
			t.Fatalf("bcrypt.Cost() error = %v", err)
		}
		if cost != bcrypt.DefaultCost {
			t.Errorf("bcrypt cost = %d, want %d", cost, bcrypt.DefaultCost)
		}
	})
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "有効なパスワード: 8文字",
			password: "abcd1234",
			wantErr:  nil,
		},
		{
			name:     "有効なパスワード: 記号を含む",
			password: "pass@word123!",
			wantErr:  nil,
		},
		{
			name:     "有効なパスワード: 小文字のみ（文字種要件なし）",
			password: "abcdefgh",
			wantErr:  nil,
		},
		{
			name:     "有効なパスワード: 数字のみ（文字種要件なし）",
			password: "12345678",
			wantErr:  nil,
		},
		{
			name:     "有効なパスワード: 128文字（最大長）",
			password: strings.Repeat("a", 128),
			wantErr:  nil,
		},
		{
			name:     "有効なパスワード: 印字可能ASCII文字全種類",
			password: "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
			wantErr:  nil,
		},
		{
			name:     "無効: 7文字（最小長未満）",
			password: "abcd123",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "無効: 129文字（最大長超過）",
			password: strings.Repeat("a", 129),
			wantErr:  ErrPasswordTooLong,
		},
		{
			name:     "無効: 空文字",
			password: "",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "無効: スペースを含む",
			password: "pass word123",
			wantErr:  ErrPasswordInvalidChars,
		},
		{
			name:     "無効: タブ文字を含む",
			password: "pass\tword123",
			wantErr:  ErrPasswordInvalidChars,
		},
		{
			name:     "無効: 改行を含む",
			password: "pass\nword123",
			wantErr:  ErrPasswordInvalidChars,
		},
		{
			name:     "無効: Unicode文字を含む",
			password: "パスワード12345",
			wantErr:  ErrPasswordInvalidChars,
		},
		{
			name:     "無効: 絵文字を含む",
			password: "password😀",
			wantErr:  ErrPasswordInvalidChars,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePasswordStrength(tt.password)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidatePasswordStrength() error = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Errorf("ValidatePasswordStrength() error = nil, want %v", tt.wantErr)
				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidatePasswordStrength() error = %v, want errors.Is(_, %v)", err, tt.wantErr)
			}
		})
	}
}
