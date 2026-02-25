# コードレビュー: validation-fix

## レビュー情報

| 項目                   | 内容                                       |
| ---------------------- | ------------------------------------------ |
| レビュー日             | 2026-02-09                                 |
| 対象ブランチ           | validation-fix                             |
| ベースブランチ         | validation                                 |
| 設計書（指定があれば） | docs/designs/1_doing/worker-unification.md |
| 変更ファイル数         | 50 ファイル                                |
| 変更行数（実装）       | +723 / -694 行                             |
| 変更行数（テスト）     | +805 / -939 行                             |

## 参照するガイドライン

### 共通

- [@CLAUDE.md](/workspace/CLAUDE.md) - プロジェクト全体のガイド（コミットメッセージ、コメント、PRガイドライン）

### Go版

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - Go版の開発ガイド（コーディング規約、テスト戦略）
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTPハンドラーガイドライン
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - リクエストバリデーションガイド
- [@go/docs/i18n-guide.md](/workspace/go/docs/i18n-guide.md) - 国際化（I18n）ガイド

### 設計書

- [docs/designs/1_doing/worker-unification.md](/workspace/docs/designs/1_doing/worker-unification.md) - Worker実装統一 設計書

## 変更ファイル一覧

### 実装ファイル

**バリデーター統合（request.go → validator.go）**:

- [x] `go/internal/handler/password/validator.go`（新規）
- [x] `go/internal/handler/password_reset/validator.go`（新規）
- [x] `go/internal/handler/sign_in_password/validator.go`（新規）
- [x] `go/internal/handler/sign_up/validator.go`（新規）
- [x] `go/internal/handler/sign_up_code/validator.go`（新規）
- [x] `go/internal/handler/sign_up_username/validator.go`（新規）
- [x] `go/internal/handler/supporters_checkout/validator.go`（新規）
- [x] `go/internal/handler/password/request.go`（削除）
- [x] `go/internal/handler/password_reset/request.go`（削除）
- [x] `go/internal/handler/sign_in_password/request.go`（削除）
- [x] `go/internal/handler/sign_up/request.go`（削除）
- [x] `go/internal/handler/sign_up_code/request.go`（削除）
- [x] `go/internal/handler/sign_up_username/request.go`（削除）
- [x] `go/internal/handler/supporters_checkout/request.go`（削除）

**ハンドラー更新（バリデーター呼び出し変更）**:

- [x] `go/internal/handler/password/update.go`
- [x] `go/internal/handler/password_reset/create.go`
- [x] `go/internal/handler/sign_in_password/create.go`
- [x] `go/internal/handler/sign_up/create.go`
- [x] `go/internal/handler/sign_up_code/create.go`
- [x] `go/internal/handler/sign_up_username/create.go`
- [x] `go/internal/handler/supporters_checkout/create.go`

**Worker統一**:

- [x] `go/internal/worker/send_email.go`（新規）
- [x] `go/internal/worker/client.go`
- [x] `go/internal/worker/cleanup_expired_tokens.go`
- [x] `go/internal/worker/cleanup_expired_sign_in_codes.go`
- [x] `go/internal/worker/send_password_reset_email.go`（削除）
- [x] `go/internal/worker/send_sign_in_code.go`（削除）
- [x] `go/internal/worker/send_sign_up_code.go`（削除）

**メールインターフェース拡張**:

- [x] `go/internal/mail/resend.go`
- [x] `go/internal/mail/noop.go`

**ユースケース変更**:

- [x] `go/internal/usecase/create_password_reset_token.go`
- [x] `go/internal/usecase/send_sign_in_code.go`
- [x] `go/internal/usecase/send_sign_up_code.go`

**エントリーポイント**:

- [x] `go/cmd/server/main.go`

### テストファイル

- [x] `go/internal/handler/password/validator_test.go`（新規、request_test.go を置換）
- [x] `go/internal/handler/password_reset/validator_test.go`（リネーム + 更新）
- [x] `go/internal/handler/password_reset/create_test.go`
- [x] `go/internal/handler/sign_in_password/validator_test.go`（リネーム + 更新）
- [x] `go/internal/handler/supporters_checkout/validator_test.go`（新規、request_test.go を置換）
- [x] `go/internal/handler/supporters_checkout/request_test.go`（削除）
- [x] `go/internal/handler/password/request_test.go`（削除）
- [x] `go/internal/worker/send_email_test.go`（新規）
- [x] `go/internal/worker/insert_opts_test.go`（新規）
- [x] `go/internal/worker/send_password_reset_email_test.go`（削除）
- [x] `go/internal/worker/send_sign_in_code_test.go`（削除）
- [x] `go/internal/worker/send_sign_up_code_test.go`（削除）
- [x] `go/internal/usecase/create_password_reset_token_test.go`

### 設定・その他

- [x] `docs/designs/1_doing/worker-unification.md`
- [x] `docs/designs/{1_doing → done/202602}/validator-consolidation.md`
- [x] `docs/reviews/done/202602/validation-fix-202602090433.md`

## ファイルごとのレビュー結果

問題のあるファイルはありません。すべてのファイルがガイドラインに適合しています。

## 設計書との適合性

### バリデーター統合

設計書 `validator-consolidation.md` の要件との適合性を確認しました：

| 要件                                     | 適合 | 備考                                      |
| ---------------------------------------- | ---- | ----------------------------------------- |
| `request.go` → `validator.go` にリネーム | ✅   | 7ファイルすべて対応済み                   |
| `{Action}Validator` 命名規則             | ✅   | 全バリデーターが準拠                      |
| `{Action}ValidatorInput` 入力構造体      | ✅   | 全バリデーターが準拠                      |
| `{Action}ValidatorResult` 結果構造体     | ✅   | 全バリデーターが準拠                      |
| `New{Action}Validator` コンストラクタ    | ✅   | 全バリデーターが準拠                      |
| i18n対応                                 | ✅   | 全エラーメッセージが `i18n.T()` を使用    |
| 正規表現のパッケージレベル定義           | ✅   | `sign_up_code`, `sign_up_username` で準拠 |
| ハンドラーの呼び出しパターン統一         | ✅   | 全ハンドラーが統一パターンを使用          |

### Worker実装統一（設計書: worker-unification.md）

| 要件                                   | 適合 | 備考                                                                                |
| -------------------------------------- | ---- | ----------------------------------------------------------------------------------- |
| タスク 2-1: InsertOptsメソッド追加     | ✅   | `SendEmailArgs`, `CleanupExpiredTokensArgs`, `CleanupExpiredSignInCodesArgs` に追加 |
| タスク 2-2: 事前レンダリング方式に変更 | ✅   | `BuildXxxEmail` 関数でテンプレートを事前レンダリング                                |
| 統一されたSendEmailArgs                | ✅   | `To`, `Subject`, `HTMLBody`, `TextBody` フィールドの統一構造体                      |
| ログ出力（開始・完了・エラー）         | ✅   | `slog.InfoContext` / `slog.ErrorContext` を使用                                     |
| DBドライバー: pgx/v5 + riverpgxv5      | ✅   | 既存のclient.goで対応済み                                                           |
| ワーカー登録: client.go内部            | ✅   | client.goで`SendEmailWorker`を登録                                                  |
| コネクションプール設定                 | ✅   | 既存のclient.goで対応済み                                                           |
| `mail.SendRaw` インターフェース        | ✅   | `Sender` インターフェースに `SendRaw` メソッドを追加                                |

## 総合評価

**評価**: Approve

**総評**:

2つの大きな変更（バリデーター統合とWorker実装統一）がきれいに実装されています。

**良かった点**:

1. **バリデーターの統一性**: 7つのバリデーターすべてが `{Action}Validator` / `{Action}ValidatorInput` / `{Action}ValidatorResult` の命名規則に完全準拠しており、バリデーションガイドとの整合性が高い
2. **Worker統一の設計書適合性**: 設計書で定義された「事前レンダリング方式」「InsertOptsメソッド」「統一されたSendEmailArgs」のすべてが正確に実装されている
3. **i18n対応の徹底**: `supporters_checkout/validator.go` の新規i18nキー `supporters_checkout_invalid_plan` も翻訳ファイルに定義済み
4. **テストの網羅性**: バリデーターテスト、ワーカーテスト、InsertOptsテストがすべて揃っている
5. **コード量の削減**: Worker統合により3つの個別ワーカーファイル（計350行）が1つの統一ファイル（200行）に集約され、コード重複が解消されている
6. **slogの一貫した使用**: ログ出力がすべて `slog.InfoContext` / `slog.ErrorContext` パターンに準拠している
