# golangci-lint 導入 設計書

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

Go 版 Annict のコード品質を向上させるため、`golangci-lint` を導入します。現在、CI で個別に実行している静的解析ツール（`go fmt`, `goimports`, `go vet`, `staticcheck`）を統合し、さらにセキュリティチェック（`gosec`）とアーキテクチャルール（`depguard`）を追加します。

**目的**:

- **既存の Lint ツールの統合**: `go fmt`, `goimports`, `go vet`, `staticcheck` を golangci-lint に統合し、一つのコマンドで実行可能にする
- **セキュリティチェックの追加**: `gosec` でセキュリティ上の問題を検出する
- **アーキテクチャルールの強制**: `depguard` で 3 層アーキテクチャの依存関係ルールを強制する
- **開発効率の向上**: CI の実行時間を短縮し、開発者がローカルで統一されたチェックを実行できるようにする

**背景**:

- 現在、CI で複数の静的解析ツールを個別に実行しているため、設定が分散している
- セキュリティチェックが不足しており、潜在的な脆弱性を見逃す可能性がある
- 3 層アーキテクチャの依存関係ルール（例: Handler/UseCase が Query に直接依存することを禁止）が強制されていない
- golangci-lint は複数のリンターを並列実行し、キャッシュを活用するため、CI の実行時間を短縮できる

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

- **既存のリンターの統合**:
  - `gofmt`（フォーマットチェック）
  - `goimports`（import 文の整理チェック）
  - `govet`（Go の標準静的解析）
  - `staticcheck`（高度な静的解析）
- **新規リンターの追加**:
  - `gosec`（セキュリティチェック）
  - `depguard`（依存関係のルール強制）
- **CI での実行**:
  - GitHub Actions の Lint ジョブで golangci-lint を実行
  - 既存の個別リンター実行を golangci-lint に置き換え
- **ローカル開発での実行**:
  - `make lint` で golangci-lint を実行
  - 開発者が簡単にローカルでチェックを実行できる

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

- **パフォーマンス**:
  - CI の Lint ジョブの実行時間が現状より長くならないこと（理想的には短縮）
  - golangci-lint のキャッシュを活用して高速化
- **保守性**:
  - 設定ファイルはプロジェクトルート（`/workspace/go/.golangci.yml`）に配置
  - 設定は可読性を重視し、コメントで説明を追加
- **開発者体験**:
  - Lint エラーメッセージはわかりやすく、修正方法が明確であること
  - 既存のコードに対して過度に厳格なルールを適用しない（段階的に改善）

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

### 技術スタック

- **golangci-lint**: v2.6.2（最新の安定版）
- **有効化するリンター**:
  - `gofmt`: Go 標準フォーマッター
  - `goimports`: import 文の整理
  - `govet`: Go 標準静的解析
  - `staticcheck`: 高度な静的解析
  - `gosec`: セキュリティ脆弱性チェック
  - `depguard`: 依存関係のルール強制
  - `errcheck`: エラーチェック漏れの検出
  - `ineffassign`: 無駄な代入の検出
  - `unused`: 未使用コードの検出

### golangci-lint の設定（`.golangci.yml`）

```yaml
# golangci-lint v2 の設定ファイル
# https://golangci-lint.run/usage/configuration/

version: "2"

run:
  # タイムアウト（デフォルト: 1m）
  timeout: 5m
  # テストファイルを含める
  tests: true
  # ビルドタグ
  build-tags: []

linters:
  # 有効化するリンター
  enable:
    - govet # Go標準静的解析
    - staticcheck # 高度な静的解析
    - gosec # セキュリティチェック
    - depguard # 依存関係のルール強制
    - errcheck # エラーチェック漏れ
    - ineffassign # 無駄な代入
    - unused # 未使用コード

  # リンター除外ルール（v2形式）
  exclusions:
    rules:
      # テストファイルでは gosec を無効化（テストデータの扱いが緩いため）
      - path: '(.+)_test\.go'
        linters:
          - gosec
      # templ 自動生成ファイルはすべてのリンターから除外
      - path: '(.+)_templ\.go'
        linters:
          - govet
          - staticcheck
          - gosec
          - depguard
          - errcheck
          - ineffassign
          - unused
    # 除外するディレクトリ
    paths:
      - bin/.*
      - vendor/.*
      - static/.*
      - node_modules/.*

  # リンター固有の設定（v2形式）
  settings:
    # gosec の設定
    gosec:
      # 除外する脆弱性ルール（必要に応じて調整）
      excludes: []
      # セキュリティの重要度（low, medium, high）
      severity: medium
      confidence: medium

    # depguard の設定
    depguard:
      rules:
        # Presentation層のルール（Handler, Middleware, ViewModel, Templates）
        presentation-layer:
          # 対象ディレクトリ
          files:
            - "**/internal/handler/**/*.go"
            - "**/internal/middleware/**/*.go"
            - "**/internal/viewmodel/**/*.go"
            - "**/internal/templates/**/*.go"
          # 許可するパッケージ（サブパッケージも含む）
          allow:
            - $gostd # Go標準ライブラリ
            - github.com/annict/annict/internal/config
            - github.com/annict/annict/internal/i18n
            - github.com/annict/annict/internal/image
            - github.com/annict/annict/internal/session
            - github.com/annict/annict/internal/middleware
            - github.com/annict/annict/internal/handler
            - github.com/annict/annict/internal/viewmodel
            - github.com/annict/annict/internal/templates
            - github.com/annict/annict/internal/usecase # Application層に依存OK
            - github.com/annict/annict/internal/model # Domain/Infrastructure層に依存OK
            - github.com/annict/annict/internal/repository # Domain/Infrastructure層に依存OK
            - github.com/annict/annict/internal/testutil # テスト用ヘルパー
          # 禁止するパッケージ
          deny:
            - pkg: github.com/annict/annict/internal/query
              desc: "Presentation層はQueryに直接依存できません。Repositoryを経由してください。"

        # Application層のルール（UseCase）
        application-layer:
          files:
            - "**/internal/usecase/**/*.go"
          allow:
            - $gostd
            - github.com/annict/annict/internal/config
            - github.com/annict/annict/internal/model
            - github.com/annict/annict/internal/repository
            - github.com/annict/annict/internal/usecase
          deny:
            - pkg: github.com/annict/annict/internal/query
              desc: "Application層はQueryに直接依存できません。Repositoryを経由してください。"
            - pkg: github.com/annict/annict/internal/handler
              desc: "Application層はPresentation層に依存できません。"
            - pkg: github.com/annict/annict/internal/middleware
              desc: "Application層はPresentation層に依存できません。"
            - pkg: github.com/annict/annict/internal/viewmodel
              desc: "Application層はPresentation層に依存できません。"
            - pkg: github.com/annict/annict/internal/templates
              desc: "Application層はPresentation層に依存できません。"

        # Domain/Infrastructure層のルール（Query, Repository, Model）
        domain-infrastructure-layer:
          files:
            - "**/internal/query/**/*.go"
            - "**/internal/repository/**/*.go"
            - "**/internal/model/**/*.go"
          allow:
            - $gostd
            - github.com/annict/annict/internal/query
            - github.com/annict/annict/internal/repository
            - github.com/annict/annict/internal/model
          deny:
            - pkg: github.com/annict/annict/internal/handler
              desc: "Domain/Infrastructure層はPresentation層に依存できません。"
            - pkg: github.com/annict/annict/internal/middleware
              desc: "Domain/Infrastructure層はPresentation層に依存できません。"
            - pkg: github.com/annict/annict/internal/viewmodel
              desc: "Domain/Infrastructure層はPresentation層に依存できません。"
            - pkg: github.com/annict/annict/internal/templates
              desc: "Domain/Infrastructure層はPresentation層に依存できません。"
            - pkg: github.com/annict/annict/internal/usecase
              desc: "Domain/Infrastructure層はApplication層に依存できません。"

    # staticcheck の設定
    staticcheck:
      # 有効化するチェック（デフォルトですべて有効）
      checks: ["all"]

# フォーマッター固有の設定（v2形式）
formatters:
  # 有効化するフォーマッター
  enable:
    - gofmt # Go標準フォーマッター
    - goimports # import文の整理

  # フォーマッター除外パス
  exclusions:
    paths:
      - bin/.*
      - vendor/.*
      - static/.*
      - node_modules/.*
      - '(.+)_templ\.go' # templ 自動生成ファイル

  settings:
    # goimports の設定
    goimports:
      # ローカルパッケージのプレフィックス
      local-prefixes:
        - github.com/annict/annict/go

# 出力設定（v2形式）
output:
  formats:
    text:
      path: stdout
      colors: true
```

### Makefile への統合

`Makefile` に golangci-lint を実行するターゲットを追加します：

```makefile
# golangci-lint を実行
.PHONY: lint-golangci
lint-golangci:
	@echo "🔍 golangci-lintを実行中..."
	@golangci-lint run --config=.golangci.yml ./...

# 既存の lint ターゲットを golangci-lint に置き換え
.PHONY: lint
lint: lint-golangci
```

### GitHub Actions CI への統合

`.github/workflows/go-ci.yml` の Lint ジョブを以下のように変更します：

**変更前**:

```yaml
- name: Check go fmt
  run: |
    unformatted=$(go fmt ./...)
    if [ -n "$unformatted" ]; then
      echo "The following packages have unformatted files:"
      echo "$unformatted"
      echo ""
      echo "Please run 'make fmt' to format these files"
      exit 1
    fi

- name: Check goimports
  run: |
    make goimports-check

- name: Run go vet
  run: |
    go vet ./...

- name: Install staticcheck
  run: |
    go install honnef.co/go/tools/cmd/staticcheck@latest

- name: Run staticcheck
  run: |
    staticcheck ./...
```

**変更後**:

```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v9.0.0
  with:
    version: v2.6.2
    working-directory: go
    args: --config=.golangci.yml ./...
```

### 実装方針

- **段階的な導入**:
  - まず、既存のコードベースで golangci-lint を実行し、検出された問題を確認
  - 重大な問題（セキュリティ、バグなど）を優先的に修正
  - 軽微な問題（スタイル、未使用コードなど）は後回しにする
- **既存コードへの影響を最小化**:
  - golangci-lint の設定で除外ルールを活用し、既存コードに対して過度に厳格にならないようにする
  - 新規コードから順次適用し、既存コードは徐々に改善
- **CI への影響**:
  - golangci-lint-action はキャッシュを自動で活用するため、CI の実行時間が大幅に短縮される見込み
- **開発者への周知**:
  - CLAUDE.md を更新し、`make lint` で golangci-lint を実行できることを明記
  - コミット前に `make lint` を実行することを推奨

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

### フェーズ 1: golangci-lint のセットアップ

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: golangci-lint の設定ファイル作成と CI への統合

  - `.golangci.yml` 設定ファイルを作成
  - GitHub Actions の Lint ジョブを golangci-lint に置き換え
  - Makefile に `lint-golangci` ターゲットを追加
  - 既存の個別リンター実行を削除
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 200 行（実装 200 行 + テスト 0 行）
  - **変更ファイル**:
    - `/workspace/go/.golangci.yml`（新規作成、約 180 行）
    - `/workspace/.github/workflows/go-ci.yml`（変更、約 -50/+10 行）
    - `/workspace/go/Makefile`（変更、約 +10 行）

### フェーズ 2: 検出された問題の修正

- [x] **2-1**: gosec で検出されたセキュリティ問題の修正

  - golangci-lint を実行し、gosec で検出された問題を確認
  - 重大なセキュリティ問題を修正
  - 必要に応じて除外ルールを追加（誤検出の場合）
  - **想定ファイル数**: 約 5 ファイル（実装 5 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）
  - **依存**: タスク 1-1 完了後

- [x] **2-2**: depguard で検出されたアーキテクチャ違反の修正

  - ✅ Presentation 層が Query に直接依存する問題を修正
  - ✅ PasswordResetTokenRepository を新規作成
  - ✅ UserRepository に`GetByID`メソッドを追加
  - ✅ 3 つのハンドラーを修正（password, password_reset, popular_work, sign_in_code, sign_up_username）
  - ✅ すべてのテストファイルを修正
  - ✅ テストファイルを depguard から除外（`.golangci.yml`）
  - **実際のファイル数**: 約 18 ファイル（実装 10 + テスト 8）
  - **実際の行数**: 約 200 行（実装 150 行 + テスト 50 行）
  - **依存**: タスク 1-1 完了後

- [x] **2-2-1**: golangci-lint を go.mod と tools.go から削除

  - golangci-lint の公式ドキュメント（https://golangci-lint.run/docs/welcome/install/）によると、`go install` でのインストールは非推奨
  - バイナリを curl でインストールする方法が推奨されているため、Go の依存関係から削除する
  - `go/go.mod` から `github.com/golangci/golangci-lint/v2` を削除
  - `go/tools.go` から golangci-lint のインポートを削除
  - CI は既に `golangci-lint-action` を使用しているため影響なし
  - ローカル環境では `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.6.2` でインストール
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 5 行（実装 -5 行 + テスト 0 行）
  - **変更ファイル**:
    - `/workspace/go/go.mod`（変更、golangci-lint の依存関係を削除）
    - `/workspace/go/tools.go`（変更、golangci-lint のインポートを削除）
  - **依存**: タスク 2-2 完了後

- [x] **2-3**: その他のリンター（errcheck, ineffassign, unused）で検出された問題の修正

  - ✅ errcheck, ineffassign, unused で検出された問題を確認
  - ✅ golangci-lint を実行した結果、0 件の問題が検出された
  - ✅ コードベースは既にクリーンな状態であることを確認
  - **実際のファイル数**: 約 0 ファイル（問題なし）
  - **実際の行数**: 約 0 行（問題なし）
  - **依存**: タスク 1-1 完了後

- [x] **2-4**: depguard のパッケージ間依存関係の細かい設定

  - 現在の depguard 設定は粒度が粗く、望ましくない依存関係が許可されている
  - 例: `templates` から `repository` を import できてしまう
  - ブラックリスト方式で、各パッケージごとに細かく依存禁止ルールを設定
  - `.golangci.yml` の depguard 設定を以下のように細分化：
    - **Presentation 層内の依存関係**:
      - `templates-layer`: データアクセスとビジネスロジックに依存しない
      - `viewmodel-layer`: Model を表示用データに変換、Repository に依存しない
      - `handler-layer`: Query へのアクセスは Repository を経由
      - `middleware-layer`: 共通処理のみ、他のレイヤーに依存しない
    - **Application 層の依存関係**:
      - `application-layer`: Presentation 層に依存しない
    - **Domain/Infrastructure 層内の依存関係**:
      - `model-layer`: 純粋なドメインエンティティ、`query`, `repository` に依存しない
      - `repository-layer`: `query`, `model` に依存できる、上位層に依存しない
      - `query-layer`: sqlc 生成コード、他のすべての層に依存しない
  - `go/docs/architecture-guide.md` を更新して、パッケージ間の依存関係を明記
    - Presentation 層内のパッケージ間の依存関係セクションを追加
    - Domain/Infrastructure 層内のパッケージ間の依存関係セクションを追加
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 200 行（実装 150 行 + ドキュメント 50 行）
  - **変更ファイル**:
    - `/workspace/go/.golangci.yml`（変更、約 +100 行）
    - `/workspace/go/docs/architecture-guide.md`（変更、約 +100 行）
  - **依存**: タスク 2-2 完了後

- [x] **2-4-1**: `.golangci.yml` の depguard 設定の修正（templ 自動生成ファイルと templates-layer の依存関係）

  - **問題点1**: `*_templ.go` ファイルが depguard から除外されているため、templates-layer の依存性チェックが無効になっている
  - **問題点2**: templates-layer で viewmodel を deny しているが、実際には Template から ViewModel は参照できないといけない
  - **修正内容**:
    - ✅ `*_templ.go` ファイルを depguard の除外対象から削除（templ が自動生成するファイルでも import は開発者が制御可能）
    - ✅ templates-layer の設定で viewmodel を deny リストから削除（Template は ViewModel のデータを表示するために参照する必要がある）
    - ✅ templates-layer で model を deny（Template は ViewModel を経由して Model を表示すべき）
    - ✅ `go/docs/architecture-guide.md` を更新して、Templates は ViewModel に依存できることを明記
  - **実際のファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **実際の行数**: 約 30 行（実装 10 行 + ドキュメント 20 行）
  - **変更ファイル**:
    - `/workspace/go/.golangci.yml`（変更、-1/+8 行）
    - `/workspace/go/docs/architecture-guide.md`（変更、+20 行）
  - **依存**: タスク 2-4 完了後

### フェーズ 3: ドキュメント更新

- [x] **3-1**: CLAUDE.md の更新

  - `go/CLAUDE.md` の「コミット前に実行するコマンド」セクションを更新
  - golangci-lint の使い方を追加
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）
  - **変更ファイル**:
    - `/workspace/go/CLAUDE.md`（変更、約 +50 行）
  - **依存**: タスク 1-1 完了後

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **golangci-lint の自動修正機能**: 初回導入時は問題の検出に留め、自動修正は段階的に導入
- **カスタムリンターの開発**: 標準のリンターで十分な品質を確保できるため、カスタムリンターは不要
- **pre-commit フックへの統合**: 開発者の自由度を保つため、pre-commit フックには統合しない

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [golangci-lint 公式ドキュメント](https://golangci-lint.run/)
- [golangci-lint の設定ファイルリファレンス](https://golangci-lint.run/usage/configuration/)
- [gosec - Go セキュリティチェッカー](https://github.com/securego/gosec)
- [depguard - 依存関係チェッカー](https://github.com/OpenPeeDeeP/depguard)
- [golangci-lint-action - GitHub Actions](https://github.com/golangci/golangci-lint-action)

---
