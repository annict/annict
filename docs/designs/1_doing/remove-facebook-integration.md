# Facebook連携機能の削除 設計書

<!--
このテンプレートの使い方:
1. このファイルを `docs/designs/2_todo/` ディレクトリにコピー
   例: cp docs/designs/template.md docs/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨
-->

## 実装ガイドラインの参照

<!--
**重要**: 設計書を作成する前に、対象プラットフォームのガイドラインを必ず確認してください。
特に以下の点に注意してください：
- ディレクトリ構造・ファイル名の命名規則
- コーディング規約
- アーキテクチャパターン
-->

### Rails版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@rails/CLAUDE.md](/workspace/rails/CLAUDE.md) - 全体的なコーディング規約
- [@rails/docs/architecture-guide.md](/workspace/rails/docs/architecture-guide.md) - アーキテクチャガイド
- [@rails/docs/testing-guide.md](/workspace/rails/docs/testing-guide.md) - テスト戦略ガイド
- [@rails/docs/security-guide.md](/workspace/rails/docs/security-guide.md) - セキュリティガイドライン

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

Rails版Annictに実装されているFacebook連携機能を完全に削除する。この機能は現在使用されておらず、多くのコードが既に`deprecated`フォルダに配置されている。

**目的**:

- 使用されていないコードを削除し、コードベースをシンプルに保つ
- 不要な依存関係（`omniauth-facebook`, `koala`）を削除し、セキュリティリスクを低減
- メンテナンスコストの削減

**背景**:

- Facebook連携機能は長期間使用されておらず、`deprecated`フォルダに多くのコードが移動済み
- `Setting`モデルでは既に`ignored_columns`でFacebook関連カラムが無視されている
- Facebook Graph APIの仕様変更に追従するコストが不要になる

## 要件

<!--
ガイドライン:
- 機能要件: 「何ができるべきか」を記述
- 非機能要件: 「どのように動くべきか」を必要に応じて記述
-->

### 機能要件

<!--
「ユーザーは〇〇できる」「システムは〇〇する」という形式で記述
箇条書きで簡潔に
-->

- Facebook OAuthログイン機能を削除する
- Facebookへの投稿シェア機能を削除する
- 設定画面からFacebook連携UIを削除する
- GraphQL APIからFacebook関連フィールドを削除する
- Facebook関連のGem依存関係を削除する（`omniauth-facebook`、`koala`）
- OmniAuth関連のGem依存関係を削除する（`omniauth-rails_csrf_protection`）- Facebookが唯一のOmniAuthプロバイダーのため不要

### 非機能要件

<!--
必要に応じて以下のような項目を追加してください：
- セキュリティ（認証、認可、暗号化、監査ログなど）
- パフォーマンス（応答時間、スループット、リソース使用量など）
- ユーザビリティ（UX）（使いやすさ、わかりやすさ、アクセシビリティなど）
- 可用性・信頼性（稼働率、障害時の挙動、エラーハンドリングなど）
- 保守性（テストのしやすさ、コードの読みやすさ、ドキュメントなど）

不要な場合はこのセクション全体を削除してください。
-->

- **後方互換性**: GraphQL APIの破壊的変更となるため、APIバージョンの検討が必要
- **データ保全**: 既存のデータベースカラムは即座に削除せず、段階的に対応する
- **他機能への影響なし**: Twitter連携など他のSNS連携機能に影響を与えない

## 設計

<!--
ガイドライン:
- 技術的な実装の詳細を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - 技術スタック（使用するライブラリ、フレームワーク、ツールなど）
  - アーキテクチャ（システム全体の構成、コンポーネント間の関係など）
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - セキュリティ設計（認証・認可、トークン管理、Rate Limitingなど）
  - コード設計（パッケージ構成、主要な構造体、インターフェースなど）
  - テスト戦略（単体テスト、統合テスト、E2Eテストの方針）
  - マイグレーション管理（データベースマイグレーションの方針）
  - 実装方針（特記事項、既存システムとの関係、制約など）

不要な場合はこのセクション全体を削除してください。
-->

### 削除対象一覧

#### 1. Gem依存関係

| Gem | 用途 | 削除方法 |
|-----|------|---------|
| `omniauth-facebook` | Facebook OAuthストラテジー | Gemfileから削除 |
| `omniauth-rails_csrf_protection` | OmniAuth CSRF保護 | Gemfileから削除（Facebookが唯一のOmniAuthプロバイダーのため不要） |
| `koala` | Facebook Graph API SDK | Gemfileから削除 |

**備考**: `omniauth`本体は`omniauth-facebook`の依存関係として含まれているため、`omniauth-facebook`を削除すると自動的に削除されます。

#### 2. 設定ファイル

| ファイル | 削除内容 |
|---------|---------|
| `config/initializers/devise.rb` | Facebook OmniAuth設定（L7-9） |
| `config/routes.rb` | `devise_for`から`omniauth_callbacks: "callbacks"`を削除（L9）、`/friends`ルートを削除（L223） |

#### 3. 環境変数

削除対象:
- `FACEBOOK_APP_ID`
- `FACEBOOK_SECRET_KEY`

#### 4. モデル

| ファイル | 削除/修正内容 |
|---------|-------------|
| `app/models/user.rb` | `:omniauthable`と`omniauth_providers`設定を削除、`facebook`メソッド削除、`social_friends`メソッド削除、`get_large_avatar_image`のFacebookケース削除 |
| `app/models/provider.rb` | `enumerize :name`から`:facebook`削除、`token_expires_at=`のFacebook条件削除 |
| `app/models/episode_record.rb` | `facebook_share_title`、`facebook_share_body`削除（`share_url_with_query`は残す - リダイレクト用に必要） |
| `app/models/work_record.rb` | `facebook_share_title`、`facebook_share_body`削除 |

#### 5. コントローラー

| ファイル | 削除内容 |
|---------|---------|
| `app/controllers/callbacks_controller.rb` | ファイル全体を削除（Facebookが唯一のOmniAuthプロバイダーのため不要） |
| `app/controllers/friends_controller.rb` | ファイル全体を削除（`UserSocialFriendsQuery`削除に伴い不要） |

**残すファイル**:
- `app/controllers/legacy/record_redirects_controller.rb` - 古いFacebook共有URLから現在のURLへのリダイレクト用。既存リンクの互換性維持のため残す

#### 6. コンポーネント・ビュー

| ファイル | 削除内容 |
|---------|---------|
| `app/components/deprecated/buttons/share_to_facebook_button_component.rb` | ファイル全体 |
| `app/components/deprecated/buttons/share_to_facebook_button_component.html.slim` | ファイル全体 |
| `app/views/settings/providers/index.html.erb` | Facebook接続UI（L13-38） |
| `app/views/friends/index.html.erb` | ファイル全体（`FriendsController`削除に伴い不要） |

#### 7. JavaScript

| ファイル | 削除内容 |
|---------|---------|
| `app/javascript/controllers/share-to-facebook-button-controller.ts` | ファイル全体 |

#### 8. サービス・クエリ

| ファイル | 削除/修正内容 |
|---------|-------------|
| `app/services/deprecated/facebook_service.rb` | ファイル全体を削除 |
| `app/queries/deprecated/user_social_friends_query.rb` | ファイル全体を削除（Twitter連携も現在は使用されていないため不要） |

**残すファイル**:
- `app/services/deprecated/sns_image_service.rb` - `og:image`を取得するRakeタスクで使用中のため残す

#### 9. ヘルパー

| ファイル | 削除内容 |
|---------|---------|
| `app/helpers/head_helper.rb` | `fb: { app_id: ... }`メタタグ設定 |

#### 10. GraphQL API

**変更なし** - 後方互換性のため、以下のフィールドはすべて残す:
- `facebook_click_count` - Beta/Canary両方で残す
- `facebook_og_image_url` - Beta/Canary両方で残す

#### 11. テスト

**残すテストファイル**:
- `spec/requests/api/v1/*_spec.rb` - REST APIは後方互換のため変更しない
- `spec/requests/legacy/record_redirects_spec.rb` - リダイレクト機能は残すためテストも残す

### データベースカラム

以下のカラムは現時点では**削除しない**（将来的に別タスクで対応）:

| テーブル | カラム |
|---------|--------|
| `episode_records` | `facebook_url_hash`, `facebook_click_count` |
| `works` | `facebook_og_image_url` |
| `records` | `facebook_click_count` |
| `settings` | `share_record_to_facebook`, `share_review_to_facebook`, `share_status_to_facebook`（既に`ignored_columns`） |

**理由**: データベースカラムの削除は破壊的変更であり、本タスクではコードの削除に集中する。カラム削除は別途計画する。

## タスクリスト

<!--
ガイドライン:
- フェーズごとに段階的な実装計画を記述
- チェックボックスで進捗を管理
- **重要**: 1タスク = 1 Pull Request の粒度で作成してください
- **重要**: 各タスクには想定ファイル数と想定行数を明記してください（PRサイズの見積もりのため）
- 想定ファイル数は「実装」と「テスト」に分けて記載してください
- 想定行数も「実装」と「テスト」に分けて記載してください
- 依存関係を明確に
- Pull Requestのガイドラインは CLAUDE.md を参照（変更ファイル数20以下、変更行数300行以下）

タスク番号の付け方:
- 各タスクには階層的な番号を付与します（例: 1-1, 1-2, 2-1, 2-2）
- フォーマット: **フェーズ番号-タスク番号**: タスク名
- **フェーズ番号は半角英数字とハイフンのみで表記**してください（ブランチ名に使用するため）
  - 例: フェーズ 1, フェーズ 2, フェーズ 5a（フェーズ 5 と 6 の間に追加する場合）
  - NG: フェーズ 5.5（ドットは使用不可）
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）
-->

### フェーズ 1: Gem依存関係とDevise設定の削除

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: Gem依存関係とDevise設定の削除

  - Gemfileから`omniauth-facebook`、`omniauth-rails_csrf_protection`、`koala`を削除
  - `bundle install`を実行して`Gemfile.lock`を更新
  - `config/initializers/devise.rb`からFacebook OmniAuth設定を削除
  - Sorbet RBIファイルが自動削除されることを確認
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 25 行（実装 15 行 + テスト 10 行）

### フェーズ 2: コアモデルの修正

- [x] **2-1**: Userモデルからfacebook関連コードとomniauthable設定を削除

  - Devise設定から`:omniauthable`と`omniauth_providers`を削除
  - `facebook`メソッドを削除
  - `social_friends`メソッドを削除（`UserSocialFriendsQuery`削除に伴い不要）
  - `get_large_avatar_image`メソッドからFacebookケースを削除
  - 関連テストの修正
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 35 行（実装 20 行 + テスト 15 行）

- [x] **2-1-1**: Userモデルから不要になったメソッドを削除

  - `build_relations`メソッドを削除（OmniAuth削除に伴い呼び出し元がなくなったため不要）
  - `get_large_avatar_image`メソッドを削除（`build_relations`削除に伴い呼び出し元がなくなったため不要）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

- [x] **2-2**: Providerモデルからfacebook関連コードを削除

  - `enumerize :name`から`:facebook`を削除
  - `token_expires_at=`メソッドからFacebook条件を削除
  - 関連テストの修正
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 20 行（実装 10 行 + テスト 10 行）

- [ ] **2-3**: EpisodeRecord・WorkRecordモデルからfacebook関連コードを削除

  - `episode_record.rb`から`facebook_share_title`、`facebook_share_body`削除（`share_url_with_query`は残す - リダイレクト用に必要）
  - `work_record.rb`から`facebook_share_title`、`facebook_share_body`削除
  - 関連テストの修正
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 40 行（実装 20 行 + テスト 20 行）

### フェーズ 3: コントローラー・ビューの修正

- [x] **3-1**: CallbacksController・FriendsControllerとルーティング設定を削除

  - `app/controllers/callbacks_controller.rb`ファイル全体を削除（Facebookが唯一のOmniAuthプロバイダーのため不要）
  - `app/controllers/friends_controller.rb`ファイル全体を削除（`UserSocialFriendsQuery`削除に伴い不要）
  - `app/views/friends/index.html.erb`を削除
  - `config/routes.rb`から`omniauth_callbacks: "callbacks"`と`/friends`ルートを削除
  - 関連テストを削除
  - **想定ファイル数**: 約 5 ファイル（実装 4 + テスト 1）
  - **想定行数**: 約 120 行（実装 100 行 + テスト 20 行）

- [x] **3-2**: 設定画面からFacebook連携UIを削除

  - `settings/providers/index.html.erb`からFacebook接続UIを削除
  - 関連テストの修正
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 25 行 + テスト 5 行）

### フェーズ 4: コンポーネント・JavaScript・ヘルパーの削除

- [ ] **4-1**: Facebook共有ボタンコンポーネントとJavaScriptを削除

  - `share_to_facebook_button_component.rb`を削除
  - `share_to_facebook_button_component.html.slim`を削除
  - `share-to-facebook-button-controller.ts`を削除
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 60 行（実装 60 行 + テスト 0 行）

- [ ] **4-2**: ヘルパーからFacebookメタタグ設定を削除

  - `head_helper.rb`から`fb: { app_id: ... }`を削除
  - 関連テストの修正
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 15 行（実装 5 行 + テスト 10 行）

### フェーズ 5: サービス・クエリの削除

- [ ] **5-1**: Deprecated Facebookサービス・クエリを削除

  - `facebook_service.rb`を削除
  - `user_social_friends_query.rb`を削除（Twitter連携も現在は使用されていないため不要）
  - `User#authorized_to?`メソッドを削除（`user_social_friends_query.rb`削除に伴い呼び出し元がなくなるため不要）
  - （`sns_image_service.rb`は`og:image`取得Rakeタスクで使用中のため残す）
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 70 行（実装 70 行 + テスト 0 行）

### フェーズ 6: 最終クリーンアップ

- [ ] **7-1**: 最終クリーンアップと動作確認

  - 全テストが通ることを確認
  - `make lint`、`make sorbet`が通ることを確認
  - 環境変数ドキュメントの更新（必要な場合）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 20 行（実装 10 行 + テスト 10 行）

- [ ] **7-2**: 不要なI18nキーの削除

  - 削除したコードで参照されていたI18nキーを検索
  - 他の箇所で使用されていないI18nキーを`config/locales/`から削除
  - `bin/rails i18n:unused`などで未使用キーを確認（利用可能な場合）
  - **想定ファイル数**: 約 4 ファイル（実装 4 + テスト 0）
  - **想定行数**: 約 20 行（実装 20 行 + テスト 0 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **データベースカラムの削除**: `facebook_url_hash`、`facebook_click_count`、`facebook_og_image_url`などのカラム削除は破壊的変更のため、別タスクで対応
- **providersテーブルのFacebookデータ削除**: 既存ユーザーのproviderレコード削除は別途検討
- **他SNS連携（Twitter/Gumroad）の削除**: 本タスクのスコープはFacebookのみ

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Devise OmniAuth](https://github.com/heartcombo/devise/wiki/OmniAuth:-Overview) - Devise OmniAuth設定ドキュメント
- [omniauth-facebook](https://github.com/simi/omniauth-facebook) - Facebook OAuthストラテジー
- [koala](https://github.com/arsduo/koala) - Facebook Graph API Ruby SDK
