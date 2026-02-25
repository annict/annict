# シードデータ改善 設計書

## 概要

開発環境で使用するシードデータの品質を向上させるための修正を行います。具体的には、画像の比率調整、プロフィール画像の追加、エピソード記録の日本語化を実施します。

**目的**:

- 本番環境のデータに近いリアルなシードデータを生成することで、開発時の動作確認を正確に行えるようにする
- ユーザープロフィールの表示確認を可能にするため、プロフィール画像を追加する
- 視聴記録の表示を自然にするため、日本語の感想文を生成する
- 作品画像の表示比率を正しくするため、縦長の画像を生成する

**背景**:

- 現在のシードデータは以下の問題があり、開発時の動作確認が不十分
  - 作品画像が正方形（1:1）で生成されているが、実際の UI は縦長（3:4）を想定している
  - プロフィール画像が設定されておらず、アバター表示の確認ができない
  - エピソード記録の感想が英語で生成されており、日本語 UI との整合性がとれていない

## 要件

### 機能要件

#### 1. 作品画像の比率を縦長（3:4）に変更

- 作品画像の生成時に、幅:高さの比率を 3:4 にする
- Rails 版の仕様（`app/assets/stylesheets/components/_work-picture.scss`の`aspect-ratio: 3 / 4`）に合わせる
- 画像サイズ: 600x800 px（幅 600px、高さ 800px）
  - 理由: 500x500 より大きくすることで、高解像度ディスプレイでも鮮明に表示できる
  - imgproxy によるリサイズを前提としているため、元画像は少し大きめにする

#### 2. プロフィール画像を生成して設定

- ユーザー作成時に、プロフィール画像を生成して`profiles.image_data`カラムに設定する
- プロフィール画像の比率: 1:1（正方形）
  - Rails 版の仕様（`app/helpers/image_helper.rb`の`ann_avatar_image_url`で`ratio: "1:1"`）に合わせる
- 画像サイズ: 400x400 px
  - 理由: アバター画像として十分な解像度を確保しつつ、ファイルサイズを抑える
- Shrine 形式の JSON 構造で`image_data`カラムに保存
  - パス: `shrine/profile/{profile_id}/image/master-{uuid}.png`
  - Shrine の pretty_location プラグインの仕様に合わせる

#### 3. エピソード記録の本文を日本語にする

- 視聴記録の感想文を英語から日本語に変更する
- 日本語の自然な感想文を生成する
  - バリエーション: 複数のパターンを用意してランダムに選択
  - 長さ: 50〜200 文字程度（自然な長さ）
  - 内容: アニメ視聴の感想として自然な内容

### 非機能要件

#### パフォーマンス

- シードデータ生成時間は現状と同等を維持（プロフィール画像生成による若干の増加は許容）
- 画像生成は既存の Worker Pool パターンを活用して並列処理

#### 保守性

- 既存のシードデータ生成 Usecase の構造を維持
- 画像生成ロジックは`internal/seed/image.go`に集約
- 日本語テキスト生成ロジックは`internal/seed/data.go`に集約

#### データ整合性

- Shrine の image_data JSON フォーマットとの互換性を維持
- プロフィール画像は必ず Cloudflare R2 にアップロード（ローカル環境でも R2 を使用）
- 既存の作品画像クリーンアップ処理を参考に、プロフィール画像もクリーンアップ対象に含める

## 設計

### 技術スタック

- **言語**: Go 1.25.1
- **画像処理**: Go 標準ライブラリ（`image`, `image/png`）
- **ストレージ**: Cloudflare R2（S3 互換 API）
- **テストデータ生成**: brianvoe/gofakeit/v6（日本語テキスト生成には使用せず、独自実装）

### アーキテクチャ

#### 1. 画像生成の変更（`internal/seed/image.go`）

**現状**:

```go
const (
    ImageWidth = 500   // 正方形
    ImageHeight = 500
)
```

**変更後**:

```go
const (
    // 作品画像のサイズ（3:4の縦長）
    WorkImageWidth = 600
    WorkImageHeight = 800

    // プロフィール画像のサイズ（1:1の正方形）
    ProfileImageWidth = 400
    ProfileImageHeight = 400
)
```

新しい関数を追加:

- `GenerateRandomWorkImage(workID int64) (*RandomImage, error)`: 作品画像生成（600x800px）
- `GenerateRandomProfileImage(profileID int64) (*RandomImage, error)`: プロフィール画像生成（400x400px）

既存の`GenerateRandomImage`は非推奨として残し、内部で`GenerateRandomWorkImage`を呼び出すようにする（後方互換性のため）。

#### 2. プロフィール画像生成 Usecase の追加（`internal/usecase/seed/create_profile_image.go`）

**新規ファイル**: `internal/usecase/seed/create_profile_image.go`

```go
package seed

type CreateProfileImageParams struct {
    ProfileID int64
    UserID    int64
}

type CreateProfileImageResult struct {
    ProfileImageID int64
    ImagePath      string
}

type CreateProfileImageUsecase struct {
    db              *sql.DB
    queries         *query.Queries
    endpoint        string
    accessKeyID     string
    secretAccessKey string
    region          string
    bucketName      string
}
```

主な処理:

1. `GenerateRandomProfileImage(profileID)`でプロフィール画像を生成
2. Cloudflare R2 にアップロード
3. Shrine 形式の JSON を生成
4. `profiles.image_data`カラムを更新

#### 3. 日本語感想文生成（`internal/seed/data.go`）

**新規関数**: `GenerateJapaneseEpisodeRecordBody(rnd *rand.Rand) string`

感想文のパターン例:

```go
var episodeBodyTemplates = []string{
    "今回の話はとても面白かったです！{character}の活躍が印象的でした。",
    "{scene}のシーンに感動しました。次回も楽しみです。",
    "展開が予想外で驚きました。{emotion}な気持ちになりました。",
    // ... 20〜30パターン用意
}

var characterWords = []string{"主人公", "ヒロイン", "敵キャラ", "サブキャラ"}
var sceneWords = []string{"戦闘", "告白", "別れ", "再会"}
var emotionWords = []string{"感動的", "切ない", "嬉しい", "驚き"}
```

プレースホルダーをランダムな単語で置換して、バリエーション豊かな感想文を生成する。

#### 4. データベース設計

**更新対象テーブル**: `profiles`

```sql
-- profilesテーブルのimage_dataカラムを更新
UPDATE profiles
SET image_data = '{Shrine JSON}',
    updated_at = NOW()
WHERE id = $1
```

Shrine JSON 形式:

```json
{
  "master": {
    "id": "profile/12345/image/master-uuid.png",
    "storage": "store",
    "metadata": {
      "filename": "master-uuid.png",
      "size": 123456,
      "mime_type": "image/png",
      "width": 400,
      "height": 400
    }
  }
}
```

### テスト戦略

#### 単体テスト

- `internal/seed/image_test.go`: 画像生成ロジックのテスト
  - `TestGenerateRandomWorkImage`: 作品画像が 600x800px で生成されることを確認
  - `TestGenerateRandomProfileImage`: プロフィール画像が 400x400px で生成されることを確認
- `internal/seed/data_test.go`: 日本語テキスト生成のテスト
  - `TestGenerateJapaneseEpisodeRecordBody`: 日本語の感想文が生成されることを確認
- `internal/usecase/seed/create_profile_image_test.go`: プロフィール画像 UseCase

#### 統合テスト

- `cmd/seed/main.go`を実行してシードデータ生成が正常に完了することを確認
- 生成されたデータをブラウザで確認
  - 作品画像が縦長で表示されること
  - プロフィール画像が正方形で表示されること
  - エピソード記録の感想が日本語で表示されること

### 実装方針

#### フェーズごとの実装順序

1. **フェーズ 1**: 画像生成ロジックの変更（`internal/seed/image.go`）
2. **フェーズ 2**: 日本語感想文生成ロジックの追加（`internal/seed/data.go`）
3. **フェーズ 3**: プロフィール画像生成 Usecase の実装（`internal/usecase/seed/create_profile_image.go`）
4. **フェーズ 4**: シードスクリプトの修正（`cmd/seed/main.go`）
5. **フェーズ 5**: 動作確認とテスト

#### 既存コードとの互換性

- 既存の`GenerateRandomImage`関数は非推奨として残し、内部で`GenerateRandomWorkImage`を呼び出す
- テストコードで既存の関数を使用している箇所があれば、新しい関数に移行する

## タスクリスト

### フェーズ 1: 画像生成ロジックの変更

- [x] **1-1**: `internal/seed/image.go`の修正
  - `WorkImageWidth`, `WorkImageHeight`, `ProfileImageWidth`, `ProfileImageHeight`定数を追加
  - `GenerateRandomWorkImage(workID)`関数を追加（600x800px）
  - `GenerateRandomProfileImage(profileID)`関数を追加（400x400px）
  - 既存の`GenerateRandomImage`を非推奨とし、内部で`GenerateRandomWorkImage`を呼び出すように変更
  - テストを追加（`TestGenerateRandomWorkImage`, `TestGenerateRandomProfileImage`）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

### フェーズ 2: 日本語感想文生成ロジックの追加

- [x] **2-1**: `internal/seed/data.go`の修正
  - `GenerateJapaneseEpisodeRecordBody(rnd *rand.Rand) string`関数を追加
  - 感想文テンプレート（20〜30 パターン）を定義
  - プレースホルダー置換ロジックを実装
  - テストを追加（`TestGenerateJapaneseEpisodeRecordBody`）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 250 行（実装 150 行 + テスト 100 行）

### フェーズ 3: プロフィール画像生成 Usecase の実装

- [x] **3-1**: `internal/usecase/seed/create_profile_image.go`の追加
  - `CreateProfileImageUsecase`構造体を定義
  - `ExecuteBatch`メソッドを実装（作品画像と同様のパターン）
  - Worker Pool パターンで並列処理
  - テストを追加（`TestCreateProfileImageUsecase_ExecuteBatch`）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 300 行（実装 200 行 + テスト 100 行）

- [x] **3-2**: sqlc クエリの追加（`internal/query/queries/profiles.sql`）
  - `UpdateProfileImageData`クエリを追加
  - sqlc コード生成（`make sqlc-generate`）
  - **想定ファイル数**: 約 2 ファイル（SQL クエリ 1 + 生成コード 1）
  - **想定行数**: 約 50 行（SQL クエリ 10 行 + 生成コード 40 行）

### フェーズ 4: シードスクリプトの修正

- [x] **4-1**: `cmd/seed/main.go`の修正
  - `generateEpisodeRecords`関数を修正
    - `gofakeit.Sentence(10)`を`seed.GenerateJapaneseEpisodeRecordBody(rnd)`に変更
  - `generateProfileImages`関数を追加（新規フェーズとして）
    - フェーズ 1 とフェーズ 2 の間に挿入（ユーザー生成直後）
    - 全ユーザーのプロフィールにプロフィール画像を設定
  - `cleanupProfileImages`関数を追加（クリーンアップ処理）
  - **想定ファイル数**: 約 1 ファイル（実装のみ）
  - **想定行数**: 約 150 行（実装 150 行）

- [x] **4-2**: `internal/usecase/seed/create_work_image.go`の修正
  - `GenerateRandomImage`の呼び出しを`GenerateRandomWorkImage`に変更
  - **想定ファイル数**: 約 1 ファイル（実装のみ）
  - **想定行数**: 約 5 行（実装 5 行）

### フェーズ 5: 動作確認とクリーンアップ

- [x] **5-1**: シードデータ生成の動作確認
  - `make db-migrate`でマイグレーション実行
  - `go run cmd/seed/main.go`でシードデータ生成
  - 生成時間を計測（現状と比較）
  - ブラウザで表示確認
    - 作品画像が縦長（3:4）で表示されること
    - プロフィール画像が正方形（1:1）で表示されること
    - エピソード記録の感想が日本語で表示されること
  - **想定ファイル数**: 約 0 ファイル（確認作業のみ）
  - **想定行数**: 約 0 行（確認作業のみ）

- [x] **5-2**: 既存テストの修正とクリーンアップ
  - 既存のテストで`GenerateRandomImage`を使用している箇所を確認
  - 必要に応じて新しい関数に移行
  - 非推奨警告のドキュメント追加
  - **想定ファイル数**: 約 2 ファイル（テスト修正）
  - **想定行数**: 約 50 行（テスト修正 50 行）

### フェーズ 6: パフォーマンス最適化

- [x] **6-1**: プロフィール画像生成の高速化（単色画像の使用）
  - `internal/seed/image.go`の`generateImage`関数を修正
  - グラデーション計算を削除し、ランダムな単色で塗りつぶす
  - ドット模様（`addRandomPattern`）の削除
  - テストデータは視覚的なバリエーションよりも速度を優先
  - **効果見込み**: 画像生成時間を 90%以上削減
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 30 行 + テスト 20 行）

- [x] **6-2**: パフォーマンス計測と比較
  - 最適化前後の画像生成時間を計測
  - 30,000 件のプロフィール画像生成時間を比較
  - ブラウザで表示確認（単色画像でも問題ないことを確認）
  - **想定ファイル数**: 約 0 ファイル（確認作業のみ）
  - **想定行数**: 約 0 行（確認作業のみ）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **背景画像の生成**: プロフィールには背景画像（`background_image_data`）も設定できるが、今回はアバター画像のみ対象
- **プロフィール画像の複数バージョン**: Shrine は複数バージョン（サムネイルなど）を管理できるが、今回は master バージョンのみ
- **既存のプロフィール画像の更新**: 既存のシードデータがあれば削除してから再生成する（クリーンアップ処理で対応）
- **作品画像のバリエーション**: 現在はグラデーション+ドット模様のみだが、他のパターンは今回追加しない

## 参考資料

- [Rails 版の画像比率定義](file:///workspace/rails/app/assets/stylesheets/components/_work-picture.scss) - `aspect-ratio: 3 / 4`
- [Rails 版の画像ヘルパー](file:///workspace/rails/app/helpers/image_helper.rb) - `ann_work_image_url`, `ann_avatar_image_url`
- [Shrine Gem Documentation](https://shrinerb.com/) - Shrine 形式の JSON 構造
- [gofakeit Documentation](https://github.com/brianvoe/gofakeit) - テストデータ生成ライブラリ
