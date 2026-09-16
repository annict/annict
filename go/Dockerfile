# syntax=docker/dockerfile:1.7

# ========================================
# ビルドステージ
# ========================================
FROM golang:1.26.2-alpine3.22 AS builder

WORKDIR /app

# ビルド時に必要なシステムパッケージ:
# - git: `go install` / `go mod download` でVCS由来のモジュールを取得するため
# - libstdc++: 公式node:alpineからコピーしたNodeバイナリの実行に必要
# - curl: dbmateをSHA256検証付きで取得するために使用
RUN apk add --no-cache git libstdc++ curl

# 公式node:alpineイメージからNodeとcorepackをコピーする。`apk add nodejs`
# ではなく公式イメージを使うことで、本番のNodeバージョンをgo/mise.tomlと
# 完全に一致させ、フロントエンドビルドを開発・本番で再現可能にする。
# /usr/local/bin/corepackのsymlinkはCOPYによって実体ファイルになるため、
# 後でsymlinkを作り直す。pnpmはcorepack経由でpackage.jsonのpackageManager
# フィールドと同じバージョンに固定する。
COPY --from=node:22.20.0-alpine3.22 /usr/local/bin/node /usr/local/bin/node
COPY --from=node:22.20.0-alpine3.22 /usr/local/lib/node_modules/corepack /usr/local/lib/node_modules/corepack
RUN ln -sf /usr/local/lib/node_modules/corepack/dist/corepack.js /usr/local/bin/corepack && \
    corepack enable && \
    corepack prepare pnpm@10.17.1 --activate

# dbmateをSHA256検証付きでインストールする。浮動参照である
# `releases/latest` は使わず、バージョンとハッシュを必ず固定する。
# DBMATE_VERSION / DBMATE_SHA256はDockerfile.devと意図的に二重定義しており
# (各Dockerfileを自己完結に保つため)、更新時は両ファイルを揃えて変更する。
ARG DBMATE_VERSION=v2.33.0
ARG DBMATE_SHA256=dfd4027141b6e4357158099bdfe6c02b286af70f8e49057937744cf80b38ea1c
RUN curl -fsSL -o /tmp/dbmate "https://github.com/amacneil/dbmate/releases/download/${DBMATE_VERSION}/dbmate-linux-amd64" && \
    echo "${DBMATE_SHA256}  /tmp/dbmate" | sha256sum -c - && \
    mv /tmp/dbmate /usr/local/bin/dbmate && \
    chmod +x /usr/local/bin/dbmate

# Goモジュールの依存関係をダウンロードする。BuildKitのキャッシュマウントで
# モジュールキャッシュとビルドキャッシュを再ビルド間で保持し、依存に変更が無ければ
# 再ダウンロード・再コンパイルを発生させない。
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# フロントエンドの依存をインストールする。pnpmストアのキャッシュマウントで
# ダウンロード済みパッケージを再ビルド間で再利用する。tools.goの変更でこの
# レイヤーが無効化されないよう、tools.goのCOPYより前に置く。
COPY package.json pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

# tools.goで宣言した開発ツール (templ / goimports) をインストールする。
# `-mod=mod` を指定してgo.modのバージョンを厳密に解決し、固定したバージョンを
# 確実にインストールする。
COPY tools.go ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install -mod=mod github.com/a-h/templ/cmd/templ && \
    go install -mod=mod golang.org/x/tools/cmd/goimports

# 残りのソースをコピーする。コミット済みの *_templ.goやテストファイルは
# .dockerignoreでビルドコンテキストから除外済み。
COPY . .

# templ生成importの二重宣言に対する二段ガード:
# (1) .dockerignoreでコミット済みの *_templ.goをビルドコンテキストから除外
# (2) それでも何かの拍子に入り込んだ場合に備え、`templ generate` 前に削除し、
#     生成後にgoimportsでtemplが再生成するimportの重複をマージ解消する
# この組み合わせにより、古い生成ファイルが新規生成と混在したときに発生する
# 「duplicate import of github.com/a-h/templ」ビルドエラーを防ぐ。
# goimportsはローカルの `make templ-generate` と挙動を揃えて `-w .` で実行する。
RUN find internal -name '*_templ.go' -delete && \
    templ generate && \
    goimports -w .

# フロントエンドアセット (CSS / JS) をビルドする。
RUN pnpm build

# Goバイナリをビルドする。
# - -trimpath: 絶対ビルドパスをバイナリから除去する (ビルドの再現性確保)
# - -ldflags="-s -w": シンボルテーブルとDWARFデバッグ情報を除いてイメージを縮小
# - CGO_ENABLED=0: glibc不要の静的バイナリを生成しalpineで動かす
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o bin/annict ./cmd/annict

# ========================================
# 実行ステージ
# ========================================
FROM alpine:3.22

WORKDIR /app

# ランタイムに必要な最小構成のパッケージ:
# - ca-certificates: 外部APIへのHTTPS通信
# - tzdata: time.LoadLocation用のタイムゾーンDB
# - bash: DokkuはProcfileを /bin/bash経由で実行する (Heroku Buildpack由来) ため、
#   bashが無いと `release: dbmate --no-dump-schema up` が起動しない
# curl / postgresql-clientはランタイムでは不要のため意図的に入れない。dbmateは
# builderステージで取得済みでcurlは使わず、Goバイナリはlib/pqでPostgreSQLに
# 接続するためpsqlクライアントも要らない。
RUN apk add --no-cache ca-certificates tzdata bash

# dbmateバイナリとビルド成果物をbuilderステージからコピーする。
# db/schema.sqlは開発時スナップショットのため意図的にコピーしない (Procfile側で
# `dbmate --no-dump-schema up` を指定しており、本番では書き戻しが発生しない)。
COPY --from=builder /usr/local/bin/dbmate /usr/local/bin/dbmate
COPY --from=builder /app/bin/annict /app/bin/annict
COPY --from=builder /app/static /app/static
COPY --from=builder /app/db/migrations /app/db/migrations
COPY --from=builder /app/internal/i18n/locales /app/internal/i18n/locales

ENV TZ=UTC

# サーバーはANNICT_PORTで指定されたポートを待ち受ける。ここでの8080は
# `docker run` でローカルスモークテストする際の目安 (Dokkuはポートを自動で
# マッピングする)。
EXPOSE 8080

# DokkuはProcfileの `web:` 行 (`bin/annict serve`) を実行するため、このCMDは
# 本番では使われない。それでも `docker run <image>` でビルド済みイメージのローカル /
# CIスモークテストができるように宣言しておく (WORKDIRに依存しない絶対パスで、かつ
# 引数なしのバイナリはusageを表示して非ゼロ終了するため `serve` サブコマンドを付ける)。
CMD ["/app/bin/annict", "serve"]
