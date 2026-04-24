---
paths:
  - "go/**/*.{go,templ}"
---

# テストガイド

このドキュメントは、Go 版 Wikino のテスト戦略とベストプラクティスを説明します。

## 基本方針

- **実データベースを使用**: 基本的にデータベースをモックせず、実際の PostgreSQL データベースを使用してテストを実行
- **DB 接続プールの共有**: `TestMain` パターンでパッケージ単位で DB 接続を 1 回だけ確立し、全テストで共有
- **トランザクションでの分離**: 各テストはトランザクション内で実行し、テスト終了時に自動ロールバックすることでデータをクリーンアップ
- **テスト用 bcrypt コストの低減**: テスト時は bcrypt コストを最小値（4）に設定し、パスワードハッシュの計算を高速化
- **テストヘルパーの活用**: `internal/testutil` パッケージのヘルパー関数とビルダーパターンを使用してテストデータを作成
- **自動スキーマセットアップ**: `make test` を実行すると、テスト用データベースが自動的にリセットされ、`db/schema.sql` が適用される

## テストの構造

- **テストファイル**: `*_test.go` という接尾辞を付けて同じディレクトリに配置
- **テスト関数**: `Test` で始まる名前（例: `TestPopularWorks`）
- **ベンチマーク関数**: `Benchmark` で始まる名前（例: `BenchmarkPopularWorks`）

## DB 接続の共有化（TestMain パターン）

各テストパッケージでは `TestMain` を使用し、DB 接続を 1 回だけ確立してパッケージ内の全テストで共有します。

**セットアップの流れ**:

```
TestMain: sql.Open → Ping → bcryptコスト設定
テスト1: Begin → テスト実行 → Rollback
テスト2: Begin → テスト実行 → Rollback
TestMain: Close
```

**新規テストパッケージの作成手順**:

1. `main_test.go` を作成し、`TestMain` で `testutil.SetupTestMain` を呼び出す
2. 各テスト関数では `testutil.SetupTx(t)` でトランザクションを取得
3. Usecase などトランザクション管理を自前で行うテストでは `testutil.GetTestDB()` を使用

```go
// main_test.go
package handler_test

import (
    "os"
    "testing"

    "github.com/wikinoapp/wikino/go/internal/testutil"
)

func TestMain(m *testing.M) {
    os.Exit(testutil.SetupTestMain(m))
}
```

```go
// create_test.go
func TestCreate_Success(t *testing.T) {
    t.Parallel()

    db, tx := testutil.SetupTx(t)
    // 以降は既存のテストコードと同じ
}
```

**`SetupTestMain` が行う初期化**:

- テスト用に bcrypt コストを下げる（DefaultCost 10 → MinCost 4 で約 64 倍高速化）
- DB 接続プールを 1 回だけ確立し、パッケージ内の全テストで共有

## `SetupTx` と `GetTestDB` の使い分け

| ヘルパー      | 用途                                           | トランザクション                   |
| ------------- | ---------------------------------------------- | ---------------------------------- |
| `SetupTx(t)`  | Handler、Repository、Validator のテスト        | 自動（テスト終了時にロールバック） |
| `GetTestDB()` | Usecase のテスト（自前でトランザクション管理） | 手動（Usecase 内で管理）           |

## レイヤーごとのテストカバレッジ

新しい実装ファイルを追加した場合は、対応するテストファイルも必ず作成する。Handler テストはエンドポイントの統合テストとして重要だが、UseCase や Validator のロジックも独立してテストする必要がある。

| レイヤー       | テスト対象                     | テストの目的                                               | テストファイルの必須度 |
| -------------- | ------------------------------ | ---------------------------------------------------------- | ---------------------- |
| **Handler**    | HTTP リクエスト・レスポンス    | 認証・認可チェック、ステータスコード、リダイレクト先の検証 | 必須                   |
| **UseCase**    | ビジネスロジック・永続化処理   | トランザクション内の永続化が正しく行われるかの検証         | 必須                   |
| **Validator**  | バリデーションルール           | 形式チェック・状態チェックの検証                           | 必須                   |
| **Repository** | データアクセス・Model への変換 | クエリ結果が正しく Model に変換されるかの検証              | 必須                   |
| **ViewModel**  | 表示用データ変換               | Model から ViewModel への変換が正しいかの検証              | 推奨                   |
| **Model**      | 純粋なドメインエンティティ     | ビジネスロジックを持つメソッドがある場合のみテスト         | 条件付き               |
| **Policy**     | 認可ルール                     | 権限判定が正しいかの検証                                   | 必須                   |

**重要**: Handler テストだけでは UseCase のロジックを十分にテストできない。Handler テストはエンドポイントの振る舞い（認証・認可・リダイレクト）に集中し、UseCase テストはビジネスロジック（永続化結果・トランザクション整合性）に集中する。両方のテストがあることで、責務ごとに独立したテストカバレッジが確保される。

## テストのベストプラクティス

- **実データベースを使用**: モックではなく実際の PostgreSQL データベースでテスト
- **TestMain パターン**: 各テストパッケージに `main_test.go` を作成し、`testutil.SetupTestMain(m)` で DB 接続を共有
- **トランザクション分離**: `testutil.SetupTx(t)` でテスト用トランザクションをセットアップ
- **テーブル駆動テスト**: 複数のテストケースを効率的に実行
- **並行テスト**: すべてのトップレベルテスト関数（`func TestXxx(t *testing.T)`）の先頭で必ず `t.Parallel()` を呼ぶ。テストデータにユニークな識別子を使用しているため、トランザクション分離パターン（`SetupTx`）でも直接 DB アクセスパターン（`GetTestDB`）でも並行実行が安全
- **テストヘルパー**: 共通のセットアップコードをヘルパー関数に抽出
- **エラーケースを必ずテスト**: 正常系だけでなく異常系も網羅

## テーブル駆動テストの書き方

複数のテストケースを効率的に実行するために、テーブル駆動テストを使用します。

**基本パターン**:

```go
func TestCreateAccount(t *testing.T) {
    t.Parallel()

    db, tx := testutil.SetupTx(t)
    ctx := context.Background()

    // テスト対象のセットアップ（共通部分）
    userRepo := repository.NewUserRepository(db).WithTx(tx)
    profileRepo := repository.NewProfileRepository(db).WithTx(tx)
    uc := usecase.NewCreateAccountUsecase(db, userRepo, profileRepo)

    // テストケースの定義
    tests := []struct {
        name    string
        input   usecase.CreateAccountInput
        wantErr bool
    }{
        {
            name: "正常系: 有効な入力でアカウントを作成できる",
            input: usecase.CreateAccountInput{
                Email:    "test@example.com",
                Atname:   "testuser",
                Password: "password123",
            },
            wantErr: false,
        },
        {
            name: "正常系: 日本語パスワードでアカウントを作成できる",
            input: usecase.CreateAccountInput{
                Email:    "japanese@example.com",
                Atname:   "japaneseuser",
                Password: "パスワード123",
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := uc.Execute(ctx, tt.input)

            if tt.wantErr {
                if err == nil {
                    t.Error("expected error but got nil")
                }
                return
            }

            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            if result == nil {
                t.Fatal("result should not be nil")
            }
        })
    }
}
```

**テーブル駆動テストを使用する場面**:

- 同じロジックに対して複数の入力パターンをテストする場合
- 正常系と異常系を網羅的にテストする場合
- バリデーションルールを複数テストする場合

**テーブル駆動テストを使用しない場面**:

- テストケースごとに異なるセットアップが必要な場合
- 各テストの検証ロジックが大きく異なる場合
- 単一のシンプルなテストケースの場合

## 実データベーステストの例

```go
func TestPopularWorks(t *testing.T) {
    // 共有DB接続プールからトランザクションをセットアップ
    db, tx := testutil.SetupTx(t)

    // テストデータを作成（ビルダーパターン）
    workID := testutil.NewWorkBuilder(t, tx).
        WithTitle("テストアニメ").
        WithSeason(2024, testutil.SeasonSpring).
        Build()

    // sqlcリポジトリを作成（トランザクションを使用）
    queries := repository.New(db).WithTx(tx)

    // ハンドラーを作成してテスト実行
    handler := &Handler{
        queries: queries,
        cfg:     cfg,
        templates: templates,
    }

    // HTTPリクエストとレスポンスのテスト
    req := httptest.NewRequest("GET", "/works/popular", nil)
    rr := httptest.NewRecorder()
    handler.PopularWorks(rr, req)

    // アサーション
    if rr.Code != http.StatusOK {
        t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
    }

    // テスト終了時にトランザクションは自動的にロールバックされる
}
```

## テストヘルパーの使用

`internal/testutil` パッケージには以下のヘルパーが用意されています：

**DB 接続・トランザクション**:

- **`SetupTestMain(m)`**: `TestMain` 内で呼び出し、パッケージ共有の DB 接続を初期化。bcrypt コストの低減も行う
- **`SetupTx(t)`**: 共有 DB 接続プールからトランザクションを取得。テスト終了時に自動ロールバック
- **`GetTestDB()`**: 共有 DB 接続プールへの参照を取得。Usecase などトランザクション管理を自前で行うテストで使用

**テストデータビルダー**:

- **`NewUserBuilder(t, tx)`**: ユーザーデータのビルダー
- **`NewWorkBuilder(t, tx)`**: 作品データのビルダー
- **`NewEpisodeBuilder(t, tx, workID)`**: エピソードデータのビルダー
- **`NewWorkImageBuilder(t, tx, workID)`**: 作品画像データのビルダー
- **`CreateTestWork(t, tx, title)`**: 簡易的な作品作成ヘルパー

## 個別テスト実行タスク

開発効率を向上させるため、特定のパッケージやテストのみを実行する Makefile タスクを用意しています。これらのタスクは自動的に 1Password CLI のラッパー経由で環境変数を設定し、DB セットアップも実行します。

**`make test-pkg PKG=<パッケージ>`** - 特定のパッケージのテストを実行:

```sh
make test-pkg PKG=internal/handler/password_reset
make test-pkg PKG=internal/turnstile
```

**`make test-run PKG=<パッケージ> RUN=<テスト名>`** - 特定のテストを実行（パターンマッチ可能）:

```sh
make test-run PKG=internal/handler/password_reset RUN=TestCreate_TurnstileVerification
make test-run PKG=internal/handler/password_reset RUN=TestCreate_RateLimiting
```

**`make test-verbose`** - 詳細ログ付きで全テストを実行（デバッグ用）:

```sh
make test-verbose
```

## テンプレートレンダリングのテスト

templ テンプレートのレンダリングも含めたテストを実装します。

### 基本方針

- **ハンドラーテスト**: HTTP リクエスト・レスポンスを含めた統合テスト
- **コンポーネント単体テスト**: `Render()` メソッドを使って直接テスト
- **テーブル駆動テスト**: 複数のロケールやケースをまとめてテスト
- **HTML 出力の検証**: 特定の要素やテキストが正しくレンダリングされているか確認

詳しい書き方やコード例は [@.claude/rules/go-templ.md](go-templ.md) のテストセクションを参照してください。
