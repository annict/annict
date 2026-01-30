# 関連テーブルへの anime_id 追加 設計書

<!--
このテンプレートの使い方:
1. このファイルを `docs/designs/2_todo/` ディレクトリにコピー
   例: cp docs/designs/template.md docs/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 実装ガイドラインの参照

<!--
**重要**: 設計書を作成する前に、対象プラットフォームのガイドラインを必ず確認してください。
特に以下の点に注意してください：
- ディレクトリ構造・ファイル名の命名規則
- コーディング規約
- アーキテクチャパターン

ガイドラインに沿わない設計は、実装時にそのまま実装されてしまうため、
設計書作成の段階でガイドラインに準拠していることを確認してください。
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

作品・エピソードを参照している関連テーブルに `anime_id` カラムを追加し、`animes` テーブルを参照できるようにする。これは [animes テーブル導入設計書](../1_doing/animes-table.md) の一部として、関連テーブルの移行を実現する。

**目的**:

- 関連テーブルが `anime_id` を参照することで、作品↔エピソード変換時の更新対象を最小化する
- 将来的に `work_id` / `episode_id` カラムを廃止できるようにする

**背景**:

- 現状の `works.id` / `episodes.id` は多数のテーブルから参照されており、変換時にすべて更新が必要
- `animes` テーブル導入により、関連テーブルは `anime_id` を参照するようになり、変換時の影響範囲が大幅に縮小される
- 移行期間中は `work_id` / `episode_id` と `anime_id` の両方を保持する

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

- システムは関連テーブルに `anime_id` カラムを追加できる
- システムは既存データの `work_id` / `episode_id` から `anime_id` を設定できる
- システムは新規レコード作成時に `anime_id` を自動設定できる（Rails モデルのコールバック）

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

- **データ整合性**: 移行中も既存機能が正常に動作すること
- **段階的移行**: テーブルごとに個別にデプロイできること
- **パフォーマンス**: インデックスを適切に設定し、クエリのパフォーマンスを維持すること

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

詳細な設計は [animes テーブル導入設計書](../1_doing/animes-table.md) を参照してください。

### 前提条件

この設計書の実装を開始する前に、以下が完了している必要があります：

- [作品に紐付く animes レコードを作成する](./animes-work-sync.md) の フェーズ 1（既存 works データの移行）が完了していること
- [エピソードに紐付く animes レコードを作成する](./animes-episode-sync.md) の フェーズ 1（既存 episodes データの移行）が完了していること

### 対象テーブル一覧

以下のテーブルに `anime_id` カラムを追加する：

| テーブル | 現在の参照 | 追加するカラム |
|---------|-----------|---------------|
| activities | work_id, episode_id | anime_id |
| casts | work_id | anime_id |
| channel_works | work_id | anime_id |
| collection_items | work_id | anime_id |
| comments | work_id | anime_id |
| episode_records | work_id, episode_id | anime_id, parent_anime_id |
| library_entries | work_id, next_episode_id | anime_id, next_anime_id |
| multiple_episode_records | work_id | anime_id |
| programs | work_id | anime_id |
| records | work_id | anime_id |
| slots | work_id, episode_id | anime_id |
| staffs | work_id | anime_id |
| statuses | work_id | anime_id |
| syobocal_alerts | work_id | anime_id |
| trailers | work_id | anime_id |
| vod_titles | work_id | anime_id |
| work_comments | work_id | anime_id |
| work_images | work_id | anime_id |
| work_records | work_id | anime_id |
| work_taggings | work_id | anime_id |
| series_works | work_id | anime_id |

### 各フェーズで実施する内容

1. **マイグレーション**: `anime_id` カラム追加 + インデックス作成
2. **データ移行スクリプト**: 既存データの `work_id` / `episode_id` から `anime_id` を設定
3. **Rails モデルへのコールバック追加**: 新規レコード作成時の `anime_id` 自動設定

### Rails モデルのコールバック実装例

```ruby
# app/models/concerns/anime_id_settable.rb
module AnimeIdSettable
  extend ActiveSupport::Concern

  included do
    before_validation :set_anime_id, if: -> { anime_id.blank? && work_id.present? }
  end

  private

  def set_anime_id
    self.anime_id = work&.anime_id
  end
end

# 各モデルで include する
class Cast < ApplicationRecord
  include AnimeIdSettable
end
```

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

### フェーズ 0: AnimeIdSettable concern の作成

- [ ] **0-1**: (Rails) AnimeIdSettable concern の作成

  - `app/models/concerns/anime_id_settable.rb` の作成
  - 新規レコード作成時に `work_id` / `episode_id` から `anime_id` を自動設定するコールバック
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 20 行 + テスト 30 行）

### フェーズ 1: activities への anime_id 追加

- [ ] **1-1**: (Go) activities テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id/episode_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **1-2**: (Rails) Activity モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 2: casts への anime_id 追加

- [ ] **2-1**: (Go) casts テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **2-2**: (Rails) Cast モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 3: channel_works への anime_id 追加

- [ ] **3-1**: (Go) channel_works テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **3-2**: (Rails) ChannelWork モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 4: collection_items への anime_id 追加

- [ ] **4-1**: (Go) collection_items テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **4-2**: (Rails) CollectionItem モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 5: comments への anime_id 追加

- [ ] **5-1**: (Go) comments テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **5-2**: (Rails) Comment モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 6: episode_records への anime_id 追加

- [ ] **6-1**: (Go) episode_records テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id, parent_anime_id カラム追加 + インデックス）
  - データ移行（work_id/episode_id から anime_id/parent_anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 40 行

- [ ] **6-2**: (Rails) EpisodeRecord モデルへのコールバック追加

  - anime_id と parent_anime_id の自動設定コールバック
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 15 行 + テスト 35 行）

### フェーズ 7: library_entries への anime_id 追加

- [ ] **7-1**: (Go) library_entries テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id, next_anime_id カラム追加 + インデックス）
  - データ移行（work_id/next_episode_id から anime_id/next_anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 40 行

- [ ] **7-2**: (Rails) LibraryEntry モデルへのコールバック追加

  - anime_id と next_anime_id の自動設定コールバック
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 15 行 + テスト 35 行）

### フェーズ 8: multiple_episode_records への anime_id 追加

- [ ] **8-1**: (Go) multiple_episode_records テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **8-2**: (Rails) MultipleEpisodeRecord モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 9: programs への anime_id 追加

- [ ] **9-1**: (Go) programs テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **9-2**: (Rails) Program モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 10: records への anime_id 追加

- [ ] **10-1**: (Go) records テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **10-2**: (Rails) Record モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 11: slots への anime_id 追加

- [ ] **11-1**: (Go) slots テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id/episode_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **11-2**: (Rails) Slot モデルへのコールバック追加

  - anime_id の自動設定コールバック（episode_id または work_id から設定）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 40 行（実装 10 行 + テスト 30 行）

### フェーズ 12: staffs への anime_id 追加

- [ ] **12-1**: (Go) staffs テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **12-2**: (Rails) Staff モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 13: statuses への anime_id 追加

- [ ] **13-1**: (Go) statuses テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **13-2**: (Rails) Status モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 14: syobocal_alerts への anime_id 追加

- [ ] **14-1**: (Go) syobocal_alerts テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **14-2**: (Rails) SyobocalAlert モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 15: trailers への anime_id 追加

- [ ] **15-1**: (Go) trailers テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **15-2**: (Rails) Trailer モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 16: vod_titles への anime_id 追加

- [ ] **16-1**: (Go) vod_titles テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **16-2**: (Rails) VodTitle モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 17: work_comments への anime_id 追加

- [ ] **17-1**: (Go) work_comments テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **17-2**: (Rails) WorkComment モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 18: work_images への anime_id 追加

- [ ] **18-1**: (Go) work_images テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **18-2**: (Rails) WorkImage モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 19: work_records への anime_id 追加

- [ ] **19-1**: (Go) work_records テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **19-2**: (Rails) WorkRecord モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 20: work_taggings への anime_id 追加

- [ ] **20-1**: (Go) work_taggings テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **20-2**: (Rails) WorkTagging モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### フェーズ 21: series_works への anime_id 追加

- [ ] **21-1**: (Go) series_works テーブルへの anime_id カラム追加

  - マイグレーションの作成（anime_id カラム追加 + インデックス）
  - データ移行（work_id から anime_id を設定）
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行

- [ ] **21-2**: (Rails) SeriesWork モデルへのコールバック追加

  - AnimeIdSettable concern の include
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 5 行 + テスト 25 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **work_id / episode_id カラムの削除**: 移行期間中は両方のカラムを保持する

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [animes テーブル導入設計書](../1_doing/animes-table.md)
- [作品に紐付く animes レコードを作成する](./animes-work-sync.md)
- [エピソードに紐付く animes レコードを作成する](./animes-episode-sync.md)
