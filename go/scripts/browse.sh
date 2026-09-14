#!/usr/bin/env bash
#
# browse.shはplaywright-cliを駆動してdevサイトのブラウザ確認を行う。
# Basic認証configの生成・マルチステップのdevサインイン・ログイン済み
# セッションでのスクショ・後片付けをまとめる。
#
# KORYLUS_BROWSING_* とANNICT_TURNSTILE_DISABLEが環境にある前提なので、op run
# ラッパー配下 (go/Makefileのbrowse-* ターゲット) から実行する。credsをop run
# 経由で読むことで、.envをシェル評価して `$` を含むcredsを壊すのを避ける。
set -euo pipefail

SESSION=dev
TMP_DIR=/workspace/tmp
CONFIG_FILE="$TMP_DIR/browse-cli.config.json"
ORIGIN_FILE="$TMP_DIR/browse-cli.origin"
PROFILE_DIR="$TMP_DIR/browse-cli-profile"
SHOT_DIR="$TMP_DIR/browse"

pw() { playwright-cli -s="$SESSION" "$@"; }

# check_environmentは資格情報の値を表示せず、ブラウザ確認に使う環境を検証する。
# 診断とログインをbrowse.shに集約し、同じop runラッパーと要件を通すため。
check_environment() {
  local n="${1:-1}"
  local email_var="KORYLUS_BROWSING_USER${n}_EMAIL"
  local pass_var="KORYLUS_BROWSING_USER${n}_PASSWORD"
  local failed=0

  if [ "${APP_ENV:-}" != "dev" ]; then
    echo "APP_ENVはdevである必要があります (Makefileのbrowseターゲット経由で実行してください)" >&2
    failed=1
  fi

  local name
  for name in KORYLUS_BROWSING_BASE_URL "$email_var" "$pass_var"; do
    if [ -z "${!name:-}" ]; then
      echo "$nameが未設定です" >&2
      failed=1
    fi
  done

  if [ "${ANNICT_TURNSTILE_DISABLE:-}" != "true" ]; then
    echo "devのブラウザログインにはANNICT_TURNSTILE_DISABLEにtrueが必要です" >&2
    failed=1
  fi

  local command_name
  for command_name in node playwright-cli; do
    if ! command -v "$command_name" >/dev/null 2>&1; then
      echo "$command_nameが使えません" >&2
      failed=1
    fi
  done

  if [ -n "${KORYLUS_BROWSING_BASE_URL:-}" ] && command -v node >/dev/null 2>&1; then
    if ! node -e '
      const raw = process.env.KORYLUS_BROWSING_BASE_URL || "";
      let url;
      try {
        url = new URL(raw);
      } catch {
        console.error("KORYLUS_BROWSING_BASE_URLが正しいURLではありません");
        process.exit(1);
      }
      if (url.protocol !== "http:" && url.protocol !== "https:") {
        console.error("KORYLUS_BROWSING_BASE_URLはhttpまたはhttpsである必要があります");
        process.exit(1);
      }
      if (!url.username || !url.password) {
        console.error("KORYLUS_BROWSING_BASE_URLにBasic認証の資格情報が含まれていません");
        process.exit(1);
      }
    '; then
      failed=1
    fi
  fi

  if [ "$failed" -ne 0 ]; then
    return 1
  fi
}

# build_configはKORYLUS_BROWSING_BASE_URLからBasic認証config
# (httpCredentials) を生成し、以降の遷移用にcredsを抜いたoriginファイルも書く。
# configはcredsを含むためgitignore済みtmpに0600で書き、ログインが
# ブラウザコンテキストに取り込んだ直後に削除する。
build_config() {
  mkdir -p "$TMP_DIR"
  node -e '
    const fs = require("fs");
    const raw = process.env.KORYLUS_BROWSING_BASE_URL || "";
    if (!raw) { console.error("KORYLUS_BROWSING_BASE_URLが未設定です"); process.exit(1); }
    const u = new URL(raw);
    const cfg = { browser: { contextOptions: { httpCredentials: {
      username: decodeURIComponent(u.username),
      password: decodeURIComponent(u.password),
    } } } };
    fs.writeFileSync(process.argv[1], JSON.stringify(cfg), { mode: 0o600 });
    u.username = u.password = "";
    fs.writeFileSync(process.argv[2], u.origin);
  ' "$CONFIG_FILE" "$ORIGIN_FILE"
}

cmd_login() {
  local n="${1:-1}"
  local email_var="KORYLUS_BROWSING_USER${n}_EMAIL"
  local pass_var="KORYLUS_BROWSING_USER${n}_PASSWORD"
  check_environment "$n"

  local email="${!email_var:-}"
  local pass="${!pass_var:-}"

  # credsを含むconfigをどの終了経路でも削除し、ログイン途中の失敗
  # (下の明示rmへ到達する前のset -e abort) でもcredsをディスクに残さない。
  trap 'rm -f "$CONFIG_FILE"' EXIT

  build_config
  local origin
  origin="$(cat "$ORIGIN_FILE")"

  # Basic認証はconfig (httpCredentials) で渡す。永続プロファイルはログイン
  # Cookieをディスクに残し、起動中のセッションが別々のシェル呼び出しをまたいで
  # 生き続けられるようにする。
  pw open "$origin/sign_in" --browser=chromium --persistent --profile="$PROFILE_DIR" --config="$CONFIG_FILE" >/dev/null

  # マルチステップのサインインはプロジェクト固有: Emailステップが
  # passwordステップへ遷移し、そこでログインを送信する。Turnstileは無効化
  # (ANNICT_TURNSTILE_DISABLE) されている必要があり、でないと送信が弾かれる。
  pw fill "getByRole('textbox', { name: 'Email' })" "$email" --submit >/dev/null
  pw fill "input[type=password]" "$pass" --submit >/dev/null

  # コンテキストがcredsを保持したので、ディスク上のconfigはもう不要。
  # credsを残さないため削除する。
  rm -f "$CONFIG_FILE"

  local state
  state="$(pw --raw run-code "async page => { await page.waitForLoadState('networkidle'); return page.url() + ' signed_in=' + (await page.locator('a[href*=sign_out], form[action*=sign_out]').count() > 0); }")"
  echo "USER${n}でログインしました: $state"
}

cmd_check() {
  local n="${1:-1}"
  check_environment "$n"
  echo "USER${n}のブラウザ環境は準備できています"
}

cmd_shot() {
  local path="${1:-/}"
  if [ ! -f "$ORIGIN_FILE" ]; then
    echo "有効なセッションがありません。先に 'browse.sh login' を実行してください" >&2
    exit 1
  fi
  mkdir -p "$SHOT_DIR"
  local origin
  origin="$(cat "$ORIGIN_FILE")"
  local name
  name="$(printf '%s' "$path" | sed 's#[^a-zA-Z0-9]#_#g; s#^_*##')"
  [ -n "$name" ] || name=home

  pw goto "$origin$path" >/dev/null
  pw run-code "async page => page.waitForLoadState('networkidle')" >/dev/null
  pw screenshot --filename="$SHOT_DIR/$name.png" >/dev/null
  echo "スクリーンショット: $SHOT_DIR/$name.png"
}

cmd_close() {
  pw close >/dev/null 2>&1 || true
  playwright-cli close-all >/dev/null 2>&1 || true
  rm -f "$CONFIG_FILE" "$ORIGIN_FILE"
  rm -rf "$PROFILE_DIR"
  echo "ブラウザセッションを閉じ、一時ファイルを削除しました"
}

case "${1:-}" in
  check)
    shift
    cmd_check "${1:-1}"
    ;;
  login)
    shift
    cmd_login "${1:-1}"
    ;;
  shot)
    shift
    cmd_shot "${1:-/}"
    ;;
  close)
    cmd_close
    ;;
  *)
    echo "使い方: browse.sh {check [user_number] | login [user_number] | shot <path> | close}" >&2
    exit 2
    ;;
esac
