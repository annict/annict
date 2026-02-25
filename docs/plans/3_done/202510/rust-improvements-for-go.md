# Rust 版の優れた実装を Go 版に持ち帰る 設計書

## 概要

別リポジトリでこのプロジェクトを Rust で書き直すとどうなるか？という実験をしていました。
結局 Rust で書き直すことはやめたのですが、Rust 版と Go 版の実装を詳細に比較した結果、Rust 版で優れている設計や実装パターンを多数発見しました。これらの優れた点を Go 版に導入することで、コード品質、型安全性、保守性を大幅に向上させることができます。

**目的**:

- Rust 版で実現されている高品質な実装パターンを Go 版に導入する
- Go 版のコード品質とメンテナンス性を向上させる
- 開発者体験（DX）を改善し、バグの早期発見を可能にする

**背景**:

- Rust 版と Go 版は同じ機能を実装しており、直接比較が可能
- Rust 版は型安全性とドキュメント充実度で優れている
- これらの優れた点の多くは Go でも実現可能

## 要件

### 機能要件

- **バリデーション強化**: Rust 版と同等の厳密なバリデーションを実装する
- **ドキュメント充実**: すべての公開関数・構造体に詳細な日本語コメントを追加する
- **テスト拡充**: Rust 版と同等のテストカバレッジとエッジケーステストを実装する
- **Fail-fast 機能**: データベース制約違反を即座に検出する仕組みを導入する

### 非機能要件

- **保守性**: コードの意図が明確で、将来のメンテナンスが容易であること
- **可読性**: 日本語コメントにより、日本語話者がコードを理解しやすいこと
- **信頼性**: エッジケースや異常系を適切にハンドリングすること
- **開発者体験**: 明示的なエラーメッセージにより、問題の特定が容易であること

## 設計

### Rust 版の優れている点（詳細分析）

#### 1. バリデーションの品質

**Rust 版の実装**:

```rust
// src/handlers/sign_in_request.rs:55-60
if self.email_or_username.trim().is_empty() {
    errors.add_field_error(
        "email_username",
        i18n.t(locale, "sign-in-error-email-username-required"),
    );
}
```

**Go 版の現状**:

```go
// internal/handler/sign_in_request.go:20-22
if req.EmailOrUsername == "" {
    errors.AddFieldError("email_username", i18n.T(ctx, "sign_in_error_email_username_required"))
}
```

**問題点**:

- Go 版は whitespace-only 入力（" "など）をスルーしてしまう
- ユーザーがスペースのみを入力した場合、エラーにならずにデータベースクエリが実行される

**改善案**:

```go
import "strings"

if strings.TrimSpace(req.EmailOrUsername) == "" {
    errors.AddFieldError("email_username", i18n.T(ctx, "sign_in_error_email_username_required"))
}
```

#### 2. ドキュメントの充実度

**Rust 版の実装例**:

```rust
// src/repositories/users.rs:42-56
/// ユーザーをメールアドレスまたはユーザー名で検索する
///
/// # 引数
///
/// * `pool` - データベース接続プール
/// * `email_or_username` - メールアドレスまたはユーザー名
///
/// # 戻り値
///
/// ユーザーが見つかった場合は`Some(User)`、見つからない場合は`None`を返します。
///
/// # エラー
///
/// データベースエラーが発生した場合は`sqlx::Error`を返します。
pub async fn find_by_email_or_username(
    pool: &PgPool,
    email_or_username: &str,
) -> Result<Option<User>, sqlx::Error>
```

**Go 版の現状**:

```go
// internal/repository/sqlc/users.sql.go:60
func (q *Queries) GetUserByEmailOrUsername(ctx context.Context, lower string) (GetUserByEmailOrUsernameRow, error)
```

**問題点**:

- sqlc が生成するコードにはコメントがない
- 引数の意味、戻り値の意味、エラー条件が不明確
- 他の開発者が使い方を理解しにくい

**改善案**:
各ハンドラー、UseCase、手動実装の関数に GoDoc コメントを追加

```go
// GetUserByEmailOrUsername はメールアドレスまたはユーザー名でユーザーを検索します。
//
// # 引数
//   - ctx: コンテキスト
//   - lower: メールアドレスまたはユーザー名（大文字小文字は区別しません）
//
// # 戻り値
//   - GetUserByEmailOrUsernameRow: ユーザー情報
//   - error: ユーザーが見つからない場合はsql.ErrNoRows、その他のエラーはDB操作エラー
func (q *Queries) GetUserByEmailOrUsername(ctx context.Context, lower string) (GetUserByEmailOrUsernameRow, error)
```

#### 3. テストの充実度

**Rust 版の実装**:

```rust
// src/handlers/sign_in_request.rs:115-128
#[test]
fn test_validate_whitespace_only_email_username() {
    let i18n = I18n::new().unwrap();
    let req = SignInRequest {
        email_or_username: "   ".to_string(),
        password: "password123".to_string(),
    };

    let result = req.validate(&i18n, "ja");
    assert!(result.is_err());

    let errors = result.unwrap_err();
    assert!(errors.has_errors());
    assert_eq!(errors.get_field_errors("email_username").len(), 1);
}
```

**Go 版の現状**:

```go
// internal/handler/sign_in_request_test.go:19-24
{
    name:            "正常系",
    emailOrUsername: "testuser",
    password:        "password123",
    wantErrors:      false,
},
```

**問題点**:

- whitespace-only 入力のテストケースがない
- I18n メッセージの実際の内容を検証していない（エラーの存在のみチェック）
- エッジケースのカバレッジが不足

**改善案**:

```go
{
    name:            "メールアドレスがwhitespaceのみ",
    emailOrUsername: "   ",
    password:        "password123",
    wantErrors:      true,
    wantFieldErrors: []string{"email_username"},
},
```

#### 4. Fail-fast 原則の徹底

**Rust 版の実装**:

```rust
// src/repositories/users.rs:22-29
let created_at = row
    .created_at
    .and_then(|dt| DateTime::from_timestamp(dt.unix_timestamp(), 0))
    .expect("created_atがNULLです（データベース制約違反）");
let updated_at = row
    .updated_at
    .and_then(|dt| DateTime::from_timestamp(dt.unix_timestamp(), 0))
    .expect("updated_atがNULLです（データベース制約違反）");
```

**Go 版の現状**:

```go
// internal/repository/sqlc/users.sql.go:56-57
CreatedAt         sql.NullTime `db:"created_at"`
UpdatedAt         sql.NullTime `db:"updated_at"`
```

**問題点**:

- `sql.NullTime`を使用しているため、NULL を許容してしまう
- データベースの NOT NULL 制約があっても、コードレベルではチェックされない
- 問題が発生してもサイレントに処理が続行される可能性がある

**改善案**:

```go
// ユーザー取得後にバリデーションを追加
user, err := queries.GetUserByID(ctx, userID)
if err != nil {
    return nil, err
}

// NOT NULL制約があるフィールドの検証
if !user.CreatedAt.Valid {
    return nil, fmt.Errorf("created_atがNULLです（データベース制約違反）: user_id=%d", userID)
}
if !user.UpdatedAt.Valid {
    return nil, fmt.Errorf("updated_atがNULLです（データベース制約違反）: user_id=%d", userID)
}
```

#### 5. エラーハンドリングの明示性

**Rust 版の実装**:

```rust
// src/usecases/create_session.rs:7-22
#[derive(Debug, Error)]
pub enum CreateSessionError {
    /// データベースエラー
    #[error("データベースエラー: {0}")]
    Database(#[from] sqlx::Error),

    /// JSON シリアライズエラー
    #[error("JSONシリアライズエラー: {0}")]
    JsonSerialization(#[from] serde_json::Error),

    /// ランダム生成エラー
    #[error("ランダム値生成エラー")]
    RandomGeneration,
}
```

**Go 版の現状**:

```go
// internal/usecase/create_session.go:37-42
publicID, err := generatePublicID()
if err != nil {
    return nil, fmt.Errorf("public ID生成エラー: %w", err)
}
```

**比較**:

- Rust 版は専用のエラー型で各エラーケースを明示的に定義
- Go 版は`fmt.Errorf`で文字列ベースのエラーラップ
- Go 版の方がシンプルだが、Rust 版の方が型安全

**評価**:
この点は**Go 版の現状を維持**することを推奨します。理由：

- Go の慣例的なエラーハンドリングパターンに従っている
- カスタムエラー型を導入すると複雑性が増す
- `%w`によるエラーラップで十分な情報が得られる

#### 6. ドキュメント内の使用例

**Rust 版の実装**:

````rust
// src/handlers/sign_in_request.rs:35-50
/// # 例
///
/// ```
/// use annict_next::forms::FormErrors;
/// use annict_next::handlers::sign_in_request::SignInRequest;
/// use annict_next::i18n::I18n;
///
/// let i18n = I18n::new().unwrap();
/// let req = SignInRequest {
///     email_or_username: "".to_string(),
///     password: "".to_string(),
/// };
///
/// let result = req.validate(&i18n, "ja");
/// assert!(result.is_err());
/// ```
````

**Go 版の現状**:
使用例なし

**改善案**:
主要な関数に GoDoc の`Example`形式で使用例を追加

```go
// Example:
//   i18n := i18n.New()
//   req := &SignInRequest{
//       EmailOrUsername: "",
//       Password:        "",
//   }
//   errors := req.Validate(ctx)
//   if errors != nil {
//       // バリデーションエラーを処理
//   }
```

### 技術スタック

- **言語**: Go 1.22 以上
- **ツール**: `go vet`, `staticcheck`, `golangci-lint`
- **テスト**: `go test`, テーブル駆動テスト
- **ドキュメント**: GoDoc コメント

## タスクリスト

### フェーズ 1: バリデーション強化

- [x] バリデーション関数に whitespace チェックを追加
  - `SignInRequest.Validate()`を修正
  - `strings.TrimSpace()`を使用
  - 既存のテストが通ることを確認
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 10 行 + テスト 40 行）

- [x] その他のバリデーション関数も同様に修正
  - パスワードリセットリクエストのバリデーション
  - その他のフォームバリデーション
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 100 行（実装 20 行 + テスト 80 行）

### フェーズ 2: テストの拡充

- [x] バリデーションテストにエッジケースを追加
  - whitespace-only 入力のテストケース追加
  - I18n メッセージの内容検証を追加
  - すべてのバリデーション関数に対して実施
  - **想定ファイル数**: 約 3 ファイル（テスト 3）
  - **想定行数**: 約 150 行（テスト 150 行）

- [x] UseCase テストに fail-fast ケースを追加
  - データベース制約違反のテストケース追加
  - NULL 値が返された場合のテストケース追加
  - **想定ファイル数**: 約 3 ファイル（テスト 3）
  - **想定行数**: 約 120 行（テスト 120 行）

### フェーズ 3: ドキュメント充実

- [ ] ハンドラーに GoDoc コメントを追加
  - すべての公開関数にコメント追加
  - 引数、戻り値、エラー条件を明記
  - 使用例を追加（主要な関数のみ）
  - **想定ファイル数**: 約 5 ファイル（実装 5）
  - **想定行数**: 約 200 行（ドキュメント 200 行）

- [ ] UseCase に GoDoc コメントを追加
  - すべての UseCase 構造体と関数にコメント追加
  - トランザクション管理の説明を追加
  - **想定ファイル数**: 約 3 ファイル（実装 3）
  - **想定行数**: 約 150 行（ドキュメント 150 行）

- [ ] Repository レイヤーに GoDoc コメントを追加
  - 手動実装の関数にコメント追加（sqlc 生成コードは除く）
  - クエリの意図と使用方法を明記
  - **想定ファイル数**: 約 2 ファイル（実装 2）
  - **想定行数**: 約 80 行（ドキュメント 80 行）

### フェーズ 4: Fail-fast 機能の導入

- [x] データベース制約違反の検出機能を追加
  - ヘルパー関数を作成（`ValidateNotNull`など）
  - UseCase でデータベースから取得後にバリデーション実行
  - エラーメッセージに詳細情報を含める
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 80 行 + テスト 120 行）

- [x] ログ出力の改善
  - fail-fast 時のログメッセージを詳細化
  - データベース制約違反を明確に記録
  - **想定ファイル数**: 約 3 ファイル（実装 3）
  - **想定行数**: 約 60 行（実装 60 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **カスタムエラー型の導入**: Go の慣例的なエラーハンドリングパターン（`fmt.Errorf`と`%w`）を維持する。Rust 版のような専用のエラー型は導入しない。
- **型レベルの Null 安全性**: Go の`sql.NullXxx`型を維持する。Rust の`Option`型のような型レベルでの Null 安全性は Go の言語仕様上実現不可能。ランタイムバリデーションで対応する。
- **コンパイル時 SQL チェック**: sqlc の機能で十分であり、Rust 版の sqlx のようなマクロベースのコンパイル時チェックは導入しない。

## 参考資料

- [Rust 版実装](/workspace/src/)
- [Go 版実装](/go-app/internal/)
- [Effective Go - Commentary](https://go.dev/doc/effective_go#commentary)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [sqlc Documentation](https://docs.sqlc.dev/)
