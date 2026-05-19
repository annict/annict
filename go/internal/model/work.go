package model

import "time"

// Work is the domain entity for an anime work, kept page-independent and generic.
// It belongs to the Domain / Infrastructure layer and must not depend on the Presentation layer.
//
// [Ja] Work は作品のドメインエンティティ (ページに依存しない汎用的な構造)。
// Domain / Infrastructure 層に属し、Presentation 層に依存しない。
type Work struct {
	ID                  WorkID
	Title               string
	TitleEn             string
	TitleKana           *string
	RecommendedImageURL string
	ImageData           string // JSON payload from the work_images.image_data column. [Ja] work_images テーブルの image_data カラム (JSON)
	WatchersCount       int32
	SeasonYear          *int32
	SeasonName          *int32 // Season number: 1=winter, 2=spring, 3=summer, 4=autumn. [Ja] シーズン番号 (1=冬、2=春、3=夏、4=秋)
	CreatedAt           time.Time

	// Related entities. Set only when the caller has explicitly loaded them; nil by default.
	// [Ja] 関連エンティティ。明示的にロードした場合のみセットされ、通常は nil。
	Casts  []*Cast
	Staffs []*Staff
}

// DBWorkListItem is the per-row data shape for the work list on the Annict DB admin screen.
// [Ja] DBWorkListItem は Annict DB 管理画面の作品一覧で 1 行ごとに表示するデータ構造。
type DBWorkListItem struct {
	ID            WorkID
	Title         string
	SeasonYear    *int32
	SeasonName    *int32
	WatchersCount int32
	Status        string
	HasImage      bool
}
