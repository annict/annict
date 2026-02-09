# コードレビュー: validation-fix

## レビュー情報

| 項目                   | 内容                                                                        |
| ---------------------- | --------------------------------------------------------------------------- |
| レビュー日             | 2026-02-09                                                                  |
| 対象ブランチ           | validation-fix                                                              |
| ベースブランチ         | validation                                                                  |
| 設計書（指定があれば） | docs/designs/3_done/202602/validator-consolidation.md                        |
| 変更ファイル数         | 28 ファイル                                                                 |
| 変更行数（実装）       | +402 / -289 行                                                              |
| 変更行数（テスト）     | +476 / -434 行                                                              |

## 参照するガイドライン

### 共通

- [@CLAUDE.md](/workspace/CLAUDE.md) - プロジェクト全体のガイド（コミットメッセージ、コメント、PRガイドライン）

### Go版

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - Go版の開発ガイド（コーディング規約、テスト戦略）
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTPハンドラーガイドライン
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - リクエストバリデーションガイド
- [@go/docs/i18n-guide.md](/workspace/go/docs/i18n-guide.md) - 国際化（I18n）ガイド
- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - セキュリティガイドライン

## 変更ファイル一覧

### 実装ファイル

- [x] `go/internal/handler/password/validator.go` (新規)
- [x] `go/internal/handler/password/update.go`
- [x] `go/internal/handler/password/request.go` (削除)
- [x] `go/internal/handler/password_reset/validator.go` (新規)
- [x] `go/internal/handler/password_reset/create.go`
- [x] `go/internal/handler/password_reset/request.go` (削除)
- [x] `go/internal/handler/sign_in_password/validator.go` (新規)
- [x] `go/internal/handler/sign_in_password/create.go`
- [x] `go/internal/handler/sign_in_password/request.go` (削除)
- [x] `go/internal/handler/sign_up/validator.go` (新規)
- [x] `go/internal/handler/sign_up/create.go`
- [x] `go/internal/handler/sign_up/request.go` (削除)
- [x] `go/internal/handler/sign_up_code/validator.go` (新規)
- [x] `go/internal/handler/sign_up_code/create.go`
- [x] `go/internal/handler/sign_up_code/request.go` (削除)
- [x] `go/internal/handler/sign_up_username/validator.go` (新規)
- [x] `go/internal/handler/sign_up_username/create.go`
- [x] `go/internal/handler/sign_up_username/request.go` (削除)
- [x] `go/internal/handler/supporters_checkout/validator.go` (新規)
- [x] `go/internal/handler/supporters_checkout/create.go`
- [x] `go/internal/handler/supporters_checkout/request.go` (削除)

### テストファイル

- [x] `go/internal/handler/password/validator_test.go` (新規)
- [x] `go/internal/handler/password/request_test.go` (削除)
- [x] `go/internal/handler/password_reset/validator_test.go` (リネーム)
- [x] `go/internal/handler/sign_in_password/validator_test.go` (リネーム)
- [x] `go/internal/handler/supporters_checkout/validator_test.go` (新規)
- [x] `go/internal/handler/supporters_checkout/request_test.go` (削除)

### 設定・その他

- [x] `docs/designs/3_done/202602/validator-consolidation.md` (設計書の軽微な修正)

## ファイルごとのレビュー結果

### `go/internal/handler/supporters_checkout/validator.go`

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド
- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - コーディング規約（コメント）

**問題点・改善提案**:

- **[@go/docs/validation-guide.md#バリデーションの分類]**: 旧 `CreateRequest.Validate()` は `context.Context` を受け取らず `map[string]string` を返していたが、新しい `CreateValidator.Validate()` は `context.Context` を受け取り `i18n.T(ctx, ...)` で国際化メッセージを使用するようになった。これは設計書の方針（Input/Result パターン + I18n対応）に合致しているが、**旧実装ではバリデーションエラーのキーとして `"invalid_plan"` というI18nキーではない文字列を使用していた**のに対し、新実装では `i18n.T(ctx, "supporters_checkout_invalid_plan")` を使用している。この I18n キー `"supporters_checkout_invalid_plan"` が翻訳ファイル（`ja.toml` / `en.toml`）に定義されているか確認が必要。

  ```go
  // 新しいコード
  formErrors.AddFieldError("plan", i18n.T(ctx, "supporters_checkout_invalid_plan"))
  ```

  **対応方針**:

  <!-- 開発者が回答を記入してください -->

  - [x] 翻訳ファイルに `supporters_checkout_invalid_plan` が既に定義済みであれば問題なし
  - [ ] 未定義の場合は翻訳ファイルに追加する
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

## 設計との整合性チェック

### 設計書の要件チェック

| 要件 | 状態 | 備考 |
|------|------|------|
| `request.go` を `validator.go` にリネーム | ✅ 完了 | 全9ハンドラーで対応済み |
| `{Action}Request` → `{Action}Validator` に変更 | ✅ 完了 | 全ハンドラーで統一済み |
| `{Action}ValidatorInput` 構造体を追加 | ✅ 完了 | 全ハンドラーで追加済み |
| `{Action}ValidatorResult` 構造体を追加 | ✅ 完了 | 全ハンドラーで追加済み |
| `NewXxxValidator()` コンストラクタを追加 | ✅ 完了 | 全ハンドラーで追加済み |
| テストファイルを `validator_test.go` にリネーム | ✅ 完了 | テストがあるハンドラーで対応済み |
| Handler での参照を更新 | ✅ 完了 | 全 `create.go` / `update.go` で更新済み |

### フェーズ3タスクの完了確認

| タスク | 状態 | 備考 |
|--------|------|------|
| 3-1: sign_in | ❓ | validation ブランチで対応済み（このPRの範囲外） |
| 3-2: sign_in_code | ❓ | validation ブランチで対応済み（このPRの範囲外） |
| 3-3: sign_in_password | ✅ | このPRで対応 |
| 3-4: sign_up | ✅ | このPRで対応 |
| 3-5: sign_up_code | ✅ | このPRで対応 |
| 3-6: sign_up_username | ✅ | このPRで対応 |
| 3-7: password | ✅ | このPRで対応 |
| 3-8: password_reset | ✅ | このPRで対応 |
| 3-9: supporters_checkout | ✅ | このPRで対応 |

### 設計書との整合性総評

設計書の方針通りに実装されている。主な変更点：

1. **ファイルリネーム**: `request.go` → `validator.go`（全ハンドラーで完了）
2. **構造体リネーム**: `{Action}Request` → `{Action}Validator` + `{Action}ValidatorInput` + `{Action}ValidatorResult`
3. **パターンの統一**: 全てのバリデーターが `NewXxxValidator()` コンストラクタ + `Validate(ctx, input) *Result` パターンを使用
4. **ハンドラーの更新**: `req.Validate()` → `v.Validate(ctx, input)` に変更、`result` 変数名を `ucResult` にリネームして衝突を回避

改善点として、`supporters_checkout` の旧実装は `map[string]string` を返す独自パターンだったが、新実装では他のバリデーターと統一された `FormErrors` パターン + I18n対応に変更されている。これは設計書の意図に沿った良い改善。

## 総合評価

**評価**: Comment

**総評**:

設計書の方針（フェーズ3: Annictリファクタリング）に沿って、全9ハンドラーの `request.go` → `validator.go` リネームと Input/Result パターンへの移行が正しく実装されている。

**良い点**:

- 全ハンドラーで一貫した命名規則（`{Action}Validator`, `{Action}ValidatorInput`, `{Action}ValidatorResult`）が適用されている
- テストも新しいパターンに合わせて適切に更新されている
- `supporters_checkout` は独自の `map[string]string` 返却パターンから、標準の `FormErrors` パターンに改善された
- `sign_up_code` では `regexp.MustCompile` がパッケージレベル変数に移動されており、ベストプラクティスに従っている
- ハンドラー側で `result` → `ucResult` のリネームにより変数名の衝突を適切に回避している

**確認が必要な点**:

- `supporters_checkout` の新しいI18nキー `supporters_checkout_invalid_plan` が翻訳ファイルに定義されているかの確認（1件）
