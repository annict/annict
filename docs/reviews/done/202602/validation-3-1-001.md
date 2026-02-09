# コードレビュー: validation-3-1

## レビュー情報

| 項目                   | 内容                                                         |
| ---------------------- | ------------------------------------------------------------ |
| レビュー日             | 2026-02-08                                                   |
| 対象ブランチ           | validation-3-1                                               |
| ベースブランチ         | validation                                                   |
| 設計書（指定があれば） | docs/designs/1_doing/validator-consolidation.md               |
| 変更ファイル数         | 4 ファイル                                                   |
| 変更行数（実装）       | +32 / -8 行                                                  |
| 変更行数（テスト）     | +157 / -0 行（旧 request_test.go の削除分は git rename 扱い） |

## 参照するガイドライン

### 共通

- [@CLAUDE.md](/workspace/CLAUDE.md) - プロジェクト全体のガイド（コミットメッセージ、コメント、PR ガイドライン）

### Go 版

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - Go 版の開発ガイド（コーディング規約、テスト戦略）
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTP ハンドラーガイドライン
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド

## 変更ファイル一覧

### 実装ファイル

- [x] `go/internal/handler/sign_in/validator.go`
- [x] `go/internal/handler/sign_in/create.go`

### テストファイル

- [x] `go/internal/handler/sign_in/validator_test.go`

### 設定・その他

- [x] `docs/designs/1_doing/validator-consolidation.md`

## ファイルごとのレビュー結果

### `go/internal/handler/sign_in/validator.go`

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド（構造体の命名規則、Input/Result パターン）

**問題点・改善提案**:

- **[@go/docs/validation-guide.md#構造体の命名規則]**: バリデーターの構造が設計書およびバリデーションガイドの推奨パターンと異なる

  現在の実装では `CreateValidator` がフィールド（`Email`）を直接持つ構造体として定義されており、`Validate` メソッドが `*session.FormErrors` を直接返しています。

  一方、バリデーションガイドの推奨パターンでは以下の構成が示されています：

  - `{Action}Validator` 構造体（依存関係を保持）
  - `{Action}ValidatorInput` 構造体（入力パラメータ）
  - `{Action}ValidatorResult` 構造体（結果を返す）
  - `Validate(ctx context.Context, input {Action}ValidatorInput) *{Action}ValidatorResult`

  ```go
  // 現在の実装
  type CreateValidator struct {
      Email string
  }
  func (v *CreateValidator) Validate(ctx context.Context) *session.FormErrors {
  ```

  ```go
  // バリデーションガイドの推奨パターン
  type CreateValidator struct{}
  type CreateValidatorInput struct {
      Email string
  }
  type CreateValidatorResult struct {
      FormErrors *session.FormErrors
  }
  func (v *CreateValidator) Validate(ctx context.Context, input CreateValidatorInput) *CreateValidatorResult {
  ```

  ただし、現在の sign_in バリデーターは形式バリデーションのみ（DB アクセスなし）で非常にシンプルなため、Input/Result パターンを導入するのはオーバーエンジニアリングになる可能性もあります。また、この実装は旧 `request.go` からのリネームであり、設計書のタスク 3-1 は「`request.go` を `validator.go` にリネーム、`CreateRequest` → `CreateValidator` に変更」と記載されているため、構造の大幅な変更はスコープ外とも解釈できます。

  **対応方針**:

  <!-- 開発者が回答を記入してください -->

  - [ ] 現状維持（リネームのみのスコープとして許容する）
  - [x] Input/Result パターンに変更する（ガイドラインとの完全な整合性を確保）
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  @docs/designs/1_doing/validator-consolidation.md のフェーズ3の他のタスクでもInput/Result パターンを採用したいです。
  ```

## 設計との整合性チェック

### タスク 3-1 の要件確認

設計書のタスク 3-1 に記載された要件：

| 要件                                          | 実装状況 |
| --------------------------------------------- | -------- |
| `request.go` を `validator.go` にリネーム     | ✅ 完了  |
| `CreateRequest` → `CreateValidator` に変更    | ✅ 完了  |
| `request_test.go` を `validator_test.go` にリネーム | ✅ 完了  |
| Handler での参照を更新                        | ✅ 完了  |
| 設計書のチェックボックスを更新                | ✅ 完了  |

すべての要件が満たされています。

## 総合評価

**評価**: Comment

**総評**:

タスク 3-1（sign_in の `request.go` を `validator.go` にリネーム）は設計書の要件通りに正しく実装されています。旧ファイル（`request.go`, `request_test.go`）は適切に削除されており、新ファイル（`validator.go`, `validator_test.go`）への構造体名の変更（`CreateRequest` → `CreateValidator`）も一貫して行われています。`create.go` での参照箇所も正しく更新されています。

1 点確認事項として、バリデーターの構造（Input/Result パターンの不採用）がバリデーションガイドの推奨パターンと異なりますが、設計書のスコープ（リネーム）を考慮すると現状で問題ないと判断できます。開発者の方針をコメントとして確認したい点です。

テストは十分にカバーされており（正常系 1 件、異常系 5 件、I18n メッセージ検証 1 件）、テーブル駆動テストパターンと `t.Parallel()` も適切に使用されています。
