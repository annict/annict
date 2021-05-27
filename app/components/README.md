# Components

## モジュール構成

実装方針が途中で変わったため、実装方針ごとにバージョンを設け、バージョンをモジュール名に含めて管理しています。

### V4::Db::*

- Annict DBでのみ使用しているコンポーネントです
- [view_component](https://github.com/github/view_component) gemを使用しています
- ActiveRecordのインスタンスを渡しています

### V4::*

- システム全体で使用しているコンポーネントです
- `view_component` gemを使用しています
- [Entityクラス](../entities)のインスタンスを渡しています

### V6::*

- システム全体で使用しているコンポーネントです
- [htmlrb](https://github.com/kiraka/htmlrb) gemを使用しています
- ActiveRecordのインスタンスを渡しています
