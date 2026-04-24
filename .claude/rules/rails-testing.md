---
paths:
  - "rails/**/*.{rb,erb}"
---

# テストガイド

このドキュメントは、Rails 版 Wikino でのテストのコーディング規約とベストプラクティスを説明します。

## RSpec

### 基本方針

- `context`, `let`, `described_class` は使用しない
- `it` ブロック内で変数を定義する
- FactoryBot で作成したレコードの変数名には `_record` サフィックスを付ける

```ruby
# ❌ context, let, described_classは使用しない
context "when xxx" do
  let(:user) { create(:user) }
end

# ✅ itブロック内で変数定義
it "xxxのとき、somethingすること" do
  user = FactoryBot.create(:user)
  # テスト実装
end

# ✅ FactoryBotで作成したレコードの変数名には_recordサフィックスを付ける
user_record = FactoryBot.create(:user_record)
space_record = FactoryBot.create(:space_record)
space_member_record = FactoryBot.create(:space_member_record, user_record:, space_record:)

# ❌ サフィックスなしの変数名は避ける
user = FactoryBot.create(:user_record)
space = FactoryBot.create(:space_record)
```

### システムテストの待機処理

```ruby
# ❌ sleepを使用した待機処理は避ける
button.click
sleep 2
expect(page).to have_current_path(some_path)

# ✅ Capybaraの待機機能を活用
button.click
# ページ上の要素の変化を待つ（Capybaraが自動的に最大5秒待機）
expect(page).not_to have_content("削除されたコンテンツ")
expect(page).to have_content("新しく表示されるコンテンツ")

# ✅ have_css/not_to have_cssで要素の出現/消失を待つ
expect(page).to have_css(".success-message")
expect(page).not_to have_css(".loading-spinner")
```

**重要**: システムテストでは`sleep`の使用を避け、Capybara の自動待機機能を活用すること。要素の出現や消失、コンテンツの変化を検証することで、適切な待機処理が自動的に行われる。
