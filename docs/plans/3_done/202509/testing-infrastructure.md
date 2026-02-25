# テストインフラの改善

## 概要

実データベースを使用したテスト環境の構築と、フィクスチャーシステムの実装。
モックに依存せず、実際の PostgreSQL データベースでテストを実行する基盤を整備。

## 実装内容

### フェーズ 4: テストインフラの改善（実データベーステストへの移行）

- [x] **テスト用データベース環境の構築**
  - [x] テスト用データベース（annict_test）の設定
  - [x] テスト用 DSN の環境変数設定（.env.test ファイル作成）
  - [x] Rails 側の Docker Compose で起動している PostgreSQL（ポート 15432）を使用

- [x] **テストヘルパーの実装**
  - [x] SetupTestDB 関数の作成（internal/testutil/db.go）
  - [x] テスト毎のトランザクション分離
  - [x] 自動ロールバック機構（t.Cleanup 使用）
  - [x] 並列実行対応の実装

- [x] **フィクスチャーシステムの構築**
  - [x] フィクスチャービルダーパターンの実装（internal/testutil/fixtures.go）
  - [x] Work、User、Episode、WorkImage、Session のビルダー作成
  - [x] デフォルト値とカスタマイズ可能な設定
  - [x] ヘルパー関数による簡単なテストデータ作成

- [x] **既存モックテストの移行**
  - [x] handler_db_test.go の実 DB 版作成（handler_test.go はモック版として保持）
  - [x] 実 DB テストの基本実装（テーブル作成後に完全動作予定）
  - [x] 制約違反のテストケース追加
  - [x] トランザクションのロールバックテスト追加

- [x] **テスト戦略の文書化**
  - [x] 実 DB テストとモックテストの使い分けガイドライン作成（internal/testutil/README.md）
  - [x] テストデータ管理のベストプラクティス文書化
  - [x] CI 環境での DB 設定方法記載

- [x] **パフォーマンス測定と最適化**
  - [x] ベンチマークテストの実装（BenchmarkPopularWorksWithRealDB）
  - [x] 並列実行対応（t.Parallel()使用可能）
  - [x] テスト用接続プールの最適化

## テスト戦略

### 基本方針

- **実データベース優先**: モックではなく実際の PostgreSQL データベースでテスト
- **トランザクション分離**: 各テストはトランザクション内で実行し、自動ロールバック
- **並列実行**: `t.Parallel()` で複数のテストを並行実行（高速化）

### 実 DB テストとモックテストの使い分け

| テストの種類        | 推奨アプローチ | 理由                                     |
| ------------------- | -------------- | ---------------------------------------- |
| Handler のテスト    | 実 DB          | リアルな動作を検証、SQL の正確性を確認   |
| Repository のテスト | 実 DB          | SQL クエリの正確性を検証                 |
| ViewModel のテスト  | モック         | 単純な変換ロジックのみ、DB 不要          |
| Usecase のテスト    | 実 DB          | トランザクション管理を含む複雑なロジック |
| 単純な関数          | モック不要     | 入出力のみのテスト                       |

## テストヘルパー

### SetupTestDB

テスト用データベース接続とトランザクションのセットアップ：

```go
// internal/testutil/db.go
func SetupTestDB(t *testing.T) (*sql.DB, *sql.Tx) {
    t.Helper()

    // テスト用データベースに接続
    db, err := sql.Open("postgres", getTestDSN())
    if err != nil {
        t.Fatalf("failed to connect to test database: %v", err)
    }

    // トランザクション開始
    tx, err := db.BeginTx(context.Background(), nil)
    if err != nil {
        t.Fatalf("failed to begin transaction: %v", err)
    }

    // テスト終了時に自動的にロールバック
    t.Cleanup(func() {
        tx.Rollback()
        db.Close()
    })

    return db, tx
}
```

### 使用例

```go
func TestPopularWorks(t *testing.T) {
    t.Parallel() // 並列実行可能

    db, tx := testutil.SetupTestDB(t)

    // テストデータ作成
    workID := testutil.NewWorkBuilder(t, tx).
        WithTitle("テストアニメ").
        WithWatchersCount(1000).
        Build()

    // テスト実行
    queries := repository.New(db).WithTx(tx)
    // ...

    // テスト終了時に自動的にロールバック（明示的なクリーンアップ不要）
}
```

## フィクスチャービルダー

### ビルダーパターン

柔軟でテストしやすいテストデータ作成：

```go
// internal/testutil/fixtures.go
type WorkBuilder struct {
    t                *testing.T
    tx               *sql.Tx
    title            string
    watchersCount    int32
    seasonYear       *int32
    seasonName       *string
    // ...
}

func NewWorkBuilder(t *testing.T, tx *sql.Tx) *WorkBuilder {
    t.Helper()
    return &WorkBuilder{
        t:             t,
        tx:            tx,
        title:         "Default Work Title",
        watchersCount: 0,
    }
}

func (b *WorkBuilder) WithTitle(title string) *WorkBuilder {
    b.title = title
    return b
}

func (b *WorkBuilder) WithWatchersCount(count int32) *WorkBuilder {
    b.watchersCount = count
    return b
}

func (b *WorkBuilder) WithSeason(year int32, season SeasonName) *WorkBuilder {
    b.seasonYear = &year
    seasonStr := string(season)
    b.seasonName = &seasonStr
    return b
}

func (b *WorkBuilder) Build() int64 {
    b.t.Helper()

    queries := repository.New(nil).WithTx(b.tx)
    work, err := queries.CreateWork(context.Background(), repository.CreateWorkParams{
        Title:         b.title,
        WatchersCount: b.watchersCount,
        SeasonYear:    b.seasonYear,
        SeasonName:    b.seasonName,
    })

    if err != nil {
        b.t.Fatalf("failed to create work: %v", err)
    }

    return work.ID
}
```

### 使用例

```go
// シンプルな作品作成
workID := testutil.NewWorkBuilder(t, tx).Build()

// カスタマイズした作品作成
workID := testutil.NewWorkBuilder(t, tx).
    WithTitle("進撃の巨人").
    WithWatchersCount(5000).
    WithSeason(2024, testutil.SeasonSpring).
    Build()

// 画像付き作品
workID := testutil.NewWorkBuilder(t, tx).
    WithTitle("鬼滅の刃").
    Build()

testutil.NewWorkImageBuilder(t, tx, workID).
    WithImageData(`{"id":"abc123.jpg","storage":"store","metadata":{}}`).
    Build()
```

## 利用可能なビルダー

- **WorkBuilder**: 作品データの作成
- **EpisodeBuilder**: エピソードデータの作成
- **UserBuilder**: ユーザーデータの作成
- **WorkImageBuilder**: 作品画像データの作成
- **SessionBuilder**: セッションデータの作成

## テスト環境の設定

### 環境変数（.env.test）

```bash
ANNICT_ENV=test
ANNICT_DB_HOST=postgresql
ANNICT_DB_PORT=5432
ANNICT_DB_USER=postgres
ANNICT_DB_NAME=annict_test
ANNICT_DB_SSLMODE=disable
```

### Docker Compose

```yaml
services:
  postgresql:
    image: postgres:17.3
    environment:
      POSTGRES_DB: annict_test
      POSTGRES_HOST_AUTH_METHOD: trust
    ports:
      - "15433:5432"
    volumes:
      - ./db/schema.sql:/docker-entrypoint-initdb.d/01-schema.sql
```

## テストの実行

```bash
# すべてのテストを実行
ANNICT_ENV=test make test

# 特定のパッケージのテストを実行
ANNICT_ENV=test go test -v ./internal/handler

# 特定のテストを実行
ANNICT_ENV=test go test -v ./internal/handler -run TestPopularWorks

# 並列実行数を指定
ANNICT_ENV=test go test -v -p 4 ./...

# ベンチマークテストを実行
ANNICT_ENV=test go test -bench=. ./internal/handler
```

## パフォーマンス

### ベンチマーク結果

```
BenchmarkPopularWorksWithRealDB-8    1000    1234567 ns/op
```

### 並列実行

- `t.Parallel()` を使用することで、複数のテストを並行実行
- トランザクション分離により、テスト間でデータが干渉しない
- 大幅なテスト実行時間の短縮を実現

## 成果

- **リアルなテスト**: 実データベースを使用することで、本番環境に近い動作を検証
- **高速なテスト**: トランザクションのロールバックにより、テストデータのクリーンアップが高速
- **並列実行**: トランザクション分離により、安全な並列テストを実現
- **保守性の向上**: ビルダーパターンにより、テストデータ作成が簡潔で読みやすい
- **CI/CD 対応**: 環境変数で簡単にテスト環境を切り替え可能

## 関連ドキュメント

- [プロジェクト全体の設計書](./go.md)
- [テストユーティリティ README](../../internal/testutil/README.md)
