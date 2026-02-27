# コードレビュー: archive-fix

## レビュー情報

| 項目                       | 内容                                            |
| -------------------------- | ----------------------------------------------- |
| レビュー日                 | 2026-02-26                                      |
| 対象ブランチ               | archive-fix                                     |
| ベースブランチ             | develop                                         |
| 作業計画書（指定があれば） | docs/plans/1_doing/work-episode-archive.md      |
| 変更ファイル数             | 46 ファイル（機能関連のみ。インフラ変更を除く） |
| 変更行数（実装）           | +1199 / -26 行                                  |
| 変更行数（テスト）         | +1149 / -35 行                                  |

## 参照するガイドライン

- [@CLAUDE.md#レビュー時に参照するガイドライン](/workspace/CLAUDE.md) - ガイドライン一覧

## 変更ファイル一覧

### Go版: マイグレーション

- [x] `go/db/migrations/20260210055715_add_status_to_works.sql`
- [x] `go/db/migrations/20260210081156_add_status_to_episodes.sql`
- [x] `go/db/schema.sql`

### Go版: 実装ファイル

- [x] `go/cmd/server/main.go`
- [x] `go/internal/handler/db_work/handler.go`
- [x] `go/internal/handler/db_work/index.go`
- [x] `go/internal/middleware/authorization.go`
- [x] `go/internal/middleware/auth.go`
- [x] `go/internal/middleware/reverse_proxy.go`
- [x] `go/internal/model/user.go`
- [x] `go/internal/model/work.go`
- [x] `go/internal/repository/work.go`
- [x] `go/internal/viewmodel/db_work.go`
- [x] `go/internal/viewmodel/pagination.go`
- [x] `go/internal/viewmodel/user.go`
- [x] `go/internal/query/queries/works.sql`
- [x] `go/internal/query/works.sql.go`
- [x] `go/internal/query/models.go`
- [x] `go/internal/query/querier.go`
- [x] `go/sqlc.yaml`

### Go版: テンプレート

- [x] `go/internal/templates/layouts/db.templ`
- [x] `go/internal/templates/components/db_sidebar.templ`
- [x] `go/internal/templates/components/pagination.templ`
- [x] `go/internal/templates/components/status_label.templ`
- [x] `go/internal/templates/pages/db_works/index.templ`

### Go版: テストファイル

- [x] `go/internal/handler/db_work/handler_test.go`
- [x] `go/internal/middleware/authorization_test.go`
- [x] `go/internal/middleware/auth_test.go`
- [x] `go/internal/viewmodel/pagination_test.go`
- [x] `go/internal/repository/work_test.go`
- [x] `go/internal/templates/components/status_label_test.go`

### Go版: 国際化

- [x] `go/internal/i18n/locales/ja.toml`
- [x] `go/internal/i18n/locales/en.toml`

### Go版: 自動生成ファイル（レビュー対象外）

- [x] `go/internal/templates/layouts/db_templ.go`
- [x] `go/internal/templates/components/db_sidebar_templ.go`
- [x] `go/internal/templates/components/pagination_templ.go`
- [x] `go/internal/templates/components/status_label_templ.go`
- [x] `go/internal/templates/pages/db_works/index_templ.go`

### Rails版: 実装ファイル

- [x] `rails/app/models/work.rb`
- [x] `rails/app/models/episode.rb`

### Rails版: テストファイル

- [x] `rails/spec/models/work_spec.rb`
- [x] `rails/spec/models/episode_spec.rb`
- [x] `rails/spec/requests/db/work_publishings/create_spec.rb`
- [x] `rails/spec/requests/db/work_publishings/destroy_spec.rb`
- [x] `rails/spec/requests/db/episode_publishings/create_spec.rb`
- [x] `rails/spec/requests/db/episode_publishings/destroy_spec.rb`

## ファイルごとのレビュー結果

### `go/internal/i18n/locales/ja.toml`

**ステータス**: 要修正

**チェックしたガイドライン**:

- [@go/docs/i18n-guide.md](/workspace/go/docs/i18n-guide.md) - 国際化ガイド

**問題点・改善提案**:

- **[@go/docs/i18n-guide.md#翻訳の追加手順]**: `watchers_count` キーの日本語翻訳が英語テキストになっている（32-35行目）

  ```toml
  # 現在のコード（ja.toml）
  [watchers_count]
  description = "アニメの視聴者数"
  one = "{{.Count}} watcher"
  other = "{{.Count}} watchers"
  ```

  **修正案**:

  ```toml
  [watchers_count]
  description = "アニメの視聴者数"
  other = "{{.Count}}人"
  ```

  日本語には単数・複数の区別がないため `one` は不要で `other` のみで十分です。

  **対応方針**:
  - [x] 修正案の通り日本語テキストに修正する
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

## 設計改善の提案

設計改善の提案はありません。

## 総合評価

**評価**: Comment

**総評**:

作品・エピソードアーカイブ機能のフェーズ1〜3（DB変更、共通基盤、作品一覧ページ）の実装全体を通してレビューしました。

**良かった点**:

- **設計書との整合性が高い**: マイグレーション、enum定義、カラム追加、インデックスなどすべて設計書通りに実装されている
- **アーキテクチャガイドラインへの準拠**: Handler/Repository/ViewModel/Templateの各レイヤーの責務分離が正しく行われている
- **テストカバレッジが充実**: 認可ミドルウェア、リポジトリ、ページネーション、ステータスラベルなど幅広くテストが書かれている
- **既存コードとの一貫性**: Repository構造体のパターン（`NewXxxRepository(queries *query.Queries)`）やテストヘルパー（`SetupTestDB(t)`）など、既存コードのパターンに従っている
- **セキュリティ**: SQLクエリはsqlcのプリペアドステートメント使用、エラーメッセージは適切（詳細を漏らさない）
- **国際化**: ユーザー向けメッセージはすべて `templates.T(ctx, ...)` で国際化対応済み
- **設計仕様に準拠した認証不要のDB一覧ページ**: 設計書の「未ログインでも閲覧可能」要件を正しく実装

**指摘事項**:

- **1件の軽微な問題**: `ja.toml` の `watchers_count` 翻訳が英語テキストになっている（修正は任意）

**補足**: この差分にはインフラ変更（Oxfmt移行、depguard修正、ドキュメント再構成、CI設定など）も含まれていますが、本レビューでは作業計画書に関連する機能実装ファイル（46ファイル）のみを対象としました。
