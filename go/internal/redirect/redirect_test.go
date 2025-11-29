package redirect

import "testing"

func TestValidateBackURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		backURL  string
		expected bool
	}{
		// 有効なケース
		{
			name:     "ルートパス",
			backURL:  "/",
			expected: true,
		},
		{
			name:     "単純なパス",
			backURL:  "/works",
			expected: true,
		},
		{
			name:     "ネストしたパス",
			backURL:  "/works/123",
			expected: true,
		},
		{
			name:     "クエリパラメータ付きパス",
			backURL:  "/oauth/authorize?client_id=xxx",
			expected: true,
		},
		{
			name:     "複数クエリパラメータ付きパス",
			backURL:  "/oauth/authorize?client_id=xxx&scope=read",
			expected: true,
		},
		{
			name:     "フラグメント付きパス",
			backURL:  "/works#section",
			expected: true,
		},
		{
			name:     "URLエンコードされたパス",
			backURL:  "/search?q=%E3%82%A2%E3%83%8B%E3%83%A1",
			expected: true,
		},

		// 無効なケース - 空文字
		{
			name:     "空文字",
			backURL:  "",
			expected: false,
		},

		// 無効なケース - 絶対URL（オープンリダイレクト攻撃）
		{
			name:     "HTTPの絶対URL",
			backURL:  "http://evil.com",
			expected: false,
		},
		{
			name:     "HTTPSの絶対URL",
			backURL:  "https://evil.com",
			expected: false,
		},
		{
			name:     "パス付きの絶対URL",
			backURL:  "https://evil.com/phishing",
			expected: false,
		},

		// 無効なケース - プロトコル相対URL（オープンリダイレクト攻撃）
		{
			name:     "プロトコル相対URL",
			backURL:  "//evil.com",
			expected: false,
		},
		{
			name:     "パス付きプロトコル相対URL",
			backURL:  "//evil.com/phishing",
			expected: false,
		},

		// 無効なケース - その他の攻撃パターン
		{
			name:     "JavaScriptスキーム",
			backURL:  "javascript:alert(1)",
			expected: false,
		},
		{
			name:     "DataスキームURL",
			backURL:  "data:text/html,<script>alert(1)</script>",
			expected: false,
		},
		{
			name:     "相対パス（スラッシュなし）",
			backURL:  "works",
			expected: false,
		},
		{
			name:     "ドットから始まるパス",
			backURL:  "../works",
			expected: false,
		},
		{
			name:     "バックスラッシュ（IEの脆弱性対策）",
			backURL:  "\\\\evil.com",
			expected: false,
		},

		// エッジケース - 紛らわしいが安全なケース
		{
			name:     "スラッシュのみ",
			backURL:  "/",
			expected: true,
		},
		{
			name:     "URLにドメイン名を含むパス",
			backURL:  "/redirect?url=http://example.com",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateBackURL(tt.backURL)
			if result != tt.expected {
				t.Errorf("ValidateBackURL(%q) = %v, want %v", tt.backURL, result, tt.expected)
			}
		})
	}
}

func TestGetSafeRedirectURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		backURL  string
		expected string
	}{
		// 有効なURLはそのまま返す
		{
			name:     "有効なパス",
			backURL:  "/works",
			expected: "/works",
		},
		{
			name:     "有効なパス（クエリパラメータ付き）",
			backURL:  "/oauth/authorize?client_id=xxx",
			expected: "/oauth/authorize?client_id=xxx",
		},
		{
			name:     "ルートパス",
			backURL:  "/",
			expected: "/",
		},

		// 無効なURLはデフォルト（"/"）を返す
		{
			name:     "空文字",
			backURL:  "",
			expected: "/",
		},
		{
			name:     "絶対URL",
			backURL:  "https://evil.com",
			expected: "/",
		},
		{
			name:     "プロトコル相対URL",
			backURL:  "//evil.com",
			expected: "/",
		},
		{
			name:     "JavaScriptスキーム",
			backURL:  "javascript:alert(1)",
			expected: "/",
		},
		{
			name:     "相対パス（スラッシュなし）",
			backURL:  "works",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := GetSafeRedirectURL(tt.backURL)
			if result != tt.expected {
				t.Errorf("GetSafeRedirectURL(%q) = %q, want %q", tt.backURL, result, tt.expected)
			}
		})
	}
}
