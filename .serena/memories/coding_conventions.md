# Annict コーディング規約

## 全般
- 説明的な命名規則を採用
- コメントを日本語で適切に追加し、コードの可読性を高める
- セキュリティのベストプラクティスに従った実装
- コードベースの一貫性を保つ
- 1行の文字数は90文字を超えないようにし、超える場合は適切に改行

## Ruby
- 文字列はダブルクオートで記述
- 最終行に改行を入れる (RuboCop Layout/TrailingEmptyLines)
- 後置ifは使用禁止
- ハッシュのキーと変数名が同じ場合は省略記法を使用
- マジックコメント: `# typed: strict` (Sorbet), `# frozen_string_literal: true`
- プライベートメソッドは `private def` として定義
- `T.must` は使わず、`#not_nil!` を使用

## Rails
### クラスの依存関係
- Component: Component, Form, Model
- Controller: Form, Model, Record, Repository, Service, View
- Form: Record, Validator
- Job: Service
- Mailer: Model, Record, Repository, View
- Model: Model
- Policy: Record
- Record: Record
- Repository: Model, Record
- Service: Job, Mailer, Record
- Validator: Record
- View: Component, Form, Model

### マイグレーション
- idの生成には `generate_ulid()` 関数を使用

### I18n
- 翻訳ファイルは用途別に分類
  - forms.(ja,en).yml: フォーム関連
  - messages.(ja,en).yml: メッセージ・説明文
  - meta.(ja,en).yml: メタデータ
  - nouns.(ja,en).yml: 名詞・ラベル
- 日本語と英語の両方をサポート

## CSS
- Tailwind CSSを使用

## RSpec
- `context` ブロックは使用禁止、`it` の中にケースを記述
- `let`, `let!` は使用禁止、`it` 内に変数を定義
- `described_class` は使用禁止、明示的にクラス名を記述
- Request specのファイルパス: `spec/requests/<コントローラー名>/<アクション名>_spec.rb`