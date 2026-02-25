# コードレビュー: email

## レビュー情報

| 項目                   | 内容                                                |
| ---------------------- | --------------------------------------------------- |
| レビュー日             | 2026-02-08                                          |
| 対象ブランチ           | email                                               |
| ベースブランチ         | develop                                             |
| 設計書（指定があれば） | docs/designs/3_done/202602/unified-email-sending.md |
| 変更ファイル数         | 11 ファイル（実装 6 + テスト 5）                    |
| 変更行数（実装）       | +119 / -176 行                                      |
| 変更行数（テスト）     | +502 / -171 行                                      |

## 参照するガイドライン

### 共通

- [@CLAUDE.md](/workspace/CLAUDE.md) - プロジェクト全体のガイド（コミットメッセージ、コメント、PRガイドライン）

### Go版

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - Go版の開発ガイド（コーディング規約、テスト戦略）
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド

## 変更ファイル一覧

### 実装ファイル

- [x] `go/internal/mail/noop.go`
- [ ] `go/internal/mail/resend.go`
- [x] `go/internal/worker/client.go`
- [x] `go/internal/worker/send_password_reset_email.go`
- [x] `go/internal/worker/send_sign_in_code.go`
- [x] `go/internal/worker/send_sign_up_code.go`

### テストファイル

- [x] `go/internal/mail/noop_test.go`
- [x] `go/internal/mail/resend_test.go`
- [x] `go/internal/worker/send_password_reset_email_test.go`
- [x] `go/internal/worker/send_sign_in_code_test.go`
- [x] `go/internal/worker/send_sign_up_code_test.go`

### 設定・その他

- [x] `docs/designs/3_done/202602/unified-email-sending.md`
- [x] `docs/reviews/template.md`

## ファイルごとのレビュー結果

### `go/internal/mail/resend.go`

**ステータス**: 要修正

**チェックしたガイドライン**:

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - コーディング規約
- [設計書](/workspace/docs/designs/3_done/202602/unified-email-sending.md) - Senderインターフェース設計

**問題点・改善提案**:

- **[設計書#ResendSender]**: `Send`メソッドで `r.client.Emails.Send(params)` を使用しているが、設計書では `s.client.Emails.SendWithContext(ctx, params)` を使用している（設計書213行目）

  ```go
  // 問題のあるコード（78行目）
  _, err := r.client.Emails.Send(params)
  ```

  `Send` メソッドは内部で `SendWithContext(context.Background(), params)` を呼び出しており、関数に渡された `ctx` が無視される。これにより、リクエストのキャンセルやタイムアウトのコンテキスト伝播が失われる。

  **修正案**:

  ```go
  _, err := r.client.Emails.SendWithContext(ctx, params)
  ```

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [ ] 修正案の通り `SendWithContext(ctx, params)` に変更する
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

## 設計との整合性チェック

### 設計書との対応状況

| 設計書の要件                   | 実装状況    | 備考                                        |
| ------------------------------ | ----------- | ------------------------------------------- |
| Senderインターフェースの統一   | ✅ 実装済み | `MailSender` → `Sender` にリネーム          |
| SendInputの導入                | ✅ 実装済み | templ.Componentベースに変更                 |
| NoopSenderの追加               | ✅ 実装済み | `SentEmails`, `Reset()` を含む              |
| ResendSenderの改善             | ✅ 実装済み | fromName, タイムアウト対応                  |
| Worker内レンダリング処理の統一 | ✅ 実装済み | 3ワーカーすべて統一パターン                 |
| テンプレート選択関数の統一     | ✅ 実装済み | `xxxTemplates()` パターンに統一             |
| SendWithContextの使用          | ❌ 未対応   | `Send` を使用しておりコンテキスト伝播が欠落 |

### 前回レビューからの改善

前回レビュー（`email-202602080920.md`）で指摘された3つのワーカーテストの問題（`NoopSender` を活用していない）がすべて修正されている：

- `TestSendPasswordResetEmailWorker_Work`: `NoopSender`を使った統合テストに改善
- `TestSendSignInCodeWorker_Work`: 同上
- `TestSendSignUpCodeWorker_Work`: 同上

### 設計書との差異（前回レビューと同じ、問題なし）

| 項目               | 設計書                   | 実装                           | 重要度                                                                                                                             |
| ------------------ | ------------------------ | ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------- |
| パッケージ名       | `email`                  | `mail`                         | 低（既存の命名を継続使用しているため妥当）                                                                                         |
| ファイル構成       | `sender.go` に全て       | `resend.go` + `noop.go` に分割 | 低（分割の方が可読性が高い）                                                                                                       |
| 構造体名           | `ResendSender`           | `ResendClient`                 | 低（既存の命名を維持）                                                                                                             |
| `SendRaw` メソッド | Workerでの利用のため設計 | 未実装                         | 低（Worker内で`Send`メソッドを使用する設計に変更。templ.Componentを直接Workerに渡す方式が採用されており、`SendRaw`は不要になった） |

## 総合評価

**評価**: Request Changes

**総評**:

前回レビューで指摘された3つのワーカーテストの問題がすべて修正されており、`NoopSender`を活用した統合テストが適切に実装されている。テストコードの品質が大幅に向上した。

**良かった点**:

- 前回レビューの指摘事項がすべて対応されている
- テーブル駆動テストパターンが適切に使用されている（日本語/英語ロケールのテスト）
- `NoopSender`を使ったテストで、メール送信の宛先・件名・テンプレート内容まで検証されている
- ログ出力が `slog` を使用しておりガイドラインに準拠している

**修正が必要な点**:

- `resend.go` の `Send` メソッドで `r.client.Emails.Send(params)` を使用しており、コンテキスト伝播が欠落している。設計書通り `SendWithContext(ctx, params)` に変更が必要（1件）
