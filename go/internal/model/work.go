package model

import "time"

// Work は作品のドメインエンティティです（ページに依存しない汎用的な構造）
// Domain/Infrastructure層に属し、Presentation層に依存しない
type Work struct {
	ID                  WorkID
	Title               string
	TitleEn             string
	TitleKana           *string
	RecommendedImageURL string
	ImageData           string // work_imagesテーブルのimage_data (JSON)
	WatchersCount       int32
	SeasonYear          *int32
	SeasonName          *int32 // シーズン番号（1=冬、2=春、3=夏、4=秋）
	CreatedAt           time.Time

	// 関連エンティティ（必要な場合のみセットされる。通常は nil）
	Casts  []*Cast
	Staffs []*Staff
}

// DBWorkListItem はDB管理画面の作品一覧用のデータ構造です
type DBWorkListItem struct {
	ID            WorkID
	Title         string
	SeasonYear    *int32
	SeasonName    *int32
	WatchersCount int32
	Status        string
	HasImage      bool
}
