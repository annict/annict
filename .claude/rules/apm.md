---
paths:
  - "**"
---

# APM (Agent Package Manager) ガイドライン

Korylus のプロダクト (Annict, Mewst, Wikino) では、AI 向けガイドラインとスキルを共有するために [APM (Agent Package Manager)](https://github.com/microsoft/apm) を利用しています。このドキュメントは、APM の基本的な使い方と運用ルールをまとめたものです。

## パッケージの構成

```
/korylus-guidelines/                   # APM パッケージのソース (Single Source of Truth)
├── apm.yml                            # パッケージ定義 (name: korylus)
└── .apm/
    ├── instructions/                  # .claude/rules/ に配置される
    │   ├── common.instructions.md
    │   ├── apm.instructions.md        # このファイル
    │   ├── go-*.instructions.md
    │   └── rails-*.instructions.md
    └── skills/                        # .claude/skills/ に配置される
        ├── commit/
        ├── review/
        ├── review-pr/
        └── ...

/workspace/                            # 各プロダクト (例: Wikino)
├── apm.yml                            # 依存定義 (../korylus-guidelines)
├── apm.lock.yaml                      # ロックファイル (自動生成)
├── apm_modules/                       # APM の作業ディレクトリ
└── .claude/
    ├── rules/                         # apm install で生成
    └── skills/                        # apm install で生成
```

## 主要コマンド

APM はプロダクトのルート (`apm.yml` があるディレクトリ) から実行します。

| コマンド                | 用途                                                                                 |
| ----------------------- | ------------------------------------------------------------------------------------ |
| `apm install`           | `apm.yml` の依存関係を取得し、`.claude/rules/` と `.claude/skills/` にファイルを配置 |
| `apm install --update`  | 依存パッケージを最新の Git 参照に更新                                                |
| `apm install --dry-run` | 実際に配置せず、何が入るかだけ確認                                                   |
| `apm audit`             | 配置済みファイルに含まれる隠し Unicode 文字をスキャン                                |
| `apm audit --strip`     | 不要な隠し Unicode 文字を削除 (絵文字と空白は保持)                                   |
| `apm prune`             | `apm.yml` に記載がなくなったパッケージを削除                                         |
| `apm uninstall <pkg>`   | パッケージと配置済みファイルを削除                                                   |
| `apm list`              | 利用可能なスクリプト一覧を表示                                                       |

## ファイルの編集可否ルール

- **編集不可**: `apm.lock.yaml` の `deployed_files` に列挙されたファイル。`apm install` で `korylus-guidelines` から上書きされるため、直接編集しても次回 `apm install` で元に戻る
- **編集可**: `deployed_files` に載っていないファイル。プロダクト固有のルールやスキルを `.claude/rules/` や `.claude/skills/` に追加しても良い
- **名前衝突は運用で回避**: プロダクト固有のファイル名を決めるときは、先に `apm.lock.yaml` の `deployed_files` を確認して重複を避ける

共通ガイドラインを変更したい場合は `korylus-guidelines` 側のソース (`.apm/instructions/` や `.apm/skills/`) を編集し、各プロダクトで `apm install` を再実行してください (単方向フロー)。

## フォーマットの運用

`.md` / `.yml` のフォーマットは **`korylus-guidelines` 側で一度かけるだけで全プロダクトに反映される**、という単方向フローで運用します。

### `korylus-guidelines` 側

- `package.json` + `.oxfmtrc.json` で Oxfmt を導入済み
- `pnpm install` 後、`pnpm fmt` で整形、`pnpm fmt:check` で確認
- `mise.toml` で Node.js と pnpm のバージョンを固定しているので、Annict / Wikino などの Docker コンテナからでも `mise install` すれば同じ環境で作業できる

### プロダクト側 (例: Wikino)

- `.oxfmtrc.json` の `ignorePatterns` に `apm.lock.yaml` と `apm_modules/` を追加することで、APM 管理下のファイルを触らないようにする
- それ以外の APM 生成物 (`.claude/rules/`, `.claude/skills/`) は `korylus-guidelines` 側で事前に整形されているため、プロダクト側で `pnpm fmt` を走らせても差分は発生しない (はず)

もし `apm install` 直後にプロダクト側の `pnpm fmt` で差分が出た場合は、`korylus-guidelines` 側のフォーマットが古いサインです。`korylus-guidelines` で `pnpm fmt` → コミット → プロダクトで `apm install` やり直し、で解消してください。

## Markdown の HTML コード例の注意

Oxfmt (Prettier 系フォーマッタ) は Markdown の `html` コードブロック内で閉じタグのない HTML を見つけると、後続の要素をネスト扱いにして自動インデントします。結果として、本来 **独立した複数のスニペット** が **1 つのネスト構造** にまとめられ、セマンティクスが壊れます。

たとえば、空行で区切った `<div hx-get="/a">` / `<div hx-get="/b">` / `<input hx-get="/c">` の 3 スニペットは、そのままだと `/a` の `<div>` の子として `/b` の `<div>` が入り、さらにその子として `/c` の `<input>` が入る 1 つのネストされた構造に変換されてしまいます。

そのため、`.instructions.md` や `SKILL.md` に HTML コードを載せる際は、**独立したスニペットには必ず閉じタグを付ける** こと:

```html
<!-- 例 A -->
<div hx-get="/a"></div>

<!-- 例 B -->
<div hx-get="/b"></div>

<!-- 例 C -->
<input hx-get="/c" />
```

こう書くと、フォーマッタは各要素が独立していると認識してネストを試みません。

## 診断メッセージの読み方

`apm install` の末尾に `-- Diagnostics --` セクションが出ることがあります。代表的なパターンと対応は以下の通り。

### `[!] N file(s) contain hidden characters`

多くの場合は **`U+FE0F` (Emoji Presentation Selector / Variation Selector-16)** が検出されています。`⚠️` や `⭕️` のような絵文字の直後に付く zero-width 文字で、絵文字としての表示を明示するために必要な Unicode コードポイントです。

- **対応**: **対応不要**。`apm audit -v` で詳細を見ると Severity が `INFO` になっており、`apm audit --strip` も「strippable character ではない」として削除を拒否します。警告っぽい `[!]` 表示ですが、実体は無害な INFO-level の通知です
- **消したい場合**: 絵文字自体を削るか、VS16 だけ剥がす (後者はプラットフォームによって絵文字が色なしのテキスト表示になる可能性がある) の 2 択。通常は現状維持を推奨

### その他の警告

ERROR / WARNING レベルの診断が出た場合は `apm audit -v` で詳細を確認してから対処してください。

## トラブルシューティング

### `apm install` してもファイルが更新されない

- `apm.lock.yaml` が古い可能性。`apm install --update` で最新に更新する
- `korylus-guidelines` 側のファイルがコミットされていない可能性。ローカルパス依存 (`../korylus-guidelines`) でもワーキングツリーの状態が反映されるので、念のため `cd /korylus-guidelines && git status` で未コミット差分を確認する

### プロダクト側で `pnpm fmt` すると APM 管理下のファイルに差分が出る

- プロダクトの `.oxfmtrc.json` に `apm.lock.yaml` と `apm_modules/` を除外追加しているか確認
- それ以外のファイル (`.claude/rules/*`, `.claude/skills/*/SKILL.md`) に差分が出る場合は、`korylus-guidelines` 側で `pnpm fmt` がかかっていないサイン。上流を直す

### `mise` が `mise.toml` を信頼していないと言われる

- 初回のみ `mise trust /korylus-guidelines/mise.toml` を実行する

## 参考資料

- [APM (Agent Package Manager) README](https://github.com/microsoft/apm)
- `/korylus-guidelines/apm.yml` — APM パッケージ定義
- 各プロダクトの `apm.yml` — 依存関係定義
- 各プロダクトの `apm.lock.yaml` — ロックファイル (編集不可ファイルの判定基準)
