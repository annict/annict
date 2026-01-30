# viewmodel 画像URL事前計算 設計書

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

viewmodel において、画像URLを構造体作成時に事前計算する方針に統一する。
現在の `viewmodel.Work` は `ImageDataJSON` フィールドと `GetImageURL` メソッドで動的にURLを生成しているが、
これをテンプレートがシンプルになる「事前計算済みURL」方式に変更する。

また、新たに `viewmodel.User` を追加し、サイドバーなどで表示するアバター画像URLを事前計算して保持する。
アバター画像の表示ロジックは `components/avatar.templ` として切り出し、再利用可能にする。

**目的**:

- テンプレートの責務を「表示」に限定し、URL生成ロジックを持たせない
- viewmodel の設計を統一し、コードの一貫性を向上させる
- アバター画像コンポーネントを再利用可能にする

**背景**:

- 現在の `viewmodel.Work` は `ImageDataJSON` を持ち、テンプレート側で `GetImageURL` を呼び出してURLを生成している
- サイドバーにアバター画像を表示する機能を追加した際、同様の議論が発生
- テンプレートをシンプルに保つため、事前計算済みURL方式に統一することを決定

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

- `viewmodel.User` を新規作成し、ユーザー情報とアバター画像URLを保持する
- `components/avatar.templ` を作成し、アバター画像の表示ロジックを共通化する
- アバター画像が未設定の場合は、ユーザー名の頭文字をフォールバック表示する
- `viewmodel.Work` から `ImageDataJSON` フィールドと `GetImageURL`/`GetSrcSet` メソッドを削除し、事前計算済みURLのみにする

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

- **保守性**: テンプレートがシンプルになり、URL生成ロジックが viewmodel に集約される
- **一貫性**: viewmodel の設計方針が統一される

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

### コード設計

#### viewmodel.User

```go
// internal/viewmodel/user.go
package viewmodel

import (
    "github.com/annict/annict/go/internal/image"
    "github.com/annict/annict/go/internal/repository"
)

// User はテンプレート表示用のユーザーデータです
type User struct {
    ID        int64
    Username  string
    AvatarURL string // 事前計算済みアバター画像URL
}

// NewUserFromQuery は repository.GetUserByIDRow から viewmodel.User に変換します
func NewUserFromQuery(row *repository.GetUserByIDRow, helper *image.Helper, avatarWidth int) *User {
    if row == nil {
        return nil
    }

    var avatarURL string
    if row.ProfileImageData.Valid && helper != nil {
        avatarURL = helper.GetAvatarImageURL(row.ProfileImageData.String, avatarWidth, "webp")
    }

    return &User{
        ID:        row.ID,
        Username:  row.Username,
        AvatarURL: avatarURL,
    }
}
```

#### components/avatar.templ

```templ
// internal/templates/components/avatar.templ
package components

// AvatarSize はアバターのサイズバリエーション
type AvatarSize string

const (
    AvatarSizeSm AvatarSize = "sm"  // 32px (w-8 h-8)
    AvatarSizeMd AvatarSize = "md"  // 40px (w-10 h-10)
    AvatarSizeLg AvatarSize = "lg"  // 48px (w-12 h-12)
)

templ Avatar(avatarURL string, username string, size AvatarSize) {
    // サイズに応じたCSSクラスを適用
    // avatarURL が空の場合はユーザー名の頭文字を表示
}
```

#### viewmodel.Work の変更（将来）

```go
// 変更前
type Work struct {
    ImageURL      string
    ImageDataJSON string
    imageHelper   *image.Helper
}

func (w *Work) GetImageURL(width int, format string) string { ... }
func (w *Work) GetSrcSet(width int, format string) string { ... }

// 変更後
type Work struct {
    ImageURL    string // 事前計算済み（デフォルトサイズ）
    ImageURL2x  string // 事前計算済み（2x Retina用）
    // ImageDataJSON, imageHelper, GetImageURL, GetSrcSet は削除
}
```

### 実装方針

1. **フェーズ1**: `viewmodel.User` と `components/avatar.templ` を作成
   - 現在のサイドバー実装をリファクタリング
   - 既存のテストを更新

2. **フェーズ2**: `viewmodel.Work` を事前計算済みURL方式に変更
   - `ImageDataJSON`、`imageHelper`、`GetImageURL`、`GetSrcSet` を削除
   - 必要なサイズのURLを事前計算してフィールドに保持
   - テンプレートを更新

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

### フェーズ 1: viewmodel.User と Avatar コンポーネントの作成

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: viewmodel.User の作成と Avatar コンポーネントの実装

  - `internal/viewmodel/user.go` を作成
  - `internal/templates/components/avatar.templ` を作成
  - `sidebar.templ` を Avatar コンポーネントを使用するようにリファクタリング
  - ハンドラー（`popular_work/index.go`, `supporters/show.go` など）を更新
  - 関連するテストを更新
  - **想定ファイル数**: 約 12 ファイル（実装 6 + テスト 6）
  - **想定行数**: 約 250 行（実装 150 行 + テスト 100 行）

### フェーズ 2: viewmodel.Work のリファクタリング

- [ ] **2-1**: viewmodel.Work から動的URL生成を削除

  - `ImageDataJSON` フィールドを削除
  - `imageHelper` フィールドを削除
  - `GetImageURL` メソッドを削除
  - `GetSrcSet` メソッドを削除
  - 必要なサイズのURLを事前計算してフィールドに保持（`ImageURL`, `ImageURL2x` など）
  - テンプレートを更新（`GetImageURL` 呼び出しをフィールド参照に変更）
  - 関連するテストを更新
  - **想定ファイル数**: 約 8 ファイル（実装 4 + テスト 4）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **キャッシュ機構**: URL生成は軽量なため、メモ化やキャッシュは不要と判断
- **複数フォーマット対応**: webp 固定で実装。将来的にブラウザ対応が必要になった場合に検討

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [go/docs/architecture-guide.md](../../go/docs/architecture-guide.md) - アーキテクチャガイド
- [go/docs/templ-guide.md](../../go/docs/templ-guide.md) - templ テンプレートガイド
