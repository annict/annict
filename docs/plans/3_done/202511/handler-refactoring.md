# ハンドラーの整理とリファクタリング 設計書

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

現在、Annict Go 版のハンドラー層は、機能ごとに異なる構成で実装されています（`handler.go` に複数の機能、`sign_in.go` に専用ハンドラー、`password_reset.go` に専用ハンドラーなど）。この状態は以下の問題を引き起こしています：

- ファイルが大きくなりすぎて可読性が低下する
- 機能追加時にどこに実装すべきか迷う
- 並行開発時にファイル衝突が発生しやすい
- テストファイルも肥大化する

**目的**:

- 一貫性のあるディレクトリ構造とファイル命名規則を確立する
- 1 ファイル 1 責務の原則を徹底し、コードの可読性・保守性を向上させる
- Go の標準的なプロジェクト構成に準拠し、新規メンバーが理解しやすくする
- 機能追加時の開発効率を向上させる

**背景**:

- 現在は機能ごとにハンドラーの構成がバラバラで、統一されたルールがない
- 今後、サインアップ機能やその他の機能を追加する際に、明確な指針が必要
- Go の中〜大規模プロジェクトで一般的な「リソースごとにディレクトリを切る」パターンを採用することで、将来的な拡張性を確保する

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

- リソース/機能ごとにディレクトリを切り、関連するハンドラーをまとめる
- 1 エンドポイント = 1 ハンドラーファイル の原則を徹底する
- ファイル名とメソッド名に一貫性のある命名規則を適用する
- 既存のハンドラーを新しい構成に移行する（動作は変更しない）
- テストファイルも同様の構成で整理する

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

#### 保守性

- ファイルサイズを小さく保ち（1 ファイル 200 行以下を目安）、可読性を維持する
- ディレクトリ構造を見れば、どこに何があるか直感的に理解できる
- テストファイルも同じ命名規則に従い、対応するハンドラーを簡単に見つけられる

#### 拡張性

- 新規機能追加時に、どのディレクトリ・ファイル名で作成すべきか明確
- Go の標準的なプロジェクト構成に準拠し、一般的な Go プロジェクトの知識が活用できる

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

### ディレクトリ構造

新しいディレクトリ構造は以下の通りです：

```
handler/
├── popular_work/
│   ├── handler.go       # Handler構造体と依存性
│   ├── index.go         # Index メソッド (GET /works/popular)
│   └── index_test.go    # Index のテスト
├── password_reset/
│   ├── handler.go       # Handler構造体と依存性
│   ├── new.go           # New (GET /password/reset) - リセット申請フォーム
│   ├── new_test.go      # New のテスト
│   ├── create.go        # Create (POST /password/reset)
│   ├── create_test.go   # Create のテスト
│   └── request.go       # Request構造体
├── password/
│   ├── handler.go       # Handler構造体と依存性
│   ├── edit.go          # Edit (GET /password/edit)
│   ├── edit_test.go     # Edit のテスト
│   ├── update.go        # Update (PUT /password)
│   ├── update_test.go   # Update のテスト
│   └── request.go       # Request構造体
├── sign_in/
│   ├── handler.go       # Handler構造体と依存性
│   ├── new.go           # New (GET /sign_in) - サインインフォーム
│   ├── new_test.go      # New のテスト
│   ├── create.go        # Create (POST /sign_in)
│   ├── create_test.go   # Create のテスト
│   └── request.go       # Request構造体
├── health/
│   ├── handler.go       # Handler構造体と依存性
│   ├── show.go          # Show (GET /health) - ヘルスチェック
│   └── show_test.go     # Show のテスト
├── home/
│   ├── handler.go       # Handler構造体と依存性
│   ├── show.go          # Show (GET /) - ホームページ
│   └── show_test.go     # Show のテスト
├── manifest/
│   ├── handler.go       # Handler構造体と依存性
│   ├── show.go          # Show (GET /manifest.json) - PWAマニフェスト
│   └── show_test.go     # Show のテスト
└── error_502/
    ├── handler.go       # Handler構造体と依存性
    ├── show.go          # Show (GET /error/502) - エラーページ
    └── show_test.go     # Show のテスト
```

**リソースディレクトリの原則**:

- **すべてのエンドポイントをディレクトリ化**: 例外なく、すべてのエンドポイントはリソースディレクトリを作成
- **統一性**: 単独のエンドポイントでも必ずディレクトリを作成（例: `health/`, `home/`, `manifest/`）
- **拡張性**: 将来的にエンドポイントが追加されても容易に対応可能
- **リソース名**: 名詞として成立するリソース名を使用（例: `popular_work/`, `password_reset/`, `sign_in/`）

**命名の原則**:

- リソース名は**名詞**にする（例: `popular_work`, `password_reset`, `sign_in`）
- 形容詞+名詞の組み合わせの場合は英語の自然な語順にする（例: `popular_work` ⭕️、`work_popular` ❌）

### ファイル命名規則

**標準ファイル名（8 種類のみ）**:

リソースディレクトリ内には、以下の標準的なファイル名**のみ**を使用します：

- `handler.go` - Handler 構造体と依存性の定義
- `index.go` - 一覧ページ表示 (GET /resources)
- `show.go` - 個別ページ表示 (GET /resources/:id)
- `new.go` - 新規作成フォーム表示 (GET /resources/new)
- `create.go` - 作成処理 (POST /resources)
- `edit.go` - 編集フォーム表示 (GET /resources/:id/edit)
- `update.go` - 更新処理 (PATCH /resources/:id)
- `delete.go` - 削除処理 (DELETE /resources/:id)

**重要な原則**:

- 上記 8 種類以外のファイル名は**使用しない**
- 複雑な名前（`show_reset_form.go`, `process_reset.go` など）が必要な場合は、**新しいリソースディレクトリを作成する**
- 例: `password/show_reset_form.go` ではなく、`password_reset/new.go` を使用

**テストファイルと Request DTO**:

- テストファイル: 対応するハンドラーファイルと同じ名前に `_test.go` を付ける（例: `index_test.go`, `show_test.go`, `new_test.go`）
- Request DTO: `request.go` と `request_test.go`（1 リソース 1 リクエスト構造体）

### メソッド命名規則

リソースディレクトリ内では以下のメソッド名を使用します：

- `Index` - 一覧ページ表示（ファイル名: `index.go`）
- `Show` - 個別ページ表示（ファイル名: `show.go`）
- `New` - 新規作成フォーム表示（ファイル名: `new.go`）
- `Create` - 作成処理（ファイル名: `create.go`）
- `Edit` - 編集フォーム表示（ファイル名: `edit.go`）
- `Update` - 更新処理（ファイル名: `update.go`）
- `Delete` - 削除処理（ファイル名: `delete.go`）

**ファイル名とメソッド名の一致**:

- すべてのファイル名とメソッド名が完全に対応しています（例: `index.go` → `Index` メソッド、`show.go` → `Show` メソッド）
- この一貫性により、コードの可読性と保守性が向上します
- HTTP メソッド（GET/POST/PUT/DELETE）とハンドラーメソッド名（Index/Show/Create/Update/Delete）は明確に区別されています

**重要**: 複雑なメソッド名（`ShowResetForm`, `ProcessReset` など）は**使用しない**。代わりに新しいリソースディレクトリを作成し、標準的なメソッド名を使用します。

**例**:

- ❌ `password.ShowResetForm()` ではなく、✅ `password_reset.New()` を使用
- ❌ `password.ProcessReset()` ではなく、✅ `password_reset.Create()` を使用

### コード設計

#### Handler 構造体の定義

各リソースディレクトリの `handler.go` に Handler 構造体と依存性を定義します。

```go
// handler/popular_work/handler.go
package popular_work

import (
    "github.com/annict/annict/internal/config"
    repository "github.com/annict/annict/internal/repository/sqlc"
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

#### 依存性注入のガイドライン

Handler 構造体には、そのリソースで必要な依存性を注入します。

**基本方針**:

1. **共通の依存性**: すべてのエンドポイントで使う依存性は必ず含める
2. **専用の依存性**: 一部のエンドポイントでしか使わない依存性も含めて OK（多少の冗長性は許容）
3. **肥大化の防止**: Handler 構造体のフィールドが**8 個を超えたら**、リソース分割を検討
4. **段階的な設計**: 最初は必要最小限から始め、必要に応じて依存性を追加

**許容される冗長性の例**:

```go
// ✅ 良い例: update.go でしか使わない依存性も含める（3-4個程度なら許容）
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
- ポインタなので未使用でもメモリオーバーヘッドは小さい（8 バイト程度）
- 将来的に他のエンドポイントでも使う可能性がある

**肥大化の警告**:

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

**リソース分割の例**:

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

#### エンドポイント実装ファイル

各エンドポイントは個別のファイルに実装します。ファイル名は標準的な 8 種類のいずれかを使用します。

```go
// handler/popular_work/index.go
package popular_work

import (
    "net/http"
)

// Index GET /works/popular - 人気作品一覧を表示
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
    // 実装
}
```

#### ルーティング登録

`cmd/server/main.go` でルーティングを登録する際は、リソースごとにハンドラーを初期化します。

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

### Request DTO の配置

Request DTO（リクエストバリデーション用の構造体）は、各リソースディレクトリに配置します。

**命名規則**:

- **1 リソース 1 リクエスト構造体**: 各リソースディレクトリには `request.go` と `request_test.go` のみを配置
- 複数のリクエスト構造体が必要な場合は、**新しいリソースディレクトリを作成する**

**例**:

```
handler/
├── sign_in/
│   ├── handler.go
│   ├── new.go
│   ├── create.go
│   ├── request.go          # Request 構造体（サインイン用）
│   └── request_test.go     # Request のテスト
├── password_reset/
│   ├── handler.go
│   ├── new.go
│   ├── create.go
│   ├── request.go          # Request 構造体（パスワードリセット用）
│   └── request_test.go     # Request のテスト
├── password/
│   ├── handler.go
│   ├── edit.go
│   ├── update.go
│   ├── request.go          # Request 構造体（パスワード更新用）
│   └── request_test.go     # Request のテスト
```

**重要**: `reset_request.go` や `update_request.go` のような複雑な名前は使用しません。常に `request.go` を使用します。

### 実装方針

#### リファクタリングの原則

- **動作を変更しない**: 既存のハンドラーの動作は一切変更せず、ファイル構成のみを変更する
- **テストを先に移行**: テストファイルを先に移行し、動作確認してからハンドラーを移行する
- **段階的に移行**: 1 リソースずつ移行し、各ステップでテストを実行して正常性を確認する
- **インポートパスの変更**: ファイル移動に伴うインポートパスの変更を漏れなく実施する

#### テスト戦略

- 既存のテストをそのまま移行（テストの内容は変更しない）
- 移行後、すべてのテストが通ることを確認する
- 各リソースの移行が完了したら、`make test` を実行して全体のテストが通ることを確認する

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

### フェーズ 0: 既存コードのPATCH統一

<!--
既存のPUTを使用している箇所をPATCHに統一するフェーズ。
本番運用前のため、今のうちにリファクタリングを実施する。
-->

- [x] **0-1**: 既存のPUTエンドポイントをPATCHに変更
  - `cmd/server/main.go` のルーティング定義を `r.Put("/password", ...)` から `r.Patch("/password", ...)` に変更
  - `internal/handler/password_reset.go` のコメントを `PUT /password` から `PATCH /password` に変更
  - `internal/handler/password_reset_integration_test.go` のテストコードを `PUT` から `PATCH` に変更
  - `internal/handler/password_reset_token_lifecycle_test.go` のテストコードを `PUT` から `PATCH` に変更
  - `internal/templates/pages/password/edit.templ` のフォームを `value="PUT"` から `value="PATCH"` に変更
  - `internal/templates/pages/password/edit_templ.go` を再生成（`make templ-generate`）
  - `go/CLAUDE.md` のドキュメント例を `r.Put` から `r.Patch` に更新
  - テストを実行して正常に動作することを確認
  - **想定ファイル数**: 約 7 ファイル（実装 5 + テスト 2）
  - **想定行数**: 約 50 行（コメントとメソッド名の変更のみ）

### フェーズ 1: ドキュメント整備

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: ハンドラー命名規則のドキュメント作成
  - `.claude/designs/1_doing/handler-refactoring.md` の作成（完了）
  - PATCHで統一する方針を明記（完了）
  - **想定ファイル数**: 約 1 ファイル（ドキュメントのみ）
  - **想定行数**: 約 450 行（ドキュメントのみ）

### フェーズ 2: 単独エンドポイントの移行

- [x] **2-1**: `health/` と `home/` ディレクトリの作成
  - `handler/health/` ディレクトリを作成
  - `handler/health/handler.go` を作成（Handler 構造体）
  - `handler/health/show.go` を作成（Show メソッド）
  - `handler/health/show_test.go` を作成
  - `handler/home/` ディレクトリを作成
  - `handler/home/handler.go` を作成（Handler 構造体）
  - `handler/home/show.go` を作成（Show メソッド）
  - `handler/home/show_test.go` を作成
  - `cmd/server/main.go` のルーティングを更新（`healthHandler.Show`, `homeHandler.Show`）
  - `handler.go` から Health と Home 関連を削除
  - **想定ファイル数**: 約 10 ファイル（実装 6 + テスト 2 + main.go 1 + handler.go 1）
  - **想定行数**: 約 200 行（実装 120 行 + テスト 80 行）

### フェーズ 3: popular_work リソースの移行

- [x] **3-1**: `popular_work/` ディレクトリの作成と `PopularWorks` の移行
  - `handler/popular_work/` ディレクトリを作成
  - `handler/popular_work/handler.go` を作成（Handler 構造体）
  - `handler/popular_work/index.go` を作成（Index メソッド）
  - `handler/popular_work/index_test.go` を作成（既存のテストを移行）
  - `cmd/server/main.go` のルーティングを更新（`popularWorkHandler.Index`）
  - `handler.go` から PopularWorks 関連を削除
  - **想定ファイル数**: 約 5 ファイル（実装 3 + テスト 1 + main.go 1）
  - **想定行数**: 約 200 行（実装 120 行 + テスト 80 行）

### フェーズ 4: sign_in リソースの移行

- [x] **4-1**: `sign_in/` ディレクトリへの移行
  - `handler/sign_in/` ディレクトリを作成
  - `handler/sign_in/handler.go` を作成（既存の SignInHandler を移動）
  - `handler/sign_in/new.go` を作成（ShowSignIn → New に改名）
  - `handler/sign_in/new_test.go` を作成
  - `handler/sign_in/create.go` を作成（ProcessSignIn → Create に改名）
  - `handler/sign_in/create_test.go` を作成
  - `handler/sign_in/request.go` を作成（SignInRequest → Request に改名、移動）
  - `handler/sign_in/request_test.go` を作成
  - `cmd/server/main.go` のルーティングを更新（`signInHandler.New`, `signInHandler.Create`）
  - `sign_in.go` と `sign_in_request.go` を削除
  - **想定ファイル数**: 約 10 ファイル（実装 4 + テスト 4 + main.go 1 + 削除 2）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

### フェーズ 5: password_reset と password リソースの移行

- [x] **5-1**: `password_reset/` ディレクトリへの移行
  - `handler/password_reset/` ディレクトリを作成
  - `handler/password_reset/handler.go` を作成（Handler 構造体）
  - `handler/password_reset/new.go` を作成（ShowResetForm → New に改名）
  - `handler/password_reset/new_test.go` を作成
  - `handler/password_reset/create.go` を作成（ProcessReset → Create に改名）
  - `handler/password_reset/create_test.go` を作成
  - `handler/password_reset/request.go` を作成（Request 構造体）
  - `handler/password_reset/request_test.go` を作成
  - `cmd/server/main.go` のルーティングを更新（`passwordResetHandler.New`, `passwordResetHandler.Create`）
  - **想定ファイル数**: 約 9 ファイル（実装 4 + テスト 4 + main.go 1）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [x] **5-2**: `password/` ディレクトリへの移行
  - `handler/password/` ディレクトリを作成
  - `handler/password/handler.go` を作成（Handler 構造体）
  - `handler/password/edit.go` を作成（ShowEditForm → Edit に改名）
  - `handler/password/edit_test.go` を作成
  - `handler/password/update.go` を作成（UpdatePassword → Update に改名）
  - `handler/password/update_test.go` を作成
  - `handler/password/request.go` を作成（Request 構造体）
  - `handler/password/request_test.go` を作成
  - `cmd/server/main.go` のルーティングを更新（`passwordHandler.Edit`, `passwordHandler.Update`）
  - `password_reset.go` と `password_reset_request.go` を削除
  - **想定ファイル数**: 約 10 ファイル（実装 4 + テスト 4 + main.go 1 + 削除 2）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

### フェーズ 6: 残りのファイルの整理

- [x] **6-1**: `handler.go` の削除と最終確認
  - `handler.go` の削除（すべてのメソッドが移行済み）
  - `handler_test.go` の削除または必要に応じて分割
  - `error_502/` ディレクトリの作成と移行
  - `manifest/` ディレクトリの作成と移行
  - すべてのテストが通ることを確認
  - **想定ファイル数**: 約 2 ファイル（削除のみ）
  - **想定行数**: 約 0 行（削除のみ）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **新規ハンドラーの追加**: このリファクタリングは既存コードの整理のみを行い、新規機能は追加しない
- **ハンドラーのロジック変更**: 既存のビジネスロジックは一切変更せず、ファイル構成のみを変更する
- **テストの追加**: 既存のテストを移行するのみで、新規テストは追加しない（カバレッジの向上は別タスク）

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Standard Go Project Layout - Handler Organization](https://github.com/golang-standards/project-layout/blob/master/internal/README.md)
- [Effective Go - Package names](https://go.dev/doc/effective_go#package-names)
- [CLAUDE.md - Pull Request のガイドライン](../../CLAUDE.md#pull-requestのガイドライン)

---

## 補足: 命名規則の例

### リソースディレクトリとエンドポイントの対応

| エンドポイント       | リソースディレクトリ | ファイル名  | メソッド名 | 説明                           |
| -------------------- | -------------------- | ----------- | ---------- | ------------------------------ |
| GET /works           | `work/`              | `index.go`  | `Index`    | 一覧表示                       |
| GET /works/:id       | `work/`              | `show.go`   | `Show`     | 個別表示                       |
| GET /works/new       | `work/`              | `new.go`    | `New`      | 新規作成フォーム               |
| POST /works          | `work/`              | `create.go` | `Create`   | 作成処理                       |
| GET /works/:id/edit  | `work/`              | `edit.go`   | `Edit`     | 編集フォーム                   |
| PATCH /works/:id     | `work/`              | `update.go` | `Update`   | 更新処理                       |
| DELETE /works/:id    | `work/`              | `delete.go` | `Delete`   | 削除処理                       |
| GET /works/popular   | `popular_work/`      | `index.go`  | `Index`    | 人気作品一覧                   |
| GET /password/reset  | `password_reset/`    | `new.go`    | `New`      | パスワードリセット申請フォーム |
| POST /password/reset | `password_reset/`    | `create.go` | `Create`   | パスワードリセット申請処理     |
| GET /password/edit   | `password/`          | `edit.go`   | `Edit`     | パスワード変更フォーム         |
| PATCH /password      | `password/`          | `update.go` | `Update`   | パスワード変更処理             |
| GET /sign_in         | `sign_in/`           | `new.go`    | `New`      | サインインフォーム             |
| POST /sign_in        | `sign_in/`           | `create.go` | `Create`   | サインイン処理                 |

### 単独エンドポイント（1 つのアクションのみ持つリソース）

| エンドポイント     | リソースディレクトリ | ファイル名 | メソッド名 | 説明             |
| ------------------ | -------------------- | ---------- | ---------- | ---------------- |
| GET /              | `home/`              | `show.go`  | `Show`     | ホームページ     |
| GET /health        | `health/`            | `show.go`  | `Show`     | ヘルスチェック   |
| GET /manifest.json | `manifest/`          | `show.go`  | `Show`     | PWA マニフェスト |
| GET /error/502     | `error_502/`         | `show.go`  | `Show`     | エラーページ     |

### ファイル名の決め方（フローチャート）

1. **リソース名を決める**
   - エンドポイントの URL（例: `/works/popular`）から名詞を抽出
   - リソースディレクトリ名を決定（例: `popular_work/`）
   - 単独エンドポイントでもディレクトリを作成（例: `/health` → `health/`）

2. **標準的な CRUD アクション（index/show/new/create/edit/update/delete）を選択**
   - 一覧表示 → `index.go`
   - 個別表示 → `show.go`
   - 新規フォーム → `new.go`
   - 作成処理 → `create.go`
   - 編集フォーム → `edit.go`
   - 更新処理 → `update.go`
   - 削除処理 → `delete.go`

### ファイル名とメソッド名の対応

RESTful 設計の標準に沿った、以下の対応関係を採用しています：

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

**設計思想**:

- **RESTful 設計**: Rails をはじめ多くの Web フレームワークが採用する標準的な 7 アクション（index, show, new, create, edit, update, delete）
- **完全な対応**: ファイル名とメソッド名が 100%一致
- **HTTP メソッドとの区別**: HTTP メソッド（GET/POST/PUT/DELETE）とハンドラーメソッド名（Index/Show/Create/Update/Delete）は明確に区別
- **Rails からの移行に最適**: 既存の Rails 開発者が即座に理解できる命名
- **MPA 向け最適化**: HTML レンダリング型 Web アプリケーションに適した命名

### リソース名の命名例

| URL パターン    | リソース名（推奨）   | 理由                         |
| --------------- | -------------------- | ---------------------------- |
| /works/popular  | `popular_work/` ⭕️   | 形容詞+名詞の自然な語順      |
| /works/popular  | `work_popular/` ❌   | 「作品\_人気」は不自然       |
| /password/reset | `password_reset/` ⭕️ | 名詞として成立               |
| /users/me       | `current_user/` ⭕️   | 「現在のユーザー」という名詞 |
| /search         | `search/` ⭕️         | 「検索」という名詞           |
