# 開発コンテナ統合 設計書

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

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

現在、Go版とRails版の開発環境は別々のDockerコンテナ（`go-app`, `rails-app`）で動作しています。Go版は `golang:1.25.1-trixie` ベース、Rails版は `ruby:3.3.10-slim-bookworm` ベースで、それぞれ異なるベースイメージを使用しています。Go版とRails版を同時に編集したい場面があり、コンテナ間の切り替えが不便です。

本設計では、開発ツールバージョンマネージャー [mise](https://mise.jdx.dev/) を導入することで、1つのDockerコンテナ内でGo版とRails版の両方を開発可能にします。

**目的**:

- Go版とRails版を同一コンテナで開発可能にし、開発体験を向上させる
- Claude Codeから両方のコードベースを同時に操作可能にする

**背景**:

- Go版とRails版が同一のPostgreSQLデータベースを共有し、段階的に機能移行を進めている
- 両方のコードを同時に参照・編集する場面が増えている
- コンテナ間の切り替えが開発効率を低下させている

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

- 1つのDockerコンテナ内でGo版の開発コマンド（`make test`, `make lint`, `air` 等）がすべて動作する
- 1つのDockerコンテナ内でRails版の開発コマンド（`make test`, `make lint`, `bin/check` 等）がすべて動作する
- Go版のディレクトリ（`/workspace/go/`）ではNode.js 22.20.0 + pnpm 10.17.1が使用される
- Rails版のディレクトリ（`/workspace/rails/`）ではNode.js 22.17.0 + Yarn 1.22.22が使用される
- 既存の`make`コマンドやワークフローが変更なく動作する

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

- Dockerイメージのビルド時間が大幅に増加しないこと（目安: 10分以内）
- コンテナ起動後の開発体験が現状と同等以上であること

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

### 現状の構成

```
docker-compose.yml
├── go-app      (golang:1.25.1-trixie ベース)
│   ├── Go 1.25.1
│   ├── Node.js 22.20.0 (NodeSource)
│   ├── pnpm 10.17.1
│   ├── golangci-lint v2.6.2, dbmate, sqlc v1.27.0, air
│   ├── Stripe CLI
│   ├── Redis Tools 7.4系
│   └── 共通ツール (postgresql-client-17, 1Password CLI, Claude Code, zsh等)
│
├── rails-app   (ruby:3.3.10-slim-bookworm ベース)
│   ├── Ruby 3.3.10
│   ├── Node.js 22.17.0 (NodeSource)
│   ├── Yarn 1.22.22
│   ├── Bundler 2.3.12, ImageMagick
│   ├── Playwright依存ライブラリ
│   └── 共通ツール (postgresql-client-17, 1Password CLI, Claude Code, zsh等)
│
├── caddy       (リバースプロキシ)
├── postgresql  (共有DB)
├── redis       (キャッシュ・セッション)
└── imgproxy    (画像処理プロキシ)
```

### 統合後の構成

```
docker-compose.yml
├── app         (debian:trixie ベース、統合コンテナ)
│   ├── mise (開発ツールバージョンマネージャー)
│   │   ├── Go 1.25.1
│   │   ├── Ruby 3.3.10
│   │   ├── Node.js 22.20.0 + pnpm 10.17.1 → /workspace/go/ で使用
│   │   └── Node.js 22.17.0 + Yarn 1.22.22 → /workspace/rails/ で使用
│   ├── Go ツール: golangci-lint v2.6.2, dbmate, sqlc v1.27.0, air
│   ├── Rails ツール: Bundler 2.3.12, ImageMagick
│   ├── Stripe CLI
│   ├── Redis Tools 7.4系
│   ├── Playwright依存ライブラリ
│   └── 共通ツール (postgresql-client-17, 1Password CLI, Claude Code, zsh等)
│
├── caddy       (変更なし)
├── postgresql  (変更なし)
├── redis       (変更なし)
└── imgproxy    (変更なし)
```

### 開発ツールバージョン管理: mise

[mise](https://mise.jdx.dev/) を使用してGo、Ruby、Node.jsのすべての開発ツールバージョンを統一管理します。

**miseを選択した理由**:

- Go、Ruby、Node.js等を統一的に管理できるポリグロットなバージョンマネージャー
- `mise.toml` ファイルによるディレクトリ単位のバージョン自動切り替え
- Docker環境での利用が公式にサポートされている（shims方式）
- 個別のインストール手順（Go tarball、ruby-install等）が不要になり、Dockerfileがシンプルになる

**miseで管理するツール**:

| ツール  | バージョン | 用途                    |
| ------- | ---------- | ----------------------- |
| Go      | 1.25.1     | Go版の開発              |
| Ruby    | 3.3.10     | Rails版の開発           |
| Node.js | 22.20.0    | Go版のフロントエンド    |
| Node.js | 22.17.0    | Rails版のフロントエンド |
| pnpm    | 10.17.1    | Go版のパッケージ管理    |

**注意**: Yarnはmiseではなくnpm経由でグローバルインストールします。miseはYarnのバージョン管理をサポートしていますが、Rails版ではNode.jsに紐づくnpmからYarnをインストールする方が既存の動作に近いためです。

**動作の仕組み**:

1. ルートの `mise.toml` でGo、Rubyのバージョンを指定（プロジェクト全体で共通）
2. 各サブプロジェクトの `mise.toml` でNode.js、pnpmのバージョンを指定（ディレクトリごとに異なる）
3. Docker環境ではshims方式を使用し、`/mise/shims` をPATHに含めることでツールバージョンを自動切り替え

```toml
# /workspace/mise.toml（プロジェクト全体で共通のツール）
[tools]
go = "1.25.1"
ruby = "3.3.10"

# /workspace/go/mise.toml（Go版固有）
[tools]
node = "22.20.0"
pnpm = "10.17.1"

# /workspace/rails/mise.toml（Rails版固有）
[tools]
node = "22.17.0"
```

### パッケージマネージャー管理

Go版はpnpm、Rails版はYarnと、異なるパッケージマネージャーを使用しています。

**Go版（pnpm）**: mise経由でインストール・管理。`go/mise.toml` でバージョンを指定。

**Rails版（Yarn）**: mise経由でNode.jsをインストールした後、npm経由でYarn 1.22.22をグローバルインストール。Yarnはmise管理外だが、Dockerfileのビルド時に固定バージョンでインストールされるため問題なし。

### ベースイメージの選定

`debian:trixie` をベースイメージとして使用し、Go・Ruby・Node.jsはすべてmise経由でインストールします。

**理由**:

- 現在のGo版Dockerfileが `golang:1.25.1-trixie`（Debian Trixie）ベースである
- `golang:` や `ruby:` の公式イメージをベースにすると、もう一方の言語を追加インストールする際に複雑になる
- ニュートラルなDebianベースにmiseを入れ、すべての言語ランタイムをmiseで管理するのが最もシンプル

### Dockerfile.dev の設計

```dockerfile
FROM debian:trixie

ARG USER_ID=1000
ARG GROUP_ID=1000
ARG USERNAME=developer

# --- システムパッケージ ---
RUN <<EOF
apt update
apt dist-upgrade -yq

PACKAGES=$(cat <<'PKGLIST' | sed 's/#.*//'
  build-essential
  ca-certificates           # HTTPS接続に必要
  curl
  file                      # Shrineによる画像アップロードに必要 (Rails)
  fzf                       # Claude Codeで使用
  git
  gnupg                     # PostgreSQL/RedisのGPGキーとaptリポジトリに必要
  imagemagick               # Shrineによる画像アップロードに必要 (Rails)
  jq                        # Claude Codeで使用
  libasound2                # Playwright: ALSA音声ライブラリ
  libatk-bridge2.0-0        # Playwright: ATKブリッジ
  libatk1.0-0               # Playwright: ATK
  libcairo2                 # Playwright: 2Dグラフィックス
  libcups2                  # Playwright: 印刷システム
  libdrm2                   # Playwright: Direct Rendering Manager
  libffi-dev                # Ruby拡張のビルドに必要
  libgbm1                   # Playwright: Generic Buffer Management
  libnspr4                  # Playwright: Netscape Portable Runtime
  libnss3                   # Playwright: Network Security Services
  libpango-1.0-0            # Playwright: テキストレンダリング
  libreadline-dev           # Rubyのreadline拡張に必要
  libssl-dev                # SSL関連のビルドに必要
  libxcomposite1            # Playwright: X11 Composite拡張
  libxdamage1               # Playwright: X11 Damage拡張
  libxfixes3                # Playwright: X11 Fixes拡張
  libxkbcommon0             # Playwright: XKBキーボード処理
  libxrandr2                # Playwright: X11 RandR拡張
  libyaml-dev               # psych gemのインストールに必要 (Rails)
  lsb-release               # aptリポジトリ設定に必要
  nano                      # Claude Codeで使用
  ripgrep                   # Claude Codeで使用
  sudo
  tree                      # Claude Codeで使用
  unzip
  vim                       # Claude Codeで使用
  zlib1g-dev                # Rubyのzlib拡張に必要
  zsh
PKGLIST
)

apt install -y --no-install-recommends $PACKAGES
rm -rf /var/lib/apt/lists/*
EOF

# --- PostgreSQL クライアント ---
RUN curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc \
      | gpg --dearmor -o /etc/apt/trusted.gpg.d/postgresql.gpg && \
    echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" \
      > /etc/apt/sources.list.d/pgdg.list && \
    apt update && apt install -y libpq-dev postgresql-client-17 && \
    rm -rf /var/lib/apt/lists/*

# --- Redis Tools ---
RUN curl -fsSL https://packages.redis.io/gpg \
      | gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" \
      | tee /etc/apt/sources.list.d/redis.list && \
    apt-get update && \
    apt-get install -y redis-tools=6:7.4.* && \
    rm -rf /var/lib/apt/lists/*

# --- Stripe CLI ---
RUN curl -s https://packages.stripe.dev/api/security/keypair/stripe-cli-gpg/public \
      | gpg --dearmor -o /usr/share/keyrings/stripe.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/stripe.gpg] https://packages.stripe.dev/stripe-cli-debian-local stable main" \
      | tee /etc/apt/sources.list.d/stripe.list && \
    apt update && apt install -y stripe && \
    rm -rf /var/lib/apt/lists/*

# --- mise (開発ツールバージョンマネージャー) ---
ENV MISE_DATA_DIR="/mise"
ENV MISE_CONFIG_DIR="/mise"
ENV MISE_CACHE_DIR="/mise/cache"
ENV MISE_INSTALL_PATH="/usr/local/bin/mise"
ENV PATH="/mise/shims:${PATH}"
RUN curl https://mise.run | sh

# Go、Ruby、Node.js、pnpmをmise経由でインストール
RUN mise install go@1.25.1 && \
    mise install ruby@3.3.10 && \
    mise install node@22.20.0 && \
    mise install node@22.17.0 && \
    mise install pnpm@10.17.1

# グローバルデフォルトを設定
RUN mise use --global go@1.25.1 && \
    mise use --global ruby@3.3.10 && \
    mise use --global node@22.20.0 && \
    mise use --global pnpm@10.17.1

# --- dbmate ---
RUN curl -fsSL -o /usr/local/bin/dbmate \
      https://github.com/amacneil/dbmate/releases/latest/download/dbmate-linux-amd64 && \
    chmod +x /usr/local/bin/dbmate

# --- golangci-lint ---
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
      | sh -s -- -b /usr/local/bin v2.6.2

# --- ユーザー作成 ---
RUN groupadd -g ${GROUP_ID} ${USERNAME} && \
    useradd -u ${USER_ID} -g ${GROUP_ID} -m -s /bin/zsh ${USERNAME} && \
    echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# --- Go キャッシュディレクトリ ---
RUN mkdir -p /go/.cache/go-build /go/pkg && \
    chown -R ${USER_ID}:${GROUP_ID} /go

# --- mise のデータディレクトリの権限設定 ---
RUN chown -R ${USER_ID}:${GROUP_ID} /mise

WORKDIR /workspace
USER ${USERNAME}

# --- mise シェル設定 ---
# Docker環境ではshims方式を使用（PATH="/mise/shims:..."で設定済み）
RUN mise reshim

# --- 1Password CLI ---
COPY --from=1password/op:2 /usr/local/bin/op /usr/local/bin/op

# --- npmグローバルインストール先を一般ユーザーのホームディレクトリに設定 ---
ENV NPM_CONFIG_PREFIX=/home/${USERNAME}/.npm-global
ENV PATH=$NPM_CONFIG_PREFIX/bin:$PATH

# --- Claude Code ---
RUN curl -fsSL https://claude.ai/install.sh | bash

# --- Yarn (Rails版で使用) ---
RUN npm install --global yarn@1.22.22

# --- Go ツール ---
ENV GOCACHE=/go/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod
ENV GOPATH=/go
ENV PATH="${GOPATH}/bin:${PATH}"
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 && \
    go install github.com/air-verse/air@latest

# --- Ruby ツール ---
RUN gem install bundler -v 2.3.12

# --- シェル設定 ---
SHELL ["/bin/zsh", "-c"]
ENV SHELL=/bin/zsh

RUN touch /home/${USERNAME}/.zshrc

# --- Git設定 ---
RUN git config --global user.email "me@shimba.co" && \
    git config --global user.name "Koji Shimba"

EXPOSE 8080 3000

CMD ["/bin/zsh"]
```

### docker-compose.yml の変更

```yaml
services:
  app:
    build:
      context: .
      dockerfile: ./Dockerfile.dev
    depends_on:
      - caddy
      - imgproxy
      - postgresql
      - redis
    volumes:
      - .:/workspace
      - op-config:/home/developer/.config/op
      - go-mod-cache:/go/pkg/mod
      - go-build-cache:/go/.cache/go-build
      - app-gems-data:/mise/installs/ruby/3.3.10/lib/ruby/gems/3.3.0
    ports:
      - "4004:8080" # Go
      - "4000:3000" # Rails
    stdin_open: true
    tty: true
    working_dir: /workspace
    environment:
      - BINDING=0.0.0.0
      - OP_SERVICE_ACCOUNT_TOKEN=${OP_SERVICE_ACCOUNT_TOKEN:-}

  caddy:
    image: caddy:2-alpine
    ports:
      - "4008:8080"
    volumes:
      - ./caddy/Caddyfile:/etc/caddy/Caddyfile:ro

  postgresql:
    image: postgres:17.5
    ports:
      - "4001:5432"
    volumes:
      - postgresql17_data:/var/lib/postgresql/data:delegated
    environment:
      POSTGRES_HOST_AUTH_METHOD: trust

  redis:
    image: redis:8.2.3-alpine
    ports:
      - "4002:6379"

  imgproxy:
    image: darthsim/imgproxy:v3.27.2
    ports:
      - "4003:8080"
    environment:
      - IMGPROXY_S3_REGION=${IMGPROXY_S3_REGION:-}
      - AWS_ACCESS_KEY_ID=${IMGPROXY_AWS_ACCESS_KEY_ID:-}
      - AWS_SECRET_ACCESS_KEY=${IMGPROXY_AWS_SECRET_ACCESS_KEY:-}
      - IMGPROXY_S3_ENDPOINT=${IMGPROXY_S3_ENDPOINT:-}
    env_file:
      - ./imgproxy/.env

volumes:
  app-gems-data:
  go-build-cache:
  go-mod-cache:
  op-config:
  postgresql17_data:
```

### mise.toml ファイルの追加

ルートとサブプロジェクトに `mise.toml` ファイルを配置し、ツールバージョンを指定します。

```toml
# /workspace/mise.toml（プロジェクト全体で共通のツール）
[tools]
go = "1.25.1"
ruby = "3.3.10"

# /workspace/go/mise.toml（Go版固有）
[tools]
node = "22.20.0"
pnpm = "10.17.1"

# /workspace/rails/mise.toml（Rails版固有）
[tools]
node = "22.17.0"
```

miseは階層的な設定を持ち、カレントディレクトリから親ディレクトリへと`mise.toml`を探索します。設定はマージされるため:

- `/workspace/go/` 内では Go 1.25.1 + Ruby 3.3.10（親から継承）+ Node.js 22.20.0 + pnpm 10.17.1（自身の設定）
- `/workspace/rails/` 内では Go 1.25.1 + Ruby 3.3.10（親から継承）+ Node.js 22.17.0（自身の設定）

バージョン更新時は `mise.toml` を変更するだけで済み、Dockerfileの再ビルドは不要です（新しいバージョンは `mise install` で追加）。

### Makefileへの影響

各サブプロジェクトの `Makefile` は基本的に変更不要です。miseのshims方式により、`/mise/shims` がPATHに含まれていれば、カレントディレクトリの `mise.toml` に基づいて自動的に正しいバージョンの `go`、`ruby`、`node`、`pnpm`、`yarn` 等が使用されます。

`make` コマンドはサブシェルで実行されますが、shims方式ではシェル初期化が不要なため、Makefile内でコマンドを呼び出すだけで正しいバージョンが使用されます。

具体的な対応はタスク実装時に検証します。

### 移行の注意点

- **既存のDockerボリューム**: `go-mod-cache`, `go-build-cache` はそのまま利用可能
- **Gemのインストール先**: mise管理のRubyはシステムとは異なるパスにインストールされるため、Gemのインストール先パスが変わる。`app-gems-data` ボリュームのマウントパスを調整する（`/mise/installs/ruby/3.3.10/lib/ruby/gems/3.3.0`）
- **GoのPATH**: mise管理のGoは `/mise/installs/go/1.25.1` 配下にインストールされる。`GOROOT` の設定や `go install` でインストールされるバイナリのパスに注意する
- **ポート**: Go（8080）とRails（3000）の両方を公開
- **旧コンテナの削除**: 移行後、`go-app` と `rails-app` サービスは削除
- **ベースイメージの違い**: Rails版は現在 Debian Bookworm ベースだが、統合後は Debian Trixie ベースになる。Playwright依存ライブラリのパッケージ名が異なる可能性があるため、動作確認が必要
- **npm グローバルパッケージ**: Yarn は引き続きnpm経由でグローバルインストール。Claude Codeはネイティブインストール方式（`curl -fsSL https://claude.ai/install.sh | bash`）を使用

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

### フェーズ 1: 統合Dockerfileの作成と検証

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
Go版/Rails版の両方を修正する場合は別タスクに分けてください
-->

- [x] **1-1**: 統合 Dockerfile.dev を作成
  - ルートに `Dockerfile.dev` を新規作成（上記設計に基づく）
  - `mise.toml`（ルート）、`go/mise.toml`、`rails/mise.toml` を作成
  - `docker-compose.yml` を更新（`go-app` + `rails-app` → `app` に統合）
  - コンテナをビルドして起動できることを確認
  - **想定ファイル数**: 約 5 ファイル（実装 5 + テスト 0）
  - **想定行数**: 約 200 行（実装 200 行 + テスト 0 行）

### フェーズ 2: Go版の動作確認と調整

- [x] **2-1**: [Go] 統合コンテナでのGo版動作確認と調整
  - `cd /workspace/go && make test` が通ることを確認
  - `cd /workspace/go && make lint` が通ることを確認
  - `cd /workspace/go && make fmt` が動作することを確認
  - Node.js/pnpm関連のコマンド（`pnpm install`, `pnpm build`）が正しいバージョンで動作することを確認
  - 必要に応じてMakefileや設定ファイルを調整
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

### フェーズ 3: Rails版の動作確認と調整

- [x] **3-1**: [Rails] 統合コンテナでのRails版動作確認と調整
  - `cd /workspace/rails && make test` が通ることを確認
  - `cd /workspace/rails && make lint` が通ることを確認
  - Node.js/Yarn関連のコマンド（`yarn install`, `yarn build`）が正しいバージョンで動作することを確認
  - Gemのインストール先パスが正しいことを確認
  - Playwright依存ライブラリがDebian Trixie上で正しく動作することを確認
  - 必要に応じてMakefileや設定ファイルを調整
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

### フェーズ 4: クリーンアップ

- [x] **4-1**: 旧Dockerfileの削除とドキュメント更新
  - `go/Dockerfile.dev` を削除
  - `rails/Dockerfile.dev` を削除
  - `CLAUDE.md` のDocker関連セクションを更新
  - `go/CLAUDE.md` のコンテナ関連セクションを更新
  - `rails/CLAUDE.md` のコンテナ関連セクションを更新
  - **想定ファイル数**: 約 5 ファイル（実装 5 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **本番用Dockerfileの統合**: 本設計は開発環境（`Dockerfile.dev`）のみを対象とする。本番用Dockerfileは引き続き個別に管理する
- **CI/CDの変更**: GitHub ActionsのCI設定は現状のまま維持する
- **devcontainerの導入**: VS Code Dev Container設定の追加は今回のスコープ外

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [mise - 開発ツールバージョンマネージャー](https://mise.jdx.dev/)
- [mise - Docker環境での使い方](https://mise.jdx.dev/containers/)
