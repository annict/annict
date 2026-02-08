# コードレビュー: email

## レビュー情報

| 項目                   | 内容                                                    |
| ---------------------- | ------------------------------------------------------- |
| レビュー日             | 2026-02-08                                              |
| 対象ブランチ           | email                                                   |
| ベースブランチ         | develop                                                 |
| 設計書（指定があれば） | docs/designs/3_done/202602/unified-email-sending.md      |
| 変更ファイル数         | 11 ファイル（実装 6 + テスト 5）                          |
| 変更行数（実装）       | +120 / -177 行                                          |
| 変更行数（テスト）     | +502 / -171 行                                          |

## 参照するガイドライン

### 共通

- [@CLAUDE.md](/workspace/CLAUDE.md) - プロジェクト全体のガイド（コミットメッセージ、コメント、PRガイドライン）

### Go版

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - Go版の開発ガイド（コーディング規約、テスト戦略）
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド

## 変更ファイル一覧

### 実装ファイル

- [x] `go/internal/mail/noop.go`
- [x] `go/internal/mail/resend.go`
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

すべてのファイルが問題なし。変更ファイル一覧のチェックボックスで完了を示す。

## 設計との整合性チェック

### 設計書との対応状況

| 設計書の要件 | 実装状況 | 備考 |
|---|---|---|
| Senderインターフェースの統一 | ✅ 実装済み | `MailSender` → `Sender` にリネーム |
| SendInputの導入 | ✅ 実装済み | templ.Componentベースに変更 |
| NoopSenderの追加 | ✅ 実装済み | `SentEmails`, `Reset()` を含む |
| ResendSenderの改善 | ✅ 実装済み | fromName, タイムアウト、SendWithContext対応 |
| Worker内レンダリング処理の統一 | ✅ 実装済み | 3ワーカーすべて統一パターン |
| テンプレート選択関数の統一 | ✅ 実装済み | `xxxTemplates()` パターンに統一 |
| SendWithContextの使用 | ✅ 対応済み | 前回レビューの指摘を修正済み |

### 前回レビュー（email-202602080930.md）からの改善

前回レビューで指摘された1件の問題がすべて修正されている：

- `resend.go` の `Send` メソッドで `r.client.Emails.Send(params)` → `r.client.Emails.SendWithContext(ctx, params)` に変更済み。コンテキスト伝播が正しく動作する

### 設計書との差異（既知、問題なし）

| 項目 | 設計書 | 実装 | 重要度 |
|---|---|---|---|
| パッケージ名 | `email` | `mail` | 低（既存の命名を継続使用しているため妥当） |
| ファイル構成 | `sender.go` に全て | `resend.go` + `noop.go` に分割 | 低（分割の方が可読性が高い） |
| 構造体名 | `ResendSender` | `ResendClient` | 低（既存の命名を維持） |
| `SendRaw` メソッド | Workerでの利用のため設計 | 未実装 | 低（Worker内で`Send`メソッドを使用する設計に変更。templ.Componentを直接Workerに渡す方式が採用されており、`SendRaw`は不要になった） |

## 総合評価

**評価**: Approve

**総評**:

前回レビュー（email-202602080930.md）で指摘された `SendWithContext` の問題が修正されており、すべてのレビュー指摘事項が対応済み。

**良かった点**:

- **インターフェースの統一**: `MailSender`/`SendMultipartEmail` → `Sender`/`Send` へのリネームが設計書の方針通りに実施されている
- **コンテキスト伝播の修正**: `SendWithContext(ctx, params)` を使用し、リクエストのキャンセルやタイムアウトが正しく伝播される
- **テンプレート選択の簡素化**: 各ワーカーの `renderXXXTemplate(ctx, locale, format, ...)` が `xxxTemplates(locale, ...)` に統一され、templ.Componentを直接返す設計に改善。コード量が大幅に削減された
- **NoopSenderの充実したテスト**: インターフェース実装チェック、基本送信、HTMLのみ、複数送信、Resetの各パターンを網羅
- **ワーカーテストの改善**: 3つのワーカーすべてで`NoopSender`を使った統合テストが実装され、メール送信の宛先・件名・テンプレート内容まで検証されている
- **ログ出力が `slog` ガイドラインに準拠**: すべてのログが `slog.InfoContext`/`slog.ErrorContext` を使用
- **エラーハンドリング**: `%w` による適切なエラーラッピングが一貫して使用されている
