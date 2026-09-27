#!/usr/bin/env bash
#
# load_snapshot_db.shは本番スナップショットからローカル開発DBを構築する。
# 本番pg_dumpバックアップをCloudflare R2から取得し、隔離した使い捨てDBに
# 復元してPIIをマスクし、マスク結果をtmp/ にダンプしたうえで、それをローカル
# 開発DBに (再) 流し込み、未適用のdbmateマイグレーションを適用する。
#
# 生PIIは隔離DB (annict_anonymize) にしか存在せず、開発DB (annict_development)
# にはマスク済みダンプしか入らない。
#
# SKIPPED_DATA_TABLESに挙げたテーブルの行は取り込まないため、これらのテーブルは
# ローカルには存在するが中身は空になる。
#
# 使い方 (秘密注入のため必ずMakefile経由で実行する):
#   make load-snapshot-db       # フルパイプライン: 取得 -> マスク -> 流し込み
#   make restore-snapshot-db    # 再流し込みのみ (tmp/ のマスク済みダンプから)

set -euo pipefail

# ローカルPostgreSQLの接続設定。db-* のMakefileターゲットと揃える
# (trust認証・パスワード無し)。DB名はローカル開発の慣習に固定する。
export PGHOST="${PGHOST:-postgresql}"
export PGUSER="${PGUSER:-postgres}"
ANONYMIZE_DB="annict_anonymize"
DEV_DB="annict_development"

# 本番ダンプの復元時に行を取り込まないテーブル。いずれも巨大でローカルには
# 不要 (Railsのセッションストア・Delayed::Jobのキュー・Riverのキュー) であり、
# 行を落とすことで本パイプラインのピーク時ディスク使用量を抑えられる。スキップ
# するのはデータのみで、テーブル定義は復元されるため空の状態で存在する。
SKIPPED_DATA_TABLES=(sessions delayed_jobs river_job)

# ダウンロードとマスク済みダンプの作業ディレクトリ (gitignore済みのtmp/)。
# cleanupトラップはbackup/ をまるごと削除するため、実行後に残したくないものは
# その配下に置く。
SNAPSHOT_DIR="tmp/snapshot"
EXPORT_FILE="${SNAPSHOT_DIR}/backup/export"
RESTORE_LIST="${SNAPSHOT_DIR}/backup/restore.list"
MASKED_DUMP="${SNAPSHOT_DIR}/masked.dump"

# 終了時に隔離DBを落とし、取得した生ダンプも削除して、途中で失敗しても
# 生PIIが残らないようにする。再復元はtmp/ のマスク済みダンプを使うため、ここで
# 生ダンプを削除して問題ない。
cleanup() {
  dropdb --if-exists "$ANONYMIZE_DB" >/dev/null 2>&1 || true
  rm -rf "${SNAPSHOT_DIR:?}/backup" "${SNAPSHOT_DIR}/backup.tgz"
}
trap cleanup EXIT

# (5) マスク済みダンプからローカル開発DBを再構築し、マイグレーションを
# 適用する。フルパイプラインと再復元専用エントリの両方から使う。
load_into_dev_db() {
  if [ ! -f "$MASKED_DUMP" ]; then
    echo "マスク済みダンプが見つかりません: ${MASKED_DUMP}。先に 'make load-snapshot-db' を実行してください。" >&2
    exit 1
  fi

  echo "==> ${DEV_DB} を削除 (存在しない場合はスキップ)"
  dropdb --if-exists "$DEV_DB"

  echo "==> ${DEV_DB} を作成"
  createdb "$DEV_DB"

  echo "==> マスク済みダンプを ${DEV_DB} に復元"
  # -Cを付けず --no-owner/--no-aclで復元し、ダンプ内の本番ロールや
  # CREATE/DROP DATABASEのTOCエントリを無視する。pg_restoreは無害なエラー
  # (既存の拡張など) を報告しうるため、その終了ステータスでスクリプト全体を
  # 止めない。
  pg_restore --no-owner --no-acl -d "$DEV_DB" "$MASKED_DUMP" \
    || echo "    pg_restoreがエラーを報告しました (ロールや拡張など、無視できるものが大半)。処理を続行します"

  # (6) ローカルの各アプリが前提とするスキーマに揃えるため、未適用のdbmate
  # マイグレーションを適用する。DATABASE_URLはop runが注入済みなのでdbmateを
  # 直接呼ぶ ('make db-migrate' だとop runが入れ子になる)。db/schema.sqlが本
  # スナップショットの内容で書き換わらないよう、スキーマのダンプは抑止する。
  echo "==> dbmateマイグレーションを適用"
  DBMATE_NO_DUMP_SCHEMA=true dbmate up
}

# 再復元モード: tmp/ に保存済みのマスク済みダンプを再利用する。
if [ "${1:-}" = "restore" ]; then
  load_into_dev_db
  echo "==> 完了 (${MASKED_DUMP} から ${DEV_DB} を復元しました)"
  exit 0
fi

# 以下はフルパイプライン。本番R2の読み取り専用資格情報とバックアップ
# バケット名 (いずれもop run経由で1Passwordから注入) を要求する。
: "${ANNICT_PROD_S3_ENDPOINT:?ANNICT_PROD_S3_ENDPOINTが未設定です}"
: "${ANNICT_PROD_S3_ACCESS_KEY_ID:?ANNICT_PROD_S3_ACCESS_KEY_IDが未設定です}"
: "${ANNICT_PROD_S3_SECRET_ACCESS_KEY:?ANNICT_PROD_S3_SECRET_ACCESS_KEYが未設定です}"
: "${ANNICT_PROD_S3_BACKUP_BUCKET:?ANNICT_PROD_S3_BACKUP_BUCKETが未設定です}"

# rcloneに空の設定ファイル (/dev/null) を指定し、起動時の
# "Config file ... not found" のNOTICEを出さないようにする。リモートは下記の
# RCLONE_CONFIG_* だけで定義しており、実体の設定ファイルは不要なため。
export RCLONE_CONFIG=/dev/null

# コミット済みのrclone.confではなく、ANNICT_PROD_S3_* から本番のS3互換
# オブジェクトストレージ (現状Cloudflare R2) 用のインラインrclone S3リモート
# ("prods3") を構成する。バックエンドがR2の間はrcloneのproviderを
# "Cloudflare" のままにする。
export RCLONE_CONFIG_PRODS3_TYPE=s3
export RCLONE_CONFIG_PRODS3_PROVIDER=Cloudflare
export RCLONE_CONFIG_PRODS3_ENDPOINT="$ANNICT_PROD_S3_ENDPOINT"
export RCLONE_CONFIG_PRODS3_ACCESS_KEY_ID="$ANNICT_PROD_S3_ACCESS_KEY_ID"
export RCLONE_CONFIG_PRODS3_SECRET_ACCESS_KEY="$ANNICT_PROD_S3_SECRET_ACCESS_KEY"

mkdir -p "$SNAPSHOT_DIR"

# (1) 本番バックアップの .tgzを取得しbackup/exportに展開する。既定では
# 辞書順で最後の *.tgzを使う (バックアップ名はタイムスタンプ付き)。特定の
# ファイルを使いたい場合はANNICT_SNAPSHOT_BACKUP_KEYで上書きする。
backup_key="${ANNICT_SNAPSHOT_BACKUP_KEY:-}"
if [ -z "$backup_key" ]; then
  backup_key=$(rclone lsf "prods3:${ANNICT_PROD_S3_BACKUP_BUCKET}" \
    --include "*.tgz" --files-only | sort | tail -n 1)
fi
if [ -z "$backup_key" ]; then
  echo "バケット ${ANNICT_PROD_S3_BACKUP_BUCKET} に .tgzのバックアップが見つかりません" >&2
  exit 1
fi

echo "==> バックアップをダウンロード: ${backup_key}"
rm -rf "${SNAPSHOT_DIR:?}/backup"
rclone copyto "prods3:${ANNICT_PROD_S3_BACKUP_BUCKET}/${backup_key}" \
  "${SNAPSHOT_DIR}/backup.tgz"

echo "==> バックアップを展開"
tar xzf "${SNAPSHOT_DIR}/backup.tgz" -C "$SNAPSHOT_DIR"
if [ ! -f "$EXPORT_FILE" ]; then
  echo "展開後に想定の位置にダンプが見つかりません: ${EXPORT_FILE}" >&2
  exit 1
fi

# 展開が済んだ時点で書庫を削除する。直後の復元がパイプラインのディスク使用量
# のピークであり、そこで必要なのは展開後のダンプだけであるため。
rm -f "${SNAPSHOT_DIR}/backup.tgz"

# (2) 本番ダンプを新規の隔離DBに復元する。
echo "==> ${ANONYMIZE_DB} を削除 (存在しない場合はスキップ)"
dropdb --if-exists "$ANONYMIZE_DB"

echo "==> ${ANONYMIZE_DB} を作成"
createdb "$ANONYMIZE_DB"

# 書庫の目次からSKIPPED_DATA_TABLESのTABLE DATAエントリを除いた復元
# リストを組み立てる。pg_restoreには --exclude-table-dataが無いため、-lの一覧を
# 加工して -Lで渡すのが、テーブル定義を残したまま行だけを落とす方法になる。
skipped_data_pattern=$(IFS='|'; echo "${SKIPPED_DATA_TABLES[*]}")
pg_restore -l "$EXPORT_FILE" \
  | grep -Ev "TABLE DATA public (${skipped_data_pattern})( |$)" > "$RESTORE_LIST"

echo "==> 隔離DB ${ANONYMIZE_DB} に復元 (データを取り込まないテーブル: ${SKIPPED_DATA_TABLES[*]})"
pg_restore --no-owner --no-acl -L "$RESTORE_LIST" -d "$ANONYMIZE_DB" "$EXPORT_FILE" \
  || echo "    pg_restoreがエラーを報告しました (ロールや拡張など、無視できるものが大半)。処理を続行します"

# (3) 隔離DBのPIIをマスクする。ON_ERROR_STOPでマスク失敗を致命的に
# 扱い、マスクが中途半端なDBをダンプ・流し込みしないようにする。
echo "==> PIIをマスク (db/anonymize.sql)"
psql -v ON_ERROR_STOP=1 -d "$ANONYMIZE_DB" -f db/anonymize.sql

# (4) マスク済みDBをtmp/ にダンプし、隔離DBを直後に破棄する。終了時の
# cleanupトラップでも破棄されるが、(5) で同じ規模のDBをもう1つ復元するため、
# 両方を同時に抱えることがディスク不足の原因になる。
echo "==> マスク済みダンプを書き出し: ${MASKED_DUMP}"
pg_dump -Fc --no-owner --no-acl -d "$ANONYMIZE_DB" -f "$MASKED_DUMP"

echo "==> ${ANONYMIZE_DB} を削除"
dropdb "$ANONYMIZE_DB"

# (5)-(6) マスク済みダンプを開発DBに流し込み、マイグレーションする。
load_into_dev_db

echo "==> 完了 (本番スナップショット ${backup_key} から ${DEV_DB} を構築しました)"
