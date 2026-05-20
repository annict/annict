package viewmodel

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

func TestNewDBWorkFormInput(t *testing.T) {
	t.Parallel()

	input := usecase.CreateWorkInput{
		Title:                 "テスト作品",
		TitleKana:             "てすとさくひん",
		TitleAlter:            "別タイトル",
		TitleEn:               "Test Work",
		TitleAlterEn:          "Alt Test Work",
		Media:                 "1",
		SeasonYear:            "2024",
		SeasonName:            "2",
		StartedOn:             "2024-04-01",
		EndedOn:               "2024-06-30",
		OfficialSiteURL:       "https://example.com",
		OfficialSiteURLEn:     "https://example.com/en",
		WikipediaURL:          "https://wikipedia.org/test",
		WikipediaURLEn:        "https://en.wikipedia.org/test",
		TwitterUsername:       "test_user",
		TwitterHashtag:        "test_hashtag",
		ScTid:                 "100",
		MalAnimeID:            "200",
		Synopsis:              "あらすじ",
		SynopsisSource:        "出典",
		SynopsisEn:            "Synopsis",
		SynopsisSourceEn:      "Source",
		ManualEpisodesCount:   "12",
		StartEpisodeRawNumber: "1",
		NumberFormatID:        "3",
		NoEpisodes:            "1",
	}

	got := NewDBWorkFormInput(input)

	if got == nil {
		t.Fatal("NewDBWorkFormInput returned nil")
	}

	tests := []struct {
		field string
		want  string
	}{
		{"title", "テスト作品"},
		{"title_kana", "てすとさくひん"},
		{"title_alter", "別タイトル"},
		{"title_en", "Test Work"},
		{"title_alter_en", "Alt Test Work"},
		{"media", "1"},
		{"season_year", "2024"},
		{"season_name", "2"},
		{"started_on", "2024-04-01"},
		{"ended_on", "2024-06-30"},
		{"official_site_url", "https://example.com"},
		{"official_site_url_en", "https://example.com/en"},
		{"wikipedia_url", "https://wikipedia.org/test"},
		{"wikipedia_url_en", "https://en.wikipedia.org/test"},
		{"twitter_username", "test_user"},
		{"twitter_hashtag", "test_hashtag"},
		{"sc_tid", "100"},
		{"mal_anime_id", "200"},
		{"synopsis", "あらすじ"},
		{"synopsis_source", "出典"},
		{"synopsis_en", "Synopsis"},
		{"synopsis_source_en", "Source"},
		{"manual_episodes_count", "12"},
		{"start_episode_raw_number", "1"},
		{"number_format_id", "3"},
		{"no_episodes", "1"},
	}

	for _, tt := range tests {
		if v := got.Val(tt.field); v != tt.want {
			t.Errorf("Val(%q) = %q, want %q", tt.field, v, tt.want)
		}
	}
}

func TestDBWorkFormInput_Val(t *testing.T) {
	t.Parallel()

	t.Run("nilレシーバはすべての項目で空文字列を返す", func(t *testing.T) {
		t.Parallel()

		var d *DBWorkFormInput
		if v := d.Val("title"); v != "" {
			t.Errorf("nil receiver Val(\"title\") = %q, want \"\"", v)
		}
		if v := d.Val("media"); v != "" {
			t.Errorf("nil receiver Val(\"media\") = %q, want \"\"", v)
		}
	})

	t.Run("未知のフィールド名は空文字列を返す", func(t *testing.T) {
		t.Parallel()

		d := &DBWorkFormInput{Title: "x"}
		if v := d.Val("unknown_field"); v != "" {
			t.Errorf("Val(\"unknown_field\") = %q, want \"\"", v)
		}
	})

	t.Run("ゼロ値レシーバはすべての項目で空文字列を返す", func(t *testing.T) {
		t.Parallel()

		d := &DBWorkFormInput{}
		fields := []string{
			"title", "title_kana", "title_alter", "title_en", "title_alter_en",
			"media", "season_year", "season_name", "started_on", "ended_on",
			"official_site_url", "official_site_url_en", "wikipedia_url", "wikipedia_url_en",
			"twitter_username", "twitter_hashtag", "sc_tid", "mal_anime_id",
			"synopsis", "synopsis_source", "synopsis_en", "synopsis_source_en",
			"manual_episodes_count", "start_episode_raw_number", "number_format_id", "no_episodes",
		}
		for _, f := range fields {
			if v := d.Val(f); v != "" {
				t.Errorf("Val(%q) on zero value = %q, want \"\"", f, v)
			}
		}
	})
}

func TestNewDBWorkListItem(t *testing.T) {
	t.Parallel()

	year := int32(2024)
	season := int32(2)

	tests := []struct {
		name            string
		work            *model.Work
		wantID          WorkID
		wantTitle       string
		wantWatchers    int32
		wantStatus      WorkStatus
		wantHasImage    bool
		wantSeasonHasJP string
	}{
		{
			name: "正常系: ImageData が設定されていれば HasImage が true になる",
			work: &model.Work{
				ID:            1,
				Title:         "画像あり作品",
				WatchersCount: 100,
				Status:        model.WorkStatusPublished,
				ImageData:     `{"id":"work_images/abc.jpg"}`,
				SeasonYear:    &year,
				SeasonName:    &season,
			},
			wantID:          WorkID(1),
			wantTitle:       "画像あり作品",
			wantWatchers:    100,
			wantStatus:      WorkStatusPublished,
			wantHasImage:    true,
			wantSeasonHasJP: "2024 春",
		},
		{
			name: "正常系: ImageData が空文字列なら HasImage は false になる",
			work: &model.Work{
				ID:            2,
				Title:         "画像なし作品",
				WatchersCount: 0,
				Status:        model.WorkStatusArchived,
				ImageData:     "",
			},
			wantID:       WorkID(2),
			wantTitle:    "画像なし作品",
			wantWatchers: 0,
			wantStatus:   WorkStatusArchived,
			wantHasImage: false,
		},
		{
			name: "正常系: シーズン未設定の場合 Season は空文字列になる",
			work: &model.Work{
				ID:     3,
				Title:  "シーズンなし作品",
				Status: model.WorkStatusPublished,
			},
			wantID:          WorkID(3),
			wantTitle:       "シーズンなし作品",
			wantStatus:      WorkStatusPublished,
			wantHasImage:    false,
			wantSeasonHasJP: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := i18n.SetLocale(context.Background(), "ja")
			got := NewDBWorkListItem(ctx, tt.work)

			if got.ID != tt.wantID {
				t.Errorf("ID = %v, want %v", got.ID, tt.wantID)
			}
			if got.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantTitle)
			}
			if got.WatchersCount != tt.wantWatchers {
				t.Errorf("WatchersCount = %d, want %d", got.WatchersCount, tt.wantWatchers)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.HasImage != tt.wantHasImage {
				t.Errorf("HasImage = %v, want %v", got.HasImage, tt.wantHasImage)
			}
			if got.Season != tt.wantSeasonHasJP {
				t.Errorf("Season = %q, want %q", got.Season, tt.wantSeasonHasJP)
			}
		})
	}
}

func TestNewDBWorkListItems(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	works := []*model.Work{
		{ID: 10, Title: "A", Status: model.WorkStatusPublished, ImageData: "{}"},
		{ID: 11, Title: "B", Status: model.WorkStatusArchived, ImageData: ""},
	}

	got := NewDBWorkListItems(ctx, works)

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].ID != WorkID(10) || !got[0].HasImage {
		t.Errorf("got[0] = %+v, want ID=10 HasImage=true", got[0])
	}
	if got[1].ID != WorkID(11) || got[1].HasImage {
		t.Errorf("got[1] = %+v, want ID=11 HasImage=false", got[1])
	}
}
