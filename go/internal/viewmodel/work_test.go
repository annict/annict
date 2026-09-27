package viewmodel

import (
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/model"
)

func TestNewWorkFromModel(t *testing.T) {
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

		m := &model.Work{
			ID:            123,
			Title:         "テストアニメ",
			TitleEn:       "Test Anime",
			ImageData:     `{"id":"test123","storage":"store","metadata":{"filename":"test.jpg","size":12345,"mime_type":"image/jpeg"}}`,
			WatchersCount: 100,
			SeasonYear:    &seasonYear,
			SeasonName:    &seasonName,
			Casts: []*model.Cast{
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
			Staffs: []*model.Staff{
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

		work := NewWorkFromModel(m, helper)

		// 基本フィールドの検証
		if work.ID != 123 {
			t.Errorf("ID = %d、期待値 = 123", work.ID)
		}
		if work.Title != "テストアニメ" {
			t.Errorf("Title = %s、期待値 = テストアニメ", work.Title)
		}
		if work.TitleEn != "Test Anime" {
			t.Errorf("TitleEn = %s、期待値 = Test Anime", work.TitleEn)
		}
		if work.WatchersCount != 100 {
			t.Errorf("WatchersCount = %d、期待値 = 100", work.WatchersCount)
		}

		// ImageDataJSONが正しく設定されていることを確認
		if work.ImageDataJSON != m.ImageData {
			t.Errorf("ImageDataJSON = %s、期待値 = %s", work.ImageDataJSON, m.ImageData)
		}

		// imageHelperが設定されていることを確認
		if work.imageHelper == nil {
			t.Error("imageHelperがnilだった")
		}

		// シーズン情報の検証
		if work.SeasonYear == nil || *work.SeasonYear != 2024 {
			t.Errorf("SeasonYear = %v、期待値 = 2024", work.SeasonYear)
		}
		if work.SeasonNumber == nil || *work.SeasonNumber != 1 {
			t.Errorf("SeasonNumber = %v、期待値 = 1", work.SeasonNumber)
		}
		if work.SeasonName == nil || *work.SeasonName != "春" {
			t.Errorf("SeasonName = %v、期待値 = 春", work.SeasonName)
		}

		// キャストの検証
		if len(work.Casts) != 1 {
			t.Fatalf("Castsの件数 = %d、期待値 = 1", len(work.Casts))
		}
		if work.Casts[0].ID != 1 {
			t.Errorf("Cast ID = %d、期待値 = 1", work.Casts[0].ID)
		}
		if work.Casts[0].Name != "キャラクター名" {
			t.Errorf("Cast Name = %s、期待値 = キャラクター名", work.Casts[0].Name)
		}

		// スタッフの検証
		if len(work.Staffs) != 1 {
			t.Fatalf("Staffsの件数 = %d、期待値 = 1", len(work.Staffs))
		}
		if work.Staffs[0].ID != 2 {
			t.Errorf("Staff ID = %d、期待値 = 2", work.Staffs[0].ID)
		}
		if work.Staffs[0].Name != "スタッフ名" {
			t.Errorf("Staff Name = %s、期待値 = スタッフ名", work.Staffs[0].Name)
		}
	})

	t.Run("タイトルのフォールバック", func(t *testing.T) {
		t.Parallel()

		m := &model.Work{
			ID:      123,
			Title:   "", // 日本語タイトルが空
			TitleEn: "Fallback Title",
		}

		work := NewWorkFromModel(m, helper)

		// 英語タイトルがフォールバックされることを確認
		if work.Title != "Fallback Title" {
			t.Errorf("TitleEnへフォールバックしたTitle = %s、期待値 = Fallback Title", work.Title)
		}
	})

	t.Run("シーズン情報がnilの場合", func(t *testing.T) {
		t.Parallel()

		m := &model.Work{
			ID:         123,
			Title:      "テストアニメ",
			SeasonYear: nil,
			SeasonName: nil,
		}

		work := NewWorkFromModel(m, helper)

		if work.SeasonYear != nil {
			t.Errorf("SeasonYear = %v、期待値 = nil", work.SeasonYear)
		}
		if work.SeasonNumber != nil {
			t.Errorf("SeasonNumber = %v、期待値 = nil", work.SeasonNumber)
		}
		if work.SeasonName != nil {
			t.Errorf("SeasonName = %v、期待値 = nil", work.SeasonName)
		}
	})

	t.Run("キャストとスタッフが空の場合", func(t *testing.T) {
		t.Parallel()

		m := &model.Work{
			ID:     123,
			Title:  "テストアニメ",
			Casts:  []*model.Cast{},
			Staffs: []*model.Staff{},
		}

		work := NewWorkFromModel(m, helper)

		if len(work.Casts) != 0 {
			t.Errorf("Castsの件数 = %d、期待値 = 0", len(work.Casts))
		}
		if len(work.Staffs) != 0 {
			t.Errorf("Staffsの件数 = %d、期待値 = 0", len(work.Staffs))
		}
	})

	t.Run("image.Helperがnilの場合", func(t *testing.T) {
		t.Parallel()

		m := &model.Work{
			ID:        123,
			Title:     "テストアニメ",
			ImageData: `{"id":"test123"}`,
		}

		work := NewWorkFromModel(m, nil)

		// ImageURLが空になることを確認
		if work.ImageURL != "" {
			t.Errorf("helperがnilのときのImageURL = %s、期待値 = 空", work.ImageURL)
		}
	})

	t.Run("シーズン番号の変換 (冬=0、春=1、夏=2、秋=3)", func(t *testing.T) {
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
			m := &model.Work{
				ID:         123,
				Title:      "テストアニメ",
				SeasonName: &seasonName,
			}

			work := NewWorkFromModel(m, helper)

			if work.SeasonName == nil || *work.SeasonName != tc.expectedName {
				t.Errorf("SeasonNumber %d = %v、期待値 = %s", tc.seasonNumber, work.SeasonName, tc.expectedName)
			}
		}
	})
}

func TestNewWorksFromModels(t *testing.T) {
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

		works := []*model.Work{
			{
				ID:         1,
				Title:      "アニメ1",
				SeasonYear: &seasonYear1,
				SeasonName: &seasonName1,
			},
			{
				ID:         2,
				Title:      "アニメ2",
				SeasonYear: &seasonYear2,
				SeasonName: &seasonName2,
			},
		}

		result := NewWorksFromModels(works, helper)

		if len(result) != 2 {
			t.Fatalf("worksの件数 = %d、期待値 = 2", len(result))
		}

		// 1つ目の作品の検証
		if result[0].ID != 1 {
			t.Errorf("works[0].ID = %d、期待値 = 1", result[0].ID)
		}
		if result[0].Title != "アニメ1" {
			t.Errorf("works[0].Title = %s、期待値 = アニメ1", result[0].Title)
		}

		// 2つ目の作品の検証
		if result[1].ID != 2 {
			t.Errorf("works[1].ID = %d、期待値 = 2", result[1].ID)
		}
		if result[1].Title != "アニメ2" {
			t.Errorf("works[1].Title = %s、期待値 = アニメ2", result[1].Title)
		}
	})

	t.Run("空のスライス", func(t *testing.T) {
		t.Parallel()

		works := []*model.Work{}
		result := NewWorksFromModels(works, helper)

		if len(result) != 0 {
			t.Errorf("worksの件数 = %d、期待値 = 0", len(result))
		}
	})
}
