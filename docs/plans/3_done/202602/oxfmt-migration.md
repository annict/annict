# Prettier から Oxfmt への移行 作業計画書

<!--
このテンプレートの使い方:
1. このファイルを `docs/plans/2_todo/` ディレクトリにコピー
   例: cp docs/plans/template.md docs/plans/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残してください

**作業計画書の性質**:
- 作業計画書は「何をどう変えるか」という変更内容を記述するドキュメントです
- 新しい機能の場合は、概要・要件・設計もこのドキュメントに記述します
- 現在のシステムの状態は `docs/specs/` の仕様書に記述されています
- タスク完了後は、仕様書を新しい状態に更新してください（設計判断や採用しなかった方針も含める）

**仕様書との関係**:
- 新しい機能の場合: タスク完了後に `docs/specs/` に仕様書を作成する
- 既存機能の変更の場合: 「仕様書」セクションに対応する仕様書へのリンクを記載し、タスク完了後に仕様書を更新する

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 仕様書

<!--
- 既存機能を変更する場合: 変更対象の仕様書へのリンクを記載してください
- 新しい機能の場合: タスク完了後に作成予定の仕様書のパスを記載してください
-->

- タスク完了後に作成: `docs/specs/formatter.md`

## 概要

<!--
ガイドライン:
- この機能が「何であるか」「なぜ必要か」を簡潔に説明
- 2-3段落程度で簡潔に
- 既存機能の変更の場合は、変更の背景と目的を記述
-->

プロジェクト全体のコードフォーマッタを Prettier から [Oxfmt](https://oxc.rs/docs/guide/usage/formatter) に移行する。

### 現状の課題

- Prettier は Rails プロジェクト（`rails/`）のみに導入されており、Go プロジェクト（`go/`）やルートレベルの Markdown ファイルはフォーマット対象外
- Rails の `.prettierignore` で `app/docs`、`app/views`、`sorbet`、`tmp` を除外しているが、プロジェクト全体で統一的なフォーマット管理ができていない
- Go プロジェクトの TOML 翻訳ファイル（`go/internal/i18n/locales/*.toml`）も統一的にフォーマットされていない
- Prettier のプラグイン（`prettier-plugin-tailwindcss`、`@ttskch/prettier-plugin-tailwindcss-anywhere`）への依存がある

### Oxfmt を選定した理由

- Prettier の約 30 倍高速（Rust 製）
- JavaScript, TypeScript, CSS, YAML, Markdown, TOML, JSON をプラグインなしでネイティブサポート
- Prettier 100% 互換（JS/TS）
- Tailwind CSS クラスソート内蔵（`prettier-plugin-tailwindcss` が不要に）
- `oxfmt --migrate=prettier` コマンドで既存設定を移行可能
- Vue.js、Turborepo、Sentry JavaScript SDK などの主要プロジェクトが採用済み

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

- プロジェクトルートから `pnpm oxfmt` を実行して、プロジェクト全体の対象ファイルをフォーマットできる
- `pnpm oxfmt --check` で CI フォーマットチェックを実行できる
- フォーマット対象: JavaScript, TypeScript, CSS, SCSS, YAML, Markdown, TOML, JSON ファイル
- Rails 版の既存フォーマット設定（printWidth: 120, tabWidth: 2, ダブルクォート, トレイリングカンマ）を引き継ぐ
- Rails 版から Prettier を完全に削除する

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

- CI でのフォーマットチェックが高速に完了すること（Prettier より改善）
- 既存のフォーマット済みファイルとの差分を最小限に抑えること

## 実装ガイドラインの参照

<!--
**重要**: 作業計画書を作成する前に、対象プラットフォームのガイドラインを必ず確認してください。
特に以下の点に注意してください：
- ディレクトリ構造・ファイル名の命名規則
- コーディング規約
- アーキテクチャパターン

ガイドラインに沿わない設計は、実装時にそのまま実装されてしまうため、
作業計画書作成の段階でガイドラインに準拠していることを確認してください。
-->

### Go 版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - 全体的なコーディング規約

### Rails 版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@rails/CLAUDE.md](/workspace/rails/CLAUDE.md) - 全体的なコーディング規約

## 設計

<!--
ガイドライン:
- 技術的な実装の設計を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - UI設計（画面構成、ユーザーフローなど）
  - セキュリティ設計（認証・認可、トークン管理など）
  - コード設計（パッケージ構成、主要な構造体など）

**重要: 設計は実装中に更新する**:
- 作業計画書内の設計は初期の方針であり、完璧ではない
- 実装中により良いアプローチが見つかった場合は、設計を積極的に更新する
- 設計に固執して実装の質を下げるよりも、実装で得た知見を設計に反映する方が重要
- 変更した場合は「採用しなかった方針」セクションに変更前の方針と変更理由を記録する
-->

### パッケージマネージャーの選定

プロジェクトルートでは **pnpm** を使用する。Go プロジェクトが既に pnpm を使用しているため統一する。

### ディレクトリ構成

```
/workspace/
├── package.json          # 新規作成: oxfmt の依存関係
├── pnpm-lock.yaml        # 新規作成: pnpm のロックファイル
├── Makefile              # 新規作成: プロジェクトルートの Makefile（make fmt / make fmt-check）
├── .oxfmtignore          # 新規作成: フォーマット除外設定
├── .github/
│   └── workflows/
│       ├── fmt-ci.yml    # 新規作成: プロジェクト全体のフォーマットチェック
│       └── rails-ci.yml  # 変更: prettier ジョブを削除
├── rails/
│   ├── package.json      # 変更: prettier 関連パッケージを削除
│   ├── .prettierrc       # 削除
│   ├── .prettierignore   # 削除
│   └── bin/check         # 変更: prettier コマンドを oxfmt に置き換え
└── ...
```

### Oxfmt 設定

`oxfmt --migrate=prettier` を使用して Rails の `.prettierrc` から設定を移行する。主要な設定値:

- **printWidth**: 120
- **tabWidth**: 2
- **semi**: true
- **singleQuote**: false（ダブルクォート）
- **trailingComma**: all

### フォーマット対象と除外

**対象ファイル**:

- `*.js`, `*.jsx`, `*.ts`, `*.tsx` — JavaScript / TypeScript
- `*.css`, `*.scss` — スタイルシート
- `*.json`, `*.jsonc` — JSON
- `*.yaml`, `*.yml` — YAML
- `*.md` — Markdown
- `*.toml` — TOML（Go の翻訳ファイルなど）

**除外対象**（`.oxfmtignore`）:

- `**/node_modules/` — 依存パッケージ
- `**/vendor/` — バンドルされた Gem
- `rails/db/` — DB スキーマ・マイグレーション（自動生成）
- `rails/app/docs/` — 現行の `.prettierignore` と同等
- `rails/app/views/` — 現行の `.prettierignore` と同等
- `rails/sorbet/` — Sorbet 自動生成ファイル
- `rails/tmp/` — 一時ファイル
- `go/internal/query/*.go` — sqlc 自動生成コード
- `go/internal/templates/*_templ.go` — templ 自動生成コード
- `go/static/` — ビルド成果物
- `rails/app/assets/builds/` — ビルド成果物

### CI ワークフロー設計

プロジェクト全体のフォーマットチェック用に新しいワークフロー `fmt-ci.yml` を作成する。

```yaml
name: Formatter
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
jobs:
  oxfmt:
    name: Oxfmt
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v5
        with:
          node-version: 22
          cache: pnpm
      - run: pnpm install
      - run: pnpm oxfmt --check
```

### プロジェクトルートの Makefile

`pnpm oxfmt` や `pnpm oxfmt --check` を直接呼ぶのは冗長なため、プロジェクトルートに Makefile を配置して `make fmt` / `make fmt-check` で実行できるようにする。Go 版が既に `make fmt` を使用しているため、プロジェクト全体で統一感のある操作になる。

```makefile
.PHONY: fmt
fmt: ## コードをフォーマット（Oxfmt）
	pnpm oxfmt

.PHONY: fmt-check
fmt-check: ## フォーマットチェック（Oxfmt）
	pnpm oxfmt --check
```

Rails ディレクトリ内からは `make -C /workspace fmt` で呼び出せる。

### `bin/check` の更新

Rails の `bin/check` スクリプトで `yarn prettier . --write` をルートの Makefile 経由に置き換える。

```ruby
# 変更前
"yarn prettier . --write",

# 変更後
"make -C /workspace fmt",
```

### ERB テンプレートのフォーマット

現在 Prettier は `@ttskch/prettier-plugin-tailwindcss-anywhere` プラグインを使って ERB テンプレート内の Tailwind CSS クラスをソートしている。Oxfmt は HTML の Tailwind クラスソートをネイティブサポートしているが、ERB テンプレート（`.html.erb`）のサポートは未確認。

ERB テンプレートのフォーマット自体は `erb_lint` が担当しているため、Oxfmt のフォーマット対象からは除外する。Tailwind クラスソートについては移行後に動作確認し、対応が必要であれば別途対応する。

## 採用しなかった方針

<!--
ガイドライン:
- 検討したが採用しなかった設計や機能を、理由とともに記述
- 将来の開発者が同じ検討を繰り返さないための判断記録
- タスク完了後、この内容は `specs/` の仕様書にも転記する
- 該当がない場合は「なし」と記載
-->

### dprint の採用

Rust 製の高速フォーマッタ dprint も検討した。プラグイン方式で柔軟性が高いが、Oxfmt は Prettier との互換性が高く移行コマンドが用意されている点、Tailwind CSS クラスソートが内蔵されている点で Oxfmt を選定した。

### Rails ディレクトリのみで Oxfmt を使用する方針

Rails の `package.json` に oxfmt を追加する方針も検討したが、プロジェクト全体で統一的にフォーマットを適用するためルートレベルに配置する方針とした。

### ERB テンプレートの Tailwind クラスソートを Oxfmt で行う方針

Oxfmt は HTML の Tailwind クラスソートをネイティブサポートしているが、ERB テンプレートは標準的な HTML ではなく Ruby コードが混在するため、対応状況が不明。ERB のフォーマットは `erb_lint` が担当しているため、初期移行では ERB を Oxfmt の対象外とする。

### Prettier の `prettier-plugin-tailwindcss` を残す方針

ERB テンプレートの Tailwind クラスソート用に Prettier プラグインを残す方針も検討したが、ERB テンプレートのフォーマットは `erb_lint` が担当しており、Prettier をこの用途だけのために残すのは保守コストが高い。Oxfmt 移行後は Prettier を完全に削除する。

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

### フェーズ 1: Oxfmt のセットアップ

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
Go版/Rails版の両方を修正する場合は別タスクに分けてください
-->

- [x] **1-1**: Oxfmt のインストールと設定ファイルの作成
  - ルートレベルの `package.json` を作成し、oxfmt を devDependencies に追加
  - `pnpm install` で `pnpm-lock.yaml` を生成
  - `oxfmt --migrate=prettier` を実行して Rails の `.prettierrc` から設定を移行（必要に応じて手動調整）
  - `.oxfmtignore` を作成（自動生成ファイル、ビルド成果物、vendor 等を除外）
  - **想定ファイル数**: 約 4 ファイル（実装 4 + テスト 0）
  - **想定行数**: 約 60 行（実装 60 行 + テスト 0 行）

### フェーズ 2: フォーマットの実行と CI の更新

- [x] **2-1**: Oxfmt によるプロジェクト全体のフォーマット実行
  - `pnpm oxfmt` を実行してプロジェクト全体をフォーマット
  - 差分を確認し、意図しない変更がないか検証
  - ※広範囲のフォーマット変更のため、ファイル数・行数の制限は例外扱い
  - **想定ファイル数**: 多数（フォーマット変更のみ）
  - **想定行数**: 多数（フォーマット変更のみ、自動整形による差分）

- [x] **2-2**: フォーマット CI ワークフローの追加と Rails CI の更新
  - `.github/workflows/fmt-ci.yml` を新規作成（プロジェクト全体のフォーマットチェック）
  - `.github/workflows/rails-ci.yml` から `prettier` ジョブを削除
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 40 行（実装 40 行 + テスト 0 行）

### フェーズ 3: Prettier の削除とクリーンアップ

- [x] **3-1**: [Rails] Prettier の削除とスクリプトの更新
  - `rails/package.json` から `prettier`, `prettier-plugin-tailwindcss`, `@ttskch/prettier-plugin-tailwindcss-anywhere` を削除
  - `yarn install` でロックファイルを更新
  - `rails/.prettierrc` を削除
  - `rails/.prettierignore` を削除
  - `rails/bin/check` の `yarn prettier . --write` を Oxfmt のコマンドに置き換え
  - **想定ファイル数**: 約 5 ファイル（実装 5 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

### フェーズ 4: ドキュメント更新

- [x] **4-1**: CLAUDE.md とガイドラインの更新、プロジェクトルート Makefile の追加
  - `/workspace/Makefile` を新規作成（`make fmt` / `make fmt-check` ターゲット）
  - `/workspace/CLAUDE.md`: コミット前のチェックセクションを更新（Prettier → `make fmt` / `make fmt-check`）
  - `/workspace/rails/CLAUDE.md`: Prettier コマンドの記述を `make -C /workspace fmt` / `make -C /workspace fmt-check` に置き換え
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

### フェーズ 5: 仕様書への反映

<!--
**重要**: 実装完了後、必ず仕様書を作成・更新してください。
- 新しい機能の場合: `docs/specs/` に仕様書を新規作成する
- 既存機能の変更の場合: 対応する仕様書を最新の状態に更新する
- 概要・仕様・設計・採用しなかった方針を作業計画書から転記・整理する
-->

- [x] **5-1**: 仕様書の作成
  - `docs/specs/formatter.md` にフォーマッタの仕様書を作成
  - 作業計画書の概要・要件・設計・採用しなかった方針を仕様書に反映する

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **ERB テンプレートの Tailwind CSS クラスソート**: Oxfmt の ERB 対応状況が未確認のため。ERB のフォーマットは引き続き `erb_lint` が担当。必要に応じて別途対応
- **Go プロジェクトの `package.json` 統合**: Go の `package.json` は Tailwind CSS / esbuild 用であり、今回の移行対象外。Go のフロントエンドビルドは現状のまま pnpm で管理

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Oxfmt 公式ドキュメント - Usage](https://oxc.rs/docs/guide/usage/formatter)
- [Oxfmt Beta アナウンス](https://oxc.rs/blog/2026-02-24-oxfmt-beta)
- [Oxfmt 設定リファレンス](https://oxc.rs/docs/guide/usage/formatter/config-file-reference)
- [Prettier から Oxfmt への移行ガイド](https://oxc.rs/docs/guide/usage/formatter/migrate-from-prettier)
