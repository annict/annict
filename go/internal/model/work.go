package model

import "time"

// Work は作品のドメインエンティティです（ページに依存しない汎用的な構造）
// Domain/Infrastructure層に属し、Presentation層に依存しない
type Work struct {
	ID                  int64
	Title               string
	TitleEn             string
	TitleKana           *string
	RecommendedImageURL string
	ImageData           string // work_imagesテーブルのimage_data (JSON)
	WatchersCount       int32
	SeasonYear          *int32
	SeasonName          *int32 // シーズン番号（0=冬、1=春、2=夏、3=秋）
	CreatedAt           time.Time
}

// WorkWithDetails は作品の詳細情報を含むデータ構造です
type WorkWithDetails struct {
	Work   Work
	Casts  []Cast
	Staffs []Staff
}
