#!/usr/bin/env bash
#
# load_snapshot_images.shは本番の作品・プロフィール画像を開発用R2バケットへ
# コピーし、ローカルアプリが本番と同じように画像を表示できるようにする。本番の
# コンテンツバケット (annict-user-contents) のshrine/ prefixを、開発用バケット
# (annict-development) へrclone syncで全件ミラーする。
#
# 取得元は本番の読み取り専用R2資格情報 (ANNICT_PROD_S3_*) を使い、宛先はアプリが
# 画像配信に既に使っている資格情報 (ANNICT_S3_*) を流用する。いずれの値もop run経由
# で1Passwordから注入されるため、必ずMakefile経由で実行する。
#
# 使い方:
#   make load-snapshot-images

set -euo pipefail

# 取得元の本番R2読み取り専用資格情報と、宛先の開発用R2資格情報の両方を
# 要求する (いずれもop run経由で1Passwordから注入)。
: "${ANNICT_PROD_S3_ENDPOINT:?ANNICT_PROD_S3_ENDPOINTが未設定です}"
: "${ANNICT_PROD_S3_ACCESS_KEY_ID:?ANNICT_PROD_S3_ACCESS_KEY_IDが未設定です}"
: "${ANNICT_PROD_S3_SECRET_ACCESS_KEY:?ANNICT_PROD_S3_SECRET_ACCESS_KEYが未設定です}"
: "${ANNICT_PROD_S3_CONTENTS_BUCKET:?ANNICT_PROD_S3_CONTENTS_BUCKETが未設定です}"
: "${ANNICT_S3_ENDPOINT:?ANNICT_S3_ENDPOINTが未設定です}"
: "${ANNICT_S3_ACCESS_KEY_ID:?ANNICT_S3_ACCESS_KEY_IDが未設定です}"
: "${ANNICT_S3_SECRET_ACCESS_KEY:?ANNICT_S3_SECRET_ACCESS_KEYが未設定です}"
: "${ANNICT_S3_BUCKET_NAME:?ANNICT_S3_BUCKET_NAMEが未設定です}"

# rcloneに空の設定ファイル (/dev/null) を指定し、起動時の
# "Config file ... not found" のNOTICEを出さないようにする。リモートは下記の
# RCLONE_CONFIG_* だけで定義しており、実体の設定ファイルは不要なため。
export RCLONE_CONFIG=/dev/null

# コミット済みのrclone.confではなくenvから2つのインラインrclone S3
# リモートを構成する: 取得元の本番用 "prods3" と宛先の開発用 "devs3"。バックエンドが
# R2の間はrcloneのproviderを "Cloudflare" のままにする。
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

# shrine/ prefixを全件ミラーする。copyではなくsyncを使い、宛先を取得元に
# 一致させて不要オブジェクトを削除する。開発用バケットは開発専用なので安全。フラグは
# 大量転送向けの調整で、--transfers/--checkersはrcloneの保守的な既定値 (4/8) より
# 並列度を上げ、--fast-listは再帰リストでR2のリスト操作を減らし、--progressは
# 長時間のsyncに進捗を表示する。
echo "==> 画像を同期: ${ANNICT_PROD_S3_CONTENTS_BUCKET}/shrine/ -> ${ANNICT_S3_BUCKET_NAME}/shrine/"
rclone sync \
  "prods3:${ANNICT_PROD_S3_CONTENTS_BUCKET}/shrine/" \
  "devs3:${ANNICT_S3_BUCKET_NAME}/shrine/" \
  --transfers 16 \
  --checkers 32 \
  --fast-list \
  --progress

echo "==> 完了 (本番の画像を ${ANNICT_S3_BUCKET_NAME}/shrine/ へ同期しました)"
