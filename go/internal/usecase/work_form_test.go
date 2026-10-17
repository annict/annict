package usecase

import (
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

func TestBuildWorkFormParams(t *testing.T) {
	t.Parallel()

	t.Run("任意入力の値を型付きのカラムに変換する", func(t *testing.T) {
		t.Parallel()

		params, err := buildWorkFormParams(WorkFormInput{
			Title:               " タイトル ",
			Media:               "1",
			SeasonYear:          "2024",
			SeasonName:          "2",
			StartedOn:           "2024-04-05",
			EndedOn:             "2024-06-28",
			TwitterUsername:     " anime_official ",
			TwitterHashtag:      " anime2024 ",
			ScTid:               "5678",
			MalAnimeID:          "1234",
			ManualEpisodesCount: "12",
			NumberFormatID:      "3",
		})
		if err != nil {
			t.Fatalf("buildWorkFormParams()のエラー = %v", err)
		}

		if params.Title != "タイトル" {
			t.Errorf("Title = %q、期待値 = タイトル", params.Title)
		}
		if params.SeasonYear == nil || *params.SeasonYear != 2024 {
			t.Errorf("SeasonYear = %v、期待値 = 2024", params.SeasonYear)
		}
		if params.SeasonName == nil || *params.SeasonName != 2 {
			t.Errorf("SeasonName = %v、期待値 = 2", params.SeasonName)
		}
		wantStartedOn := time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC)
		if params.StartedOn == nil || !params.StartedOn.Equal(wantStartedOn) {
			t.Errorf("StartedOn = %v、期待値 = %v", params.StartedOn, wantStartedOn)
		}
		wantEndedOn := time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)
		if params.EndedOn == nil || !params.EndedOn.Equal(wantEndedOn) {
			t.Errorf("EndedOn = %v、期待値 = %v", params.EndedOn, wantEndedOn)
		}
		if params.TwitterUsername == nil || *params.TwitterUsername != "anime_official" {
			t.Errorf("TwitterUsername = %v、期待値 = anime_official", params.TwitterUsername)
		}
		if params.TwitterHashtag == nil || *params.TwitterHashtag != "anime2024" {
			t.Errorf("TwitterHashtag = %v、期待値 = anime2024", params.TwitterHashtag)
		}
		if params.ScTid == nil || *params.ScTid != 5678 {
			t.Errorf("ScTid = %v、期待値 = 5678", params.ScTid)
		}
		if params.MalAnimeID == nil || *params.MalAnimeID != 1234 {
			t.Errorf("MalAnimeID = %v、期待値 = 1234", params.MalAnimeID)
		}
		if params.ManualEpisodesCount == nil || *params.ManualEpisodesCount != 12 {
			t.Errorf("ManualEpisodesCount = %v、期待値 = 12", params.ManualEpisodesCount)
		}
		if params.NumberFormatID == nil || *params.NumberFormatID != model.NumberFormatID(3) {
			t.Errorf("NumberFormatID = %v、期待値 = 3", params.NumberFormatID)
		}
	})

	t.Run("空欄の任意入力はnilにする", func(t *testing.T) {
		t.Parallel()

		params, err := buildWorkFormParams(WorkFormInput{
			Title: "タイトル",
			Media: "1",
		})
		if err != nil {
			t.Fatalf("buildWorkFormParams()のエラー = %v", err)
		}

		if params.SeasonYear != nil {
			t.Errorf("SeasonYear = %v、期待値 = nil", *params.SeasonYear)
		}
		if params.SeasonName != nil {
			t.Errorf("SeasonName = %v、期待値 = nil", *params.SeasonName)
		}
		if params.StartedOn != nil {
			t.Errorf("StartedOn = %v、期待値 = nil", *params.StartedOn)
		}
		if params.EndedOn != nil {
			t.Errorf("EndedOn = %v、期待値 = nil", *params.EndedOn)
		}
		if params.TwitterUsername != nil {
			t.Errorf("TwitterUsername = %q、期待値 = nil", *params.TwitterUsername)
		}
		if params.TwitterHashtag != nil {
			t.Errorf("TwitterHashtag = %q、期待値 = nil", *params.TwitterHashtag)
		}
		if params.ScTid != nil {
			t.Errorf("ScTid = %v、期待値 = nil", *params.ScTid)
		}
		if params.MalAnimeID != nil {
			t.Errorf("MalAnimeID = %v、期待値 = nil", *params.MalAnimeID)
		}
		if params.ManualEpisodesCount != nil {
			t.Errorf("ManualEpisodesCount = %v、期待値 = nil", *params.ManualEpisodesCount)
		}
		if params.NumberFormatID != nil {
			t.Errorf("NumberFormatID = %v、期待値 = nil", *params.NumberFormatID)
		}
		// 開始話数はNOT NULLのカラムのため、空欄でもnilにせず既定値の1にする。
		if params.StartEpisodeRawNumber != 1 {
			t.Errorf("StartEpisodeRawNumber = %v、期待値 = 1", params.StartEpisodeRawNumber)
		}
	})
}
