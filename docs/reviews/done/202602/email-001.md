# コードレビュー: email

## レビュー情報

| 項目                   | 内容                                                    |
| ---------------------- | ------------------------------------------------------- |
| レビュー日             | 2026-02-08                                              |
| 対象ブランチ           | email                                                   |
| ベースブランチ         | develop                                                 |
| 設計書（指定があれば） | docs/designs/3_done/202602/unified-email-sending.md      |
| 変更ファイル数         | 11 ファイル（実装+テスト。設計書・テンプレートを除く）       |
| 変更行数（実装）       | +153 / -176 行                                          |
| 変更行数（テスト）     | +235 / -131 行                                          |

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
- [ ] `go/internal/worker/send_password_reset_email_test.go`
- [ ] `go/internal/worker/send_sign_in_code_test.go`
- [ ] `go/internal/worker/send_sign_up_code_test.go`

### 設定・その他

- [x] `docs/designs/3_done/202602/unified-email-sending.md`
- [x] `docs/reviews/template.md`

## ファイルごとのレビュー結果

### `go/internal/worker/send_password_reset_email_test.go`

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/CLAUDE.md#テスト戦略](/workspace/go/CLAUDE.md) - テストのベストプラクティス

**問題点・改善提案**:

- **[@go/CLAUDE.md#テスト戦略]**: `TestSendPasswordResetEmailWorker_Work` がワーカーの `Work` メソッドを実際にテストしていない

  ```go
  // 問題のあるコード（84-105行目）
  func TestSendPasswordResetEmailWorker_Work(t *testing.T) {
      _, tx := testutil.SetupTestDB(t)
      // ユーザーを作成するが、Workerを呼び出さない
      _ = testutil.NewUserBuilder(t, tx).WithUsername("ja_user_worker_test")...Build()
      _ = testutil.NewUserBuilder(t, tx).WithUsername("en_user_worker_test")...Build()
      t.Log("メール送信ワーカーの基本的な動作を確認しました")
  }
  ```

  テスト用の `NoopSender` が導入されたため、ワーカーの `Work` メソッドを `NoopSender` を使って実際に呼び出し、メールが送信されたことを検証できるはず。現状はDBセットアップのみで実質何もテストしていない。

  **修正案**: `NoopSender` を使って `Work` メソッドを呼び出し、`SentEmails` の内容を検証するテストを追加する。

  **対応方針**:

  - [x] NoopSenderを使ったワーカーの統合テストを追加する
  - [ ] 現状のテストで十分と判断し、変更しない
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

### `go/internal/worker/send_sign_in_code_test.go`

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/CLAUDE.md#テスト戦略](/workspace/go/CLAUDE.md) - テストのベストプラクティス

**問題点・改善提案**:

- **[@go/CLAUDE.md#テスト戦略]**: `TestSendSignInCodeWorker_Work` が `send_password_reset_email_test.go` と同様に、ワーカーの `Work` メソッドを実際にテストしていない

  ```go
  // 問題のあるコード（83-105行目）
  func TestSendSignInCodeWorker_Work(t *testing.T) {
      _, tx := testutil.SetupTestDB(t)
      // ユーザーを作成するが、Workerを呼び出さない
      _ = testutil.NewUserBuilder(t, tx)...Build()
      _ = testutil.NewUserBuilder(t, tx)...Build()
      t.Log("ログインコード送信ワーカーの基本的な動作を確認しました")
  }
  ```

  **修正案**: 上記と同じく `NoopSender` を使った統合テストに変更する。

  **対応方針**: `send_password_reset_email_test.go` と同じ方針に従う。

### `go/internal/worker/send_sign_up_code_test.go`

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/CLAUDE.md#テスト戦略](/workspace/go/CLAUDE.md) - テストのベストプラクティス

**問題点・改善提案**:

- **[@go/CLAUDE.md#テスト戦略]**: `TestSendSignUpCodeWorker_Work` が実質何もテストしていない

  ```go
  // 問題のあるコード（85-90行目）
  func TestSendSignUpCodeWorker_Work(t *testing.T) {
      t.Log("新規登録確認コード送信ワーカーの基本的な動作を確認しました")
  }
  ```

  DBセットアップすらなく、ログ出力のみ。

  **修正案**: 上記と同じく `NoopSender` を使った統合テストに変更する。

  **対応方針**: `send_password_reset_email_test.go` と同じ方針に従う。

## 設計との整合性チェック

### 設計書との対応状況

| 設計書の要件 | 実装状況 | 備考 |
|---|---|---|
| Senderインターフェースの統一 | ✅ 実装済み | `MailSender` → `Sender` にリネーム |
| SendInputの導入 | ✅ 実装済み | templ.Componentベースに変更 |
| NoopSenderの追加 | ✅ 実装済み | `SentEmails`, `Reset()` を含む |
| ResendSenderの改善 | ✅ 実装済み | fromName, タイムアウト対応 |
| Worker内レンダリング処理の統一 | ✅ 実装済み | 3ワーカーすべて統一パターン |
| テンプレート選択関数の統一 | ✅ 実装済み | `xxxTemplates()` パターンに統一 |

### 設計書との差異

| 項目 | 設計書 | 実装 | 重要度 |
|---|---|---|---|
| パッケージ名 | `email` | `mail` | 低（既存の命名を継続使用しているため妥当） |
| ファイル構成 | `sender.go` に全て | `resend.go` + `noop.go` に分割 | 低（分割の方が可読性が高い） |
| 構造体名 | `ResendSender` | `ResendClient` | 低（既存の命名を維持） |
| `SendRaw` メソッド | Workerでの利用のため設計 | 未実装 | 低（Worker内で`Send`メソッドを使用する設計に変更。templ.Componentを直接Workerに渡す方式が採用されており、`SendRaw`は不要になった） |

上記の差異はいずれも設計書を改善する方向の判断であり、問題はない。

## 総合評価

**評価**: Comment

**総評**:

設計書の要件（フェーズ2: Annictの改善）に対して、実装は適切に行われている。主な変更点は以下の通り：

**良かった点**:
- `MailSender` → `Sender`、`SendMultipartEmail` → `Send` へのインターフェース統一が明確
- `NoopSender` の実装が設計書通りで、テストも充実している（`noop_test.go`）
- 3つのワーカーのテンプレート選択ロジックが統一パターン（`xxxTemplates()` 関数）で一貫性がある
- テンプレートのレンダリング責務を `ResendClient.Send` に移動し、ワーカー側はシンプルになった
- ログ出力が `slog` を使用しておりガイドラインに準拠している

**改善が望ましい点**:
- 3つのワーカーテストファイルで `TestXxxWorker_Work` テスト関数が `NoopSender` を活用しておらず、ワーカーの `Work` メソッドが実質テストされていない。`NoopSender` が導入された今、これらのテストを改善する絶好の機会
