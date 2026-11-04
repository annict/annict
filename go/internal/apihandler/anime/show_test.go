package anime_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/apihandler/anime"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

func TestHandler_GetAnime(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	classificationRepo := repository.NewAnimeClassificationRepository(queries)
	satelliteRepos := usecase.WorkSatelliteRepos{
		ExternalID:      repository.NewAnimeExternalIDRepository(queries),
		Link:            repository.NewAnimeLinkRepository(queries),
		OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
		Hashtag:         repository.NewAnimeHashtagRepository(queries),
		Season:          repository.NewAnimeSeasonRepository(queries),
		Event:           repository.NewAnimeEventRepository(queries),
	}
	h := anime.NewHandler(usecase.NewGetAPIAnimeUsecase(
		animeRepo,
		classificationRepo,
		repository.NewAnimeRedirectRepository(queries),
		repository.NewAnimeLegacyIDRepository(queries),
		satelliteRepos,
	))

	work, err := animeRepo.Create(ctx, repository.CreateAnimeParams{
		Title: sql.NullString{String: "作品", Valid: true},
		Media: model.AnimeMediaTV,
	})
	if err != nil {
		t.Fatalf("作品のanimeの作成に失敗しました: %v", err)
	}
	if _, err := classificationRepo.Create(ctx, repository.CreateAnimeClassificationParams{
		AnimeID: work.ID, Kind: model.AnimeClassificationKindWork,
	}); err != nil {
		t.Fatalf("作品の分類の作成に失敗しました: %v", err)
	}
	if _, err := satelliteRepos.Season.Create(ctx, repository.CreateAnimeSeasonParams{AnimeID: work.ID, Year: 2026}); err != nil {
		t.Fatalf("季節の作成に失敗しました: %v", err)
	}
	endedOn := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	if _, err := satelliteRepos.Event.Create(ctx, repository.CreateAnimeEventParams{
		AnimeID: work.ID, Kind: model.AnimeEventKindBroadcast,
		StartedOn: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), EndedOn: &endedOn,
	}); err != nil {
		t.Fatalf("予定の作成に失敗しました: %v", err)
	}

	episode, err := animeRepo.Create(ctx, repository.CreateAnimeParams{})
	if err != nil {
		t.Fatalf("エピソードのanimeの作成に失敗しました: %v", err)
	}
	if _, err := classificationRepo.Create(ctx, repository.CreateAnimeClassificationParams{
		AnimeID:       episode.ID,
		Kind:          model.AnimeClassificationKindEpisode,
		ParentAnimeID: &work.ID,
		Number:        sql.NullString{String: "3.5", Valid: true},
		NumberText:    sql.NullString{String: "第3.5話", Valid: true},
		SortNumber:    sql.NullInt32{Int32: 1, Valid: true},
	}); err != nil {
		t.Fatalf("エピソードの分類の作成に失敗しました: %v", err)
	}

	spec, err := apigen.GetSpec()
	if err != nil {
		t.Fatalf("GetSpec()のエラー = %v", err)
	}
	animeSchema := spec.Components.Schemas["Anime"].Value

	// getAnimeJSONは応答を、利用者が受け取るJSONのオブジェクトにして返す
	getAnimeJSON := func(t *testing.T, id model.AnimeID) map[string]any {
		t.Helper()
		res, err := h.GetAnime(ctx, apigen.GetAnimeRequestObject{AnimeId: int64(id)})
		if err != nil {
			t.Fatalf("GetAnime()のエラー = %v", err)
		}
		jsonRes, ok := res.(apigen.GetAnime200JSONResponse)
		if !ok {
			t.Fatalf("レスポンスの型 = %T、期待値 = GetAnime200JSONResponse", res)
		}
		b, err := json.Marshal(jsonRes)
		if err != nil {
			t.Fatalf("応答をJSONにできない: %v", err)
		}
		var body map[string]any
		if err := json.Unmarshal(b, &body); err != nil {
			t.Fatalf("応答のJSONを読めない: %v", err)
		}
		// 応答がOpenAPI記述の `Anime` のスキーマに合うことを確かめる
		if err := animeSchema.VisitJSON(body); err != nil {
			t.Errorf("応答がAnimeのスキーマに合わない: %v", err)
		}
		return body
	}

	// サブテストは親のトランザクションを共有するため、並列にしない

	t.Run("作品はkindがworkで、workのキーに固有の情報を入れる", func(t *testing.T) {
		body := getAnimeJSON(t, work.ID)

		if body["kind"] != "work" || body["id"] != float64(work.ID) || body["title"] != "作品" || body["media"] != "tv" {
			t.Errorf("kind・id・title・media = %v・%v・%v・%v", body["kind"], body["id"], body["title"], body["media"])
		}
		// 値の無いフィールドもキーを省かずnullで返す
		for _, key := range []string{"title_en", "release_status", "synopsis", "legacy_work_id", "legacy_episode_id"} {
			v, ok := body[key]
			if !ok || v != nil {
				t.Errorf("%s = %v (キーの有無 %v)、期待値 = null", key, v, ok)
			}
		}
		// 値の無い配列は空の配列で返す
		for _, key := range []string{"links", "external_ids", "official_accounts", "hashtags"} {
			if v, ok := body[key].([]any); !ok || len(v) != 0 {
				t.Errorf("%s = %v、期待値 = []", key, body[key])
			}
		}
		seasons, _ := body["seasons"].([]any)
		if len(seasons) != 1 || seasons[0].(map[string]any)["name"] != nil {
			t.Errorf("seasons = %v、期待値 = 季節の分からない1件", body["seasons"])
		}
		events, _ := body["events"].([]any)
		if len(events) != 1 || events[0].(map[string]any)["started_on"] != "2026-04-01" || events[0].(map[string]any)["ended_on"] != "2026-06-30" {
			t.Errorf("events = %v、期待値 = YYYY-MM-DDの日付の1件", body["events"])
		}
		w, _ := body["work"].(map[string]any)
		if w == nil || w["standalone"] != false || w["expected_episodes_count"] != nil || w["episodes_count"] != float64(1) {
			t.Errorf("work = %v", body["work"])
		}
		if _, ok := body["episode"]; ok {
			t.Error("作品の応答にepisodeのキーがある")
		}
	})

	t.Run("エピソードはkindがepisodeで、episodeのキーに固有の情報を入れる", func(t *testing.T) {
		body := getAnimeJSON(t, episode.ID)

		if body["kind"] != "episode" || body["title"] != nil {
			t.Errorf("kind・title = %v・%v、期待値 = episode・null", body["kind"], body["title"])
		}
		e, _ := body["episode"].(map[string]any)
		if e == nil {
			t.Fatalf("episode = %v", body["episode"])
		}
		if e["number"] != 3.5 || e["number_text"] != "第3.5話" || e["sort_number"] != float64(1) {
			t.Errorf("number・number_text・sort_number = %v・%v・%v", e["number"], e["number_text"], e["sort_number"])
		}
		parent, _ := e["work"].(map[string]any)
		if parent == nil || parent["id"] != float64(work.ID) || parent["title"] != "作品" {
			t.Errorf("work = %v、期待値 = 親の作品", e["work"])
		}
		for _, key := range []string{"prev_episode", "next_episode"} {
			if v, ok := e[key]; !ok || v != nil {
				t.Errorf("%s = %v (キーの有無 %v)、期待値 = null", key, v, ok)
			}
		}
		if _, ok := body["work"]; ok {
			t.Error("エピソードの応答にworkのキーがある")
		}
	})

	t.Run("UseCaseのエラーはそのまま返し、共通のヘルパーに変換を任せる", func(t *testing.T) {
		_, err := h.GetAnime(ctx, apigen.GetAnimeRequestObject{AnimeId: 0})
		if ae := model.AsAppError(err); ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
			t.Errorf("エラー = %v、期待値 = 未存在のAppError", err)
		}
	})

	t.Run("関連データの値とラベルを返し、編集者が決めた順に並べる", func(t *testing.T) {
		// 自然キーと作成順のどちらとも逆のsort_numberを指定し、応答の順序を検証する。
		// ラベルと並び順は作品の同期用のCreateでは指定できないため、直接設定する
		for _, setup := range []struct {
			name  string
			query string
		}{
			{
				name: "リンク",
				query: `INSERT INTO anime_links (anime_id, kind, language, url, label, label_en, sort_number, created_at, updated_at)
					VALUES ($1, 'official_site', 'ja', 'https://example.com', '公式サイト', NULL, 20, NOW(), NOW()),
					       ($1, 'other', 'en', 'https://example.com/guide', NULL, 'Guide', 10, NOW(), NOW())`,
			},
			{
				name: "外部ID",
				query: `INSERT INTO anime_external_ids (anime_id, service, external_id, created_at, updated_at)
					VALUES ($1, 'mal', '12345', NOW(), NOW())`,
			},
			{
				name: "公式アカウント",
				query: `INSERT INTO anime_official_accounts (anime_id, service, account, label, label_en, sort_number, created_at, updated_at)
					VALUES ($1, 'bluesky', 'anime.bsky.social', '公式', NULL, 20, NOW(), NOW()),
					       ($1, 'x', 'anime_info', NULL, 'News', 10, NOW(), NOW())`,
			},
			{
				name: "ハッシュタグ",
				query: `INSERT INTO anime_hashtags (anime_id, hashtag, sort_number, created_at, updated_at)
					VALUES ($1, 'アニメ', 20, NOW(), NOW()), ($1, 'anime_info', 10, NOW(), NOW())`,
			},
		} {
			if _, err := tx.ExecContext(ctx, setup.query, work.ID); err != nil {
				t.Fatalf("%sの作成に失敗しました: %v", setup.name, err)
			}
		}

		body := getAnimeJSON(t, work.ID)
		want := map[string]any{
			"links": []any{
				map[string]any{"kind": "other", "language": "en", "url": "https://example.com/guide", "label": nil, "label_en": "Guide"},
				map[string]any{"kind": "official_site", "language": "ja", "url": "https://example.com", "label": "公式サイト", "label_en": nil},
			},
			"external_ids": []any{
				map[string]any{"service": "mal", "external_id": "12345"},
			},
			"official_accounts": []any{
				map[string]any{"service": "x", "account": "anime_info", "label": nil, "label_en": "News"},
				map[string]any{"service": "bluesky", "account": "anime.bsky.social", "label": "公式", "label_en": nil},
			},
			"hashtags": []any{
				map[string]any{"hashtag": "anime_info"},
				map[string]any{"hashtag": "アニメ"},
			},
		}
		for key, expected := range want {
			if !reflect.DeepEqual(body[key], expected) {
				t.Errorf("%s = %#v、期待値 = %#v", key, body[key], expected)
			}
		}
	})
}
