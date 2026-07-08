#!/usr/bin/env bash
#
# browse.sh drives playwright-cli for browser verification of the dev site:
# it generates the Basic-auth config, runs the multi-step dev sign-in, reuses
# the logged-in session for screenshots, and cleans up.
#
# It expects KORYLUS_BROWSING_* and ANNICT_TURNSTILE_DISABLE in the environment,
# so run it under the op-run wrapper (see the browse-* targets in go/Makefile).
# Reading credentials through op-run avoids evaluating the .env in a shell,
# which would corrupt any credential containing a `$`.
#
# [Ja] browse.sh は playwright-cli を駆動して dev サイトのブラウザ確認を行う。
# Basic 認証 config の生成・マルチステップの dev サインイン・ログイン済み
# セッションでのスクショ・後片付けをまとめる。
#
# KORYLUS_BROWSING_* と ANNICT_TURNSTILE_DISABLE が環境にある前提なので、op run
# ラッパー配下 (go/Makefile の browse-* ターゲット) から実行する。creds を op run
# 経由で読むことで、.env をシェル評価して `$` を含む creds を壊すのを避ける。
set -euo pipefail

SESSION=dev
TMP_DIR=/workspace/tmp
CONFIG_FILE="$TMP_DIR/browse-cli.config.json"
ORIGIN_FILE="$TMP_DIR/browse-cli.origin"
PROFILE_DIR="$TMP_DIR/browse-cli-profile"
SHOT_DIR="$TMP_DIR/browse"

pw() { playwright-cli -s="$SESSION" "$@"; }

# build_config writes the Basic-auth config (httpCredentials) parsed from
# KORYLUS_BROWSING_BASE_URL, plus a credential-free origin file for later
# navigation. The config carries the credentials, so it is written 0600 under
# the gitignored tmp dir and removed as soon as login captures it in the
# browser context.
#
# [Ja] build_config は KORYLUS_BROWSING_BASE_URL から Basic 認証 config
# (httpCredentials) を生成し、以降の遷移用に creds を抜いた origin ファイルも書く。
# config は creds を含むため gitignore 済み tmp に 0600 で書き、ログインが
# ブラウザコンテキストに取り込んだ直後に削除する。
build_config() {
  mkdir -p "$TMP_DIR"
  node -e '
    const fs = require("fs");
    const raw = process.env.KORYLUS_BROWSING_BASE_URL || "";
    if (!raw) { console.error("KORYLUS_BROWSING_BASE_URL is not set"); process.exit(1); }
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
  local email="${!email_var:-}"
  local pass="${!pass_var:-}"
  if [ -z "$email" ] || [ -z "$pass" ]; then
    echo "USER${n} credentials are not set (${email_var} / ${pass_var})" >&2
    exit 1
  fi

  # Remove the credential-bearing config on any exit, so a mid-login failure
  # (a set -e abort before the explicit rm below) never leaves credentials at
  # rest.
  #
  # [Ja] creds を含む config をどの終了経路でも削除し、ログイン途中の失敗
  # (下の明示 rm へ到達する前の set -e abort) でも creds をディスクに残さない。
  trap 'rm -f "$CONFIG_FILE"' EXIT

  build_config
  local origin
  origin="$(cat "$ORIGIN_FILE")"

  # Basic auth is passed via the config (httpCredentials); the persistent
  # profile keeps the login cookies on disk so a still-running session survives
  # across separate shell invocations.
  #
  # [Ja] Basic 認証は config (httpCredentials) で渡す。永続プロファイルはログイン
  # Cookie をディスクに残し、起動中のセッションが別々のシェル呼び出しをまたいで
  # 生き続けられるようにする。
  pw open "$origin/sign_in" --browser=chromium --persistent --profile="$PROFILE_DIR" --config="$CONFIG_FILE" >/dev/null

  # Multi-step sign-in is project-specific: the email step navigates to the
  # password step, which submits the login. Turnstile must be disabled
  # (ANNICT_TURNSTILE_DISABLE) or the submit is blocked.
  #
  # [Ja] マルチステップのサインインはプロジェクト固有: Email ステップが
  # password ステップへ遷移し、そこでログインを送信する。Turnstile は無効化
  # (ANNICT_TURNSTILE_DISABLE) されている必要があり、でないと送信が弾かれる。
  pw fill "getByRole('textbox', { name: 'Email' })" "$email" --submit >/dev/null
  pw fill "input[type=password]" "$pass" --submit >/dev/null

  # The context now holds the credentials, so the on-disk config is no longer
  # needed; drop it to avoid leaving credentials at rest.
  #
  # [Ja] コンテキストが creds を保持したので、ディスク上の config はもう不要。
  # creds を残さないため削除する。
  rm -f "$CONFIG_FILE"

  local state
  state="$(pw --raw run-code "async page => { await page.waitForLoadState('networkidle'); return page.url() + ' signed_in=' + (await page.locator('a[href*=sign_out], form[action*=sign_out]').count() > 0); }")"
  echo "logged in as USER${n}: $state"
}

cmd_shot() {
  local path="${1:-/}"
  if [ ! -f "$ORIGIN_FILE" ]; then
    echo "no active session; run 'browse.sh login' first" >&2
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
  echo "screenshot: $SHOT_DIR/$name.png"
}

cmd_close() {
  pw close >/dev/null 2>&1 || true
  playwright-cli close-all >/dev/null 2>&1 || true
  rm -f "$CONFIG_FILE" "$ORIGIN_FILE"
  rm -rf "$PROFILE_DIR"
  echo "browser session closed and temp files removed"
}

case "${1:-}" in
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
    echo "usage: browse.sh {login [user_number] | shot <path> | close}" >&2
    exit 2
    ;;
esac
