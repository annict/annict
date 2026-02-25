# バリデーション責務のリファクタリング 設計書

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

Handler、UseCase、Request DTO におけるバリデーション処理の責務を明確化し、各プロジェクト（Mewst、Annict、Wikino）で一貫性のある設計に統一するリファクタリング作業です。

**目的**:

- UseCase をトランザクションを伴う永続化処理に特化させ、処理の流れをわかりやすくする
- バリデーションの責務を明確にし、エラーハンドリングを一貫させる
- 3 つのプロジェクトで共通の設計方針を適用する

**背景**:

- 現状、UseCase にビジネスルール検証が含まれている箇所があり、責務が曖昧になっている
- プロジェクト間、さらには同一プロジェクト内でも設計に一貫性がない
- Rails 版では Form オブジェクトで入力値検証を行い、UseCase を呼ぶという明確な流れがあった

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

- バリデーションの責務が明確に分離され、各層の役割がわかりやすくなる
- エラー発生時に適切な表示方法（フィールドエラー、グローバルエラー、Flash）が選択される
- 3 つのプロジェクト（Mewst、Annict、Wikino）で一貫した設計が適用される

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
- **テスト容易性**: 各層の責務が明確になることで、テストの範囲が明確になる

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

### バリデーションの分類と責務

バリデーションを 3 つのカテゴリに分類し、それぞれの責務と表示方法を明確にします。

#### 1. 形式バリデーション（format_validator.go）

| 項目           | 内容                                                                              |
| -------------- | --------------------------------------------------------------------------------- |
| **責務**       | 入力値の形式チェック                                                              |
| **場所**       | `internal/handler/{resource}/format_validator.go`                                 |
| **例**         | 必須チェック、メール形式、文字数制限、正規表現                                    |
| **特徴**       | DB アクセス不要                                                                   |
| **構造体名**   | `{Action}FormatValidator`（例: `CreateFormatValidator`, `UpdateFormatValidator`） |
| **エラー表示** | フォーム内のフィールドエラー（`FormErrors.AddFieldError()`）                      |

#### 2. 状態バリデーション（state_validator.go または UseCase）

| 項目           | 内容                                                            |
| -------------- | --------------------------------------------------------------- |
| **責務**       | DB の状態を使った検証                                           |
| **場所**       | `internal/handler/{resource}/state_validator.go` または UseCase |
| **例**         | ユーザー存在チェック、メールアドレス重複、パスワード照合        |
| **特徴**       | DB アクセス必要、Repository への依存あり                        |
| **構造体名**   | `{Action}StateValidator`（例: `CreateStateValidator`）          |
| **エラー表示** | グローバルエラー（`FormErrors.AddGlobalError()`）               |

**配置場所の判断基準**: **「検証失敗時に DB を更新する必要があるか？」**

| 検証失敗時の DB 更新 | 配置場所        | 理由                                               |
| -------------------- | --------------- | -------------------------------------------------- |
| 不要                 | state_validator | UseCase をシンプルに保つため                       |
| 必要                 | UseCase         | トランザクション内で検証と更新を行う必要があるため |

**state_validator.go で行うべき検証**:

| 検証内容                   | 失敗時の動作     | 理由                   |
| -------------------------- | ---------------- | ---------------------- |
| ユーザー存在チェック       | エラーメッセージ | DB 更新なし            |
| メールアドレス重複チェック | エラーメッセージ | DB 更新なし            |
| アットネーム重複チェック   | エラーメッセージ | DB 更新なし            |
| メール確認完了チェック     | エラーメッセージ | DB 更新なし            |
| コード一致チェック         | エラーメッセージ | DB 更新なし（※注参照） |
| パスワード照合             | エラーメッセージ | DB 更新なし            |

※注: コード検証で「試行回数インクリメント」が必要な場合は UseCase で行う

**UseCase で行うべき検証**:

| 検証内容           | 失敗時の動作           | 理由                   |
| ------------------ | ---------------------- | ---------------------- |
| ログインコード検証 | 試行回数インクリメント | 失敗時に DB 更新が必要 |

※注: 「リカバリーコード消費」や「トークン使用済みマーク」は検証成功後の処理であり、バリデーションではなく UseCase の永続化処理として扱う。検証自体は state_validator で行い、成功後に UseCase を呼び出す。

#### 3. システムエラー（全層）

| 項目           | 内容                                                   |
| -------------- | ------------------------------------------------------ |
| **責務**       | 永続化処理中の予期せぬエラー                           |
| **場所**       | UseCase、Repository                                    |
| **例**         | DB 接続エラー、トランザクション失敗                    |
| **特徴**       | ユーザーが対処困難                                     |
| **エラー表示** | Flash メッセージまたはエラーページ、ログには詳細を記録 |

### エラー表示方法の使い分け

| エラー種類           | 表示方法            | 使い分け                                           |
| -------------------- | ------------------- | -------------------------------------------------- |
| **フィールドエラー** | `FormErrors.Fields` | 特定の入力フィールドに関連するエラー               |
| **グローバルエラー** | `FormErrors.Global` | フォーム全体に関連するエラー（同じページに留まる） |
| **Flash メッセージ** | `session.Flash`     | リダイレクト後に表示するメッセージ（成功/エラー）  |
| **ログのみ**         | `slog.Error`        | 開発者向け情報（ユーザーには一般メッセージを表示） |

**判断フローチャート**:

```
フォームを再表示する？
├─ Yes → FormErrors（Fields または Global）
│    └─ 特定フィールドに関連？ → AddFieldError（例: ユーザー名重複）
│    └─ フォーム全体に関連？  → AddGlobalError（例: 確認コード不一致）
└─ No（リダイレクトする）→ Flash
     └─ 成功 → FlashSuccess
     └─ エラー → FlashError
```

### 現状分析

#### Mewst (`/workspace/go/`)

**現状**:

| ファイル                       | 内容                                        | 問題点                   |
| ------------------------------ | ------------------------------------------- | ------------------------ |
| `confirm_email.go` (UseCase)   | コード検証ロジックが含まれている            | Handler で使われていない |
| `email_confirmation/create.go` | Handler でコード検証、Repository を直接使用 | UseCase を使っていない   |
| `update_password.go` (UseCase) | 純粋な永続化のみ                            | 良い例                   |

**修正方針**:

- `email_confirmation/create.go` で UseCase を使うように統一
- Handler では形式バリデーションとビジネスルール検証（コード一致チェック）を行い、UseCase は永続化のみを担当
- UseCase 名を `mark_email_as_confirmed.go` に変更（責務を明確化）

#### Annict (`/annict/go/`)

**現状**:

| ファイル                             | 内容                                       | 判定    |
| ------------------------------------ | ------------------------------------------ | ------- |
| `verify_sign_in_code.go` (UseCase)   | コード検証、失敗時に試行回数インクリメント | ✅ 妥当 |
| `verify_sign_up_code.go` (UseCase)   | コード検証、失敗時に試行回数インクリメント | ✅ 妥当 |
| `update_password_reset.go` (UseCase) | トークン検証、成功時に使用済みマーク       | ✅ 妥当 |

**分析**:
Annict の UseCase は「検証失敗時に DB 更新が必要」なケースであり、現状の設計は妥当です。

**修正方針**:

- コード修正は不要
- ドキュメントに設計方針を明記

#### Wikino (`/wikino/go/`)

**現状**:

| ファイル                                 | 内容                                         | 判定                       |
| ---------------------------------------- | -------------------------------------------- | -------------------------- |
| `verify_email_confirmation.go` (UseCase) | コード検証、有効期限チェック後に永続化       | ⚠️ Handler に移動可能      |
| `verify_two_factor.go` (UseCase)         | TOTP 検証、リカバリーコード消費              | ✅ 妥当                    |
| `create_account.go` (UseCase)            | メール確認チェック、アットネーム重複チェック | ⚠️ 一部 Handler に移動可能 |

**修正方針**:

- `verify_email_confirmation.go`: 検証を Handler に移動、UseCase は永続化（`MarkAsSucceeded`）のみに
- `create_account.go`: アットネーム重複チェックを Handler に移動

### Handler での検証パターン（推奨実装）

```go
// Handler: 検証 + UseCaseの呼び出し
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. 形式バリデーション（Request DTO）
    req := &CreateRequest{Code: r.FormValue("code")}
    if formErrors := req.Validate(ctx); formErrors.HasErrors() {
        h.renderForm(w, ctx, csrfToken, formErrors)
        return
    }

    // 2. ビジネスルール検証（Handler）
    // 2-1. レコード存在チェック
    emailConfirmation, err := h.emailConfirmationRepo.GetActiveByID(ctx, id)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            formErrors := session.NewFormErrors()
            formErrors.AddGlobalError(templates.T(ctx, "error_not_found"))
            h.renderForm(w, ctx, csrfToken, formErrors)
            return
        }
        // システムエラー
        slog.ErrorContext(ctx, "レコード取得に失敗", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 2-2. コード一致チェック
    if emailConfirmation.Code != req.Code {
        formErrors := session.NewFormErrors()
        formErrors.AddGlobalError(templates.T(ctx, "error_code_incorrect"))
        h.renderForm(w, ctx, csrfToken, formErrors)
        return
    }

    // 3. UseCase で永続化のみ
    if err := h.markEmailAsConfirmedUC.Execute(ctx, id); err != nil {
        slog.ErrorContext(ctx, "確認処理に失敗", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 4. 成功レスポンス
    h.flashMgr.SetSuccess(w, templates.T(ctx, "flash_email_confirmed"))
    http.Redirect(w, r, "/next-page", http.StatusFound)
}
```

### UseCase での検証パターン（試行回数インクリメントが必要な場合）

```go
// UseCase: 検証失敗時にもDB更新が必要なケース
func (uc *VerifySignInCodeUsecase) Execute(ctx context.Context, userID int64, code string) error {
    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("トランザクション開始に失敗: %w", err)
    }
    defer func() { _ = tx.Rollback() }()

    queriesWithTx := uc.queries.WithTx(tx)

    // コードを取得
    signInCode, err := queriesWithTx.GetValidSignInCode(ctx, userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ErrCodeNotFound
        }
        return fmt.Errorf("コードの取得に失敗: %w", err)
    }

    // コード検証（失敗時は試行回数をインクリメント）
    if !auth.VerifyCode(code, signInCode.CodeDigest) {
        // 失敗時もDB更新が必要 → UseCaseで検証するのが妥当
        if err := queriesWithTx.IncrementSignInCodeAttempts(ctx, signInCode.ID); err != nil {
            return fmt.Errorf("試行回数のインクリメントに失敗: %w", err)
        }
        if err := tx.Commit(); err != nil {
            return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
        }
        return ErrCodeInvalid
    }

    // 成功時はコードを使用済みに
    if err := queriesWithTx.MarkSignInCodeAsUsed(ctx, signInCode.ID); err != nil {
        return fmt.Errorf("コードの使用済み設定に失敗: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
    }

    return nil
}
```

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

- [x] **1-1**: [Go/Mewst] validation-guide.md と handler-guide.md の更新
  - バリデーションの 2 分類（形式バリデーション、状態バリデーション）に整理
  - ファイル名を `format_validator.go`（形式）と `state_validator.go`（状態）に変更
  - 構造体名を `{Action}FormatValidator` と `{Action}StateValidator` に変更
  - テスト方針（`format_validator_test.go`, `state_validator_test.go`, `handler_test.go`）を追記
  - 「検証失敗時に DB を更新する必要があるか？」という判断基準を追記
  - エラー表示方法の使い分け（フィールドエラー、グローバルエラー、Flash）を追記
  - Handler での検証パターンのコード例を追記
  - handler-guide.md の標準ファイル名を 8 種類から 10 種類に拡張
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 300 行（実装 300 行 + テスト 0 行）

### フェーズ 2: Mewst リファクタリング

- [x] **2-1**: [Go/Mewst] email_confirmation の UseCase 統一
  - `internal/usecase/confirm_email.go` を `mark_email_as_confirmed.go` にリネーム
  - UseCase から検証ロジックを削除し、永続化（`MarkAsSucceeded`）のみに変更
  - `email_confirmation/create.go` で UseCase を使用するように修正
  - Handler でコード一致チェックを行い、成功時に UseCase を呼び出す
  - テストの更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [x] **2-2**: [Go/Mewst] request.go を format_validator.go にリネーム
  - 各ハンドラーの `request.go` を `format_validator.go` にリネーム
  - 構造体名を `{Action}Request` から `{Action}FormatValidator` に変更
  - 対象ファイル:
    - `internal/handler/sign_in/request.go` → `format_validator.go`
    - `internal/handler/email_confirmation/request.go` → `format_validator.go`
    - `internal/handler/password/request.go` → `format_validator.go`
    - `internal/handler/password_reset/request.go` → `format_validator.go`
  - 各ハンドラーの create.go, update.go 等で使用している箇所も修正
  - `format_validator_test.go` を作成
  - テスト内容: 正常系、異常系（必須チェック、形式チェック）、境界値
  - **想定ファイル数**: 約 8 ファイル（実装 4 + テスト 4）
  - **想定行数**: 約 450 行（実装 50 行 + テスト 400 行）

- [x] **2-3**: [Go/Mewst] state_validator.go の作成
  - ハンドラーのメソッド内にあるビジネスルール検証を `state_validator.go` に抽出
  - 対象ファイル:
    - `internal/handler/sign_in/state_validator.go`（ユーザー検索、パスワード照合）
    - `internal/handler/email_confirmation/state_validator.go`（レコード存在チェック、コード一致チェック）
  - Handler の依存性を更新（stateValidator を追加）
  - `state_validator_test.go` を作成
  - **想定ファイル数**: 約 8 ファイル（実装 4 + テスト 4）
  - **想定行数**: 約 400 行（実装 150 行 + テスト 250 行）

- [x] **2-4**: [Go/Mewst] golangci-lint に depguard ルールを追加
  - `.golangci.yml` にバリデーターの依存関係ルールを追加
  - **format_validator.go のルール**（DB アクセス禁止）:
    - query への依存を禁止
    - repository への依存を禁止
    - model への依存を禁止
    - usecase への依存を禁止
  - **state_validator.go のルール**（Repository/Model は許可）:
    - query への依存を禁止（Repository 経由を強制）
    - usecase への依存を禁止
  - 依存関係まとめ:
    | パッケージ | format_validator | state_validator |
    |------------|:----------------:|:---------------:|
    | query | ❌ | ❌ |
    | repository | ❌ | ✅ |
    | model | ❌ | ✅ |
    | usecase | ❌ | ❌ |
    | session | ✅ | ✅ |
    | templates | ✅ | ✅ |
  - **前提**: タスク 2-2、2-3 が完了していること
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

### フェーズ 3: Wikino リファクタリング

- [x] **3-1**: [Go/Wikino] verify_email_confirmation.go のリファクタリング
  - Handler (`email_confirmation/update.go` または該当ファイル) で検証を行うように変更
  - UseCase は永続化（`MarkAsSucceeded`）のみに変更
  - UseCase 名を `mark_email_as_confirmed.go` などに変更することを検討
  - テストの更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 100 行（実装 50 行 + テスト 50 行）

- [x] **3-2**: [Go/Wikino] create_account.go のリファクタリング
  - アットネーム重複チェックを Handler (`account/create.go`) に移動
  - UseCase からアットネーム重複チェックを削除
  - テストの更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [x] **3-2a**: [Go/Wikino] request.go を format_validator.go にリネーム
  - 各ハンドラーの `request.go` を `format_validator.go` にリネーム
  - 構造体名を `{Action}Request` から `{Action}FormatValidator` に変更
  - テストファイルも同様にリネーム（`request_test.go` → `format_validator_test.go`）
  - 対象ファイル:
    - `internal/handler/sign_in/request.go` → `format_validator.go`
    - `internal/handler/sign_in_two_factor/request.go` → `format_validator.go`
    - `internal/handler/sign_in_two_factor_recovery/request.go` → `format_validator.go`
    - `internal/handler/email_confirmation/request.go` → `format_validator.go`
    - `internal/handler/account/request.go` → `format_validator.go`
  - 各ハンドラーの create.go, update.go 等で使用している箇所も修正
  - **想定ファイル数**: 約 15 ファイル（実装 10 + テスト 5）
  - **想定行数**: 約 100 行（実装 50 行 + テスト 50 行）

- [x] **3-2b**: [Go/Wikino] sign_in の state_validator.go 作成
  - `create.go` の状態バリデーション（ユーザー検索、パスワード照合）を `state_validator.go` に抽出
  - `CreateStateValidator` 構造体を作成
  - Handler の依存性を更新（stateValidator を追加）
  - `state_validator_test.go` を作成
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [x] **3-2c**: [Go/Wikino] email_confirmation の CreateStateValidator 追加
  - `create.go` のメールアドレス重複チェックを `state_validator.go` に追加
  - `CreateStateValidator` 構造体を作成（既存の `UpdateStateValidator` と同じファイル）
  - Handler の依存性を更新
  - テストを追加
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 150 行（実装 70 行 + テスト 80 行）

- [x] **3-2d**: [Go/Wikino] account の state_validator.go にメール確認チェック追加
  - `create.go` のメール確認情報取得・状態チェックを既存の `state_validator.go` に追加
  - `CreateStateValidator` の入力パラメータと検証内容を拡張
  - テストを追加
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 150 行（実装 70 行 + テスト 80 行）

- [x] **3-2e**: [Go/Wikino] sign_in_two_factor の state_validator.go 作成
  - TOTP 検証を UseCase から `state_validator.go` に移動（試行回数インクリメント不要のため）
  - `CreateStateValidator` 構造体を作成
  - UseCase の `VerifyTOTP` メソッドを削除または永続化処理のみに変更
  - Handler と UseCase の依存関係を調整
  - テストを更新
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 250 行（実装 120 行 + テスト 130 行）

- [x] **3-2f**: [Go/Wikino] sign_in_two_factor_recovery の state_validator.go 作成
  - リカバリーコード検証を `state_validator.go` に移動
  - リカバリーコード消費処理は UseCase に残す（`ConsumeRecoveryCode` メソッドに分離）
  - `CreateStateValidator` 構造体を作成
  - Handler と UseCase の依存関係を調整
  - テストを更新
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 250 行（実装 120 行 + テスト 130 行）

- [x] **3-3**: [Go/Wikino] golangci-lint に depguard ルールを追加
  - `.golangci.yml` にバリデーターの依存関係ルールを追加
  - **format_validator.go のルール**（DB アクセス禁止）:
    - query への依存を禁止
    - repository への依存を禁止
    - model への依存を禁止
    - usecase への依存を禁止
  - **state_validator.go のルール**（Repository/Model は許可）:
    - query への依存を禁止（Repository 経由を強制）
    - usecase への依存を禁止
  - 依存関係まとめ:
    | パッケージ | format_validator | state_validator |
    |------------|:----------------:|:---------------:|
    | query | ❌ | ❌ |
    | repository | ❌ | ✅ |
    | model | ❌ | ✅ |
    | usecase | ❌ | ❌ |
    | session | ✅ | ✅ |
    | templates | ✅ | ✅ |
  - **前提**: タスク 3-2a 〜 3-2f が完了していること
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

### フェーズ 4: Annict ドキュメント整備

- [x] **4-1**: [Go/Annict] CLAUDE.md へのバリデーション方針追記
  - Annict の CLAUDE.md にバリデーション方針のセクションを追加
  - 「試行回数インクリメントが必要なため UseCase で検証」という設計判断を明記
  - Mewst の validation-guide.md へのリンクを追加
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **Annict のコード修正**: 現状の設計は「検証失敗時に DB 更新が必要」なケースであり、UseCase に検証ロジックがあるのは妥当なため
- **バリデーションライブラリの導入**: 現状の手動バリデーションで十分対応できるため
- **共通バリデーション関数の抽出**: プロジェクト間でバリデーションロジックを共有する仕組みは、今後の課題とする

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Mewst validation-guide.md](/workspace/go/docs/validation-guide.md) - 現在のバリデーションガイド
- [Mewst architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
