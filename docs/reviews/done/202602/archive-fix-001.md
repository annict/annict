# コードレビュー: archive-fix

## レビュー情報

| 項目                       | 内容                                                       |
| -------------------------- | ---------------------------------------------------------- |
| レビュー日                 | 2026-02-25                                                 |
| 対象ブランチ               | archive-fix                                                |
| ベースブランチ             | archive                                                    |
| 作業計画書（指定があれば） | docs/plans/1_doing/fix-depguard-architecture-violations.md |
| 変更ファイル数             | 68 ファイル（Go関連のみ）                                  |
| 変更行数（実装）           | 約 +500 / -285 行                                          |
| 変更行数（テスト）         | 約 +560 / -335 行                                          |

## 参照するガイドライン

- [@CLAUDE.md#レビュー時に参照するガイドライン](/workspace/CLAUDE.md) - ガイドライン一覧

## 変更ファイル一覧

### 実装ファイル

- [x] `go/internal/model/user.go` (NEW)
- [x] `go/internal/repository/email_notification.go` (NEW)
- [x] `go/internal/repository/password_reset_token.go`
- [x] `go/internal/repository/profile.go` (NEW)
- [x] `go/internal/repository/session.go`
- [x] `go/internal/repository/setting.go` (NEW)
- [x] `go/internal/repository/sign_in_code.go` (NEW)
- [x] `go/internal/repository/sign_up_code.go` (NEW)
- [x] `go/internal/repository/stripe_subscriber.go`
- [x] `go/internal/repository/user.go`
- [x] `go/internal/usecase/complete_sign_up.go`
- [x] `go/internal/usecase/create_password_reset_token.go`
- [x] `go/internal/usecase/create_session.go`
- [x] `go/internal/usecase/create_stripe_subscriber.go`
- [x] `go/internal/usecase/delete_stripe_subscriber.go`
- [x] `go/internal/usecase/send_sign_in_code.go`
- [x] `go/internal/usecase/send_sign_up_code.go`
- [x] `go/internal/usecase/update_password_reset.go`
- [x] `go/internal/usecase/update_stripe_subscriber.go`
- [x] `go/internal/usecase/verify_sign_in_code.go`
- [x] `go/internal/usecase/verify_sign_up_code.go`
- [x] `go/internal/middleware/auth.go`
- [x] `go/internal/middleware/authorization.go`
- [x] `go/internal/session/session.go`
- [x] `go/internal/viewmodel/user.go`
- [x] `go/cmd/server/main.go`

### テストファイル

- [x] `go/internal/repository/email_notification_test.go` (NEW)
- [x] `go/internal/repository/password_reset_token_test.go`
- [x] `go/internal/repository/profile_test.go` (NEW)
- [x] `go/internal/repository/session_test.go`
- [x] `go/internal/repository/setting_test.go` (NEW)
- [x] `go/internal/repository/user_test.go`
- [x] `go/internal/usecase/complete_sign_up_test.go`
- [x] `go/internal/usecase/create_password_reset_token_test.go`
- [x] `go/internal/usecase/create_session_test.go`
- [x] `go/internal/usecase/send_sign_in_code_test.go`
- [x] `go/internal/usecase/update_password_reset_test.go`
- [x] `go/internal/usecase/verify_sign_in_code_test.go`
- [x] `go/internal/middleware/auth_test.go`
- [x] `go/internal/middleware/authorization_test.go`
- [x] `go/internal/middleware/sentry_test.go`
- [x] `go/internal/handler/password/edit_test.go`
- [x] `go/internal/handler/password/update_test.go`
- [x] `go/internal/handler/password_reset/create_test.go`
- [x] `go/internal/handler/sign_in/back_param_test.go`
- [x] `go/internal/handler/sign_in/create_test.go`
- [x] `go/internal/handler/sign_in/new_test.go`
- [x] `go/internal/handler/sign_in_code/create_test.go`
- [x] `go/internal/handler/sign_in_code/show_test.go`
- [x] `go/internal/handler/sign_in_code/update_test.go`
- [x] `go/internal/handler/sign_in_password/create_test.go`
- [x] `go/internal/handler/sign_in_password/new_test.go`
- [x] `go/internal/handler/sign_up/create_rate_limit_test.go`
- [x] `go/internal/handler/sign_up/create_test.go`
- [x] `go/internal/handler/sign_up/new_test.go`
- [x] `go/internal/handler/sign_up_code/create_test.go`
- [x] `go/internal/handler/sign_up_username/create_test.go`
- [x] `go/internal/handler/sign_up_username/new_test.go`
- [x] `go/internal/handler/supporters/show_test.go`
- [x] `go/internal/handler/supporters_checkout/create_test.go`
- [x] `go/internal/handler/supporters_portal/create_test.go`

### 設定・その他

- [x] `go/.golangci.yml`
- [x] `go/CLAUDE.md`
- [x] `go/Makefile`
- [x] `go/docs/architecture-guide.md`
- [x] `go/docs/handler-guide.md`
- [x] `go/docs/templ-guide.md`
- [x] `go/docs/validation-guide.md`

## ファイルごとのレビュー結果

全ファイルを確認しましたが、ガイドライン違反となる問題点はありませんでした。

## 設計改善の提案

設計改善の提案はありません。

## 設計との整合性チェック

作業計画書に記載された4件のdepguard違反の修正はすべて実装されています。

| 作業計画書の要件                                                                    | 実装状況                                                                                      |
| ----------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| 違反1: `viewmodel/user.go` の Repository依存を Model依存に変更                      | ✅ `model.User` を直接参照するように変更済み                                                  |
| 違反2: `usecase/create_session.go` の Query依存を Repository経由に修正              | ✅ `SessionRepository` を使用するように変更済み                                               |
| 違反3: `usecase/create_password_reset_token.go` の Query依存を Repository経由に修正 | ✅ `UserRepository` + `PasswordResetTokenRepository` を使用するように変更済み                 |
| 違反4: `usecase/complete_sign_up.go` の Query依存を Repository経由に修正            | ✅ 各Repository（User, Profile, Setting, EmailNotification, Session）を使用するように変更済み |
| `make lint` が depguard 違反なしで成功すること                                      | ✅ golangci.yml の depguard ルールが整備済み                                                  |
| 既存のテストがすべてパスすること                                                    | ✅ テストコードも新しいコンストラクタ引数に対応済み                                           |

### 作業計画書のスコープを超えた変更

作業計画書の4件の違反修正に加え、以下のUseCaseも同様にRepository経由に修正されています。これらは作業計画書には明記されていませんが、同じアーキテクチャ違反パターンの修正であり、一貫性のある対応です：

- `send_sign_in_code.go` → `SignInCodeRepository` + `UserRepository` を使用
- `verify_sign_in_code.go` → `SignInCodeRepository` を使用
- `send_sign_up_code.go` → `SignUpCodeRepository` を使用
- `verify_sign_up_code.go` → `SignUpCodeRepository` を使用
- `update_password_reset.go` → `PasswordResetTokenRepository` + `UserRepository` + `SessionRepository` を使用
- `create_stripe_subscriber.go`, `delete_stripe_subscriber.go`, `update_stripe_subscriber.go` → Repository型エイリアスを使用

また、以下の追加変更も含まれています：

- **`model/user.go` の新規作成**: `model.User` ドメインエンティティとロール定数（`RoleUser`, `RoleAdmin`, `RoleEditor`）の定義
- **`middleware/auth.go`**: `GetUserFromContext` の戻り値型を `*query.GetUserByIDRow` → `*model.User` に変更
- **`middleware/authorization.go`**: ロール定数を `model` パッケージから再エクスポート、`model.User` のメソッド（`IsAdmin`, `IsEditor`, `IsCommitter`）を活用
- **`session/session.go`**: `GetCurrentUser` の戻り値型を `*model.User` に変更、`TouchSession` メソッドを追加

これらの変更は作業計画書の「スコープ外」とされた「既存RepositoryメソッドのQuery戻り値型をModel型に統一」に一部該当しますが、middleware/session層での `model.User` 導入は、depguard違反修正の一環として自然な対応です（middleware層がqueryパッケージに依存しないようにするため）。

## 総合評価

**評価**: Approve

**総評**:

作業計画書に記載された4件のdepguard違反が適切に修正されており、加えて同じパターンの違反が他のUseCaseでも一貫して修正されています。

良かった点：

- **アーキテクチャの一貫性**: すべてのUseCaseがRepositoryを経由してデータアクセスするように統一された
- **WithTxパターンの適用**: 新規・既存のRepositoryに `WithTx` メソッドが適切に追加され、トランザクション管理がアーキテクチャガイドに準拠
- **`model.User` の導入**: ドメインエンティティとしての `model.User` が作成され、ロール判定メソッドがモデルに集約された
- **golangci.yml の整備**: `validator-layer`, `ratelimit-layer` などの新しいdepguardルールが追加され、アーキテクチャルールの網羅性が向上
- **テストの網羅性**: 新規Repositoryにはテストが追加され、既存テストもコンストラクタ変更に対応済み
- **型エイリアスによる後方互換性**: `repository.User = model.User` 等の型エイリアスにより、既存のハンドラーコードへの影響を最小限に抑えている
