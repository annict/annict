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
			"updated_at",
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
		wantStatus      PublishingStatus
		wantHasImage    bool
		wantSeasonHasJP string
	}{
		{
			name: "正常系: 画像がある作品は Image が実サムネイルを解決する",
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
			wantStatus:      PublishingStatusPublished,
			wantHasImage:    true,
			wantSeasonHasJP: "2024年春",
		},
		{
			name: "正常系: unpublished_at があれば archived になり、title_kana 未設定は空文字列・画像なしは Image がプレースホルダーになる",
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
			wantStatus:    PublishingStatusArchived,
			wantHasImage:  false,
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
			wantStatus:      PublishingStatusPublished,
			wantHasImage:    false,
			wantSeasonHasJP: "",
		},
		{
			// season_name that falls outside the known enum (1..4) has no label key, so
			// the season display falls back to the same year-only display as a work with
			// no season at all.
			//
			// [Ja] season_name が既知の enum (1..4) の範囲外だとラベルキーが無いため、
			// 季節が未登録の作品と同じ年のみの表示にフォールバックする。
			name: "正常系: season_name が範囲外 enum のとき Season は年のみの表示になる",
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
			wantStatus:      PublishingStatusPublished,
			wantHasImage:    false,
			wantSeasonHasJP: "2024年 (季節未登録)",
		},
		{
			name: "正常系: 年だけが登録された作品は Season が年のみの表示になる",
			work: &model.Work{
				ID:         7,
				Title:      "年のみ登録作品",
				Media:      1,
				SeasonYear: &year,
			},
			wantID:          WorkID(7),
			wantTitle:       "年のみ登録作品",
			wantMedia:       "TV",
			wantStatus:      PublishingStatusPublished,
			wantHasImage:    false,
			wantSeasonHasJP: "2024年 (季節未登録)",
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
			wantStatus: PublishingStatusPublished,
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
			wantStatus: PublishingStatusPublished,
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
			if got.Image.Exists() != tt.wantHasImage {
				t.Errorf("Image.Exists() = %v (URL %q), want %v", got.Image.Exists(), got.Image.URL(70, "jpg"), tt.wantHasImage)
			}
			if got.Season != tt.wantSeasonHasJP {
				t.Errorf("Season = %q, want %q", got.Season, tt.wantSeasonHasJP)
			}
		})
	}
}

// TestFormatSeason pins the release-season display for every season_year / season_name
// combination in both locales, including the year-only display a work gets when the
// season is unregistered and the empty string that makes the list render a "-".
//
// [Ja] TestFormatSeason は season_year / season_name の全ての組み合わせに対するリリース
// 時期の表示を、両ロケールで固定する。季節が未登録の作品が受け取る年のみの表示と、一覧に
// "-" を描かせる空文字列も含む。
func TestFormatSeason(t *testing.T) {
	t.Parallel()

	year := int32(2024)
	spring := int32(2)
	unknownSeason := int32(5)

	tests := []struct {
		name   string
		locale string
		year   *int32
		season *int32
		want   string
	}{
		{name: "正常系: ja で年と季節が揃っていれば両方を表示する", locale: "ja", year: &year, season: &spring, want: "2024年春"},
		{name: "正常系: en で年と季節が揃っていれば両方を表示する", locale: "en", year: &year, season: &spring, want: "Spring 2024"},
		{name: "正常系: ja で季節が未登録なら年のみを表示する", locale: "ja", year: &year, season: nil, want: "2024年 (季節未登録)"},
		{name: "正常系: en で季節が未登録なら年のみを表示する", locale: "en", year: &year, season: nil, want: "2024 (No Season)"},
		{name: "正常系: 範囲外 enum の季節は未登録と同じ表示になる", locale: "ja", year: &year, season: &unknownSeason, want: "2024年 (季節未登録)"},
		{name: "正常系: 年が未登録なら空文字列を返す", locale: "ja", year: nil, season: nil, want: ""},
		{name: "正常系: 季節だけが登録されていても年が無ければ空文字列を返す", locale: "ja", year: nil, season: &spring, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := i18n.SetLocale(context.Background(), tt.locale)

			if got := formatSeason(ctx, tt.year, tt.season); got != tt.want {
				t.Errorf("formatSeason() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNewDBWorkListItem_StatusFromTimestamps verifies that the display status is
// derived from the work's unpublished_at / deleted_at timestamps, with deleted_at
// taking precedence over unpublished_at.
//
// [Ja] TestNewDBWorkListItem_StatusFromTimestamps は表示ステータスが work の
// unpublished_at / deleted_at タイムスタンプから導出され、deleted_at が unpublished_at
// より優先されることを検証する。
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
		want          PublishingStatus
	}{
		{
			name: "両方 nil なら published",
			want: PublishingStatusPublished,
		},
		{
			name:          "unpublished_at のみなら archived",
			unpublishedAt: &unpublishedAt,
			want:          PublishingStatusArchived,
		},
		{
			name:      "deleted_at のみなら deleted",
			deletedAt: &deletedAt,
			want:      PublishingStatusDeleted,
		},
		{
			name:          "両方あれば deleted_at が優先される",
			unpublishedAt: &unpublishedAt,
			deletedAt:     &deletedAt,
			want:          PublishingStatusDeleted,
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

// TestDBWorkListItem_Image verifies that the list item wires the work's image_data into
// its WorkImage, so a work with an image resolves to a real thumbnail and one without
// falls back to the placeholder.
//
// [Ja] TestDBWorkListItem_Image は一覧アイテムが作品の image_data を WorkImage に配線し、
// 画像がある作品は実サムネイルに、無い作品はプレースホルダーに解決されることを検証する。
func TestDBWorkListItem_Image(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	helper := testutil.NewTestImageHelper()

	withImage := NewDBWorkListItem(ctx, &model.Work{
		ID:        1,
		Title:     "画像あり作品",
		ImageData: `{"master":{"id":"workimage/1/image/master-abc.jpg","storage":"store"}}`,
	}, helper)
	if !withImage.Image.Exists() {
		t.Error("画像がある作品では Image.Exists() が true になるべきです")
	}
	if withImage.Image.SrcSet(70, "webp") == "" {
		t.Error("画像がある作品では SrcSet が非空を返すべきです")
	}

	withoutImage := NewDBWorkListItem(ctx, &model.Work{
		ID:        2,
		Title:     "画像なし作品",
		ImageData: "",
	}, helper)
	if withoutImage.Image.Exists() {
		t.Error("画像がない作品では Image.Exists() が false になるべきです")
	}
	if got := withoutImage.Image.URL(70, "jpg"); got != NoWorkImagePath {
		t.Errorf("画像がない作品の URL = %q, want %q", got, NoWorkImagePath)
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
	if got[0].ID != WorkID(10) || !got[0].Image.Exists() {
		t.Errorf("got[0] = %+v, want ID=10 で画像あり", got[0])
	}
	if got[1].ID != WorkID(11) || got[1].Image.Exists() {
		t.Errorf("got[1] = %+v, want ID=11 で画像なし", got[1])
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

// TestDBWorkFormInput_Version covers the version the work edit form round-trips: opening the
// form takes it from the stored row, and a rejected submit hands back the one the editor sent
// rather than whatever the server holds by then.
//
// [Ja] TestDBWorkFormInput_Version は作品編集フォームが往復させる版を対象とする。フォームを開く
// ときは保存済みの行から取り、却下された送信では、その時点でサーバーが持つ値ではなく編集者が
// 送った版を返す。
func TestDBWorkFormInput_Version(t *testing.T) {
	t.Parallel()

	t.Run("保存済みの updated_at をフォームの版に射影する", func(t *testing.T) {
		t.Parallel()

		updatedAt := time.Date(2026, 8, 17, 1, 2, 3, 456789000, time.UTC)
		got := NewDBWorkFormInputFromWork(&model.Work{Title: "版あり作品", UpdatedAt: &updatedAt})

		// The form's value has to parse back to the instant it came from, since the update
		// matches it against the stored column.
		//
		// [Ja] フォームの値は元の時刻へパースし直せる必要がある。更新側が保存済みのカラムと
		// 照合するため。
		parsed, err := time.Parse(formVersionLayout, got.Val("updated_at"))
		if err != nil {
			t.Fatalf("版のパースに失敗: %v (value=%q)", err, got.UpdatedAt)
		}
		if !parsed.Equal(updatedAt) {
			t.Errorf("version = %v, want %v", parsed, updatedAt)
		}
	})

	t.Run("updated_at を持たない work はセンチネルを運ぶ", func(t *testing.T) {
		t.Parallel()

		got := NewDBWorkFormInputFromWork(&model.Work{Title: "版なし作品"})
		if got.UpdatedAt != FormNullVersion {
			t.Errorf("UpdatedAt = %q, want %q", got.UpdatedAt, FormNullVersion)
		}
	})

	t.Run("却下された送信は送られた版をそのまま返す", func(t *testing.T) {
		t.Parallel()

		submitted := "2026-08-17T01:02:03.456789Z"
		got := NewDBWorkFormInputFromSubmit(usecase.UpdateWorkInput{
			WorkID:        model.WorkID(1),
			UpdatedAt:     submitted,
			WorkFormInput: usecase.WorkFormInput{Title: "送信されたタイトル"},
		})

		if got.UpdatedAt != submitted {
			t.Errorf("UpdatedAt = %q, want %q", got.UpdatedAt, submitted)
		}
		if got.Title != "送信されたタイトル" {
			t.Errorf("Title = %q, want 送信されたタイトル", got.Title)
		}
	})

	t.Run("作成フォームは版を運ばない", func(t *testing.T) {
		t.Parallel()

		got := NewDBWorkFormInput(usecase.WorkFormInput{Title: "作成中の作品"})
		if got.UpdatedAt != "" {
			t.Errorf("UpdatedAt = %q, want empty", got.UpdatedAt)
		}
	})
}
