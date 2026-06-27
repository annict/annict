#!/usr/bin/env bash
#
# load_snapshot_images.sh copies the production work/profile images into the
# development R2 bucket so the local app renders images exactly like production.
# It mirrors the shrine/ prefix from the production contents bucket
# (annict-user-contents) into the development bucket (annict-development) with a
# full rclone sync.
#
# The source uses the production read-only R2 credentials (ANNICT_PROD_S3_*);
# the destination reuses the credentials the app already uses for image delivery
# (ANNICT_S3_*). All values are injected from 1Password via op run, so always
# run this through the Makefile.
#
# Usage:
#   make load-snapshot-images
#
# [Ja] load_snapshot_images.sh は本番の作品・プロフィール画像を開発用 R2 バケットへ
# コピーし、ローカルアプリが本番と同じように画像を表示できるようにする。本番の
# コンテンツバケット (annict-user-contents) の shrine/ prefix を、開発用バケット
# (annict-development) へ rclone sync で全件ミラーする。
#
# 取得元は本番の読み取り専用 R2 資格情報 (ANNICT_PROD_S3_*) を使い、宛先はアプリが
# 画像配信に既に使っている資格情報 (ANNICT_S3_*) を流用する。いずれの値も op run 経由
# で 1Password から注入されるため、必ず Makefile 経由で実行する。
#
# 使い方:
#   make load-snapshot-images

set -euo pipefail

# Require both the production R2 read-only credentials (source) and the
# development R2 credentials (destination), all injected from 1Password via
# op run.
#
# [Ja] 取得元の本番 R2 読み取り専用資格情報と、宛先の開発用 R2 資格情報の両方を
# 要求する (いずれも op run 経由で 1Password から注入)。
: "${ANNICT_PROD_S3_ENDPOINT:?ANNICT_PROD_S3_ENDPOINT is required}"
: "${ANNICT_PROD_S3_ACCESS_KEY_ID:?ANNICT_PROD_S3_ACCESS_KEY_ID is required}"
: "${ANNICT_PROD_S3_SECRET_ACCESS_KEY:?ANNICT_PROD_S3_SECRET_ACCESS_KEY is required}"
: "${ANNICT_PROD_S3_CONTENTS_BUCKET:?ANNICT_PROD_S3_CONTENTS_BUCKET is required}"
: "${ANNICT_S3_ENDPOINT:?ANNICT_S3_ENDPOINT is required}"
: "${ANNICT_S3_ACCESS_KEY_ID:?ANNICT_S3_ACCESS_KEY_ID is required}"
: "${ANNICT_S3_SECRET_ACCESS_KEY:?ANNICT_S3_SECRET_ACCESS_KEY is required}"
: "${ANNICT_S3_BUCKET_NAME:?ANNICT_S3_BUCKET_NAME is required}"

# Point rclone at an empty config file (/dev/null) so it does not emit the
# "Config file ... not found" notice on startup. The remotes are defined entirely
# by the RCLONE_CONFIG_* variables below, so no real config file is needed.
#
# [Ja] rclone に空の設定ファイル (/dev/null) を指定し、起動時の
# "Config file ... not found" の NOTICE を出さないようにする。リモートは下記の
# RCLONE_CONFIG_* だけで定義しており、実体の設定ファイルは不要なため。
export RCLONE_CONFIG=/dev/null

# Configure two inline rclone S3 remotes from the env vars instead of a committed
# rclone.conf: "prods3" for the production source and "devs3" for the development
# destination. The rclone provider stays "Cloudflare" while the backend is R2.
#
# [Ja] コミット済みの rclone.conf ではなく env から 2 つのインライン rclone S3
# リモートを構成する: 取得元の本番用 "prods3" と宛先の開発用 "devs3"。バックエンドが
# R2 の間は rclone の provider を "Cloudflare" のままにする。
export RCLONE_CONFIG_PRODS3_TYPE=s3
export RCLONE_CONFIG_PRODS3_PROVIDER=Cloudflare
export RCLONE_CONFIG_PRODS3_ENDPOINT="$ANNICT_PROD_S3_ENDPOINT"
export RCLONE_CONFIG_PRODS3_ACCESS_KEY_ID="$ANNICT_PROD_S3_ACCESS_KEY_ID"
export RCLONE_CONFIG_PRODS3_SECRET_ACCESS_KEY="$ANNICT_PROD_S3_SECRET_ACCESS_KEY"

export RCLONE_CONFIG_DEVS3_TYPE=s3
export RCLONE_CONFIG_DEVS3_PROVIDER=Cloudflare
export RCLONE_CONFIG_DEVS3_ENDPOINT="$ANNICT_S3_ENDPOINT"
export RCLONE_CONFIG_DEVS3_ACCESS_KEY_ID="$ANNICT_S3_ACCESS_KEY_ID"
export RCLONE_CONFIG_DEVS3_SECRET_ACCESS_KEY="$ANNICT_S3_SECRET_ACCESS_KEY"

# Mirror the shrine/ prefix in full. sync (not copy) makes the destination match
# the source, deleting stale objects; the development bucket is dev-only so that
# is safe. The flags tune a bulk transfer: --transfers/--checkers raise the
# concurrency over rclone's conservative defaults (4/8), --fast-list cuts R2 list
# operations by listing recursively, and --progress shows live progress for a
# long-running sync.
#
# [Ja] shrine/ prefix を全件ミラーする。copy ではなく sync を使い、宛先を取得元に
# 一致させて不要オブジェクトを削除する。開発用バケットは開発専用なので安全。フラグは
# 大量転送向けの調整で、--transfers/--checkers は rclone の保守的な既定値 (4/8) より
# 並列度を上げ、--fast-list は再帰リストで R2 のリスト操作を減らし、--progress は
# 長時間の sync に進捗を表示する。
echo "==> Syncing images: ${ANNICT_PROD_S3_CONTENTS_BUCKET}/shrine/ -> ${ANNICT_S3_BUCKET_NAME}/shrine/"
rclone sync \
  "prods3:${ANNICT_PROD_S3_CONTENTS_BUCKET}/shrine/" \
  "devs3:${ANNICT_S3_BUCKET_NAME}/shrine/" \
  --transfers 16 \
  --checkers 32 \
  --fast-list \
  --progress

echo "==> Done (synced production images into ${ANNICT_S3_BUCKET_NAME}/shrine/)"
