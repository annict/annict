# セッションタイムアウト問題の修正 設計書

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

ユーザーから「ログインしてもすぐログアウトされる」という報告が寄せられています。調査の結果、**Go版とRails版のセッション管理の動作の違い**が原因であることが判明しました。

Rails版ではActiveRecord SessionStoreがアクセスのたびにセッションの`updated_at`を自動更新しますが、Go版では`GetSession()`でセッションを読み取るだけで`updated_at`を更新しません。その結果、Go版のページだけを使い続けるユーザーのセッションが10日後に削除され、突然ログアウトされる問題が発生しています。

**目的**:

- Go版とRails版でセッション管理の動作を統一し、ユーザーが突然ログアウトされる問題を解決する
- セッション期限の設定（30日）と実際の削除タイミング（10日）の不整合を解消する
- Rails版との完全な互換性を保つ

**背景**:

- 毎日19:00に`rake session:sweep`タスクが実行され、`updated_at`が10日以上前のセッションを削除している
- Rails版では各リクエストで`updated_at`が自動更新されるため問題は発生しない
- Go版では`GetSession()`時に`updated_at`を更新していないため、Go版のページだけを使い続けるユーザーが10日後にログアウトされる
- セッション設定では`expire_after: 30.days`だが、実際には10日で削除される不整合がある
- Annictの使用パターン（あとでまとめて記録するユーザーが多い）を考慮し、**30日**を維持する

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

- Go版でセッションを読み取る際に、`sessions`テーブルの`updated_at`カラムを自動的に更新する
- Rails版と同様に、ユーザーがアクセスするたびにセッションの有効期限が延長される
- セッションクリーンアップタスク（`rake session:sweep`）によってアクティブなセッションが削除されないようにする
- Rails版とGo版でCookieのドメイン設定を統一する

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

**パフォーマンス**:

- セッションの`updated_at`更新は各リクエストで1回のみ実行する
- 更新処理は非同期で実行し、レスポンス時間への影響を最小化する（オプション）
- データベースへの負荷増加を最小限に抑える

**保守性**:

- Rails版のActiveRecord SessionStoreと同じ動作を実装し、将来のメンテナンスを容易にする
- セッション管理のロジックを1箇所に集約し、変更を容易にする

**互換性**:

- Rails版とGo版でセッションを完全に共有できる状態を維持する
- 既存のセッションデータを壊さない

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

**Rails版の動作**:

- `config/initializers/session_store.rb`:
  ```ruby
  Annict::Application.config.session_store :active_record_store,
    key: "_annict_session_v201904",
    domain: :all,
    expire_after: 30.days
  ```
- ActiveRecord SessionStoreが**各リクエストで`updated_at`を自動更新**
- セッションクリーンアップ: 毎日19:00に`rake session:sweep`が実行され、`updated_at < 10.days.ago`のセッションを削除

**Go版の動作**:

- `go/internal/session/session.go`の`GetSession()`メソッド:
  - セッションを読み取るだけで`updated_at`を**更新しない**
  - `SetValue()`などの書き込み時のみ`updated_at`を更新
- Cookieの`domain`属性を設定していない（デフォルトは現在のドメインのみ）

**問題のシナリオ**:

1. ユーザーがRails版でログイン → セッション作成（`updated_at`が現在時刻）
2. ユーザーがGo版のページ（`/sign_in`, `/password/*`など）にアクセス → セッションを読み取るが、`updated_at`は更新されない
3. ユーザーが10日以上Go版のページだけを使い続ける → `updated_at`は10日以上前のまま
4. 毎日19:00に`rake session:sweep`が実行される → セッションが削除される
5. ユーザーが次にアクセス → セッションが存在せず、突然ログアウトされる

### 実装方針

#### 1. Go版のセッション更新機能

**方針**: Session Repositoryを作成し、認証ミドルウェアでセッションの`updated_at`を更新する

- `internal/repository/session.go`にSessionRepositoryを作成
- SessionRepositoryが`TouchSession()`メソッドを提供（Queryへの依存を集約）
- `internal/middleware/auth.go`でSessionRepositoryを使用してセッション更新
- 認証が成功した場合（ユーザーがログインしている場合）、セッションの`updated_at`を更新
- エラー時はログに記録するだけで処理を継続（パフォーマンスへの影響を最小化）

**SQLクエリ**:

```sql
-- name: TouchSession :exec
UPDATE sessions
SET updated_at = NOW()
WHERE session_id = $1;
```

**SessionRepository実装例**:

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
	// Rails/Rackの実装と互換性のある形式でprivate IDを生成
	privateID := r.generatePrivateID(sessionID)

	// セッションのupdated_atのみを更新
	return r.queries.TouchSession(ctx, privateID)
}

// generatePrivateID はpublic IDからprivate IDを生成
// Rails/Rackの実装と互換性のある形式: "2::" + SHA256(publicID)
func (r *SessionRepository) generatePrivateID(publicID string) string {
	hash := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("2::%s", hex.EncodeToString(hash[:]))
}
```

**Middleware実装例**:

```go
// internal/middleware/auth.go

// AuthMiddleware は認証ミドルウェア
type AuthMiddleware struct {
	sessionRepo    *repository.SessionRepository
	sessionManager *session.Manager
	// ...その他のフィールド
}

// RequireAuthミドルウェア内で呼び出す
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// セッションIDを取得
		sessionID, err := m.sessionManager.GetSessionID(r)
		if err != nil || sessionID == "" {
			// 未ログイン
			http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
			return
		}

		// ユーザー情報を取得
		user, err := m.sessionManager.GetCurrentUser(ctx, r)
		if err != nil || user == nil {
			// 未ログイン
			http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
			return
		}

		// セッションのupdated_atを更新
		// エラーが発生してもログに記録するだけで処理を継続
		if err := m.sessionRepo.TouchSession(ctx, sessionID); err != nil {
			slog.WarnContext(ctx, "セッション更新エラー", "error", err)
		}

		// 次のハンドラーを呼び出す
		next.ServeHTTP(w, r)
	})
}
```

#### 2. Cookieドメイン設定の統一

**問題点**:

- Rails版: `domain: :all`を使用（Rails 7.0以降では非推奨）
- Go版: `domain`属性を設定していない

**方針**:

- Rails版の`domain: :all`を明示的なドメイン名に変更（例: `ENV.fetch("ANNICT_DOMAIN")`）
- Go版のCookie設定にも同じドメインを設定
- 環境変数`ANNICT_COOKIE_DOMAIN`を新設し、両方で共有

**実装例**:

Rails版（`config/initializers/session_store.rb`）:

```ruby
Annict::Application.config.session_store :active_record_store,
  key: "_annict_session_v201904",
  domain: ENV.fetch("ANNICT_COOKIE_DOMAIN", ".annict.com"),
  expire_after: 30.days  # 既存のまま維持
```

Go版（`internal/session/session.go`）:

```go
// SetValue セッションに任意の値を保存
func (m *Manager) SetValue(ctx context.Context, w http.ResponseWriter, r *http.Request, key, value string) error {
    // ... 既存のコード ...

    // Cookieを設定
    cookie := &http.Cookie{
        Name:     SessionKey,
        Value:    publicID,
        Path:     "/",
        Domain:   m.cfg.CookieDomain, // 環境変数から取得
        Secure:   true,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    }

    // 開発環境ではSecureフラグをオフにする（HTTPアクセスを許可）
    if m.cfg.Env == "development" {
        cookie.Secure = false
    }

    http.SetCookie(w, cookie)

    // ... 既存のコード ...
}
```

#### 3. セッションクリーンアップタスクの変更

**現在の問題**:

- セッション設定は`expire_after: 30.days`だが、クリーンアップタスクは10日で削除
- この不整合が今回の問題の直接的な原因

**方針**:

- `lib/tasks/session.rake`の削除期間を`expire_after`設定と統一する
- セッション期限の妥当性を検討する（後述の「セッション期限の検討」を参照）

**実装**:

Rails版のセッションクリーンアップタスク:

```ruby
# lib/tasks/session.rake
namespace :session do
  task sweep: :environment do
    # 30日以上前のセッションを削除（10日から変更）
    Session.where("updated_at < '#{30.days.ago.to_s(:db)}'").delete_all
  end
end
```

Rails版のセッションストア設定（変更なし）:

```ruby
# config/initializers/session_store.rb
Annict::Application.config.session_store :active_record_store,
  key: "_annict_session_v201904",
  domain: ENV.fetch("ANNICT_COOKIE_DOMAIN", ".annict.com"),
  expire_after: 30.days  # 既存のまま維持
```

### セッション期限の検討

**現在の設定**:

- `expire_after: 30.days`（Rails設定）
- `updated_at < 10.days.ago`で削除（実際の動作）→ **不整合**

**業界標準の比較**:

| サービス | セッション期限 | 備考                         |
| -------- | -------------- | ---------------------------- |
| Gmail    | 14日           | 非アクティブで自動ログアウト |
| GitHub   | 14日           | 非アクティブで自動ログアウト |
| Facebook | 30日程度       | 長期間維持                   |
| Twitter  | 無期限         | セキュリティイベントで無効化 |
| Amazon   | 短い           | 頻繁にログアウトされる       |

**Annictの使用パターン分析**:

- **アクセス頻度**: アニメは基本的に週1回放送されるため、アクティブなユーザーは**週1回程度**アクセスすると想定
- **シーズン性**: 一部のユーザーはシーズンごと（3ヶ月ごと）にアクセスする可能性もある
- **記録サービス**: リアルタイムでの利用が必須ではなく、後からまとめて記録することもある

**セキュリティリスク**:

- **セッション乗っ取り**: 期限が長いほど、盗まれたセッションが長期間有効になるリスクが高まる
- **共有端末**: ネットカフェや図書館などの共有端末でログアウトし忘れた場合、期限が長いほど危険
- **個人情報**: Annictには視聴履歴などの個人情報が含まれるため、適切なセッション管理が必要

**推奨される期限**:

#### 提案1: 14日（2週間）【推奨】

**メリット**:

- セキュリティとユーザビリティのバランスが良い
- 週1回程度のアクセスでログイン状態を維持できる
- 業界標準（Gmail、GitHub）と同等
- セッション乗っ取りのリスクを低減

**デメリット**:

- 2週間以上アクセスしないユーザーは再ログインが必要
- 現在の30日から短くなるため、ユーザー体験の変化が発生

**実装**:

```ruby
expire_after: 14.days
Session.where("updated_at < '#{14.days.ago.to_s(:db)}'").delete_all
```

#### 提案2: 30日（1ヶ月）

**メリット**:

- 現在の設定を維持できるため、ユーザー体験の変化がない
- 月1回程度のアクセスでログイン状態を維持できる
- ユーザビリティ重視

**デメリット**:

- セッション乗っ取りのリスクがやや高い
- 業界標準よりも長い

**実装**:

```ruby
expire_after: 30.days  # 既存のまま
Session.where("updated_at < '#{30.days.ago.to_s(:db)}'").delete_all  # 10日から変更
```

#### 提案3: 7日（1週間）

**メリット**:

- セキュリティ重視
- 週1回のアクセスでログイン状態を維持できる

**デメリット**:

- 週1回のアクセスが必須となり、少し厳しい
- ユーザビリティがやや低下

**最終決定: 30日（1ヶ月）**

以下の理由により、セッション期限を**30日（1ヶ月）**とすることに決定しました：

1. **Annictの使用パターンに適合**: あとでまとめて記録するユーザーが多く、月1回程度のアクセスでもログイン状態を維持できることが重要
2. **既存ユーザー体験の維持**: 現在の30日設定を維持することで、ユーザー体験の変化を避ける
3. **ユーザビリティ重視**: 頻繁な再ログインを避け、ユーザーの利便性を優先
4. **設定との整合性**: 既存の`expire_after: 30.days`設定と一致させる

**考慮した点**:

- **セキュリティリスク**: 30日は業界標準（14日）よりも長いが、Annictの使用パターン（記録サービス）を考慮すると許容範囲
- **14日との比較**: セキュリティは向上するが、まとめて記録するユーザーが再ログインを強いられる可能性が高い
- **実際の問題**: 今回の問題は「30日が長すぎる」ではなく「10日と30日の不整合」であるため、30日に統一すれば解決する

**既存ユーザーへの影響**:

- セッション期限は30日のまま維持されるため、既存ユーザーへの影響はなし
- 今回の修正により、Go版のページを使い続けても突然ログアウトされる問題が解消される

### テスト戦略

**単体テスト**:

- `internal/session/session_test.go`:
  - `TouchSession()`メソッドのテスト
  - セッションの`updated_at`が正しく更新されることを確認

**統合テスト**:

- `internal/middleware/auth_test.go`:
  - 認証ミドルウェアでセッションが更新されることを確認
  - エラー時も処理が継続されることを確認

**手動テスト**:

- 開発環境でGo版のページにアクセスし、セッションの`updated_at`が更新されることを確認
- 10日以上Go版のページだけを使い続けても、セッションが削除されないことを確認

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

### フェーズ 1: Go版のセッション更新機能の実装

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: SQLクエリの実装
  - `internal/query/queries/sessions.sql`に`TouchSession`クエリを追加
  - `sqlc generate`を実行して、Goコードを生成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + 自動生成 1）
  - **想定行数**: 約 10 行（実装 5 行 + 自動生成 5 行）

- [x] **1-2**: Session Repository の実装
  - `internal/repository/session.go`を作成
  - `SessionRepository`構造体と`NewSessionRepository()`コンストラクタを実装
  - `TouchSession()`メソッドを実装（`generatePrivateID()`を内部メソッドとして含む）
  - 単体テストを追加（`internal/repository/session_test.go`）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 120 行（実装 40 行 + テスト 80 行）

### フェーズ 2: 認証ミドルウェアの更新

- [x] **2-1**: 認証ミドルウェアでセッション更新を実装
  - `internal/middleware/auth.go`の`AuthMiddleware`構造体に`SessionRepository`フィールドを追加
  - `RequireAuth`ミドルウェア内でSessionRepositoryの`TouchSession()`を呼び出す
  - エラー時はログに記録するだけで処理を継続（`slog.WarnContext`を使用）
  - 統合テストを追加（`internal/middleware/auth_test.go`）
  - `cmd/server/main.go`でSessionRepositoryを初期化してAuthMiddlewareに渡す
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 150 行（実装 60 行 + テスト 90 行）

### フェーズ 3: Cookieドメイン設定の統一

- [x] **3-1**: Go版のCookieドメイン設定を追加
  - `internal/config/config.go`に`CookieDomain`フィールドを追加
  - `.env.example`に`ANNICT_COOKIE_DOMAIN`を追加
  - `internal/session/session.go`の`SetValue()`メソッドでCookieの`Domain`属性を設定
  - 単体テストを更新
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [x] **3-2**: Rails版のCookieドメイン設定を更新
  - `rails/config/initializers/session_store.rb`の`domain: :all`を`ENV.fetch("ANNICT_COOKIE_DOMAIN")`に変更
  - `rails/.env.development`に`ANNICT_COOKIE_DOMAIN`を追加
  - `rails/.env.test`に`ANNICT_COOKIE_DOMAIN`を追加
  - **想定ファイル数**: 約 3 ファイル（実装 3）
  - **想定行数**: 約 10 行（実装 10 行）

### フェーズ 4: セッションクリーンアップタスクの更新

- [x] **4-1**: セッションクリーンアップタスクの削除期間を30日に変更
  - `rails/lib/tasks/session.rake`の削除期間を10日から30日に変更
  - `expire_after: 30.days`設定と統一する
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 5 行（実装 5 行）

### フェーズ 5: ドキュメント更新

- [x] **5-1**: CLAUDE.mdの更新
  - `CLAUDE.md`の「セッションストア（Redis）」の記述を修正（実際はPostgreSQLを使用）
  - セッション管理の動作を正確に記述
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 20 行（実装 20 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **Redisベースのセッションストアへの移行**: 現在のActiveRecordベースのセッションストアは正常に動作しており、移行の必要性は低い。将来的にパフォーマンス問題が発生した場合に検討する。なお、CLAUDE.mdに「セッションストア（Redis）」と記載されているのは誤りで、実際はPostgreSQLのsessionsテーブルを使用している
- **セッション更新の非同期化**: 現在のセッション更新処理は軽量（UPDATE文1つ）であり、非同期化の必要性は低い。将来的にパフォーマンス問題が発生した場合に検討する
- **セッション更新の条件付き実行（1時間ごと）**: 実装の複雑さを考慮し、毎回更新する方式を採用する。Rails版と同じ動作を実現するため、条件付き実行は不要
- **"Remember Me"（ログイン状態を保持）機能**: セッション期限を短くする場合、長期間ログイン状態を維持したいユーザー向けに別途実装を検討できるが、今回はスコープ外とする
- **session.Managerのリファクタリング**: 現在の`internal/session/session.go`の`Manager`は`query.Queries`に直接依存している（アーキテクチャガイドライン違反）。将来的にはSessionRepositoryを使用するようにリファクタリングすべきだが、既存コードへの影響が大きいため今回はスコープ外とする

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Rails ActiveRecord SessionStore](https://github.com/rails/activerecord-session_store)
- [Rails 7.0 Session Store Guide](https://guides.rubyonrails.org/action_controller_overview.html#session)
- [Go Chi Router Middleware](https://github.com/go-chi/chi)

---

## 調査結果サマリー

### 問題の発見経緯

ユーザーから「ログインしてもすぐログアウトされる」という報告があり、セッション管理の調査を実施しました。

### 調査で判明した事実

1. **セッションストア**: Rails版とGo版の両方が**PostgreSQLの`sessions`テーブル**を使用（Redisではない）
2. **セッションクリーンアップ**: 毎日19:00に`rake session:sweep`が実行され、`updated_at < 10.days.ago`のセッションを削除
3. **Rails版の動作**: ActiveRecord SessionStoreが各リクエストで`updated_at`を自動更新
4. **Go版の動作**: `GetSession()`でセッションを読み取るだけで`updated_at`を更新しない
5. **Cookieドメイン設定**: Rails版は`domain: :all`、Go版は`domain`属性なし（不整合）

### 問題の原因

Go版のページだけを10日以上使い続けるユーザーのセッションが、`rake session:sweep`によって削除され、突然ログアウトされる。

### 解決策

1. Go版の認証ミドルウェアでセッションの`updated_at`を更新する
2. Rails版とGo版でCookieのドメイン設定を統一する
3. セッションクリーンアップタスクの削除期間を30日に変更し、`expire_after`設定と統一する

### 影響範囲

**現在の問題の影響**:

- **影響を受けるユーザー**: Go版のページ（`/sign_in`, `/password/*`など）だけを10日以上使い続けるユーザー
- **影響の頻度**: 毎日19:00のクリーンアップタスク実行後に発生
- **影響の深刻度**: 高（ユーザーが突然ログアウトされ、UXが著しく低下）

**修正後の影響**:

- **修正の効果**: Go版のページを使い続けても、30日間ログイン状態が維持される
- **既存ユーザーへの影響**: なし（セッション期限は30日のまま維持）
- **メリット**: まとめて記録するユーザーも快適に利用できる
