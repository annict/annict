package auth

import (
	"strconv"
	"testing"
)

func TestGenerateVerificationCode(t *testing.T) {
	t.Run("6桁の数字コードが生成されること", func(t *testing.T) {
		code, err := GenerateVerificationCode()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// 6桁であることを確認
		if len(code) != 6 {
			t.Errorf("expected 6 digits, got %d digits: %s", len(code), code)
		}

		// 数字のみで構成されていることを確認
		if _, err := strconv.Atoi(code); err != nil {
			t.Errorf("expected numeric code, got non-numeric: %s", code)
		}
	})

	t.Run("コードが100000～999999の範囲内であること", func(t *testing.T) {
		code, err := GenerateVerificationCode()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		codeInt, err := strconv.Atoi(code)
		if err != nil {
			t.Fatalf("failed to parse code: %v", err)
		}

		if codeInt < 100000 || codeInt > 999999 {
			t.Errorf("code out of range: got %d, want between 100000 and 999999", codeInt)
		}
	})

	t.Run("複数回呼び出しても異なるコードが生成されること", func(t *testing.T) {
		codes := make(map[string]bool)
		duplicates := 0

		// 100回生成してランダム性を確認
		for i := 0; i < 100; i++ {
			code, err := GenerateVerificationCode()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if codes[code] {
				duplicates++
			}
			codes[code] = true
		}

		// 重複が多すぎる場合はランダム性が不十分
		// 100回中10回以上重複したら警告（理論上は重複する可能性はある）
		if duplicates > 10 {
			t.Errorf("too many duplicates: %d out of 100 codes", duplicates)
		}

		// 少なくとも50種類以上のコードが生成されることを確認
		if len(codes) < 50 {
			t.Errorf("insufficient randomness: only %d unique codes out of 100", len(codes))
		}
	})
}

func TestHashCode(t *testing.T) {
	t.Parallel()

	code := "123456"

	// コードをハッシュ化
	hashedCode, err := HashCode(code)
	if err != nil {
		t.Fatalf("HashCode failed: %v", err)
	}

	// ハッシュが空でないことを確認
	if hashedCode == "" {
		t.Error("HashedCode should not be empty")
	}

	// ハッシュが元のコードと異なることを確認
	if hashedCode == code {
		t.Error("HashedCode should be different from the original code")
	}

	// bcryptハッシュは "$2a$" または "$2b$" で始まる
	if len(hashedCode) < 4 || (hashedCode[:4] != "$2a$" && hashedCode[:4] != "$2b$") {
		t.Errorf("HashedCode format is invalid: %s", hashedCode[:4])
	}
}

func TestVerifyCode_Success(t *testing.T) {
	t.Parallel()

	code := "123456"

	// コードをハッシュ化
	hashedCode, err := HashCode(code)
	if err != nil {
		t.Fatalf("HashCode failed: %v", err)
	}

	// 正しいコードで検証
	if !VerifyCode(code, hashedCode) {
		t.Error("VerifyCode should return true for correct code")
	}
}

func TestVerifyCode_Failure(t *testing.T) {
	t.Parallel()

	code := "123456"
	wrongCode := "654321"

	// コードをハッシュ化
	hashedCode, err := HashCode(code)
	if err != nil {
		t.Fatalf("HashCode failed: %v", err)
	}

	// 間違ったコードで検証
	if VerifyCode(wrongCode, hashedCode) {
		t.Error("VerifyCode should return false for incorrect code")
	}
}

func TestHashCode_DifferentHashesForSameInput(t *testing.T) {
	t.Parallel()

	code := "123456"

	// 同じコードを2回ハッシュ化
	hash1, err := HashCode(code)
	if err != nil {
		t.Fatalf("First HashCode failed: %v", err)
	}

	hash2, err := HashCode(code)
	if err != nil {
		t.Fatalf("Second HashCode failed: %v", err)
	}

	// bcryptはソルトを使用するため、同じ入力でも異なるハッシュが生成される
	if hash1 == hash2 {
		t.Error("HashCode should generate different hashes due to salt")
	}

	// ただし、両方とも検証は成功する
	if !VerifyCode(code, hash1) {
		t.Error("VerifyCode should return true for first hash")
	}
	if !VerifyCode(code, hash2) {
		t.Error("VerifyCode should return true for second hash")
	}
}
