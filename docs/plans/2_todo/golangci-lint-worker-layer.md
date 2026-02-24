# golangci-lint Worker 層ルール追加・glob パターン修正 設計書

## 実装ガイドラインの参照

### Go版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/annict/go/CLAUDE.md) - 全体的なコーディング規約
- [@go/docs/architecture-guide.md](/annict/go/docs/architecture-guide.md) - アーキテクチャガイド

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

golangci-lint の depguard 設定を改善する：

1. Worker 層のアーキテクチャルールを追加
2. 各層の glob パターンを適切に修正

**目的**:

- Worker 層が不適切な層に依存することを防ぐ
- glob パターンを実際のディレクトリ構造に合わせて最適化する
- アーキテクチャの一貫性を静的解析で担保する

**背景**:

- 現在 `internal/worker` パッケージが存在するが、depguard にルールが定義されていない
- 一部の層で `**/*.go` パターンが使用されているが、サブディレクトリを持たない層では `*.go` で十分
- Wikino で同様の修正を行った

## 要件

### 機能要件

<!--
「ユーザーは〇〇できる」「システムは〇〇する」という形式で記述
箇条書きで簡潔に
-->

- Worker 層は Query に直接依存できない（Repository を経由する）
- Worker 層は Presentation 層（Handler, Middleware, ViewModel）に依存できない
- Worker 層は templates への依存を許可する（メールレンダリングのため）
- usecase 層のみサブディレクトリ（`seed/`）を持つため `**/*.go` パターンを維持
- その他の層はフラット構造のため `*.go` パターンに変更

## 設計

### アーキテクチャ

Worker は Application 層として位置づける。

```
┌─────────────────────────────────────────────────────────┐
│ Presentation層                                          │
│ - Handler, ViewModel, Template, Middleware            │
└─────────────────────────────────────────────────────────┘
         ↓ 依存
┌─────────────────────────────────────────────────────────┐
│ Application層                                           │
│ - UseCase（ビジネスフロー、トランザクション管理）           │
│ - Worker（バックグラウンドジョブ、非同期処理）             │
└─────────────────────────────────────────────────────────┘
         ↓ 依存
┌─────────────────────────────────────────────────────────┐
│ Domain/Infrastructure層（統合）                          │
│ - Query (sqlc), Repository, Model                     │
└─────────────────────────────────────────────────────────┘
```

**Worker の依存関係**:

- Worker は **UseCase を呼び出し可能**（定期実行処理などで永続化が必要な場合）
- Worker は **templates に依存可能**（メールレンダリングのため、例外として許可）
- Worker は **Presentation 層（Handler, Middleware, ViewModel）に依存不可**

### コード設計

#### 1. Worker 層ルールの追加

`.golangci.yml` に以下のルールを追加する：

```yaml
# Worker層のルール: バックグラウンドジョブ処理（Application層）
# Presentation層（Handler等）に依存しない
# 例外: templatesへの依存はメールレンダリングのため許可
worker-layer:
  files:
    - "**/internal/worker/*.go"
  deny:
    - pkg: github.com/annict/annict/go/internal/query
      desc: "WorkerはQueryに直接依存できません。Repositoryを経由してください。"
    - pkg: github.com/annict/annict/go/internal/handler
      desc: "WorkerはPresentation層に依存できません。"
    - pkg: github.com/annict/annict/go/internal/middleware
      desc: "WorkerはPresentation層に依存できません。"
    - pkg: github.com/annict/annict/go/internal/viewmodel
      desc: "WorkerはPresentation層に依存できません。"
```

#### 2. glob パターンの修正

以下の層の glob パターンを修正する：

| 層         | 変更前                           | 変更後                        | 理由                                     |
| ---------- | -------------------------------- | ----------------------------- | ---------------------------------------- |
| viewmodel  | `**/internal/viewmodel/**/*.go`  | `**/internal/viewmodel/*.go`  | サブディレクトリなし                     |
| middleware | `**/internal/middleware/**/*.go` | `**/internal/middleware/*.go` | サブディレクトリなし                     |
| model      | `**/internal/model/**/*.go`      | `**/internal/model/*.go`      | サブディレクトリなし                     |
| repository | `**/internal/repository/**/*.go` | `**/internal/repository/*.go` | サブディレクトリなし                     |
| query      | `**/internal/query/**/*.go`      | `**/internal/query/*.go`      | サブディレクトリなし                     |
| usecase    | `**/internal/usecase/**/*.go`    | **変更なし**                  | `seed/` サブディレクトリが存在するため   |
| handler    | `**/internal/handler/**/*.go`    | **変更なし**                  | リソースごとのサブディレクトリがあるため |
| templates  | `**/internal/templates/**/*.go`  | **変更なし**                  | サブディレクトリ構造を持つため           |
| worker     | -                                | `**/internal/worker/*.go`     | 新規追加                                 |

## タスクリスト

### フェーズ 1: depguard 設定の改善

- [ ] **1-1**: [Go] Worker 層の depguard ルールを追加し、glob パターンを修正

  - `.golangci.yml` に worker-layer ルールを追加
  - viewmodel, middleware, model, repository, query の glob パターンを `*.go` に変更
  - `make lint` を実行して既存コードが違反していないことを確認
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 25 行（実装 25 行 + テスト 0 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **CLAUDE.md の更新**: Worker 層に関するドキュメントは別途対応

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Wikino の .golangci.yml](/workspace/go/.golangci.yml) - 参考実装
