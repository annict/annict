package model

import "time"

// WorkStatus represents the lifecycle state of a work and mirrors the work_status PostgreSQL enum.
// [Ja] WorkStatus は作品のライフサイクル状態を表し、PostgreSQL の work_status enum と対応する。
type WorkStatus string

const (
	WorkStatusPublished WorkStatus = "published"
	WorkStatusArchived  WorkStatus = "archived"
	WorkStatusDeleted   WorkStatus = "deleted"
)

// String returns the textual representation of the status.
// [Ja] ステータスの文字列表現を返す。
func (s WorkStatus) String() string { return string(s) }

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
	ImageData           string // JSON payload from the work_images.image_data column; empty when no work_images row is joined. [Ja] work_images テーブルの image_data カラム (JSON)。LEFT JOIN で行が無い場合は空文字列。
	WatchersCount       int32
	SeasonYear          *int32
	SeasonName          *int32     // Season number: 1=winter, 2=spring, 3=summer, 4=autumn. [Ja] シーズン番号 (1=冬、2=春、3=夏、4=秋)
	Status              WorkStatus // Populated only by loaders that select the works.status column (e.g. ListForDB). [Ja] works.status カラムを select するロード経路 (例: ListForDB) でのみ値が入る。
	CreatedAt           time.Time

	// Related entities. Set only when the caller has explicitly loaded them; nil by default.
	// [Ja] 関連エンティティ。明示的にロードした場合のみセットされ、通常は nil。
	Casts  []*Cast
	Staffs []*Staff
}
