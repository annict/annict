---
name: fix-dependabot-alert
description: GitHub の Dependabot alert URL を受け取り、対象パッケージのアップデート・検証・コミット・コードレビューを一括で行う。レビュー本体は `/review` スキルに委譲する。
argument-hint: "<Dependabot alert URL>"
---

# Dependabot alert 対応 <Dependabot alert URL>

GitHub の Dependabot alert URL を受け取り、以下を自動で行ってください:

1. Alert 情報の取得
2. 対象パッケージとプロジェクトの特定
3. 新規ブランチの作成
4. パッケージのアップデート(直接依存ならそのまま、間接依存ならロックファイル更新→ダメなら parent もアップデート)
5. 対象プロジェクトの「コミット前に実行するコマンド」での検証
6. コミット作成
7. PR 情報・対応内容を一時ドキュメントにまとめて `/review` スキルに渡す
8. `/review` のレビュー結果に基づいてマージ可否を判断し報告する

## 引数

`$ARGUMENTS` は以下の形式で渡されます:

- 形式: `<Dependabot alert URL>`
- 例: `/fix-dependabot-alert https://github.com/owner/repo/security/dependabot/152`

不正な URL が渡された場合はエラーを報告して終了してください。

## 原則

- **プロジェクト固有の情報を含めない**: このスキルは複数プロジェクトで共有されるため、プロジェクト固有の絶対パスや専用コマンドを埋め込まない。必要な情報は各プロジェクトの `CLAUDE.md` や `Makefile` から動的に判断する
- **push と PR 作成は行わない**: コミットまでで止め、push と PR 作成はユーザーが手動で実施する
- **なるべく対象パッケージだけアップデートする**: 関係ないパッケージまで巻き込まないよう、最小限のアップデートに留める
- **コマンド完了後は作業ブランチに留まる**: 元のブランチには戻らない。ユーザーが作業ブランチで継続作業できる状態で終了する

## 手順

### ステップ 1: 引数の解析

- `$ARGUMENTS` から Dependabot alert URL を取得する
- `https://github.com/<owner>/<repo>/security/dependabot/<number>` の形式でない場合はエラーを報告して終了する
- URL から `<owner>`, `<repo>`, `<number>` を抽出する

### ステップ 2: 作業ディレクトリのクリーン状態チェック

**重要**: GitHub への API 呼び出しを行う前に、ローカルの状態を確認する。

- `git status --porcelain` で未コミット変更がないか確認する
- 未コミット変更がある場合はエラーを報告して終了する
  - ユーザーに `git commit` または `git stash` を促すメッセージを表示する
  - GitHub への不要な API リクエストを避けるため、Alert 情報の取得より前にローカルで弾く
- 現在のブランチ名を `git rev-parse --abbrev-ref HEAD` で取得し、後の `/review` 呼び出しの比較対象ブランチとして記録する

### ステップ 3: Alert 情報の取得

```sh
gh api repos/<owner>/<repo>/dependabot/alerts/<number>
```

取得する情報:

- `state`: alert の状態(`open` / `dismissed` / `auto_dismissed` / `fixed`)
- `dependency.package.ecosystem`: パッケージエコシステム(`npm`, `rubygems`, `go`, ...)
- `dependency.package.name`: パッケージ名
- `dependency.manifest_path`: 検出ファイルのパス(プロジェクト判定に使う)
- `dependency.scope`: `development` / `runtime`
- `dependency.relationship`: `direct` / `transitive`
- `security_advisory.severity`: severity(`critical` / `high` / `medium` / `low`)
- `security_advisory.summary`: 概要
- `security_advisory.description`: 詳細
- `security_advisory.cve_id` / `security_advisory.ghsa_id`: CVE/GHSA ID
- `security_vulnerability.first_patched_version.identifier`: 修正バージョン
- `security_vulnerability.vulnerable_version_range`: 脆弱なバージョン範囲
- `html_url`: alert の HTML URL

**state チェック**: `state` が `open` でない場合は警告を表示して終了する。

**first_patched_version チェック**: 修正バージョンが存在しない場合は自動アップデートが不可能なため、その旨を一時ドキュメントに記載してレビューに進む(ステップ 9 へ)。アップデート手順(ステップ 5〜8)はスキップする。

### ステップ 4: 対象プロジェクトとパッケージマネージャーの判定

`dependency.manifest_path` から対象プロジェクトを判定する。

- `manifest_path` の最初のディレクトリ(例: `rails`, `go`)を「対象プロジェクトのルート」として扱う
- ルート直下に manifest がある場合(例: `package.json`)はリポジトリ全体を対象プロジェクトとする
- 対象プロジェクトのルートに `CLAUDE.md` がある場合は、そこから「コミット前に実行するコマンド」を読み取る(ステップ 7 で使用)

`dependency.package.ecosystem` から使用するパッケージマネージャーを判定する。実際のコマンドは対象プロジェクトの慣例(`Makefile`、lockfile、`package.json` の `packageManager` フィールドなど)から動的に判断する。

| ecosystem      | 想定されるコマンド例(プロジェクトに応じて判断)                  |
| -------------- | --------------------------------------------------------------- |
| `npm`          | `pnpm update <pkg>` / `npm update <pkg>` / `yarn upgrade <pkg>` |
| `rubygems`     | `bundle update <pkg> --conservative`                            |
| `go` (`gomod`) | `go get <pkg>@<version> && go mod tidy`                         |
| `pip`          | `pip install --upgrade <pkg>`                                   |
| `composer`     | `composer update <pkg>`                                         |

### ステップ 5: 新規ブランチの作成

ベースブランチ(ステップ 2 で記録した現在のブランチ)から新規ブランチを作成する。

```sh
git checkout -b fix-dependabot-alert/<番号>
```

ブランチ名は `fix-dependabot-alert/<alert 番号>` で固定する。Dependabot 自身が作る `dependabot/...` プレフィックスとの衝突を避けるため、別プレフィックスを使う。

### ステップ 6: パッケージのアップデート

#### 直接依存(`relationship == "direct"`)の場合

対象プロジェクトのルートで、ecosystem に応じたコマンドを実行して対象パッケージのみをアップデートする。

- 必ず `first_patched_version` 以上のバージョンに上がることを目標にする
- アップデート後、ロックファイル(`pnpm-lock.yaml` / `Gemfile.lock` / `go.sum` など)を読んで実際のバージョンが目標を満たしているか確認する

#### 間接依存(`relationship == "transitive"`)の場合

以下の順で試行する:

**ステップ 6-A: ロックファイルだけアップデートを試みる**

- npm/pnpm の場合: `pnpm update <pkg>` を試す。`pnpm update` で transitive 依存が更新されない場合は `pnpm update --depth Infinity <pkg>` も試す
- bundler の場合: `bundle update <pkg> --conservative`(`--conservative` で他の gem を巻き込まない)
- go の場合: `go get <pkg>@<version> && go mod tidy`

ロックファイルを読み、対象パッケージのバージョンが `first_patched_version` 以上になっているかを確認する(例: `pnpm-lock.yaml` / `Gemfile.lock` / `go.sum` を grep)。

**ステップ 6-B: ロックファイル更新で解決しない場合は parent パッケージもアップデートを試みる**

- npm/pnpm: `pnpm why <pkg>` で依存元を特定
- bundler: `bundle info <pkg>` または `gem dependency <pkg>` で依存元を特定
- go: `go mod why -m <pkg>` で依存元を特定

特定した parent のうち、対象パッケージを transitive で含むものを 1 つずつアップデートする。アップデート後、再度ロックファイルで対象パッケージのバージョンを確認する。複数の parent がある場合は、まず `direct` 依存になっているものから順に試す。

**ステップ 6-C: それでも解決しない場合**

- 自動アップデート不能と判断し、一時ドキュメントの「アップデート不能」セクションに以下を記載してステップ 9 に進む:
  - 試行した内容(コマンドと結果)
  - 推奨される手動対応(parent パッケージのリリース待ち / pnpm overrides / Gemfile での直接指定 / alert の dismiss など)
- 検証(ステップ 7)とコミット(ステップ 8)はスキップする

### ステップ 7: 検証コマンドの実行

対象プロジェクトの `CLAUDE.md` の「コミット前に実行するコマンド」セクションを参照し、すべてのチェックを順に実行する。コマンドはプロジェクトごとに異なるため、必ず `CLAUDE.md` から動的に読み取ること。

各コマンドの成否を記録する。失敗したコマンドがあっても、修正ループは行わずに次のステップへ進む(失敗内容は一時ドキュメントに残してレビューに委ねる)。

`CLAUDE.md` に明示的なコマンドがない場合は、対象プロジェクトの `Makefile` のターゲット(`fmt`, `lint`, `test`, `build` など)を確認してそれらを実行する。

### ステップ 8: コミットの作成

アップデートで変更されたファイルをまとめて 1 つのコミットにする。同じ alert に対するアップデートは 1 コミットで完結させる。

**コミットメッセージ**(日本語):

```
<package> を <old> から <new> に更新

GHSA-XXXX-YYYY-ZZZZ / CVE-XXXX-NNNNN への対応。
- severity: <severity>
- alert: #<番号>
- <summary を 1 行で>
```

例:

```
flatted を 3.4.1 から 3.4.2 に更新

GHSA-rf6f-7fwh-wjgh / CVE-2026-33228 への対応。
- severity: high
- alert: #152
- Prototype Pollution via parse() in NodeJS flatted
```

**コミットルール**:

- `--no-verify` は使わない。pre-commit hook が失敗したら根本原因を修正するか、修正できないなら一時ドキュメントに記載してレビューに委ねる
- **push はしない**。push はユーザーが後で手動で実施する
- 検証(ステップ 7)で失敗があった場合でも、アップデートの変更自体はコミットする(失敗内容は一時ドキュメントに記載)

### ステップ 9: 一時ドキュメントの作成

`tmp/fix-dependabot-alert/alert-<番号>.md` に alert 情報・対応内容・検証結果をまとめた一時ドキュメントを作成する。

- `tmp/fix-dependabot-alert/` ディレクトリが存在しない場合は `mkdir -p` で作成する
- `<番号>` は alert 番号

ドキュメントの内容:

```markdown
# Dependabot Alert #<番号>: <CVE / GHSA>

## 基本情報

- **Alert 番号**: #<番号>
- **対象パッケージ**: `<package>` (<ecosystem>)
- **manifest path**: `<manifest_path>`
- **scope**: <development / runtime>
- **relationship**: <direct / transitive>
- **severity**: <severity>
- **CVE**: <cve_id>
- **GHSA**: <ghsa_id>
- **summary**: <summary>
- **vulnerable range**: <vulnerable_version_range>
- **first patched version**: <first_patched_version>
- **URL**: <html_url>

## 概要

<security_advisory.description から要約>

## 対応内容

- **アップデート方法**: <ロックファイルのみ更新 / parent もアップデート / 自動アップデート不能>
- **変更されたパッケージ**:
  - `<package>` <old> → <new>
  - (parent もアップデートした場合)
  - `<parent package>` <old> → <new>
- **変更されたファイル**:
  - <ファイル名>
  - ...

## 検証結果

- **対象プロジェクト**: <プロジェクトのパス>
- **実行コマンド**:
  - `<コマンド 1>`: <成功 / 失敗>
  - `<コマンド 2>`: <成功 / 失敗>
  - ...
- **失敗があった場合**:
  - `<コマンド名>`: <失敗内容の要約>

## アップデート不能な場合(該当する場合のみ)

- **理由**: <ロックファイル更新が効かない / parent もアップデート不能 / first_patched_version なし / その他>
- **試行した内容**:
  - <試行 1>
  - <試行 2>
- **推奨される手動対応**:
  - <ユーザーが手動で行うべき内容>

## コミット

- <コミット 1 のメッセージ>
```

この一時ドキュメントを `/review` スキルに参考ドキュメントとして渡すことで、レビュー実行時に Alert 情報・CVE・対応方針・検証結果を考慮したレビューが可能になる。

### ステップ 10: `/review` スキルの呼び出し

Skill ツール経由で `/review` を呼び出し、以下の引数を渡す:

- 比較対象ブランチ名: ステップ 2 で記録した「元のブランチ名」
- 参考ドキュメントのパス: ステップ 9 で作成した `tmp/fix-dependabot-alert/alert-<番号>.md`

引数形式: `<元のブランチ名> tmp/fix-dependabot-alert/alert-<番号>.md`

`/review` スキルがローカルブランチ(チェックアウト済みのアップデート用ブランチ)を対象にレビューを実行し、`docs/private/reviews/` 配下にレビュードキュメントを作成する。

**自動アップデート不能で終了した場合**: 変更がないため `/review` の呼び出しはスキップし、ステップ 12 の報告だけを行う。

### ステップ 11: レビュー結果の受け取り

`/review` スキルから以下を受け取る:

- レビュードキュメントのパス: `docs/private/reviews/<ファイル名>.md`
- 評価: Approve / Request Changes / Comment
- 主要な問題点のサマリー

### ステップ 12: マージ可否の判断と報告

「マージ可否判断の基準」に従ってマージ可否を決定し、「報告フォーマット」に従って結果を報告する。

## マージ可否判断の基準

| 状況                                                                     | 判断            |
| ------------------------------------------------------------------------ | --------------- |
| 直接依存のパッチ/マイナー、検証成功、`/review` で問題なし                | Approve         |
| 間接依存のロックファイルのみ更新、検証成功、`/review` で問題なし         | Approve         |
| parent もアップデート、検証成功、`/review` で問題なし                    | Approve         |
| 検証コマンドに失敗があるが、`/review` で軽微な問題のみ                   | Comment         |
| 検証コマンドに失敗があり、`/review` でも問題あり                         | Request Changes |
| メジャーバージョンアップで破壊的変更を含む                               | Comment         |
| 自動アップデート不能(parent も更新できない / first_patched_version なし) | Comment         |
| `/review` が Request Changes を出した                                    | Request Changes |

**判断の補助情報**:

- Dependabot alert はセキュリティ修正であるため、検証が通って破壊的変更がなければ基本は Approve 寄りで判断する
- 検証で失敗があっても、その内容が「アップデートに起因しない既存の問題」なら Approve に近い判断にしてよい(根拠を報告に書く)
- `severity` が `critical` や `high` の場合は、多少の検証失敗があっても適用を優先する判断もあり得る(`/review` 結果と総合的に判断)

## 報告フォーマット

### アップデート成功 + レビュー完了の場合

```
## Dependabot Alert 対応完了

**Alert**: #<番号> <CVE / GHSA>
**対象パッケージ**: `<package>` (<ecosystem>)
**変更**: <old> → <new>
**ブランチ**: fix-dependabot-alert/<番号> ← <元のブランチ名>
**対応方法**: <ロックファイルのみ / parent もアップデート>
**検証結果**: <すべて成功 / 失敗あり>

**評価**: [Approve / Request Changes / Comment]
**判断理由**: <理由>

**レビュードキュメント**: docs/private/reviews/<ファイル名>.md
**参考ドキュメント**: tmp/fix-dependabot-alert/alert-<番号>.md

**主要な確認ポイント**:
- ...

**次のアクション**:
1. 現在のブランチ: `fix-dependabot-alert/<番号>`(作業ブランチにいます)
2. コミット内容を確認してください: `git log <元のブランチ名>..HEAD`
3. 問題なければ push してください: `git push -u origin fix-dependabot-alert/<番号>`
4. PR を作成してください: `gh pr create`
5. CI が通ることを GitHub 上で確認してください
```

### 検証失敗ありの場合

成功時の報告に加えて、以下を追記する:

```
**検証失敗の詳細**:
- `<コマンド名>`: <失敗内容の要約>
- ...

**判断**: <失敗を許容して Approve / Comment / Request Changes> + 理由
```

### 自動アップデート不能で終了した場合

```
## Dependabot Alert 対応不能

**Alert**: #<番号> <CVE / GHSA>
**対象パッケージ**: `<package>` (<ecosystem>)
**relationship**: <direct / transitive>
**理由**: <ロックファイル更新が効かない / parent もアップデート不能 / first_patched_version なし>

**試行内容**:
- <試行 1>
- <試行 2>

**参考ドキュメント**: tmp/fix-dependabot-alert/alert-<番号>.md

**現在のブランチ**: `fix-dependabot-alert/<番号>`(変更なし)

**推奨される手動対応**:
- <例: parent パッケージの新リリースを待つ>
- <例: pnpm overrides で強制バージョン指定する>
- <例: alert を dismiss する(リスク許容できる場合)>
```

### state チェックで終了した場合(open でない)

```
## Dependabot Alert 対応中断

**Alert**: #<番号> <CVE / GHSA>
**state**: <現在の state>

Alert がすでに `<state>` 状態のため、対応はスキップしました。

**現在のブランチ**: <元のブランチ名>(変更なし)
```

### 未コミット変更で終了した場合

```
## Dependabot Alert 対応中断

作業ディレクトリに未コミット変更があるため、コマンドを終了しました。

**対応方法**:
1. 変更をコミットするか stash してから再実行してください
   - `git stash` または `git commit`
2. 再実行: `/fix-dependabot-alert <Dependabot alert URL>`

**現在のブランチ**: <元のブランチ名>
```
