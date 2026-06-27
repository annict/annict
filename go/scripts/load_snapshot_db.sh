#!/usr/bin/env bash
#
# load_snapshot_db.sh builds a local development database from a production
# snapshot. It downloads the production pg_dump backup from Cloudflare R2,
# restores it into an isolated throwaway database, masks PII there, dumps the
# masked result to tmp/, and finally (re)loads that masked dump into the local
# development database and applies pending dbmate migrations.
#
# Raw PII only ever lives in the isolated annict_anonymize database; the
# development database (annict_development) only ever receives the masked dump.
#
# Usage (always via the Makefile so 1Password injects the secrets):
#   make load-snapshot-db       # full pipeline: download -> mask -> load
#   make restore-snapshot-db    # reload only, from the masked dump in tmp/
#
# [Ja] load_snapshot_db.sh は本番スナップショットからローカル開発 DB を構築する。
# 本番 pg_dump バックアップを Cloudflare R2 から取得し、隔離した使い捨て DB に
# 復元して PII をマスクし、マスク結果を tmp/ にダンプしたうえで、それをローカル
# 開発 DB に (再) 流し込み、未適用の dbmate マイグレーションを適用する。
#
# 生 PII は隔離 DB (annict_anonymize) にしか存在せず、開発 DB (annict_development)
# にはマスク済みダンプしか入らない。
#
# 使い方 (秘密注入のため必ず Makefile 経由で実行する):
#   make load-snapshot-db       # フルパイプライン: 取得 -> マスク -> 流し込み
#   make restore-snapshot-db    # 再流し込みのみ (tmp/ のマスク済みダンプから)

set -euo pipefail

# Local PostgreSQL connection. Matches the db-* Makefile targets (trust auth, no
# password). The database names are fixed to the local development conventions.
#
# [Ja] ローカル PostgreSQL の接続設定。db-* の Makefile ターゲットと揃える
# (trust 認証・パスワード無し)。DB 名はローカル開発の慣習に固定する。
export PGHOST="${PGHOST:-postgresql}"
export PGUSER="${PGUSER:-postgres}"
ANONYMIZE_DB="annict_anonymize"
DEV_DB="annict_development"

# Working directory for downloads and the masked dump (gitignored tmp/).
#
# [Ja] ダウンロードとマスク済みダンプの作業ディレクトリ (gitignore 済みの tmp/)。
SNAPSHOT_DIR="tmp/snapshot"
EXPORT_FILE="${SNAPSHOT_DIR}/backup/export"
MASKED_DUMP="${SNAPSHOT_DIR}/masked.dump"

# On exit, drop the isolated database and remove the downloaded raw dump so raw
# PII never lingers, even when the pipeline fails partway through. Restores reuse
# the masked dump in tmp/, so the raw dump is safe to delete here.
#
# [Ja] 終了時に隔離 DB を落とし、取得した生ダンプも削除して、途中で失敗しても
# 生 PII が残らないようにする。再復元は tmp/ のマスク済みダンプを使うため、ここで
# 生ダンプを削除して問題ない。
cleanup() {
  dropdb --if-exists "$ANONYMIZE_DB" >/dev/null 2>&1 || true
  rm -rf "${SNAPSHOT_DIR:?}/backup" "${SNAPSHOT_DIR}/backup.tgz"
}
trap cleanup EXIT

# (5) Reload the local development database from the masked dump and apply
# migrations. Shared by the full pipeline and the restore-only entry point.
#
# [Ja] (5) マスク済みダンプからローカル開発 DB を再構築し、マイグレーションを
# 適用する。フルパイプラインと再復元専用エントリの両方から使う。
load_into_dev_db() {
  if [ ! -f "$MASKED_DUMP" ]; then
    echo "Masked dump not found: ${MASKED_DUMP}. Run 'make load-snapshot-db' first." >&2
    exit 1
  fi

  echo "==> Rebuilding ${DEV_DB} from the masked dump"
  dropdb --if-exists "$DEV_DB"
  createdb "$DEV_DB"
  # Restore without -C and with --no-owner/--no-acl so the production roles and
  # the CREATE/DROP DATABASE TOC entries in the dump are ignored. pg_restore can
  # report ignorable errors (already-present extensions etc.), so do not abort
  # the whole script on its exit status.
  #
  # [Ja] -C を付けず --no-owner/--no-acl で復元し、ダンプ内の本番ロールや
  # CREATE/DROP DATABASE の TOC エントリを無視する。pg_restore は無害なエラー
  # (既存の拡張など) を報告しうるため、その終了ステータスでスクリプト全体を
  # 止めない。
  pg_restore --no-owner --no-acl -d "$DEV_DB" "$MASKED_DUMP" \
    || echo "    pg_restore reported errors (often ignorable: roles/extensions); continuing"

  # (6) Apply pending dbmate migrations so the schema matches what the local
  # apps expect. DATABASE_URL is already injected by op run, so call dbmate
  # directly (calling 'make db-migrate' would nest op run). Skip the schema dump
  # so db/schema.sql is not rewritten with this snapshot's contents.
  #
  # [Ja] (6) ローカルの各アプリが前提とするスキーマに揃えるため、未適用の dbmate
  # マイグレーションを適用する。DATABASE_URL は op run が注入済みなので dbmate を
  # 直接呼ぶ ('make db-migrate' だと op run が入れ子になる)。db/schema.sql が本
  # スナップショットの内容で書き換わらないよう、スキーマのダンプは抑止する。
  echo "==> Applying dbmate migrations"
  DBMATE_NO_DUMP_SCHEMA=true dbmate up
}

# Restore-only mode: reuse the masked dump already saved in tmp/.
#
# [Ja] 再復元モード: tmp/ に保存済みのマスク済みダンプを再利用する。
if [ "${1:-}" = "restore" ]; then
  load_into_dev_db
  echo "==> Done (restored ${DEV_DB} from ${MASKED_DUMP})"
  exit 0
fi

# Full pipeline below. Require the production R2 read-only credentials and the
# backup bucket name (all injected from 1Password via op run).
#
# [Ja] 以下はフルパイプライン。本番 R2 の読み取り専用資格情報とバックアップ
# バケット名 (いずれも op run 経由で 1Password から注入) を要求する。
: "${ANNICT_PROD_S3_ENDPOINT:?ANNICT_PROD_S3_ENDPOINT is required}"
: "${ANNICT_PROD_S3_ACCESS_KEY_ID:?ANNICT_PROD_S3_ACCESS_KEY_ID is required}"
: "${ANNICT_PROD_S3_SECRET_ACCESS_KEY:?ANNICT_PROD_S3_SECRET_ACCESS_KEY is required}"
: "${ANNICT_PROD_S3_BACKUP_BUCKET:?ANNICT_PROD_S3_BACKUP_BUCKET is required}"

# Point rclone at an empty config file (/dev/null) so it does not emit the
# "Config file ... not found" notice on startup. The remote is defined entirely
# by the RCLONE_CONFIG_* variables below, so no real config file is needed.
#
# [Ja] rclone に空の設定ファイル (/dev/null) を指定し、起動時の
# "Config file ... not found" の NOTICE を出さないようにする。リモートは下記の
# RCLONE_CONFIG_* だけで定義しており、実体の設定ファイルは不要なため。
export RCLONE_CONFIG=/dev/null

# Configure an inline rclone S3 remote ("prods3") for the production
# S3-compatible object storage (currently Cloudflare R2) from the
# ANNICT_PROD_S3_* variables, instead of a committed rclone.conf. The rclone
# provider stays "Cloudflare" while the backend is R2.
#
# [Ja] コミット済みの rclone.conf ではなく、ANNICT_PROD_S3_* から本番の S3 互換
# オブジェクトストレージ (現状 Cloudflare R2) 用のインライン rclone S3 リモート
# ("prods3") を構成する。バックエンドが R2 の間は rclone の provider を
# "Cloudflare" のままにする。
export RCLONE_CONFIG_PRODS3_TYPE=s3
export RCLONE_CONFIG_PRODS3_PROVIDER=Cloudflare
export RCLONE_CONFIG_PRODS3_ENDPOINT="$ANNICT_PROD_S3_ENDPOINT"
export RCLONE_CONFIG_PRODS3_ACCESS_KEY_ID="$ANNICT_PROD_S3_ACCESS_KEY_ID"
export RCLONE_CONFIG_PRODS3_SECRET_ACCESS_KEY="$ANNICT_PROD_S3_SECRET_ACCESS_KEY"

mkdir -p "$SNAPSHOT_DIR"

# (1) Download the production backup .tgz and unpack it to backup/export. By
# default the lexicographically last *.tgz is used (backup file names are
# timestamped); override with ANNICT_SNAPSHOT_BACKUP_KEY to pick a specific one.
#
# [Ja] (1) 本番バックアップの .tgz を取得し backup/export に展開する。既定では
# 辞書順で最後の *.tgz を使う (バックアップ名はタイムスタンプ付き)。特定の
# ファイルを使いたい場合は ANNICT_SNAPSHOT_BACKUP_KEY で上書きする。
backup_key="${ANNICT_SNAPSHOT_BACKUP_KEY:-}"
if [ -z "$backup_key" ]; then
  backup_key=$(rclone lsf "prods3:${ANNICT_PROD_S3_BACKUP_BUCKET}" \
    --include "*.tgz" --files-only | sort | tail -n 1)
fi
if [ -z "$backup_key" ]; then
  echo "No .tgz backup found in bucket ${ANNICT_PROD_S3_BACKUP_BUCKET}" >&2
  exit 1
fi

echo "==> Downloading backup: ${backup_key}"
rm -rf "${SNAPSHOT_DIR:?}/backup"
rclone copyto "prods3:${ANNICT_PROD_S3_BACKUP_BUCKET}/${backup_key}" \
  "${SNAPSHOT_DIR}/backup.tgz"

echo "==> Extracting backup"
tar xzf "${SNAPSHOT_DIR}/backup.tgz" -C "$SNAPSHOT_DIR"
if [ ! -f "$EXPORT_FILE" ]; then
  echo "Expected dump not found at ${EXPORT_FILE} after extraction" >&2
  exit 1
fi

# (2) Restore the production dump into a fresh isolated database.
#
# [Ja] (2) 本番ダンプを新規の隔離 DB に復元する。
echo "==> Restoring into isolated ${ANONYMIZE_DB}"
dropdb --if-exists "$ANONYMIZE_DB"
createdb "$ANONYMIZE_DB"
pg_restore --no-owner --no-acl -d "$ANONYMIZE_DB" "$EXPORT_FILE" \
  || echo "    pg_restore reported errors (often ignorable: roles/extensions); continuing"

# (3) Mask PII in the isolated database. ON_ERROR_STOP makes masking failures
# fatal so we never dump and load a partially-masked database.
#
# [Ja] (3) 隔離 DB の PII をマスクする。ON_ERROR_STOP でマスク失敗を致命的に
# 扱い、マスクが中途半端な DB をダンプ・流し込みしないようにする。
echo "==> Masking PII (db/anonymize.sql)"
psql -v ON_ERROR_STOP=1 -d "$ANONYMIZE_DB" -f db/anonymize.sql

# (4) Dump the masked database to tmp/. The isolated database is dropped by the
# cleanup trap on exit.
#
# [Ja] (4) マスク済み DB を tmp/ にダンプする。隔離 DB は終了時の cleanup トラップ
# で破棄される。
echo "==> Writing masked dump: ${MASKED_DUMP}"
pg_dump -Fc --no-owner --no-acl -d "$ANONYMIZE_DB" -f "$MASKED_DUMP"

# (5)-(6) Load the masked dump into the development database and migrate.
#
# [Ja] (5)-(6) マスク済みダンプを開発 DB に流し込み、マイグレーションする。
load_into_dev_db

echo "==> Done (loaded ${DEV_DB} from production snapshot ${backup_key})"
