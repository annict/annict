# `app/components/`

- コンポーネントが管理されているディレクトリです
- 下記の実装方針に従っていないコードは `deprecated/` ディレクトリに格納されています

## 実装方針

- `ApplicationComponent` を継承すること
  - [ViewComponent](https://github.com/ViewComponent/view_component) を使用しています
- 以下のどちらかがレンダリングされること
  - ERBで記述されたテンプレート
  - ヘルパー
    - 例: [MalLinkComponent](https://github.com/annict/annict/blob/8f2d4374fc4ea1b378edd7ca3ef70d4f94e9822f/app/components/mal_link_component.rb) は `link_to` を返している
