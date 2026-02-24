# Go モジュールパス修正 設計書

## 概要

Go 版 Annict の `go.mod` ファイルに記載されているモジュールパスを、モノレポのディレクトリ構造に合わせて修正します。

現在のモジュールパス:

```
module github.com/annict/annict
```

修正後のモジュールパス:

```
module github.com/annict/annict/go
```

**目的**:

- Go Modules の慣習に従った正しいモジュールパスに修正する
- 外部からの `go get` が正しく動作するようにする
- 将来的にパッケージを公開する際の問題を回避する

**背景**:

- 現在、Go プロジェクトはモノレポ内の `/workspace/go/` サブディレクトリに配置されている
- `go get github.com/annict/annict` を実行すると、リポジトリのルートを取得しようとするが、`go.mod` は `go/` ディレクトリにあるため正しく動作しない
- Go Modules の仕様上、モジュールパスとリポジトリ内の位置は一致すべきである

## 要件

### 機能要件

- `go.mod` のモジュールパスを `github.com/annict/annict/go` に変更する
- すべての Go ソースファイルの import 文を新しいモジュールパスに更新する
- `.golangci.yml` の depguard 設定を新しいモジュールパスに更新する
- ドキュメント内のモジュールパス参照を更新する
- 変更後もすべてのテストが通ることを確認する
- 変更後もビルドが成功することを確認する

### 非機能要件

- **保守性**: 一括置換で対応可能な変更であること
- **互換性**: 内部使用のみのため、外部への破壊的変更は考慮不要

## 設計

### 変更対象ファイル

影響範囲の調査結果：

| カテゴリ            | ファイル数 | 変更箇所数 |
| ------------------- | ---------- | ---------- |
| `go.mod`            | 1          | 1          |
| `.golangci.yml`     | 1          | 43         |
| `cmd/`              | 2          | 28         |
| `internal/`         | 約 177     | 約 560     |
| `docs/`             | 6          | 約 11      |
| **合計**            | **187**    | **約 643** |

### 変更方法

`sed` コマンドによる一括置換を使用：

```sh
# Go ソースファイル (.go) の import 文を置換
find /workspace/go -name "*.go" -exec sed -i 's|github.com/annict/annict/internal|github.com/annict/annict/go/internal|g' {} \;

# templ ファイル (.templ) の import 文を置換
find /workspace/go -name "*.templ" -exec sed -i 's|github.com/annict/annict/internal|github.com/annict/annict/go/internal|g' {} \;

# go.mod のモジュールパスを置換
sed -i 's|module github.com/annict/annict$|module github.com/annict/annict/go|' /workspace/go/go.mod

# .golangci.yml の depguard 設定を置換
sed -i 's|github.com/annict/annict/internal|github.com/annict/annict/go/internal|g' /workspace/go/.golangci.yml

# ドキュメントの参照を置換
find /workspace/go/docs -name "*.md" -exec sed -i 's|github.com/annict/annict/internal|github.com/annict/annict/go/internal|g' {} \;
```

### 検証方法

1. `go mod tidy` が正常に完了すること
2. `make fmt` が正常に完了すること
3. `make lint` が正常に完了すること
4. `go build ./...` が正常に完了すること
5. `make test` が正常に完了すること

## タスクリスト

### フェーズ 1: モジュールパスの修正

- [x] **1-1**: Go ソースファイルの import 文を一括置換

  - `.go` ファイルの import 文を `github.com/annict/annict/internal` から `github.com/annict/annict/go/internal` に変更
  - `.templ` ファイルの import 文も同様に変更
  - `go.mod` のモジュールパスを変更
  - **想定ファイル数**: 約 180 ファイル（実装 180 + テスト 0）
  - **想定行数**: 約 600 行（実装 600 行 + テスト 0 行）
  - **備考**: 一括置換のため、実質的な作業は sed コマンドの実行のみ

- [x] **1-2**: 設定ファイルとドキュメントの更新

  - `.golangci.yml` の depguard 設定を更新
  - `docs/` 配下の Markdown ファイルを更新
  - **想定ファイル数**: 約 7 ファイル（実装 7 + テスト 0）
  - **想定行数**: 約 54 行（実装 54 行 + テスト 0 行）

- [x] **1-3**: templ コードの再生成と検証
  - `make templ-generate` で templ コードを再生成
  - `go mod tidy` で依存関係を整理
  - `make fmt` でフォーマット
  - `make lint` でリントチェック
  - `go build ./...` でビルド確認
  - `make test` でテスト実行
  - **想定ファイル数**: 約 15 ファイル（自動生成される `*_templ.go` ファイル）
  - **想定行数**: 約 45 行（自動生成されるファイルの import 文）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **外部公開の考慮**: このモジュールは内部使用のみのため、外部への破壊的変更対応は不要
- **バージョニング**: パッケージの公開予定がないため、バージョンタグの付与は不要

## 参考資料

- [Go Modules Reference](https://go.dev/ref/mod)
- [Go Wiki: Modules](https://go.dev/wiki/Modules)
