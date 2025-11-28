package session

import (
	"encoding/base64"
	"testing"
)

func TestGenerateCSRFToken(t *testing.T) {
	t.Run("トークンが生成されること", func(t *testing.T) {
		token, err := GenerateCSRFToken()
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if token == "" {
			t.Error("トークンが空文字列です")
		}
	})

	t.Run("トークンの長さが正しいこと", func(t *testing.T) {
		token, err := GenerateCSRFToken()
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}

		// 32バイトのランダムデータをBase64エンコードすると44文字になる
		expectedLen := 44
		if len(token) != expectedLen {
			t.Errorf("トークンの長さが正しくありません: got %d want %d", len(token), expectedLen)
		}
	})

	t.Run("Base64エンコードされた文字列であること", func(t *testing.T) {
		token, err := GenerateCSRFToken()
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}

		// Base64デコードが成功することを確認
		decoded, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			t.Errorf("Base64デコードに失敗しました: %v", err)
		}

		// デコード後のバイト数が32バイトであることを確認
		if len(decoded) != 32 {
			t.Errorf("デコード後のバイト数が正しくありません: got %d want 32", len(decoded))
		}
	})

	t.Run("2回呼び出すと異なるトークンが生成されること", func(t *testing.T) {
		token1, err := GenerateCSRFToken()
		if err != nil {
			t.Fatalf("1回目の呼び出しでエラー: %v", err)
		}

		token2, err := GenerateCSRFToken()
		if err != nil {
			t.Fatalf("2回目の呼び出しでエラー: %v", err)
		}

		if token1 == token2 {
			t.Error("同じトークンが生成されました（ランダム性が失われています）")
		}
	})

	t.Run("複数回呼び出してもすべて異なるトークンが生成されること", func(t *testing.T) {
		tokens := make(map[string]bool)
		iterations := 100

		for i := 0; i < iterations; i++ {
			token, err := GenerateCSRFToken()
			if err != nil {
				t.Fatalf("%d回目の呼び出しでエラー: %v", i+1, err)
			}

			if tokens[token] {
				t.Errorf("重複したトークンが生成されました: %s", token)
			}
			tokens[token] = true
		}

		if len(tokens) != iterations {
			t.Errorf("生成されたユニークなトークン数が正しくありません: got %d want %d", len(tokens), iterations)
		}
	})
}
