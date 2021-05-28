# Controllers

## モジュール構成

実装方針が途中で変わったため、実装方針ごとにバージョンを設け、バージョンをモジュール名に含めて管理しています。

### V3::*

- ActiveRecordのインスタンスをビューに渡しています
- テンプレートエンジンに[Slim](https://github.com/slim-template/slim)を使用しています
- モバイル用のビューファイルがあります
  - レスポンシブ対応していない

### V4::*

- GraphQL APIを内部で使用しています
- [Entityクラス](../entities)のインスタンスをビューに渡しています
- テンプレートエンジンにERBを使用しています
- PCとモバイルどちらも同じビューファイルを参照します
  - レスポンシブ対応している

### V6::*

- ActiveRecordのインスタンスをビューに渡しています
- テンプレートエンジンにERBを使用しています
- PCとモバイルどちらも同じビューファイルを参照します
  - レスポンシブ対応している