# session.Managerリファクタリング 設計書

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

現在の`internal/session/session.go`の`Manager`構造体が`*query.Queries`に直接依存しており、アーキテクチャガイドラインに違反しています。アーキテクチャガイドラインによれば、**Presentation層のヘルパー（`internal/session`）がDomain/Infrastructure層のQuery（`query.Queries`）に直接依存してはいけません**。すべてのデータアクセスはRepositoryを経由すべきです。

**目的**:

- アーキテクチャガイドラインに準拠したコード構造の実現
- データアクセスロジックのSessionRepositoryへの集約
- テスタビリティとメンテナンス性の向上

**背景**:

- 現在、`session.Manager`が`query.Queries`に直接依存している（アーキテクチャ違反）
- SessionRepositoryは`TouchSession()`メソッドのみを提供しており、他のセッション関連操作が含まれていない
- データアクセスロジックが散在している（SessionRepositoryとsession.Managerの両方に存在）
- 今回のセッションタイムアウト問題の修正（#170-173）で、この問題が顕在化した

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

- SessionRepositoryがすべてのセッション関連のデータアクセスを提供する
- session.Managerは`query.Queries`に直接依存せず、SessionRepositoryのみに依存する
- 既存のセッション管理機能（取得、作成、更新、削除）の動作は完全に保たれる
- リファクタリング前後でAPIの変更がない（内部実装のみの変更）

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

**保守性**:

- データアクセスロジックがSessionRepositoryに集約され、変更が容易になる
- アーキテクチャガイドラインに準拠し、将来のメンテナンスが容易になる
- テストが書きやすくなる（SessionRepositoryをモックすればsession.Managerのテストが容易）

**互換性**:

- 既存のセッション管理機能の動作を完全に保つ
- 既存のAPI（session.Managerの公開メソッド）を変更しない
- 既存のテストがリグレッションを検出できる

**パフォーマンス**:

- リファクタリング前後でパフォーマンスの変化がない（単なる依存先の変更）

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

### 現在の問題点

**session.Managerの依存関係（アーキテクチャ違反）**:

```go
// internal/session/session.go
type Manager struct {
    repo       *query.Queries  // ← Presentation層がQueryに直接依存（アーキテクチャ違反）
    cfg        *config.Config
    CookieName string
}
```

**アーキテクチャガイドラインの要求**:

- ✅ **正**: Presentation層 → Repository → Query
- ❌ **誤**: Presentation層 → Query（現在の実装）

### 実装方針

#### 1. SessionRepositoryの拡張

現在のSessionRepositoryには`TouchSession()`しかないため、session.Managerが必要とする以下のメソッドを追加します：

- `GetSessionByID(ctx context.Context, sessionID string) (*query.Session, error)`
- `GetUserByID(ctx context.Context, userID int64) (*query.GetUserByIDRow, error)`
- `UpdateSession(ctx context.Context, params query.UpdateSessionParams) error`
- `CreateSession(ctx context.Context, params query.CreateSessionParams) (query.Session, error)`

**実装例**:

```go
// internal/repository/session.go
package repository

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"

    "github.com/annict/annict/internal/query"
)

// SessionRepository はSession関連のデータアクセスを担当します
type SessionRepository struct {
    queries *query.Queries
}

// NewSessionRepository はSessionRepositoryを作成します
func NewSessionRepository(queries *query.Queries) *SessionRepository {
    return &SessionRepository{queries: queries}
}

// TouchSession はセッションのupdated_atを更新します
func (r *SessionRepository) TouchSession(ctx context.Context, sessionID string) error {
    privateID := r.generatePrivateID(sessionID)
    return r.queries.TouchSession(ctx, privateID)
}

// GetSessionByID はセッションIDからセッションを取得します
func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*query.Session, error) {
    privateID := r.generatePrivateID(sessionID)
    session, err := r.queries.GetSessionByID(ctx, privateID)
    if err != nil {
        return nil, err
    }
    return &session, nil
}

// GetUserByID はユーザーIDからユーザー情報を取得します
func (r *SessionRepository) GetUserByID(ctx context.Context, userID int64) (*query.GetUserByIDRow, error) {
    user, err := r.queries.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// UpdateSession はセッションを更新します
func (r *SessionRepository) UpdateSession(ctx context.Context, params query.UpdateSessionParams) error {
    return r.queries.UpdateSession(ctx, params)
}

// CreateSession はセッションを作成します
func (r *SessionRepository) CreateSession(ctx context.Context, params query.CreateSessionParams) (query.Session, error) {
    return r.queries.CreateSession(ctx, params)
}

// generatePrivateID はpublic IDからprivate IDを生成
// Rails/Rackの実装と互換性のある形式: "2::" + SHA256(publicID)
func (r *SessionRepository) generatePrivateID(publicID string) string {
    hash := sha256.Sum256([]byte(publicID))
    return fmt.Sprintf("2::%s", hex.EncodeToString(hash[:]))
}
```

#### 2. session.Managerの修正

依存先を`query.Queries`から`repository.SessionRepository`に変更します：

**変更前**:

```go
type Manager struct {
    repo       *query.Queries
    cfg        *config.Config
    CookieName string
}

func NewManager(repo *query.Queries, cfg *config.Config) *Manager {
    return &Manager{
        repo:       repo,
        cfg:        cfg,
        CookieName: SessionKey,
    }
}
```

**変更後**:

```go
type Manager struct {
    sessionRepo *repository.SessionRepository  // ← repoから変更
    cfg         *config.Config
    CookieName  string
}

func NewManager(sessionRepo *repository.SessionRepository, cfg *config.Config) *Manager {
    return &Manager{
        sessionRepo: sessionRepo,
        cfg:         cfg,
        CookieName:  SessionKey,
    }
}
```

**各メソッドの修正**:

- `GetSession()` (line 70): `m.repo.GetSessionByID` → `m.sessionRepo.GetSessionByID`
- `GetCurrentUser()` (line 132): `m.repo.GetUserByID` → `m.sessionRepo.GetUserByID`
- `SetValue()` (line 158, 210, 217): `m.repo.*` → `m.sessionRepo.*`
- `GetValue()` (line 240): `m.repo.GetSessionByID` → `m.sessionRepo.GetSessionByID`
- `DeleteValue()` (line 279, 303): `m.repo.*` → `m.sessionRepo.*`
- `getAndDeleteSessionValue()` (line 325, 360): `m.repo.*` → `m.sessionRepo.*`

**`generatePrivateID()`の移動**:

- `internal/session/session.go`の`generatePrivateID()`はSessionRepositoryに移動済み
- session.Manager内では`generatePrivateID()`を直接呼び出さず、SessionRepositoryのメソッドを使用

#### 3. 初期化コードの修正

`cmd/server/main.go`でSessionRepositoryを作成してManagerに渡すように変更します：

**変更前**:

```go
sessionManager := session.NewManager(queries, cfg)
```

**変更後**:

```go
// SessionRepositoryを作成
sessionRepo := repository.NewSessionRepository(queries)

// session.Managerを作成（queriesの代わりにsessionRepoを渡す）
sessionManager := session.NewManager(sessionRepo, cfg)
```

#### 4. 影響範囲

**修正が必要なファイル**:

- `internal/repository/session.go` - SessionRepositoryの拡張
- `internal/repository/session_test.go` - SessionRepositoryのテスト
- `internal/session/session.go` - Managerの修正
- `internal/session/session_test.go` - Managerのテスト（モックの修正）
- `cmd/server/main.go` - 初期化コードの修正

**修正が不要なファイル**:

- `internal/middleware/auth.go` - Managerのインターフェースは変更されないため修正不要
- その他のハンドラー - Managerのインターフェースは変更されないため修正不要

### テスト戦略

**単体テスト**:

- `internal/repository/session_test.go`:
  - 新しく追加されたメソッド（`GetSessionByID`, `GetUserByID`, `UpdateSession`, `CreateSession`）のテスト
  - `generatePrivateID()`のテスト
- `internal/session/session_test.go`:
  - SessionRepositoryをモックしてManagerの各メソッドをテスト
  - 既存のテストを修正（モックの対象が`query.Queries`から`repository.SessionRepository`に変更）

**統合テスト**:

- 既存のテストがリグレッションを検出できることを確認
- セッション管理機能（取得、作成、更新、削除）が正常に動作することを確認

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

### フェーズ 1: SessionRepositoryの拡張

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: SessionRepositoryのメソッド追加
  - `GetSessionByID()`, `GetUserByID()`, `UpdateSession()`, `CreateSession()`メソッドを実装
  - `generatePrivateID()`メソッドを追加（`internal/session/session.go`から移動）
  - 単体テストを追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 70 行 + テスト 80 行）

### フェーズ 2: session.Managerのリファクタリング

- [x] **2-1**: session.Managerの依存関係変更
  - `Manager`構造体の`repo`フィールドを`sessionRepo`に変更
  - `NewManager()`の引数を`*query.Queries`から`*repository.SessionRepository`に変更
  - 各メソッドで`m.repo.*`を`m.sessionRepo.*`に変更
  - `generatePrivateID()`の呼び出しを削除（SessionRepositoryのメソッドを使用）
  - 単体テストを修正（モックの対象を変更）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 40 行 + テスト 60 行）

### フェーズ 3: 初期化コードの修正

- [x] **3-1**: cmd/server/main.goの修正
  - SessionRepositoryを作成してManagerに渡す
  - 既存のテストを実行して動作確認
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 10 行（実装 10 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **SessionRepositoryのインターフェース化**: 現時点では具象型で十分であり、過度な抽象化を避ける。将来的にモックが必要になった場合に検討する
- **session.Managerのインターフェース化**: 現時点では具象型で十分であり、過度な抽象化を避ける
- **セッション管理の機能追加**: 既存の機能を保つリファクタリングのみを実施し、新機能の追加は行わない

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [@go/docs/architecture-guide.md](../../go/docs/architecture-guide.md) - アーキテクチャガイド
- [セッションタイムアウト問題の修正 設計書](../1_doing/session-timeout-fix.md) - 関連する修正内容

---

## 補足情報

### リスク評価

**難易度**: ★★★☆☆（中程度）
**工数**: 約2-3時間
**リスク**: ★★☆☆☆（低-中）
**優先度**: 低（今すぐやる必要はないが、将来的にはやるべき）

**低リスクの理由**：

- ✅ ロジックの変更がない（単なる依存先の変更）
- ✅ 既存のテストがリグレッションを検出できる
- ✅ SessionRepositoryがQueryへの依存を一箇所に集約

**注意が必要な理由**：

- ⚠️ セッション管理は重要な機能（ログイン・認証に影響）
- ⚠️ すべてのセッション関連機能に影響するため、慎重なテストが必要

### メリット

- **アーキテクチャの整合性**: アーキテクチャガイドラインに準拠したコード構造
- **保守性の向上**: データアクセスロジックがSessionRepositoryに集約される
- **テスタビリティの向上**: SessionRepositoryをモックすればsession.Managerのテストが容易
- **一貫性**: 他のコードと同じパターンを使用（例: WorkRepository, UserRepositoryなど）

### デメリット

- **初期コスト**: リファクタリングに時間がかかる（2-3時間程度）
- **影響範囲**: セッション管理という重要な機能への変更

### 実施タイミング

このリファクタリングは、セッションタイムアウト問題の修正（#170-173）とは独立しているため、以下のタイミングで実施することを推奨します：

- セッションタイムアウト問題の修正が完了し、本番環境で安定稼働している
- 他の緊急度の高いタスクがない
- アーキテクチャの整合性を高めるタイミング
