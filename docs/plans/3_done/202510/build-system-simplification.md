# ビルドシステムの簡素化（Vite → Tailwind CLI + esbuild）設計書

## 概要

現在 Vite を使用して CSS（Tailwind CSS）と JS のビルドを行っていますが、システムの複雑さを軽減するため、より軽量でシンプルな構成に移行します。

**目的**:

- ビルドシステムの複雑さを軽減し、理解しやすく保守しやすい構成にする
- 設定ファイルと依存関係を最小限に抑える
- ビルドプロセスを明確にし、何が行われているかを見えやすくする
- 将来のメンテナンス負担を軽減する

**背景**:

- Vite は優れたツールだが、このプロジェクトの要件（HMR 不要、シンプルな CSS/JS ビルド）に対して大仰
- 設定ファイル（vite.config.ts, postcss.config.js）が多く、依存パッケージも多い
- 開発サーバーが不要（Go の Air だけで十分）
- Datastar を使った JS は import に対応していれば十分（TypeScript 不要）

## 要件

### 機能要件

- Tailwind CSS v4 を使用した CSS のビルド
- JS のバンドル（複数ファイルの import に対応）
- ファイル監視モード（watch）での自動再ビルド
- 本番ビルド時の minify
- basecoat-css などの npm パッケージの import
- CDN キャッシュ対策（クエリパラメータ方式）

### 非機能要件

#### シンプルさ

- **設定ファイルの最小化**: tailwind.config.js のみ（vite.config.ts, postcss.config.js を削除）
- **依存関係の最小化**: Tailwind CLI と esbuild のみ
- **ビルドプロセスの透明性**: コマンドが明確に見える

#### 開発者体験

- `pnpm watch` で CSS/JS を監視して自動再ビルド
- Air が static/ ディレクトリの変更を検知してページをリロード
- ビルドエラーが明確にわかる

## 設計

### 技術スタック

- **CSS ビルド**: Tailwind CLI (v4)
- **JS ビルド**: esbuild
- **パッケージマネージャー**: pnpm

### 移行前後の比較

| 項目                   | 移行前（Vite）                                                | 移行後（Tailwind CLI + esbuild）         |
| ---------------------- | ------------------------------------------------------------- | ---------------------------------------- |
| 設定ファイル           | vite.config.ts, postcss.config.js, tailwind.config.js         | tailwind.config.js のみ                  |
| 依存パッケージ         | vite, @vitejs/\*, postcss, @tailwindcss/postcss, autoprefixer | tailwindcss, esbuild のみ                |
| CSS ビルド             | Vite + PostCSS + Tailwind                                     | Tailwind CLI                             |
| JS ビルド              | Vite + esbuild（内部）                                        | esbuild（直接）                          |
| TypeScript             | main.ts                                                       | main.js（JavaScript のみ）               |
| 開発サーバー           | Vite dev server (port 5173)                                   | 不要（Air のみ）                         |
| ビルドコマンドの可視性 | `vite build`（内部処理が見えにくい）                          | `tailwindcss ...`, `esbuild ...`（明確） |
| ファイル監視           | `vite`                                                        | `pnpm watch`（tailwindcss + esbuild）    |
| 出力ディレクトリ       | web/dist/                                                     | static/css/, static/js/                  |
| CDN キャッシュ対策     | なし                                                          | クエリパラメータ（?v=commit-hash）       |

### ディレクトリ構成

```
web/
├── style.css         # Tailwind のソース（@import "tailwindcss" など）
└── main.js           # JS のエントリポイント（basecoat-css などを import）

static/
├── css/
│   └── style.css     # ビルド済み CSS（Tailwind CLI が生成）
└── js/
    └── main.js       # バンドル済み JS（esbuild が生成）

tailwind.config.js    # Tailwind 設定（変更なし）
package.json          # ビルドスクリプトを更新
```

### ビルドスクリプト

**package.json**:

```json
{
  "scripts": {
    "build:css": "tailwindcss -i web/style.css -o static/css/style.css --minify",
    "build:js": "esbuild web/main.js --bundle --outfile=static/js/main.js --minify",
    "build": "npm run build:css && npm run build:js",
    "watch:css": "tailwindcss -i web/style.css -o static/css/style.css --watch",
    "watch:js": "esbuild web/main.js --bundle --outfile=static/js/main.js --watch",
    "watch": "npm run watch:css & npm run watch:js"
  },
  "devDependencies": {
    "tailwindcss": "^4.1.13",
    "esbuild": "^0.24.2",
    "basecoat-css": "^0.3.2"
  }
}
```

### CDN キャッシュ対策（クエリパラメータ方式）

**仕組み**:

- ファイル名は固定（`style.css`, `main.js`）
- URL にバージョン番号を付与：`style.css?v=abc123`
- バージョンは Git コミットハッシュを使用

**実装**:

1. **Config 構造体に AssetVersion フィールドを追加**:

```go
// internal/config/config.go
type Config struct {
    // ...
    AssetVersion string // Gitコミットハッシュ
}

// 起動時にGitコミットハッシュを取得
func Load() (*Config, error) {
    // ...
    cfg.AssetVersion = getGitCommitHash()
    return cfg, nil
}

func getGitCommitHash() string {
    cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
    out, err := cmd.Output()
    if err != nil {
        return "dev" // Gitがない場合のフォールバック
    }
    return strings.TrimSpace(string(out))
}
```

2. **テンプレートヘルパー関数**:

```go
// テンプレートデータに AssetVersion を含める
data := map[string]interface{}{
    "AssetVersion": cfg.AssetVersion,
    // ...
}
```

3. **テンプレートでの使用**:

```html
<link rel="stylesheet" href="/static/css/style.css?v={{.AssetVersion}}" />
<script src="/static/js/main.js?v={{.AssetVersion}}"></script>
```

### 静的ファイル配信

**Go ハンドラー**:

```go
// cmd/server/main.go
r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
```

### テスト戦略

- ビルドスクリプトが正しく動作することを確認
- CSS が正しくビルドされることを確認（Tailwind のクラスが展開されているか）
- JS が正しくバンドルされることを確認（import が解決されているか）
- watch モードが動作することを確認
- テンプレートで CSS/JS が正しく読み込まれることを確認
- AssetVersion がテンプレートに正しく渡されることを確認

## タスクリスト

### フェーズ 1: 依存関係とビルドスクリプトの更新

- [x] 依存関係の整理
  - Vite 関連パッケージの削除（vite, @vitejs/plugin-legacy, typescript）
  - PostCSS 関連パッケージの削除（postcss, @tailwindcss/postcss, autoprefixer）
  - esbuild の追加
  - package.json の更新
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

- [x] ビルドスクリプトの更新
  - package.json scripts を Tailwind CLI + esbuild に変更
  - build:css（Tailwind CLI）、build:js（esbuild）、build スクリプトの追加
  - watch:css、watch:js、watch スクリプトの追加
  - Makefile の更新（必要に応じて）
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 40 行（実装 40 行 + テスト 0 行）

### フェーズ 2: 設定ファイルとソースファイルの整理

- [x] 設定ファイルの整理
  - vite.config.ts の削除
  - postcss.config.js の削除
  - tailwind.config.js は保持（変更なし）
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 0 行（削除のみ）

- [x] ソースファイルとディレクトリ構成の変更
  - web/main.ts → web/main.js に変更（TypeScript から JavaScript へ）
  - basecoat-css の import 方法を調整
  - static/css/, static/js/ ディレクトリの作成
  - .gitignore の更新（static/css/\*, static/js/\*を追加）
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 20 行（実装 20 行 + テスト 0 行）

### フェーズ 3: テンプレートの更新

- [x] テンプレートの更新
  - CSS/JS 読み込みパスを /static/css/, /static/js/ に変更
  - 全 HTML テンプレートの確認と更新
  - レイアウトテンプレート（base.html など）の更新
  - **想定ファイル数**: 約 5 ファイル（実装 5 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

### フェーズ 4: CDN キャッシュ対策

- [x] CDN キャッシュ対策（クエリパラメータ方式）
  - Git コミットハッシュを取得する関数の実装（getGitCommitHash）
  - Config 構造体に AssetVersion フィールドを追加
  - 起動時に AssetVersion を設定
  - テンプレートデータに AssetVersion を含める
  - 全 HTML テンプレートで CSS/JS 読み込みにバージョンパラメータ（?v=xxx）を追加
  - 開発環境と本番環境での動作確認
  - **想定ファイル数**: 約 5 ファイル（実装 4 + テスト 1）
  - **想定行数**: 約 80 行（実装 60 行 + テスト 20 行）

### フェーズ 5: 動作確認とドキュメント更新

- [x] 動作確認とドキュメント更新
  - pnpm watch での自動ビルド確認
  - Air との連携確認（ファイル監視）
  - CSS が正しくビルドされることを確認
  - JS が正しくバンドルされることを確認
  - ブラウザでページが正しく表示されることを確認
  - AssetVersion がクエリパラメータに反映されることを確認
  - CLAUDE.md の技術スタック更新
  - 開発フローのドキュメント更新
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 20 行 + テスト 10 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **ハッシュ付きファイル名（manifest 方式）**: クエリパラメータ方式で十分なため
- **TypeScript サポート**: JavaScript のみで十分（Datastar は JS で使用）
- **HMR（Hot Module Replacement）**: Air のリロードで十分
- **CSS Modules**: Tailwind CSS を使用するため不要
- **画像最適化**: 既存の imgproxy を使用

## 参考資料

- [Tailwind CSS CLI Documentation](https://tailwindcss.com/docs/installation)
- [esbuild Documentation](https://esbuild.github.io/)
- [Datastar Documentation](https://data-star.dev/)
