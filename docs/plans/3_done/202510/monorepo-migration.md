# モノレポ化対応 設計書

## 概要

Rails版とGo版を一つのリポジトリで管理するモノレポ化が行われたことにより、CI設定とドキュメント（CLAUDE.md）の構造が実際のプロジェクト構造と合わなくなっています。これを修正し、モノレポ構造に適した形に再構成します。

**目的**:

- Go版とRails版のCIを正常に動作させる
- CLAUDE.mdをモノレポ構造に合わせて再構成し、プロジェクト全体の情報とサブプロジェクト固有の情報を適切に分離する
- 開発者がプロジェクト構造を理解しやすくする

**背景**:

- `/workspace/go/` と `/workspace/rails/` という形でモノレポ化されたが、CI設定がこの構造に対応していない
- Go版のCI設定が `/workspace/go/.github/workflows/ci.yml` にあり、GitHub Actionsから認識されない
- Rails版のCI設定がモノレポ構造を考慮せず、パス指定が間違っている
- `/workspace/go/CLAUDE.md` にプロジェクト全体に関係する情報とGo版固有の情報が混在している
- `/workspace/rails/CLAUDE.md` がほぼ空で、Rails版の開発ガイドが不足している

## 要件

### 機能要件

- Go版のCIが正常に動作する（lint、test、buildの全ジョブが成功する）
- Rails版のCIが正常に動作する（zeitwerk、sorbet、standard、erb_lint、eslint、rspecの全ジョブが成功する）
- ルートの `/workspace/CLAUDE.md` でプロジェクト全体の概要とモノレポ構造を説明する
- `/workspace/go/CLAUDE.md` でGo版固有の開発ガイドを提供する
- `/workspace/rails/CLAUDE.md` でRails版固有の開発ガイドを提供する
- 各CLAUDE.mdファイルが相互にリンクされており、開発者が必要な情報に辿り着ける

### 非機能要件

- **保守性**: CI設定とドキュメントがプロジェクト構造と一致しており、将来的な変更に対応しやすい
- **可読性**: 開発者が迷わず必要な情報を見つけられる構造
- **再利用性**: Go版とRails版のCIジョブが独立しており、個別に実行・メンテナンスできる

## 設計

### CI設計

#### 現状の問題

1. **Go版CI**: `/workspace/go/.github/workflows/ci.yml` にあるが、GitHub Actionsはリポジトリルートの `.github/workflows/` のみを認識する
2. **Rails版CI**: `/workspace/.github/workflows/lint-and-test.yml` にあるが、パス設定がモノレポ構造を考慮していない（例: `bin/rails` → `rails/bin/rails` が必要）

#### 解決策

1. **統合CI設定ファイルの作成**:
   - `/workspace/.github/workflows/go-ci.yml` - Go版専用のCI
   - `/workspace/.github/workflows/rails-ci.yml` - Rails版専用のCI（既存の `lint-and-test.yml` をリネーム）

2. **パス設定の修正**:
   - Go版CI: `working-directory: go` を各ステップに追加
   - Rails版CI: `working-directory: rails` を各ステップに追加、またはコマンドのパスを修正

3. **既存ファイルの削除**:
   - `/workspace/go/.github/workflows/ci.yml` を削除（不要になる）

#### CI設定の詳細

**Go版CI (`/workspace/.github/workflows/go-ci.yml`)**:

- トリガー: pushとpull_request（Go関連ファイルの変更時のみ）
- ジョブ: lint、test、build
- 各ステップで `working-directory: go` を指定
- パス指定: `db/schema.sql` → `go/db/schema.sql`

**Rails版CI (`/workspace/.github/workflows/rails-ci.yml`)**:

- トリガー: pushとpull_request（Rails関連ファイルの変更時のみ）
- ジョブ: zeitwerk、sorbet、standard、erb_lint、eslint、rspec
- 各ステップで `working-directory: rails` を指定
- コマンドパス: `bin/rails` は `working-directory: rails` により正しく動作

### Dependabot設計

#### 現状の問題

1. **依存関係ファイルのパス**: 現在の設定ではルートディレクトリ (`/`) を参照しているが、実際のファイルは `go/` と `rails/` に分散している
2. **Go版の依存関係管理が未設定**: Go Modulesの更新が自動化されていない

#### 解決策

Dependabotの設定を以下のように更新：

**`/workspace/.github/dependabot.yml`**:

- **gomod**: `/go` ディレクトリの `go.mod` を監視（新規追加）
- **npm (Go版)**: `/go` ディレクトリの `package.json` を監視（pnpmもnpmとして扱える）
- **npm (Rails版)**: `/rails` ディレクトリの `package.json` を監視
- **bundler**: `/rails` ディレクトリの `Gemfile` を監視
- **github-actions**: `/` ディレクトリ（変更なし）

各パッケージエコシステムごとに適切なディレクトリを指定することで、Go版とRails版の依存関係を独立して管理できます。

### CLAUDE.md設計

#### 現状の問題

1. `/workspace/CLAUDE.md` - プロジェクト全体の説明が薄い
2. `/workspace/go/CLAUDE.md` - プロジェクト全体とGo版固有の情報が混在
3. `/workspace/rails/CLAUDE.md` - ほぼ空

#### 解決策: 3層構造

```
/workspace/CLAUDE.md (Layer 1: プロジェクト全体)
├── 概要（Annictとは何か）
├── モノレポ構造の説明
├── 共通の開発環境セットアップ
├── 共通のインフラ（PostgreSQL、Redis、MinIO、imgproxyなど）
└── 各サブプロジェクトへのリンク

/workspace/go/CLAUDE.md (Layer 2: Go版固有)
├── Go版プロジェクト構造
├── Go版の技術スタック
├── Go版の開発環境セットアップ
├── Go版のコーディング規約
├── Go版のテスト戦略
└── ルートCLAUDE.mdへのリンク

/workspace/rails/CLAUDE.md (Layer 3: Rails版固有)
├── Rails版プロジェクト構造
├── Rails版の技術スタック
├── Rails版の開発環境セットアップ
├── Rails版のコーディング規約
├── Rails版のテスト戦略
└── ルートCLAUDE.mdへのリンク
```

#### 移行方針

1. **プロジェクト全体の情報を抽出**: 現在の `/workspace/go/CLAUDE.md` から以下を抽出して `/workspace/CLAUDE.md` に移動
   - Annictの概要
   - RailsからGoへの移行について（これはプロジェクト全体の文脈）
   - 画像配信アーキテクチャ（Rails版とGo版で共通）
   - 共通のインフラ（PostgreSQL、Redis、MinIO、imgproxy）

2. **Go版固有の情報を整理**: `/workspace/go/CLAUDE.md` に残す内容
   - Go版の技術スタック（chi、sqlcなど）
   - Go版のプロジェクト構造（internal/handler、internal/usecaseなど）
   - Go版のコーディング規約
   - Go版のテスト戦略

3. **Rails版の情報を充実**: `/workspace/rails/CLAUDE.md` に追加する内容
   - Rails版の技術スタック
   - Rails版のプロジェクト構造
   - Rails版のコーディング規約
   - Rails版のテスト戦略

### ディレクトリ構造

```
/workspace/
├── .github/
│   ├── workflows/
│   │   ├── go-ci.yml          # Go版CI（新規作成）
│   │   └── rails-ci.yml       # Rails版CI（lint-and-test.ymlからリネーム）
│   └── dependabot.yml         # Dependabot設定（モノレポ対応に更新）
├── CLAUDE.md                  # プロジェクト全体のガイド（更新）
├── go/
│   ├── .github/               # 削除する
│   │   └── workflows/
│   │       └── ci.yml         # 削除する（不要になる）
│   └── CLAUDE.md              # Go版固有のガイド（更新）
└── rails/
    └── CLAUDE.md              # Rails版固有のガイド（充実させる）
```

## タスクリスト

### フェーズ 1: CI設定の修正

- [x] Go版CI設定ファイルの作成
  - `/workspace/go/.github/workflows/ci.yml` を `/workspace/.github/workflows/go-ci.yml` として移動・修正
  - 各ステップに `working-directory: go` を追加
  - パストリガーを追加（Go関連ファイルの変更時のみ実行）
  - データベーススキーマパスを `go/db/schema.sql` に修正
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 160 行（実装 160 行、既存ファイルの移動・修正）

- [x] Rails版CI設定ファイルの修正
  - `/workspace/.github/workflows/lint-and-test.yml` を `/workspace/.github/workflows/rails-ci.yml` にリネーム
  - 各ジョブのステップに `working-directory: rails` を追加
  - パストリガーを追加（Rails関連ファイルの変更時のみ実行）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 180 行（実装 180 行、既存ファイルの修正）

- [x] Dependabot設定の更新
  - `/workspace/.github/dependabot.yml` をモノレポ構造に対応させる
  - Go版の依存関係管理を追加（gomod: `/go`、npm: `/go`）
  - Rails版の依存関係管理を修正（bundler: `/rails`、npm: `/rails`）
  - GitHub Actionsの設定は維持（directory: `/`）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行（実装 30 行、既存ファイルの修正）

- [x] 不要なCI設定ファイルの削除
  - `/workspace/go/.github/` ディレクトリ全体を削除
  - **想定ファイル数**: 約 0 ファイル（削除のみ）
  - **想定行数**: 約 0 行（削除のみ）

### フェーズ 2: CLAUDE.mdの再構成

- [x] ルートCLAUDE.mdの更新
  - プロジェクト全体の概要を充実
  - モノレポ構造の説明を追加
  - RailsからGoへの移行の背景を追加
  - 共通インフラ（PostgreSQL、Redis、MinIO、imgproxy）の説明を追加
  - 開発環境のセットアップ（Docker Composeなど）を追加
  - Go版とRails版へのリンクを追加
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 150 行（実装 150 行）

- [x] Go版CLAUDE.mdの更新
  - プロジェクト全体の情報を削除（ルートCLAUDE.mdに移動）
  - Go版固有の情報のみを残す
  - ルートCLAUDE.mdへのリンクを追加
  - モノレポ構造に合わせてパスやコマンドを修正
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 200 行（実装 200 行、既存内容の整理）

- [x] Rails版CLAUDE.mdの充実
  - Rails版の技術スタック説明を追加
  - Rails版のプロジェクト構造説明を追加
  - Rails版の開発環境セットアップを追加
  - Rails版のコーディング規約を追加
  - Rails版のテスト戦略を追加
  - ルートCLAUDE.mdへのリンクを追加
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 200 行（実装 200 行）

### フェーズ 3: 動作確認とドキュメント整備

- [ ] CI動作確認
  - Go版CIが正常に動作することを確認（テストPRを作成して確認）
  - Rails版CIが正常に動作することを確認（テストPRを作成して確認）
  - **想定ファイル数**: 約 0 ファイル（確認のみ）
  - **想定行数**: 約 0 行（確認のみ）

- [ ] ドキュメントの最終確認
  - 各CLAUDE.mdファイルのリンクが正しく機能することを確認
  - 誤字脱字や不正確な情報がないかチェック
  - **想定ファイル数**: 約 0 ファイル（確認のみ）
  - **想定行数**: 約 0 行（確認のみ）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **CI/CDパイプラインの最適化**: 今回はCIの修正のみで、パフォーマンス最適化（キャッシュ戦略など）は対象外
- **monorepoツールの導入**: Nx、Turborepo、Lerna などのmonorepo管理ツールは導入せず、シンプルなディレクトリ構造で管理
- **共通コードの抽出**: Go版とRails版で共有できるコード（設定ファイルなど）の抽出は今回対象外

## 参考資料

- [GitHub Actions - Working Directory](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_iddefaultsrun)
- [GitHub Actions - Path Filtering](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onpushpull_requestpull_request_targetpathspaths-ignore)
- [Monorepo 設計ベストプラクティス](https://monorepo.tools/)
