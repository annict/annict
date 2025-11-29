# templ テンプレートガイド

このドキュメントは、Go版Annictで使用している [templ](https://templ.guide/) テンプレートエンジンの使い方を説明します。

## 概要

Go版では、型安全なテンプレートエンジン **templ** を使用してHTMLを生成します。

### なぜtemplを使うのか？

- **型安全性**: コンパイル時に型チェックとエラー検出
- **IDEサポート**: 自動補完、リファクタリング、Go to Definitionが使える
- **パフォーマンス**: Goコードにコンパイルされるため高速
- **保守性**: テンプレートがGoコードとして生成されるため、静的解析やテストが容易

## ファイル配置

```
/workspace/go/internal/templates/
├── layouts/              # レイアウトテンプレート
│   ├── default.templ
│   ├── default_templ.go  # 自動生成
│   ├── simple.templ
│   └── simple_templ.go   # 自動生成
├── components/           # 再利用可能なコンポーネント
│   ├── head.templ
│   ├── head_templ.go     # 自動生成
│   ├── flash.templ
│   ├── flash_templ.go    # 自動生成
│   ├── form_errors.templ
│   └── form_errors_templ.go  # 自動生成
├── pages/                # ページテンプレート
│   ├── sign_in.templ
│   ├── sign_in_templ.go  # 自動生成
│   ├── works/
│   │   ├── popular.templ
│   │   └── popular_templ.go  # 自動生成
│   └── errors/
│       ├── 502.templ
│       └── 502_templ.go  # 自動生成
├── emails/               # メールテンプレート
│   └── password_reset/
│       ├── ja_html.templ
│       ├── ja_html_templ.go  # 自動生成
│       ├── en_html.templ
│       └── en_html_templ.go  # 自動生成
└── helper.go             # テンプレートヘルパー関数
```

### 命名規則

- **templファイル**: 小文字のスネークケース（例: `sign_in.templ`, `popular.templ`）
- **生成ファイル**: `*_templ.go` ファイルは自動生成されるため、**手動編集禁止**

## 基本構文

### シンプルなテンプレート

```templ
// パッケージ宣言（必須）
package pages

// インポート
import (
    "context"
    "github.com/annict/annict/internal/templates"
)

// テンプレートコンポーネントの定義
templ SignIn(ctx context.Context, csrfToken string) {
    <div>
        // 翻訳の呼び出し
        <h2>{ templates.T(ctx, "sign_in_heading") }</h2>

        // フォーム
        <form method="POST" action="/sign_in">
            <input type="hidden" name="csrf_token" value={ csrfToken } />

            // ラベルと入力フィールド
            <label for="email_username">{ templates.T(ctx, "sign_in_email_username_label") }</label>
            <input type="text" id="email_username" name="email_username" />

            <button type="submit">{ templates.T(ctx, "sign_in_submit") }</button>
        </form>
    </div>
}
```

### 条件分岐

```templ
templ SignIn(ctx context.Context, formErrors *session.FormErrors) {
    <div>
        // if文
        if formErrors != nil && formErrors.HasErrors() {
            <div class="alert alert-danger">
                { templates.T(ctx, "form_errors_found") }
            </div>
        }

        // if-else
        if user != nil {
            <p>ようこそ、{ user.Username }さん</p>
        } else {
            <p>ログインしてください</p>
        }
    </div>
}
```

### 繰り返し

```templ
templ PopularWorks(ctx context.Context, works []viewmodel.Work) {
    <div>
        <h2>{ templates.T(ctx, "popular_anime") }</h2>

        // for文
        for _, work := range works {
            <div class="work-card">
                <h3>{ work.Title }</h3>
                <p>{ fmt.Sprintf("%d watchers", work.WatchersCount) }</p>

                // ネストした条件分岐
                if work.ImageURL != "" {
                    <img src={ work.ImageURL } alt={ work.Title } />
                }
            </div>
        }
    </div>
}
```

## レイアウトの継承

templではレイアウトの継承を**コンポーネント化**で実現します。

### レイアウトテンプレート

```templ
// layouts/default.templ
package layouts

import (
    "context"
    "github.com/a-h/templ"
    "github.com/annict/annict/internal/repository"
    "github.com/annict/annict/internal/templates/components"
    "github.com/annict/annict/internal/viewmodel"
)

templ Default(ctx context.Context, meta viewmodel.PageMeta, user *repository.GetUserByIDRow, content templ.Component) {
    <!DOCTYPE html>
    <html lang={ templates.Locale(ctx) }>
        <head>
            @components.Head(meta)
        </head>
        <body>
            <header>
                <nav>
                    // ナビゲーション
                </nav>
            </header>

            <main>
                // コンテンツを挿入
                @content
            </main>

            <footer>
                // フッター
            </footer>
        </body>
    </html>
}
```

### ページテンプレート

```templ
// pages/sign_in.templ
package pages

import (
    "context"
    "github.com/annict/annict/internal/templates"
)

templ SignIn(ctx context.Context, csrfToken string) {
    <div>
        <h2>{ templates.T(ctx, "sign_in_heading") }</h2>
        <form method="POST" action="/sign_in">
            // フォームフィールド
        </form>
    </div>
}
```

### ハンドラーでの使用

```go
func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // ページメタデータを作成
    meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
    meta.SetTitle(ctx, "sign_in_title")

    // ログインユーザーを取得（未ログインの場合はnil）
    user := authMiddleware.GetUserFromContext(ctx)

    // CSRFトークンを取得
    csrfToken := middleware.GetCSRFToken(r)

    // レイアウトにコンテンツを渡してレンダリング
    layouts.Default(ctx, meta, user, pages.SignIn(ctx, csrfToken)).Render(ctx, w)
}
```

## コンポーネントの再利用

他のコンポーネントを呼び出すには `@` を使用します。

### コンポーネントの定義

```templ
// components/form_errors.templ
package components

import (
    "context"
    "github.com/annict/annict/internal/session"
    "github.com/annict/annict/internal/templates"
)

templ FormErrors(ctx context.Context, formErrors *session.FormErrors) {
    if formErrors != nil && formErrors.HasErrors() {
        <div class="alert alert-danger">
            <p><strong>{ templates.T(ctx, "form_errors_found") }</strong></p>
            <ul>
                for _, err := range formErrors.GetFieldErrors() {
                    <li>{ err }</li>
                }
            </ul>
        </div>
    }
}
```

### コンポーネントの呼び出し

```templ
// pages/sign_in.templ
package pages

import (
    "context"
    "github.com/annict/annict/internal/session"
    "github.com/annict/annict/internal/templates/components"
)

templ SignIn(ctx context.Context, formErrors *session.FormErrors) {
    <div>
        // コンポーネントを呼び出し
        @components.FormErrors(ctx, formErrors)

        // フォーム
        <form method="POST" action="/sign_in">
            // フォームフィールド
        </form>
    </div>
}
```

## 国際化対応

テンプレート内で翻訳を使用する場合は、`templates.T(ctx, "message_id")` を呼び出します。

### 基本的な翻訳

```templ
// シンプルな翻訳
<h2>{ templates.T(ctx, "sign_in_heading") }</h2>

// ラベル
<label>{ templates.T(ctx, "sign_in_email_username_label") }</label>

// ボタン
<button type="submit">{ templates.T(ctx, "sign_in_submit") }</button>
```

### プレースホルダー付き翻訳

```templ
// 翻訳ファイル (ja.toml)
// [watchers_count]
// other = "{{.Count}} 人がウォッチ"

// テンプレート
<p>{ templates.T(ctx, "watchers_count", map[string]any{"Count": work.WatchersCount}) }</p>
```

### 条件に応じた翻訳

```templ
templ WorkCard(ctx context.Context, work viewmodel.Work) {
    <div>
        <h3>{ work.Title }</h3>

        // シーズン表示
        if work.SeasonYear != nil && work.SeasonName != nil {
            <p>
                { fmt.Sprintf("%d年", *work.SeasonYear) }
                switch *work.SeasonName {
                    case "spring":
                        { templates.T(ctx, "season_spring") }
                    case "summer":
                        { templates.T(ctx, "season_summer") }
                    case "autumn":
                        { templates.T(ctx, "season_autumn") }
                    case "winter":
                        { templates.T(ctx, "season_winter") }
                }
            </p>
        }
    </div>
}
```

## ヘルパー関数

`internal/templates/helper.go` にテンプレートで使用するヘルパー関数を定義しています。

### 利用可能なヘルパー

```go
// 翻訳を取得
templates.T(ctx, "message_id")
templates.T(ctx, "message_id", map[string]any{"Key": value})

// 現在のロケールを取得
templates.Locale(ctx)  // "ja" または "en"

// ポインタの参照外し
templates.Deref(work.SeasonYear)  // *int32 -> int32

// アイコンを表示（SVG）
templates.Icon("check", "icon-sm")
```

## コード生成

templファイルを編集したら、以下のコマンドでGoコードを生成します。

```sh
# 手動生成
templ generate

# air（ホットリロード）使用時は自動的に実行される
```

生成された `*_templ.go` ファイルは自動生成されるため、**絶対に手動編集しないでください**。

## テスト

### ハンドラーテスト

```go
func TestSignInPage(t *testing.T) {
    // テストDBとトランザクションをセットアップ
    db, tx := testutil.SetupTestDB(t)
    queries := repository.New(db).WithTx(tx)

    // 設定とハンドラーを作成
    cfg := &config.Config{Domain: "localhost"}
    handler := &Handler{queries: queries, cfg: cfg}

    // HTTPリクエストを作成
    req := httptest.NewRequest("GET", "/sign_in", nil)
    req = req.WithContext(i18n.WithLocale(req.Context(), "ja"))
    rr := httptest.NewRecorder()

    // ハンドラーを実行
    handler.SignIn(rr, req)

    // ステータスコードの確認
    if rr.Code != http.StatusOK {
        t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
    }

    // Content-Typeヘッダーの確認
    if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
        t.Errorf("wrong content type: got %v want text/html", contentType)
    }

    // HTML出力の検証
    body := rr.Body.String()
    expectedChecks := []string{
        "ログイン",                              // タイトル
        "メールアドレスまたはユーザー名",        // ラベル
        "<form",                                 // フォームタグ
        "csrf_token",                            // CSRFトークン
    }

    for _, expected := range expectedChecks {
        if !strings.Contains(body, expected) {
            t.Errorf("response doesn't contain: %q", expected)
        }
    }
}
```

### コンポーネント単体テスト

```go
func TestFormErrorsComponent(t *testing.T) {
    ctx := context.Background()
    ctx = i18n.WithLocale(ctx, "ja")

    // エラーを含むFormErrorsを作成
    formErrors := &session.FormErrors{}
    formErrors.AddFieldError("email", "メールアドレスを入力してください")

    // コンポーネントをレンダリング
    var buf bytes.Buffer
    err := components.FormErrors(ctx, formErrors).Render(ctx, &buf)
    if err != nil {
        t.Fatalf("failed to render: %v", err)
    }

    // HTML出力の検証
    html := buf.String()
    if !strings.Contains(html, "メールアドレスを入力してください") {
        t.Errorf("error message not found in output")
    }
    if !strings.Contains(html, "alert-danger") {
        t.Errorf("alert class not found in output")
    }
}
```

### テーブル駆動テスト（多言語対応）

```go
func TestSignInPageMultipleLocales(t *testing.T) {
    tests := []struct {
        name     string
        locale   string
        expected string
    }{
        {
            name:     "Japanese",
            locale:   "ja",
            expected: "ログイン",
        },
        {
            name:     "English",
            locale:   "en",
            expected: "Sign In",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db, tx := testutil.SetupTestDB(t)
            queries := repository.New(db).WithTx(tx)

            cfg := &config.Config{Domain: "localhost"}
            handler := &Handler{queries: queries, cfg: cfg}

            req := httptest.NewRequest("GET", "/sign_in", nil)
            req = req.WithContext(i18n.WithLocale(req.Context(), tt.locale))
            rr := httptest.NewRecorder()

            handler.SignIn(rr, req)

            if !strings.Contains(rr.Body.String(), tt.expected) {
                t.Errorf("expected %q not found in response", tt.expected)
            }
        })
    }
}
```

## ベストプラクティス

### 型安全性を活用

```templ
// ✅ Good: 引数の型を明示的に指定
templ WorkCard(ctx context.Context, work viewmodel.Work, showCast bool) {
    <div>
        <h3>{ work.Title }</h3>
        if showCast {
            // キャスト情報を表示
        }
    </div>
}
```

### コンポーネントの再利用

```templ
// ✅ Good: 共通のUIパターンはコンポーネントとして切り出す
@components.Button(ctx, "sign_in_submit", "/sign_in")
@components.FormErrors(ctx, formErrors)
@components.Flash(ctx, flash)
```

### 国際化を徹底

```templ
// ❌ Bad: ハードコードされた日本語
<h2>ログイン</h2>

// ✅ Good: 翻訳を使用
<h2>{ templates.T(ctx, "sign_in_heading") }</h2>
```

### シンプルなロジック

```templ
// ❌ Bad: 複雑なロジックをテンプレートに書く
templ WorkCard(ctx context.Context, work viewmodel.Work) {
    <div>
        // 複雑な計算やビジネスロジック
        { calculateComplexValue(work) }
    </div>
}

// ✅ Good: 複雑なロジックはGoコード側（viewmodel）で処理
templ WorkCard(ctx context.Context, work viewmodel.Work) {
    <div>
        // viewmodelで事前に計算された値を表示
        { work.DisplayValue }
    </div>
}
```

### セキュリティ

```templ
// ✅ Good: templは自動的にエスケープ処理を行う
<p>{ user.Comment }</p>  // XSS対策済み

// ⚠️ 注意: templ.Raw()を使う場合は信頼できるソースのみ
<div>{ templ.Raw(trustedHTMLContent) }</div>
```

## 参考資料

- [templ公式サイト](https://templ.guide/)
- [templ GitHubリポジトリ](https://github.com/a-h/templ)
- [templ Examples](https://github.com/a-h/templ/tree/main/examples)
