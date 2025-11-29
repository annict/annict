package viewmodel

import (
	"github.com/annict/annict/internal/image"
	"github.com/annict/annict/internal/model"
)

// Work はテンプレート表示用の作品データです
type Work struct {
	ID            int64
	Title         string
	TitleEn       string // 英語タイトルも保持
	ImageURL      string // デフォルトの画像URL (280px, jpg)
	ImageDataJSON string // work_imagesテーブルのimage_data (JSON)
	WatchersCount int32
	SeasonYear    *int32
	SeasonName    *string
	SeasonNumber  *int32        // シーズン番号も保持（翻訳キー用）
	Casts         []Cast        // キャスト情報
	Staffs        []Staff       // スタッフ情報
	imageHelper   *image.Helper // 画像URL生成用のヘルパー
}

// GetImageURL は指定されたサイズとフォーマットで画像URLを取得します
func (w *Work) GetImageURL(width int, format string) string {
	if w.imageHelper == nil {
		return ""
	}
	return w.imageHelper.GetWorkImageURL(w.ImageDataJSON, width, format)
}

// GetSrcSet は1xと2xの画像URLセットを取得します
func (w *Work) GetSrcSet(width int, format string) string {
	if w.imageHelper == nil {
		return ""
	}
	originalURL := w.imageHelper.ExtractImageURL(w.ImageDataJSON)
	return w.imageHelper.GetSrcSet(originalURL, width, format)
}

// Cast はキャスト情報を表します
type Cast struct {
	ID              int64
	Name            string
	NameEn          string
	CharacterName   string
	CharacterNameEn string
	PersonName      string
	PersonNameEn    string
}

// Staff はスタッフ情報を表します
type Staff struct {
	ID          int64
	Name        string
	NameEn      string
	Role        string
	RoleOther   string
	RoleOtherEn string
}

// NewWorksFromModelDetails は model.WorkWithDetails から viewmodel.Work に変換します
func NewWorksFromModelDetails(details []model.WorkWithDetails, helper *image.Helper) []Work {
	works := make([]Work, len(details))
	for i, detail := range details {
		works[i] = NewWorkFromModelDetail(detail, helper)
	}
	return works
}

// NewWorkFromModelDetail は model.WorkWithDetails から viewmodel.Work に変換します
func NewWorkFromModelDetail(detail model.WorkWithDetails, helper *image.Helper) Work {
	// imgproxy用の画像URL生成（280pxサイズ、jpg形式）
	imageURL := ""
	if helper != nil {
		imageURL = helper.GetWorkImageURL(detail.Work.ImageData, 280, "jpg")
	}

	work := Work{
		ID:            detail.Work.ID,
		Title:         detail.Work.Title,
		TitleEn:       detail.Work.TitleEn,
		ImageURL:      imageURL,
		ImageDataJSON: detail.Work.ImageData,
		WatchersCount: detail.Work.WatchersCount,
		imageHelper:   helper,
	}

	// タイトルのフォールバック処理
	if work.Title == "" && detail.Work.TitleEn != "" {
		work.Title = detail.Work.TitleEn
	}

	// シーズン情報の変換
	if detail.Work.SeasonYear != nil {
		work.SeasonYear = detail.Work.SeasonYear
	}

	if detail.Work.SeasonName != nil {
		work.SeasonNumber = detail.Work.SeasonName
		// 日本語のシーズン名に変換
		seasonNames := []string{"冬", "春", "夏", "秋"}
		// seasonNamesは固定長（4要素）のため、int32への変換は安全
		seasonNamesLen := int32(len(seasonNames)) // #nosec G115
		if *detail.Work.SeasonName >= 0 && *detail.Work.SeasonName < seasonNamesLen {
			seasonStr := seasonNames[*detail.Work.SeasonName]
			work.SeasonName = &seasonStr
		}
	}

	// キャストとスタッフの変換
	work.Casts = make([]Cast, len(detail.Casts))
	for i, cast := range detail.Casts {
		work.Casts[i] = Cast{
			ID:              cast.ID,
			Name:            cast.Name,
			NameEn:          cast.NameEn,
			CharacterName:   cast.CharacterName,
			CharacterNameEn: cast.CharacterNameEn,
			PersonName:      cast.PersonName,
			PersonNameEn:    cast.PersonNameEn,
		}
	}

	work.Staffs = make([]Staff, len(detail.Staffs))
	for i, staff := range detail.Staffs {
		work.Staffs[i] = Staff{
			ID:          staff.ID,
			Name:        staff.Name,
			NameEn:      staff.NameEn,
			Role:        staff.Role,
			RoleOther:   staff.RoleOther,
			RoleOtherEn: staff.RoleOtherEn,
		}
	}

	return work
}
