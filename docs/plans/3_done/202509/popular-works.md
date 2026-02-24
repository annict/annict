# 人気アニメページ実装

## 概要

最初の実装ページとして、視聴者数の多い作品を一覧表示する「人気アニメ」ページを実装。
このページの実装を通じて、Go プロジェクトの全体的なアーキテクチャとフローを確立。

## 実装内容

### フェーズ 2: 最初のページ実装（人気アニメ）で全体の流れを掴む

- [x] **sqlc の導入とセットアップ**
  - [x] sqlc.yaml の設定（Rails DB スキーマ参照）
  - [x] 作品取得クエリの作成（queries/works.sql）
  - [x] タイプセーフなコード生成

- [x] **人気アニメページの実装（/works/popular）**
  - [x] watchers_count でソートした works テーブルから取得
  - [x] work_images テーブルとの JOIN
  - [x] HTML テンプレート（html/template）で表示
  - [x] go.example.dev で動作確認

- [x] **アーキテクチャ基盤の構築**
  - [x] ViewModel パターンの実装（internal/viewmodel）
  - [x] テンプレートヘルパー関数の追加
  - [x] ベースレイアウトテンプレートの作成

- [x] **画像配信システムの実装**
  - [x] imgproxy 連携（署名付き URL 生成）
  - [x] S3 互換ストレージ対応（MinIO/Cloudflare R2）
  - [x] 画像ヘルパー関数の実装

- [x] **国際化（i18n）対応**
  - [x] go-i18n ライブラリの導入
  - [x] 日本語・英語のロケールファイル作成
  - [x] シーズン名の翻訳機能

- [x] **テスト基盤の構築**
  - [x] ハンドラーの単体テスト
  - [x] モックリポジトリパターンの実装
  - [x] テンプレートレンダリングテスト

- [x] **フロントエンド基盤の構築**
  - [x] Vite + TypeScript + Tailwind CSS 設定
  - [x] PostCSS 設定
  - [x] ホットリロード対応（air）
  - [x] npm init と package.json 作成
  - [x] Tailwind 設定（必要最小限）
  - [x] ビルドスクリプト作成

- [x] **運用環境の準備**
  - [x] Cloudflare Tunnel 設定
  - [x] go.example.dev の設定
  - [x] HTTPS アクセスの確認

## 技術スタック

### バックエンド
- **sqlc**: SQL クエリから Go コードを生成
- **html/template**: Go 標準テンプレートエンジン
- **go-i18n**: 国際化ライブラリ
- **imgproxy**: 画像リサイズ・最適化プロキシ

### フロントエンド
- **Vite**: 高速フロントエンドビルドツール
- **TypeScript**: 型安全な JavaScript
- **Tailwind CSS**: ユーティリティファースト CSS

## アーキテクチャパターン

### レイヤー構成

```
Handler (HTTP層)
    ↓
Repository (データアクセス層、sqlcが生成)
    ↓
ViewModel (プレゼンテーション層)
    ↓
Template (表示層)
```

### ファイル構成

```
internal/
├── handler/
│   └── popular_works.go          # HTTPハンドラー
├── repository/
│   ├── queries/
│   │   └── works.sql             # SQLクエリ定義
│   └── sqlc/                     # sqlcが生成したコード
│       ├── db.go
│       ├── models.go
│       └── works.sql.go
├── viewmodel/
│   └── work.go                   # ViewModel変換ロジック
└── templates/
    ├── layouts/
    │   └── base.html             # ベースレイアウト
    └── works/
        └── popular.html          # 人気アニメページ
```

## 主要な実装

### SQL クエリ（queries/works.sql）

```sql
-- name: GetPopularWorks :many
SELECT
    w.id,
    w.title,
    w.watchers_count,
    w.season_year,
    w.season_name,
    wi.image_data
FROM works w
LEFT JOIN work_images wi ON w.id = wi.work_id AND wi.is_deleted = false
WHERE w.deleted_at IS NULL
ORDER BY w.watchers_count DESC
LIMIT $1 OFFSET $2;
```

### ViewModel（internal/viewmodel/work.go）

```go
type Work struct {
    ID            int64
    Title         string
    ImageURL      string
    WatchersCount int32
    SeasonYear    *int32
    SeasonName    *string
}

func NewWorkFromPopularRow(ctx context.Context, cfg *config.Config, work repository.GetPopularWorksRow) Work {
    // リポジトリの型からViewModelに変換
    // 画像URLの生成なども含む
}
```

### ハンドラー（internal/handler/popular_works.go）

```go
func (h *Handler) PopularWorks(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // データ取得
    works, err := h.queries.GetPopularWorks(ctx, repository.GetPopularWorksParams{
        Limit:  100,
        Offset: 0,
    })

    // ViewModelに変換
    viewWorks := viewmodel.NewWorksFromPopularRows(ctx, h.cfg, works)

    // テンプレートレンダリング
    h.renderTemplate(w, r, "works/popular", map[string]interface{}{
        "Works": viewWorks,
    })
}
```

## 画像配信

### imgproxy 署名付き URL 生成

```go
func GenerateImageURL(cfg *config.Config, imagePath string, width int, height int) string {
    // S3プロトコルのURL生成
    sourceURL := fmt.Sprintf("s3://%s/shrine/%s", cfg.S3Bucket, imagePath)

    // imgproxyの署名付きURL生成
    path := fmt.Sprintf("/resize:fill:%d:%d/plain/%s", width, height, sourceURL)
    signature := generateHMAC(cfg.ImgproxyKey, cfg.ImgproxySalt, path)

    return fmt.Sprintf("%s/%s%s", cfg.ImgproxyURL, signature, path)
}
```

## 国際化

### ロケールファイル（internal/i18n/locales/ja.toml）

```toml
[popular_works_title]
description = "人気アニメページのタイトル"
other = "人気アニメ"

[season_spring]
description = "春シーズン"
other = "春"
```

### テンプレートでの使用

```html
<h1>{{call .T "popular_works_title"}}</h1>
<p>{{.Work.SeasonYear}}年 {{call .T (printf "season_%s" .Work.SeasonName)}}</p>
```

## テスト

### ハンドラーのテスト

```go
func TestPopularWorks(t *testing.T) {
    db, tx := testutil.SetupTestDB(t)

    // テストデータ作成
    workID := testutil.NewWorkBuilder(t, tx).
        WithTitle("テストアニメ").
        WithWatchersCount(1000).
        Build()

    // ハンドラー実行
    handler := &Handler{queries: repository.New(db).WithTx(tx)}
    req := httptest.NewRequest("GET", "/works/popular", nil)
    rr := httptest.NewRecorder()
    handler.PopularWorks(rr, req)

    // アサーション
    assert.Equal(t, http.StatusOK, rr.Code)
    assert.Contains(t, rr.Body.String(), "テストアニメ")
}
```

## 成果

- **アーキテクチャの確立**: Handler → Repository → ViewModel → Template の流れを確立
- **sqlc の活用**: タイプセーフなデータベースアクセスを実現
- **画像配信**: imgproxy を使った効率的な画像配信システム
- **国際化対応**: 日本語・英語の切り替えに対応
- **テスト基盤**: 実データベースを使ったテストパターンを確立
- **フロントエンド**: Vite + TypeScript + Tailwind CSS の開発環境構築

## 関連ドキュメント

- [プロジェクト全体の設計書](./go.md)
- [テストインフラ](./testing-infrastructure.md)
