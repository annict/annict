package viewmodel

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

func TestNewDBWorkFormInput(t *testing.T) {
	t.Parallel()

	input := usecase.WorkFormInput{
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

func TestDBWorkFormInput_LabelLinkURL(t *testing.T) {
	t.Parallel()

	t.Run("nilレシーバは空文字列を返す", func(t *testing.T) {
		t.Parallel()

		var d *DBWorkFormInput
		if got := d.LabelLinkURL("official_site_url"); got != "" {
			t.Errorf("nil receiver LabelLinkURL(\"official_site_url\") = %q, want \"\"", got)
		}
	})

	t.Run("値が入っていればリンク先を返す", func(t *testing.T) {
		t.Parallel()

		d := &DBWorkFormInput{
			OfficialSiteURL:   "https://example.com",
			OfficialSiteURLEn: "https://example.com/en",
			WikipediaURL:      "https://ja.wikipedia.org/wiki/x",
			WikipediaURLEn:    "https://en.wikipedia.org/wiki/x",
			TwitterUsername:   "annict_com",
			TwitterHashtag:    "annict",
			ScTid:             "3524",
			MalAnimeID:        "20",
		}
		tests := map[string]string{
			"official_site_url":    "https://example.com",
			"official_site_url_en": "https://example.com/en",
			"wikipedia_url":        "https://ja.wikipedia.org/wiki/x",
			"wikipedia_url_en":     "https://en.wikipedia.org/wiki/x",
			"twitter_username":     "https://x.com/annict_com",
			"twitter_hashtag":      "https://x.com/search?q=%23annict",
			"sc_tid":               "http://cal.syoboi.jp/tid/3524",
			"mal_anime_id":         "https://myanimelist.net/anime/20",
		}
		for field, want := range tests {
			if got := d.LabelLinkURL(field); got != want {
				t.Errorf("LabelLinkURL(%q) = %q, want %q", field, got, want)
			}
		}
	})

	t.Run("値が空ならリンク先も空", func(t *testing.T) {
		t.Parallel()

		d := &DBWorkFormInput{}
		linkable := []string{
			"official_site_url", "official_site_url_en", "wikipedia_url", "wikipedia_url_en",
			"twitter_username", "twitter_hashtag", "sc_tid", "mal_anime_id",
		}
		for _, field := range linkable {
			if got := d.LabelLinkURL(field); got != "" {
				t.Errorf("LabelLinkURL(%q) on empty input = %q, want \"\"", field, got)
			}
		}
	})

	t.Run("リンク非対象フィールドは空を返す", func(t *testing.T) {
		t.Parallel()

		d := &DBWorkFormInput{Title: "x", Synopsis: "y"}
		for _, field := range []string{"title", "synopsis", "media", "unknown"} {
			if got := d.LabelLinkURL(field); got != "" {
				t.Errorf("LabelLinkURL(%q) = %q, want \"\"", field, got)
			}
		}
	})
}

func TestNewDBWorkListItem(t *testing.T) {
	t.Parallel()

	helper := testutil.NewTestImageHelper()

	year := int32(2024)
	season := int32(2)
	unknownSeason := int32(5)
	titleKana := "がぞうありさくひん"
	unpublishedAt := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		work            *model.Work
		wantID          WorkID
		wantTitle       string
		wantTitleKana   string
		wantTitleEn     string
		wantMedia       string
		wantWatchers    int32
		wantStatus      WorkStatus
		wantImageURL    bool
		wantSeasonHasJP string
	}{
		{
			name: "正常系: 画像がある作品は ImageURL が生成される",
			work: &model.Work{
				ID:            1,
				Title:         "画像あり作品",
				TitleKana:     &titleKana,
				TitleEn:       "Work With Image",
				Media:         1,
				WatchersCount: 100,
				ImageData:     `{"master":{"id":"workimage/1/image/master-abc.jpg","storage":"store"}}`,
				SeasonYear:    &year,
				SeasonName:    &season,
			},
			wantID:          WorkID(1),
			wantTitle:       "画像あり作品",
			wantTitleKana:   "がぞうありさくひん",
			wantTitleEn:     "Work With Image",
			wantMedia:       "TV",
			wantWatchers:    100,
			wantStatus:      WorkStatusPublished,
			wantImageURL:    true,
			wantSeasonHasJP: "2024 春",
		},
		{
			name: "正常系: unpublished_at があれば archived になり、title_kana 未設定は空文字列・画像なしは ImageURL が空になる",
			work: &model.Work{
				ID:            2,
				Title:         "画像なし作品",
				Media:         2,
				WatchersCount: 0,
				UnpublishedAt: &unpublishedAt,
				ImageData:     "",
			},
			wantID:        WorkID(2),
			wantTitle:     "画像なし作品",
			wantTitleKana: "",
			wantTitleEn:   "",
			wantMedia:     "OVA",
			wantWatchers:  0,
			wantStatus:    WorkStatusArchived,
			wantImageURL:  false,
		},
		{
			name: "正常系: シーズン未設定の場合 Season は空文字列になる",
			work: &model.Work{
				ID:    3,
				Title: "シーズンなし作品",
				Media: 0,
			},
			wantID:          WorkID(3),
			wantTitle:       "シーズンなし作品",
			wantMedia:       "その他",
			wantStatus:      WorkStatusPublished,
			wantImageURL:    false,
			wantSeasonHasJP: "",
		},
		{
			// season_name that falls outside the known enum (1..4) has no label key, so
			// the season display falls back to the year alone.
			//
			// [Ja] season_name が既知の enum (1..4) の範囲外だとラベルキーが無いため、
			// シーズン表示は年のみにフォールバックする。
			name: "正常系: season_name が範囲外 enum のとき Season は年のみになる",
			work: &model.Work{
				ID:         6,
				Title:      "範囲外シーズン作品",
				Media:      1,
				SeasonYear: &year,
				SeasonName: &unknownSeason,
			},
			wantID:          WorkID(6),
			wantTitle:       "範囲外シーズン作品",
			wantMedia:       "TV",
			wantStatus:      WorkStatusPublished,
			wantImageURL:    false,
			wantSeasonHasJP: "2024",
		},
		{
			name: "正常系: media = 3 は 映画 に変換される",
			work: &model.Work{
				ID:    4,
				Title: "映画作品",
				Media: 3,
			},
			wantID:     WorkID(4),
			wantTitle:  "映画作品",
			wantMedia:  "映画",
			wantStatus: WorkStatusPublished,
		},
		{
			name: "正常系: media = 4 は Web に変換される",
			work: &model.Work{
				ID:    5,
				Title: "Web作品",
				Media: 4,
			},
			wantID:     WorkID(5),
			wantTitle:  "Web作品",
			wantMedia:  "Web",
			wantStatus: WorkStatusPublished,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := i18n.SetLocale(context.Background(), "ja")
			got := NewDBWorkListItem(ctx, tt.work, helper)

			if got.ID != tt.wantID {
				t.Errorf("ID = %v, want %v", got.ID, tt.wantID)
			}
			if got.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantTitle)
			}
			if got.TitleKana != tt.wantTitleKana {
				t.Errorf("TitleKana = %q, want %q", got.TitleKana, tt.wantTitleKana)
			}
			if got.TitleEn != tt.wantTitleEn {
				t.Errorf("TitleEn = %q, want %q", got.TitleEn, tt.wantTitleEn)
			}
			if got.Media != tt.wantMedia {
				t.Errorf("Media = %q, want %q", got.Media, tt.wantMedia)
			}
			if got.WatchersCount != tt.wantWatchers {
				t.Errorf("WatchersCount = %d, want %d", got.WatchersCount, tt.wantWatchers)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if (got.ImageURL != "") != tt.wantImageURL {
				t.Errorf("ImageURL presence = %v (%q), want %v", got.ImageURL != "", got.ImageURL, tt.wantImageURL)
			}
			if got.Season != tt.wantSeasonHasJP {
				t.Errorf("Season = %q, want %q", got.Season, tt.wantSeasonHasJP)
			}
		})
	}
}

// TestNewDBWorkListItem_StatusFromTimestamps verifies that the display status is
// derived from the work's unpublished_at / deleted_at timestamps (not the dormant
// works.status), with deleted_at taking precedence over unpublished_at.
//
// [Ja] TestNewDBWorkListItem_StatusFromTimestamps は表示ステータスが work の
// unpublished_at / deleted_at タイムスタンプ (休眠している works.status ではない) から
// 導出され、deleted_at が unpublished_at より優先されることを検証する。
func TestNewDBWorkListItem_StatusFromTimestamps(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	helper := testutil.NewTestImageHelper()

	unpublishedAt := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		unpublishedAt *time.Time
		deletedAt     *time.Time
		want          WorkStatus
	}{
		{
			name: "両方 nil なら published",
			want: WorkStatusPublished,
		},
		{
			name:          "unpublished_at のみなら archived",
			unpublishedAt: &unpublishedAt,
			want:          WorkStatusArchived,
		},
		{
			name:      "deleted_at のみなら deleted",
			deletedAt: &deletedAt,
			want:      WorkStatusDeleted,
		},
		{
			name:          "両方あれば deleted_at が優先される",
			unpublishedAt: &unpublishedAt,
			deletedAt:     &deletedAt,
			want:          WorkStatusDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDBWorkListItem(ctx, &model.Work{
				ID:            1,
				Title:         "作品",
				UnpublishedAt: tt.unpublishedAt,
				DeletedAt:     tt.deletedAt,
			}, helper)

			if got.Status != tt.want {
				t.Errorf("Status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

// TestNewDBWorkListItem_ExternalServices verifies that sc_tid / mal_anime_id map to
// the Syoboi Calendar / MyAnimeList links, and that an unset id yields an empty link.
//
// [Ja] TestNewDBWorkListItem_ExternalServices は sc_tid / mal_anime_id が
// しょぼかる / MyAnimeList リンクに写像されること、未設定の ID では空リンクになることを検証する。
func TestNewDBWorkListItem_ExternalServices(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	helper := testutil.NewTestImageHelper()

	scTid := int32(3524)
	malAnimeID := int32(20)

	t.Run("sc_tid / mal_anime_id があればラベルと URL を持つ", func(t *testing.T) {
		t.Parallel()

		got := NewDBWorkListItem(ctx, &model.Work{
			ID:         1,
			Title:      "外部サービスあり作品",
			ScTid:      &scTid,
			MalAnimeID: &malAnimeID,
		}, helper)

		if got.Syobocal.Label != "3524" || got.Syobocal.URL != "http://cal.syoboi.jp/tid/3524" {
			t.Errorf("Syobocal = %+v, want label 3524 / しょぼかる URL", got.Syobocal)
		}
		if got.MalAnime.Label != "20" || got.MalAnime.URL != "https://myanimelist.net/anime/20" {
			t.Errorf("MalAnime = %+v, want label 20 / MyAnimeList URL", got.MalAnime)
		}
	})

	t.Run("sc_tid / mal_anime_id が未設定なら空リンクになる", func(t *testing.T) {
		t.Parallel()

		got := NewDBWorkListItem(ctx, &model.Work{
			ID:    2,
			Title: "外部サービスなし作品",
		}, helper)

		if got.Syobocal != (ExternalServiceLink{}) {
			t.Errorf("Syobocal = %+v, want zero value", got.Syobocal)
		}
		if got.MalAnime != (ExternalServiceLink{}) {
			t.Errorf("MalAnime = %+v, want zero value", got.MalAnime)
		}
	})
}

// TestDBWorkListItem_GetSrcSet verifies that the thumbnail srcset is produced only
// when the work has an image and an image helper is wired.
//
// [Ja] TestDBWorkListItem_GetSrcSet はサムネイルの srcset が、作品に画像があり
// 画像ヘルパーが配線されているときにのみ生成されることを検証する。
func TestDBWorkListItem_GetSrcSet(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	helper := testutil.NewTestImageHelper()

	withImage := NewDBWorkListItem(ctx, &model.Work{
		ID:        1,
		Title:     "画像あり作品",
		ImageData: `{"master":{"id":"workimage/1/image/master-abc.jpg","storage":"store"}}`,
	}, helper)
	if withImage.GetSrcSet(70, "webp") == "" {
		t.Error("画像がある作品では GetSrcSet が非空を返すべきです")
	}

	withoutImage := NewDBWorkListItem(ctx, &model.Work{
		ID:        2,
		Title:     "画像なし作品",
		ImageData: "",
	}, helper)
	if got := withoutImage.GetSrcSet(70, "webp"); got != "" {
		t.Errorf("画像がない作品では GetSrcSet が空を返すべきです: %q", got)
	}

	// A nil image helper (e.g. a struct literal) yields an empty srcset.
	//
	// [Ja] 画像ヘルパーが nil (例: 構造体リテラル) の場合は空の srcset を返す。
	noHelper := DBWorkListItem{ImageDataJSON: `{"master":{"id":"x","storage":"store"}}`}
	if got := noHelper.GetSrcSet(70, "webp"); got != "" {
		t.Errorf("画像ヘルパーが nil の場合は空を返すべきです: %q", got)
	}
}

func TestNewDBWorkListItems(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	helper := testutil.NewTestImageHelper()
	works := []*model.Work{
		{ID: 10, Title: "A", ImageData: `{"master":{"id":"workimage/10/image/master-a.jpg","storage":"store"}}`},
		{ID: 11, Title: "B", ImageData: ""},
	}

	got := NewDBWorkListItems(ctx, works, helper)

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].ID != WorkID(10) || got[0].ImageURL == "" {
		t.Errorf("got[0] = %+v, want ID=10 で ImageURL が非空", got[0])
	}
	if got[1].ID != WorkID(11) || got[1].ImageURL != "" {
		t.Errorf("got[1] = %+v, want ID=11 で ImageURL が空", got[1])
	}
}

func TestNewDBWorkFormInputFromWork(t *testing.T) {
	t.Parallel()

	t.Run("全フィールドが埋まった work を文字列フォーム値に射影する", func(t *testing.T) {
		titleKana := "てすとさくひん"
		twitterUsername := "test_user"
		twitterHashtag := "test_hashtag"
		var scTid int32 = 100
		var malAnimeID int32 = 200
		var manualEpisodesCount int32 = 12
		var seasonYear int32 = 2024
		var seasonName int32 = 2
		startedOn := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
		endedOn := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		numberFormatID := model.NumberFormatID(3)

		work := &model.Work{
			ID:                    1,
			Title:                 "テスト作品",
			TitleKana:             &titleKana,
			TitleAlter:            "別タイトル",
			TitleEn:               "Test Work",
			TitleAlterEn:          "Alt Test Work",
			Media:                 1,
			SeasonYear:            &seasonYear,
			SeasonName:            &seasonName,
			StartedOn:             &startedOn,
			EndedOn:               &endedOn,
			OfficialSiteURL:       "https://example.com",
			OfficialSiteURLEn:     "https://example.com/en",
			WikipediaURL:          "https://wikipedia.org/test",
			WikipediaURLEn:        "https://en.wikipedia.org/test",
			TwitterUsername:       &twitterUsername,
			TwitterHashtag:        &twitterHashtag,
			ScTid:                 &scTid,
			MalAnimeID:            &malAnimeID,
			Synopsis:              "あらすじ",
			SynopsisSource:        "出典",
			SynopsisEn:            "Synopsis",
			SynopsisSourceEn:      "Source",
			ManualEpisodesCount:   &manualEpisodesCount,
			StartEpisodeRawNumber: 2.5,
			NumberFormatID:        &numberFormatID,
			NoEpisodes:            true,
		}

		got := NewDBWorkFormInputFromWork(work)
		if got == nil {
			t.Fatal("NewDBWorkFormInputFromWork returned nil")
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
			{"start_episode_raw_number", "2.5"},
			{"number_format_id", "3"},
			{"no_episodes", "1"},
		}
		for _, tt := range tests {
			if v := got.Val(tt.field); v != tt.want {
				t.Errorf("Val(%q) = %q, want %q", tt.field, v, tt.want)
			}
		}
	})

	t.Run("nullable が未設定の work は空文字列で返す", func(t *testing.T) {
		work := &model.Work{
			ID:                    2,
			Title:                 "最小作品",
			Media:                 0,
			StartEpisodeRawNumber: 1,
			NoEpisodes:            false,
		}

		got := NewDBWorkFormInputFromWork(work)

		emptyFields := []string{
			"title_kana", "season_year", "season_name", "started_on", "ended_on",
			"twitter_username", "twitter_hashtag", "sc_tid", "mal_anime_id",
			"manual_episodes_count", "number_format_id", "no_episodes",
		}
		for _, field := range emptyFields {
			if v := got.Val(field); v != "" {
				t.Errorf("Val(%q) = %q, want empty string", field, v)
			}
		}
		if v := got.Val("media"); v != "0" {
			t.Errorf("Val(media) = %q, want 0", v)
		}
		if v := got.Val("start_episode_raw_number"); v != "1" {
			t.Errorf("Val(start_episode_raw_number) = %q, want 1", v)
		}
		if v := got.Val("title"); v != "最小作品" {
			t.Errorf("Val(title) = %q, want 最小作品", v)
		}
	})
}
