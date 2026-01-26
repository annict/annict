# テストデータ生成機能 設計書

## 概要

開発環境での動作確認用に大量のテストデータを生成する機能を実装します。Rails版とGo版の両方で正常に参照できるデータを作成し、パフォーマンス問題の検出やUIの動作確認を可能にします。

**目的**:

- 開発環境で大量のデータを使った動作確認を可能にする
- パフォーマンス問題を早期に発見する（特にN+1問題、遅いクエリなど）
- UIの表示確認を実際のデータに近い状態で行う
- Rails版との互換性を検証する（特にOAuthトークン、セッション、画像など）
- 将来の機能実装で再利用可能なUsecaseを作成する

**背景**:

- 現在の開発環境にはテストデータが少なく、パフォーマンス問題に気づきにくい
- Rails版で使用されている全テーブルのうち、Go版でテストデータを生成できるのは一部のみ
- 大量データでの動作確認が必要（作品10,000件、ユーザー30,000件、視聴記録1,000,000件）
- ヘビーユーザー（大量の視聴記録、フォロー/フォロワー、アクティビティを持つユーザー）の動作確認が必要

## 要件

### 機能要件

- **データ生成コマンド**: `make seed` を実行すると、テストデータが生成される
- **データ量**:
  - ユーザー（users）: 30,000件
  - ユーザープロフィール（profiles）: 30,000件（usersに紐づく）
  - 作品（works）: 10,000件
  - エピソード（episodes）: 約120,000件（1作品あたり平均12話）
  - 視聴記録（episode_records + records）: 1,000,000件
  - 作品画像（work_images）: 10,000件（全作品に画像を付与）
  - フォロー関係（follows）: 約100,000件
  - アクティビティ（activities）: 1,000,000件（視聴記録に対応）
  - OAuth アクセストークン（oauth_access_tokens）: 100件（動作確認用）
- **リアルなデータ**: 実際のアニメタイトル、シーズン情報（2020年〜2025年）など、それらしいデータを生成
- **ヘビーユーザー**: 以下の特徴を持つテストユーザーを1人生成
  - ユーザー名: `heavy_user`（ログインしやすいように）
  - 視聴記録: 10,000件（全体の1%）
  - フォロワー: 1,000人
  - フォロー: 500人
  - フォロイーもアクティブ（各フォロイーが100件以上の視聴記録を持つ）
  - アクティビティ: 10,000件（タイムラインが充実）
- **作品画像**: ランダムな模様を機械的に生成（実際の画像ファイルを作成し、MinIOにアップロード）
- **Rails版互換性**: 生成したデータはRails版でも正常に参照・操作できる
  - OAuth トークンでWeb APIが実行可能
  - セッション情報が正しく読み取れる
  - 画像URLが正しく生成される（Shrine + imgproxy）
- **Usecase設計**: データ生成ロジックはシード専用Usecaseとして実装（`internal/usecase/seed/`配下）
  - シード用はバルクインサート・高速化に特化
  - 将来のユーザー操作用Usecase（単発処理）は別途実装（YAGNI原則に従い今は不要）
  - Repository層（sqlc）や共通ヘルパー（`internal/seed/`）は再利用可能

### 非機能要件

- **パフォーマンス**: 大量データ生成を効率的に行う
  - バルクインサートを活用（1,000件ごとにコミット）
  - トランザクション管理を適切に行う
  - 進捗表示（プログレスバー）を表示する
- **冪等性**: 同じコマンドを複数回実行しても安全（既存データを削除してから生成）
- **データ整合性**: 外部キー制約を満たすデータを生成する
  - user_id, work_id, episode_id などの参照整合性
  - created_at, updated_at のタイムスタンプ
  - NOT NULL制約、UNIQUE制約の遵守
- **ログ出力**: 生成中の進捗とエラーをログに出力する

## 設計

### 技術スタック

- **Go 1.25.1**: seedプログラムの実装言語
- **sqlc**: データベースアクセス（既存のクエリを活用）
- **github.com/schollz/progressbar/v3**: 進捗表示
- **github.com/brianvoe/gofakeit/v6**: フェイクデータ生成
- **image/png**: 画像生成（Goの標準ライブラリ）
- **AWS SDK for Go v2 (aws-sdk-go-v2/service/s3)**: Cloudflare R2への画像アップロード（S3互換API）

### アーキテクチャ

```
cmd/seed/
└── main.go                    # エントリポイント（seedコマンド）

internal/usecase/seed/         # シード専用Usecase（バルクインサート・高速化特化）
├── create_user.go             # ユーザー生成
├── create_work.go             # 作品生成
├── create_episode.go          # エピソード生成
├── create_episode_record.go   # 視聴記録生成
├── cleanup_work_images.go     # 作品画像のクリーンアップ（S3バケット + DB）
├── create_work_image.go       # 作品画像生成
├── create_follow.go           # フォロー関係生成
├── create_oauth_token.go      # OAuth トークン生成
└── create_heavy_user.go       # ヘビーユーザー生成

internal/seed/                 # 共通ヘルパー（シード・本番両方で使える）
├── data.go                    # リアルなデータソース（アニメタイトル、シーズン情報など）
├── image.go                   # ランダム画像生成ロジック
└── progress.go                # 進捗表示ヘルパー

internal/testutil/
└── builder.go                 # 既存のビルダー（テスト用、シードでも活用可能）
```

### データベース設計

既存のテーブルをそのまま使用します。主要なテーブルと生成順序：

1. **users, profiles**: ユーザーとプロフィール（bcryptでパスワードハッシュ化）
2. **works**: 作品（season_year, season_name, media などを設定）
3. **episodes**: エピソード（work_idに紐づく）
4. **records**: 視聴記録の親レコード（user_id, work_id）
5. **episode_records**: エピソード単位の視聴記録（record_id, episode_id, rating など）
6. **work_images**: 作品画像（Shrine形式のimage_dataカラム）
7. **follows**: フォロー関係（follower_id, following_id）
8. **activities**: アクティビティ（episode_recordsに対応）
9. **oauth_applications**: OAuthアプリケーション（動作確認用に1件）
10. **oauth_access_tokens**: アクセストークン（動作確認用に100件）

### 作品画像生成

Shrine gem + Cloudflare R2 の仕様に従って画像を生成・アップロードします：

1. **ランダム画像生成**: Goの`image`パッケージで800x600pxのPNG画像を生成
   - ランダムな色のグラデーションやパターンを描画
   - `shrine/workimage/{UUID}.png` のパスで保存
2. **Cloudflare R2へアップロード**: AWS SDK for Go v2 (S3互換API) を使用してアップロード
3. **image_dataカラムの生成**: Shrine形式のJSON
   ```json
   {
     "id": "shrine/workimage/abc123.png",
     "storage": "s3",
     "metadata": {
       "size": 102400,
       "mime_type": "image/png",
       "width": 800,
       "height": 600
     }
   }
   ```

### Usecase設計

各データ生成ロジックはシード専用Usecaseとして実装します（`internal/usecase/seed/`パッケージ）。
バルクインサートや高速化に特化し、将来のユーザー操作用Usecaseとは分離します。

**例: CreateUserUsecase (シード専用)**

```go
// internal/usecase/seed/create_user.go
package seed

type CreateUserUsecase struct {
    db      *sql.DB
    queries *repository.Queries
}

type CreateUserParams struct {
    Username string
    Email    string
    Password string
}

type CreateUserResult struct {
    UserID int64
}

// バルクインサート用メソッド（1000件まとめて処理）
func (uc *CreateUserUsecase) ExecuteBatch(ctx context.Context, users []CreateUserParams, progressBar *progressbar.ProgressBar) ([]CreateUserResult, error) {
    // パスワードハッシュ化（bcrypt、Rails互換）
    // ユーザー + プロフィールをバルクインサート
    // 1000件ごとにコミット
    // 進捗表示を更新
}
```

**例: CreateEpisodeRecordUsecase (シード専用)**

```go
// internal/usecase/seed/create_episode_record.go
package seed

type CreateEpisodeRecordUsecase struct {
    db      *sql.DB
    queries *repository.Queries
}

type CreateEpisodeRecordParams struct {
    UserID    int64
    EpisodeID int64
    WorkID    int64
    Rating    float64
    Body      *string
    WatchedAt time.Time
}

type CreateEpisodeRecordResult struct {
    RecordID        int64
    EpisodeRecordID int64
}

// バルクインサート用メソッド
func (uc *CreateEpisodeRecordUsecase) ExecuteBatch(ctx context.Context, records []CreateEpisodeRecordParams, progressBar *progressbar.ProgressBar) ([]CreateEpisodeRecordResult, error) {
    tx, _ := uc.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // 1. Record作成（親レコード、バルクインサート）
    // 2. EpisodeRecord作成（子レコード、バルクインサート）
    // 3. Activity作成（バルクインサート）
    // 4. カウンター更新（まとめて更新）
    // 5. 進捗表示を更新

    tx.Commit()
    return results, nil
}
```

**将来のユーザー操作用Usecase（今は実装不要）**

ユーザー登録フォームなどで単発作成が必要になったら、`internal/usecase/create_user.go` として別途実装します。
シード用とは責務が異なるため、分離することで保守性が向上します。

### 実装方針

- **段階的な実装**: 基盤→ユーザー→作品→視聴記録→画像→ヘビーユーザーの順に実装
- **既存コードの活用**: `internal/testutil` の既存ビルダーを拡張して使用
- **バルクインサート**: 1,000件ごとにコミットして効率化
- **進捗表示**: progressbar で進捗を可視化
- **エラーハンドリング**: 途中でエラーが発生したらロールバックして中断
- **Rails互換性の検証**: seedコマンド実行後、Rails版でデータが正しく表示されるか手動確認

## タスクリスト

### フェーズ 1: 基盤整備

- [x] **seedコマンドの基盤実装**

  - `cmd/seed/main.go` を作成（エントリポイント）
  - `make seed` コマンドをMakefileに追加
  - データベース接続、トランザクション管理
  - 既存データのクリーンアップ処理（`DELETE FROM` で全削除）
  - 進捗表示ライブラリ（progressbar）の統合
  - **想定ファイル数**: 約 3 ファイル（実装 2 + Makefile 1）
  - **想定行数**: 約 150 行（実装 120 行 + Makefile 30 行）

- [x] **リアルなデータソースの準備**

  - `internal/seed/data.go` を作成
  - アニメタイトルのリスト（100〜200件程度を用意し、ランダムに組み合わせて10,000件生成）
  - シーズン情報（2020年〜2025年、春夏秋冬）
  - メディアタイプ（TV, OVA, Movie, Web など）
  - ユーザー名のパターン（faker + 連番）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

- [x] **testutil ビルダーの拡張**
  - 既存の `internal/testutil/builder.go` を拡張
  - バルクインサート対応（`BatchInsert` メソッドの追加）
  - 進捗コールバック対応（生成中の進捗を通知）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

### フェーズ 2: 基本データ生成Usecase実装

- [x] **ユーザー生成Usecaseの実装**

  - `internal/usecase/seed/create_user.go` を作成
  - bcryptでパスワードハッシュ化（Rails互換）
  - users + profiles の同時生成
  - バルクインサート対応（1,000件ごとにコミット）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

- [x] **作品生成Usecaseの実装**

  - `internal/usecase/seed/create_work.go` を作成
  - リアルなタイトル、シーズン情報の設定
  - season_year, season_name, media, official_site_url などを設定
  - バルクインサート対応
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

- [x] **エピソード生成Usecaseの実装**

  - `internal/usecase/seed/create_episode.go` を作成
  - 1作品あたり12話（平均）を生成
  - number, title, sort_number などを設定
  - prev_episode_id の設定（エピソード連鎖）
  - バルクインサート対応
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

- [x] **視聴記録生成Usecaseの実装**
  - `internal/usecase/seed/create_episode_record.go` を作成
  - Record（親） + EpisodeRecord（子）の同時生成
  - rating, rating_state, body の設定
  - Activity の同時生成
  - カウンター更新（User#episode_records_count, Work#records_count など）
  - バルクインサート対応（最も時間がかかる処理）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 250 行（実装 200 行 + テスト 50 行）

### フェーズ 3: 高度な機能（画像、ヘビーユーザー、OAuth）

- [x] **ランダム画像生成機能の実装**

  - `internal/seed/image.go` を作成
  - Goの`image`パッケージで800x600pxのPNG画像を生成
  - ランダムな色のグラデーションやパターンを描画
  - AWS SDK for Go v2 (aws-sdk-go-v2/service/s3) の統合（Cloudflare R2へアップロード）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

- [x] **作品画像のクリーンアップUsecaseの実装**

  - `internal/usecase/seed/cleanup_work_images.go` を作成
  - Cloudflare R2 上の `shrine/workimage/` プレフィックス配下のすべての画像を削除
  - データベース上の `work_images` テーブルも削除（既に `cmd/seed/main.go` で実装済み）
  - AWS SDK for Go v2 (aws-sdk-go-v2/service/s3) の ListObjectsV2 と DeleteObjects を使用
  - **実行タイミング**: 作品画像生成の**前**に実行（`cmd/seed/main.go` で呼び出し）
  - **冪等性の確保**: `make seed` を何度実行しても、S3バケット内に孤立した画像ファイルが残らないようにする
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [x] **作品画像生成Usecaseの実装**

  - `internal/usecase/seed/create_work_image.go` を作成
  - ランダム画像を生成してCloudflare R2にアップロード
  - Shrine形式のimage_dataカラムを生成
  - work_images テーブルにレコードを作成
  - バルクインサート対応
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

- [x] **フォロー関係生成Usecaseの実装**

  - `internal/usecase/seed/create_follow.go` を作成
  - ランダムなユーザー間でフォロー関係を作成
  - カウンター更新（User#followers_count, User#following_count）
  - バルクインサート対応
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [x] **ヘビーユーザー生成Usecaseの実装**

  - `internal/usecase/seed/create_heavy_user.go` を作成
  - ユーザー名: `heavy_user`、パスワード: `password`（ログインしやすいように）
  - 視聴記録: 10,000件
  - フォロワー: 1,000人、フォロー: 500人
  - フォロイーもアクティブ（各フォロイーが100件以上の視聴記録）
  - 既存のUsecaseを組み合わせて実装
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

- [x] **OAuth トークン生成Usecaseの実装**
  - `internal/usecase/seed/create_oauth_token.go` を作成
  - oauth_applications テーブルに動作確認用アプリケーションを1件作成
  - oauth_access_tokens テーブルにトークンを100件作成
  - Rails版との互換性確認（Web APIが実行可能か）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

### フェーズ 4: 統合・検証・最適化

- [x] **seedコマンドの統合とテスト**

  - すべてのUsecaseを`cmd/seed/main.go`で統合
  - 生成順序の最適化（外部キー制約を考慮）
  - エラーハンドリングの改善
  - 進捗表示の改善（各フェーズの進捗を表示）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [x] **作品画像のimage_data構造の修正**

  - 現状のimage_data構造が実際のShrineの仕様と異なる
  - 実際の構造では、トップレベルに `"master"` キーがあり、その中にバージョン情報が格納される
  - 修正内容:
    - `"master"` キーでラップする
    - `id` から `shrine/` プレフィックスを削除
    - `storage` の値を `"s3"` から `"store"` に変更
    - `metadata` に `filename` フィールドを追加
  - 対象ファイル: `internal/seed/image.go`, テストファイル2件
  - **想定ファイル数**: 約 3 ファイル（実装 1 + テスト 2）
  - **想定行数**: 約 50 行（構造体とJSON生成ロジックの修正 + テスト修正）

- [ ] **Rails版互換性の検証**

  - seedコマンド実行後、Rails版で以下を確認:
    - 作品一覧ページが表示される
    - ユーザーページが表示される
    - 視聴記録が表示される
    - 作品画像が正しく表示される（imgproxy経由）
    - OAuth トークンでWeb APIが実行できる
    - ヘビーユーザーでログインして大量データの表示を確認
  - 問題があれば修正
  - **想定ファイル数**: 約 1 ファイル（ドキュメント）
  - **想定行数**: 約 50 行（検証手順書）

- [x] **作品画像生成のパフォーマンス最適化**
  - 作品画像生成が遅い（10,000件が完全にシーケンシャル処理）
  - 以下の4つの改善策を段階的に実装することで、20-50倍の高速化を目指す

- [x] **画像サイズの最適化（500x500に変更）**
  - 現在の800x600pxを500x500pxに変更
  - Annictの作品画像は1:1で表示されるため、500x500pxで十分
  - ピクセル数が約48%削減されるため、生成・エンコード時間が改善
  - 対象ファイル: `internal/seed/image.go`
  - **想定ファイル数**: 約 1 ファイル
  - **想定行数**: 約 10 行（定数変更）

- [x] **PNG圧縮レベルの調整（BestSpeed）**
  - デフォルトの圧縮レベルを`png.BestSpeed`に変更
  - テストデータなので画質より速度を優先
  - エンコード時間が2-3倍高速化
  - 対象ファイル: `internal/seed/image.go` の `GenerateRandomImage` 関数
  - **想定ファイル数**: 約 1 ファイル
  - **想定行数**: 約 10 行（エンコーダー設定変更）

- [x] **ランダムパターンの簡略化**
  - 現在のランダムパターン描画（10-30個の円）を3-8個に削減
  - パターン描画の計算量を削減
  - 1.5-2倍の高速化
  - 対象ファイル: `internal/seed/image.go` の `addRandomPattern` 関数
  - **想定ファイル数**: 約 1 ファイル
  - **想定行数**: 約 5 行（パターン数の調整）

- [x] **並列処理の導入（Worker Pool パターン）**
  - goroutineを使った並列処理で画像生成・アップロードを並列化
  - 10並列（調整可能）でネットワークI/Oの待ち時間を埋める
  - 10-20倍の高速化が期待できる（最も効果が高い）
  - 対象ファイル: `internal/usecase/seed/create_work_image.go` の `executeBatchWithTx` 関数
  - **想定ファイル数**: 約 1 ファイル
  - **想定行数**: 約 100 行（Worker Pool実装）

- [x] **フォロイー視聴記録生成のパフォーマンス最適化**
  - フォロイーの視聴記録生成（1,000人 × 100件 = 100,000件）が遅い
  - 原因: `getRandomEpisodes` メソッドが100,000回のクエリを実行している（1件ずつ `LIMIT 1 OFFSET N`）
  - 改善策: `ORDER BY RANDOM() LIMIT N` を使って1〜2回のクエリに削減
  - 期待される効果: 100〜1000倍の高速化（数分〜数十分 → 数秒〜数十秒）
  - 対象ファイル: `internal/usecase/seed/create_heavy_user.go` の `getRandomEpisodes` メソッド
  - **想定ファイル数**: 約 1 ファイル
  - **想定行数**: 約 30 行（メソッド全体の書き換え）

- [x] **エピソード生成ロジックの修正**
  - `GenerateEpisodeParamsForWork` 関数のエピソード生成ロジックを修正
  - 修正内容:
    - `Number` フィールドに「第1話」「第2話」といった形式の文字列を設定（数値のみではなく）
    - `Title` フィールドにはサブタイトルのみを設定（「第N話」を含めない）
  - 対象ファイル: `internal/usecase/seed/create_episode.go` の `GenerateEpisodeParamsForWork` 関数
  - **想定ファイル数**: 約 1 ファイル
  - **想定行数**: 約 20 行（関数の修正）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **複雑な関連データ**: キャスト、スタッフ、キャラクター、放送情報など（将来的に必要に応じて追加）
- **WorkRecord**: 作品全体のレビュー（EpisodeRecordのみ生成）
- **コメント、いいね**: コメントやいいね機能（優先度が低い）
- **通知**: 通知データの生成（複雑なため後回し）
- **コレクション**: ユーザーコレクション（優先度が低い）
- **フォーラム**: フォーラム投稿（使用頻度が低い）
- **本番環境での実行**: seedコマンドは開発環境専用（本番では実行しない）

## 参考資料

- [gofakeit - Go Fake Data Generator](https://github.com/brianvoe/gofakeit)
- [progressbar - Terminal Progress Bar for Go](https://github.com/schollz/progressbar)
- [AWS SDK for Go v2 - S3 Service](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3)
- [Cloudflare R2 - S3 API Compatibility](https://developers.cloudflare.com/r2/api/s3/api/)
- [Shrine - File Attachment toolkit for Ruby](https://shrinerb.com/)
- [Rails調査レポート（テーブル構造）](./.claude/designs/1_doing/rails-database-schema-report.md)（上記の調査結果を参照）

---

## 実装時の注意点

### Rails互換性を保つために

1. **bcryptのコスト**: Rails版と同じコスト（デフォルト10）を使用
2. **Shrineのimage_data形式**: JSON形式を正確に再現
3. **タイムスタンプ**: created_at, updated_at を適切に設定
4. **外部キー制約**: 参照整合性を必ず守る
5. **カウンターカラム**: counter_cultureで管理されるカウンターを正しく更新
6. **Enum値**: Rails版のenum定義と一致させる（rating_state: bad/average/good/great など）

### パフォーマンス最適化のポイント

1. **バルクインサート**: 1,000件ごとにコミット
2. **トランザクション**: 適切なトランザクション境界を設定
3. **インデックス**: 既存のインデックスを活用
4. **進捗表示**: ユーザーに進捗を見せることで安心感を与える
5. **並列処理**: 可能な箇所はgoroutineで並列化（ただし複雑にしすぎない）
