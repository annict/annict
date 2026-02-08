# バリデーターファイル統合 設計書

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

### Go 版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - 全体的なコーディング規約
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTP ハンドラーガイドライン（**ファイル名は標準の 8 種類のみ**）
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

Handler におけるバリデーション処理のファイル構成をシンプル化し、各プロジェクト（Mewst、Wikino、Annict）で一貫性のある設計に統一するリファクタリング作業です。

**目的**:

- バリデーション関連のファイルを 1 ファイル（`validator.go`）に統合し、「どこに書くべきか」の判断コストを削減
- 構造体名を `{Action}Validator` に統一し、命名規則をシンプルにする
- 3 つのプロジェクトで共通の設計方針を適用する

**背景**:

- 現状、Wikino では `format_validator.go` と `state_validator.go` に分離しているが、分離のメリットが明確でない
- Annict では `request.go` という別の命名規則を使用している
- 形式バリデーションと状態バリデーションの分類は、その後の処理が内容によって異なるため、厳密に分ける必要性が低い
- YAGNI 原則に従い、必要になったときにファイル分割すればよい

## 要件

<!--
ガイドライン:
- 機能要件: 「何ができるべきか」を記述
- 非機能要件: 「どのように動くべきか」を必要に応じて記述
-->

### 機能要件

<!--
「ユーザーは〇〇できる」「システムは〇〇する」という形式で記述
箇条書きで簡潔に
-->

- バリデーション関連のコードが `validator.go` に統一され、どこに書くべきか迷わない
- 構造体名が `{Action}Validator` に統一され、命名規則が明確になる
- 3 つのプロジェクト（Mewst、Wikino、Annict）で一貫した設計が適用される

### 非機能要件

<!--
必要に応じて以下のような項目を追加してください：
- セキュリティ（認証、認可、暗号化、監査ログなど）
- パフォーマンス（応答時間、スループット、リソース使用量など）
- ユーザビリティ（UX）（使いやすさ、わかりやすさ、アクセシビリティなど）
- 可用性・信頼性（稼働率、障害時の挙動、エラーハンドリングなど）
- 保守性（テストのしやすさ、コードの読みやすさ、ドキュメントなど）

不要な場合はこのセクション全体を削除してください。
-->

- **保守性**: バリデーションロジックの配置場所が明確になり、将来の機能追加時に迷わない
- **シンプルさ**: ファイル数が減り、ナビゲーションが楽になる

## 設計

<!--
ガイドライン:
- 技術的な実装の詳細を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - 技術スタック（使用するライブラリ、フレームワーク、ツールなど）
  - アーキテクチャ（システム全体の構成、コンポーネント間の関係など）
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - セキュリティ設計（認証・認可、トークン管理、Rate Limitingなど）
  - コード設計（パッケージ構成、主要な構造体、インターフェースなど）
  - テスト戦略（単体テスト、統合テスト、E2Eテストの方針）
  - マイグレーション管理（データベースマイグレーションの方針）
  - 実装方針（特記事項、既存システムとの関係、制約など）

不要な場合はこのセクション全体を削除してください。
-->

### 新しいバリデーター構成

#### ファイル構成

| 変更前 | 変更後 |
|--------|--------|
| `format_validator.go` + `state_validator.go` | `validator.go` |
| `request.go` | `validator.go` |

#### 構造体命名規則

| 変更前 | 変更後 |
|--------|--------|
| `{Action}FormatValidator` | `{Action}Validator` |
| `{Action}StateValidator` | `{Action}Validator` に統合 |
| `{Action}Request` | `{Action}Validator` |

#### validator.go の構成例

```go
package sign_in

import (
    "context"
    "errors"
    "net/mail"

    "github.com/example/app/internal/i18n"
    "github.com/example/app/internal/model"
    "github.com/example/app/internal/repository"
    "github.com/example/app/internal/session"
)

// バリデーションのエラー定義
var (
    ErrUserNotFound    = errors.New("ユーザーが見つかりません")
    ErrInvalidPassword = errors.New("パスワードが正しくありません")
)

// CreateValidator はサインインのバリデーションを行う
type CreateValidator struct {
    userRepo         *repository.UserRepository
    userPasswordRepo *repository.UserPasswordRepository
}

// NewCreateValidator は CreateValidator を生成する
func NewCreateValidator(
    userRepo *repository.UserRepository,
    userPasswordRepo *repository.UserPasswordRepository,
) *CreateValidator {
    return &CreateValidator{
        userRepo:         userRepo,
        userPasswordRepo: userPasswordRepo,
    }
}

// CreateValidatorInput はバリデーションの入力パラメータ
type CreateValidatorInput struct {
    Email    string
    Password string
}

// CreateValidatorResult はバリデーションの結果
type CreateValidatorResult struct {
    User       *model.User
    FormErrors *session.FormErrors
    Err        error
}

// Validate はバリデーションを行う
func (v *CreateValidator) Validate(ctx context.Context, input CreateValidatorInput) *CreateValidatorResult {
    // 1. 形式バリデーション
    formErrors := session.NewFormErrors()

    if input.Email == "" {
        formErrors.AddField("email", i18n.T(ctx, "validation_required"))
    } else if !isValidEmail(input.Email) {
        formErrors.AddField("email", i18n.T(ctx, "validation_email_invalid"))
    }

    if input.Password == "" {
        formErrors.AddField("password", i18n.T(ctx, "validation_required"))
    }

    if formErrors.HasErrors() {
        return &CreateValidatorResult{FormErrors: formErrors}
    }

    // 2. 状態バリデーション（DB検証）
    user, err := v.userRepo.FindByEmail(ctx, input.Email)
    if err != nil {
        return &CreateValidatorResult{Err: err}
    }
    if user == nil {
        formErrors.AddGlobal(i18n.T(ctx, "validation_email_or_password_invalid"))
        return &CreateValidatorResult{FormErrors: formErrors, Err: ErrUserNotFound}
    }

    // パスワード検証...

    return &CreateValidatorResult{User: user}
}

func isValidEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}
```

### 現状分析

#### Mewst (`/mewst/go/`)

**現状**:

| ハンドラー | 現在のファイル構成 |
|------------|-------------------|
| sign_in | format_validator.go, state_validator.go |
| email_confirmation | format_validator.go, state_validator.go |
| password | format_validator.go のみ |
| password_reset | format_validator.go のみ |

**修正方針**:

- `format_validator.go` と `state_validator.go` を `validator.go` に統合
- `format_validator.go` のみのハンドラーは `validator.go` にリネーム
- 構造体名を `{Action}Validator` に変更
- テストファイルも `validator_test.go` に統合/リネーム

#### Wikino (`/workspace/go/`)

**現状**:

| ハンドラー | 現在のファイル構成 |
|------------|-------------------|
| sign_in | format_validator.go, state_validator.go |
| sign_in_two_factor | format_validator.go, state_validator.go |
| sign_in_two_factor_recovery | format_validator.go, state_validator.go |
| email_confirmation | format_validator.go, state_validator.go |
| account | format_validator.go, state_validator.go |

**修正方針**:

- `format_validator.go` と `state_validator.go` を `validator.go` に統合
- 構造体名を `{Action}FormatValidator` + `{Action}StateValidator` → `{Action}Validator` に変更
- テストファイルも `validator_test.go` に統合

#### Annict (`/annict/go/`)

**現状**:

| ハンドラー | 現在のファイル構成 |
|------------|-------------------|
| sign_in | request.go |
| sign_in_code | request.go |
| sign_in_password | request.go |
| sign_up | request.go |
| sign_up_code | request.go |
| sign_up_username | request.go |
| password | request.go |
| password_reset | request.go |
| supporters_checkout | request.go |

**修正方針**:

- `request.go` を `validator.go` にリネーム
- 構造体名を `{Action}Request` → `{Action}Validator` に変更
- テストファイルも `validator_test.go` にリネーム

## タスクリスト

<!--
ガイドライン:
- フェーズごとに段階的な実装計画を記述
- チェックボックスで進捗を管理
- **重要**: 1タスク = 1 Pull Request の粒度で作成してください
- **重要**: 各タスクには想定ファイル数と想定行数を明記してください（PRサイズの見積もりのため）
- 想定ファイル数は「実装」と「テスト」に分けて記載してください
- 想定行数も「実装」と「テスト」に分けて記載してください
- 依存関係を明確に
- Pull Requestのガイドラインは CLAUDE.md を参照（変更ファイル数20以下、変更行数300行以下）

タスク番号の付け方:
- 各タスクには階層的な番号を付与します（例: 1-1, 1-2, 2-1, 2-2）
- フォーマット: **フェーズ番号-タスク番号**: タスク名
- **フェーズ番号は半角英数字とハイフンのみで表記**してください（ブランチ名に使用するため）
  - 例: フェーズ 1, フェーズ 2, フェーズ 5a（フェーズ 5 と 6 の間に追加する場合）
  - NG: フェーズ 5.5（ドットは使用不可）
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）

プラットフォームプレフィックス:
- Go版またはRails版の修正を行うタスクには、タスク名の先頭にプラットフォームを示すプレフィックスを付けてください
- フォーマット: **フェーズ番号-タスク番号**: [Go] タスク名 または **フェーズ番号-タスク番号**: [Rails] タスク名
- Go版とRails版の両方を修正する場合は、別々のタスクに分けてください
- 例:
  - `- [ ] **1-1**: [Go] マイグレーション作成`
  - `- [ ] **1-2**: [Rails] モデルへのコールバック追加`
-->

### フェーズ 1: ドキュメント整備

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
Go版/Rails版の両方を修正する場合は別タスクに分けてください
-->

- [x] **1-1**: [Go/Wikino] validation-guide.md と handler-guide.md の更新

  - バリデーション関連ファイルを `validator.go` に統一する方針に変更
  - 構造体名を `{Action}Validator` に統一
  - handler-guide.md の標準ファイル名を 10 種類から 9 種類に変更（`format_validator.go` と `state_validator.go` を `validator.go` に統合）
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 200 行（実装 200 行 + テスト 0 行）

### フェーズ 2: Wikino リファクタリング

- [x] **2-1**: [Go/Wikino] sign_in の validator.go 統合

  - `format_validator.go` と `state_validator.go` を `validator.go` に統合
  - `CreateFormatValidator` + `CreateStateValidator` → `CreateValidator` に変更
  - `format_validator_test.go` と `state_validator_test.go` を `validator_test.go` に統合
  - Handler の依存性を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [x] **2-2**: [Go/Wikino] sign_in_two_factor の validator.go 統合

  - `format_validator.go` と `state_validator.go` を `validator.go` に統合
  - `CreateFormatValidator` + `CreateStateValidator` → `CreateValidator` に変更
  - テストファイルも統合
  - Handler の依存性を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [x] **2-3**: [Go/Wikino] sign_in_two_factor_recovery の validator.go 統合

  - `format_validator.go` と `state_validator.go` を `validator.go` に統合
  - `CreateFormatValidator` + `CreateStateValidator` → `CreateValidator` に変更
  - テストファイルも統合
  - Handler の依存性を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [x] **2-4**: [Go/Wikino] email_confirmation の validator.go 統合

  - `format_validator.go` と `state_validator.go` を `validator.go` に統合
  - `CreateFormatValidator` + `CreateStateValidator` → `CreateValidator` に変更
  - `UpdateFormatValidator` + `UpdateStateValidator` → `UpdateValidator` に変更
  - テストファイルも統合
  - Handler の依存性を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 250 行（実装 120 行 + テスト 130 行）

- [x] **2-5**: [Go/Wikino] account の validator.go 統合

  - `format_validator.go` と `state_validator.go` を `validator.go` に統合
  - `CreateFormatValidator` + `CreateStateValidator` → `CreateValidator` に変更
  - テストファイルも統合
  - Handler の依存性を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

### フェーズ 3: Annict リファクタリング

- [x] **3-1**: [Go/Annict] sign_in の request.go を validator.go にリネーム

  - `request.go` を `validator.go` にリネーム
  - `CreateRequest` → `CreateValidator` + `CreateValidatorInput` + `CreateValidatorResult` に変更（Input/Result パターン）
  - `request_test.go` を `validator_test.go` にリネームし、Input/Result パターンに合わせて更新
  - Handler での参照を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [ ] **3-2**: [Go/Annict] sign_in_code の request.go を validator.go にリネーム

  - `request.go` を `validator.go` にリネーム
  - `{Action}Request` → `{Action}Validator` + `{Action}ValidatorInput` + `{Action}ValidatorResult` に変更（Input/Result パターン）
  - テストファイルもリネームし、Input/Result パターンに合わせて更新
  - Handler での参照を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [ ] **3-3**: [Go/Annict] sign_in_password の request.go を validator.go にリネーム

  - `request.go` を `validator.go` にリネーム
  - `{Action}Request` → `{Action}Validator` + `{Action}ValidatorInput` + `{Action}ValidatorResult` に変更（Input/Result パターン）
  - テストファイルもリネームし、Input/Result パターンに合わせて更新
  - Handler での参照を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [ ] **3-4**: [Go/Annict] sign_up の request.go を validator.go にリネーム

  - `request.go` を `validator.go` にリネーム
  - `{Action}Request` → `{Action}Validator` + `{Action}ValidatorInput` + `{Action}ValidatorResult` に変更（Input/Result パターン）
  - テストファイルもリネームし、Input/Result パターンに合わせて更新
  - Handler での参照を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [ ] **3-5**: [Go/Annict] sign_up_code の request.go を validator.go にリネーム

  - `request.go` を `validator.go` にリネーム
  - `{Action}Request` → `{Action}Validator` + `{Action}ValidatorInput` + `{Action}ValidatorResult` に変更（Input/Result パターン）
  - Handler での参照を更新
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

- [ ] **3-6**: [Go/Annict] sign_up_username の request.go を validator.go にリネーム

  - `request.go` を `validator.go` にリネーム
  - `{Action}Request` → `{Action}Validator` + `{Action}ValidatorInput` + `{Action}ValidatorResult` に変更（Input/Result パターン）
  - Handler での参照を更新
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

- [ ] **3-7**: [Go/Annict] password の request.go を validator.go にリネーム

  - `request.go` を `validator.go` にリネーム
  - `{Action}Request` → `{Action}Validator` + `{Action}ValidatorInput` + `{Action}ValidatorResult` に変更（Input/Result パターン）
  - テストファイルもリネームし、Input/Result パターンに合わせて更新
  - Handler での参照を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [ ] **3-8**: [Go/Annict] password_reset の request.go を validator.go にリネーム

  - `request.go` を `validator.go` にリネーム
  - `{Action}Request` → `{Action}Validator` + `{Action}ValidatorInput` + `{Action}ValidatorResult` に変更（Input/Result パターン）
  - テストファイルもリネームし、Input/Result パターンに合わせて更新
  - Handler での参照を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [ ] **3-9**: [Go/Annict] supporters_checkout の request.go を validator.go にリネーム

  - `request.go` を `validator.go` にリネーム
  - `{Action}Request` → `{Action}Validator` + `{Action}ValidatorInput` + `{Action}ValidatorResult` に変更（Input/Result パターン）
  - テストファイルもリネームし、Input/Result パターンに合わせて更新
  - Handler での参照を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

### フェーズ 4: Mewst リファクタリング

- [x] **4-1**: [Go/Mewst] sign_in の validator.go 統合

  - `format_validator.go` と `state_validator.go` を `validator.go` に統合
  - `CreateFormatValidator` + `CreateStateValidator` → `CreateValidator` に変更
  - `format_validator_test.go` と `state_validator_test.go` を `validator_test.go` に統合
  - Handler の依存性を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [x] **4-2**: [Go/Mewst] email_confirmation の validator.go 統合

  - `format_validator.go` と `state_validator.go` を `validator.go` に統合
  - `CreateFormatValidator` + `CreateStateValidator` → `CreateValidator` に変更
  - テストファイルも統合
  - Handler の依存性を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [x] **4-3**: [Go/Mewst] password の format_validator.go を validator.go にリネーム

  - `format_validator.go` を `validator.go` にリネーム
  - `UpdateFormatValidator` → `UpdateValidator` に変更
  - `format_validator_test.go` を `validator_test.go` にリネーム
  - Handler での参照を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 50 行（実装 25 行 + テスト 25 行）

- [x] **4-4**: [Go/Mewst] password_reset の format_validator.go を validator.go にリネーム

  - `format_validator.go` を `validator.go` にリネーム
  - `CreateFormatValidator` → `CreateValidator` に変更
  - `format_validator_test.go` を `validator_test.go` にリネーム
  - Handler での参照を更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 50 行（実装 25 行 + テスト 25 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **depguard ルールの追加**: バリデーターの依存関係を静的に検証する仕組みは、今回は導入しない（必要性が低い）
- **バリデーションライブラリの導入**: 現状の手動バリデーションで十分対応できるため
- **共通バリデーション関数の抽出**: プロジェクト間でバリデーションロジックを共有する仕組みは、今後の課題とする

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Wikino validation-guide.md](/workspace/go/docs/validation-guide.md) - 現在のバリデーションガイド
- [Wikino handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTP ハンドラーガイドライン
- [validation-responsibility-refactoring.md](/workspace/docs/designs/1_doing/validation-responsibility-refactoring.md) - 旧設計書（形式/状態バリデーションの分離）
