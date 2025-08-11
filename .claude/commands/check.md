---
description: "変更点のチェック"
argument-hint: "ベースブランチ名 (default: main)"
---

以下のコマンドを実行し、異常終了したらエラー内容をもとに修正してください。

```bash
bin/rails zeitwerk:check
bin/rails sorbet:update
yarn prettier . --write
yarn eslint . --fix
yarn tsc
bin/erb_lint --lint-all
bin/standardrb
bin/srb tc
bin/rspec
```
