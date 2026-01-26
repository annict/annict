# iCalendar連携機能 Go版移行 設計書

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

ガイドラインに沿わない設計は、実装時にそのまま実装されてしまうため、
設計書作成の段階でガイドラインに準拠していることを確認してください。
-->

### Go版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - 全体的なコーディング規約
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTPハンドラーガイドライン（**ファイル名は標準の8種類のみ**）
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド
- [@go/docs/i18n-guide.md](/workspace/go/docs/i18n-guide.md) - 国際化ガイド
- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - セキュリティガイドライン
- [@go/docs/templ-guide.md](/workspace/go/docs/templ-guide.md) - templテンプレートガイド

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

iCalendar連携機能は、ユーザーの視聴リスト（「見たい」「見てる」ステータスのアニメ）に基づいて、放送スケジュールをiCalendar形式（.ics）で出力する機能です。ユーザーはこのURLをGoogleカレンダーやAppleカレンダーなどに登録することで、視聴予定のアニメの放送時間を自分のカレンダーで確認できます。

現在Rails版で実装されているこの機能をGo版に移行します。移行に伴い、既存の不具合も修正します。

**目的**:

- Rails版からGo版への段階的移行の一環として、iCalendar連携機能をGo版で再実装する
- 既存の深夜帯放送枠が消える不具合を修正し、ユーザー体験を向上させる
- ユーザーのタイムゾーンを正しく考慮した放送スケジュール表示を実現する

**背景**:

- 現在のRails版実装では、`Date.today.beginning_of_day` を使用しているため、日本時間の午前0時を過ぎると当日の深夜帯（例: 25時放送）の番組がまだ放送前なのにカレンダーから消えてしまう
- ユーザーのタイムゾーン（`users.time_zone`）を考慮せずにフィルタリングを行っているため、海外ユーザーにとって不正確な結果となる可能性がある

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

- ユーザーは `/@:username/ics` または `/ics` エンドポイントからiCalendar形式のファイルを取得できる
- システムは以下の条件でSlot（放送枠）イベントを出力する：
  - ユーザーの視聴リストに登録されている番組（「見たい」「見てる」ステータス）
  - **現在時刻以降かつ7日後までの放送枠**（不具合修正：`Date.today` ではなく現在時刻を基準とする）
  - エピソード情報が存在する
  - ユーザーがまだ視聴していないエピソード
  - 削除されていない放送枠
- システムは以下の条件でWork（作品）イベントを出力する：
  - ユーザーの視聴リストに登録されている作品（「見たい」「見てる」ステータス）
  - 放送開始日（`started_on`）が設定されている
  - 削除されていない作品
- システムはユーザーのタイムゾーン（`users.time_zone`）を考慮してイベントの日時を出力する
- システムはユーザーのロケール（`users.locale`）に基づいて作品タイトルを出力する（日本語または英語）

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

- **パフォーマンス**: 大量のライブラリエントリを持つユーザーでも1秒以内にレスポンスを返す
- **互換性**: RFC 5545（iCalendar）仕様に準拠する
- **互換性**: Apple カレンダーが `@` を含むURLを正しく処理できないため、`/ics` という代替パスを提供する
- **セキュリティ**: 認証不要（URLを知っている人は誰でもアクセス可能）だが、削除されたユーザーのカレンダーは404を返す

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

### 技術スタック

- **iCalendarライブラリ**: 標準ライブラリで実装（外部ライブラリ不要、RFC 5545形式のテキスト生成）
- **タイムゾーン処理**: `time.LoadLocation()` でIANAタイムゾーンを処理

### API設計

#### エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/@:username/ics` | ユーザーのiCalendarを取得（メイン） |
| GET | `/ics?username=:username` | Apple カレンダー互換の代替パス |

#### レスポンス

- **Content-Type**: `text/calendar; charset=utf-8`
- **Content-Disposition**: `attachment; filename="annict.ics"`
- **本文**: iCalendar形式のテキスト

#### エラーレスポンス

| 状況 | ステータスコード |
|------|-----------------|
| ユーザーが存在しない | 404 Not Found |
| ユーザーが削除されている | 404 Not Found |

### コード設計

#### ディレクトリ構造

```
internal/
├── handler/
│   └── ics/
│       ├── handler.go      # Handler構造体と依存性の定義
│       └── show.go         # Show メソッド (GET /@:username/ics, GET /ics)
├── model/
│   └── user_calendar.go    # UserCalendarモデル
├── repository/
│   └── user_calendar.go    # UserCalendarRepository（カレンダーデータ取得）
├── query/
│   └── user_calendar.sql   # カレンダー用クエリ
└── ical/
    └── calendar.go         # iCalendar形式生成ユーティリティ
```

#### 主要な構造体

```go
// internal/model/user_calendar.go

// UserCalendar はユーザーのカレンダーデータを表す
type UserCalendar struct {
    Username string
    TimeZone string
    Locale   string
    Slots    []CalendarSlot
    Works    []CalendarWork
}

// CalendarSlot はカレンダーに表示する放送枠を表す
type CalendarSlot struct {
    ID            int64
    StartedAt     time.Time
    WorkID        int64
    WorkTitle     string
    WorkTitleEn   string
    EpisodeID     int64
    EpisodeTitle  string
    EpisodeNumber string
    ChannelName   string
}

// CalendarWork はカレンダーに表示する作品（放送開始日）を表す
type CalendarWork struct {
    ID        int64
    Title     string
    TitleEn   string
    StartedOn time.Time
}
```

```go
// internal/repository/user_calendar.go

// UserCalendarRepository はカレンダーデータの取得を担当する
type UserCalendarRepository struct {
    queries *query.Queries
}

// GetByUsername はユーザー名からカレンダーデータを取得する
// 読み取り専用の処理のためUsecaseは使用せず、Repositoryで完結する
func (r *UserCalendarRepository) GetByUsername(ctx context.Context, username string) (*model.UserCalendar, error)
```

```go
// internal/ical/calendar.go

// Calendar はiCalendar形式のカレンダーを表す
type Calendar struct {
    TimeZone string
    CalName  string
    Events   []Event
}

// Event はカレンダーイベントを表す
type Event struct {
    UID         string
    Summary     string
    Description string
    Start       time.Time
    End         time.Time
    AllDay      bool  // trueの場合はDate形式、falseの場合はDateTime形式
}

// ToICS はiCalendar形式の文字列を生成する
func (c *Calendar) ToICS() string
```

### 不具合修正の設計

#### 現在の問題

```ruby
# Rails版の問題のあるコード
.where("started_at >= ?", Date.today.beginning_of_day)
```

この実装では、サーバーのタイムゾーン（日本時間）で「今日の0時」を基準にしているため：
- 日本時間 0:00 〜 6:00 頃に放送される深夜アニメが、0:00を過ぎた時点で消えてしまう
- 例: 1月15日 25:00（= 1月16日 01:00）放送の番組は、1月16日 00:00を過ぎると消える

#### 修正方針

```go
// Go版の修正後のコード
// 現在時刻を基準にして、まだ放送されていない番組を取得
now := time.Now()
slots := repo.FindSlotsForCalendar(ctx, FindSlotsInput{
    ProgramIDs:       programIDs,
    StartedAtFrom:    now,                         // 現在時刻以降
    StartedAtTo:      now.AddDate(0, 0, 7),        // 7日後まで
    ExcludeEpisodeIDs: watchedEpisodeIDs,
})
```

この修正により：
- 現在時刻を基準にするため、深夜帯の放送枠が正しく表示される
- ユーザーのタイムゾーンに関係なく、まだ放送されていない番組が常に表示される

### データベースクエリ

```sql
-- 放送枠の取得（修正後）
SELECT
    s.id,
    s.started_at,
    s.work_id,
    s.episode_id,
    s.channel_id,
    w.title AS work_title,
    w.title_en AS work_title_en,
    e.title AS episode_title,
    e.number AS episode_number,
    e.number_text AS episode_number_text,
    c.name AS channel_name
FROM slots s
JOIN works w ON w.id = s.work_id
JOIN episodes e ON e.id = s.episode_id
JOIN channels c ON c.id = s.channel_id
WHERE s.deleted_at IS NULL
  AND w.deleted_at IS NULL
  AND e.deleted_at IS NULL
  AND s.program_id = ANY(@program_ids)
  AND s.started_at >= @started_at_from    -- 現在時刻以降
  AND s.started_at <= @started_at_to      -- 7日後まで
  AND s.episode_id IS NOT NULL
  AND s.episode_id != ALL(@exclude_episode_ids)
ORDER BY s.started_at ASC;
```

### iCalendar出力形式

```ics
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Annict//Annict Calendar//EN
CALSCALE:GREGORIAN
METHOD:PUBLISH
X-WR-TIMEZONE:Asia/Tokyo
X-WR-CALNAME:Annict@username
BEGIN:VTIMEZONE
TZID:Asia/Tokyo
BEGIN:STANDARD
TZOFFSETFROM:+0900
TZOFFSETTO:+0900
TZNAME:JST
DTSTART:19700101T000000
END:STANDARD
END:VTIMEZONE
BEGIN:VEVENT
UID:slot-12345@annict.com
DTSTART;TZID=Asia/Tokyo:20250120T250000
DTEND;TZID=Asia/Tokyo:20250120T253000
SUMMARY:作品タイトル 第1話 サブタイトル (TOKYO MX)
DESCRIPTION:作品タイトル 第1話 サブタイトル\nhttps://annict.com/works/123/episodes/456
END:VEVENT
BEGIN:VEVENT
UID:work-789@annict.com
DTSTART;VALUE=DATE:20250401
DTEND;VALUE=DATE:20250402
SUMMARY:新作アニメタイトル
DESCRIPTION:新作アニメタイトル\nhttps://annict.com/works/789
END:VEVENT
END:VCALENDAR
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

### フェーズ 1: インフラ層の実装

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: iCalendar生成ユーティリティの実装

  - `internal/ical/calendar.go` を新規作成
  - Calendar, Event 構造体の定義
  - `ToICS()` メソッドでRFC 5545準拠のiCalendar形式を生成
  - タイムゾーン対応（VTIMEZONE, TZID）
  - DateTime形式とDate形式の両方をサポート
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 250 行（実装 150 行 + テスト 100 行）

- [x] **1-2**: SQLクエリとUserCalendarRepositoryの実装

  - `internal/query/user_calendar.sql` を新規作成（カレンダー用クエリ）
  - `internal/repository/user_calendar.go` を新規作成
  - ユーザーの視聴リスト取得
  - 放送枠と作品の取得・フィルタリング
  - 現在時刻を基準としたフィルタリング（不具合修正）
  - 視聴済みエピソードの除外
  - ユーザーのタイムゾーン・ロケール対応
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

### フェーズ 2: プレゼンテーション層の実装

- [x] **2-1**: ハンドラーとルーティングの実装

  - `internal/handler/ics/handler.go` を新規作成
  - `internal/handler/ics/show.go` を新規作成
  - ルーティング設定（`/@:username/ics`, `/ics`）
  - レスポンスヘッダー設定（Content-Type, Content-Disposition）
  - エラーハンドリング（404）
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 250 行（実装 120 行 + テスト 130 行）

### フェーズ 3: リバースプロキシ設定

- [x] **3-1**: Go版で処理するパスの追加

  - `internal/middleware/reverse_proxy.go` のホワイトリストに `/@:username/ics` と `/ics` を追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 10 行 + テスト 40 行）

### フェーズ 4: Rails版の削除

- [x] **4-1**: Rails版iCalendar機能の削除

  - `rails/app/controllers/ics_controller.rb` を削除
  - `rails/app/helpers/icalendar_helper.rb` を削除
  - `rails/spec/requests/ics/show_spec.rb` を削除
  - `rails/config/routes.rb` から `/@:username/ics` と `/ics` のルーティングを削除
  - `rails/Gemfile` から `icalendar` gem を削除
  - `bundle install` を実行して `Gemfile.lock` を更新
  - **想定ファイル数**: 約 5 ファイル（実装 5 + テスト 0）
  - **想定行数**: 約 -280 行（削除のみ）

### フェーズ 4a: 不具合修正

- [x] **4a-1**: iCalendar出力時のタイムゾーン変換不具合を修正

  - **問題**: DB から取得した `started_at` は UTC で保持されているが、iCalendar 出力時にユーザーのタイムゾーンに変換せずに出力しているため、時刻が9時間ずれて表示される
  - **例**: 日本時間の 2025-01-21 00:30 のイベントが、カレンダーアプリで 2025-01-20 15:30 と表示される
  - **修正内容**:
    - `internal/ical/calendar.go` の `generateVEvent` 関数で、イベントの開始・終了時刻をカレンダーのタイムゾーンに変換してからフォーマットする
    - UTC で渡された時刻が正しく変換されることを確認するテストを追加
  - **影響範囲**:
    - `internal/ical/calendar.go` - `generateVEvent` 関数の修正
    - `internal/ical/calendar_test.go` - UTC 時刻のテストケース追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 10 行 + テスト 40 行）

- [x] **4a-2**: エピソード番号の`#`重複問題を修正

  - **問題**: データベースの`episodes.number`カラムには`第4話`形式と`#4`形式の2種類が存在する。現在の実装では全てのエピソード番号に`#`を付加しているため、`#4`が`##4`と表示される
  - **例**: `#4`というエピソード番号に対して`#`を付加すると、カレンダーアプリで`##4`と表示される
  - **修正内容**:
    - `internal/handler/ics/show.go`で、エピソード番号が既に`#`で始まっている場合は`#`を追加しないようにする
    - 同様にDESCRIPTIONの構築部分も修正
  - **影響範囲**:
    - `internal/handler/ics/show.go` - サマリーと説明の構築ロジック
    - `internal/handler/ics/show_test.go` - テストケース追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 20 行 + テスト 30 行）
  - **備考**: RFC 5545では`#`のエスケープは不要（エスケープが必要なのはバックスラッシュ、セミコロン、カンマ、改行のみ）

### フェーズ 5: Go版テストの拡充

- [x] **5-1**: Rails版から移行したテストケースの追加

  - Rails版で実装されていた以下のテストケースをGo版に追加:
    - 削除されたユーザーにアクセスした場合、404エラーを返すこと
    - 視聴リストに追加済みのアニメがない場合、空のカレンダーを返すこと
    - 8日以降の放送枠は含まれないこと（現在7日後までを取得）
    - 番組（program）が設定されていないライブラリエントリは無視されること
    - 開始日（started_on）が設定されているアニメがイベントとして含まれること
    - 削除済みの放送枠は含まれないこと
    - wanna_watch/watchingステータスのテスト（両ステータスでカレンダーに含まれること）
  - **対象ファイル**: `go/internal/handler/ics/show_test.go`, `go/internal/repository/user_calendar_test.go`
  - **想定ファイル数**: 約 2 ファイル（テスト 2）
  - **想定行数**: 約 200 行（テスト 200 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **認証付きカレンダーURL**: 現在のRails版と同様に、URLを知っている人は誰でもアクセス可能とする。将来的にトークンベースの認証を検討。
- **カレンダー購読の設定画面**: Go版での設定画面は別途移行予定。
- **リマインダー（VALARM）**: iCalendarのアラーム機能は実装しない。カレンダーアプリ側の設定に委ねる。
- **繰り返しイベント（RRULE）**: 毎週放送のアニメを繰り返しイベントとして表現する機能は実装しない。個別のSlotイベントとして出力する。

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [RFC 5545 - Internet Calendaring and Scheduling Core Object Specification (iCalendar)](https://tools.ietf.org/html/rfc5545)
- [Rails版実装: ics_controller.rb](/workspace/rails/app/controllers/ics_controller.rb)
- [Rails版実装: icalendar_helper.rb](/workspace/rails/app/helpers/icalendar_helper.rb)
- [Rails版テスト: ics/show_spec.rb](/workspace/rails/spec/requests/ics/show_spec.rb)
