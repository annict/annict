# templ への移行 設計書

## 概要

Go版Annictで現在使用しているGo標準の`html/template`を、型安全なテンプレートエンジン[templ](https://github.com/a-h/templ)に移行します。

**目的**:

- **型安全性の向上**: コンパイル時に型チェックとエラー検出を行い、実行時エラーを防ぐ
- **開発体験の向上**: IDEの自動補完、リファクタリング、Go to Definitionなどの機能をフル活用
- **保守性の向上**: テンプレートがGoコードとして生成されるため、静的解析やテストが容易
- **パフォーマンスの向上**: Go言語にコンパイルされるため、実行時のオーバーヘッドが少ない
- **XSS対策の強化**: templがデフォルトでエスケープ処理を行うため、セキュリティが向上

**背景**:

- 現在のテンプレートシステムでは、実行時までエラーが検出されない
- テンプレート内の変数名や関数名のタイポが実行時エラーの原因になる
- テンプレートの国際化対応でヘルパー関数（`FuncMap`）に依存しており、型安全性が低い
- templを使用することで、これらの課題を解決し、よりGoらしい実装が可能になる

## 要件

### 機能要件

- **既存テンプレートの移行**: 現在の16個のHTMLテンプレートファイルをすべて`.templ`形式に変換する
- **国際化対応の維持**: `i18n.T(ctx, "message_id")`を使用した翻訳機能を維持する
- **レイアウトの継承**: レイアウトテンプレート（`default`, `simple`）とコンテンツテンプレートの構造を維持する
- **パーシャルの再利用**: `head`, `flash`, `form_errors`などのパーシャルコンポーネントを維持する
- **ヘルパー関数の移行**: `icon`, `deref`, `dict`, `locale`などのヘルパー関数をtempl対応に変換する
- **ビルドシステムの統合**: `templ generate`をビルドプロセスに組み込む
- **テストの動作保証**: 既存のテンプレートレンダリングテストが正常に動作すること

### 非機能要件

#### パフォーマンス

- テンプレートレンダリングの速度が現在と同等以上であること
- ビルド時間が大幅に増加しないこと（templ生成の追加オーバーヘッドは許容範囲内）

#### 保守性

- テンプレートコードが読みやすく、理解しやすいこと
- 国際化対応のパターンが統一されていること
- コンポーネントの再利用性が高いこと

#### 開発体験

- ホットリロード（air）が正常に動作すること
- IDEの自動補完とエラー検出が機能すること
- コンパイル時に型エラーが検出されること

## 設計

### 技術スタック

- **templ**: v0.3.960（最新安定版）
- **Go**: 1.25.1（現在のバージョンを維持）
- **ビルドツール**: Makefile, air（ホットリロード）

### アーキテクチャ

#### ディレクトリ構造

```
/workspace/go/
├── internal/
│   ├── templates/
│   │   ├── layouts/
│   │   │   ├── default.templ          # デフォルトレイアウト
│   │   │   ├── default_templ.go       # 生成されたGoコード（自動生成）
│   │   │   ├── simple.templ           # シンプルレイアウト
│   │   │   └── simple_templ.go        # 生成されたGoコード（自動生成）
│   │   ├── components/                # パーシャルをcomponentsに改名
│   │   │   ├── head.templ
│   │   │   ├── head_templ.go
│   │   │   ├── flash.templ
│   │   │   ├── flash_templ.go
│   │   │   ├── form_errors.templ
│   │   │   └── form_errors_templ.go
│   │   ├── pages/                     # ページテンプレート（新規ディレクトリ）
│   │   │   ├── works/
│   │   │   │   ├── popular.templ
│   │   │   │   └── popular_templ.go
│   │   │   ├── sign_in.templ
│   │   │   ├── sign_in_templ.go
│   │   │   ├── password/
│   │   │   │   ├── reset.templ
│   │   │   │   ├── reset_templ.go
│   │   │   │   ├── edit.templ
│   │   │   │   ├── edit_templ.go
│   │   │   │   ├── reset_sent.templ
│   │   │   │   └── reset_sent_templ.go
│   │   │   └── errors/
│   │   │       ├── 502.templ
│   │   │       └── 502_templ.go
│   │   ├── emails/
│   │   │   └── password_reset/
│   │   │       ├── ja_html.templ
│   │   │       ├── ja_html_templ.go
│   │   │       ├── en_html.templ
│   │   │       ├── en_html_templ.go
│   │   │       ├── ja_text.templ
│   │   │       ├── ja_text_templ.go
│   │   │       ├── en_text.templ
│   │   │       └── en_text_templ.go
│   │   └── helper.go                   # テンプレートヘルパー関数（i18n.T, iconなど）
│   └── handler/
│       ├── handler.go
│       └── template_loader.go          # 削除予定（templ移行後は不要）
```

#### 国際化対応の設計

templでは`html/template`の`FuncMap`が使えないため、国際化対応を以下のように変更します：

**Before（html/template）**:

```html
<h2>{{t "popular_anime"}}</h2>
```

**After（templ）**:

```templ
templ PopularWorks(ctx context.Context, works []viewmodel.Work) {
  <h2>{ templates.T(ctx, "popular_anime") }</h2>
}
```

#### レイアウトの継承

templではレイアウトの継承を**コンポーネント化**で実現します：

**Before（html/template）**:

```html
{{define "default"}}
<!DOCTYPE html>
<html>
  <body>
    {{template "content" .}}
  </body>
</html>
{{end}}
```

**After（templ）**:

```templ
// layouts/default.templ
package layouts

templ Default(ctx context.Context, meta viewmodel.PageMeta, user *repository.GetUserByIDRow, content templ.Component) {
  <!DOCTYPE html>
  <html lang={ templates.Locale(ctx) }>
    <head>
      @components.Head(meta)
    </head>
    <body>
      <header>
        <!-- ナビゲーション -->
      </header>
      <main>
        @content
      </main>
      <footer>
        <!-- フッター -->
      </footer>
    </body>
  </html>
}
```

```templ
// pages/works/popular.templ
package works

templ Popular(ctx context.Context, works []viewmodel.Work) {
  <div>
    <h2>{ templates.T(ctx, "popular_anime") }</h2>
    <!-- 作品リスト -->
  </div>
}
```

ハンドラーでの使用：

```go
func (h *Handler) PopularWorks(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  works, _ := h.queries.GetPopularWorks(ctx)

  meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
  meta.SetTitle(ctx, "popular_anime")
  user := authMiddleware.GetUserFromContext(ctx)

  // レイアウトにコンテンツを渡す
  layouts.Default(ctx, meta, user, works.Popular(ctx, worksView)).Render(ctx, w)
}
```

#### ヘルパー関数の移行

`internal/templates/helper.go`を以下のように変更：

```go
package templates

import (
  "context"
  "github.com/annict/annict/internal/i18n"
)

// T は翻訳を取得する（templ用）
func T(ctx context.Context, messageID string, data ...map[string]any) string {
  return i18n.T(ctx, messageID, data...)
}

// Locale は現在のロケールを取得する
func Locale(ctx context.Context) string {
  return i18n.GetLocale(ctx)
}

// Deref はポインタを参照外しする
func Deref[T any](v *T) T {
  if v != nil {
    return *v
  }
  var zero T
  return zero
}

// Icon はアイコン名からSVGを返す（templ.ComponentWrapper経由で返す）
func Icon(name string, class ...string) templ.Component {
  // SVGをtemplコンポーネントとして定義
}
```

### テスト戦略

- **単体テスト**: 各templコンポーネントが正しくレンダリングされることを確認
- **統合テスト**: レイアウトとコンテンツの組み合わせが正しく動作することを確認
- **リグレッションテスト**: 既存のテンプレートレンダリングテストを維持し、HTML出力が変わらないことを確認

## タスクリスト

### フェーズ 1: 環境構築とプロトタイプ

- [x] **1-1**: templのインストールとセットアップ

  - `go get github.com/a-h/templ`
  - `go install github.com/a-h/templ/cmd/templ@latest`
  - `go.mod`と`go.sum`の更新
  - **想定ファイル数**: 約2ファイル（実装2）
  - **想定行数**: 約10行（実装10行）

- [x] **1-2**: Makefileへのtempl generateタスク追加

  - `make templ-generate`タスクを追加
  - `make build`に`templ generate`を統合
  - `make clean`に生成ファイルのクリーンアップを追加
  - **想定ファイル数**: 約1ファイル（実装1）
  - **想定行数**: 約15行（実装15行）

- [x] **1-3**: airの設定更新（ホットリロード）

  - `.air.toml`に`.templ`ファイルの監視を追加
  - `templ generate`を自動実行するように設定
  - **想定ファイル数**: 約1ファイル（実装1）
  - **想定行数**: 約10行（実装10行）

- [x] **1-4**: プロトタイプの作成（1ページ分）

  - シンプルなページ（例: 502エラーページ）をtemplで実装
  - レンダリングが正常に動作することを確認
  - ハンドラーからの呼び出しを確認
  - **想定ファイル数**: 約3ファイル（実装2 + テスト1）
  - **想定行数**: 約100行（実装60行 + テスト40行）

### フェーズ 2: コアコンポーネントの移行

- [x] **2-1**: テンプレートヘルパーの移行

  - `internal/templates/helper.go`を更新
  - `T()`, `Locale()`, `Deref()`関数をtempl対応に変更
  - `Icon()`関数を`templ.Component`として実装
  - ヘルパー関数のテストを追加
  - **想定ファイル数**: 約2ファイル（実装1 + テスト1）
  - **想定行数**: 約150行（実装80行 + テスト70行）

- [x] **2-2**: パーシャルコンポーネントの移行

  - `components/head.templ`を作成
  - `components/flash.templ`を作成
  - `components/form_errors.templ`を作成
  - 各コンポーネントのテストを追加
  - **想定ファイル数**: 約6ファイル（実装3 + テスト3）
  - **想定行数**: 約200行（実装120行 + テスト80行）

- [x] **2-3**: レイアウトテンプレートの移行

  - `layouts/default.templ`を作成
  - `layouts/simple.templ`を作成
  - レイアウトのテストを追加
  - **想定ファイル数**: 約4ファイル（実装2 + テスト2）
  - **想定行数**: 約250行（実装150行 + テスト100行）

### フェーズ 3: ページテンプレートの移行

- [x] **3-1**: エラーページの移行

  - `pages/errors/502.templ`を作成
  - ハンドラーを更新
  - テストを追加
  - **想定ファイル数**: 約3ファイル（実装2 + テスト1）
  - **想定行数**: 約80行（実装50行 + テスト30行）

- [x] **3-2**: ログインページの移行

  - `pages/sign_in.templ`を作成
  - `SignInHandler`を更新
  - テストを追加
  - **想定ファイル数**: 約3ファイル（実装2 + テスト1）
  - **想定行数**: 約150行（実装90行 + テスト60行）

- [x] **3-3**: パスワードリセットページの移行

  - `pages/password/reset.templ`を作成
  - `pages/password/edit.templ`を作成
  - `pages/password/reset_sent.templ`を作成
  - ハンドラーを更新
  - テストを追加
  - **想定ファイル数**: 約6ファイル（実装4 + テスト2）
  - **想定行数**: 約300行（実装180行 + テスト120行）

- [x] **3-4**: 人気作品ページの移行

  - `pages/works/popular.templ`を作成
  - `PopularWorks`ハンドラーを更新
  - テストを追加
  - **想定ファイル数**: 約3ファイル（実装2 + テスト1）
  - **想定行数**: 約250行（実装150行 + テスト100行）

### フェーズ 4: メールテンプレートの移行

- [x] **4-1**: パスワードリセットメールの移行

  - `emails/password_reset/ja_html.templ`を作成
  - `emails/password_reset/en_html.templ`を作成
  - `emails/password_reset/ja_text.templ`を作成
  - `emails/password_reset/en_text.templ`を作成
  - メール送信処理を更新
  - テストを追加
  - **想定ファイル数**: 約5ファイル（実装4 + テスト1）
  - **想定行数**: 約200行（実装140行 + テスト60行）

### フェーズ 5: クリーンアップとドキュメント更新

- [x] **5-1**: 旧テンプレートファイルの削除

  - `internal/templates/*.html`を削除
  - `internal/handler/template_loader.go`を削除
  - 旧テンプレート関連のコードをクリーンアップ
  - **想定ファイル数**: 約17ファイル削除（実装17）
  - **想定行数**: 約-1000行（実装-1000行）

- [x] **5-2**: CLAUDE.mdの更新

  - テンプレートセクションをtempl用に更新
  - コーディング規約を更新
  - テストガイドを更新
  - **想定ファイル数**: 約1ファイル（実装1）
  - **想定行数**: 約100行（実装100行）

- [x] **5-3**: CIの更新

  - `.github/workflows/go-ci.yml`に`templ generate`を追加
  - ビルドステップを更新
  - **想定ファイル数**: 約1ファイル（実装1）
  - **想定行数**: 約10行（実装10行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **templのLSP（Language Server Protocol）のセットアップ**: 開発環境によって異なるため、各開発者が個別に設定
- **既存のRails版テンプレートの移行**: Go版のテンプレートのみを対象とする
- **Datastarとの統合**: 別タスクとして将来的に検討

## 参考資料

- [templ公式サイト](https://templ.guide/)
- [templ GitHubリポジトリ](https://github.com/a-h/templ)
- [templ Examples](https://github.com/a-h/templ/tree/main/examples)
- [Go html/templateドキュメント](https://pkg.go.dev/html/template)（移行前の参考）

---

## 実装の注意点

### templの基本構文

```templ
// パッケージ宣言（必須）
package pages

// インポート（必要に応じて）
import "github.com/annict/annict/internal/viewmodel"

// テンプレートコンポーネントの定義
templ Popular(ctx context.Context, works []viewmodel.Work) {
  <div>
    // 変数の出力
    <h2>{ templates.T(ctx, "popular_anime") }</h2>

    // 繰り返し
    for _, work := range works {
      <div>
        <h3>{ work.Title }</h3>

        // 条件分岐
        if work.ImageURL != "" {
          <img src={ work.ImageURL } alt={ work.Title } />
        }
      </div>
    }
  </div>
}
```

### 型安全性の活用

```templ
// 引数の型を明示的に指定
templ WorkCard(work viewmodel.Work, showCast bool) {
  <div class="card">
    <h3>{ work.Title }</h3>

    if showCast && len(work.Casts) > 0 {
      <div class="casts">
        for _, cast := range work.Casts {
          <span>{ cast.Name }</span>
        }
      </div>
    }
  </div>
}
```

### コンポーネントの再利用

```templ
// 他のコンポーネントを呼び出す
@layouts.Default(ctx, meta, user, Popular(ctx, works))
```

### 国際化対応のパターン

```templ
// テンプレート内で翻訳関数を呼び出す
<h2>{ templates.T(ctx, "popular_anime") }</h2>

// プレースホルダー付き翻訳
<p>{ templates.T(ctx, "watchers_count", map[string]any{"Count": work.WatchersCount}) }</p>

// 条件に応じて翻訳キーを変更
if work.SeasonNumber != nil {
  { templates.T(ctx, "season_spring") }
}
```
