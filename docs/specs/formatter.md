# コードフォーマッター 仕様書

<!--
このテンプレートの使い方:
1. 操作対象のモデルに対応するディレクトリを `docs/specs/` 配下に作成（例: `docs/specs/page/`）
2. このファイルをそのディレクトリにコピー（例: cp docs/specs/template.md docs/specs/page/create.md）
3. [機能名] などのプレースホルダーを実際の内容に置き換え
4. 各セクションのガイドラインに従って記述
5. コメント（ `\<!-- ... --\>` ）はガイドラインとして残してください

**ファイルの配置ルール**:
- 仕様書は操作対象のモデル（名詞）ごとにディレクトリを分け、機能（動詞）をファイル名にする
  - 例: `docs/specs/user/sign-up.md`、`docs/specs/page/create.md`
- モデルに分類しにくい横断的な機能は、その機能自体を名詞としてディレクトリにする
  - 例: `docs/specs/search/full-text.md`
- モデルの定義・状態遷移・他モデルとの関係を記述する場合は `overview.md` を作成する
  - `overview.md` はモデルの静的な性質（「これは何か」）を書く場所
  - 操作に紐づく仕様（バリデーション、権限など）は各機能の仕様書に書く
- 詳細は [@docs/README.md](/workspace/docs/README.md) を参照

**仕様書の性質**:
- 仕様書は「現在のシステムの状態」を記述するドキュメントです
- 実装が完了したら、仕様書を最新の状態に更新してください
- 過去の状態はGit履歴で参照できるため、仕様書には常に現在の状態のみを記述します

**作業計画書との関係**:
- 新しい機能の場合: `docs/plans/` の作業計画書に概要・要件・設計を記述し、タスク完了後にこの仕様書を作成します
- 既存機能の変更の場合: `docs/plans/` の作業計画書に変更内容を記述し、タスク完了後にこの仕様書を更新します

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 概要

<!--
ガイドライン:
- この機能が現在「どのように動いているか」を簡潔に説明
- なぜこの仕組みになっているかの背景も記述
- 2-3段落程度で簡潔に
-->

プロジェクト全体のコードフォーマッターとして [Oxfmt](https://oxc.rs/docs/guide/usage/formatter) を使用している。Oxfmt はプロジェクトルートの `package.json` で pnpm パッケージとして管理されており、ルートから実行することで Go 版・Rails 版の両方のファイルを統一的にフォーマットする。

**目的**:

- プロジェクト全体で統一的なコードフォーマットを維持する
- JavaScript, TypeScript, CSS, SCSS, YAML, Markdown, TOML, JSON ファイルを一括フォーマットする
- CI でフォーマットチェックを自動実行し、フォーマットの乱れを防止する

**背景**:

- Oxfmt は Rust 製の高速フォーマッターで、Prettier の約 30 倍高速に動作する
- Prettier 100% 互換（JS/TS）で、`oxfmt --migrate=prettier` による設定移行が可能
- Tailwind CSS クラスソートが内蔵されており、プラグインが不要
- Go プロジェクトの TOML ファイルもフォーマット対象にできるため、プロジェクト全体の統一管理が可能

## 仕様

<!--
ガイドライン:
- 現在のシステムの振る舞いを記述
- 「ユーザーは〇〇できる」「システムは〇〇する」という形式で記述
- 必要に応じて非機能的な仕様（セキュリティ、パフォーマンスなど）も記述
-->

- 開発者はプロジェクトルートで `make fmt` を実行してプロジェクト全体のファイルをフォーマットできる
- 開発者はプロジェクトルートで `make fmt-check` を実行してフォーマットチェックを行える
- Rails ディレクトリ内からは `make -C /workspace fmt` で実行できる
- CI（GitHub Actions）がプッシュおよび Pull Request 時に自動的にフォーマットチェックを実行する

### フォーマット対象

- `*.js`, `*.jsx`, `*.ts`, `*.tsx` — JavaScript / TypeScript
- `*.css`, `*.scss` — スタイルシート
- `*.json`, `*.jsonc` — JSON
- `*.yaml`, `*.yml` — YAML
- `*.md` — Markdown
- `*.toml` — TOML（Go の翻訳ファイルなど）

### フォーマット除外対象

以下のディレクトリ・ファイルはフォーマット対象から除外される（`.oxfmtrc.json` の `ignorePatterns` で設定）:

| パターン                           | 理由                                      |
| ---------------------------------- | ----------------------------------------- |
| `**/node_modules/`                 | 依存パッケージ                            |
| `**/vendor/`                       | バンドルされた Gem                        |
| `rails/db/`                        | DB スキーマ・マイグレーション（自動生成） |
| `rails/app/docs/`                  | ドキュメント用テンプレート                |
| `rails/app/views/`                 | ERB/Slim テンプレート（erb_lint が担当）  |
| `rails/sorbet/`                    | Sorbet 自動生成ファイル                   |
| `rails/tmp/`                       | 一時ファイル                              |
| `go/internal/query/*.go`           | sqlc 自動生成コード                       |
| `go/internal/templates/*_templ.go` | templ 自動生成コード                      |
| `go/static/`                       | ビルド成果物                              |
| `rails/app/assets/builds/`         | ビルド成果物                              |
| `rails/public/`                    | 静的ファイル（ミニファイ済み JS 含む）    |

### フォーマット設定

| 設定項目          | 値    | 説明                        |
| ----------------- | ----- | --------------------------- |
| `printWidth`      | 120   | 1 行の最大文字数            |
| `tabWidth`        | 2     | インデント幅                |
| `semi`            | true  | セミコロンを付ける          |
| `singleQuote`     | false | ダブルクォートを使用        |
| `trailingComma`   | all   | 末尾カンマを付ける          |
| `sortTailwindcss` | 有効  | Tailwind CSS クラスソート   |
| `sortPackageJson` | false | package.json のソートは無効 |

## 設計

<!--
ガイドライン:
- 現在の技術的な実装の詳細を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - 技術スタック（使用するライブラリ、フレームワーク、ツールなど）
  - アーキテクチャ（システム全体の構成、コンポーネント間の関係など）
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - セキュリティ設計（認証・認可、トークン管理、Rate Limitingなど）
  - コード設計（パッケージ構成、主要な構造体、インターフェースなど）
- 該当がない場合も、セクション自体は残しておく（後から追加しやすくするため）
-->

### ファイル構成

```
/workspace/
├── package.json          # oxfmt の依存関係（pnpm で管理）
├── pnpm-lock.yaml        # pnpm のロックファイル
├── .oxfmtrc.json         # Oxfmt 設定ファイル
├── Makefile              # make fmt / make fmt-check
└── .github/
    └── workflows/
        └── fmt-ci.yml    # CI フォーマットチェック
```

### パッケージマネージャー

プロジェクトルートでは **pnpm** を使用する。Go プロジェクトが pnpm を使用しているため統一。Rails プロジェクト（`rails/`）は引き続き **yarn** を使用する。

### Makefile

プロジェクトルートの `Makefile` に以下のターゲットを定義:

- `make fmt` — `pnpm run fmt`（`oxfmt --write`）を実行
- `make fmt-check` — `pnpm run fmt:check`（`oxfmt --check`）を実行

### CI ワークフロー

`.github/workflows/fmt-ci.yml` で以下を実行:

- `main` ブランチへの push および Pull Request で実行
- pnpm のセットアップ → `pnpm install` → `pnpm run fmt:check`

### Rails の bin/check との統合

`rails/bin/check` スクリプトから `make -C /workspace fmt` を呼び出し、Rails の検証フローに統合。

### ERB テンプレートのフォーマット

ERB テンプレート（`.html.erb`）のフォーマットは Oxfmt の対象外。ERB のフォーマットは `erb_lint` が担当する。Tailwind CSS クラスソートについては、ERB テンプレート内での対応は今後検討。

## 採用しなかった方針

<!--
ガイドライン:
- 検討したが採用しなかった設計や機能を、理由とともに記述
- 将来の開発者が同じ検討を繰り返さないための判断記録として活用する
- 後から実装された場合は、該当項目を削除する
- 該当がない場合も、セクション自体は残しておく（後から追加しやすくするため）
-->

### dprint の採用

Rust 製の高速フォーマッタ dprint も検討した。プラグイン方式で柔軟性が高いが、Oxfmt は Prettier との互換性が高く移行コマンドが用意されている点、Tailwind CSS クラスソートが内蔵されている点で Oxfmt を選定した。

### Rails ディレクトリのみで Oxfmt を使用する方針

Rails の `package.json` に oxfmt を追加する方針も検討したが、プロジェクト全体で統一的にフォーマットを適用するためルートレベルに配置する方針とした。

### ERB テンプレートの Tailwind クラスソートを Oxfmt で行う方針

Oxfmt は HTML の Tailwind クラスソートをネイティブサポートしているが、ERB テンプレートは標準的な HTML ではなく Ruby コードが混在するため、対応状況が不明。ERB のフォーマットは `erb_lint` が担当しているため、ERB を Oxfmt の対象外とした。

### Prettier の `prettier-plugin-tailwindcss` を残す方針

ERB テンプレートの Tailwind クラスソート用に Prettier プラグインを残す方針も検討したが、ERB テンプレートのフォーマットは `erb_lint` が担当しており、Prettier をこの用途だけのために残すのは保守コストが高いため不採用とした。

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Oxfmt 公式ドキュメント - Usage](https://oxc.rs/docs/guide/usage/formatter)
- [Oxfmt Beta アナウンス](https://oxc.rs/blog/2026-02-24-oxfmt-beta)
- [Oxfmt 設定リファレンス](https://oxc.rs/docs/guide/usage/formatter/config-file-reference)
- [Prettier から Oxfmt への移行ガイド](https://oxc.rs/docs/guide/usage/formatter/migrate-from-prettier)
