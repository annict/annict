# タスク完了時のチェックリスト

## コード実装後に必ず実行すること

### 1. Lintとフォーマット
- `bin/standardrb` - Rubyコードのlintチェック
- `pnpm eslint . --fix` - JavaScriptコードのlintと自動修正
- `pnpm prettier . --write` - コードの自動フォーマット
- `bin/erb_lint --lint-all` - ERBファイルのlintチェック

### 2. 型チェック
- `bin/srb tc` - Sorbetによる型チェック
- `pnpm tsc` - TypeScriptの型チェック

### 3. テスト
- 新機能追加時: 対応するテストを作成
- 既存コード修正時: 関連するテストを実行
- `bin/rspec spec/requests/...` - 該当するテストを実行

### 4. 最終確認
- `bin/check` - 各種チェックをまとめて実行

## 注意事項
- コミット前に必ず上記のチェックを行う
- エラーや警告が出た場合は修正してから次に進む
- 型定義が不足している場合は `bin/rails sorbet:update` で更新
- CLAUDE.mdに記載されているコマンドがある場合は優先的に実行

## コミット時の注意
- ユーザーから明示的に要求された場合のみコミットを作成
- コミットメッセージは日本語でも英語でも可
- 変更内容を正確に表現するメッセージを作成