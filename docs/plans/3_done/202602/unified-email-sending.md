# メール送信機能の統一設計書

<!--
このテンプレートの使い方:
1. このファイルを `docs/designs/2_todo/` ディレクトリにコピー
   例: cp docs/designs/template.md docs/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 実装ガイドラインの参照

<!--
**重要**: 設計書を作成する前に、対象プラットフォームのガイドラインを必ず確認してください。
特に以下の点に注意してください：
- ディレクトリ構造・ファイル名の命名規則
- コーディング規約
- アーキテクチャパターン

ガイドラインに沿わない設計は、実装時にそのまま実装されてしまうため、
設計書作成の段階でガイドラインに準拠していることを確認してください。
-->

### Go版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - 全体的なコーディング規約
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTPハンドラーガイドライン（**ファイル名は標準の8種類のみ**）
- [@go/docs/i18n-guide.md](/workspace/go/docs/i18n-guide.md) - 国際化ガイド
- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - セキュリティガイドライン
- [@go/docs/templ-guide.md](/workspace/go/docs/templ-guide.md) - templテンプレートガイド
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド

## 概要

3つのプロジェクト（Annict、Wikino、Mewst）のメール送信処理を統一し、各プロジェクトの良いところを取り入れた共通設計を実現する。

**目的**:

- 3プロジェクト間でメール送信処理のコードを統一し、保守性を向上させる
- 各プロジェクトの良い設計パターンを他のプロジェクトに展開する
- テスタビリティを向上させる（NoopSenderの導入）

**背景**:

現在、3つのプロジェクトでメール送信処理が異なる実装になっている：

| 項目             | Annict         | Wikino         | Mewst        |
| ---------------- | -------------- | -------------- | ------------ |
| インターフェース | `MailSender`   | なし           | `Sender`     |
| テスト用Noop     | なし           | なし           | あり         |
| 非同期処理       | Worker (river) | Worker (river) | なし（同期） |
| HTML+テキスト    | 両方           | HTMLのみ       | 両方         |
| 送信者名         | サポート       | なし           | なし         |

これらを統一することで、コードの一貫性と品質を向上させる。

## 要件

### 機能要件

- メール送信のインターフェースを統一する（`Sender`インターフェース）
- 本番用の`ResendSender`とテスト用の`NoopSender`を提供する
- HTML形式とテキスト形式の両方をサポートする（マルチパートメール）
- 送信者名（fromName）をサポートする（例: `"Mewst <noreply@mewst.com>"`）
- 非同期メール送信をサポートする（Worker経由、オプション）
- 多言語テンプレートをサポートする（日本語・英語）

### 非機能要件

- **テスタビリティ**: NoopSenderで送信内容を検証可能にする
- **拡張性**: 新しいメール種別を簡単に追加できる構造にする
- **保守性**: 3プロジェクトで同じ設計パターンを採用する

## 設計

### 統一設計の方針

各プロジェクトの良いところを取り入れた統一設計：

| 採用元              | 採用する設計要素                                       |
| ------------------- | ------------------------------------------------------ |
| Mewst               | インターフェースベース設計（`Sender`インターフェース） |
| Mewst               | NoopSenderによるテスト可能性                           |
| Mewst               | templ.Componentを引数に取る設計                        |
| Mewst               | 言語×形式でテンプレートファイル分離                    |
| Annict              | 送信者名（fromName）のサポート                         |
| Annict              | カスタムHTTPクライアント（タイムアウト設定）           |
| Annict/Wikino/Mewst | Worker（river）を使った非同期メール送信                |
| Annict              | 機能別テンプレートディレクトリ構成                     |

### パッケージ構成

```
internal/
├── email/
│   ├── sender.go           # Senderインターフェース、ResendSender、NoopSender
│   └── sender_test.go      # テスト
├── worker/                  # 非同期処理が必要な場合
│   ├── client.go
│   └── send_email.go       # メール送信Worker
└── templates/
    └── emails/
        ├── email_confirmation/
        │   ├── ja_html.templ
        │   ├── ja_text.templ
        │   ├── en_html.templ
        │   └── en_text.templ
        ├── password_reset/      # 機能別ディレクトリ
        │   ├── ja_html.templ
        │   ├── ja_text.templ
        │   ├── en_html.templ
        │   └── en_text.templ
        └── sign_in/             # 必要に応じて追加
            └── ...
```

### Senderインターフェース

```go
// Package email はメール送信機能を提供します
package email

import (
    "bytes"
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/a-h/templ"
    "github.com/resend/resend-go/v2"
)

// Sender はメール送信を行うインターフェース
type Sender interface {
    // Send はメールを送信する
    Send(ctx context.Context, input SendInput) error
}

// SendInput はメール送信の入力
type SendInput struct {
    To       string          // 送信先メールアドレス
    Subject  string          // 件名
    HTMLBody templ.Component // メール本文（HTML形式）
    TextBody templ.Component // メール本文（テキスト形式、nilの場合はHTMLのみ）
}
```

### ResendSender（本番用）

```go
// ResendSender はResend APIを使用してメールを送信する
type ResendSender struct {
    client    *resend.Client
    fromEmail string
    fromName  string
}

// NewResendSender は新しいResendSenderを作成する
func NewResendSender(apiKey, fromEmail, fromName string) *ResendSender {
    // カスタムHTTPクライアント（タイムアウト設定）
    httpClient := &http.Client{
        Timeout: 30 * time.Second,
    }
    client := resend.NewCustomClient(httpClient, apiKey)
    return &ResendSender{
        client:    client,
        fromEmail: fromEmail,
        fromName:  fromName,
    }
}

// from はFromアドレスを生成する
func (s *ResendSender) from() string {
    if s.fromName != "" {
        return fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
    }
    return s.fromEmail
}

// Send はメールを送信する
func (s *ResendSender) Send(ctx context.Context, input SendInput) error {
    // HTMLテンプレートをレンダリング
    var htmlBuf bytes.Buffer
    if err := input.HTMLBody.Render(ctx, &htmlBuf); err != nil {
        return fmt.Errorf("HTMLテンプレートのレンダリングに失敗しました: %w", err)
    }

    params := &resend.SendEmailRequest{
        From:    s.from(),
        To:      []string{input.To},
        Subject: input.Subject,
        Html:    htmlBuf.String(),
    }

    // テキストテンプレートがある場合はレンダリング
    if input.TextBody != nil {
        var textBuf bytes.Buffer
        if err := input.TextBody.Render(ctx, &textBuf); err != nil {
            return fmt.Errorf("テキストテンプレートのレンダリングに失敗しました: %w", err)
        }
        params.Text = textBuf.String()
    }

    _, err := s.client.Emails.SendWithContext(ctx, params)
    if err != nil {
        return fmt.Errorf("メール送信に失敗しました: %w", err)
    }

    return nil
}
```

### NoopSender（テスト用）

```go
// NoopSender はメールを送信しないダミー実装（テスト用）
type NoopSender struct {
    // SentEmails は送信されたメールを記録する（テスト用）
    SentEmails []SendInput
}

// NewNoopSender は新しいNoopSenderを作成する
func NewNoopSender() *NoopSender {
    return &NoopSender{
        SentEmails: make([]SendInput, 0),
    }
}

// Send はメールを送信せず、記録のみ行う
func (s *NoopSender) Send(_ context.Context, input SendInput) error {
    s.SentEmails = append(s.SentEmails, input)
    return nil
}

// Reset は送信記録をクリアする
func (s *NoopSender) Reset() {
    s.SentEmails = make([]SendInput, 0)
}
```

### テンプレート構造

**ディレクトリ構造**:

```
templates/emails/
├── email_confirmation/     # メール確認
│   ├── ja_html.templ      # 日本語HTML
│   ├── ja_text.templ      # 日本語テキスト
│   ├── en_html.templ      # 英語HTML
│   └── en_text.templ      # 英語テキスト
├── password_reset/         # パスワードリセット
│   ├── ja_html.templ
│   ├── ja_text.templ
│   ├── en_html.templ
│   └── en_text.templ
└── sign_in/                # ログインコード（Annictのみ）
    └── ...
```

**テンプレート関数のシグネチャ**:

```templ
// 各テンプレートは関数として定義
// 引数はメール種別ごとに異なる

// email_confirmation/ja_html.templ
templ JaHTML(email, code string) { ... }

// password_reset/ja_html.templ
templ JaHTML(email, resetURL string) { ... }
```

### 初期化方法

```go
// main.go での初期化

var emailSender email.Sender
if cfg.ResendAPIKey != "" && cfg.EmailFrom != "" {
    emailSender = email.NewResendSender(
        cfg.ResendAPIKey,
        cfg.EmailFrom,
        cfg.EmailFromName,  // 新規追加: 送信者名
    )
} else {
    emailSender = email.NewNoopSender()
    slog.Warn("Resend APIキーまたは送信元メールアドレスが設定されていないため、メール送信は無効です")
}
```

### 環境変数

```env
# メール送信（Resend API）
MEWST_RESEND_API_KEY=re_xxxxx
MEWST_EMAIL_FROM=noreply@mewst.com
MEWST_EMAIL_FROM_NAME=Mewst      # 新規追加
```

### ユースケースでの使用例

```go
// internal/usecase/create_email_confirmation.go

type CreateEmailConfirmationUsecase struct {
    emailConfirmRepo *repository.EmailConfirmationRepository
    emailSender      email.Sender
}

func (uc *CreateEmailConfirmationUsecase) Execute(ctx context.Context, input CreateEmailConfirmationInput) (*CreateEmailConfirmationResult, error) {
    // 確認コード生成
    code, err := generateConfirmationCode()
    if err != nil {
        return nil, fmt.Errorf("確認コードの生成に失敗: %w", err)
    }

    // DB保存
    ec, err := uc.emailConfirmRepo.Create(ctx, repository.CreateEmailConfirmationParams{
        Email: input.Email,
        Event: input.Event,
        Code:  code,
    })
    if err != nil {
        return nil, fmt.Errorf("メール確認レコードの作成に失敗: %w", err)
    }

    // テンプレート選択（ロケールに基づく）
    htmlBody, textBody := getEmailTemplates(input.Locale, input.Email, code)

    // メール送信
    if err := uc.emailSender.Send(ctx, email.SendInput{
        To:       input.Email,
        Subject:  getEmailSubject(input.Locale),
        HTMLBody: htmlBody,
        TextBody: textBody,
    }); err != nil {
        return nil, fmt.Errorf("確認メールの送信に失敗: %w", err)
    }

    return &CreateEmailConfirmationResult{
        EmailConfirmation: ec,
    }, nil
}
```

### テスト例

```go
func TestCreateEmailConfirmationUsecase_Execute(t *testing.T) {
    t.Parallel()

    _, tx := testutil.SetupTestDB(t)
    ctx := context.Background()

    // テスト用NoopSenderを使用
    emailSender := email.NewNoopSender()
    emailConfirmRepo := repository.NewEmailConfirmationRepository(tx)

    uc := usecase.NewCreateEmailConfirmationUsecase(emailConfirmRepo, emailSender)
    result, err := uc.Execute(ctx, usecase.CreateEmailConfirmationInput{
        Email:  "test@example.com",
        Event:  model.EmailConfirmationEventPasswordReset,
        Locale: "ja",
    })

    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    // メールが送信されたことを確認
    if len(emailSender.SentEmails) != 1 {
        t.Fatalf("SentEmails count = %v, want 1", len(emailSender.SentEmails))
    }

    sentEmail := emailSender.SentEmails[0]
    if sentEmail.To != "test@example.com" {
        t.Errorf("SentEmail.To = %v, want test@example.com", sentEmail.To)
    }
    if sentEmail.Subject != "[Mewst] 確認用コード" {
        t.Errorf("SentEmail.Subject = %v, want [Mewst] 確認用コード", sentEmail.Subject)
    }
}
```

### 非同期処理（Worker）が必要な場合

大量のメール送信や、メール送信の失敗をリトライしたい場合は、Workerを使った非同期処理を採用する。

```go
// internal/worker/send_email.go

type SendEmailArgs struct {
    To       string `json:"to"`
    Subject  string `json:"subject"`
    HTMLBody string `json:"html_body"`  // レンダリング済みHTML
    TextBody string `json:"text_body"`  // レンダリング済みテキスト
}

func (SendEmailArgs) Kind() string {
    return "send_email"
}

type SendEmailWorker struct {
    river.WorkerDefaults[SendEmailArgs]
    sender email.Sender
}

func (w *SendEmailWorker) Work(ctx context.Context, job *river.Job[SendEmailArgs]) error {
    // Workerの場合はレンダリング済み文字列を受け取るため、
    // templ.Componentではなくstring版のSendメソッドを使用
    return w.sender.SendRaw(ctx, email.SendRawInput{
        To:       job.Args.To,
        Subject:  job.Args.Subject,
        HTMLBody: job.Args.HTMLBody,
        TextBody: job.Args.TextBody,
    })
}
```

## 各プロジェクトへの適用

### Mewstへの適用

**現状**: インターフェースベース設計、NoopSender、同期処理
**変更点**:

- 送信者名（fromName）のサポートを追加
- HTTPクライアントのタイムアウト設定を追加
- インターフェースメソッド名を`SendEmailConfirmation`から`Send`に変更（汎用化）
- Worker（river）を使った非同期メール送信のサポートを追加

### Annictへの適用

**現状**: MailSenderインターフェース、Worker非同期処理、fromNameサポート
**変更点**:

- NoopSenderを追加（テスト可能性向上）
- インターフェースを`Sender`に統一
- `SendMultipartEmail`を`Send`に変更（templ.Componentを引数に）

### Wikinoへの適用

**現状**: 2層構造（Client + ConfirmationSender）、Worker非同期処理、HTMLのみ
**変更点**:

- `Sender`インターフェースを導入
- NoopSenderを追加
- テキスト形式のサポートを追加
- 送信者名（fromName）のサポートを追加
- テンプレートを言語×形式で分離

## タスクリスト

### フェーズ 1: Mewstの改善

- [x] **1-1**: [Go] 送信者名（fromName）サポートの追加
  - `email.ResendSender`に`fromName`フィールドを追加
  - `from()`メソッドで「名前 <メール>」形式を生成
  - 環境変数`MEWST_EMAIL_FROM_NAME`を追加
  - config.goの更新
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 50 行（実装 30 行 + テスト 20 行）

- [x] **1-2**: [Go] HTTPクライアントのタイムアウト設定追加
  - `NewResendSender`でカスタムHTTPクライアントを使用
  - タイムアウトを30秒に設定
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 15 行 + テスト 15 行）

- [x] **1-3**: [Go] インターフェースメソッド名の汎用化
  - `SendEmailConfirmation`を`Send`にリネーム
  - `SendEmailConfirmationInput`を`SendInput`にリネーム
  - 呼び出し元の更新
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 80 行（実装 50 行 + テスト 30 行）

- [x] **1-4**: [Go] Worker（river）を使った非同期メール送信のサポート追加
  - riverパッケージの導入（go.mod更新）
  - `internal/worker/`ディレクトリの作成
  - `internal/worker/client.go`: Workerクライアントの初期化
  - `internal/worker/send_email.go`: メール送信Worker
  - `email.Sender`に`SendRaw`メソッドを追加（レンダリング済み文字列用）
  - Usecaseからの呼び出しをWorker経由に変更
  - main.goでWorkerの初期化と起動
  - **想定ファイル数**: 約 8 ファイル（実装 6 + テスト 2）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

### フェーズ 2: Annictの改善

- [x] **2-1**: [Go] NoopSenderの追加
  - `email.NoopSender`構造体を追加
  - `SentEmails`スライスで送信内容を記録
  - テストでの使用例を追加
  - **想定ファイル数**: 約 3 ファイル（実装 1 + テスト 2）
  - **想定行数**: 約 60 行（実装 20 行 + テスト 40 行）

- [x] **2-2**: [Go] インターフェース名の統一
  - `MailSender`を`Sender`にリネーム
  - `SendMultipartEmail`を`Send`にリネーム
  - templ.Componentを引数に取るように変更
  - Worker内でのレンダリング処理を調整
  - **想定ファイル数**: 約 8 ファイル（実装 5 + テスト 3）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

### フェーズ 3: Wikinoの改善

- [x] **3-1**: [Go] Senderインターフェースの導入
  - `email.Sender`インターフェースを追加
  - `Client`を`ResendSender`にリネーム
  - `ConfirmationSender`を削除（Senderに統合）
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 100 行（実装 60 行 + テスト 40 行）

- [x] **3-2**: [Go] NoopSenderの追加
  - `email.NoopSender`構造体を追加
  - テストでの使用例を追加
  - **想定ファイル数**: 約 3 ファイル（実装 1 + テスト 2）
  - **想定行数**: 約 60 行（実装 20 行 + テスト 40 行）

- [x] **3-3**: [Go] テキスト形式のサポート追加
  - `SendInput`に`TextBody`フィールドを追加
  - テンプレートにテキスト版を追加（ja_text, en_text）
  - **想定ファイル数**: 約 5 ファイル（実装 4 + テスト 1）
  - **想定行数**: 約 80 行（実装 60 行 + テスト 20 行）

- [x] **3-4**: [Go] 送信者名（fromName）サポートの追加
  - Mewstと同じ実装
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 50 行（実装 30 行 + テスト 20 行）

- [x] **3-5**: [Go] テンプレートの言語×形式分離
  - 単一ファイルから4ファイル構成に変更
  - `IsJapanese`フラグをロケール文字列に変更
  - **想定ファイル数**: 約 6 ファイル（実装 5 + テスト 1）
  - **想定行数**: 約 120 行（実装 100 行 + テスト 20 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **メール送信のリトライ機能**: 必要に応じてWorker側で対応
- **メール送信の優先度制御**: 現時点では不要
- **添付ファイルのサポート**: 現時点では不要
- **複数宛先への同時送信**: 現時点では不要

## 参考資料

- [Resend Go SDK](https://github.com/resend/resend-go)
- [River - Background job processing for Go](https://github.com/riverqueue/river)
- [templ - HTML templates in Go](https://templ.guide/)
