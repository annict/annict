# コードレビュー: feature-flag-1-1

## レビュー情報

| 項目                       | 内容                                        |
| -------------------------- | ------------------------------------------- |
| レビュー日                 | 2026-03-22                                  |
| 対象ブランチ               | feature-flag-1-1                            |
| ベースブランチ             | archive                                     |
| 作業計画書（指定があれば） | docs/plans/1_doing/feature-flag.md          |
| 変更ファイル数             | 11 ファイル（Go 実装関連、schema.sql 含む） |
| 変更行数（実装）           | +144 / -0 行                                |
| 変更行数（テスト）         | +220 / -0 行                                |

## 参照するガイドライン

- [@CLAUDE.md#レビュー時に参照するガイドライン](/workspace/CLAUDE.md) - ガイドライン一覧
- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - Go 版の開発ガイド
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド

## 変更ファイル一覧

### 実装ファイル

- [x] `go/db/migrations/20260322083140_create_feature_flags.sql`
- [x] `go/internal/model/feature_flag.go`
- [x] `go/internal/model/id.go`
- [x] `go/internal/query/queries/feature_flags.sql`
- [x] `go/internal/query/feature_flags.sql.go`（sqlc 自動生成）
- [x] `go/internal/query/models.go`（sqlc 自動生成）
- [x] `go/internal/query/querier.go`（sqlc 自動生成）
- [x] `go/internal/repository/feature_flag.go`
- [x] `go/internal/testutil/feature_flag_builder.go`

### テストファイル

- [x] `go/internal/repository/feature_flag_test.go`

### 設定・その他

- [x] `go/db/schema.sql`（dbmate 自動生成）

## ファイルごとのレビュー結果

問題のあるファイルはありません。すべてのファイルがガイドラインに準拠しています。

## 設計との整合性チェック

作業計画書（タスク 1-1）に記載された要件との整合性を確認しました。

### 確認結果

| 要件                                                          | 実装状況 |
| ------------------------------------------------------------- | -------- |
| `db/migrations/` にマイグレーションを作成                     | ✅       |
| `internal/model/feature_flag.go` に FeatureFlag モデルを定義  | ✅       |
| `internal/model/id.go` にドメイン ID 型を定義                 | ✅       |
| `internal/query/queries/feature_flags.sql` に sqlc クエリ定義 | ✅       |
| `sqlc generate` を実行                                        | ✅       |
| `internal/repository/feature_flag.go` に Repository を実装    | ✅       |
| `IsEnabled` メソッド                                          | ✅       |
| `IsEnabledByDeviceOrUser` メソッド                            | ✅       |
| `internal/testutil/feature_flag_builder.go` にビルダーを作成  | ✅       |
| `internal/repository/feature_flag_test.go` にテストを作成     | ✅       |

### 作業計画書のインデックスとの差異

作業計画書では以下の個別インデックスが定義されていますが、実装では省略されています：

```sql
CREATE INDEX idx_feature_flags_device_token ON feature_flags(device_token);
CREATE INDEX idx_feature_flags_user_id ON feature_flags(user_id);
CREATE INDEX idx_feature_flags_name ON feature_flags(name);
```

これは問題ありません。`UNIQUE(device_token, name)` と `UNIQUE(user_id, name)` が PostgreSQL で自動的にインデックスを生成するため、`device_token` と `user_id` のルックアップは UNIQUE インデックスの左端カラムで十分にカバーされます。`name` 単体のインデックスは `IsFeatureFlagEnabled` クエリが常に `name` を `device_token` または `user_id` と組み合わせて使用するため不要です。

## 設計改善の提案

設計改善の提案はありません。

## 総合評価

**評価**: Approve

**総評**:

作業計画書タスク 1-1 の要件をすべて満たした実装です。主な評価ポイント：

- **アーキテクチャ準拠**: 3 層アーキテクチャの依存関係ルール（Query → Repository → Model）に従っている
- **既存パターンとの一貫性**: `queries *query.Queries` フィールド名、`WithTx` パターン、`SetupTestDB` の使用、ビルダーの戻り値 `int64` など、既存コードベースのパターンと完全に一致
- **マイグレーション**: `VARCHAR`（長さ指定なし）、`TIMESTAMP WITH TIME ZONE`、`BIGSERIAL` 主キーのガイドラインに準拠
- **ドメイン ID 型**: `FeatureFlagID` と `FeatureFlagName` を適切に定義し、Model で使用
- **SQL インジェクション対策**: sqlc のプリペアドステートメントを使用
- **テスト網羅性**: デバイストークン/ユーザー ID/両方一致/不一致/異なるフラグ名/空パラメータ/`WithTx`/`IsEnabled` の 8 ケースを網羅
- **コメント**: 日本語で記述されており、意図を明確に説明している
