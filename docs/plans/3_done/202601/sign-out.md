# ログアウト機能 設計書

## 概要

Go版Annictにおけるログアウト機能を実装します。ユーザーがサイドバーからログアウトボタンをクリックすることで、セッションを削除してログアウト状態にする機能です。

**目的**:

- ユーザーが安全にログアウトできるようにする
- Rails版との段階的移行において、Go版でセッション管理を完結させる

**背景**:

- 現在Go版のサイドバーにはログアウトリンクが存在するが、Rails版のエンドポイントを呼び出す形になっている
- Rails版とGo版でCSRFトークンの生成・検証ロジックが異なるため、Go版からRails版のログアウト機能を直接呼び出すことが困難
- セッションストアは共有されているため、Go版でセッションを削除すればRails版でもログアウトされる
- Rails版のサイドバーからもログアウトできる必要がある（Rails版のページにはGo版のCSRFトークンが含まれないため、CSRFミドルウェアは適用しない）

## 要件

### 機能要件

- ユーザーはGo版サイドバーからログアウトできる
- ユーザーはRails版サイドバーからもログアウトできる
- ログアウト時に確認ダイアログを表示する（Go版: Datastarの`confirm()`を使用、Rails版: 既存の`data_confirm`を使用）
- ログアウト後はホームページ（`/`）にリダイレクトする
- セッションをDBから削除する
- セッションCookieを削除する

### 非機能要件

- **セキュリティ**: ログアウトはCSRFミドルウェアを適用しない（Rails版からのリクエスト対応のため。ログアウトは破壊的操作ではないため安全）
- **互換性**: Rails版サイドバーとGo版サイドバーの両方からログアウトできる
- **UX**: 確認ダイアログで誤操作を防止する

## 設計

### API設計

| メソッド | パス | 説明 |
|----------|------|------|
| DELETE | `/sign_out` | ログアウト処理（Rails UJSからのリクエスト） |
| POST | `/sign_out` | ログアウト処理（Go版HTMLフォームからのリクエスト） |

**リクエスト（Go版サイドバー）**:
- POSTメソッド + `_method=DELETE` パラメータ（HTMLフォームからの送信時）
- MethodOverrideミドルウェアがDELETEに変換するが、Chiのルーティング順序によりPOSTとして登録が必要

**リクエスト（Rails版サイドバー）**:
- DELETEメソッド（Rails UJSが`data_method: :delete`を直接DELETEリクエストとして送信）

**レスポンス**:
- 成功時: 302リダイレクト → `/`
- 未ログイン時: 302リダイレクト → `/`

### コード設計

**新規ファイル**:
- `internal/handler/sign_out/handler.go` - ハンドラー構造体
- `internal/handler/sign_out/delete.go` - ログアウト処理

**修正ファイル**:
- `internal/query/queries/sessions.sql` - DeleteSessionクエリ追加
- `internal/query/sessions.sql.go` - sqlc生成ファイル（自動生成）
- `internal/repository/session.go` - DeleteSessionメソッド追加
- `internal/session/session.go` - DestroySessionメソッド追加
- `internal/templates/components/sidebar.templ` - ログアウトボタンをフォームに変更
- `internal/i18n/locales/ja.toml` - 確認メッセージの翻訳追加
- `internal/i18n/locales/en.toml` - 確認メッセージの翻訳追加
- `cmd/server/main.go` - ルーティング追加

### CSRFトークンについて

ログアウト機能ではCSRFトークン検証を行いません。理由:
- Rails版のサイドバーからもログアウトできる必要がある
- Rails版のページにはGo版のCSRFトークンが含まれない
- ログアウトは破壊的操作ではないためCSRF攻撃のリスクが低い

### セッション削除の流れ

**Go版サイドバーからの場合**:
1. ユーザーがログアウトボタンをクリック
2. 確認ダイアログを表示（「ログアウトしますか？」）
3. OKをクリックするとフォームが送信される（POST + `_method=DELETE`）
4. セッションIDをCookieから取得
5. DBからセッションレコードを削除
6. セッションCookieを削除（MaxAge=-1）
7. ホームページにリダイレクト

**Rails版サイドバーからの場合**:
1. ユーザーがログアウトリンクをクリック
2. Rails UJSが確認ダイアログを表示（既存の`data_confirm`）
3. OKをクリックするとDELETEリクエストが送信される
4. 以降はGo版と同じ（セッション削除、Cookie削除、リダイレクト）

## タスクリスト

### フェーズ 1: セッション削除機能の実装

- [x] **1-1**: セッション削除クエリとRepositoryの追加
  - `internal/query/queries/sessions.sql` に `DeleteSession` クエリを追加
  - `sqlc generate` でGoコードを生成
  - `internal/repository/session.go` に `DeleteSession` メソッドを追加
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 60 行（実装 20 行 + テスト 40 行）

- [x] **1-2**: SessionManagerにDestroySessionメソッドを追加
  - `internal/session/session.go` に `DestroySession` メソッドを追加
  - セッションDBレコードの削除とCookie削除を行う
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 80 行（実装 30 行 + テスト 50 行）

### フェーズ 2: ハンドラー実装

- [x] **2-1**: ログアウトハンドラーの実装
  - `internal/handler/sign_out/handler.go` - Handler構造体の定義
  - `internal/handler/sign_out/delete.go` - DELETE処理の実装
  - `cmd/server/main.go` にルーティング追加
    - CSRFミドルウェアは適用しない（Rails版からのリクエスト対応のため）
    - DELETEとPOST両方のメソッドを登録
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

### フェーズ 3: UI修正

- [x] **3-1**: 翻訳キーの追加とサイドバーのログアウトボタン修正
  - `internal/i18n/locales/ja.toml` に `nav_sign_out_confirm` を追加
  - `internal/i18n/locales/en.toml` に `nav_sign_out_confirm` を追加
  - `internal/templates/components/sidebar.templ` のログアウト部分をフォームに変更
  - CSRFトークンは含めない（検証しないため、Rails版との一貫性のため）
  - Datastarの`confirm()`で確認ダイアログを表示
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

### フェーズ 4: Rails版のログアウト機能削除

- [x] **4-1**: Rails版のログアウト機能を削除
  - Go版でログアウト機能が完成したため、Rails版のログアウトエンドポイントは不要
  - Rails版のサイドバーからはGo版のログアウトエンドポイントを呼び出すように変更済み
  - **想定ファイル数**: 約 2-3 ファイル
  - **想定行数**: 約 30 行（削除）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **ヘッダーのログアウトボタン**: 現在のヘッダーは一時的なものであとで削除予定のため
- **他端末からの強制ログアウト**: 全セッション削除機能は今回のスコープ外
- **ログアウト時の確認画面ページ**: 確認ダイアログで十分

## 参考資料

- [Datastar Delete Row Example](https://data-star.dev/examples/delete_row) - confirm()の使い方
- `internal/session/session.go` - 既存のセッション管理実装
- `internal/repository/session.go` - 既存のSessionRepository実装
- `/mewst/go/cmd/server/main.go` - MewstのログアウトエンドポイントのRails対応実装例
