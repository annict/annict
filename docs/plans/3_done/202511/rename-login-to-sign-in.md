# login → sign_in への命名変更 設計書

<!--
このテンプレートの使い方:
1. このファイルを `.claude/designs/2_todo/` ディレクトリにコピー
   例: cp .claude/designs/template.md .claude/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨
-->

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

Go 版のコードベース内で `login` と書かれている箇所を `sign_in` に統一的に変更します。これにより、Rails のコーディング規約や Go 版の命名規則（`sign_in`）との整合性を保ち、コードベース全体の可読性と保守性を向上させます。

**目的**:

- Go 版の命名規則の一貫性を保つ（既存のハンドラーは `sign_in`, `sign_in_password`, `sign_in_code`）
- Rails の慣習との整合性を保つ（Rails では `sign_in` が一般的）
- 将来のコードレビューや新規開発者のオンボーディングを容易にする

**背景**:

- 現在、メールログインコード機能では `email_login_codes` というテーブル名や `SendEmailLoginCodeUsecase` といった命名を使用しているが、これは既存の `sign_in` ハンドラーや Rails の慣習と不整合
- go.md の設計書（line 786-787）にも「`email_login_codes` テーブルは `sign_in_codes` にする」という TODO が記載されている
- この変更により、コードベース全体で `sign_in` という用語に統一される

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

- すべての `email_login` という命名を `sign_in` に変更する
- データベーステーブル `email_login_codes` を `sign_in_codes` にリネームする
- ファイル名、ディレクトリ名、関数名、構造体名、変数名など、すべての箇所で `login` を `sign_in` に変更する
- 国際化ファイル（ja.toml, en.toml）内のキー名を `email_login_code_*` から `sign_in_code_*` に変更する
- テストコードも含めて、すべての関連ファイルを更新する

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

- **後方互換性**: データベースマイグレーションは本番環境でダウンタイムなしで実行可能であること
- **テストカバレッジ**: 既存のテストカバレッジを維持すること
- **保守性**: すべての命名が一貫していること

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

### データベース設計

- **テーブル名**: `email_login_codes` → `sign_in_codes`
- **インデックス名**: `idx_email_login_codes_*` → `idx_sign_in_codes_*`
- **外部キー制約名**: `email_login_codes_user_id_fkey` → `sign_in_codes_user_id_fkey`

マイグレーションは `ALTER TABLE ... RENAME TO ...` を使用してテーブルとインデックスをリネームします。

### コード設計

変更が必要な箇所（洗い出し結果）：

#### 1. データベース関連
- `db/migrations/20251109101724_create_email_login_codes.sql` → 新しいマイグレーション作成でリネーム
- `db/schema.sql` → dbmate による自動更新

#### 2. SQL クエリファイル
- `internal/repository/queries/email_login_codes.sql` → `sign_in_codes.sql`

#### 3. リポジトリ層（sqlc 生成コード）
- `internal/repository/sqlc/email_login_codes.sql.go` → `sign_in_codes.sql.go` (sqlc 自動生成)
- `internal/repository/email_login_codes_test.go` → `sign_in_codes_test.go`

#### 4. usecase 層
- `internal/usecase/send_email_login_code.go` → `send_sign_in_code.go`
- `internal/usecase/send_email_login_code_test.go` → `send_sign_in_code_test.go`
- `internal/usecase/verify_email_login_code.go` → `verify_sign_in_code.go`
- `internal/usecase/verify_email_login_code_test.go` → `verify_sign_in_code_test.go`

#### 5. worker 層
- `internal/worker/send_email_login_code.go` → `send_sign_in_code.go`
- `internal/worker/send_email_login_code_test.go` → `send_sign_in_code_test.go`
- `internal/worker/cleanup_expired_email_login_codes.go` → `cleanup_expired_sign_in_codes.go`
- `internal/worker/cleanup_expired_email_login_codes_test.go` → `cleanup_expired_sign_in_codes_test.go`

#### 6. テンプレート（メール）
- `internal/templates/emails/email_login/` ディレクトリ → `sign_in/`
  - `ja_text.templ`, `ja_html.templ`, `en_text.templ`, `en_html.templ`

#### 7. handler 層
- `internal/handler/sign_in/*.go` → import 文の更新
- `internal/handler/sign_in_code/*.go` → import 文の更新
- `cmd/server/main.go` → 変数名、import 文の更新

#### 8. 国際化ファイル
- `internal/i18n/locales/ja.toml` → `email_login_code_*` → `sign_in_code_*`
- `internal/i18n/locales/en.toml` → `email_login_code_*` → `sign_in_code_*`

### 実装方針

- **段階的な変更**: データベース → リポジトリ → usecase → worker → テンプレート → handler → 国際化の順で変更
- **テストの維持**: 各ステップでテストが通ることを確認しながら進める
- **自動生成コードの再生成**: sqlc と templ の自動生成コードは適切なタイミングで再生成する

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
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）
-->

### フェーズ 1: データベース層の変更

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: データベーステーブルとインデックスのリネーム

  - 新しいマイグレーションファイルを作成（`rename_email_login_codes_to_sign_in_codes.sql`）
  - テーブル名を `email_login_codes` から `sign_in_codes` にリネーム
  - すべてのインデックスをリネーム（`idx_email_login_codes_*` → `idx_sign_in_codes_*`）
  - 外部キー制約をリネーム（`email_login_codes_user_id_fkey` → `sign_in_codes_user_id_fkey`）
  - マイグレーションを実行して `db/schema.sql` を更新
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
    - `db/migrations/YYYYMMDDHHMMSS_rename_email_login_codes_to_sign_in_codes.sql`
    - `db/schema.sql` (自動生成)
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

- [x] **1-2**: SQL クエリファイルとリポジトリ層の更新

  - `internal/repository/queries/email_login_codes.sql` → `sign_in_codes.sql` にリネーム
  - SQL ファイル内のテーブル名を `sign_in_codes` に更新
  - sqlc で Go コードを再生成（`sqlc generate`）
  - `internal/repository/email_login_codes_test.go` → `sign_in_codes_test.go` にリネーム
  - テスト内のすべての命名を `EmailLoginCode` → `SignInCode` に変更
  - テストを実行して動作確認
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
    - `internal/repository/queries/sign_in_codes.sql`（リネーム）
    - `internal/repository/sqlc/sign_in_codes.sql.go`（sqlc 自動生成）
    - `internal/repository/sign_in_codes_test.go`（リネーム）
    - （削除）`internal/repository/queries/email_login_codes.sql`
    - （削除）`internal/repository/email_login_codes_test.go`
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

### フェーズ 2: ビジネスロジック層の変更

- [x] **2-1**: usecase 層のリネームと更新

  - `send_email_login_code.go` → `send_sign_in_code.go` にリネーム
  - `send_email_login_code_test.go` → `send_sign_in_code_test.go` にリネーム
  - `verify_email_login_code.go` → `verify_sign_in_code.go` にリネーム
  - `verify_email_login_code_test.go` → `verify_sign_in_code_test.go` にリネーム
  - すべての構造体名、関数名、変数名を `EmailLoginCode` → `SignInCode` に変更
  - テストを実行して動作確認
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
    - `internal/usecase/send_sign_in_code.go`（リネーム）
    - `internal/usecase/send_sign_in_code_test.go`（リネーム）
    - `internal/usecase/verify_sign_in_code.go`（リネーム）
    - `internal/usecase/verify_sign_in_code_test.go`（リネーム）
  - **想定行数**: 約 250 行（実装 120 行 + テスト 130 行）

- [x] **2-2**: worker 層のリネームと更新

  - `send_email_login_code.go` → `send_sign_in_code.go` にリネーム
  - `send_email_login_code_test.go` → `send_sign_in_code_test.go` にリネーム
  - `cleanup_expired_email_login_codes.go` → `cleanup_expired_sign_in_codes.go` にリネーム
  - `cleanup_expired_email_login_codes_test.go` → `cleanup_expired_sign_in_codes_test.go` にリネーム
  - すべての構造体名、関数名、変数名、ジョブ Kind を変更
  - `internal/worker/client.go` を更新
  - テストを実行して動作確認
  - **想定ファイル数**: 約 5 ファイル（実装 3 + テスト 2）
    - `internal/worker/send_sign_in_code.go`（リネーム）
    - `internal/worker/send_sign_in_code_test.go`（リネーム）
    - `internal/worker/cleanup_expired_sign_in_codes.go`（リネーム）
    - `internal/worker/cleanup_expired_sign_in_codes_test.go`（リネーム）
    - `internal/worker/client.go`（更新）
  - **想定行数**: 約 280 行（実装 130 行 + テスト 150 行）

### フェーズ 3: プレゼンテーション層の変更

- [x] **3-1**: メールテンプレートのリネームと更新

  - `internal/templates/emails/email_login/` → `sign_in/` にディレクトリをリネーム
  - `ja_text.templ`, `ja_html.templ`, `en_text.templ`, `en_html.templ` を更新
  - パッケージ名を `email_login` → `sign_in` に変更
  - templ でコード生成（`make templ-generate`）
  - worker 層の import 文を更新
  - **想定ファイル数**: 約 5 ファイル（実装 5 + テスト 0）
    - `internal/templates/emails/sign_in/ja_text.templ`（リネーム）
    - `internal/templates/emails/sign_in/ja_html.templ`（リネーム）
    - `internal/templates/emails/sign_in/en_text.templ`（リネーム）
    - `internal/templates/emails/sign_in/en_html.templ`（リネーム）
    - `internal/worker/send_sign_in_code.go`（import 更新）
  - **想定行数**: 約 80 行（実装 80 行 + テスト 0 行）

- [x] **3-2**: handler 層と main.go の更新

  - `internal/handler/sign_in/*.go` の import 文を更新
  - `internal/handler/sign_in_code/*.go` の import 文と変数名を更新
  - `cmd/server/main.go` の変数名、import 文、periodic job 名を更新
  - テストを実行して動作確認
  - **想定ファイル数**: 約 8 ファイル（実装 8 + テスト 0）
    - `internal/handler/sign_in/*.go`（約 2 ファイル）
    - `internal/handler/sign_in_code/*.go`（約 5 ファイル）
    - `cmd/server/main.go`
  - **想定行数**: 約 80 行（実装 80 行 + テスト 0 行）

- [x] **3-3**: 国際化ファイルの更新

  - `internal/i18n/locales/ja.toml` 内の `email_login_code_*` キーを `sign_in_code_*` に変更
  - `internal/i18n/locales/en.toml` 内の `email_login_code_*` キーを `sign_in_code_*` に変更
  - すべての国際化キーを使用している箇所を更新（handler, usecase, worker, template）
  - テストを実行して動作確認
  - **想定ファイル数**: 約 10 ファイル（実装 10 + テスト 0）
    - `internal/i18n/locales/ja.toml`
    - `internal/i18n/locales/en.toml`
    - 国際化キーを使用している各ファイル（handler, usecase, worker, template）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

### フェーズ 4: 統合とテスト

- [ ] **4-1**: エンドツーエンドテストと動作確認

  - すべてのテストを実行（`make test`）
  - 開発サーバーを起動して手動で動作確認
    - サインインフォームでメールアドレスを入力
    - メールコードの受信を確認
    - コード入力でサインイン成功を確認
  - 不具合があれば修正
  - **想定ファイル数**: 約 0 ファイル（実装 0 + テスト 0）
  - **想定行数**: 約 0 行（実装 0 行 + テスト 0 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **API エンドポイントの変更**: 現時点では Web API は実装されていないため、エンドポイントの変更は不要
- **Rails 版との同期**: Rails 版には `email_login_codes` テーブルは存在しないため、Rails 側の変更は不要

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Rails の Devise における sign_in の慣習](https://github.com/heartcombo/devise)
- [Go 版 Annict の既存 sign_in ハンドラー](internal/handler/sign_in/)
- [go.md の設計書（line 786-787）](../.claude/designs/1_doing/go.md#L786-L787)

---

## テンプレート使用例

実際の使用例は以下を参照してください：

- [パスワードリセット機能](doing/password-reset.md)
