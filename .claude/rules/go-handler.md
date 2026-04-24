---
paths:
  - "go/**/*.{go,templ}"
---

# HTTP ハンドラーのガイドライン

このドキュメントは、Wikino Go 版で HTTP ハンドラーを実装する際のガイドラインを提供します。

## 目次

- [基本方針](#基本方針)
- [ディレクトリ構造](#ディレクトリ構造)
- [ファイル命名規則](#ファイル命名規則)
- [メソッド命名規則](#メソッド命名規則)
- [Handler 構造体の定義](#handler構造体の定義)
- [依存性注入のガイドライン](#依存性注入のガイドライン)
- [バリデーターの配置](#バリデーターの配置)
- [ルーティング登録](#ルーティング登録)
- [実装例](#実装例)
- [エラーハンドリング](#エラーハンドリング)

## 基本方針

HTTP ハンドラーは**薄い Adapter**として、HTTP の入出力変換のみを行います。認可・バリデーション・ビジネスロジックはすべて UseCase に委譲します。

- **Handler は薄い Adapter**: リクエストのパース → UseCase 呼び出し → レスポンス（リダイレクト or テンプレート描画）のみ
- **認可・バリデーションは UseCase 経由**: Handler から Policy・Validator への直接依存は禁止（depguard で強制）
- **リソースごとにディレクトリを切る**: すべてのエンドポイントはリソースディレクトリを作成
- **1 エンドポイント = 1 ハンドラーファイル**: 各エンドポイントは個別のファイルに実装
- **統一された命名規則**: ファイル名とメソッド名に一貫性を持たせる
- **例外なくディレクトリ化**: 単独のエンドポイントでも必ずディレクトリを作成

### なぜこの方針を採用するのか

- **可読性の向上**: ファイルサイズが小さく保たれ、コードが理解しやすい
- **保守性の向上**: 1 ファイル 1 責務の原則により、変更箇所が明確
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
│   └── handler_test.go         # ハンドラーの統合テスト
├── password/
│   ├── handler.go              # Handler構造体と依存性
│   ├── edit.go                 # Edit (GET /password/edit)
│   ├── update.go               # Update (PATCH /password)
│   └── handler_test.go         # ハンドラーの統合テスト
├── sign_in/
│   ├── handler.go              # Handler構造体と依存性
│   ├── new.go                  # New (GET /sign_in) - サインインフォーム
│   ├── create.go               # Create (POST /sign_in)
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

internal/validator/
├── sign_in.go                  # サインインのバリデーション（形式チェック + ユーザー検索、パスワード照合）
├── sign_in_test.go             # サインインのバリデーションテスト
├── password_reset.go           # パスワードリセットのバリデーション（形式チェックのみ）
├── password_reset_test.go      # パスワードリセットのバリデーションテスト
├── password.go                 # パスワード変更のバリデーション
└── password_test.go            # パスワード変更のバリデーションテスト
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

| URL パターン    | リソース名（推奨）   | 理由                         |
| --------------- | -------------------- | ---------------------------- |
| /works/popular  | `popular_work/` ⭕️   | 形容詞+名詞の自然な語順      |
| /works/popular  | `work_popular/` ❌   | 「作品\_人気」は不自然       |
| /password/reset | `password_reset/` ⭕️ | 名詞として成立               |
| /users/me       | `current_user/` ⭕️   | 「現在のユーザー」という名詞 |
| /search         | `search/` ⭕️         | 「検索」という名詞           |

## ファイル命名規則

### 標準ファイル名（8 種類のみ）

リソースディレクトリ内には、以下の標準的なファイル名**のみ**を使用します：

- `handler.go` - Handler 構造体と依存性の定義
- `index.go` - 一覧ページ表示 (GET /resources)
- `show.go` - 個別ページ表示 (GET /resources/:id)
- `new.go` - 新規作成フォーム表示 (GET /resources/new)
- `create.go` - 作成処理 (POST /resources)
- `edit.go` - 編集フォーム表示 (GET /resources/:id/edit)
- `update.go` - 更新処理 (PATCH /resources/:id)
- `delete.go` - 削除処理 (DELETE /resources/:id)

**バリデーション**: すべてのバリデーターは `internal/validator/` パッケージに配置します。詳細は [@.claude/rules/go-validation.md](go-validation.md) を参照してください。

### 重要な原則

- 上記 8 種類以外のファイル名は**使用しない**
- 複雑な名前（`show_reset_form.go`, `process_reset.go` など）が必要な場合は、**新しいリソースディレクトリを作成する**
- 例: `password/show_reset_form.go` ではなく、`password_reset/new.go` を使用

### テストファイル

テストファイル（`*_test.go`）は上記 8 種類の制限に含まれず、必要に応じて作成できます。

- **ハンドラーテスト**: `handler_test.go`（ハンドラー全体の統合テスト）
- **個別ハンドラーテスト**: `{action}_test.go`（例: `index_test.go`, `show_test.go`）も許可
- **バリデーションテスト**: `internal/validator/` パッケージに配置（`{resource}_test.go`）

**テスト関数の配置ルール**: 個別テストファイル（`{action}_test.go`）を使用する場合、テスト関数は対応するアクションのファイルに配置する。異なるアクションのテストを 1 つの `{action}_test.go` に混在させない。

- ✅ `TestUpdate_*` → `update_test.go`、`TestEdit_*` → `edit_test.go`
- ❌ `TestUpdate_*` を `edit_test.go` に配置する

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

#### 単独エンドポイント（1 つのアクションのみ持つリソース）

| エンドポイント     | リソースディレクトリ | ファイル名 | メソッド名 | 説明             |
| ------------------ | -------------------- | ---------- | ---------- | ---------------- |
| GET /              | `home/`              | `show.go`  | `Show`     | ホームページ     |
| GET /health        | `health/`            | `show.go`  | `Show`     | ヘルスチェック   |
| GET /manifest.json | `manifest/`          | `show.go`  | `Show`     | PWA マニフェスト |

## Handler 構造体の定義

各リソースディレクトリの `handler.go` に Handler 構造体と依存性を定義します。

### 基本的な構造

Handler は UseCase を受け取り、HTTP の入出力変換のみを行います。

```go
// handler/suggestion_comment/handler.go
package suggestion_comment

import (
    "github.com/wikinoapp/wikino/go/internal/session"
    "github.com/wikinoapp/wikino/go/internal/usecase"
)

// Handler は編集提案コメント関連のHTTPハンドラーです
type Handler struct {
    flashMgr                       *session.FlashManager
    createSuggestionCommentUsecase *usecase.CreateSuggestionCommentUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
    flashMgr *session.FlashManager,
    createSuggestionCommentUsecase *usecase.CreateSuggestionCommentUsecase,
) *Handler {
    return &Handler{
        flashMgr:                       flashMgr,
        createSuggestionCommentUsecase: createSuggestionCommentUsecase,
    }
}
```

### 命名規則

- **構造体名**: `Handler` （すべてのリソースで統一）
- **コンストラクタ名**: `NewHandler` （すべてのリソースで統一）
- **パッケージ名**: リソース名と同じ（例: `popular_work`, `password_reset`）

## 依存性注入のガイドライン

Handler 構造体には、そのリソースで必要な依存性を注入します。

### 基本方針

1. **共通の依存性**: すべてのエンドポイントで使う依存性は必ず含める
2. **専用の依存性**: 一部のエンドポイントでしか使わない依存性も含めて OK（多少の冗長性は許容）
3. **肥大化の防止**: Handler 構造体のフィールドが**8 個を超えたら**、リソース分割を検討
4. **段階的な設計**: 最初は必要最小限から始め、必要に応じて依存性を追加

### 許容される冗長性の例

```go
// ✅ 良い例: update.goでしか使わない依存性も含める（3-4個程度なら許容）
type Handler struct {
    cfg                        *config.Config
    flashMgr                   *session.FlashManager
    getPasswordEditUsecase     *usecase.GetPasswordEditUsecase
    updatePasswordResetUsecase *usecase.UpdatePasswordResetUsecase  // update.goでしか使わない
}
```

**理由**:

- 依存性が明示的で、このリソースが何に依存しているか一目でわかる
- ポインタなので未使用でもメモリオーバーヘッドは小さい（8 バイト程度）
- 将来的に他のエンドポイントでも使う可能性がある

### 肥大化の警告

```go
// ⚠️ 注意: フィールドが8個を超えたらリソース分割を検討
type Handler struct {
    cfg               *config.Config
    flashMgr          *session.FlashManager
    sidebarHelper     *sidebar.Helper
    getListUsecase    *usecase.GetSuggestionListUsecase
    getDetailUsecase  *usecase.GetSuggestionDetailUsecase
    getEditUsecase    *usecase.GetSuggestionEditUsecase
    getNewUsecase     *usecase.GetSuggestionNewUsecase
    createUsecase     *usecase.CreateSuggestionUsecase
    updateUsecase     *usecase.UpdateSuggestionUsecase
    closeUsecase      *usecase.CloseSuggestionUsecase
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

すべてのバリデーターは `internal/validator/` パッケージに配置します。形式バリデーションのみの場合も、状態バリデーション（DB を使った検証）を含む場合も同じパッケージです。

- **構造体名**: `{Resource}{Action}Validator`（例: `SignInCreateValidator`, `PasswordResetCreateValidator`）
- **コンストラクタ**: `New{Resource}{Action}Validator`
- **メソッド**: Go の慣習に従った `(data, error)` の 2 値返し。詳細は [@.claude/rules/go-validation.md](go-validation.md) を参照
- **呼び出し元は UseCase**: `main.go` で Validator を構築し、UseCase のコンストラクタに渡す。Handler は Validator に直接依存しない

### 配置例

```
internal/validator/
├── sign_in.go                  # バリデーション（形式チェック + ユーザー検索、パスワード照合）
├── sign_in_test.go             # バリデーションのテスト
├── password_reset.go           # バリデーション（形式チェックのみ）
├── password_reset_test.go      # バリデーションのテスト
├── password.go                 # パスワード変更のバリデーション
└── password_test.go            # パスワード変更のバリデーションテスト

internal/handler/
├── sign_in/
│   ├── handler.go              # UseCaseを外部から受け取る
│   ├── new.go
│   ├── create.go
│   └── handler_test.go         # ハンドラーの統合テスト
├── password_reset/
│   ├── handler.go              # UseCaseを外部から受け取る
│   ├── new.go
│   ├── create.go
│   └── handler_test.go         # ハンドラーの統合テスト
```

詳細なバリデーション実装については、[@.claude/rules/go-validation.md](go-validation.md) を参照してください。

## ルーティング登録

`cmd/server/main.go` でルーティングを登録する際は、リソースごとにハンドラーを初期化します。

### 基本的な登録方法

`main.go` で Validator → UseCase → Handler の順に構築します。

```go
// 1. Validator の構築
signInValidator := validator.NewSignInCreateValidator(userRepo, userPasswordRepo, userTwoFactorAuthRepo)

// 2. UseCase の構築（Validator を注入）
createSignInUC := usecase.NewCreateSignInUsecase(signInValidator, userSessionRepo)
getSignInNewUC := usecase.NewGetSignInNewUsecase()

// 3. Handler の構築（UseCase を注入）
signInHandler := sign_in.NewHandler(flashMgr, getSignInNewUC, createSignInUC)

// ルーティング登録
r.Get("/sign_in", signInHandler.New)
r.Post("/sign_in", signInHandler.Create)
```

### ルーティングの原則

- **標準的な HTTP メソッドを使用**: GET/POST/PATCH/DELETE
- **更新処理は PATCH を使用**: 部分更新を表現し、Rails との整合性も保つ
- **Method Override パターン**: HTML フォームから PATCH/DELETE を実現（詳細は [HTTPメソッドとルーティング](#httpメソッドとルーティング) を参照）

## 実装例

### 例 1: 単独エンドポイント（health/）

```go
// internal/handler/health/handler.go
package health

import (
    "github.com/wikinoapp/wikino/go/internal/config"
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

### 例 2: 複数エンドポイント（sign_in/）

```go
// internal/handler/sign_in/handler.go
package sign_in

import (
    "github.com/wikinoapp/wikino/go/internal/session"
    "github.com/wikinoapp/wikino/go/internal/usecase"
)

// Handler はサインイン関連のHTTPハンドラーです
type Handler struct {
    flashMgr           *session.FlashManager
    getSignInNewUC     *usecase.GetSignInNewUsecase
    createSignInUC     *usecase.CreateSignInUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
    flashMgr *session.FlashManager,
    getSignInNewUC *usecase.GetSignInNewUsecase,
    createSignInUC *usecase.CreateSignInUsecase,
) *Handler {
    return &Handler{
        flashMgr:       flashMgr,
        getSignInNewUC: getSignInNewUC,
        createSignInUC: createSignInUC,
    }
}
```

```go
// internal/handler/sign_in/new.go
package sign_in

import "net/http"

// New GET /sign_in - サインインフォーム
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
    // テンプレートをレンダリング
}
```

```go
// internal/handler/sign_in/create.go
package sign_in

import (
    "errors"
    "log/slog"
    "net/http"

    "github.com/wikinoapp/wikino/go/internal/model"
    "github.com/wikinoapp/wikino/go/internal/usecase"
)

// Create POST /sign_in - サインイン処理
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. リクエストのパース
    email := r.FormValue("email")
    password := r.FormValue("password")

    // 2. UseCase 呼び出し（認可・バリデーション・永続化はすべて UseCase 内で実行）
    output, err := h.createSignInUC.Execute(ctx, usecase.CreateSignInInput{
        Email:    email,
        Password: password,
    })
    if err != nil {
        // エラー型に応じたレスポンス
        var ve *model.ValidationError
        if errors.As(err, &ve) {
            w.WriteHeader(http.StatusUnprocessableEntity)
            h.renderForm(w, r, ve, email)
            return
        }
        slog.ErrorContext(ctx, "サインインに失敗", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 3. レスポンス
    h.flashMgr.SetSuccess(w, "サインインしました")
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

## エラーハンドリング

Handler は UseCase から返されるエラーを `errors.As` で型判別し、適切な HTTP レスポンスを返します。

### エラー型の使い分け

| エラー型                 | 生成元    | 意味                             | Handler の対応                          |
| ------------------------ | --------- | -------------------------------- | --------------------------------------- |
| `*model.ValidationError` | Validator | 入力が不正（ユーザーが修正可能） | フォーム再描画（422）                   |
| `*model.AppError`        | UseCase   | 業務レベルの既知の失敗           | エラーコードに応じた処理（403, 404 等） |
| 素の `error`             | どこでも  | 予期しないシステムエラー         | 500                                     |

### 実装パターン

```go
output, err := h.createSuggestionUC.Execute(ctx, input)
if err != nil {
    // 1. バリデーションエラー → フォーム再描画（422）
    var ve *model.ValidationError
    if errors.As(err, &ve) {
        w.WriteHeader(http.StatusUnprocessableEntity)
        h.renderNewForm(w, r, ve)
        return
    }

    // 2. アプリケーションエラー → エラーコードに応じた処理
    var ae *model.AppError
    if errors.As(err, &ae) {
        switch ae.Code {
        case model.AppErrCodeResourceNotFound:
            handler.NotFound(w, r)
        case model.AppErrCodeForbidden:
            handler.NotFound(w, r)
        default:
            slog.ErrorContext(ctx, ae.LogString())
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        }
        return
    }

    // 3. 予期しないエラー → 500
    slog.ErrorContext(ctx, "予期しないエラー", "error", err)
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    return
}
```

### 重要なルール

- **Handler はエラーの型を判別するだけ**: どのエラーを返すかは UseCase と Validator の責務
- **`*model.ValidationError`**: Validator が生成し、UseCase がそのまま返す
- **`*model.AppError`**: UseCase が業務文脈に基づいて生成する（例: リソースが見つからない、権限がない）
- **素の `error`**: ユーザーには汎用的なエラーメッセージを表示し、詳細はログに記録する

## HTTP メソッドとルーティング

HTML フォームと Web API（JSON）で同じルーティングを使用するため、**Method Override パターン**を採用します。

### 基本方針

- **Web API（JSON）**: 標準的な HTTP メソッド（GET/POST/PATCH/DELETE）を使用
- **HTML フォーム**: POST に`_method`パラメータを追加して PATCH/DELETE を実現
- **ルーティング**: HTML と API で同じ URL・HTTP メソッドを共有
- **更新処理**: PATCH を使用（部分更新を表現、Rails との整合性も保つ）

### 実装方法

**ミドルウェア**: `internal/middleware/method_override.go`

```go
r.Use(authMiddleware.MethodOverride) // POSTリクエストの_methodパラメータを読み取り、HTTPメソッドを上書き
```

**HTML フォームの例**:

```html
<!-- PATCH /password として処理される -->
<form method="POST" action="/password">
  <input type="hidden" name="_method" value="PATCH" />
  <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
</form>
```

**ルーティング**:

```go
r.Patch("/password", h.UpdatePassword)
r.Patch("/users/{id}", h.PartialUpdateUser)
r.Delete("/posts/{id}", h.DeletePost)
```

## 設計思想

この命名規則は、以下の設計思想に基づいています：

- **RESTful 設計**: Rails をはじめ多くの Web フレームワークが採用する標準的な 7 アクション（index, show, new, create, edit, update, delete）
- **完全な対応**: ファイル名とメソッド名が 100%一致
- **HTTP メソッドとの区別**: HTTP メソッド（GET/POST/PATCH/DELETE）とハンドラーメソッド名（Index/Show/Create/Update/Delete）は明確に区別
- **Rails からの移行に最適**: 既存の Rails 開発者が即座に理解できる命名
- **MPA 向け最適化**: HTML レンダリング型 Web アプリケーションに適した命名

## まとめ

HTTP ハンドラーを実装する際は、以下のポイントを守ってください：

1. **Handler は薄い Adapter**: リクエストのパース → UseCase 呼び出し → レスポンス生成のみ。認可・バリデーションは UseCase に委譲する
2. **すべてのエンドポイントをディレクトリ化**: 例外なく、リソースディレクトリを作成
3. **標準的なファイル名を使用**: 8 種類のファイル名のみを使用
   - `handler.go`, `index.go`, `show.go`, `new.go`, `create.go`, `edit.go`, `update.go`, `delete.go`
4. **ファイル名とメソッド名を一致させる**: 可読性と保守性を向上
5. **依存性注入を適切に管理**: 8 個以下のフィールドを目安に、必要に応じてリソースを分割
6. **UseCase のエラーを `errors.As` で判別**: `ValidationError` → 422、`AppError` → エラーコードに応じた処理、素の `error` → 500
7. **Handler から Policy・Validator への直接依存は禁止**: depguard で強制。すべて UseCase を経由する

これらの規則を守ることで、以下のメリットが得られます：

- コードの可読性と保守性が向上する
- 新規機能追加時に迷わない
- エントリーポイントが増えても認可・バリデーションが漏れない
- テストが書きやすくなる
