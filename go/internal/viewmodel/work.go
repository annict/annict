package viewmodel

import (
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/model"
)

// Work はテンプレート表示用の作品データです
type Work struct {
	ID            WorkID
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
	ID              CastID
	Name            string
	NameEn          string
	CharacterName   string
	CharacterNameEn string
	PersonName      string
	PersonNameEn    string
}

// Staff はスタッフ情報を表します
type Staff struct {
	ID          StaffID
	Name        string
	NameEn      string
	Role        string
	RoleOther   string
	RoleOtherEn string
}

// NewWorksFromModels は []*model.Work から []viewmodel.Work に変換します
func NewWorksFromModels(works []*model.Work, helper *image.Helper) []Work {
	result := make([]Work, len(works))
	for i, w := range works {
		result[i] = NewWorkFromModel(w, helper)
	}
	return result
}

// NewWorkFromModel は *model.Work から viewmodel.Work に変換します
func NewWorkFromModel(m *model.Work, helper *image.Helper) Work {
	// imgproxy用の画像URL生成（280pxサイズ、jpg形式）
	imageURL := ""
	if helper != nil {
		imageURL = helper.GetWorkImageURL(m.ImageData, 280, "jpg")
	}

	work := Work{
		ID:            WorkID(m.ID),
		Title:         m.Title,
		TitleEn:       m.TitleEn,
		ImageURL:      imageURL,
		ImageDataJSON: m.ImageData,
		WatchersCount: m.WatchersCount,
		imageHelper:   helper,
	}

	// タイトルのフォールバック処理
	if work.Title == "" && m.TitleEn != "" {
		work.Title = m.TitleEn
	}

	// シーズン情報の変換
	if m.SeasonYear != nil {
		work.SeasonYear = m.SeasonYear
	}

	if m.SeasonName != nil {
		work.SeasonNumber = m.SeasonName
		// 日本語のシーズン名に変換
		seasonNames := []string{"冬", "春", "夏", "秋"}
		// seasonNamesは固定長（4要素）のため、int32への変換は安全
		seasonNamesLen := int32(len(seasonNames)) // #nosec G115
		if *m.SeasonName >= 0 && *m.SeasonName < seasonNamesLen {
			seasonStr := seasonNames[*m.SeasonName]
			work.SeasonName = &seasonStr
		}
	}

	// キャストとスタッフの変換
	work.Casts = make([]Cast, len(m.Casts))
	for i, cast := range m.Casts {
		work.Casts[i] = Cast{
			ID:              CastID(cast.ID),
			Name:            cast.Name,
			NameEn:          cast.NameEn,
			CharacterName:   cast.CharacterName,
			CharacterNameEn: cast.CharacterNameEn,
			PersonName:      cast.PersonName,
			PersonNameEn:    cast.PersonNameEn,
		}
	}

	work.Staffs = make([]Staff, len(m.Staffs))
	for i, staff := range m.Staffs {
		work.Staffs[i] = Staff{
			ID:          StaffID(staff.ID),
			Name:        staff.Name,
			NameEn:      staff.NameEn,
			Role:        staff.Role,
			RoleOther:   staff.RoleOther,
			RoleOtherEn: staff.RoleOtherEn,
		}
	}

	return work
}
