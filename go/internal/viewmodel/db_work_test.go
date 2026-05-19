package viewmodel

import (
	"testing"

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
