# HTTPハンドラーのガイドライン

このドキュメントは、Annict Go版でHTTPハンドラーを実装する際のガイドラインを提供します。

## 目次

- [基本方針](#基本方針)
- [ディレクトリ構造](#ディレクトリ構造)
- [ファイル命名規則](#ファイル命名規則)
- [メソッド命名規則](#メソッド命名規則)
- [Handler構造体の定義](#handler構造体の定義)
- [依存性注入のガイドライン](#依存性注入のガイドライン)
- [バリデーターの配置](#バリデーターの配置)
- [ルーティング登録](#ルーティング登録)
- [実装例](#実装例)

## 基本方針

HTTPハンドラーは以下の原則に従って実装します：

- **リソースごとにディレクトリを切る**: すべてのエンドポイントはリソースディレクトリを作成
- **1エンドポイント = 1ハンドラーファイル**: 各エンドポイントは個別のファイルに実装
- **統一された命名規則**: ファイル名とメソッド名に一貫性を持たせる
- **例外なくディレクトリ化**: 単独のエンドポイントでも必ずディレクトリを作成

### なぜこの方針を採用するのか

- **可読性の向上**: ファイルサイズが小さく保たれ、コードが理解しやすい
- **保守性の向上**: 1ファイル1責務の原則により、変更箇所が明確
- **拡張性の確保**: 新規機能追加時にどこに実装すべきか迷わない
- **並行開発の促進**: ファイル衝突が発生しにくい

## ディレクトリ構造

すべてのハンドラーは `internal/handler/` 配下にリソースディレクトリを作成します：

```
internal/handler/
├── popular_work/
│   ├── handler.go       # Handler構造体と依存性
│   ├── index.go         # Index メソッド (GET /works/popular)
│   └── index_test.go    # Index のテスト
├── password_reset/
│   ├── handler.go              # Handler構造体と依存性
│   ├── new.go                  # New (GET /password/reset) - リセット申請フォーム
│   ├── create.go               # Create (POST /password/reset)
│   ├── validator.go            # バリデーション（形式チェック + DBを使った検証）
│   ├── validator_test.go       # バリデーションのテスト
│   └── handler_test.go         # ハンドラーの統合テスト
├── password/
│   ├── handler.go              # Handler構造体と依存性
│   ├── edit.go                 # Edit (GET /password/edit)
│   ├── update.go               # Update (PATCH /password)
│   ├── validator.go            # バリデーション
│   ├── validator_test.go       # バリデーションのテスト
│   └── handler_test.go         # ハンドラーの統合テスト
├── sign_in/
│   ├── handler.go              # Handler構造体と依存性
│   ├── new.go                  # New (GET /sign_in) - サインインフォーム
│   ├── create.go               # Create (POST /sign_in)
│   ├── validator.go            # バリデーション（形式チェック + ユーザー検索、パスワード照合）
│   ├── validator_test.go       # バリデーションのテスト
│   └── handler_test.go         # ハンドラーの統合テスト
├── health/
│   ├── handler.go       # Handler構造体と依存性
│   ├── show.go          # Show (GET /health) - ヘルスチェック
│   └── show_test.go     # Show のテスト
├── home/
│   ├── handler.go       # Handler構造体と依存性
│   ├── show.go          # Show (GET /) - ホームページ
│   └── show_test.go     # Show のテスト
└── manifest/
    ├── handler.go       # Handler構造体と依存性
    ├── show.go          # Show (GET /manifest.json) - PWAマニフェスト
    └── show_test.go     # Show のテスト
```

### リソースディレクトリの原則

- **すべてのエンドポイントをディレクトリ化**: 例外なく、すべてのエンドポイントはリソースディレクトリを作成
- **統一性**: 単独のエンドポイントでも必ずディレクトリを作成（例: `health/`, `home/`, `manifest/`）
- **拡張性**: 将来的にエンドポイントが追加されても容易に対応可能
- **リソース名**: 名詞として成立するリソース名を使用（例: `popular_work/`, `password_reset/`, `sign_in/`）

### リソース名の命名規則

- リソース名は**名詞**にする（例: `popular_work`, `password_reset`, `sign_in`）
- 形容詞+名詞の組み合わせの場合は英語の自然な語順にする（例: `popular_work` ⭕️、`work_popular` ❌）
- URL パターンから名詞を抽出してリソース名を決定

**命名例**:

| URL パターン    | リソース名（推奨）    | 理由                         |
| --------------- | --------------------- | ---------------------------- |
| /works/popular  | `popular_work/` ⭕️   | 形容詞+名詞の自然な語順      |
| /works/popular  | `work_popular/` ❌    | 「作品\_人気」は不自然       |
| /password/reset | `password_reset/` ⭕️ | 名詞として成立               |
| /users/me       | `current_user/` ⭕️   | 「現在のユーザー」という名詞 |
| /search         | `search/` ⭕️         | 「検索」という名詞           |

## ファイル命名規則

### 標準ファイル名（9 種類のみ）

リソースディレクトリ内には、以下の標準的なファイル名**のみ**を使用します：

**ハンドラー関連**:

- `handler.go` - Handler 構造体と依存性の定義
- `index.go` - 一覧ページ表示 (GET /resources)
- `show.go` - 個別ページ表示 (GET /resources/:id)
- `new.go` - 新規作成フォーム表示 (GET /resources/new)
- `create.go` - 作成処理 (POST /resources)
- `edit.go` - 編集フォーム表示 (GET /resources/:id/edit)
- `update.go` - 更新処理 (PATCH /resources/:id)
- `delete.go` - 削除処理 (DELETE /resources/:id)

**バリデーション関連**:

- `validator.go` - バリデーション（形式チェック + DB を使った検証を統合）

### 重要な原則

- 上記 9 種類以外のファイル名は**使用しない**
- 複雑な名前（`show_reset_form.go`, `process_reset.go` など）が必要な場合は、**新しいリソースディレクトリを作成する**
- 例: `password/show_reset_form.go` ではなく、`password_reset/new.go` を使用

### テストファイル

- **ハンドラーテスト**: `handler_test.go`（ハンドラー全体の統合テスト）
- **バリデーションテスト**: `validator_test.go`（形式バリデーション + 状態バリデーション）
- **個別ハンドラーテスト**: `{action}_test.go`（例: `index_test.go`, `show_test.go`）も許可

## メソッド命名規則

リソースディレクトリ内では以下のメソッド名を使用します：

- `Index` - 一覧ページ表示（ファイル名: `index.go`）
- `Show` - 個別ページ表示（ファイル名: `show.go`）
- `New` - 新規作成フォーム表示（ファイル名: `new.go`）
- `Create` - 作成処理（ファイル名: `create.go`）
- `Edit` - 編集フォーム表示（ファイル名: `edit.go`）
- `Update` - 更新処理（ファイル名: `update.go`）
- `Delete` - 削除処理（ファイル名: `delete.go`）

### ファイル名とメソッド名の対応

| HTTP メソッド | URL 例          | ファイル名  | メソッド名 | 説明         |
| ------------- | --------------- | ----------- | ---------- | ------------ |
| GET           | /works          | `index.go`  | `Index`    | 一覧取得     |
| GET           | /works/:id      | `show.go`   | `Show`     | 個別取得     |
| GET           | /works/new      | `new.go`    | `New`      | 新規フォーム |
| POST          | /works          | `create.go` | `Create`   | 作成処理     |
| GET           | /works/:id/edit | `edit.go`   | `Edit`     | 編集フォーム |
| PATCH         | /works/:id      | `update.go` | `Update`   | 更新処理     |
| DELETE        | /works/:id      | `delete.go` | `Delete`   | 削除処理     |

**ファイル名とメソッド名の完全一致**:

- すべてのファイル名とメソッド名が完全に対応しています（例: `index.go` → `Index` メソッド、`show.go` → `Show` メソッド）
- この一貫性により、コードの可読性と保守性が向上します
- HTTP メソッド（GET/POST/PATCH/DELETE）とハンドラーメソッド名（Index/Show/Create/Update/Delete）は明確に区別されています

**重要**: 複雑なメソッド名（`ShowResetForm`, `ProcessReset` など）は**使用しない**。代わりに新しいリソースディレクトリを作成し、標準的なメソッド名を使用します。

**例**:

- ❌ `password.ShowResetForm()` ではなく、✅ `password_reset.New()` を使用
- ❌ `password.ProcessReset()` ではなく、✅ `password_reset.Create()` を使用

### エンドポイントとリソースの対応例

#### 複数アクションを持つリソース

| エンドポイント       | リソースディレクトリ | ファイル名  | メソッド名 | 説明                           |
| -------------------- | -------------------- | ----------- | ---------- | ------------------------------ |
| GET /works/popular   | `popular_work/`      | `index.go`  | `Index`    | 人気作品一覧                   |
| GET /password/reset  | `password_reset/`    | `new.go`    | `New`      | パスワードリセット申請フォーム |
| POST /password/reset | `password_reset/`    | `create.go` | `Create`   | パスワードリセット申請処理     |
| GET /password/edit   | `password/`          | `edit.go`   | `Edit`     | パスワード変更フォーム         |
| PATCH /password      | `password/`          | `update.go` | `Update`   | パスワード変更処理             |
| GET /sign_in         | `sign_in/`           | `new.go`    | `New`      | サインインフォーム             |
| POST /sign_in        | `sign_in/`           | `create.go` | `Create`   | サインイン処理                 |

#### 単独エンドポイント（1つのアクションのみ持つリソース）

| エンドポイント     | リソースディレクトリ | ファイル名 | メソッド名 | 説明             |
| ------------------ | -------------------- | ---------- | ---------- | ---------------- |
| GET /              | `home/`              | `show.go`  | `Show`     | ホームページ     |
| GET /health        | `health/`            | `show.go`  | `Show`     | ヘルスチェック   |
| GET /manifest.json | `manifest/`          | `show.go`  | `Show`     | PWA マニフェスト |

## Handler構造体の定義

各リソースディレクトリの `handler.go` に Handler構造体と依存性を定義します。

### 基本的な構造

```go
// handler/popular_work/handler.go
package popular_work

import (
    "github.com/annict/annict/go/internal/config"
    repository "github.com/annict/annict/go/internal/repository/sqlc"
)

// Handler は人気作品関連のHTTPハンドラーです
type Handler struct {
    cfg     *config.Config
    queries *repository.Queries
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, queries *repository.Queries) *Handler {
    return &Handler{
        cfg:     cfg,
        queries: queries,
    }
}
```

### 命名規則

- **構造体名**: `Handler` （すべてのリソースで統一）
- **コンストラクタ名**: `NewHandler` （すべてのリソースで統一）
- **パッケージ名**: リソース名と同じ（例: `popular_work`, `password_reset`）

## 依存性注入のガイドライン

Handler構造体には、そのリソースで必要な依存性を注入します。

### 基本方針

1. **共通の依存性**: すべてのエンドポイントで使う依存性は必ず含める
2. **専用の依存性**: 一部のエンドポイントでしか使わない依存性も含めてOK（多少の冗長性は許容）
3. **肥大化の防止**: Handler構造体のフィールドが**8個を超えたら**、リソース分割を検討
4. **段階的な設計**: 最初は必要最小限から始め、必要に応じて依存性を追加

### 許容される冗長性の例

```go
// ✅ 良い例: update.goでしか使わない依存性も含める（3-4個程度なら許容）
type Handler struct {
    cfg                   *config.Config
    queries               *repository.Queries
    db                    *sql.DB
    sessionMgr            *session.Manager
    updatePasswordUsecase *usecase.UpdatePasswordUsecase  // update.goでしか使わない
}
```

**理由**:

- 依存性が明示的で、このリソースが何に依存しているか一目でわかる
- ポインタなので未使用でもメモリオーバーヘッドは小さい（8バイト程度）
- 将来的に他のエンドポイントでも使う可能性がある

### 肥大化の警告

```go
// ⚠️ 注意: フィールドが8個を超えたらリソース分割を検討
type Handler struct {
    cfg            *config.Config
    queries        *repository.Queries
    db             *sql.DB
    sessionMgr     *session.Manager
    limiter        *ratelimit.Limiter
    riverClient    *worker.Client
    createUsecase  *usecase.CreatePasswordUsecase
    updateUsecase  *usecase.UpdatePasswordUsecase
    resetUsecase   *usecase.ResetPasswordUsecase
    // ↑ このような場合は、リソースを更に細分化すべき
}
```

### リソース分割の例

依存性が多い場合は、以下のようにリソースを分割します：

```
# 分割前（肥大化）
password/
├── handler.go  # Handler構造体（10個の依存性）
├── edit.go
├── update.go
├── reset.go
└── confirm.go

# 分割後（適切なサイズ）
password_edit/
├── handler.go  # 2-3個の依存性
└── edit.go

password_update/
├── handler.go  # 2-3個の依存性
└── update.go

password_reset/
├── handler.go  # 2-3個の依存性
├── new.go
└── create.go
```

## バリデーターの配置

バリデーション用のファイルは、各リソースディレクトリに `validator.go` として配置します。

### バリデーター（validator.go）

形式チェック（必須チェック、メール形式、文字数制限など）と DB を使った検証（ユーザー存在チェック、パスワード照合など）を 1 つのファイルに統合します。

- **構造体名**: `{Action}Validator`（例: `CreateValidator`, `UpdateValidator`）
- **メソッド**: `Validate(ctx context.Context, input {Action}ValidatorInput) *{Action}ValidatorResult`

### 配置例

```
internal/handler/
├── sign_in/
│   ├── handler.go
│   ├── new.go
│   ├── create.go
│   ├── validator.go            # バリデーション（形式チェック + ユーザー検索、パスワード照合）
│   ├── validator_test.go       # バリデーションのテスト
│   └── handler_test.go         # ハンドラーの統合テスト
├── password_reset/
│   ├── handler.go
│   ├── new.go
│   ├── create.go
│   ├── validator.go            # バリデーション（形式チェックのみの場合もあり）
│   ├── validator_test.go       # バリデーションのテスト
│   └── handler_test.go         # ハンドラーの統合テスト
├── password/
│   ├── handler.go
│   ├── edit.go
│   ├── update.go
│   ├── validator.go            # バリデーション
│   ├── validator_test.go       # バリデーションのテスト
│   └── handler_test.go         # ハンドラーの統合テスト
```

**注**: すべてのリソースに `validator.go` が必要なわけではありません。バリデーションが必要な場合のみ作成します。

詳細なバリデーション実装については、[@go/docs/validation-guide.md](validation-guide.md) を参照してください。

## ルーティング登録

`cmd/server/main.go` でルーティングを登録する際は、リソースごとにハンドラーを初期化します。

### 基本的な登録方法

```go
// ハンドラーの初期化
popularWorkHandler := popular_work.NewHandler(cfg, queries)
passwordResetHandler := password_reset.NewHandler(cfg, db, queries, sessionMgr, limiter, riverClient)
passwordHandler := password.NewHandler(cfg, db, queries, sessionMgr)
signInHandler := sign_in.NewHandler(cfg, queries, sessionMgr)

// ルーティング登録
r.Get("/works/popular", popularWorkHandler.Index)
r.Get("/password/reset", passwordResetHandler.New)
r.Post("/password/reset", passwordResetHandler.Create)
r.Get("/password/edit", passwordHandler.Edit)
r.Patch("/password", passwordHandler.Update)
r.Get("/sign_in", signInHandler.New)
r.Post("/sign_in", signInHandler.Create)
```

### ルーティングの原則

- **標準的なHTTPメソッドを使用**: GET/POST/PATCH/DELETE
- **更新処理はPATCHを使用**: 部分更新を表現し、Rails との整合性も保つ
- **Method Overrideパターン**: HTMLフォームからPATCH/DELETEを実現（詳細は [@go/CLAUDE.md](../CLAUDE.md#httpメソッドとルーティング) を参照）

## 実装例

### 例1: 単独エンドポイント（health/）

```go
// internal/handler/health/handler.go
package health

import (
    "github.com/annict/annict/go/internal/config"
)

// Handler はヘルスチェック関連のHTTPハンドラーです
type Handler struct {
    cfg *config.Config
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config) *Handler {
    return &Handler{
        cfg: cfg,
    }
}
```

```go
// internal/handler/health/show.go
package health

import (
    "net/http"
)

// Show GET /health - ヘルスチェック
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

### 例2: 複数エンドポイント（password_reset/）

```go
// internal/handler/password_reset/handler.go
package password_reset

import (
    "database/sql"
    "github.com/annict/annict/go/internal/config"
    repository "github.com/annict/annict/go/internal/repository/sqlc"
    "github.com/annict/annict/go/internal/session"
    "github.com/annict/annict/pkg/ratelimit"
    "github.com/riverqueue/river"
)

// Handler はパスワードリセット関連のHTTPハンドラーです
type Handler struct {
    cfg         *config.Config
    db          *sql.DB
    queries     *repository.Queries
    sessionMgr  *session.Manager
    limiter     *ratelimit.Limiter
    riverClient *river.Client
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
    cfg *config.Config,
    db *sql.DB,
    queries *repository.Queries,
    sessionMgr *session.Manager,
    limiter *ratelimit.Limiter,
    riverClient *river.Client,
) *Handler {
    return &Handler{
        cfg:         cfg,
        db:          db,
        queries:     queries,
        sessionMgr:  sessionMgr,
        limiter:     limiter,
        riverClient: riverClient,
    }
}
```

```go
// internal/handler/password_reset/new.go
package password_reset

import (
    "net/http"
    "github.com/annict/annict/go/internal/templates/pages/password_reset"
)

// New GET /password/reset - パスワードリセット申請フォーム
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
    // 実装
}
```

```go
// internal/handler/password_reset/create.go
package password_reset

import (
    "net/http"
)

// Create POST /password/reset - パスワードリセット申請処理
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    // 実装
}
```

```go
// internal/handler/password_reset/validator.go
package password_reset

import (
    "context"
    "regexp"

    "github.com/annict/annict/go/internal/session"
    "github.com/annict/annict/go/internal/templates"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// CreateValidator はパスワードリセット申請のバリデーションを行う
type CreateValidator struct{}

// NewCreateValidator は CreateValidator を生成する
func NewCreateValidator() *CreateValidator {
    return &CreateValidator{}
}

// CreateValidatorInput はバリデーションの入力パラメータ
type CreateValidatorInput struct {
    Email string
}

// CreateValidatorResult はバリデーションの結果
type CreateValidatorResult struct {
    FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateValidator) Validate(ctx context.Context, input CreateValidatorInput) *CreateValidatorResult {
    formErrors := session.NewFormErrors()

    if input.Email == "" {
        formErrors.AddFieldError("email", templates.T(ctx, "error_required"))
        return &CreateValidatorResult{FormErrors: formErrors}
    }

    if !emailRegex.MatchString(input.Email) {
        formErrors.AddFieldError("email", templates.T(ctx, "error_invalid_email_format"))
    }

    return &CreateValidatorResult{FormErrors: formErrors}
}
```

## 設計思想

この命名規則は、以下の設計思想に基づいています：

- **RESTful設計**: Railsをはじめ多くのWebフレームワークが採用する標準的な7アクション（index, show, new, create, edit, update, delete）
- **完全な対応**: ファイル名とメソッド名が100%一致
- **HTTPメソッドとの区別**: HTTPメソッド（GET/POST/PATCH/DELETE）とハンドラーメソッド名（Index/Show/Create/Update/Delete）は明確に区別
- **Railsからの移行に最適**: 既存のRails開発者が即座に理解できる命名
- **MPA向け最適化**: HTMLレンダリング型Webアプリケーションに適した命名

## まとめ

HTTPハンドラーを実装する際は、以下のポイントを守ってください：

1. **すべてのエンドポイントをディレクトリ化**: 例外なく、リソースディレクトリを作成
2. **標準的なファイル名を使用**: 9 種類のファイル名のみを使用
   - ハンドラー関連: `handler.go`, `index.go`, `show.go`, `new.go`, `create.go`, `edit.go`, `update.go`, `delete.go`
   - バリデーション関連: `validator.go`
3. **ファイル名とメソッド名を一致させる**: 可読性と保守性を向上
4. **依存性注入を適切に管理**: 8 個以下のフィールドを目安に、必要に応じてリソースを分割
5. **バリデーションは `validator.go` に統合**: 形式バリデーションと状態バリデーションを 1 ファイルで管理

これらの規則を守ることで、以下のメリットが得られます：

- コードの可読性と保守性が向上する
- 新規機能追加時に迷わない
- 並行開発がスムーズになる
- テストが書きやすくなる

## 関連ドキュメント

- [@go/CLAUDE.md](../CLAUDE.md) - Go版開発ガイド
- [@go/docs/validation-guide.md](validation-guide.md) - リクエストバリデーションガイド
- [@go/docs/templ-guide.md](templ-guide.md) - templテンプレートガイド
- [@go/docs/architecture-guide.md](architecture-guide.md) - アーキテクチャガイド
