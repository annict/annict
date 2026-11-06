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

		// 無効なケース - 絶対URL (オープンリダイレクト攻撃)
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

		// 無効なケース - プロトコル相対URL (オープンリダイレクト攻撃)
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
			name:     "相対パス (スラッシュなし)",
			backURL:  "works",
			expected: false,
		},
		{
			name:     "ドットから始まるパス",
			backURL:  "../works",
			expected: false,
		},
		{
			name:     "バックスラッシュ (IEの脆弱性対策)",
			backURL:  "\\\\evil.com",
			expected: false,
		},
		{name: "スラッシュとバックスラッシュ", backURL: `/\example.org/phishing`, expected: false},
		{name: "ドットセグメントの後のバックスラッシュ", backURL: `/db/../\example.org/phishing`, expected: false},
		{name: "水平タブ", backURL: "/\t/example.org", expected: false},
		{name: "改行", backURL: "/\n/example.org", expected: false},
		{name: "復帰", backURL: "/\r/example.org", expected: false},
		{name: "NULL文字", backURL: "/works\x00", expected: false},
		{name: "DEL文字", backURL: "/works\x7f", expected: false},
		{name: "不正なパーセントエンコード", backURL: "/works/%", expected: false},
		{name: "フラグメント内の水平タブとドットセグメント", backURL: "/a#/../\t/example.org/phishing", expected: false},
		{name: "フラグメント内の改行", backURL: "/works#\n", expected: false},

		// エッジケース - 紛らわしいが安全なケース
		{name: "フラグメント付きのパス", backURL: "/works?q=1#frag", expected: true},
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
				t.Errorf("ValidateBackURL(%q) = %v、期待値 = %v", tt.backURL, result, tt.expected)
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
			name:     "有効なパス (クエリパラメータ付き)",
			backURL:  "/oauth/authorize?client_id=xxx",
			expected: "/oauth/authorize?client_id=xxx",
		},
		{
			name:     "ルートパス",
			backURL:  "/",
			expected: "/",
		},

		// 無効なURLはデフォルト ("/") を返す
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
			name:     "相対パス (スラッシュなし)",
			backURL:  "works",
			expected: "/",
		},
		{name: "バックスラッシュで外部URLになるパス", backURL: `/\example.org/phishing`, expected: "/"},
		{name: "水平タブを取り除くと外部URLになるパス", backURL: "/\t/example.org", expected: "/"},
		{name: "フラグメントを正規化すると外部URLになるパス", backURL: "/a#/../\t/example.org/phishing", expected: "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := GetSafeRedirectURL(tt.backURL)
			if result != tt.expected {
				t.Errorf("GetSafeRedirectURL(%q) = %q、期待値 = %q", tt.backURL, result, tt.expected)
			}
		})
	}
}

// TestValidateDBReturnURLはAnnict DBの確認画面が戻り先の値に適用する許可リストを
// 検証する。Annict DBのパスは受け入れ、サイト内の別のパスを含めそれ以外は拒否する。
func TestValidateDBReturnURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		returnTo string
		expected bool
	}{
		{name: "Annict DBのトップ", returnTo: "/db", expected: true},
		{name: "Annict DBの一覧", returnTo: "/db/works", expected: true},
		{name: "クエリパラメータ付きの検索結果", returnTo: "/db/search?q=%E6%A4%9C%E7%B4%A2", expected: true},
		{name: "フラグメント付きのパス", returnTo: "/db/works#row-1", expected: true},
		{name: "クエリとフラグメント付きのパス", returnTo: "/db/works?page=2#row-1", expected: true},
		{name: "末尾にスラッシュのあるパス", returnTo: "/db/works/", expected: true},
		{name: "クエリ内のドットセグメント", returnTo: "/db/search?q=/../settings", expected: true},
		{name: "クエリ内のエンコードしたスラッシュ", returnTo: "/db/search?q=a%2fb", expected: true},
		{name: "パスの末尾のエンコードしたスペース", returnTo: "/db/..%20", expected: true},
		{name: "クエリ内のエンコードしたスペース", returnTo: "/db/search?q=hello%20world", expected: true},
		{name: "先頭の半角スペース", returnTo: " /db/works", expected: false},
		{name: "末尾の半角スペースで隠れたドットセグメント", returnTo: "/db/.. ", expected: false},
		{name: "末尾の複数の半角スペース", returnTo: "/db/..  ", expected: false},
		{name: "パスの末尾の半角スペース", returnTo: "/db/works ", expected: false},
		{name: "フラグメントの末尾の半角スペース", returnTo: "/db/works#row-1 ", expected: false},
		{name: "正規化するとAnnict DBの中になるドットセグメント", returnTo: "/db/works/../people?q=test", expected: false},
		{name: "空のセグメント", returnTo: "/db//works", expected: false},
		{name: "空文字", returnTo: "", expected: false},
		{name: "Annict DBの外", returnTo: "/settings", expected: false},
		{name: "接頭辞だけ一致する別のパス", returnTo: "/dbx/works", expected: false},
		{name: "相対パス", returnTo: "db/works", expected: false},
		{name: "外部URL", returnTo: "https://example.com/db/works", expected: false},
		{name: "プロトコル相対URL", returnTo: "//example.com/db/works", expected: false},
		{name: "バックスラッシュ始まり", returnTo: "/\\example.com", expected: false},
		{name: "正規化後に外部URLになるパス", returnTo: `/db/../\example.org/phishing`, expected: false},
		{name: "正規化後にAnnict DBの外になるパス", returnTo: "/db/../settings", expected: false},
		{name: "エンコードされたドットセグメント", returnTo: "/db/%2e%2e/settings", expected: false},
		{name: "混在するドットセグメント", returnTo: "/db/.%2E/settings", expected: false},
		{name: "フラグメントを正規化すると外部URLになるパス", returnTo: "/db/works#/../../\t/example.org/phishing", expected: false},
		{name: "フラグメントを正規化するとAnnict DBの外になるパス", returnTo: "/db/works#/../../settings", expected: false},
		{name: "サーバーとブラウザの正規化でAnnict DBの外になるパス", returnTo: "/db/a#/../%2e%2e/settings", expected: false},
		{name: "正規化でAnnict DBの外になるクエリ付きパス", returnTo: "/db/a#/../%2e%2e/settings?q=1", expected: false},
		{name: "エンコードしたスラッシュとドットセグメント", returnTo: "/db/x%2fy/%2e%2e/%2e%2e/settings", expected: false},
		{name: "大文字でエンコードしたスラッシュとクエリ", returnTo: "/db/x%2Fy/%2e%2e/%2e%2e/settings?q=1", expected: false},
		{name: "フラグメントの直前のドットセグメント", returnTo: "/db/..#settings", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateDBReturnURL(tt.returnTo); got != tt.expected {
				t.Errorf("ValidateDBReturnURL(%q) = %v、期待値 = %v", tt.returnTo, got, tt.expected)
			}
		})
	}
}

// TestGetSafeDBReturnURLは、受け入れた値はそのまま返し、拒否した値は呼び出し元の
// フォールバックに差し替えることを検証する。
func TestGetSafeDBReturnURL(t *testing.T) {
	t.Parallel()

	const fallback = "/db/works"

	tests := []struct {
		name     string
		returnTo string
		expected string
	}{
		{name: "Annict DBのパスはそのまま", returnTo: "/db/search?q=test", expected: "/db/search?q=test"},
		{name: "空文字はフォールバック", returnTo: "", expected: fallback},
		{name: "Annict DBの外はフォールバック", returnTo: "/settings", expected: fallback},
		{name: "外部URLはフォールバック", returnTo: "https://example.com/", expected: fallback},
		{name: "バックスラッシュはフォールバック", returnTo: `/db/../\example.org/phishing`, expected: fallback},
		{name: "Annict DBの外に正規化されるパスはフォールバック", returnTo: "/db/../settings", expected: fallback},
		{name: "エンコードされたドットセグメントはフォールバック", returnTo: "/db/%2e%2e/settings", expected: fallback},
		{name: "フラグメント付きのパスはそのまま", returnTo: "/db/works#row-1", expected: "/db/works#row-1"},
		{name: "パスの末尾のエンコードしたスペースはそのまま", returnTo: "/db/..%20", expected: "/db/..%20"},
		{name: "クエリ内のエンコードしたスペースはそのまま", returnTo: "/db/search?q=hello%20world", expected: "/db/search?q=hello%20world"},
		{name: "末尾の半角スペースで隠れたドットセグメントはフォールバック", returnTo: "/db/.. ", expected: fallback},
		{name: "フラグメントを正規化すると外部URLになるパスはフォールバック", returnTo: "/db/works#/../../\t/example.org/phishing", expected: fallback},
		{name: "フラグメントを正規化するとAnnict DBの外になるパスはフォールバック", returnTo: "/db/works#/../../settings", expected: fallback},
		{name: "サーバーとブラウザの正規化でAnnict DBの外になるパスはフォールバック", returnTo: "/db/a#/../%2e%2e/settings", expected: fallback},
		{name: "エンコードしたスラッシュでAnnict DBの外になるパスはフォールバック", returnTo: "/db/x%2fy/%2e%2e/%2e%2e/settings", expected: fallback},
		{name: "フラグメントの直前のドットセグメントはフォールバック", returnTo: "/db/..#settings", expected: fallback},
		{name: "ドットセグメントを含むパスはフォールバック", returnTo: "/db/works/../people?q=test", expected: fallback},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSafeDBReturnURL(tt.returnTo, fallback); got != tt.expected {
				t.Errorf("GetSafeDBReturnURL(%q, %q) = %q、期待値 = %q", tt.returnTo, fallback, got, tt.expected)
			}
		})
	}
}
