# コードレビュー: archive-fix

## レビュー情報

| 項目                       | 内容                                       |
| -------------------------- | ------------------------------------------ |
| レビュー日                 | 2026-02-25                                 |
| 対象ブランチ               | archive-fix                                |
| ベースブランチ             | develop                                    |
| 作業計画書（指定があれば） | docs/plans/1_doing/work-episode-archive.md |
| 変更ファイル数             | 222 ファイル                               |
| 変更行数（実装）           | +7776 / -1935 行                           |
| 変更行数（テスト）         | +1793 / -130 行                            |

## 参照するガイドライン

- [@CLAUDE.md#レビュー時に参照するガイドライン](/workspace/CLAUDE.md) - ガイドライン一覧

## 変更ファイル一覧

### 実装ファイル（Go）

- [x] `go/cmd/server/main.go`
- [x] `go/db/migrations/20260210055715_add_status_to_works.sql`
- [x] `go/db/migrations/20260210081156_add_status_to_episodes.sql`
- [x] `go/db/schema.sql`
- [x] `go/internal/handler/db_work/handler.go`
- [ ] `go/internal/handler/db_work/index.go`
- [x] `go/internal/middleware/authorization.go`
- [x] `go/internal/middleware/reverse_proxy.go`
- [x] `go/internal/model/user.go`
- [ ] `go/internal/model/work.go`
- [x] `go/internal/query/models.go`
- [x] `go/internal/query/querier.go`
- [x] `go/internal/query/queries/works.sql`
- [x] `go/internal/query/works.sql.go`
- [x] `go/internal/repository/work.go`
- [x] `go/internal/session/session.go`
- [x] `go/internal/templates/components/db_sidebar.templ`
- [x] `go/internal/templates/components/db_sidebar_templ.go`
- [x] `go/internal/templates/components/pagination.templ`
- [x] `go/internal/templates/components/pagination_templ.go`
- [x] `go/internal/templates/components/status_label.templ`
- [x] `go/internal/templates/components/status_label_templ.go`
- [x] `go/internal/templates/layouts/db.templ`
- [x] `go/internal/templates/layouts/db_templ.go`
- [ ] `go/internal/templates/pages/db_works/index.templ`
- [x] `go/internal/templates/pages/db_works/index_templ.go`
- [x] `go/internal/viewmodel/pagination.go`
- [x] `go/internal/viewmodel/user.go`
- [x] `go/internal/i18n/locales/ja.toml`
- [x] `go/internal/i18n/locales/en.toml`
- [x] `go/sqlc.yaml`

### 実装ファイル（Rails）

- [ ] `rails/app/models/work.rb`
- [ ] `rails/app/models/episode.rb`

### テストファイル（Go）

- [x] `go/internal/handler/db_work/handler_test.go`
- [x] `go/internal/middleware/authorization_test.go`
- [x] `go/internal/repository/work_test.go`
- [x] `go/internal/templates/components/status_label_test.go`
- [x] `go/internal/viewmodel/pagination_test.go`
- [x] `go/internal/middleware/auth_test.go`
- [x] `go/internal/middleware/sentry_test.go`
- [x] `go/internal/testutil/fixtures.go`

### テストファイル（Rails）

- [ ] `rails/spec/models/work_spec.rb`
- [ ] `rails/spec/models/episode_spec.rb`
- [ ] `rails/spec/requests/db/work_publishings/create_spec.rb`
- [ ] `rails/spec/requests/db/work_publishings/destroy_spec.rb`
- [ ] `rails/spec/requests/db/episode_publishings/create_spec.rb`
- [ ] `rails/spec/requests/db/episode_publishings/destroy_spec.rb`

### 設定・その他

- [x] `.github/workflows/rails-ci.yml`
- [x] `.github/workflows/go-ci.yml`
- [x] `.github/workflows/fmt-ci.yml`
- [x] `rails/package.json`
- [x] `package.json`
- [x] `go/Makefile`
- [x] `go/.golangci.yml`
- [x] `go/.air.toml`
- [x] `Makefile`
- [x] `Dockerfile.dev`
- [x] `.gitignore`
- [x] `.oxfmtrc.json`
- [x] `CLAUDE.md`
- [x] `go/CLAUDE.md`
- [x] `rails/CLAUDE.md`
- [x] `go/docs/*.md`
- [x] `docs/plans/**/*.md`
- [x] `docs/reviews/**/*.md`
- [x] `docs/specs/*.md`

## ファイルごとのレビュー結果

### `go/internal/templates/pages/db_works/index.templ` (バグ: シーズン番号のマッピングが不正)

**ステータス**: 要修正

**チェックしたガイドライン**:

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - コーディング規約

**問題点・改善提案**:

- **[バグ]**: `formatSeason` 関数のシーズン番号マッピングが、Rails の `Season` モデル (`{winter: 1, spring: 2, summer: 3, autumn: 4}`) およびテストフィクスチャ (`testutil/fixtures.go`) と一致していない。現在は0-indexedだが、DBには1-indexedで格納されている。

  ```go
  // 問題のあるコード (index.templ:125-136)
  func formatSeason(ctx context.Context, year int32, name int32) string {
  	seasonKey := ""
  	switch name {
  	case 0:
  		seasonKey = "season_winter"
  	case 1:
  		seasonKey = "season_spring"
  	case 2:
  		seasonKey = "season_summer"
  	case 3:
  		seasonKey = "season_autumn"
  	}
  ```

  **修正案**:

  ```go
  func formatSeason(ctx context.Context, year int32, name int32) string {
  	seasonKey := ""
  	switch name {
  	case 1:
  		seasonKey = "season_winter"
  	case 2:
  		seasonKey = "season_spring"
  	case 3:
  		seasonKey = "season_summer"
  	case 4:
  		seasonKey = "season_autumn"
  	}
  ```

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [x] 修正案の通り1-indexedに変更する
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

### `go/internal/model/work.go` (コメント: シーズン番号のマッピングが不正)

**ステータス**: 要修正

**チェックしたガイドライン**:

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - コメントのガイドライン

**問題点・改善提案**:

- **[コメント誤り]**: `SeasonName` フィールドのコメントがRailsのDBデータと一致していない。

  ```go
  // 問題のあるコード (work.go:16)
  SeasonName          *int32 // シーズン番号（0=冬、1=春、2=夏、3=秋）
  ```

  **修正案**:

  ```go
  SeasonName          *int32 // シーズン番号（1=冬、2=春、3=夏、4=秋）
  ```

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [x] 修正案の通り修正する
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

### `go/internal/templates/pages/db_works/index.templ` (アーキテクチャ: Template が Model に直接依存)

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド

**問題点・改善提案**:

- **[@go/docs/architecture-guide.md#レイヤー間の依存関係]**: `IndexPageData` 構造体が `model.DBWorkListItem` に直接依存している。ガイドラインでは「Templates は ViewModel に依存できますが、Model に直接依存することは禁止」とされている。

  ```go
  // 問題のあるコード (index.templ:14-15)
  type IndexPageData struct {
  	Works            []model.DBWorkListItem
  ```

  **修正案**: ViewModel を作成し、`model.DBWorkListItem` → `viewmodel.DBWorkListItem` の変換をハンドラーで行う。`formatSeason` ロジックも ViewModel のフィールドとして事前計算できる。

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [x] ViewModel を作成して依存関係を修正する
  - [ ] 現状のまま（理由を回答欄に記入）
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

### `rails/app/models/work.rb` (設計: Unpublishable concern と status enum の競合)

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@CLAUDE.md](/workspace/CLAUDE.md) - 設計との整合性
- 作業計画書 `docs/plans/1_doing/work-episode-archive.md`

**問題点・改善提案**:

- **[設計との整合性]**: `Work` モデルが `Unpublishable` concern を include しつつ、`status` enum で `published?` / `archived?` をオーバーライドしている。しかし `Unpublishable` が定義する以下のメソッド/スコープは依然として `unpublished_at` カラムを参照しており、二重状態管理になっている:
  - `scope :published` → `unpublished_at: nil` でフィルタ
  - `scope :unpublished` → `unpublished_at IS NOT NULL` でフィルタ
  - `only_kept` → `without_deleted.published` を呼び出し
  - `publish` メソッド → `unpublished_at = nil` を設定
  - `unpublish` メソッド → `unpublished_at = Time.zone.now` を設定

  これにより `work.published?` が `true` を返す（`status == "published"`）が、`Work.published` スコープには含まれない（`unpublished_at` が nil でない）という矛盾が発生し得る。

  作業計画書には「既存の `unpublished_at` カラムからデータを移行」「移行後、`unpublished_at` カラムは将来的に削除」と記載されているが、現時点でスコープやメソッドの移行が不完全。

  **修正案**: `Unpublishable` concern のスコープ（`published`, `unpublished`, `only_kept`）と操作メソッド（`publish`, `unpublish`）を `status` カラムベースに移行するか、concern を削除して `Work` モデルに直接実装する。

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [ ] 今のフェーズで `Unpublishable` のスコープ・メソッドも `status` ベースに移行する
  - [x] 後続フェーズで対応する（計画書に追記）
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

### `rails/app/models/episode.rb` (同上 + private メソッドスタイル)

**ステータス**: 要修正

**チェックしたガイドライン**:

- [@rails/CLAUDE.md](/workspace/rails/CLAUDE.md) - コーディング規約

**問題点・改善提案**:

- **[Unpublishable 競合]**: `Work` と同じく、`Episode` モデルでも `Unpublishable` concern と `status` enum の二重状態管理が発生している。

- **[@rails/CLAUDE.md#Rubyコード]**: `private` ブロックを使用している。`private def` を使用するべき。

  ```ruby
  # 問題のあるコード (episode.rb:134-147)
  private

  def unset_prev_episode_id
    ...
  end

  def update_prev_episode
    ...
  end
  ```

  **修正案**:

  ```ruby
  private def unset_prev_episode_id
    ...
  end

  private def update_prev_episode
    ...
  end
  ```

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [x] `private def` スタイルに修正する
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

### `rails/app/models/work.rb`, `rails/app/models/episode.rb` (コメント: 実装の変遷を説明するコメント)

**ステータス**: 要修正

**チェックしたガイドライン**:

- [@CLAUDE.md#コメントのガイドライン](/workspace/CLAUDE.md) - コメントのガイドライン

**問題点・改善提案**:

- **[@CLAUDE.md#コメントのガイドライン]**: 実装の変遷を説明するコメントが含まれている。「Unpublishable の unpublished_at ベースをオーバーライド」という表現は過去との比較であり、避けるべきコメントに該当する。

  ```ruby
  # 問題のあるコード (work.rb:360, episode.rb:125)
  # status カラムベースの判定メソッド（Unpublishable の unpublished_at ベースをオーバーライド）
  ```

  **修正案**:

  ```ruby
  # status カラムの値で公開状態を判定する
  ```

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [x] 修正案の通りコメントを修正する
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

### `rails/spec/requests/db/work_publishings/create_spec.rb`, `destroy_spec.rb`, `rails/spec/requests/db/episode_publishings/create_spec.rb` (テスト規約違反)

**ステータス**: 要修正

**チェックしたガイドライン**:

- [@rails/CLAUDE.md#RSpec](/workspace/rails/CLAUDE.md) - RSpec コーディング規約

**問題点・改善提案**:

- **[@rails/CLAUDE.md#RSpec]**: `create(...)` を bare で使用している。`FactoryBot.create(...)` を使用するべき。

  ```ruby
  # 問題のあるコード (例: work_publishings/create_spec.rb:7)
  work = create(:work, :unpublished)
  user = create(:registered_user)
  ```

  **修正案**:

  ```ruby
  work_record = FactoryBot.create(:work, :unpublished)
  user_record = FactoryBot.create(:registered_user)
  ```

- **[@rails/CLAUDE.md#RSpec]**: FactoryBot で作成したレコードの変数名に `_record` サフィックスがない。

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [ ] `FactoryBot.create` と `_record` サフィックスに修正する（work_publishings/create, destroy, episode_publishings/create の3ファイル）
  - [x] その他（下の回答欄に記入）

  **回答**:

  ```
  `_record` サフィックスを付けるというガイドラインは他のプロジェクト特有のものでした。Annictのガイドラインからは削除をお願いします。
  ```

### `rails/spec/requests/db/episode_publishings/destroy_spec.rb` (テスト規約: `_record` サフィックス)

**ステータス**: 要修正

**チェックしたガイドライン**:

- [@rails/CLAUDE.md#RSpec](/workspace/rails/CLAUDE.md) - RSpec コーディング規約

**問題点・改善提案**:

- **[@rails/CLAUDE.md#RSpec]**: このファイルは `FactoryBot.create` を正しく使用しているが、変数名に `_record` サフィックスがない。

  ```ruby
  # 問題のあるコード (例: episode_publishings/destroy_spec.rb:6)
  episode = FactoryBot.create(:episode, :published)
  user = FactoryBot.create(:registered_user)
  ```

  **修正案**:

  ```ruby
  episode_record = FactoryBot.create(:episode, :published)
  user_record = FactoryBot.create(:registered_user)
  ```

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [ ] `_record` サフィックスに修正する
  - [x] その他（下の回答欄に記入）

  **回答**:

  ```
  `_record` サフィックスを付けるというガイドラインは他のプロジェクト特有のものでした。Annictのガイドラインからは削除をお願いします。
  ```

### Pending テスト（7件）

**ステータス**: 要確認

**チェックしたガイドライン**:

- 作業計画書 `docs/plans/1_doing/work-episode-archive.md`

**問題点・改善提案**:

- **[設計との整合性]**: 以下のテストファイルに計7件の `pending` テストが存在する。すべて「`published?` が status enum ベースに変更されたため、`unpublished_at` ベースの publish/unpublish 処理との整合が必要」というメッセージ。これは上記の `Unpublishable` concern 競合に起因する。

  | ファイル                              | pending 数               |
  | ------------------------------------- | ------------------------ |
  | `work_publishings/create_spec.rb`     | 3                        |
  | `work_publishings/destroy_spec.rb`    | 1                        |
  | `episode_publishings/create_spec.rb`  | 3                        |
  | `episode_publishings/destroy_spec.rb` | 1 (ただし既存行のみ変更) |

  **対応方針**:

  <!-- 開発者が回答を記入してください -->
  - [x] `Unpublishable` concern の移行と合わせて pending テストを解消する
  - [ ] 後続フェーズで対応する（計画書に追記）
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

## 設計改善の提案

### `go/internal/handler/db_work/handler_test.go`: TestIndex_Empty の重複チェック

**ステータス**: 要確認

**現状**:

`TestIndex_Empty` テスト関数内でHTTPステータスコードを2回チェックしている（同一コード）。

**提案**:

2つ目の重複チェック（約103-105行目）を削除する。

**メリット**:

- テストコードの冗長性を削減

**トレードオフ**:

- なし

**対応方針**:

- [x] 提案通り変更する
- [ ] 現状のまま（理由を回答欄に記入）
- [ ] その他（下の回答欄に記入）

**回答**:

```
（ここに回答を記入）
```

## 総合評価

**評価**: Request Changes

**総評**:

作品・エピソードアーカイブ機能のフェーズ1〜3の実装がレビュー対象です。Go版のDB変更（マイグレーション）、認可ミドルウェア、共通コンポーネント、作品一覧ページの実装は、全体的にアーキテクチャガイドラインに沿った良い実装です。SQL クエリはパラメータ化されており、セキュリティ面も問題ありません。I18nも適切に対応されています。

**必須対応が必要な問題点**:

1. **シーズン番号マッピングのバグ** (`formatSeason` 関数): DBのデータと異なる0-indexedマッピングを使用しており、シーズン名が誤表示される
2. **`Unpublishable` concern と `status` enum の競合**: インスタンスメソッド（`published?`）はstatusを参照するが、スコープや操作メソッドは依然として `unpublished_at` を参照しており、矛盾が発生し得る。これに伴う7件の pending テストも残存
3. **Rails テスト規約違反**: bare `create` の使用、`_record` サフィックスの欠如

**推奨対応**:

4. **Template が Model に直接依存**: ガイドラインでは ViewModel を経由すべきとされている
5. **コメントガイドライン違反**: 実装の変遷を説明するコメント
6. **`private def` スタイル**: `episode.rb` の private メソッド
7. **テストの重複チェック**: `handler_test.go` の TestIndex_Empty
