---
name: sync
description: プロジェクト間ファイル同期。Annict・Mewst・Wikino の3プロジェクト間で共通ファイルの差分を調査し、採用方針を決定するためのレポートを作成する。
argument-hint: "annict | mewst | wikino"
---

# プロジェクト間ファイル同期

**自プロジェクト**: $ARGUMENTS

## 対象プロジェクト

引数で渡されたプロジェクトが自プロジェクト（`/workspace/`）として扱われます。

各プロジェクトには **public リポジトリ**と **private リポジトリ**の 2 つのルートディレクトリがあります。

| 引数     | 自プロジェクト（public / private） | 他プロジェクト（public / private）                                                 |
| -------- | ---------------------------------- | ---------------------------------------------------------------------------------- |
| `mewst`  | `/workspace/` / `/mewst-private/`  | Annict (`/annict/` / `/annict-private/`), Wikino (`/wikino/` / `/wikino-private/`) |
| `annict` | `/workspace/` / `/annict-private/` | Mewst (`/mewst/` / `/mewst-private/`), Wikino (`/wikino/` / `/wikino-private/`)    |
| `wikino` | `/workspace/` / `/wikino-private/` | Mewst (`/mewst/` / `/mewst-private/`), Annict (`/annict/` / `/annict-private/`)    |

**ルール**:

- 引数のプロジェクト → public: `/workspace/`、private: `/{プロジェクト名}-private/`
- それ以外 → public: `/{プロジェクト名}/`、private: `/{プロジェクト名}-private/`

## 比較対象ファイル

比較対象ファイルは **public リポジトリ**と **private リポジトリ**の 2 つのルートディレクトリに分かれています。各ファイルをまず public リポジトリから探し、見つからない場合は private リポジトリに存在します。

### public リポジトリ（`/workspace/` または `/{プロジェクト名}/`）

- `.github/dependabot.yml`
- `.github/workflows/*.yml`
- `.oxfmtrc.json`
- `apm.yml`
- `docker-compose.yml`
- `Dockerfile.dev`
- `docs/README.md`
- `docs/specs/template.md`
- `docs/system/template.md`
- `go/.golangci.yml`
- `go/Makefile`
- `Makefile`
- `rails/.prettierrc`
- `rails/.rspec`
- `rails/eslint.config.mjs`
- `rails/Makefile`

### private リポジトリ（`/{プロジェクト名}-private/`）

- `docker-compose.override.yml`
- `docs/private/plans/template.md`
- `docs/private/reviews/template.md`
- `docs/private/README.md`

## スキップ済みの差分

過去の /sync 実行で「スキップ（対応不要）」と判断された差分です。
次回以降の /sync ではこの一覧に該当する差分はレポートに含めません。
再度確認が必要になった場合は該当行を削除してください。

**適用ルール**: `/sync mewst` → 共通 + Mewst のスキップを適用

### 共通（全プロジェクトでスキップ）

（現在なし）

### Annict

- `rails/CLAUDE.md`: ビジネスロジック層の名称の違い（Mewst は Use Cases、他は Services。プロジェクト固有の命名のため）
- `Dockerfile.dev`: imagemagick の有無（プロジェクト固有の依存のため）
- `rails/.rspec`: require 対象の違い（プロジェクト固有の設定のため）
- `Dockerfile.dev`: Bundler バージョンの違い（プロジェクト固有の依存のため）
- `.github/workflows/rails-ci.yml`: RSpec runs-on の違い（プロジェクト固有の事情のため）
- `.github/workflows/rails-ci.yml`: DB セットアップ方式の違い（db:setup vs db:schema:load。プロジェクト固有の事情のため）
- `rails/docs/architecture-guide.md`: Mewst に存在しない（Mewst 固有の事情のため）
- `go/CLAUDE.md`: Worker パッケージの記述（Wikino 固有の機能のため）
- `go/.golangci.yml`: 追加レイヤールール（policy-layer, markup-layer, worker-layer。Wikino 固有のアーキテクチャのため）
- `go/docs/security-guide.md`: スペース ID によるクエリスコープ（Wikino 固有のドメイン要件のため）

### Mewst

- `rails/CLAUDE.md`: ビジネスロジック層の名称の違い（Mewst は Use Cases、他は Services。プロジェクト固有の命名のため）
- `Dockerfile.dev`: imagemagick の有無（プロジェクト固有の依存のため）
- `rails/.rspec`: require 対象の違い（プロジェクト固有の設定のため）
- `go/CLAUDE.md`: Go 版で処理するパスの例の詳細度の違い（プロジェクト固有のため）
- `go/docs/architecture-guide.md`: Presentation 層ヘルパーの image パッケージの有無（プロジェクト固有のため）
- `docs/plans/template.md`: Rails 版ガイドラインの参照リスト数の違い（プロジェクト固有のため）
- `rails/.prettierrc`: Mewst ではプロジェクトルートの Oxfmt を使用しており、rails-ci.yml に prettier ジョブがないため不要
- `Dockerfile.dev`: hivemind curl -L フラグの差異（`-fsSL` にすでに `-L` が含まれており実質差分なし）
- `rails/docs/testing-guide.md`: 詳細度の違い（プロジェクト固有の要素が多いため）
- `.claude/commands/review.md`: Annict のみ存在（Mewst/Wikino は /review スキルで対応済み）

### Wikino

- `rails/.rspec`: require 対象の違い（プロジェクト固有のため）
- `Dockerfile.dev`: Mewst の imagemagick 欠落（プロジェクト固有のため）
- `go/docs/templ-guide.md`: Annict のテストコメントが古い記述（Annict 側で対応すべきため）
- `go/docs/validation-guide.md`: Annict のテストコメントが古い記述（Annict 側で対応すべきため）
- `go/CLAUDE.md`: SetupTx/GetTestDB 使い分けテーブルが Annict に欠落（Annict 側で対応すべきため）
- `Dockerfile.dev`: Annict に hivemind 欠落（Annict 側で対応すべきため）
- `Dockerfile.dev`: Bundler バージョンの違い（プロジェクト固有のため）
- `go/.golangci.yml`: depguard の policy-layer、markup-layer ルール差異（プロジェクト固有のため）
- `CLAUDE.md`: Annict に fmt-ci.yml 記載なし（プロジェクト固有の CI 構成のため）
- `go/CLAUDE.md`: Worker/ドメインID型/スペースIDセクション（Wikino 固有機能のため）
- `go/docs/architecture-guide.md`: Mewst で image ヘルパー欠落・UserCalendar 例残存（Mewst 固有の問題）
- `rails/eslint.config.mjs`: Annict の globalIgnores 追加（Annict 固有の設定のため）
- `go/docs/security-guide.md`: スペースIDクエリスコープセクション（Wikino 固有機能のため）
- `.github/workflows/go-ci.yml`: Go/PostgreSQL バージョン差異（プロジェクト固有のため）

## 手順

### ステップ 1: レビュードキュメントの作成

1. `.claude/skills/sync/review-template.md` をコピーして `docs/reviews/sync-YYYYMMDD-{3桁連番}.md` を作成する（YYYYMMDD は実行日）
2. 連番は `docs/private/reviews/` と `docs/private/reviews/done/` 配下から `sync-YYYYMMDD-*` のファイルを検索し、既存の最大連番 + 1 を付与する
3. レビュー情報テーブルを記入する（実行日、自プロジェクト名）

### ステップ 2: 並列差分調査

**事前準備**: 本ファイルの「スキップ済みの差分」セクションを確認し、共通 + 自プロジェクトのスキップ対象を把握する。

**重要**: コンテキスト制限を回避するため、比較対象ファイルを **5件ずつのバッチ** に分割し、バッチごとに Agent tool（subagent_type: `general-purpose`）を使って並列で差分調査を行う。

各 Agent tool には以下を指示する：

- 指定されたバッチのファイルを全プロジェクト間で比較する
- **比較対象ファイルの探索**: public リポジトリをまず確認し、見つからない場合は private リポジトリから読み取る
- 存在しないファイルはスキップする
- 差分がないファイルは「差分なし」と明記する
- プロジェクト名のみの差異は差分として報告しない
- 実質的な内容の違い（設計方針、設定値、セクションの有無等）を報告する
- **スキップ済みの差分一覧に記載されている差分は無視する**（レポートに含めない）

### ステップ 3: 差分レポートの記入

レビュードキュメントに以下を記入する：

1. **比較対象ファイル一覧**: 差分がないファイルにはチェックを入れる
2. **ファイルごとの差分**: 差分があるファイルのみ、レビューテンプレートの形式に従って記入する
3. 各差分に「推奨ガイダンス」に基づいて推奨（Recommended）を付記する
4. **プロジェクト固有の差分**: 同期対象外の差分をチェックボックス付きで記載する（各項目に 3 プロジェクトの現状を要約）
5. レビュー情報の差分件数を更新する

### ステップ 4: ユーザーへの報告

以下の内容をユーザーに報告する：

1. 作成したレビュードキュメントのパス
2. 発見した差分の件数（同期対象 + プロジェクト固有）
3. 「各差分の **対応方針** のチェックボックスにチェックを入れてください。**プロジェクト固有の差分** で共通化したい項目にもチェックを入れてください。記入後に変更を適用します。」というメッセージ

### ステップ 5: 変更の適用

ユーザーが対応方針にチェックを入れて対応を依頼した後、以下を実行する：

- チェックされた方針に基づき、**3つのプロジェクト全て**（public リポジトリ・private リポジトリ）のファイルを直接編集する
  - 例：「Annict の方式を採用」→ Annict の該当箇所の内容を Mewst・Wikino にも反映する
- **git コミットは行わない**（ユーザーが各プロジェクトで確認後にコミットする）
- プロジェクト名の差異は各プロジェクトに合わせて維持する

### ステップ 6: スキップ済み差分の更新

ユーザーが対応方針で「スキップ（対応不要）」を選んだ差分について、本ファイルの「スキップ済みの差分」セクションの適切な見出し（共通 or プロジェクト名）に概要を追記する。

```markdown
- `go/CLAUDE.md`: Go バージョンの違い（プロジェクト固有のため）
```

### ステップ 7: プロジェクト固有の差分の作業計画書作成

「プロジェクト固有の差分」セクションでユーザーがチェックを入れた項目がある場合、以下を実行する：

1. **調査**: チェックされた各項目について、3 プロジェクト間の現状を Agent tool で並列調査する
   - 各プロジェクトのコード実装の有無と内容
   - ドキュメントの有無と内容
   - 統一する場合の推奨方針

2. **作業計画書の作成**: 各プロジェクトの private リポジトリに作業計画書を作成する
   - パス: `/{プロジェクト名}-private/docs/plans/1_doing/cross-project-unification-YYYYMMDD.md`（YYYYMMDD は実行日）
   - `docs/private/plans/template.md` をベースに作成する
   - **各プロジェクトには自プロジェクトで必要な修正のみをタスクリストに記載する**
   - 既に対応済みの項目はタスクリストに含めず、「実装しない機能（スコープ外）」に記載する
   - 設計セクションには全項目の現状比較表と統一方針を記載する

3. チェックされなかった項目はスキップ済み差分に追記する（ステップ 6 と同様）

### ステップ 8: 完了通知

ユーザーに以下を報告する：

1. 変更を適用したファイルの一覧（プロジェクトごと）
2. スキップ済みの差分に追記した内容（該当がある場合）
3. 作業計画書を作成した場合はそのパス一覧
4. 「各プロジェクトで差分を確認し、コミットしてください。」というメッセージ

## 採用方針の推奨ガイダンス

差分レポートを作成する際、各差分に対して以下の推奨を付記する。ユーザーが判断しやすくなるよう、**推奨（Recommended）** タグを適切な選択肢に追加する。

### 推奨の判断基準

**重要**: プロジェクト数の多数決で推奨を決めない。内容の正確さ・新しさで判断する。

| 状況                               | 推奨                         |
| ---------------------------------- | ---------------------------- |
| 一方がより正確・最新の内容を含む   | その方式を推奨               |
| 一方がより限定的・安全な設定である | その方式を推奨               |
| プロジェクト固有の事情がある       | スキップを推奨               |
| どちらでも良い（好みの問題）       | 推奨なし（ユーザーに委ねる） |

### 推奨の付け方

差分レポートの対応方針チェックボックスに、推奨する選択肢に `(Recommended)` を付記する。

```markdown
**対応方針**:

- [ ] Mewst の方式を採用（他プロジェクトに反映）
- [ ] Annict/Wikino の方式を採用（他プロジェクトに反映）(Recommended)
- [ ] スキップ（対応不要）
- [ ] その他: <!-- 記入 -->
```

### 推奨の例

- **バージョンの違い**: 最新バージョンを使用しているプロジェクトを推奨
- **glob パターンの精度**: より限定的・正確なパターンを推奨
- **ドキュメントの充実度**: より詳細・正確に記載しているプロジェクトを推奨
- **数値の不一致**: 実際のファイル数と合っている方を推奨
- **CI 設定の記載**: 実際のファイル構成と合っている方を推奨
- **新しい概念の追加**: 追加したプロジェクトを推奨（他プロジェクトに未反映なだけ）

## 注意事項

- 各プロジェクトには public リポジトリと private リポジトリの 2 つのルートディレクトリがある
- 比較対象ファイルはまず public リポジトリから探し、見つからない場合は private リポジトリから読み取る
- 存在しないファイルはスキップする
- 差分がないファイルは「差分なし」と明記する
- プロジェクト名のみの差異は差分として報告しない
- 差分調査は Agent tool で並列処理（5件ずつのバッチ）してコンテキスト制限を回避する
- 変更の適用時は 3 プロジェクト全て（public・private 両方）のファイルを直接編集し、git コミットは行わない
- 3 プロジェクト間で完全一致が確認されたファイル（`go/.air.toml`、`rails/.standard.yml` 等）は比較対象外
