# Annict プロジェクト構造

## ディレクトリ構成

基本的にRailsプロジェクトの構成に従っていますが、一部独自のディレクトリがあります。

### app/ ディレクトリ

- `app/assets` - 画像やCSSファイル
- `app/components` - ViewComponentを使用したコンポーネント（再利用可能なUI要素）
- `app/controllers` - Railsのコントローラー（HTTPリクエスト処理）
- `app/forms` - フォームオブジェクト（フォーム関連のロジック）
- `app/javascript` - Hotwireで実装されたフロントエンド
- `app/jobs` - Active Job（非同期処理）
- `app/mailers` - Action Mailer（メール送信）
- `app/models` - PORO（Plain Old Ruby Object）や構造体など
- `app/policies` - 認可ルールが書かれたクラス
- `app/records` - ActiveRecord::Baseを継承したクラス（DBテーブルと1:1）
- `app/repositories` - RecordをModelに変換するクラス
- `app/services` - サービスクラス（ビジネスロジック）
- `app/validators` - カスタムバリデーション
- `app/views` - ViewComponentを使用したビュー

### その他の重要なディレクトリ

- `config/` - 設定ファイル
  - `config/locales/` - 国際化ファイル（日本語・英語）
- `db/` - データベース関連（マイグレーション等）
- `spec/` - RSpecテストファイル
  - `spec/requests/` - Request spec
- `docs/` - ドキュメント
  - `docs/claude/` - Claude向けのドキュメント

## 設計思想

- RecordとModelを分離（RecordはDB操作、Modelはドメインロジック）
- ViewComponentを使用したコンポーネント指向のUI設計
- RepositoryパターンでRecordとModelの変換を行う
- Serviceクラスでビジネスロジックをカプセル化
- 明確な依存関係の管理（各クラスが依存できるクラスを制限）
