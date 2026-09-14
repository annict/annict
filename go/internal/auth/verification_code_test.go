package auth

import (
	"strconv"
	"testing"
)

func TestGenerateVerificationCode(t *testing.T) {
	t.Run("6桁の数字コードが生成されること", func(t *testing.T) {
		code, err := GenerateVerificationCode()
		if err != nil {
			t.Fatalf("エラー = %v、期待値 = nil", err)
		}

		// 6桁であることを確認
		if len(code) != 6 {
			t.Errorf("桁数 = %d、期待値 = 6 (%s)", len(code), code)
		}

		// 数字のみで構成されていることを確認
		if _, err := strconv.Atoi(code); err != nil {
			t.Errorf("コード = %s、期待値 = 数字のみ", code)
		}
	})

	t.Run("コードが100000～999999の範囲内であること", func(t *testing.T) {
		code, err := GenerateVerificationCode()
		if err != nil {
			t.Fatalf("エラー = %v、期待値 = nil", err)
		}

		codeInt, err := strconv.Atoi(code)
		if err != nil {
			t.Fatalf("コードの解析エラー = %v", err)
		}

		if codeInt < 100000 || codeInt > 999999 {
			t.Errorf("コード = %d、期待値 = 100000以上999999以下", codeInt)
		}
	})

	t.Run("複数回呼び出しても異なるコードが生成されること", func(t *testing.T) {
		codes := make(map[string]bool)
		duplicates := 0

		// 100回生成してランダム性を確認
		for i := 0; i < 100; i++ {
			code, err := GenerateVerificationCode()
			if err != nil {
				t.Fatalf("エラー = %v、期待値 = nil", err)
			}

			if codes[code] {
				duplicates++
			}
			codes[code] = true
		}

		// 重複が多すぎる場合はランダム性が不十分
		// 100回中10回以上重複したら警告 (理論上は重複する可能性はある)
		if duplicates > 10 {
			t.Errorf("重複が多すぎる。100件中の重複数 = %d", duplicates)
		}

		// 少なくとも50種類以上のコードが生成されることを確認
		if len(codes) < 50 {
			t.Errorf("ランダム性が不足している。100件中の一意なコード数 = %d", len(codes))
		}
	})
}

func TestHashCode(t *testing.T) {
	t.Parallel()

	code := "123456"

	// コードをハッシュ化
	hashedCode, err := HashCode(code)
	if err != nil {
		t.Fatalf("HashCode()のエラー = %v", err)
	}

	// ハッシュが空でないことを確認
	if hashedCode == "" {
		t.Error("HashedCodeが空だった")
	}

	// ハッシュが元のコードと異なることを確認
	if hashedCode == code {
		t.Error("HashedCodeが元のコードと同一だった")
	}

	// bcryptハッシュは "$2a$" または "$2b$" で始まる
	if len(hashedCode) < 4 || (hashedCode[:4] != "$2a$" && hashedCode[:4] != "$2b$") {
		t.Errorf("HashedCodeの先頭4文字 = %s、期待値 = $2a$または$2b$", hashedCode[:4])
	}
}

func TestVerifyCode_Success(t *testing.T) {
	t.Parallel()

	code := "123456"

	// コードをハッシュ化
	hashedCode, err := HashCode(code)
	if err != nil {
		t.Fatalf("HashCode()のエラー = %v", err)
	}

	// 正しいコードで検証
	if !VerifyCode(code, hashedCode) {
		t.Error("正しいコードでVerifyCode()がtrueを返さなかった")
	}
}

func TestVerifyCode_Failure(t *testing.T) {
	t.Parallel()

	code := "123456"
	wrongCode := "654321"

	// コードをハッシュ化
	hashedCode, err := HashCode(code)
	if err != nil {
		t.Fatalf("HashCode()のエラー = %v", err)
	}

	// 間違ったコードで検証
	if VerifyCode(wrongCode, hashedCode) {
		t.Error("誤ったコードでVerifyCode()がfalseを返さなかった")
	}
}

func TestHashCode_DifferentHashesForSameInput(t *testing.T) {
	t.Parallel()

	code := "123456"

	// 同じコードを2回ハッシュ化
	hash1, err := HashCode(code)
	if err != nil {
		t.Fatalf("1回目のHashCode()のエラー = %v", err)
	}

	hash2, err := HashCode(code)
	if err != nil {
		t.Fatalf("2回目のHashCode()のエラー = %v", err)
	}

	// bcryptはソルトを使用するため、同じ入力でも異なるハッシュが生成される
	if hash1 == hash2 {
		t.Error("ソルトにより異なるハッシュになるべきだが、同一だった")
	}

	// ただし、両方とも検証は成功する
	if !VerifyCode(code, hash1) {
		t.Error("1つ目のハッシュでVerifyCode()がtrueを返さなかった")
	}
	if !VerifyCode(code, hash2) {
		t.Error("2つ目のハッシュでVerifyCode()がtrueを返さなかった")
	}
}
