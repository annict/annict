package password_reset

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("トークンの生成エラー = %v", err)
	}

	// トークン長の検証 (32バイト → 43文字 (Base64 RawURL))
	if len(token) < 43 {
		t.Errorf("トークンの文字数 = %d、期待値 = 43以上", len(token))
	}

	// 2つのトークンが異なることを確認 (衝突しないこと)
	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("2つ目のトークンの生成エラー = %v", err)
	}

	if token == token2 {
		t.Error("生成したトークンが重複した")
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token"
	hash1 := HashToken(token)
	hash2 := HashToken(token)

	// 同じトークンは同じハッシュを生成
	if hash1 != hash2 {
		t.Errorf("ハッシュ = %s != %s", hash1, hash2)
	}

	// ハッシュ長の検証 (SHA-256 → 64文字のhex)
	if len(hash1) != 64 {
		t.Errorf("ハッシュの文字数 = %d、期待値 = 64", len(hash1))
	}

	// 異なるトークンは異なるハッシュを生成
	differentToken := "different-token"
	hash3 := HashToken(differentToken)

	if hash1 == hash3 {
		t.Error("異なるトークンのハッシュが同一だった")
	}
}
