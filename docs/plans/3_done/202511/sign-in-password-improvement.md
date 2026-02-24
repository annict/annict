# パスワードログインページのUX改善 設計書

## 概要

`GET /sign_in/password` ページでメールアドレスの再入力を不要にし、ユーザー体験を改善する。

**目的**:

- `/sign_in` で入力済みのメールアドレスを `/sign_in/password` で再度入力させない
- ログインフローをシンプルにし、ユーザーの入力負担を軽減

**背景**:

- 現在のフローでは、`/sign_in` でメールアドレスを入力した後、`/sign_in/password` でまたメールアドレス入力を求めている
- `POST /sign_in` でセッションに `sign_in_email` を保存しているが、`/sign_in/password` で活用されていない
- 同じ情報を2度入力させるのは冗長でユーザー体験が悪い

## 要件

### 機能要件

- `GET /sign_in/password` でセッションから `sign_in_email` を取得し、メールアドレスを表示のみ（編集不可）にする
- ユーザーはパスワードのみ入力すればログインできる
- 「別のアカウントでログイン」リンクを表示し、`/sign_in` に戻れるようにする
- セッションに `sign_in_email` がない場合（直接アクセスなど）は `/sign_in` にリダイレクト
- `POST /sign_in/password` はセッションの `sign_in_email` を使用して認証する

### 非機能要件

#### セキュリティ

- セッションはサーバーサイドで管理されており、クライアントからの改ざんは不可能

#### ユーザビリティ

- パスワード入力フィールドに自動フォーカス
- 表示されるメールアドレスが明確にわかるデザイン

## 設計

### 画面設計

```
┌─────────────────────────────────┐
│         [Annictロゴ]            │
│                                 │
│           ログイン              │
│                                 │
│  example@example.com でログイン  │
│                                 │
│  パスワード                      │
│  [________________________]     │
│                  パスワードを忘れた │
│                                 │
│  [ログイン]                      │
│                                 │
│  別のアカウントでログイン         │
└─────────────────────────────────┘
```

### API設計

#### GET /sign_in/password

**変更内容**:

1. セッションから `sign_in_email` を取得
2. 取得できない場合は `/sign_in` にリダイレクト
3. テンプレートに `email` を渡す

#### POST /sign_in/password

**変更内容**:

1. フォームから `email_username` を受け取る代わりに、セッションから `sign_in_email` を取得
2. `sign_in_email` がない場合は `/sign_in` にリダイレクト
3. メールアドレスでDBからユーザーを取得
4. パスワード検証後、セッションの `sign_in_email` をクリア

### コード設計

#### 変更対象ファイル

1. `internal/handler/sign_in_password/new.go` - セッションからメールアドレス取得
2. `internal/handler/sign_in_password/create.go` - セッションからメールアドレス取得
3. `internal/handler/sign_in_password/request.go` - `EmailOrUsername` フィールド削除
4. `internal/templates/pages/sign_in_password/show.templ` - UI変更
5. `internal/i18n/locales/ja.toml` - 翻訳追加
6. `internal/i18n/locales/en.toml` - 翻訳追加

## タスクリスト

### フェーズ 1: バックエンド実装

- [x] **1-1**: ハンドラーの修正

  - `new.go`: セッションから `sign_in_email` を取得し、テンプレートに渡す
  - `new.go`: セッションに `sign_in_email` がない場合は `/sign_in` にリダイレクト
  - `create.go`: セッションから `sign_in_email` を取得
  - `create.go`: `sign_in_email` がない場合は `/sign_in` にリダイレクト
  - `create.go`: メールアドレスでDBからユーザーを取得
  - `create.go`: 認証成功後、`sign_in_email` をセッションから削除
  - `request.go`: `EmailOrUsername` フィールドを削除
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 250 行（実装 100 行 + テスト 150 行）

### フェーズ 2: フロントエンド実装

- [x] **2-1**: テンプレートの修正

  - `show.templ`: メールアドレス入力フィールドを削除
  - `show.templ`: メールアドレスを表示のみに変更（「〇〇 でログイン」）
  - `show.templ`: 「別のアカウントでログイン」リンクを追加
  - `show.templ`: パスワードフィールドに `autofocus` を設定
  - 翻訳ファイル: 新しいメッセージキーを追加
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **メールアドレスの編集機能**: 別のアカウントでログインしたい場合は「別のアカウントでログイン」リンクを使用
- **セッションタイムアウト後の自動リダイレクト**: 既存のセッションタイムアウト処理に依存

## 参考資料

- [現在のログインフロー実装](/workspace/go/internal/handler/sign_in/create.go)
- [パスワードログインハンドラー](/workspace/go/internal/handler/sign_in_password/)
