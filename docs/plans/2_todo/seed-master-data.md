# マスターデータシード機能 設計書

<!--
このテンプレートの使い方:
1. このファイルを `.claude/designs/2_todo/` ディレクトリにコピー
   例: cp .claude/designs/template.md .claude/designs/2_todo/new-feature.md
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

Go 版の `make seed` コマンドで、Rails 版の `db/seeds.rb` で生成していたマスターデータ（チャンネルグループ、チャンネル、エピソード番号フォーマット、都道府県）も生成できるようにする。

開発環境とテスト環境で異なるシードデータを生成できるようにする：
- **開発環境**: 既存の大量データ（ユーザー、作品、エピソード、視聴記録など）＋ マスターデータ
- **テスト環境**: マスターデータのみ

**目的**:

- Rails 版で生成していたマスターデータを Go 版でも生成可能にする
- 開発環境とテスト環境で必要なデータを柔軟に切り替えられるようにする
- Go 版への移行をスムーズに進める

**背景**:

- Rails 版の `db/seeds.rb` は 4 つのマスターデータテーブル（`channel_groups`, `channels`, `number_formats`, `prefectures`）に CSV ファイルからデータを投入している
- Go 版のシードコマンドは現在、大量のテストデータ（ユーザー、作品、視聴記録など）のみを生成しており、マスターデータは生成していない
- テスト環境では大量データは不要だが、マスターデータ（特にチャンネル情報）はテストに必要な場合がある

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

- システムは、Rails 版の CSV ファイル（`rails/db/data/csv/*.csv`）からマスターデータを読み込む
- システムは、以下の 4 つのテーブルにマスターデータを投入する：
  - `channel_groups`（チャンネルグループ）
  - `channels`（チャンネル）
  - `number_formats`（エピソード番号フォーマット）
  - `prefectures`（都道府県）
- ユーザーは `make seed` コマンドで開発環境にシードデータを生成できる（大量データ ＋ マスターデータ）
- ユーザーは `make seed-test` コマンドでテスト環境にマスターデータのみを生成できる
- システムは、既存のレコードが存在する場合は更新（upsert）を行う
- システムは、Rails 版と同様にシーケンス値を最後の ID ＋ 1 にリセットする

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

- **保守性**: マスターデータのシード処理は独立したパッケージ・関数として実装し、再利用可能にする
- **互換性**: Rails 版の CSV ファイルをそのまま使用し、データの一貫性を保つ
- **可用性**: CSV ファイルの読み込みエラーや DB 接続エラーを適切にハンドリングする

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

マスターデータのシード処理を新しいパッケージとして実装する：

```
go/internal/seed/
├── master/
│   ├── master.go           # エントリポイント（SeedMasterData関数）
│   ├── channel_group.go    # チャンネルグループのシード
│   ├── channel.go          # チャンネルのシード
│   ├── number_format.go    # エピソード番号フォーマットのシード
│   └── prefecture.go       # 都道府県のシード
```

#### 主要な構造体・関数

```go
// master.go
package master

// SeedMasterData はすべてのマスターデータをシードする
func SeedMasterData(ctx context.Context, db *sql.DB) error

// SeedChannelGroups はチャンネルグループをシードする
func SeedChannelGroups(ctx context.Context, db *sql.DB) error

// SeedChannels はチャンネルをシードする
func SeedChannels(ctx context.Context, db *sql.DB) error

// SeedNumberFormats はエピソード番号フォーマットをシードする
func SeedNumberFormats(ctx context.Context, db *sql.DB) error

// SeedPrefectures は都道府県をシードする
func SeedPrefectures(ctx context.Context, db *sql.DB) error
```

### CSVファイルの配置

Rails 版の CSV ファイル（`/workspace/rails/db/data/csv/*.csv`）を Go 版からも参照する。
Go コードからは相対パス `../rails/db/data/csv/` でアクセスする。

### Makefile の変更

```makefile
.PHONY: seed
seed: ## 開発環境用のシードデータを生成（大量データ + マスターデータ）
	@echo "シードデータを生成しています..."
	APP_ENV=dev op run --env-file=".env" -- go run cmd/seed/main.go

.PHONY: seed-test
seed-test: db-setup-test ## テスト環境用のマスターデータを生成
	@echo "テスト環境にマスターデータを生成しています..."
	APP_ENV=test op run --env-file=".env" -- go run cmd/seed/main.go --master-only
```

### cmd/seed/main.go の変更

コマンドライン引数 `--master-only` を追加し、マスターデータのみを生成するモードを実装する：

```go
var masterOnly = flag.Bool("master-only", false, "マスターデータのみを生成する")

func main() {
    flag.Parse()

    if *masterOnly {
        // マスターデータのみ生成
        master.SeedMasterData(ctx, db)
    } else {
        // 既存の大量データ生成 + マスターデータ生成
        // ... 既存のコード ...
        master.SeedMasterData(ctx, db)
    }
}
```

### データベース設計（参考：既存テーブル）

#### channel_groups テーブル

| カラム | 型 | 説明 |
|--------|-----|------|
| id | bigint | 主キー |
| sc_chgid | varchar(510) | しょぼいカレンダーのチャンネルグループID |
| name | varchar(510) | 名前 |
| sort_number | integer | ソート順 |
| created_at | timestamp with time zone | 作成日時 |
| updated_at | timestamp with time zone | 更新日時 |
| deleted_at | timestamp without time zone | 削除日時 |
| unpublished_at | timestamp without time zone | 非公開日時 |

#### channels テーブル

| カラム | 型 | 説明 |
|--------|-----|------|
| id | bigint | 主キー |
| channel_group_id | bigint | チャンネルグループID（外部キー） |
| sc_chid | integer | しょぼいカレンダーのチャンネルID |
| name | varchar | 名前 |
| created_at | timestamp with time zone | 作成日時 |
| updated_at | timestamp with time zone | 更新日時 |
| vod | boolean | VOD（ビデオオンデマンド）かどうか |
| aasm_state | varchar | 状態（published など） |
| sort_number | integer | ソート順 |
| deleted_at | timestamp without time zone | 削除日時 |
| name_alter | varchar | 別名 |
| unpublished_at | timestamp without time zone | 非公開日時 |

#### number_formats テーブル

| カラム | 型 | 説明 |
|--------|-----|------|
| id | bigint | 主キー |
| name | varchar | 名前（例: "第1話"） |
| data | varchar[] | エピソード番号のリスト（配列） |
| sort_number | integer | ソート順 |
| created_at | timestamp without time zone | 作成日時 |
| updated_at | timestamp without time zone | 更新日時 |
| format | varchar | フォーマット文字列（例: "第%d話"） |

#### prefectures テーブル

| カラム | 型 | 説明 |
|--------|-----|------|
| id | bigint | 主キー |
| name | varchar | 都道府県名 |
| created_at | timestamp without time zone | 作成日時 |
| updated_at | timestamp without time zone | 更新日時 |

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

### フェーズ 1: マスターデータシード機能の実装

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [ ] **1-1**: 都道府県（prefectures）のシード機能を実装

  - `internal/seed/master/prefecture.go` の作成
  - CSV ファイルの読み込み処理
  - UPSERT による都道府県データの投入
  - シーケンスリセット処理
  - 単体テストの作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [ ] **1-2**: エピソード番号フォーマット（number_formats）のシード機能を実装

  - `internal/seed/master/number_format.go` の作成
  - CSV ファイルの読み込み処理（JSON 配列のパース含む）
  - UPSERT によるデータ投入
  - シーケンスリセット処理
  - 単体テストの作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 250 行（実装 130 行 + テスト 120 行）

- [ ] **1-3**: チャンネルグループ（channel_groups）のシード機能を実装

  - `internal/seed/master/channel_group.go` の作成
  - CSV ファイルの読み込み処理
  - UPSERT によるデータ投入
  - シーケンスリセット処理
  - 単体テストの作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 220 行（実装 110 行 + テスト 110 行）

- [ ] **1-4**: チャンネル（channels）のシード機能を実装

  - `internal/seed/master/channel.go` の作成
  - CSV ファイルの読み込み処理
  - UPSERT によるデータ投入
  - シーケンスリセット処理
  - 単体テストの作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 280 行（実装 150 行 + テスト 130 行）

### フェーズ 2: エントリポイントの統合と Makefile の更新

- [ ] **2-1**: マスターデータシードのエントリポイントを実装

  - `internal/seed/master/master.go` の作成（SeedMasterData 関数）
  - 各シード関数の呼び出し順序の制御（外部キー制約を考慮）
  - 進捗表示の追加
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 80 行（実装 80 行）

- [ ] **2-2**: cmd/seed/main.go の更新と Makefile の変更

  - `--master-only` フラグの追加
  - 既存シード処理後にマスターデータシードを呼び出す
  - Makefile に `seed-test` ターゲットを追加
  - 統合テストの作成
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **CSV ファイルの Go 版へのコピー**: Rails 版の CSV ファイルをそのまま参照し、データの一貫性を保つ
- **マスターデータの削除機能**: 既存データのクリーンアップは `cleanupExistingData` 関数の責務外とする
- **マスターデータのバリデーション**: CSV ファイルのデータは Rails 版で検証済みのため、追加のバリデーションは行わない

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Rails 版 seeds.rb](/workspace/rails/db/seeds.rb)
- [Go 版 seed/main.go](/workspace/go/cmd/seed/main.go)
- [encoding/csv パッケージ](https://pkg.go.dev/encoding/csv)
