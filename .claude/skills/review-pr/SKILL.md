---
name: review-pr
description: GitHub の PR URL を受け取り、PR のチェックアウト・ベースブランチへの rebase・コードレビューを一括で行い、マージ可否を判断する。レビュー本体は `/review` スキルに委譲する。
argument-hint: "<PR URL>"
---

# PR レビュー <PR URL>

GitHub の PR URL を受け取り、以下を自動で行ってください:

1. PR 情報の取得
2. ローカルへのチェックアウト・最新 base への rebase
3. CI 状態の取得
4. **Dependabot PR の場合は、PR 本文の "compare view" リンクから依存ライブラリの実差分を取得する**(後述)
5. PR 情報・CI 状態・(Dependabot の場合は)実差分サマリーを一時ドキュメントにまとめて `/review` スキルに渡す
6. Dependabot かつ CI 失敗時は修正ループを実行してから `/review` を呼び出す
7. `/review` のレビュー結果に基づいてマージ可否を判断し報告する

## 引数

`$ARGUMENTS` は以下の形式で渡されます:

- 形式: `<PR URL>`
- 例: `/review-pr https://github.com/owner/repo/pull/123`

不正な URL が渡された場合はエラーを報告して終了してください。

## 原則

- **プロジェクト固有の情報を含めない**: このスキルは複数プロジェクトで共有されるため、プロジェクト固有の絶対パスや専用コマンドを埋め込まない。必要な情報は各プロジェクトの `CLAUDE.md` や `Makefile` から動的に判断する
- **自動マージは行わない**: `gh pr merge` の自動実行はしない。マージはユーザーが GitHub 上で手動で実施する
- **コマンド完了後は PR のブランチに留まる**: 元のブランチには戻らない。ユーザーが PR のブランチで継続作業できる状態で終了する

## 手順

### ステップ 1: 引数の解析

- `$ARGUMENTS` から PR URL を取得する
- `https://github.com/<owner>/<repo>/pull/<number>` の形式でない場合はエラーを報告して終了する

### ステップ 2: 作業ディレクトリのクリーン状態チェック

**重要**: GitHub への API 呼び出しを行う前に、ローカルの状態を確認する。

- `git status --porcelain` で未コミット変更がないか確認する
- 未コミット変更がある場合はエラーを報告して終了する
  - ユーザーに `git commit` または `git stash` を促すメッセージを表示する
  - GitHub への不要な API リクエストを避けるため、PR 情報の取得より前にローカルで弾く

### ステップ 3: PR 情報の取得

```sh
gh pr view <PR URL> --json number,title,author,body,headRefName,baseRefName,state,url,additions,deletions,changedFiles
```

取得する情報:

- `number`: PR 番号
- `title`: PR タイトル
- `author.login` / `author.is_bot`: 作成者の GitHub ユーザー名と Bot 判定
- `body`: PR 本文（Dependabot の場合は Changelog・リリースノート・CVE 情報などを含む）
- `headRefName` / `baseRefName`: ヘッド/ベースブランチ名
- `state`: PR の状態（OPEN/CLOSED/MERGED）
- `additions` / `deletions` / `changedFiles`: 変更規模

**state チェック**: `state` が `OPEN` でない場合は警告を表示して終了する。

### ステップ 4: PR のチェックアウト

```sh
gh pr checkout <PR URL>
```

`gh` が自動で PR のブランチに切り替え、リモートの追跡設定を行う。

### ステップ 5: ベースブランチへの rebase

PR ブランチを最新の base の上に乗せ、マージ後の状態でレビューを行えるようにする。

```sh
git fetch origin <baseRefName>
git rebase origin/<baseRefName>
```

**コンフリクトが発生した場合**:

- `git rebase --abort` で中断する
- 自動解決は試みず、ユーザーに手動での rebase を促して終了する
- 報告フォーマットは「rebase コンフリクトで終了した場合」を参照する

rebase 失敗時にここで中断することで、後続の API 呼び出しを無駄にしない。

### ステップ 6: CI 状態の取得

rebase が成功した後、CI ジョブの状態を取得する。rebase 失敗時の API 呼び出しを無駄にしないため、このステップは rebase 成功後に実行する。

```sh
gh pr checks <PR URL>
```

取得する情報:

- 各 CI ジョブの名前と状態（成功 / 失敗 / 実行中 / スキップ）
- 失敗しているジョブがあれば、そのジョブ名と `gh pr checks` の出力に含まれるログへのリンク

**CI がまだ実行中の場合**: ユーザーに状態を報告し、そのままレビューに進む（実行中のジョブは「未確定」として扱う）。CI 失敗判定のトリガーにはしない。

CI 状態は後続のフロー分岐（ステップ 8）と一時ドキュメント（ステップ 7）で使用する。

### ステップ 7: 一時ドキュメントの作成

`tmp/review-pr/pr-<番号>.md` に PR 情報と CI 状態をまとめた一時ドキュメントを作成する。

- `tmp/review-pr/` ディレクトリが存在しない場合は `mkdir -p` で作成する
- `<番号>` は PR 番号

#### Dependabot PR の場合の追加処理（必須）

**重要**: Dependabot PR の場合（ステップ 8 の判定と同じく `author.is_bot == true`）は、PR 本文の Release notes / Commits セクションが **truncated される** ことがあるため、PR 本文だけでは破壊的変更の有無を判断できない。**Dependabot PR を Approve するかどうかの判断には、依存ライブラリの実差分の確認が必須** である。

> パッチアップデートでも behavior change や Breaking change を取り込むライブラリは多い（Rails、Sentry、grpc、AWS SDK など）。SemVer の表面的な情報だけで「パッチだから安全」と判断してはいけない。

以下を一時ドキュメントに含めるために実施する:

1. **Compare view URL の抽出**:
   - PR 本文（`gh pr view --json body` の `body` フィールド）から `https://github.com/<owner>/<repo>/compare/<base>...<head>` 形式の URL を探す。Dependabot は本文中の "Commits" セクションで `compare view` リンクとして必ず提示する
   - URL から `<owner>`, `<repo>`, `<base_tag>`, `<head_tag>` を抽出する

2. **実差分の規模把握**:

   ```sh
   gh api repos/<owner>/<repo>/compare/<base_tag>...<head_tag> --jq '{total_commits: .total_commits, files_count: (.files | length)}'
   ```

   - 総コミット数とファイル数を取得し、規模感を把握する
   - パッチアップデートでも数百コミット含まれることがある（マイナーリリースを跨ぐ場合）

3. **CHANGELOG.md / CHANGES.md の差分取得**:

   ```sh
   gh api repos/<owner>/<repo>/compare/<base_tag>...<head_tag> \
     --jq '.files[] | select(.filename | test("(?i)CHANGELOG\\.md$|CHANGES\\.md$|HISTORY\\.md$")) | "=== \(.filename) ===\n\(.patch)\n"'
   ```

   - 多くのライブラリでは CHANGELOG ファイルにバージョンごとの変更点・CVE・breaking change が記載されている
   - 出力が大きい場合（数千行を超える）は `head` などで分割して取得する

4. **CHANGELOG から以下を抽出して一時ドキュメントに記載**:
   - **追加されている CVE 修正**: PR 本文の Release notes が truncated されている場合、本文に出ていない CVE がここで見つかることがある
   - **Breaking change と明記された変更**: 「BREAKING」「breaking change」「removed」「Note that this change breaks」などの文言を探す
   - **挙動が変わる修正**: 「Fix」「Change」のうち、ユーザーコードの動作に影響する可能性があるもの（例: `Fix XX to return ...` のような戻り値変更、SQL 生成パターンの変更、バリデーション条件の変更）
   - **新規依存・依存削除**: 推移的依存の追加・削除

5. **CHANGELOG が存在しない場合**: GitHub Releases（`gh api repos/<owner>/<repo>/releases`）や リリースタグの commit log を代替として確認する

6. **PR 本文に compare view URL がない場合**: PR 本文に十分な情報がないことを一時ドキュメントに明記し、コードベースで grep による広めの確認を `/review` 側に促す

ドキュメントの内容:

```markdown
# PR #<番号>: <タイトル>

## 基本情報

- **番号**: #<番号>
- **タイトル**: <タイトル>
- **作成者**: <author.login>（<Bot or Human>）
- **ベースブランチ**: <baseRefName>
- **ヘッドブランチ**: <headRefName>
- **変更規模**: <changedFiles> ファイル / +<additions> -<deletions>
- **URL**: <url>

## CI 状態

- **全体**: <成功 / 失敗 / 実行中>
- **ジョブごとの状態**:
  - `<ジョブ名 1>`: <状態>
  - `<ジョブ名 2>`: <状態>
  - ...

## PR 本文

<PR の body をそのまま転記>

## Dependabot 実差分サマリー（Dependabot PR の場合のみ）

- **Compare view URL**: <https://github.com/...compare/...>
- **実差分の規模**: <N commits / N files>
- **間に含まれるリリース**: <例: 8.0.3, 8.0.4, 8.0.4.1>

### CHANGELOG から抽出した CVE 修正

| CVE            | 内容         | 出典 component                   |
| -------------- | ------------ | -------------------------------- |
| CVE-XXXX-XXXXX | <内容の要約> | <component 名（例: actionview）> |
| ...            | ...          | ...                              |

### CHANGELOG から抽出した Breaking change / 挙動変更

| 変更         | 出典 component / バージョン | 影響を受けるパターン（grep キーワード等） |
| ------------ | --------------------------- | ----------------------------------------- |
| <変更の要約> | <component@version>         | <検索キーワード>                          |
| ...          | ...                         | ...                                       |

### 推移的依存の追加・削除・メジャーアップ

- <依存名> <旧バージョン> → <新バージョン>（<patch / minor / **major**>）— <備考>
- ...
```

この一時ドキュメントを `/review` スキルに参考ドキュメントとして渡すことで、レビュー実行時に PR 本文（Dependabot の Changelog・CVE 情報など）と CI 状態、そして **依存ライブラリの実差分** を考慮したレビューが可能になる。`/review` 側ではこの「Breaking change / 挙動変更」テーブルの各項目について、コードベースで `Grep` による影響範囲確認を行う。

### ステップ 8: フロー分岐

PR の作成者と CI 状態に応じて処理を分岐する:

| 条件                                                                     | 次のステップ                                             |
| ------------------------------------------------------------------------ | -------------------------------------------------------- |
| 作成者が Dependabot（`author.is_bot == true`）かつ CI に失敗ジョブがある | ステップ 9（修正ループ）を実行してからステップ 10 へ進む |
| それ以外                                                                 | 直接ステップ 10（`/review` 呼び出し）へ進む              |

**判定ロジック**:

- Dependabot 判定: `author.is_bot == true` を主判定とする。`author.login` が `dependabot[bot]` などの場合も該当するが、`is_bot` を優先する
- CI 失敗判定: `gh pr checks` で取得した状態に `fail` / `failure` のジョブが 1 つでも含まれる場合

### ステップ 9: 修正ループ（Dependabot かつ CI 失敗時のみ）

修正ループの目的: Dependabot PR が CI 失敗を起こした場合、機械的に修正可能な問題（フォーマット崩れ・lint 警告・API リネームに伴うビルドエラー・機械的なテスト失敗など）をローカルで自動修正し、CI が通る状態にしてからレビューを行う。

修正ループは rebase でベースを取り込んだ状態で実行するため、古いベースで通っていた CI が新しいベースで通らなくなったケースもここで検出・修正される。

#### 失敗ジョブの特定

1. ステップ 6 で取得した `gh pr checks` の結果から失敗しているチェックと run ID を特定する
2. `gh run view <run-id> --log-failed` で失敗ジョブのログを取得する
3. ジョブ名と失敗内容から、ローカルで再現するコマンドを判定する
   - **実行するコマンドは各プロジェクトの `CLAUDE.md`（「コミット前に実行するコマンド」セクション）や `Makefile` から動的に判断する**
   - プロジェクト固有の絶対パスや専用コマンドをスキル内にハードコードしない
   - 一般的な対応例（プロジェクトによって実際のコマンドは異なる）:
     - lint ジョブが失敗 → `make lint` など
     - test ジョブが失敗 → `make test` など
     - フォーマット関連のジョブが失敗 → `make fmt` など
     - ビルドエラー → `go build ./...` / `tsc` など

#### 自動修正の対象範囲

以下は一般的なパターンの例示。プロジェクトによって実際のコマンドは異なるため、`CLAUDE.md` や `Makefile` から適切なコマンドを判断する。

| 種類                                       | 自動修正 | 想定コマンド例                |
| ------------------------------------------ | -------- | ----------------------------- |
| フォーマット崩れ(gofmt, prettier, rubocop) | する     | `make fmt` など               |
| Lint 警告(linter 全般)                     | する     | `make lint` など              |
| ビルドエラー(API シグネチャ変更程度)       | する     | `go build ./...` / `tsc` など |
| テスト失敗(API リネームなど機械的な原因)   | する     | `make test` など              |
| 大規模リファクタリング                     | しない   | -                             |
| 仕様変更を伴う修正                         | しない   | -                             |
| テスト失敗が示す本質的なバグ               | しない   | -                             |

判断が難しい場合は修正を諦め、修正ループを中断してレビューに進む。最終的なマージ判断を Request Changes とし、ユーザーに委ねる。

#### 修正ループの手順

各試行で以下を順に実行する:

1. 失敗しているジョブを特定する
2. 失敗内容を分類する（フォーマット / lint / ビルド / テスト / その他）
3. 「自動修正の対象範囲」を参照して、自動修正すべきかどうか判断する
4. ローカルで該当する CI コマンドを実行して失敗を再現する
5. 自動修正を試行する（対象範囲内の場合のみ）
6. ローカルで CI 相当のコマンドが通ることを再度確認する
7. 修正内容ごとに分けてコミットする（push はしない）
8. 全ジョブが通過したか確認する

#### 試行回数の上限と中断条件

- 修正ループの試行回数は **最大 3 回**
- 以下のいずれかで修正ループを終了する:
  - **成功で終了**: 全ジョブが通過した
  - **上限到達で終了**: 3 回試行しても CI が通らなかった
  - **対象外で終了**: 失敗内容が自動修正の対象範囲外と判断した
- いずれの終了条件でもレビュー（ステップ 10）には進む。マージ可否判断はレビュー結果と合わせて決定する

#### 修正コミットの粒度

修正内容ごとにコミットを分ける。同じ種類の修正はまとめてよい。

- 例 1: `フォーマット修正`
- 例 2: `golangci-lint の警告を修正`
- 例 3: `API リネームに追従`

**コミットルール**:

- コミットメッセージは日本語で、何を修正したかを簡潔に記述する
- `--no-verify` は使わない。pre-commit hook が失敗したら根本原因を修正する
- **push はしない**。push はユーザーが後で手動で実施する

#### 一時ドキュメントへの追記

修正ループが終了したら（成功 / 上限到達 / 対象外のいずれでも）、`tmp/review-pr/pr-<番号>.md` に以下を追記する:

```markdown
## 修正ループ結果

- **試行回数**: <N> / 3 回
- **結果**: <成功 / 上限到達 / 対象外で終了>
- **追加コミット**:
  - <コミット1のメッセージ>
  - <コミット2のメッセージ>
  - ...
- **残っている失敗**（あれば）:
  - `<ジョブ名>`: <失敗内容の要約>
```

追記後、ステップ 10（`/review` 呼び出し）に進む。

### ステップ 10: `/review` スキルの呼び出し

Skill ツール経由で `/review` を呼び出し、以下の引数を渡す:

- 比較対象ブランチ名: `<baseRefName>`
- 参考ドキュメントのパス: ステップ 7 で作成し（修正ループが走った場合はステップ 9 で追記した）`tmp/review-pr/pr-<番号>.md`

引数形式: `<baseRefName> tmp/review-pr/pr-<番号>.md`

`/review` スキルがローカルブランチ（チェックアウト済みの PR ブランチ）を対象にレビューを実行し、`docs/private/reviews/` 配下にレビュードキュメントを作成する。

### ステップ 11: レビュー結果の受け取り

`/review` スキルから以下を受け取る:

- レビュードキュメントのパス: `docs/private/reviews/<ファイル名>.md`
- 評価: Approve / Request Changes / Comment
- 主要な問題点のサマリー

### ステップ 12: マージ可否の判断と報告

「マージ可否判断の基準」に従ってマージ可否を決定し、「報告フォーマット」に従って結果を報告する。

## マージ可否判断の基準

| 状況                                                                    | 判断            |
| ----------------------------------------------------------------------- | --------------- |
| 依存関係のパッチ/マイナーバージョンアップ、破壊的変更なし、CI 成功      | Approve         |
| Dependabot のセキュリティ更新（CVE 修正）、破壊的変更なし               | Approve         |
| Dependabot で CI 失敗 → 修正ループで CI 通過に成功                      | Approve         |
| Dependabot で CI 失敗 → 修正ループで CI 通過に失敗（上限到達 / 対象外） | Request Changes |
| メジャーバージョンアップで破壊的変更を含む                              | Comment         |
| `/review` が Request Changes を出した（ガイドライン違反等）             | Request Changes |
| 判断材料が不足している、または専門的判断が必要な変更                    | Comment         |

**「破壊的変更なし」の判定基準**: PR 本文の Release notes だけでは不十分。**ステップ 7 で取得した「Dependabot 実差分サマリー」（Compare view から取得した CHANGELOG / CVE / Breaking change）** と、`/review` 側で行ったコードベースへの影響範囲調査（Grep）の両方を根拠とする。Compare view を確認せずに「パッチアップデートだから安全」と判断してはいけない。

**判断の補助情報**: Dependabot PR の場合は、PR 本文に含まれるリリースノート・Changelog・CVE 情報に加えて、ステップ 7 で取得した実差分サマリーを `/review` の結果と突き合わせて判断する。実差分サマリーが取得できなかった場合（PR 本文に compare view URL がない、CHANGELOG が存在しないなど）は判断材料が不足しているとみなし、Comment で報告してユーザーに委ねる。

## 報告フォーマット

### 正常にレビューが完了した場合

```
## PR レビュー完了

**PR**: #<番号> <タイトル>
**作成者**: <author>（<Bot or Human>）
**ブランチ**: <headRefName> ← <baseRefName>
**変更**: <changedFiles> ファイル / +<additions> -<deletions>
**CI 状態**: <成功 / 失敗（失敗ジョブ名） / 実行中>

**評価**: [Approve / Request Changes / Comment]
**判断理由**: <理由>

**レビュードキュメント**: docs/private/reviews/<ファイル名>.md
**参考ドキュメント**: tmp/review-pr/pr-<番号>.md

**主要な確認ポイント**:
- ...

**次のアクション**:
- マージする場合は GitHub 上で実施してください
- 現在のブランチ: <headRefName>（PR のブランチにいます）
```

### 修正コミットありの場合（Dependabot の修正ループが走った場合）

```
## PR レビュー完了（修正コミットあり）

**PR**: #<番号> <タイトル>
**作成者**: <author>
**ブランチ**: <headRefName> ← <baseRefName>
**CI 状態（修正前）**: 失敗（<失敗ジョブ名>）

**修正ループ**: <試行回数> / 3 回（<成功 / 上限到達 / 対象外で終了>）
**追加コミット**: <コミット数> 件
- <コミット1のメッセージ>
- <コミット2のメッセージ>
- ...

**評価**: [Approve / Request Changes / Comment]
**判断理由**: <理由>

**レビュードキュメント**: docs/private/reviews/<ファイル名>.md
**参考ドキュメント**: tmp/review-pr/pr-<番号>.md（PR 情報 + 修正ループ結果）

**次のアクション**:
1. 現在のブランチ: <headRefName>（PR のブランチにいます）
2. コミット内容を確認してください: `git log <baseRefName>..HEAD`
3. 問題なければ push してください: `git push --force-with-lease`
   - rebase によりベースが書き換わっているため force push が必要です
4. push 後、CI が通ることを GitHub 上で確認してください
5. CI 通過後、マージしてください
```

### rebase コンフリクトで終了した場合

```
## PR レビュー中断（rebase コンフリクト）

**PR**: #<番号> <タイトル>
**ブランチ**: <headRefName> ← <baseRefName>

ベースブランチを取り込む rebase 中にコンフリクトが発生したため、コマンドを終了しました。

**対応方法**:
1. 手動でコンフリクトを解決してから再実行してください
2. または、Dependabot の PR の場合は GitHub 上で `@dependabot rebase` をコメントし、Dependabot 側で rebase が完了してから再実行してください

**現在のブランチ**: <headRefName>（rebase は中断済み）
```
