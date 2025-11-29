# 国際化（I18n）ガイド

このドキュメントは、Go版Annictでの国際化（Internationalization）のベストプラクティスを説明します。

## 概要

すべてのユーザー向けメッセージは**必ず国際化対応**します。

### 対応言語

- **日本語**（デフォルト）
- **英語**

### 翻訳ファイル

- `internal/i18n/locales/ja.toml` - 日本語翻訳
- `internal/i18n/locales/en.toml` - 英語翻訳

## 使用方法

### テンプレートでの使用

```templ
// pages/sign_in.templ
package pages

import (
    "context"
    "github.com/annict/annict/internal/templates"
)

templ SignIn(ctx context.Context, csrfToken string) {
    <div>
        // 基本的な翻訳
        <h2>{ templates.T(ctx, "sign_in_heading") }</h2>

        // ラベル
        <label for="email_username">
            { templates.T(ctx, "sign_in_email_username_label") }
        </label>

        // ボタン
        <button type="submit">{ templates.T(ctx, "sign_in_submit") }</button>
    </div>
}
```

### Goコードでの使用

```go
// internal/handler/sign_in.go
package handler

import (
    "github.com/annict/annict/internal/i18n"
)

func (h *Handler) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 翻訳を取得
    message := i18n.T(ctx, "sign_in_success")

    // フラッシュメッセージに設定
    sessionManager := session.GetSessionManager(r)
    sessionManager.SetFlash(ctx, "notice", message)

    http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

### プレースホルダー付き翻訳

```templ
// テンプレート
<p>{ templates.T(ctx, "watchers_count", map[string]any{"Count": work.WatchersCount}) }</p>
```

```toml
# ja.toml
[watchers_count]
description = "ウォッチ数の表示"
other = "{{.Count}} 人がウォッチ"

# en.toml
[watchers_count]
description = "Display watchers count"
other = "{{.Count}} watchers"
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

## 対象メッセージ

### 国際化が必要なもの

- ✅ ページタイトル、見出し
- ✅ ラベル、ボタンテキスト
- ✅ エラーメッセージ、成功メッセージ
- ✅ ヘルプテキスト、説明文
- ✅ バリデーションメッセージ

### 国際化が不要なもの

- ❌ ログメッセージ（開発者向け）
- ❌ panic メッセージ（想定外のエラー）
- ❌ 内部エラー（ユーザーに見せないエラー）
- ❌ コメント（コード内のコメント）

## 運用・開発者向けメッセージ

運用・開発者向けのメッセージは**日本語で統一**します（国際化不要）。

### ログメッセージ

```go
// ✅ OK: ログメッセージは日本語でOK（開発者向け）
slog.InfoContext(ctx, "パスワードリセット申請を受け付けました", "user_id", userID)
slog.ErrorContext(ctx, "データベース接続エラー", "error", err)
```

### panicメッセージ

```go
// ✅ OK: panicメッセージは日本語でOK
panic("設定ファイルの読み込みに失敗しました")
```

### 内部エラー

```go
// ✅ OK: 内部エラーは日本語でOK（開発者向け）
return fmt.Errorf("トークンのハッシュ化に失敗: %w", err)
```

## 判断基準

```go
// ❌ NG: ユーザー向けメッセージなのに日本語ハードコード
http.Error(w, "メールアドレスを入力してください", http.StatusBadRequest)

// ✅ OK: ユーザー向けメッセージは国際化
errors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))

// ✅ OK: ログメッセージは日本語でOK（開発者向け）
slog.InfoContext(ctx, "パスワードリセットトークンを生成しました", "user_id", userID)

// ✅ OK: 内部エラーは日本語でOK（開発者向け）
return fmt.Errorf("セッションの作成に失敗: %w", err)
```

## 翻訳の追加手順

### 1. メッセージIDを決定

命名規則: `{機能名}_{種別}_{詳細}` の形式

例:
- `password_reset_email_required`
- `sign_in_heading`
- `popular_anime_title`

### 2. ja.tomlに追加

```toml
# internal/i18n/locales/ja.toml

[password_reset_title]
description = "パスワードリセットページのタイトル"
other = "パスワードリセット"

[password_reset_email_required]
description = "メールアドレス必須エラー"
other = "メールアドレスを入力してください"

[password_reset_email_invalid]
description = "メールアドレス形式エラー"
other = "メールアドレスの形式が正しくありません"
```

### 3. en.tomlに追加

```toml
# internal/i18n/locales/en.toml

[password_reset_title]
description = "Password reset page title"
other = "Password Reset"

[password_reset_email_required]
description = "Email required error"
other = "Please enter your email address"

[password_reset_email_invalid]
description = "Email format error"
other = "Email address format is invalid"
```

### 4. テンプレートまたはGoコードで使用

```templ
// テンプレート
<h2>{ templates.T(ctx, "password_reset_title") }</h2>
```

```go
// Goコード
errors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))
```

## 翻訳の命名規則

### ページタイトル

**形式**: `{page_name}_title`

例:
- `password_reset_title` - パスワードリセット
- `sign_in_title` - ログイン
- `popular_anime_title` - 人気アニメ

### 見出し

**形式**: `{page_name}_heading`

例:
- `sign_in_heading` - ログイン
- `password_reset_heading` - パスワードリセット

### ラベル

**形式**: `{page_name}_{field_name}_label`

例:
- `sign_in_email_username_label` - メールアドレスまたはユーザー名
- `password_reset_email_label` - メールアドレス

### ボタン

**形式**: `{page_name}_submit`

例:
- `password_reset_submit` - 送信
- `sign_in_submit` - ログイン

### エラーメッセージ

**形式**: `{page_name}_error_{detail}`

例:
- `password_reset_email_required` - メールアドレスを入力してください
- `password_reset_email_invalid` - メールアドレスの形式が正しくありません
- `sign_in_error_invalid_credentials` - メールアドレスまたはパスワードが正しくありません

### 成功メッセージ

**形式**: `{page_name}_success`

例:
- `password_reset_success` - パスワードリセットメールを送信しました
- `sign_in_success` - ログインしました

### ヘルプテキスト

**形式**: `{page_name}_{detail}_help`

例:
- `password_reset_email_not_received_help` - メールが届かない場合は迷惑メールフォルダをご確認ください

## バリデーションエラーメッセージの国際化

Request DTOのバリデーションメッセージも必ず国際化します。

### 例

```go
// ❌ NG: ハードコードされた日本語メッセージ
errors.AddFieldError("email", "メールアドレスを入力してください")

// ✅ OK: I18n経由で翻訳されたメッセージ
errors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))
```

## ロケールの取得と設定

### ロケールの取得

```go
// 現在のロケールを取得
locale := i18n.GetLocale(ctx)  // "ja" または "en"
```

### ロケールの設定

```go
// ミドルウェアで自動的に設定される
// Accept-Languageヘッダーまたはクッキーから判定
```

### テンプレートでロケールを使用

```templ
templ Default(ctx context.Context, meta viewmodel.PageMeta, content templ.Component) {
    <!DOCTYPE html>
    <html lang={ templates.Locale(ctx) }>
        <head>
            <!-- ... -->
        </head>
        <body>
            @content
        </body>
    </html>
}
```

## テスト

### 翻訳のテスト

```go
func TestTranslations(t *testing.T) {
    tests := []struct {
        name     string
        locale   string
        key      string
        expected string
    }{
        {
            name:     "Japanese password reset title",
            locale:   "ja",
            key:      "password_reset_title",
            expected: "パスワードリセット",
        },
        {
            name:     "English password reset title",
            locale:   "en",
            key:      "password_reset_title",
            expected: "Password Reset",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            ctx = i18n.WithLocale(ctx, tt.locale)

            got := i18n.T(ctx, tt.key)
            if got != tt.expected {
                t.Errorf("i18n.T(%q) = %q, want %q", tt.key, got, tt.expected)
            }
        })
    }
}
```

### プレースホルダー付き翻訳のテスト

```go
func TestTranslationsWithPlaceholders(t *testing.T) {
    ctx := context.Background()
    ctx = i18n.WithLocale(ctx, "ja")

    got := i18n.T(ctx, "watchers_count", map[string]any{"Count": 100})
    expected := "100 人がウォッチ"

    if got != expected {
        t.Errorf("got %q, want %q", got, expected)
    }
}
```

## ベストプラクティス

### 1. 翻訳キーは具体的に

```toml
# ❌ Bad: 汎用的すぎる
[error]
other = "エラーが発生しました"

# ✅ Good: 具体的
[sign_in_error_invalid_credentials]
other = "メールアドレスまたはパスワードが正しくありません"
```

### 2. 翻訳キーに日本語を含めない

```toml
# ❌ Bad
[パスワードリセット_タイトル]
other = "パスワードリセット"

# ✅ Good
[password_reset_title]
other = "パスワードリセット"
```

### 3. descriptionを必ず記述

```toml
# ❌ Bad: descriptionがない
[password_reset_title]
other = "パスワードリセット"

# ✅ Good: descriptionがある
[password_reset_title]
description = "パスワードリセットページのタイトル"
other = "パスワードリセット"
```

### 4. 複数形の扱い

```toml
# 日本語は複数形がないため、プレースホルダーで対応
[watchers_count]
description = "ウォッチ数の表示"
other = "{{.Count}} 人がウォッチ"

# 英語は単数形・複数形を区別する場合がある
[watchers_count]
description = "Display watchers count"
one = "{{.Count}} watcher"
other = "{{.Count}} watchers"
```

## トラブルシューティング

### 翻訳が表示されない

1. **翻訳キーが間違っている**: ja.toml/en.tomlのキーを確認
2. **ロケールが設定されていない**: `i18n.WithLocale(ctx, "ja")` を確認
3. **TOMLファイルの構文エラー**: TOMLファイルをバリデート

### 翻訳が英語になってしまう

1. **日本語翻訳が未定義**: ja.tomlに翻訳を追加
2. **デフォルトロケールが英語**: ミドルウェアの設定を確認
3. **Accept-Languageヘッダー**: ブラウザの言語設定を確認

### プレースホルダーが展開されない

1. **構文エラー**: `{{.Key}}` の形式を確認
2. **データが渡されていない**: `map[string]any` の内容を確認
3. **キー名の大文字小文字**: 大文字で始まることを確認（例: `{{.Count}}`）
