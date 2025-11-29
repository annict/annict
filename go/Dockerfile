# ========================================
# ビルドステージ
# ========================================
FROM golang:1.25.1-alpine AS builder

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

# dbmate のインストール（マイグレーション用）
RUN curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/latest/download/dbmate-linux-amd64 && \
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
