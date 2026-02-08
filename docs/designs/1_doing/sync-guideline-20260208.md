# ガイドライン同期 2026/02/08

## 概要

Annict・mewst・wikino の3プロジェクト間で、CLAUDE.md およびガイドラインドキュメントの差分を調査し、同期する。

## 設計

### 差分 1: バリデーション方針の設計パターン（最重要）

**対象ファイル**: `go/CLAUDE.md`, `go/docs/validation-guide.md`, `go/docs/handler-guide.md`

**差分の内容**:

- **Annict**: 旧方式（Request DTO パターン）。`format_validator.go` と `state_validator.go` の2ファイル分離、または `request.go` に形式バリデーションのみを配置し、DB アクセスを含むロジックはハンドラーに記述。標準ファイル名は8種類
- **mewst**: 新方式（Validator パターン）。`validator.go` 1ファイルに形式チェック + DB を使った検証を統合。構造体名は `{Action}Validator`、入出力は `{Action}ValidatorInput` / `{Action}ValidatorResult`。標準ファイル名は9種類
- **wikino**: mewst と同一の新方式（Validator パターン）。標準ファイル名は9種類

**採用方針**: mewst, wikinoの方式を採用します

---

### 差分 2: 「実装時のガイドライン」セクションの有無

**対象ファイル**: `CLAUDE.md`（トップレベル）

**差分の内容**:

- **Annict**: 該当セクションなし
- **mewst**: 該当セクションなし
- **wikino**: 「実装時のガイドライン」セクションがあり、既存コードとの一貫性について記述（類似処理の確認、エラーハンドリング・ログ出力・バリデーションのパターン踏襲）

**採用方針**: wikinoのガイドラインを参考にセクションを追加してください

---

### 差分 3: 「レビュー時に参照するガイドライン」セクションの有無

**対象ファイル**: `CLAUDE.md`（トップレベル）

**差分の内容**:

- **Annict**: 該当セクションなし
- **mewst**: 該当セクションなし
- **wikino**: 「レビュー時に参照するガイドライン」セクションがあり、共通・Go版・Rails版それぞれの参照すべきドキュメントとチェックポイントを詳細にリスト化

**採用方針**: wikinoのガイドラインを参考にセクションを追加してください

---

### 差分 4: コミット前チェックの記述スタイル（トップレベル CLAUDE.md）

**対象ファイル**: `CLAUDE.md`（トップレベル）

**差分の内容**:

- **Annict**: 具体的なコマンド名を記載（`make fmt`, `make lint`, `make test`）
- **mewst**: 抽象的な記述のみ（「コードフォーマット」「リント」「テスト」）。具体的なコマンドはサブプロジェクトの CLAUDE.md に委譲
- **wikino**: mewst と同様の抽象的な記述

**採用方針**: これはAnnictの方針とMewst, Wikinoの方針とどちらがわかりやすいですか？わかりやすいほうを採用してください

---

### 差分 5: 「避けるべきコメント」の書式

**対象ファイル**: `CLAUDE.md`（トップレベル）

**差分の内容**:

- **Annict**: 各項目に ❌ マークを付けている
- **mewst**: Annict と同じく ❌ マークあり
- **wikino**: ❌ マークがなく、太字のみ

**採用方針**: ❌ マークを付けるのを正とします。

---

### 差分 6: wikino トラブルシューティングの DB ポート番号不整合

**対象ファイル**: `CLAUDE.md`（トップレベル）

**差分の内容**:

- **Annict**: 該当なし（ポート番号は Annict 固有）
- **mewst**: 該当なし（ポート番号は mewst 固有）
- **wikino**: 「共通インフラ」セクションではポート `4204` だが、「トラブルシューティング」セクションでは `4104`（mewst と同じ値）と記載。コピペミスの可能性

**採用方針**: Wikinoのポートは `4204` なので、`4204` に修正してください

---

### 差分 7: tparse に関する記載

**対象ファイル**: `go/CLAUDE.md`

**差分の内容**:

- **Annict**: `make test` / `make test-verbose` に tparse の言及なし
- **mewst**: tparse を使用してテスト結果を整形表示する旨を記載（失敗テストの強調、パッケージサマリーテーブル）
- **wikino**: mewst と同様に tparse の記載あり

**採用方針**: mewst, wikinoの方針を採用してください

---

### 差分 8: Go バージョンの違い

**対象ファイル**: `go/CLAUDE.md`

**差分の内容**:

- **Annict**: Go 1.25.1
- **mewst**: Go 1.25.4
- **wikino**: Go 1.25.5

**採用方針**: これはプロジェクトによって異なるのでスルーで大丈夫です

---

### 差分 9: APP_ENV の例外規則

**対象ファイル**: `go/CLAUDE.md`

**差分の内容**:

- **Annict**: 記載なし
- **mewst**: 記載なし
- **wikino**: `APP_ENV` は他プロジェクトでも使われている名前のため、プレフィックス（`WIKINO_`）なしで使用するという例外ルールが明記

**採用方針**: wikinoで書かれていることを記載するようにしてください

---

### 差分 10: テーブル駆動テストの書き方セクション

**対象ファイル**: `go/CLAUDE.md`

**差分の内容**:

- **Annict**: 記載なし
- **mewst**: 「テーブル駆動テストの書き方」セクションあり（完全なコード例、使用する場面/使用しない場面のガイドライン付き）
- **wikino**: 記載なし

**採用方針**: mewstで書かれていることを記載するようにしてください

---

### 差分 11: UseCase のトランザクション管理（WithTx パターン）のコード例

**対象ファイル**: `go/CLAUDE.md`, `go/docs/architecture-guide.md`

**差分の内容**:

- **Annict**: WithTx パターンの詳細なコード例なし
- **mewst**: `WithTx` パターンの詳細なコード例あり（`CreateAccountUsecase`）。トランザクション管理の項目が追加
- **wikino**: mewst と同様の `WithTx` パターンに加え、Worker セクションが追加。architecture-guide.md に「Repository の WithTx パターン」セクションとして詳細に記載（なぜ使うのか、実装方法、重要なポイント）

**採用方針**: wikinoで書かれていることを記載するようにしてください

---

### 差分 12: architecture-guide.md の命名規則セクションと Query ファイル命名セクション

**対象ファイル**: `go/docs/architecture-guide.md`

**差分の内容**:

- **Annict**: Model/Repository の命名規則テーブルと Query ファイルの命名パターンセクションあり
- **mewst**: Annict と同様に記載あり
- **wikino**: これらのセクションが欠落

**採用方針**: Annict, Mewstで書かれていることを記載するようにしてください

---

### 差分 13: architecture-guide.md の「Usecase と Repository の使い分け」セクション

**対象ファイル**: `go/docs/architecture-guide.md`

**差分の内容**:

- **Annict**: 「Usecase と Repository の使い分け」セクションあり（判断基準を詳細に記載）
- **mewst**: Annict と同様に記載あり
- **wikino**: このセクションが欠落

**採用方針**: Annict, Mewstで書かれていることを記載するようにしてください

---

### 差分 14: templ-guide.md の Flash コンポーネント引数

**対象ファイル**: `go/docs/templ-guide.md`

**差分の内容**:

- **Annict**: `@components.Flash(ctx, flash)` — ctx を明示的に渡している
- **mewst**: Annict と同じく ctx を渡している
- **wikino**: `@components.Flash(flash)` — ctx を渡していない（templ の暗黙的 ctx を利用、CLAUDE.md の「context.Context を明示的に渡さない」方針に合致）

**採用方針**: Wikinoの方針を採用してください

---

### 差分 15: rails/CLAUDE.md の「作業完了ガイドライン」セクション

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: 該当セクションなし
- **mewst**: 「作業完了ガイドライン」セクションあり（タスク実装フロー、リトライポリシー（最大5回）、完了報告の禁止条件、検証コマンド一覧）
- **wikino**: mewst とほぼ同内容に加え、リトライポリシーにユーザーへの報告メッセージ例、`bin/check` による全検証の一括実行などが追加

**採用方針**: wikinoと同等の記述になるようにしてください

---

### 差分 16: rails/CLAUDE.md のコミット前チェックの網羅性

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: ERB リント (`bin/erb_lint --lint-all`) と Prettier チェックがコミット前チェックに含まれていない。`sorbet-update` → `zeitwerk` の順
- **mewst**: ERB リント、Prettier チェック含む。`zeitwerk` → `sorbet-update` の順
- **wikino**: mewst と同様に加え、TypeScript 型チェック (`pnpm tsc`)、`bin/check` による一括検証あり

**採用方針**: wikinoと同等の記述になるようにしてください

---

### 差分 17: rails/CLAUDE.md の RSpec コーディング規約

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: RSpec の固有規約なし
- **mewst**: `context` / `let` / `described_class` の使用禁止、`it` ブロック内で変数を定義、FactoryBot レコードに `_record` サフィックス、システムテストでの `sleep` 禁止
- **wikino**: mewst と同一の規約

**採用方針**: mewst, wikino同等の記述になるようにしてください

---

### 差分 18: rails/CLAUDE.md の ActiveRecord `includes` 禁止規約

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: 記載なし
- **mewst**: 記載なし
- **wikino**: `includes` 使用禁止。明示的に `preload` または `eager_load` を使用する規約

**採用方針**: wikinoと同等の記述になるようにしてください

---

### 差分 19: rails/CLAUDE.md のサービスクラスルール詳細度

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: 簡潔に記載（配置、命名、`call` メソッドのみ）
- **mewst**: UseCase パターンで `call` メソッドを実装
- **wikino**: 詳細なルール（使用する場合/しない場合の基準、`#with_transaction` メソッド、`ApplicationRecord.transaction` 直接使用禁止、Service と Job の依存関係ルール）

**採用方針**: wikinoと同等の記述になるようにしてください

---

### 差分 20: rails/CLAUDE.md のクラス間依存関係ルール

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: 記載なし
- **mewst**: 記載なし
- **wikino**: Controller, Service, Repository, Model, Record, Component, View, Form, Job, Mailer, Policy, Validator の依存可能先を表形式で定義

**採用方針**: wikinoと同等の記述になるようにしてください

---

### 差分 21: rails/CLAUDE.md の Ruby 細かいコーディング規約

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: 基本的な規約のみ（インデント、Standard、Sorbet `typed: true`）
- **mewst**: Annict とほぼ同等
- **wikino**: 追加規約多数（ダブルクオート使用、ハッシュ省略記法、`private def`、後置 `if` 不使用、`not_nil!` 使用、`typed: strict`、1行100文字以内）

**採用方針**: wikinoと同等の記述になるようにしてください

---

### 差分 22: rails/CLAUDE.md の「重要な原則」セクション

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: 記載なし
- **mewst**: 記載なし
- **wikino**: ネストトランザクション回避、レコードコールバック回避、View/Component での DB アクセス防止、説明的な命名規則、コメントは日本語、1行100文字以内

**採用方針**: wikinoと同等の記述になるようにしてください

---

### 差分 23: rails/CLAUDE.md のトラブルシューティングセクション

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: 記載なし
- **mewst**: 「デバッグ・トラブルシューティング」セクションあり（Sorbet エラー、オートローディングエラー、フォーマットエラー、Lint エラーの解決方法）
- **wikino**: mewst とほぼ同内容

**採用方針**: mewst, wikinoと同等の記述になるようにしてください

---

### 差分 24: rails/CLAUDE.md の 1Password CLI 注意書き

**対象ファイル**: `rails/CLAUDE.md`

**差分の内容**:

- **Annict**: 記載なし
- **mewst**: 「環境変数は 1Password CLI 経由で自動設定されるため、`make` コマンドを使用してください」
- **wikino**: 同様の注意書きあり

**採用方針**: mewst, wikinoと同等の記述になるようにしてください

---

### 差分なしのファイル

以下のファイルは3プロジェクト間で差分なし（プロジェクト固有のカスタマイズのみ）:

- `go/docs/security-guide.md` — 差分なし
- `go/docs/i18n-guide.md` — 差分なし

## タスクリスト

### フェーズ 1: Annict (自プロジェクト)

- [x] **1-1**: トップレベル CLAUDE.md の更新
  - 差分2: 「実装時のガイドライン」セクション追加（Wikino を参考）
  - 差分3: 「レビュー時に参照するガイドライン」セクション追加（Wikino を参考）
  - 差分4: コミット前チェックの記述を抽象化（Mewst/Wikino 方式: 具体的なコマンドはサブプロジェクトの CLAUDE.md に委譲）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

- [x] **1-2**: go/CLAUDE.md の更新
  - 差分1: バリデーション方針セクションを Validator パターンに更新、標準ファイル名を9種類に変更
  - 差分7: tparse の記載追加（Mewst/Wikino を参考）
  - 差分9: APP_ENV 例外規則追加（Wikino を参考）
  - 差分10: テーブル駆動テストの書き方セクション追加（Mewst を参考）
  - 差分11: WithTx パターンのコード例追加（Wikino を参考）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

- [x] **1-3**: go/docs/validation-guide.md の全面更新
  - 差分1: Request DTO パターンから Validator パターンへ書き換え（Mewst/Wikino を参考）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 200 行（実装 200 行 + テスト 0 行）

- [x] **1-4**: go/docs/handler-guide.md の更新
  - 差分1: request.go → validator.go へ更新、テストファイル構成の変更（Mewst/Wikino を参考）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

- [x] **1-5**: go/docs/architecture-guide.md, templ-guide.md の更新
  - 差分11: architecture-guide.md に WithTx パターンセクション追加（Wikino を参考）
  - 差分14: templ-guide.md の Flash コンポーネント引数から ctx を削除（Wikino を参考）
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

- [x] **1-6**: rails/CLAUDE.md コーディング規約・原則の追加
  - 差分17: RSpec コーディング規約追加（Mewst/Wikino を参考）
  - 差分18: ActiveRecord `includes` 禁止規約追加（Wikino を参考）
  - 差分21: Ruby 細かいコーディング規約追加（Wikino を参考）
  - 差分22: 「重要な原則」セクション追加（Wikino を参考）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 80 行（実装 80 行 + テスト 0 行）

- [x] **1-7**: rails/CLAUDE.md アーキテクチャ・依存関係ルールの更新
  - 差分19: サービスクラスルール詳細化（Wikino を参考）
  - 差分20: クラス間依存関係ルール追加（Wikino を参考）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 60 行（実装 60 行 + テスト 0 行）

- [x] **1-8**: rails/CLAUDE.md ワークフロー・運用ガイドラインの更新
  - 差分15: 作業完了ガイドライン追加（Wikino を参考）
  - 差分16: コミット前チェック網羅性向上（Wikino を参考）
  - 差分23: トラブルシューティングセクション追加（Mewst/Wikino を参考）
  - 差分24: 1Password CLI 注意書き追加（Mewst/Wikino を参考）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

### フェーズ 2: Mewst

- [ ] **2-1**: CLAUDE.md（トップレベル）の更新
  - 差分2: 「実装時のガイドライン」セクション追加（Wikino を参考）
  - 差分3: 「レビュー時に参照するガイドライン」セクション追加（Wikino を参考）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

- [ ] **2-2**: go/CLAUDE.md, go/docs/architecture-guide.md, go/docs/templ-guide.md の更新
  - 差分9: APP_ENV 例外規則追加
  - 差分11: go/CLAUDE.md に WithTx パターンのコード例追加、architecture-guide.md に WithTx パターンセクション追加
  - 差分14: templ-guide.md の Flash コンポーネント引数から ctx を削除
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 60 行（実装 60 行 + テスト 0 行）

- [ ] **2-3**: rails/CLAUDE.md の更新
  - 差分18: ActiveRecord `includes` 禁止規約追加
  - 差分19: サービスクラスルール詳細化
  - 差分20: クラス間依存関係ルール追加
  - 差分21: Ruby 細かいコーディング規約追加
  - 差分22: 「重要な原則」セクション追加
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

### フェーズ 3: Wikino

- [ ] **3-1**: CLAUDE.md（トップレベル）の更新
  - 差分5: 「避けるべきコメント」に ❌ マークを追加
  - 差分6: トラブルシューティングの DB ポート番号を `4204` に修正
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

- [ ] **3-2**: go/CLAUDE.md, go/docs/architecture-guide.md の更新
  - 差分10: テーブル駆動テストの書き方セクション追加（Mewst を参考）
  - 差分12: architecture-guide.md に命名規則セクション・Query ファイル命名セクション追加（Annict/Mewst を参考）
  - 差分13: architecture-guide.md に「Usecase と Repository の使い分け」セクション追加（Annict/Mewst を参考）
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 80 行（実装 80 行 + テスト 0 行）

### 対応不要

以下の差分はスキップ:

- **差分8（Go バージョン）**: プロジェクト固有のため対応不要
