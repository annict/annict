package viewmodel

import (
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/image"
	"github.com/annict/annict/internal/model"
)

func TestNewWorkFromModelDetail(t *testing.T) {
	t.Parallel()

	// テスト用の設定とimage.Helperを作成
	cfg := &config.Config{
		ImgproxyEndpoint: "http://localhost:18080",
		ImgproxyKey:      "test-key",
		ImgproxySalt:     "test-salt",
	}
	helper := image.NewHelper(cfg)

	t.Run("基本的な変換", func(t *testing.T) {
		t.Parallel()

		seasonYear := int32(2024)
		seasonName := int32(1) // 春

		detail := model.WorkWithDetails{
			Work: model.Work{
				ID:            123,
				Title:         "テストアニメ",
				TitleEn:       "Test Anime",
				ImageData:     `{"id":"test123","storage":"store","metadata":{"filename":"test.jpg","size":12345,"mime_type":"image/jpeg"}}`,
				WatchersCount: 100,
				SeasonYear:    &seasonYear,
				SeasonName:    &seasonName,
			},
			Casts: []model.Cast{
				{
					ID:              1,
					WorkID:          123,
					Name:            "キャラクター名",
					NameEn:          "Character Name",
					CharacterName:   "田中太郎",
					CharacterNameEn: "Taro Tanaka",
					PersonName:      "声優名",
					PersonNameEn:    "Voice Actor",
				},
			},
			Staffs: []model.Staff{
				{
					ID:          2,
					WorkID:      123,
					Name:        "スタッフ名",
					NameEn:      "Staff Name",
					Role:        "director",
					RoleOther:   "監督",
					RoleOtherEn: "Director",
				},
			},
		}

		work := NewWorkFromModelDetail(detail, helper)

		// 基本フィールドの検証
		if work.ID != 123 {
			t.Errorf("ID: got %d, want 123", work.ID)
		}
		if work.Title != "テストアニメ" {
			t.Errorf("Title: got %s, want テストアニメ", work.Title)
		}
		if work.TitleEn != "Test Anime" {
			t.Errorf("TitleEn: got %s, want Test Anime", work.TitleEn)
		}
		if work.WatchersCount != 100 {
			t.Errorf("WatchersCount: got %d, want 100", work.WatchersCount)
		}

		// ImageDataJSONが正しく設定されていることを確認
		if work.ImageDataJSON != detail.Work.ImageData {
			t.Errorf("ImageDataJSON: got %s, want %s", work.ImageDataJSON, detail.Work.ImageData)
		}

		// imageHelperが設定されていることを確認
		if work.imageHelper == nil {
			t.Error("imageHelper should not be nil")
		}

		// シーズン情報の検証
		if work.SeasonYear == nil || *work.SeasonYear != 2024 {
			t.Errorf("SeasonYear: got %v, want 2024", work.SeasonYear)
		}
		if work.SeasonNumber == nil || *work.SeasonNumber != 1 {
			t.Errorf("SeasonNumber: got %v, want 1", work.SeasonNumber)
		}
		if work.SeasonName == nil || *work.SeasonName != "春" {
			t.Errorf("SeasonName: got %v, want 春", work.SeasonName)
		}

		// キャストの検証
		if len(work.Casts) != 1 {
			t.Fatalf("Casts length: got %d, want 1", len(work.Casts))
		}
		if work.Casts[0].ID != 1 {
			t.Errorf("Cast ID: got %d, want 1", work.Casts[0].ID)
		}
		if work.Casts[0].Name != "キャラクター名" {
			t.Errorf("Cast Name: got %s, want キャラクター名", work.Casts[0].Name)
		}

		// スタッフの検証
		if len(work.Staffs) != 1 {
			t.Fatalf("Staffs length: got %d, want 1", len(work.Staffs))
		}
		if work.Staffs[0].ID != 2 {
			t.Errorf("Staff ID: got %d, want 2", work.Staffs[0].ID)
		}
		if work.Staffs[0].Name != "スタッフ名" {
			t.Errorf("Staff Name: got %s, want スタッフ名", work.Staffs[0].Name)
		}
	})

	t.Run("タイトルのフォールバック", func(t *testing.T) {
		t.Parallel()

		detail := model.WorkWithDetails{
			Work: model.Work{
				ID:      123,
				Title:   "", // 日本語タイトルが空
				TitleEn: "Fallback Title",
			},
		}

		work := NewWorkFromModelDetail(detail, helper)

		// 英語タイトルがフォールバックされることを確認
		if work.Title != "Fallback Title" {
			t.Errorf("Title should fallback to TitleEn: got %s, want Fallback Title", work.Title)
		}
	})

	t.Run("シーズン情報が nil の場合", func(t *testing.T) {
		t.Parallel()

		detail := model.WorkWithDetails{
			Work: model.Work{
				ID:         123,
				Title:      "テストアニメ",
				SeasonYear: nil,
				SeasonName: nil,
			},
		}

		work := NewWorkFromModelDetail(detail, helper)

		if work.SeasonYear != nil {
			t.Errorf("SeasonYear should be nil: got %v", work.SeasonYear)
		}
		if work.SeasonNumber != nil {
			t.Errorf("SeasonNumber should be nil: got %v", work.SeasonNumber)
		}
		if work.SeasonName != nil {
			t.Errorf("SeasonName should be nil: got %v", work.SeasonName)
		}
	})

	t.Run("キャストとスタッフが空の場合", func(t *testing.T) {
		t.Parallel()

		detail := model.WorkWithDetails{
			Work: model.Work{
				ID:    123,
				Title: "テストアニメ",
			},
			Casts:  []model.Cast{},
			Staffs: []model.Staff{},
		}

		work := NewWorkFromModelDetail(detail, helper)

		if len(work.Casts) != 0 {
			t.Errorf("Casts should be empty: got %d", len(work.Casts))
		}
		if len(work.Staffs) != 0 {
			t.Errorf("Staffs should be empty: got %d", len(work.Staffs))
		}
	})

	t.Run("image.Helper が nil の場合", func(t *testing.T) {
		t.Parallel()

		detail := model.WorkWithDetails{
			Work: model.Work{
				ID:        123,
				Title:     "テストアニメ",
				ImageData: `{"id":"test123"}`,
			},
		}

		work := NewWorkFromModelDetail(detail, nil)

		// ImageURLが空になることを確認
		if work.ImageURL != "" {
			t.Errorf("ImageURL should be empty when helper is nil: got %s", work.ImageURL)
		}
	})

	t.Run("シーズン番号の変換（冬=0、春=1、夏=2、秋=3）", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			seasonNumber int32
			expectedName string
		}{
			{0, "冬"},
			{1, "春"},
			{2, "夏"},
			{3, "秋"},
		}

		for _, tc := range testCases {
			seasonName := tc.seasonNumber
			detail := model.WorkWithDetails{
				Work: model.Work{
					ID:         123,
					Title:      "テストアニメ",
					SeasonName: &seasonName,
				},
			}

			work := NewWorkFromModelDetail(detail, helper)

			if work.SeasonName == nil || *work.SeasonName != tc.expectedName {
				t.Errorf("SeasonNumber %d: got %v, want %s", tc.seasonNumber, work.SeasonName, tc.expectedName)
			}
		}
	})
}

func TestNewWorksFromModelDetails(t *testing.T) {
	t.Parallel()

	// テスト用の設定とimage.Helperを作成
	cfg := &config.Config{
		ImgproxyEndpoint: "http://localhost:18080",
		ImgproxyKey:      "test-key",
		ImgproxySalt:     "test-salt",
	}
	helper := image.NewHelper(cfg)

	t.Run("複数の作品を変換", func(t *testing.T) {
		t.Parallel()

		seasonYear1 := int32(2024)
		seasonName1 := int32(1)
		seasonYear2 := int32(2023)
		seasonName2 := int32(3)

		details := []model.WorkWithDetails{
			{
				Work: model.Work{
					ID:         1,
					Title:      "アニメ1",
					SeasonYear: &seasonYear1,
					SeasonName: &seasonName1,
				},
			},
			{
				Work: model.Work{
					ID:         2,
					Title:      "アニメ2",
					SeasonYear: &seasonYear2,
					SeasonName: &seasonName2,
				},
			},
		}

		works := NewWorksFromModelDetails(details, helper)

		if len(works) != 2 {
			t.Fatalf("works length: got %d, want 2", len(works))
		}

		// 1つ目の作品の検証
		if works[0].ID != 1 {
			t.Errorf("works[0].ID: got %d, want 1", works[0].ID)
		}
		if works[0].Title != "アニメ1" {
			t.Errorf("works[0].Title: got %s, want アニメ1", works[0].Title)
		}

		// 2つ目の作品の検証
		if works[1].ID != 2 {
			t.Errorf("works[1].ID: got %d, want 2", works[1].ID)
		}
		if works[1].Title != "アニメ2" {
			t.Errorf("works[1].Title: got %s, want アニメ2", works[1].Title)
		}
	})

	t.Run("空のスライス", func(t *testing.T) {
		t.Parallel()

		details := []model.WorkWithDetails{}
		works := NewWorksFromModelDetails(details, helper)

		if len(works) != 0 {
			t.Errorf("works should be empty: got %d", len(works))
		}
	})
}
