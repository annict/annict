# Resend メールクライアントのリファクタリング 設計書

## 概要

`internal/mail/resend.go` に定義されている `ResendClient` を活用し、メール送信ロジックを集約するリファクタリングを行う。

**目的**:

- Resend クライアント生成ロジックの重複を排除する
- メール送信処理を一元化し、保守性を向上させる
- Worker から直接 `resend.Client` を使用するのではなく、`mail.ResendClient` 経由でメールを送信する

**背景**:

- 現在 `internal/mail/resend.go` に `ResendClient` が定義されているが、使用されていない
- `internal/worker/client.go` で Resend クライアントを直接作成しており、重複が発生している
- 各 Worker が `*resend.Client` を直接操作し、From アドレス設定などのロジックが重複している

## 要件

### 機能要件

- `mail.ResendClient` を使用してメール送信を行う
- `worker/client.go` での Resend クライアント直接生成を廃止し、`mail.NewResendClient` を使用する
- 各 Worker（`SendPasswordResetEmailWorker`, `SendSignInCodeWorker`, `SendSignUpCodeWorker`）は `mail.ResendClient` を依存として受け取る
- From アドレス（`Annict <xxx@example.com>`）の設定を `mail.ResendClient` に集約する
- 既存のテストが引き続きパスすること

### 非機能要件

- 後方互換性: 外部からの呼び出しインターフェースは変更しない（Worker の引数型は内部変更のみ）
- テスト容易性: `mail.ResendClient` をインターフェース化し、モック可能にする

## 設計

### アーキテクチャ

```
┌─────────────────────────────────────────────────────────┐
│ worker/client.go                                        │
│ - mail.NewResendClient() を呼び出し                      │
│ - mail.ResendClient を各 Worker に渡す                   │
└─────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────┐
│ mail/resend.go                                          │
│ - ResendClient 構造体                                    │
│ - SendMultipartEmail() メソッド                          │
│ - From アドレス設定を内部で管理                           │
└─────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────┐
│ resend-go/v2                                            │
│ - Resend API クライアント                                │
└─────────────────────────────────────────────────────────┘
```

### コード設計

#### `internal/mail/resend.go` の変更

```go
// Package mail はmail機能を提供します
package mail

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/resend/resend-go/v2"
)

// MailSender はメール送信のインターフェース
type MailSender interface {
	SendMultipartEmail(ctx context.Context, to, subject, text, html string) error
}

// ResendClient はResend APIを使用したメール送信クライアント
type ResendClient struct {
	client    *resend.Client
	fromEmail string
	fromName  string
}

// NewResendClient は新しいResendClientを作成します
func NewResendClient(apiKey, fromEmail, fromName string) *ResendClient {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	client := resend.NewCustomClient(httpClient, apiKey)
	return &ResendClient{
		client:    client,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// from はFromアドレスを生成します
func (r *ResendClient) from() string {
	if r.fromName != "" {
		return fmt.Sprintf("%s <%s>", r.fromName, r.fromEmail)
	}
	return r.fromEmail
}

// SendMultipartEmail はテキストとHTMLの両方を含むメールを送信します
func (r *ResendClient) SendMultipartEmail(ctx context.Context, to, subject, text, html string) error {
	params := &resend.SendEmailRequest{
		From:    r.from(),
		To:      []string{to},
		Subject: subject,
		Text:    text,
		Html:    html,
	}

	_, err := r.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
```

#### `internal/worker/client.go` の変更

```go
// mail.NewResendClient を使用するように変更
var mailClient mail.MailSender
if cfg.ResendAPIKey != "" {
	mailClient = mail.NewResendClient(cfg.ResendAPIKey, cfg.ResendFromEmail, "Annict")
	slog.InfoContext(ctx, "Resend クライアントを初期化しました")
} else {
	slog.WarnContext(ctx, "Resend API キーが設定されていません。メール送信機能は利用できません")
}

// Worker への受け渡し
if mailClient != nil {
	river.AddWorker(workers, NewSendPasswordResetEmailWorker(queries, mailClient, cfg))
	// ...
}
```

#### 各 Worker の変更

Worker は `mail.MailSender` インターフェースを受け取るように変更：

```go
type SendPasswordResetEmailWorker struct {
	river.WorkerDefaults[SendPasswordResetEmailArgs]
	queries    *query.Queries
	mailClient mail.MailSender  // *resend.Client から変更
	cfg        *config.Config
}

// Work メソッド内
err := w.mailClient.SendMultipartEmail(ctx, email, subject, textBody, htmlBody)
```

### 削除対象

- `mail.ResendClient.SendTextEmail()` - 使用されていないため削除
- `mail.ResendClient.SendHTMLEmail()` - 使用されていないため削除

### テスト戦略

- `mail.MailSender` インターフェースを使用し、Worker のテストではモックを注入可能にする
- `mail/resend_test.go` は既存のテストを更新（不要な関数テストを削除）

## タスクリスト

### フェーズ 1: mail パッケージの修正

- [x] **1-1**: `mail/resend.go` の修正
  - `MailSender` インターフェースを追加
  - `NewResendClient` の引数に `fromName` を追加
  - `from()` ヘルパーメソッドを追加
  - `SendTextEmail`, `SendHTMLEmail` を削除（未使用）
  - `SendMultipartEmail` を更新（ctx は現状使っていないが、将来の拡張のために残す）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 50 行 + テスト 50 行）

### フェーズ 2: Worker の修正

- [x] **2-1**: `worker/client.go` の修正
  - `mail.NewResendClient` を使用するように変更
  - `mail.MailSender` を各 Worker に渡すように変更
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行（実装 30 行）

- [x] **2-2**: 各 Worker の修正
  - `SendPasswordResetEmailWorker` を `mail.MailSender` を使用するように変更
  - `SendSignInCodeWorker` を `mail.MailSender` を使用するように変更
  - `SendSignUpCodeWorker` を `mail.MailSender` を使用するように変更
  - From アドレス設定ロジックを削除（`mail.ResendClient` に委譲）
  - **想定ファイル数**: 約 3 ファイル（実装 3）
  - **想定行数**: 約 60 行（実装 60 行、各 Worker で約 20 行の削減/変更）

### フェーズ 3: テストとクリーンアップ

- [x] **3-1**: テストの修正と実行
  - 既存のテストを更新
  - 全テストがパスすることを確認
  - **想定ファイル数**: 約 2 ファイル（テスト 2）
  - **想定行数**: 約 50 行（テスト 50 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **メール送信のリトライ機能**: River のリトライ機能で対応済み
- **メールテンプレートの共通化**: 現状のテンプレート構造を維持
- **メール送信結果の永続化**: Resend 側で管理されるため不要

## 参考資料

- [Resend Go SDK](https://github.com/resend/resend-go)
- [River - PostgreSQL ベースのジョブキュー](https://riverqueue.com/)
