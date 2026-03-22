# コードレビュー: feature-flag-1-1

## レビュー情報

| 項目                       | 内容                               |
| -------------------------- | ---------------------------------- |
| レビュー日                 | 2026-03-22                         |
| 対象ブランチ               | feature-flag-1-1                   |
| ベースブランチ             | archive                            |
| 作業計画書（指定があれば） | docs/plans/1_doing/feature-flag.md |
| 変更ファイル数             | 13 ファイル（ドキュメント含む）    |
| 変更行数（実装）           | +171 / -0 行                       |
| 変更行数（テスト）         | +220 / -0 行                       |

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

- [x] `go/internal/repository/feature_flag_test.go`
- [x] `go/internal/testutil/feature_flag_builder.go`

### 自動生成ファイル（レビュー対象外）

- [x] `go/internal/query/feature_flags.sql.go`
- [x] `go/internal/query/models.go`
- [x] `go/internal/query/querier.go`
- [x] `go/db/schema.sql`

### ドキュメント

- [x] `docs/plans/1_doing/feature-flag.md`
- [x] `docs/reviews/done/202603/feature-flag-1-1-001.md`

## ファイルごとのレビュー結果

問題のあるファイルはありません。すべてのファイルがガイドラインに準拠しています。

## 設計改善の提案

### `go/db/migrations/20260322083140_create_feature_flags.sql`: インデックスの冗長性

**ステータス**: 要確認

**現状**:

マイグレーションでは UNIQUE 制約とは別に単一カラムのインデックスを作成している:

```sql
UNIQUE(device_token, name),
UNIQUE(user_id, name)
-- 上記に加えて:
CREATE INDEX idx_feature_flags_device_token ON feature_flags(device_token);
CREATE INDEX idx_feature_flags_user_id ON feature_flags(user_id);
CREATE INDEX idx_feature_flags_name ON feature_flags(name);
```

**提案**:

`idx_feature_flags_device_token` と `idx_feature_flags_user_id` は、それぞれ `UNIQUE(device_token, name)` と `UNIQUE(user_id, name)` の UNIQUE 制約が作成するインデックスの先頭カラムと一致する。PostgreSQL は複合インデックスの先頭カラムによる検索にもそのインデックスを使用できるため、単一カラムのインデックスは冗長である。

`idx_feature_flags_name` は `name` 単独での検索が必要な場合に有用だが、現在のクエリ（`IsFeatureFlagEnabled`）は常に `name` と `device_token`/`user_id` の組み合わせで検索するため、UNIQUE 制約のインデックスで十分カバーされる。

```sql
-- 削除候補
-- CREATE INDEX idx_feature_flags_device_token ON feature_flags(device_token);
-- CREATE INDEX idx_feature_flags_user_id ON feature_flags(user_id);
-- CREATE INDEX idx_feature_flags_name ON feature_flags(name);
```

**メリット**:

- ストレージ使用量の削減
- INSERT/UPDATE 時のインデックス更新コストの削減
- スキーマのシンプル化

**トレードオフ**:

- 将来 `device_token` のみや `name` のみでの検索が必要になった場合、インデックスの追加が必要になる
- 現時点ではレコード数が少なく、パフォーマンスへの影響は軽微

**対応方針**:

<!-- 開発者が回答を記入してください -->

- [x] 冗長なインデックス（3つとも）を削除する
- [ ] `idx_feature_flags_name` のみ残し、`device_token` と `user_id` のインデックスを削除する
- [ ] 現状のまま（将来の拡張性を考慮して残す）
- [ ] その他（下の回答欄に記入）

**回答**:

```
（ここに回答を記入）
```

## 設計との整合性チェック

作業計画書（タスク 1-1）に記載された以下の要件がすべて実装されていることを確認:

| 要件                                                      | 状態 |
| --------------------------------------------------------- | ---- |
| `db/migrations/` にマイグレーションを作成                 | ✅   |
| `internal/model/feature_flag.go` にモデルとフラグ名定数   | ✅   |
| `internal/model/id.go` にドメインID型                     | ✅   |
| `internal/query/queries/feature_flags.sql` に sqlc クエリ | ✅   |
| `internal/repository/feature_flag.go` にリポジトリ        | ✅   |
| `IsEnabled` メソッド                                      | ✅   |
| `IsEnabledByDeviceOrUser` メソッド                        | ✅   |
| `WithTx` メソッド                                         | ✅   |
| `internal/testutil/feature_flag_builder.go` にビルダー    | ✅   |
| `internal/repository/feature_flag_test.go` にテスト       | ✅   |
| DB カラム定義が `VARCHAR`（長さ指定なし）                 | ✅   |
| タイムスタンプが `TIMESTAMP WITH TIME ZONE`               | ✅   |
| 主キーが `BIGSERIAL`                                      | ✅   |
| CHECK 制約（`device_token` or `user_id` が NOT NULL）     | ✅   |
| UNIQUE 制約（`device_token, name` と `user_id, name`）    | ✅   |
| sqlc 生成コードが最新                                     | ✅   |

設計との乖離はありません。

## 総合評価

**評価**: Approve

**総評**:

作業計画書タスク 1-1 の要件がすべて正確に実装されている。コードは以下の点で優れている:

- **既存パターンとの一貫性**: Repository の構造（`queries` フィールド、`WithTx` メソッド）、テストの書き方（`SetupTestDB`、ビルダーパターン）がすべて既存コードベースのパターンに従っている
- **アーキテクチャ準拠**: 3層アーキテクチャの依存関係ルールに従い、Repository は Query のみに依存している
- **テストの充実**: 8 つのテストケースで正常系・異常系を網羅（デバイストークン単独、ユーザーID単独、両方マッチ、フラグなし、異なるフラグ名、空パラメータ、`IsEnabled` ラッパー、`WithTx` 動作）
- **SQL設計**: CHECK 制約と UNIQUE 制約で整合性を保証し、NULLable カラムの扱いも正しい
- **セキュリティ**: プリペアドステートメント（sqlc）使用、パラメータバインディングにより SQL インジェクション対策済み

設計改善の提案としてインデックスの冗長性について 1 件挙げたが、機能的な問題はなく、そのまま Approve とする。
