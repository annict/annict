# Annict 開発用コマンドリファレンス

## テスト実行

- `bin/rspec` - テストを実行する
- `bin/rspec spec/requests/...` - 特定のテストファイルを実行

## 型チェック

- `bin/rails sorbet:update` - Sorbetの型定義を更新
- `bin/srb tc` - Sorbetの型チェックを実行
- `pnpm tsc` - TypeScriptの型をチェック

## コードフォーマット・Lint

- `pnpm prettier . --write` - Prettierで整形
- `pnpm eslint . --fix` - ESLintを実行
- `bin/erb_lint --lint-all` - ERB Lintを実行
- `bin/standardrb` - Ruby用のStandardを実行

## 統合チェック

- `bin/check` - 各種チェックをまとめて実行

## ビルド関連

- `pnpm build` - JavaScriptをビルド
- `pnpm build:css` - CSSをビルド

## Git操作

- `git status` - 変更状況を確認
- `git diff` - 変更内容を確認
- `git add .` - 変更をステージング
- `git commit -m "メッセージ"` - コミット作成

## システムコマンド (Darwin/macOS)

- `ls` - ファイル一覧表示
- `cd` - ディレクトリ移動
- `grep` - テキスト検索
- `find` - ファイル検索

## タスク完了時の推奨コマンド

1. `bin/standardrb` - Rubyコードのlint
2. `pnpm eslint . --fix` - JavaScriptコードのlint
3. `pnpm prettier . --write` - コード整形
4. `bin/srb tc` - 型チェック
5. `bin/rspec` - テスト実行（該当する場合）
6. `bin/check` - 統合チェック（最終確認）
