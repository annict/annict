# Simpleレイアウトのリファクタリング 設計書

<!--
このテンプレートの使い方:
1. このファイルを `docs/designs/2_todo/` ディレクトリにコピー
   例: cp docs/designs/template.md docs/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨
-->

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

Go版Annictの`Simple`レイアウトテンプレートのシグネチャを簡素化し、templ のベストプラクティスに沿った設計に改善する。

現在の`Simple`レイアウトは5つのパラメータを受け取っており、呼び出し側の負担が大きく、関心の分離も不十分である。Mewstプロジェクトの設計を参考に、パラメータを整理し、コンテキストを活用した設計に変更する。

**目的**:

- レイアウトテンプレートの呼び出しを簡素化し、ハンドラーのコードを読みやすくする
- レイアウトがセッション層（`*session.Flash`）の詳細に直接依存しないようにする
- `assetVersion`を`PageMeta`に統合し、パラメータ数を削減する

**背景**:

- 現在の`Simple`レイアウトは`ctx`, `meta`, `flash`, `assetVersion`, `content`の5つのパラメータを受け取っている
- templでは`ctx`は暗黙的に利用可能であり、明示的に渡す必要がない
- Flashメッセージはコンテキストから取得する設計にすることで、レイアウトの責務を軽減できる
- `assetVersion`は`PageMeta`に含めることで、関連するメタ情報を一箇所で管理できる

## 要件

<!--
ガイドライン:
- 機能要件: 「何ができるべきか」を記述
- 非機能要件: 「どのように動くべきか」を必要に応じて記述
-->

### 機能要件

<!--
「ユーザーは〇〇できる」「システムは〇〇する」という形式で記述
箇条書きで簡潔に
-->

- `Simple`レイアウトは`meta`と`content`の2つのパラメータのみを受け取る
- `PageMeta`構造体に`AssetVersion`フィールドを追加する
- `Flash`コンポーネントはコンテキストからFlashデータを取得する
- `Default`レイアウトも同様にリファクタリングする
- 既存の動作（Flashメッセージ表示、ページメタ情報、アセットバージョニング）は維持する

### 非機能要件

<!--
必要に応じて以下のような項目を追加してください：
- セキュリティ（認証、認可、暗号化、監査ログなど）
- パフォーマンス（応答時間、スループット、リソース使用量など）
- ユーザビリティ（UX）（使いやすさ、わかりやすさ、アクセシビリティなど）
- 可用性・信頼性（稼働率、障害時の挙動、エラーハンドリングなど）
- 保守性（テストのしやすさ、コードの読みやすさ、ドキュメントなど）

不要な場合はこのセクション全体を削除してください。
-->

- **保守性**: ハンドラーでのレイアウト呼び出しがシンプルになり、コードの可読性が向上する
- **一貫性**: `Simple`と`Default`の両レイアウトで同じパターンを使用する

## 設計

<!--
ガイドライン:
- 技術的な実装の詳細を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - 技術スタック（使用するライブラリ、フレームワーク、ツールなど）
  - アーキテクチャ（システム全体の構成、コンポーネント間の関係など）
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - セキュリティ設計（認証・認可、トークン管理、Rate Limitingなど）
  - コード設計（パッケージ構成、主要な構造体、インターフェースなど）
  - テスト戦略（単体テスト、統合テスト、E2Eテストの方針）
  - マイグレーション管理（データベースマイグレーションの方針）
  - 実装方針（特記事項、既存システムとの関係、制約など）

不要な場合はこのセクション全体を削除してください。
-->

### 変更前後の比較

#### Simpleレイアウト

**変更前**:

```go
templ Simple(ctx context.Context, meta viewmodel.PageMeta, flash *session.Flash, assetVersion string, content templ.Component)
```

**変更後**:

```go
templ Simple(meta viewmodel.PageMeta, content templ.Component)
```

#### PageMeta構造体

**変更前**:

```go
type PageMeta struct {
    Title       string
    Description string
    OGType      string
    OGURL       string
    OGImage     string
}
```

**変更後**:

```go
type PageMeta struct {
    Title        string
    Description  string
    OGType       string
    OGURL        string
    OGImage      string
    AssetVersion string  // 追加
}
```

#### Flashコンポーネント

**変更前**:

```go
templ Flash(ctx context.Context, flash *session.Flash)
```

**変更後**:

```go
templ Flash()
// コンテキストからFlashを取得
```

### コンテキストへのFlash格納

ミドルウェアまたはハンドラーでFlashをコンテキストに格納する。

```go
// internal/session/context.go（新規作成）
type contextKey string

const flashKey contextKey = "flash"

func WithFlash(ctx context.Context, flash *Flash) context.Context {
    return context.WithValue(ctx, flashKey, flash)
}

func FlashFromContext(ctx context.Context) *Flash {
    flash, _ := ctx.Value(flashKey).(*Flash)
    return flash
}
```

### 影響を受けるファイル

#### レイアウトテンプレート

- `internal/templates/layouts/simple.templ`
- `internal/templates/layouts/default.templ`

#### コンポーネント

- `internal/templates/components/flash.templ`
- `internal/templates/components/head.templ`（assetVersionの取得方法変更）

#### ビューモデル

- `internal/viewmodel/page_meta.go`

#### セッション

- `internal/session/context.go`（新規作成）

#### ハンドラー（11箇所）

- `internal/handler/sign_in/new.go`
- `internal/handler/sign_in_code/show.go`
- `internal/handler/sign_in_password/new.go`
- `internal/handler/sign_up/new.go`
- `internal/handler/sign_up_code/new.go`
- `internal/handler/sign_up_username/new.go`
- `internal/handler/password/edit.go`
- `internal/handler/password_reset/new.go`
- `internal/handler/password_reset/create.go`

### ハンドラーの変更例

**変更前**:

```go
meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
meta.Title = i18n.T(ctx, "sign_in_title") + " | Annict"
// ...
component := layouts.Simple(ctx, meta, flash, h.cfg.GetAssetVersion(), sign_in.New(ctx, formErrors, csrfToken, h.cfg.TurnstileSiteKey, backURL))
```

**変更後**:

```go
meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
meta.Title = i18n.T(ctx, "sign_in_title") + " | Annict"
ctx = session.WithFlash(ctx, flash)
// ...
component := layouts.Simple(meta, sign_in.New(ctx, formErrors, csrfToken, h.cfg.TurnstileSiteKey, backURL))
```

## タスクリスト

<!--
ガイドライン:
- フェーズごとに段階的な実装計画を記述
- チェックボックスで進捗を管理
- **重要**: 1タスク = 1 Pull Request の粒度で作成してください
- **重要**: 各タスクには想定ファイル数と想定行数を明記してください（PRサイズの見積もりのため）
- 想定ファイル数は「実装」と「テスト」に分けて記載してください
- 想定行数も「実装」と「テスト」に分けて記載してください
- 依存関係を明確に
- Pull Requestのガイドラインは CLAUDE.md を参照（変更ファイル数20以下、変更行数300行以下）

タスク番号の付け方:
- 各タスクには階層的な番号を付与します（例: 1-1, 1-2, 2-1, 2-2）
- フォーマット: **フェーズ番号-タスク番号**: タスク名
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）
-->

### フェーズ 1: インフラストラクチャ準備

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [ ] **1-1**: PageMetaにAssetVersionフィールドを追加
  - `internal/viewmodel/page_meta.go`に`AssetVersion`フィールドを追加
  - `DefaultPageMeta`関数の戻り値に`AssetVersion`を含める
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 15 行 + テスト 15 行）

- [ ] **1-2**: セッションコンテキストヘルパーを作成
  - `internal/session/context.go`を新規作成
  - `WithFlash`と`FlashFromContext`関数を実装
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 25 行 + テスト 25 行）

### フェーズ 2: コンポーネント・レイアウトの更新

- [ ] **2-1**: Flashコンポーネントをリファクタリング
  - `internal/templates/components/flash.templ`を更新
  - コンテキストからFlashを取得するように変更
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 40 行（実装 20 行 + テスト 20 行）

- [ ] **2-2**: Headコンポーネントを更新
  - `internal/templates/components/head.templ`を更新
  - `assetVersion`を`PageMeta`から取得するように変更
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 20 行（実装 20 行）

- [ ] **2-3**: Simpleレイアウトをリファクタリング
  - `internal/templates/layouts/simple.templ`を更新
  - パラメータを`meta`と`content`の2つに削減
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 20 行（実装 20 行）

- [ ] **2-4**: Defaultレイアウトをリファクタリング
  - `internal/templates/layouts/default.templ`を更新
  - Simpleレイアウトと同様のパターンに変更
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 25 行（実装 25 行）

### フェーズ 3: ハンドラーの更新

- [ ] **3-1**: 認証関連ハンドラーの更新（Part 1）
  - `internal/handler/sign_in/new.go`
  - `internal/handler/sign_in_code/show.go`
  - `internal/handler/sign_in_password/new.go`
  - レイアウト呼び出しを新しいシグネチャに変更
  - コンテキストにFlashを格納
  - **想定ファイル数**: 約 3 ファイル（実装 3）
  - **想定行数**: 約 30 行（実装 30 行）

- [ ] **3-2**: 認証関連ハンドラーの更新（Part 2）
  - `internal/handler/sign_up/new.go`
  - `internal/handler/sign_up_code/new.go`
  - `internal/handler/sign_up_username/new.go`
  - レイアウト呼び出しを新しいシグネチャに変更
  - コンテキストにFlashを格納
  - **想定ファイル数**: 約 3 ファイル（実装 3）
  - **想定行数**: 約 30 行（実装 30 行）

- [ ] **3-3**: パスワード関連ハンドラーの更新
  - `internal/handler/password/edit.go`
  - `internal/handler/password_reset/new.go`
  - `internal/handler/password_reset/create.go`
  - レイアウト呼び出しを新しいシグネチャに変更
  - コンテキストにFlashを格納
  - **想定ファイル数**: 約 3 ファイル（実装 3）
  - **想定行数**: 約 40 行（実装 40 行）

### フェーズ 4: ドキュメント更新

- [ ] **4-1**: CLAUDE.mdにレイアウトテンプレートのガイドラインを追加
  - `go/CLAUDE.md`の「templ テンプレート」セクションにレイアウトの使い方を追加
  - レイアウトのシグネチャ、コンテキストへのFlash格納方法、使用例を記載
  - 新しいレイアウト設計のベストプラクティスを文書化
  - **想定ファイル数**: 約 1 ファイル（ドキュメント 1）
  - **想定行数**: 約 50 行（ドキュメント 50 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **ミドルウェアでのFlash自動格納**: 各ハンドラーで明示的にFlashをコンテキストに格納する。将来的にミドルウェアで自動化することは検討可能
- **JavaScriptベースのFlash表示**: Mewstのような完全にJavaScript制御のFlash表示への移行。現在のサーバーサイドレンダリング方式を維持する

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [templ公式ドキュメント](https://templ.guide/)
- Mewstプロジェクト `/mewst/go/internal/templates/layouts/simple.templ`

---

## テンプレート使用例

実際の使用例は以下を参照してください：

- [パスワードリセット機能](doing/password-reset.md)
