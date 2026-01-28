# Canary GraphQL API 削除 設計書

## 実装ガイドラインの参照

### Rails版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@rails/CLAUDE.md](/workspace/rails/CLAUDE.md) - 全体的なコーディング規約
- [@rails/docs/architecture-guide.md](/workspace/rails/docs/architecture-guide.md) - アーキテクチャガイド
- [@rails/docs/testing-guide.md](/workspace/rails/docs/testing-guide.md) - テスト戦略ガイド
- [@rails/docs/security-guide.md](/workspace/rails/docs/security-guide.md) - セキュリティガイドライン

## 概要

Canary GraphQL API (`Canary::AnnictSchema`) を削除する。

**目的**:

- 使用されていない API エンドポイントを削除し、コードベースを整理する

**背景**:

- Canary API は本番環境でエンドポイントとして公開されているが、実際のリクエストはない
- Beta API (`Beta::AnnictSchema`) は外部から利用されており、引き続き維持する
- Canary の設計思想は REST API V2 の参考として別途ドキュメント化した

## 要件

### 機能要件

- Canary GraphQL API のエンドポイント (`api.annict.com/canary/graphql`) を削除する
- Canary GraphQL API 関連のすべてのコードを削除する
- ローカル開発用 API (`/api/local/graphql`) も Canary を使用しているため削除する
- `GraphqlResolvable` concern を削除する（`#global_id` メソッドは実際には使用されていないため）
- `Episode` モデルから `include GraphqlResolvable` を削除する
- 関連するテストコードを削除する

### 非機能要件

- **後方互換性**: Beta API (`api.annict.com/graphql`) に影響を与えない
- **保守性**: 削除後、Canary 関連のコードが残っていないことを確認

## 設計

### 削除対象ファイル

#### 1. コントローラー (2 ファイル)

- `app/controllers/api/canary/graphql_controller.rb`
- `app/controllers/v4/local_api/graphql_controller.rb`

#### 2. ルーティング (2 ファイル修正)

- `config/routes/api.rb` - Canary エンドポイントの削除
- `config/routes/local_api.rb` - ファイル全体を削除

#### 3. GraphQL 実装 (123 ファイル)

- `app/graphql/canary/` ディレクトリ全体

#### 4. テストファイル (7 ファイル)

- `spec/requests/api/canary/graphql_spec.rb`
- `spec/graphql/canary/` ディレクトリ全体
  - `spec/graphql/canary/mutations/add_reaction_spec.rb`
  - `spec/graphql/canary/mutations/create_episode_record_spec.rb`
  - `spec/graphql/canary/mutations/create_work_record_spec.rb`
  - `spec/graphql/canary/mutations/delete_record_spec.rb`
  - `spec/graphql/canary/mutations/update_episode_record_spec.rb`
  - `spec/graphql/canary/mutations/update_work_record_spec.rb`

#### 5. 関連ファイル修正・削除 (4 ファイル)

- `app/models/concerns/graphql_resolvable.rb` - 削除（`#global_id` メソッドは使用されていない）
- `app/models/episode.rb` - `include GraphqlResolvable` を削除
- `spec/graphql/beta/mutations/update_record_spec.rb` - Canary 参照の削除
- `spec/graphql/beta/mutations/update_review_spec.rb` - Canary 参照の削除

### 修正内容

#### GraphqlResolvable の削除

`GraphqlResolvable` concern は `Episode` モデルでのみ include されているが、`#global_id` メソッドは実際には呼び出されていない。そのため、concern ファイルごと削除し、`Episode` モデルからも `include` を削除する。

**削除対象**:
- `app/models/concerns/graphql_resolvable.rb` - ファイル削除
- `app/models/episode.rb` - `include GraphqlResolvable` 行を削除

#### ルーティングの修正 (`config/routes/api.rb`)

**変更前**:
```ruby
scope module: :api do
  constraints(subdomain: "api") do
    post :graphql, to: "graphql#execute"

    namespace :canary do
      post :graphql, to: "graphql#execute"
    end

    namespace :v1 do
      # ...
    end
  end
end
```

**変更後**:
```ruby
scope module: :api do
  constraints(subdomain: "api") do
    post :graphql, to: "graphql#execute"

    namespace :v1 do
      # ...
    end
  end
end
```

## タスクリスト

### フェーズ 1: 参考ドキュメントの作成

- [x] **1-1**: Beta と Canary の比較ドキュメントを作成
  - REST API V2 設計の参考資料として `/docs/graphql-api-comparison.md` を作成
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 400 行（実装 400 行 + テスト 0 行）

### フェーズ 2: Canary GraphQL API の削除

- [x] **2-1**: ルーティングとコントローラーの削除
  - `config/routes/api.rb` から Canary エンドポイントを削除
  - `config/routes/local_api.rb` を削除
  - `app/controllers/api/canary/graphql_controller.rb` を削除
  - `app/controllers/v4/local_api/graphql_controller.rb` を削除
  - **想定ファイル数**: 約 4 ファイル（実装 4 + テスト 0）
  - **想定行数**: 約 130 行削除（実装 130 行 + テスト 0 行）

- [x] **2-2**: GraphQL 実装ファイルの削除
  - `app/graphql/canary/` ディレクトリを削除
  - **想定ファイル数**: 約 123 ファイル（実装 123 + テスト 0）
  - **想定行数**: 約 3,000 行削除（実装 3,000 行 + テスト 0 行）

- [x] **2-3**: テストファイルの削除
  - `spec/requests/api/canary/` ディレクトリを削除
  - `spec/graphql/canary/` ディレクトリを削除
  - **想定ファイル数**: 約 7 ファイル（実装 0 + テスト 7）
  - **想定行数**: 約 500 行削除（実装 0 行 + テスト 500 行）

- [x] **2-4**: 関連ファイルの修正・削除
  - `app/models/concerns/graphql_resolvable.rb` を削除（`#global_id` は使用されていない）
  - `app/models/episode.rb` から `include GraphqlResolvable` を削除
  - `spec/graphql/beta/mutations/update_record_spec.rb` から Canary 参照を削除
  - `spec/graphql/beta/mutations/update_review_spec.rb` から Canary 参照を削除
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 15 行削除（実装 11 行 + テスト 4 行）

### フェーズ 3: 検証

- [ ] **3-1**: CI の確認
  - すべてのテストがパスすることを確認
  - Sorbet 型チェックがパスすることを確認
  - **想定ファイル数**: 約 0 ファイル（実装 0 + テスト 0）
  - **想定行数**: 約 0 行（実装 0 行 + テスト 0 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **Beta API の変更**: Beta API は外部から利用されているため、変更しない
- **REST API V2 の実装**: 今回は削除のみ。V2 の実装は別タスクとして行う

## 参考資料

- [graphql-ruby ドキュメント](https://graphql-ruby.org/)
