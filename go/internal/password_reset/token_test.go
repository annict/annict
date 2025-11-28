package password_reset

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// トークン長の検証（32バイト → 43文字（Base64 RawURL））
	if len(token) < 43 {
		t.Errorf("token too short: got %d, want at least 43", len(token))
	}

	// 2つのトークンが異なることを確認（衝突しないこと）
	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("failed to generate second token: %v", err)
	}

	if token == token2 {
		t.Error("generated tokens should be unique")
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token"
	hash1 := HashToken(token)
	hash2 := HashToken(token)

	// 同じトークンは同じハッシュを生成
	if hash1 != hash2 {
		t.Errorf("hash mismatch: %s != %s", hash1, hash2)
	}

	// ハッシュ長の検証（SHA-256 → 64文字のhex）
	if len(hash1) != 64 {
		t.Errorf("hash length should be 64, got %d", len(hash1))
	}

	// 異なるトークンは異なるハッシュを生成
	differentToken := "different-token"
	hash3 := HashToken(differentToken)

	if hash1 == hash3 {
		t.Error("different tokens should produce different hashes")
	}
}
