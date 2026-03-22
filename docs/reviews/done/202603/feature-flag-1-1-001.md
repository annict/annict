# コードレビュー: feature-flag-1-1

## レビュー情報

| 項目                       | 内容                                          |
| -------------------------- | --------------------------------------------- |
| レビュー日                 | 2026-03-22                                    |
| 対象ブランチ               | feature-flag-1-1                              |
| ベースブランチ             | archive                                       |
| 作業計画書（指定があれば） | docs/plans/1_doing/feature-flag.md            |
| 変更ファイル数             | 12 ファイル                                   |
| 変更行数（実装）           | +111 行（手書きコード） / +140 行（自動生成） |
| 変更行数（テスト）         | +264 行                                       |

## 参照するガイドライン

- [@CLAUDE.md#レビュー時に参照するガイドライン](/workspace/CLAUDE.md) - ガイドライン一覧
- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - Go版の開発ガイド
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド

## 変更ファイル一覧

### 実装ファイル

- [x] `go/db/migrations/20260322083140_create_feature_flags.sql`
- [x] `go/internal/model/feature_flag.go`
- [x] `go/internal/model/id.go`
- [x] `go/internal/query/queries/feature_flags.sql`
- [x] `go/internal/repository/feature_flag.go`

### テストファイル

- [ ] `go/internal/repository/feature_flag_test.go`
- [x] `go/internal/testutil/feature_flag_builder.go`

### 設定・その他

- [x] `go/db/schema.sql`（自動生成）
- [x] `go/internal/query/feature_flags.sql.go`（sqlc自動生成）
- [x] `go/internal/query/models.go`（sqlc自動生成）
- [x] `go/internal/query/querier.go`（sqlc自動生成）
- [x] `docs/plans/1_doing/feature-flag.md`

## ファイルごとのレビュー結果

### `go/internal/repository/feature_flag_test.go`

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/CLAUDE.md#テスト戦略](/workspace/go/CLAUDE.md) - テストのベストプラクティス

**問題点・改善提案**:

- **[@go/CLAUDE.md#テストのベストプラクティス]**: `t.Parallel()` が使用されていない

  テスト関数に `t.Parallel()` が追加されていません。CLAUDE.md では「`t.Parallel()` で並行実行可能なテストを高速化（トランザクション分離により安全）」と記載されています。

  ただし、同じ `repository` パッケージの既存テスト（`session_test.go`, `user_test.go` 等）でも `t.Parallel()` を使用していないものが多く、`work_test.go` のサブテストでのみ使用されています。既存パターンとの一貫性を考慮すると、今回追加しなくても大きな問題ではありません。

  **修正案**:

  各テスト関数の先頭に `t.Parallel()` を追加する。

  ```go
  func TestFeatureFlagRepository_IsEnabledByDeviceOrUser_DeviceToken(t *testing.T) {
      t.Parallel()
      // ...
  }
  ```

  **対応方針**:
  - [x] 全テスト関数に `t.Parallel()` を追加する
  - [ ] 既存の repository テストに合わせて今回は追加しない
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

## 設計改善の提案

設計改善の提案はありません。

## 設計との整合性チェック

作業計画書（タスク 1-1）に記載されたすべての要件が実装されていることを確認しました：

| 計画書の要件                                                                | 実装状況 |
| --------------------------------------------------------------------------- | -------- |
| `db/migrations/` にマイグレーションを作成                                   | ✅       |
| `internal/model/feature_flag.go` に FeatureFlag モデルとフラグ名定数        | ✅       |
| `internal/model/id.go` に FeatureFlagID, FeatureFlagName 型                 | ✅       |
| `internal/query/queries/feature_flags.sql` に sqlc クエリ                   | ✅       |
| `internal/repository/feature_flag.go` に IsEnabled, IsEnabledByDeviceOrUser | ✅       |
| `internal/testutil/feature_flag_builder.go` にテスト用ビルダー              | ✅       |
| `internal/repository/feature_flag_test.go` にリポジトリのテスト             | ✅       |

設計との乖離はありません。

## 総合評価

**評価**: Approve

**総評**:

フィーチャーフラグ機能のデータベース・モデル・リポジトリ層の実装として、作業計画書の要件を正確に満たしています。

**良かった点**:

- マイグレーションのテーブル定義がカラム定義ガイドライン（`VARCHAR` 長さ指定なし、`TIMESTAMP WITH TIME ZONE`）に準拠している
- CHECK制約とUNIQUE制約で適切なデータ整合性を確保している
- `IsEnabledByDeviceOrUser` の `sql.NullString` / `sql.NullInt64` による NULL ハンドリングが正確
- テストケースが正常系・異常系・境界値を網羅している（8テスト関数）
- 3層アーキテクチャの依存関係ルール（Repository → Query）に従っている
- 既存の repository テストパターン（`SetupTestDB`, `query.New(db).WithTx(tx)`）と一貫している
- ドメインID型（`FeatureFlagID`, `FeatureFlagName`）の導入で `id.go` ファイルが整備された

**軽微な指摘**:

- テストに `t.Parallel()` が未使用（既存パターンとの一貫性があるため任意）
