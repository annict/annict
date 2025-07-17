---
description: "Request specの修正をします"
---

- `$ARGUMENTS` のrequest specを書いてください
`$ARGUMENTS` のテストファイルを見つけて現在の実装を確認してください
- テストファイル内のトップレベルの `describe` は、 `RSpec.describe "<エンドポイント (例: GET /)>", type: :request do` で始まるようにしてください
- `docs/claude/base/coding-conventions/rspec.md` に沿っていない書き方をしているテストがあるので、その場合は沿うように修正してください
- その他 `docs/claude/base/coding-conventions/` 配下のファイルに記載のコーディング規約も守ってください
- 必要に応じて不足しているテストパターンを追加してください
  - 追加したテストが失敗したときは実装を変更せず、テストだけを修正してください
- 既存のテストでチェックしている内容は消さないでください
- 修正後、修正したテストを実行しテストが通ることを確認してください
- テストの説明を日本語に変更してください
- 修正後、`bin/standardrb` が成功することを確認してください
