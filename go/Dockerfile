# syntax=docker/dockerfile:1.7

# ========================================
# Build stage
# [Ja] ビルドステージ
# ========================================
FROM golang:1.26.2-alpine3.22 AS builder

WORKDIR /app

# Build-time system packages.
# - git: required by `go install` / `go mod download` for VCS-backed modules.
# - libstdc++: required by the Node binary copied from the official node:alpine image.
# - curl: used to fetch the dbmate binary with SHA256 verification.
#
# [Ja] ビルド時に必要なシステムパッケージ:
# - git: `go install` / `go mod download` で VCS 由来のモジュールを取得するため
# - libstdc++: 公式 node:alpine からコピーした Node バイナリの実行に必要
# - curl: dbmate を SHA256 検証付きで取得するために使用
RUN apk add --no-cache git libstdc++ curl

# Copy Node and corepack from the official node:alpine image. Using the official
# image (rather than `apk add nodejs`) keeps the production Node version exactly
# in sync with go/mise.toml, so the frontend build is reproducible across dev and
# prod. The corepack symlink at /usr/local/bin/corepack is materialized as a plain
# file by COPY, so we recreate it afterwards. pnpm is then pinned via corepack to
# the version declared in package.json's packageManager field.
#
# [Ja] 公式 node:alpine イメージから Node と corepack をコピーする。`apk add nodejs`
# ではなく公式イメージを使うことで、本番の Node バージョンを go/mise.toml と
# 完全に一致させ、フロントエンドビルドを開発・本番で再現可能にする。
# /usr/local/bin/corepack の symlink は COPY によって実体ファイルになるため、
# 後で symlink を作り直す。pnpm は corepack 経由で package.json の packageManager
# フィールドと同じバージョンに固定する。
COPY --from=node:22.20.0-alpine3.22 /usr/local/bin/node /usr/local/bin/node
COPY --from=node:22.20.0-alpine3.22 /usr/local/lib/node_modules/corepack /usr/local/lib/node_modules/corepack
RUN ln -sf /usr/local/lib/node_modules/corepack/dist/corepack.js /usr/local/bin/corepack && \
    corepack enable && \
    corepack prepare pnpm@10.17.1 --activate

# Install dbmate with SHA256 verification. Never use `releases/latest` (a floating
# pointer) - pin both the version and the digest. DBMATE_VERSION / DBMATE_SHA256
# are intentionally duplicated in Dockerfile.dev to keep each Dockerfile
# self-contained; bump them together when upgrading dbmate.
#
# [Ja] dbmate を SHA256 検証付きでインストールする。浮動参照である
# `releases/latest` は使わず、バージョンとハッシュを必ず固定する。
# DBMATE_VERSION / DBMATE_SHA256 は Dockerfile.dev と意図的に二重定義しており
# (各 Dockerfile を自己完結に保つため)、更新時は両ファイルを揃えて変更する。
ARG DBMATE_VERSION=v2.33.0
ARG DBMATE_SHA256=dfd4027141b6e4357158099bdfe6c02b286af70f8e49057937744cf80b38ea1c
RUN curl -fsSL -o /tmp/dbmate "https://github.com/amacneil/dbmate/releases/download/${DBMATE_VERSION}/dbmate-linux-amd64" && \
    echo "${DBMATE_SHA256}  /tmp/dbmate" | sha256sum -c - && \
    mv /tmp/dbmate /usr/local/bin/dbmate && \
    chmod +x /usr/local/bin/dbmate

# Download Go module dependencies. BuildKit cache mounts persist the module cache
# and build cache across image rebuilds, so unchanged dependencies are not
# re-downloaded or re-compiled.
#
# [Ja] Go モジュールの依存関係をダウンロードする。BuildKit のキャッシュマウントで
# モジュールキャッシュとビルドキャッシュを再ビルド間で保持し、依存に変更が無ければ
# 再ダウンロード・再コンパイルを発生させない。
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Install frontend dependencies. The pnpm store cache mount reuses downloaded
# packages across rebuilds. Ordered before tools.go so that touching tools.go
# does not invalidate this layer.
#
# [Ja] フロントエンドの依存をインストールする。pnpm ストアのキャッシュマウントで
# ダウンロード済みパッケージを再ビルド間で再利用する。tools.go の変更でこの
# レイヤーが無効化されないよう、tools.go の COPY より前に置く。
COPY package.json pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

# Install the dev tools declared in tools.go (templ / goimports). `-mod=mod`
# resolves versions strictly from go.mod so the installed binaries match the
# project's pinned versions.
#
# [Ja] tools.go で宣言した開発ツール (templ / goimports) をインストールする。
# `-mod=mod` を指定して go.mod のバージョンを厳密に解決し、固定したバージョンを
# 確実にインストールする。
COPY tools.go ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install -mod=mod github.com/a-h/templ/cmd/templ && \
    go install -mod=mod golang.org/x/tools/cmd/goimports

# Copy the rest of the source tree. .dockerignore already excludes committed
# *_templ.go and test files from the build context.
#
# [Ja] 残りのソースをコピーする。コミット済みの *_templ.go やテストファイルは
# .dockerignore でビルドコンテキストから除外済み。
COPY . .

# Two-stage guard against duplicate templ-generated imports:
# (1) .dockerignore excludes committed *_templ.go from the build context.
# (2) Even if any sneak in, delete them before `templ generate`, then run
#     goimports to merge the package/runtime imports that templ regenerates.
# This pairing prevents the "duplicate import of github.com/a-h/templ" build
# error that occurs when stale generated files are merged with fresh ones.
# goimports runs with `-w .` to match the local `make templ-generate` behavior.
#
# [Ja] templ 生成 import の二重宣言に対する二段ガード:
# (1) .dockerignore でコミット済みの *_templ.go をビルドコンテキストから除外
# (2) それでも何かの拍子に入り込んだ場合に備え、`templ generate` 前に削除し、
#     生成後に goimports で templ が再生成する import の重複をマージ解消する
# この組み合わせにより、古い生成ファイルが新規生成と混在したときに発生する
# 「duplicate import of github.com/a-h/templ」ビルドエラーを防ぐ。
# goimports はローカルの `make templ-generate` と挙動を揃えて `-w .` で実行する。
RUN find internal -name '*_templ.go' -delete && \
    templ generate && \
    goimports -w .

# Build frontend assets (CSS / JS).
# [Ja] フロントエンドアセット (CSS / JS) をビルドする。
RUN pnpm build

# Build the Go binary.
# - -trimpath: strip absolute build paths from the binary (reproducible builds).
# - -ldflags="-s -w": drop symbol table and DWARF debug info to shrink the image.
# - CGO_ENABLED=0: produce a statically linked binary that runs on alpine without glibc.
#
# [Ja] Go バイナリをビルドする。
# - -trimpath: 絶対ビルドパスをバイナリから除去する (ビルドの再現性確保)
# - -ldflags="-s -w": シンボルテーブルと DWARF デバッグ情報を除いてイメージを縮小
# - CGO_ENABLED=0: glibc 不要の静的バイナリを生成し alpine で動かす
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o bin/annict ./cmd/annict

# ========================================
# Runtime stage
# [Ja] 実行ステージ
# ========================================
FROM alpine:3.22

WORKDIR /app

# Runtime packages, kept to the minimum the server and dbmate actually need:
# - ca-certificates: outbound HTTPS to external APIs.
# - tzdata: time zone database for time.LoadLocation.
# - bash: Dokku runs Procfile commands via /bin/bash (Heroku Buildpack
#   convention), so `release: dbmate --no-dump-schema up` fails to launch
#   without it.
# Note: `curl` and `postgresql-client` are intentionally NOT installed.
# dbmate is fetched in the builder stage (no need for curl at runtime) and the
# Go binary talks to PostgreSQL via lib/pq (no need for the psql client).
#
# [Ja] ランタイムに必要な最小構成のパッケージ:
# - ca-certificates: 外部 API への HTTPS 通信
# - tzdata: time.LoadLocation 用のタイムゾーン DB
# - bash: Dokku は Procfile を /bin/bash 経由で実行する (Heroku Buildpack 由来) ため、
#   bash が無いと `release: dbmate --no-dump-schema up` が起動しない
# curl / postgresql-client はランタイムでは不要のため意図的に入れない。dbmate は
# builder ステージで取得済みで curl は使わず、Go バイナリは lib/pq で PostgreSQL に
# 接続するため psql クライアントも要らない。
RUN apk add --no-cache ca-certificates tzdata bash

# Copy the dbmate binary and the build artifacts from the builder stage.
# db/schema.sql is intentionally NOT copied: it is a dev-time snapshot, and the
# Procfile uses `dbmate --no-dump-schema up` so production never needs to write
# it back.
#
# [Ja] dbmate バイナリとビルド成果物を builder ステージからコピーする。
# db/schema.sql は開発時スナップショットのため意図的にコピーしない (Procfile 側で
# `dbmate --no-dump-schema up` を指定しており、本番では書き戻しが発生しない)。
COPY --from=builder /usr/local/bin/dbmate /usr/local/bin/dbmate
COPY --from=builder /app/bin/annict /app/bin/annict
COPY --from=builder /app/static /app/static
COPY --from=builder /app/db/migrations /app/db/migrations
COPY --from=builder /app/internal/i18n/locales /app/internal/i18n/locales

ENV TZ=UTC

# The server listens on the port given by ANNICT_PORT; 8080 here is just
# documentation for `docker run`-based local smoke tests (Dokku maps the
# port automatically).
#
# [Ja] サーバーは ANNICT_PORT で指定されたポートを待ち受ける。ここでの 8080 は
# `docker run` でローカルスモークテストする際の目安 (Dokku はポートを自動で
# マッピングする)。
EXPOSE 8080

# Dokku itself runs the Procfile `web:` line (`bin/annict serve`), so this CMD is
# not used in production. We declare it anyway (with an absolute path so it works
# regardless of WORKDIR, and with the `serve` subcommand since the bare binary
# prints usage and exits non-zero) so that `docker run <image>` can be used for
# local / CI sanity checks of the built image.
#
# [Ja] Dokku は Procfile の `web:` 行 (`bin/annict serve`) を実行するため、この CMD は
# 本番では使われない。それでも `docker run <image>` でビルド済みイメージのローカル /
# CI スモークテストができるように宣言しておく (WORKDIR に依存しない絶対パスで、かつ
# 引数なしのバイナリは usage を表示して非ゼロ終了するため `serve` サブコマンドを付ける)。
CMD ["/app/bin/annict", "serve"]
