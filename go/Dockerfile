# ========================================
# ビルドステージ
# ========================================
FROM golang:1.26.2-alpine AS builder

# 作業ディレクトリを設定
WORKDIR /app

# システム依存関係のインストール
RUN apk add --no-cache \
    git \
    nodejs \
    npm \
    curl

# pnpm のインストール
RUN npm install -g pnpm@10.17.1

# Go 依存関係ファイルをコピー
COPY go.mod go.sum ./
RUN go mod download

# pnpm 依存関係ファイルをコピー
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# tools.go で定義された開発ツールをインストール
COPY tools.go ./
RUN go install github.com/a-h/templ/cmd/templ && \
    go install golang.org/x/tools/cmd/goimports

# ソースコードをコピー
COPY . .

# templ ファイルから Go コードを生成
RUN templ generate

# import 文を整理
RUN goimports -w .

# フロントエンドアセットをビルド
RUN pnpm build

# Go バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server cmd/server/main.go

# ========================================
# 実行ステージ
# ========================================
FROM alpine:3.21

# 作業ディレクトリを設定
WORKDIR /app

# システム依存関係のインストール
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    postgresql-client \
    curl \
    bash

# Install dbmate (used for migrations) with SHA256 verification. Never use
# `releases/latest` (a floating pointer) - pin both the version and the digest.
# DBMATE_VERSION / DBMATE_SHA256 are intentionally duplicated in Dockerfile.dev
# to keep each Dockerfile self-contained; bump them together when upgrading dbmate.
#
# [Ja] dbmate を SHA256 検証付きでインストールする (マイグレーション用)。浮動参照である
# `releases/latest` は使わず、バージョンとハッシュを必ず固定する。
# DBMATE_VERSION / DBMATE_SHA256 は Dockerfile.dev と意図的に二重定義しており
# (各 Dockerfile を自己完結に保つため)、更新時は両ファイルを揃えて変更する。
ARG DBMATE_VERSION=v2.33.0
ARG DBMATE_SHA256=dfd4027141b6e4357158099bdfe6c02b286af70f8e49057937744cf80b38ea1c
RUN curl -fsSL -o /tmp/dbmate "https://github.com/amacneil/dbmate/releases/download/${DBMATE_VERSION}/dbmate-linux-amd64" && \
    echo "${DBMATE_SHA256}  /tmp/dbmate" | sha256sum -c - && \
    mv /tmp/dbmate /usr/local/bin/dbmate && \
    chmod +x /usr/local/bin/dbmate

# ビルドステージから必要なファイルをコピー
COPY --from=builder /app/bin/server /app/bin/server
COPY --from=builder /app/static /app/static
COPY --from=builder /app/db /app/db
COPY --from=builder /app/internal/i18n/locales /app/internal/i18n/locales

# タイムゾーンを UTC に設定
ENV TZ=UTC

# ポートを公開（Dokku が自動的にマッピング）
EXPOSE 8080
