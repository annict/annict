# ワークスペースディレクトリ名の統一 設計書

## 実装ガイドラインの参照

この設計書はドキュメントと設定ファイルの修正であり、Go版・Rails版のアプリケーションコードの変更は伴いません。

## 概要

全プロジェクトのマウントポイントを `/workspace/` から `/{プロジェクト名}/` にリネームし、ディレクトリ名を統一する。

現在、各プロジェクトは Docker Compose で自プロジェクトを `/workspace/` にマウントしているが、他プロジェクトからは `/mewst/`、`/wikino/`、`/groobb/` のようにプロジェクト名で参照している。自プロジェクトのマウントポイントもプロジェクト名に統一することで、クロスプロジェクトのコマンドやドキュメントがシンプルになる。

**目的**:

- 全プロジェクトで `/{プロジェクト名}/` という統一的な命名規則にする
- `/sync-guideline` 等のクロスプロジェクトコマンドで「自プロジェクト」の特別扱いを不要にする

**背景**:

- 全プロジェクトが Docker Compose で `/workspace/` を使っているが、他プロジェクトからの参照はプロジェクト名（`/annict/`、`/mewst/` 等）になっており不整合がある
- クロスプロジェクトのドキュメントやコマンドで `/workspace/` と書く必要があり、直感的でない

## 要件

### 機能要件

- 各プロジェクトの Docker Compose のボリュームマウントを `/workspace` から `/{プロジェクト名}` に変更する
- 各プロジェクト内のドキュメントで `/workspace/` を参照している箇所を `/{プロジェクト名}/` に更新する
- 変更後も各プロジェクトの開発環境が正常に動作する

## 設計

### 各プロジェクトの変更内容

各プロジェクトで共通して以下の変更を行う：

1. **docker-compose.yml**: ボリュームマウントと working_dir を `/workspace` から `/{プロジェクト名}` に変更
2. **ドキュメント内の `/workspace/` 参照**: 自プロジェクトを指す `/workspace/` を `/{プロジェクト名}/` に置換

#### Annict（`/workspace/` → `/annict/`）

**設定ファイル（1ファイル）**:

| ファイル             | 変更箇所                                                                                  |
| -------------------- | ----------------------------------------------------------------------------------------- |
| `docker-compose.yml` | `.:/workspace` → `.:/annict`、`working_dir: /workspace` → `working_dir: /annict`（2箇所） |

**ドキュメント（7ファイル）**:

| ファイル                             | 参照箇所数 |
| ------------------------------------ | ---------- |
| `CLAUDE.md`                          | 4箇所      |
| `go/CLAUDE.md`                       | 7箇所      |
| `rails/CLAUDE.md`                    | 2箇所      |
| `docs/designs/template.md`           | 11箇所     |
| `docs/reviews/template.md`           | 14箇所     |
| `go/docs/templ-guide.md`             | 1箇所      |
| `.claude/commands/sync-guideline.md` | 1箇所      |

#### mewst（`/workspace/` → `/mewst/`）

**設定ファイル（1ファイル）**:

| ファイル             | 変更箇所                                                                                |
| -------------------- | --------------------------------------------------------------------------------------- |
| `docker-compose.yml` | `.:/workspace` → `.:/mewst`、`working_dir: /workspace` → `working_dir: /mewst`（4箇所） |

**ドキュメント（19ファイルが `/workspace/` を参照）**:

更新対象（テンプレート・ガイドライン・進行中の設計書）:

| ファイル                    | 種別           |
| --------------------------- | -------------- |
| `CLAUDE.md`                 | ガイドライン   |
| `go/CLAUDE.md`              | ガイドライン   |
| `rails/CLAUDE.md`           | ガイドライン   |
| `docs/designs/template.md`  | テンプレート   |
| `docs/reviews/template.md`  | テンプレート   |
| `go/docs/templ-guide.md`    | ガイドライン   |
| `docs/designs/1_doing/*.md` | 進行中の設計書 |
| `docs/designs/2_todo/*.md`  | 未着手の設計書 |

更新しない（完了済みの設計書・レビュー記録は履歴として残す）:

- `docs/designs/3_done/**/*.md`
- `docs/reviews/done/**/*.md`

#### wikino（`/workspace/` → `/wikino/`）

**設定ファイル（1ファイル）**:

| ファイル             | 変更箇所                                                                                  |
| -------------------- | ----------------------------------------------------------------------------------------- |
| `docker-compose.yml` | `.:/workspace` → `.:/wikino`、`working_dir: /workspace` → `working_dir: /wikino`（2箇所） |

**ドキュメント（24ファイルが `/workspace/` を参照）**:

mewst と同様の方針で更新対象を選定する。

更新対象: CLAUDE.md 系、テンプレート、ガイドライン、進行中・未着手の設計書

更新しない: 完了済みの設計書・レビュー記録

#### groobb（`/workspace/` → `/groobb/`）

**設定ファイル（1ファイル）**:

| ファイル             | 変更箇所                                                                                  |
| -------------------- | ----------------------------------------------------------------------------------------- |
| `docker-compose.yml` | `.:/workspace` → `.:/groobb`、`working_dir: /workspace` → `working_dir: /groobb`（2箇所） |

**ドキュメント**: `/workspace/` を参照するドキュメントなし。設定ファイルの変更のみ。

### 置換ルール

各プロジェクトで以下の文字列置換を行う：

- `/workspace/` → `/{プロジェクト名}/`
- `/workspace` → `/{プロジェクト名}`（末尾スラッシュなしのケース）

## タスクリスト

### フェーズ 1: Annict

- [ ] **1-1**: Annict の docker-compose.yml のマウントポイントを `/workspace` → `/annict` に変更
  - ボリュームマウント `.:/workspace` → `.:/annict`
  - `working_dir: /workspace` → `working_dir: /annict`
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 4 行（実装 4 行 + テスト 0 行）

- [ ] **1-2**: Annict のドキュメント・コマンドファイル内の `/workspace/` 参照を `/annict/` に更新
  - `CLAUDE.md`、`go/CLAUDE.md`、`rails/CLAUDE.md`
  - `docs/designs/template.md`、`docs/reviews/template.md`
  - `go/docs/templ-guide.md`
  - `.claude/commands/sync-guideline.md`
  - **想定ファイル数**: 約 7 ファイル（実装 7 + テスト 0）
  - **想定行数**: 約 40 行（実装 40 行 + テスト 0 行）

### フェーズ 2: mewst

- [ ] **2-1**: mewst の docker-compose.yml のマウントポイントを `/workspace` → `/mewst` に変更
  - ボリュームマウント `.:/workspace` → `.:/mewst`
  - `working_dir: /workspace` → `working_dir: /mewst`
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 8 行（実装 8 行 + テスト 0 行）

- [ ] **2-2**: mewst のドキュメント内の `/workspace/` 参照を `/mewst/` に更新
  - CLAUDE.md 系（3ファイル）、テンプレート（2ファイル）、ガイドライン（1ファイル）
  - 進行中・未着手の設計書
  - 完了済みの設計書・レビュー記録は更新しない
  - **想定ファイル数**: 約 10 ファイル（実装 10 + テスト 0）
  - **想定行数**: 約 60 行（実装 60 行 + テスト 0 行）

### フェーズ 3: wikino

- [ ] **3-1**: wikino の docker-compose.yml のマウントポイントを `/workspace` → `/wikino` に変更
  - ボリュームマウント `.:/workspace` → `.:/wikino`
  - `working_dir: /workspace` → `working_dir: /wikino`
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 4 行（実装 4 行 + テスト 0 行）

- [ ] **3-2**: wikino のドキュメント内の `/workspace/` 参照を `/wikino/` に更新
  - CLAUDE.md 系（3ファイル）、テンプレート（2ファイル）、ガイドライン（1ファイル）
  - 進行中・未着手の設計書
  - 完了済みの設計書・レビュー記録は更新しない
  - **想定ファイル数**: 約 10 ファイル（実装 10 + テスト 0）
  - **想定行数**: 約 60 行（実装 60 行 + テスト 0 行）

### フェーズ 4: groobb

- [ ] **4-1**: groobb の docker-compose.yml のマウントポイントを `/workspace` → `/groobb` に変更
  - ボリュームマウント `.:/workspace` → `.:/groobb`
  - `working_dir: /workspace` → `working_dir: /groobb`
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 4 行（実装 4 行 + テスト 0 行）

### 実装しない機能（スコープ外）

以下は今回の実装では**実装しません**：

- **完了済みドキュメントの更新**: `3_done/` や `done/` 配下の履歴ドキュメントは更新しない（過去の記録として保持）
- **アプリケーションコードの変更**: Go/Rails のソースコード内に `/workspace/` への依存はないため対象外
