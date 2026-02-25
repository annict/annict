# セッションCookie設定処理の統一化 設計書

## 概要

セッションCookieの設定処理が複数箇所に散らばっており、設定値の不統一やメンテナンス性の低下を招いている。この問題を解決するため、Cookie設定の責務を `session.Manager` に集約し、ハンドラー層から直接 `http.SetCookie()` を呼び出すパターンを廃止する。

**目的**:

- Cookie設定ロジックを一箇所に集約し、設定値の不統一を解消する
- ハンドラー層の責務を明確化し、セッション管理の詳細を隠蔽する
- 将来の変更（Cookie属性の追加など）に対する保守性を向上させる

**背景**:

- 現在、`http.SetCookie()` が5箇所で呼び出されており、それぞれで設定値が異なる
- `Secure` 属性の判定ロジックが不統一（`== "true"` vs `!= "false"`）
- `Domain` 属性に異なるConfigフィールドが使用されている箇所がある
- `MaxAge` 属性がハンドラー層では設定されていない
- `_csrf_initialized` という使用されていないフラグが存在する

## 要件

### 機能要件

- セッションCookie設定処理を `session.Manager` に集約する
- ハンドラー層では `session.Manager` のメソッドを呼び出すだけで、Cookie設定が完了する
- Cookie属性（Domain, Secure, HttpOnly, SameSite, MaxAge）を統一する
- 不要な `_csrf_initialized` フラグを削除する

### 非機能要件

- **後方互換性**: 既存のセッション（Railsとの共有セッション）が引き続き動作すること
- **テスト容易性**: Cookie設定のテストが容易であること
- **保守性**: Cookie属性の変更が一箇所で完結すること

## 設計

### 現状の問題点

#### 1. Cookie設定箇所の散在

| ファイル                                      | 行      | 用途                              |
| --------------------------------------------- | ------- | --------------------------------- |
| `internal/session/session.go`                 | 176-186 | `CreateSession()`                 |
| `internal/session/session.go`                 | 239-249 | `SetValue()` 新規セッション作成時 |
| `internal/handler/sign_in_password/create.go` | 124-134 | パスワードログイン                |
| `internal/handler/sign_in_code/create.go`     | 203-213 | メールコードログイン              |
| `internal/handler/sign_up_username/create.go` | 80-89   | ユーザー登録                      |

#### 2. 設定値の不統一

| 属性   | session.go                    | sign_in_password          | sign_in_code              | sign_up_username                                |
| ------ | ----------------------------- | ------------------------- | ------------------------- | ----------------------------------------------- |
| Secure | `cfg.SessionSecure == "true"` | X-Forwarded-Proto対応あり | X-Forwarded-Proto対応あり | `cfg.SessionSecure != "false"` ← **逆ロジック** |
| Domain | `cfg.CookieDomain`            | `cfg.CookieDomain`        | `cfg.CookieDomain`        | `cfg.Domain` ← **異なるフィールド**             |
| MaxAge | 30日                          | **なし**                  | **なし**                  | **なし**                                        |

#### 3. 不要なコード

`internal/middleware/csrf.go:96` の `_csrf_initialized` は `SetValue()` を呼び出すためだけのダミーキーであり、実際には参照されていない。

### 解決策

#### 1. `session.Manager` に `SetSessionCookie()` メソッドを追加

```go
// SetSessionCookie はセッションCookieを設定する（内部ヘルパー）
func (m *Manager) setSessionCookie(w http.ResponseWriter, r *http.Request, publicID string) {
    secure := m.cfg.SessionSecure == "true"
    // リバースプロキシ経由のHTTPS接続を検出
    if r.Header.Get("X-Forwarded-Proto") == "https" {
        secure = true
    }

    cookie := &http.Cookie{
        Name:     SessionKey,
        Value:    publicID,
        Path:     "/",
        Domain:   m.cfg.CookieDomain,
        Secure:   secure,
        HttpOnly: m.cfg.SessionHTTPOnly == "true",
        SameSite: http.SameSiteLaxMode,
        MaxAge:   30 * 24 * 60 * 60, // 30日
    }
    http.SetCookie(w, cookie)
}
```

#### 2. ハンドラー向けのCookie設定メソッドを提供

usecase層とsession.Managerの責務を明確に分離する方針を採用する。

**責務の分離:**

- **usecase層**: ビジネスロジック（セッションDB保存、flashメッセージ設定など）を担当し、`PublicID` を返す
- **session.Manager**: Cookie設定ロジックを集約し、統一されたCookie属性で設定する
- **ハンドラー**: usecaseを呼び出した後、session.ManagerでCookieを設定

**実装方法:**

`session.Manager` に `SetSessionCookieByPublicID()` 公開メソッドを追加する。これは内部ヘルパー `setSessionCookie()` のラッパーとして機能する。

```go
// SetSessionCookieByPublicID はセッションCookieを設定する
// ハンドラーからusecase呼び出し後に使用
func (m *Manager) SetSessionCookieByPublicID(w http.ResponseWriter, r *http.Request, publicID string) {
    m.setSessionCookie(w, r, publicID)
}
```

**ハンドラーでの使用例:**

```go
// 変更前（Cookie設定ロジックがハンドラーに散らばっている）
sessionResult, err := h.createSessionUC.Execute(ctx, nil, userID, ...)
secure := h.cfg.SessionSecure == "true"
if r.Header.Get("X-Forwarded-Proto") == "https" {
    secure = true
}
cookie := &http.Cookie{
    Name:     session.SessionKey,
    Value:    sessionResult.PublicID,
    // ... 各ハンドラーで異なる設定 ...
}
http.SetCookie(w, cookie)

// 変更後（Cookie設定がsession.Managerに集約）
sessionResult, err := h.createSessionUC.Execute(ctx, nil, userID, ...)
h.sessionMgr.SetSessionCookieByPublicID(w, r, sessionResult.PublicID)
```

**メリット:**

- Cookie設定ロジックが `session.Manager` に完全に集約される
- ハンドラーは1行でCookie設定が完了
- Cookie属性（Domain, Secure, HttpOnly, SameSite, MaxAge）が統一される
- usecase層はHTTP層に依存しない（アーキテクチャの整合性を維持）

#### 3. `_csrf_initialized` の削除

`SetValue()` を呼び出す際に、CSRFトークン生成のためだけにダミーキーを使用するのではなく、専用のメソッドを追加する。

```go
// EnsureCSRFToken はセッションが存在しない場合に新規作成し、CSRFトークンを返す
// ログインフォームなど、セッションがまだ存在しない可能性があるページで使用
func (m *Manager) EnsureCSRFToken(ctx context.Context, w http.ResponseWriter, r *http.Request) (string, error) {
    // ... CSRFトークンの取得または生成 ...
}
```

### ディレクトリ構成（変更なし）

変更対象ファイル:

- `internal/session/session.go` - Cookie設定ロジックの集約
- `internal/middleware/csrf.go` - `GetOrCreateCSRFToken` の改善
- `internal/handler/sign_in_password/create.go` - `http.SetCookie()` の削除
- `internal/handler/sign_in_code/create.go` - `http.SetCookie()` の削除
- `internal/handler/sign_up_username/create.go` - `http.SetCookie()` の削除

## タスクリスト

### フェーズ 1: session.Manager の改善

- [x] **1-1**: `setSessionCookie()` 内部ヘルパーメソッドの追加
  - `session.Manager` に `setSessionCookie(w, r, publicID)` メソッドを追加
  - X-Forwarded-Proto ヘッダーによる Secure 属性の判定を含める
  - 既存の `CreateSession()` と `SetValue()` 内のCookie設定をこのメソッドに統一
  - テストを追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [x] **1-2**: `SetSessionCookieByPublicID()` 公開メソッドの追加
  - `session.Manager` に `SetSessionCookieByPublicID(w, r, publicID)` 公開メソッドを追加
  - 内部ヘルパー `setSessionCookie()` のラッパーとして実装
  - ハンドラーからusecase呼び出し後にCookieを設定するために使用
  - テストを追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 10 行 + テスト 20 行）

- [x] **1-3**: `EnsureCSRFToken()` メソッドの追加
  - CSRFトークンを取得または生成する専用メソッドを追加
  - セッションが存在しない場合は新規作成してCookieを設定
  - `_csrf_initialized` ダミーキーの使用を廃止
  - テストを追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 50 行 + テスト 50 行）

### フェーズ 2: ハンドラーの修正

- [x] **2-1**: sign_in_password ハンドラーの修正
  - `http.SetCookie()` の直接呼び出しを削除
  - `sessionMgr.SetSessionCookieByPublicID(w, r, sessionResult.PublicID)` を使用
  - Cookie設定ロジック（Secure属性判定など）の重複コードを削除
  - 既存テストが通ることを確認
  - **想定ファイル数**: 約 1 ファイル（実装のみ、テストは既存を流用）
  - **想定行数**: 約 -15 行（削除が多い）

- [x] **2-2**: sign_in_code ハンドラーの修正
  - `http.SetCookie()` の直接呼び出しを削除
  - `sessionMgr.SetSessionCookieByPublicID(w, r, sessionResult.PublicID)` を使用
  - Cookie設定ロジック（Secure属性判定など）の重複コードを削除
  - 既存テストが通ることを確認
  - **想定ファイル数**: 約 1 ファイル（実装のみ、テストは既存を流用）
  - **想定行数**: 約 -15 行（削除が多い）

- [x] **2-3**: sign_up_username ハンドラーの修正
  - `http.SetCookie()` の直接呼び出しを削除
  - `sessionMgr.SetSessionCookieByPublicID(w, r, sessionResult.PublicID)` を使用
  - `h.cfg.Domain` → `h.cfg.CookieDomain` の修正が自動的に解決される
  - 既存テストが通ることを確認
  - **想定ファイル数**: 約 1 ファイル（実装のみ、テストは既存を流用）
  - **想定行数**: 約 -15 行（削除が多い）

### フェーズ 3: CSRF ミドルウェアの修正

- [x] **3-1**: `GetOrCreateCSRFToken()` の改善
  - `_csrf_initialized` ダミーキーの使用を廃止
  - `session.Manager.EnsureCSRFToken()` を使用するように変更
  - 既存テストが通ることを確認
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 60 行（実装 20 行 + テスト 40 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **セッションストアの変更**: PostgreSQL から他のストアへの移行は行わない
- **Cookie名の変更**: Rails互換性のため `_annict_session_v201904` を維持
- **セッションデータ形式の変更**: Rails互換のJSON形式を維持

## 参考資料

- [Go net/http Cookie](https://pkg.go.dev/net/http#Cookie)
- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
- [RFC 6265 - HTTP State Management Mechanism](https://tools.ietf.org/html/rfc6265)
