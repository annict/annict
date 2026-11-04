package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// nullStrは非NULLのsql.NullStringを組み立てるテストヘルパー。
func nullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// fullCreateAnimeParamsは全カラムを埋めたCreateAnimeParamsを返し、
// ラウンドトリップのアサーションで各フィールドの永続化を検証できるようにする。
func fullCreateAnimeParams() repository.CreateAnimeParams {
	return repository.CreateAnimeParams{
		Title:            nullStr("テストアニメ"),
		TitleKana:        nullStr("てすとあにめ"),
		TitleRo:          nullStr("Test Anime"),
		TitleEn:          nullStr("Test Anime EN"),
		TitleAlter:       nullStr("別名"),
		TitleAlterRo:     nullStr("Betsumei"),
		TitleAlterEn:     nullStr("Alternative"),
		TitleAlterOther:  nullStr("测试动画"),
		Media:            model.AnimeMediaTV,
		ReleaseStatus:    model.ReleaseStatusReleased,
		Synopsis:         nullStr("あらすじ"),
		SynopsisEn:       nullStr("synopsis"),
		SynopsisSource:   nullStr("出典"),
		SynopsisSourceEn: nullStr("source"),
		Status:           model.AnimeStatusPublished,
		ArchiveMessage:   nullStr(""),
	}
}

func TestAnimeRepository_Create(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 全カラムを永続化して作成した行を返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewAnimeRepository(query.New(db).WithTx(tx))

		created, err := repo.Create(context.Background(), fullCreateAnimeParams())
		if err != nil {
			t.Fatalf("Create()のエラー = %v", err)
		}
		if created.ID == 0 {
			t.Error("created.IDが採番されていない")
		}

		// 行が単にエコーされたのでなく永続化されたことを確認するため再取得する。
		got, err := repo.GetByID(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("GetByID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("既存のanimeに対してGetByID()がnilを返した")
		}

		if got.Title.String != "テストアニメ" {
			t.Errorf("Title = %q、期待値 = テストアニメ", got.Title.String)
		}
		if got.TitleAlterOther.String != "测试动画" {
			t.Errorf("TitleAlterOther = %q、期待値 = 测试动画", got.TitleAlterOther.String)
		}
		if got.Media != model.AnimeMediaTV {
			t.Errorf("Media = %q、期待値 = tv", got.Media)
		}
		if got.ReleaseStatus != model.ReleaseStatusReleased {
			t.Errorf("ReleaseStatus = %q、期待値 = released", got.ReleaseStatus)
		}
		if got.Status != model.AnimeStatusPublished {
			t.Errorf("Status = %q、期待値 = published", got.Status)
		}
	})

	t.Run("正常系: NULL許容のenumとステータス既定値を扱う", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewAnimeRepository(query.New(db).WithTx(tx))

		// media / release_statusの空値はNULLになり、statusの空値は
		// 'published' に既定される (カラム既定値に一致)。
		created, err := repo.Create(context.Background(), repository.CreateAnimeParams{
			Title: nullStr("最小アニメ"),
		})
		if err != nil {
			t.Fatalf("Create()のエラー = %v", err)
		}

		got, err := repo.GetByID(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("GetByID()のエラー = %v", err)
		}
		if got.Media != "" {
			t.Errorf("Media = %q、期待値 = 空 (NULL)", got.Media)
		}
		if got.ReleaseStatus != "" {
			t.Errorf("ReleaseStatus = %q、期待値 = 空 (NULL)", got.ReleaseStatus)
		}
		if got.Status != model.AnimeStatusPublished {
			t.Errorf("Status = %q、期待値 = published", got.Status)
		}
		if got.TitleEn.Valid {
			t.Errorf("TitleEn = %q、期待値 = NULL", got.TitleEn.String)
		}
	})
}

func TestAnimeRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRepository(query.New(db).WithTx(tx))

	got, err := repo.GetByID(context.Background(), model.AnimeID(999999999))
	if err != nil {
		t.Fatalf("GetByID()のエラー = %v", err)
	}
	if got != nil {
		t.Errorf("GetByID() = %+v、期待値 = nil (存在しないanimeのため)", got)
	}
}

func TestAnimeRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRepository(query.New(db).WithTx(tx))

	created, err := repo.Create(context.Background(), fullCreateAnimeParams())
	if err != nil {
		t.Fatalf("Create()のエラー = %v", err)
	}

	err = repo.Update(context.Background(), repository.UpdateAnimeParams{
		ID:             created.ID,
		Title:          nullStr("更新後タイトル"),
		Media:          model.AnimeMediaMovie,
		Status:         model.AnimeStatusArchived,
		ArchiveMessage: nullStr("凍結しました"),
	})
	if err != nil {
		t.Fatalf("Update()のエラー = %v", err)
	}

	got, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID()のエラー = %v", err)
	}
	if got.Title.String != "更新後タイトル" {
		t.Errorf("Title = %q、期待値 = 更新後タイトル", got.Title.String)
	}
	if got.Media != model.AnimeMediaMovie {
		t.Errorf("Media = %q、期待値 = movie", got.Media)
	}
	if got.Status != model.AnimeStatusArchived {
		t.Errorf("Status = %q、期待値 = archived", got.Status)
	}
	if got.ArchiveMessage.String != "凍結しました" {
		t.Errorf("ArchiveMessage = %q、期待値 = 凍結しました", got.ArchiveMessage.String)
	}
	// 更新で指定しなかったフィールドはゼロ値/NULLに上書きされる。
	if got.ReleaseStatus != "" {
		t.Errorf("ReleaseStatus = %q、期待値 = 空 (NULLで上書きされること)", got.ReleaseStatus)
	}
}

// TestAnimeRepository_UpdateStatusはライフサイクル状態の遷移がすべての内容属性をそのまま
// 保持することを検証する。
func TestAnimeRepository_UpdateStatus(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRepository(query.New(db).WithTx(tx))
	params := fullCreateAnimeParams()
	params.ArchiveMessage = nullStr("保持するメッセージ")

	created, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create()のエラー = %v", err)
	}
	if err := repo.UpdateStatus(context.Background(), created.ID, model.AnimeStatusArchived); err != nil {
		t.Fatalf("UpdateStatus()のエラー = %v", err)
	}

	got, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID()のエラー = %v", err)
	}
	if got.Status != model.AnimeStatusArchived {
		t.Errorf("Status = %q、期待値 = archived", got.Status)
	}
	if got.Title.String != "テストアニメ" {
		t.Errorf("Title = %q、期待値 = テストアニメ", got.Title.String)
	}
	if got.Media != model.AnimeMediaTV {
		t.Errorf("Media = %q、期待値 = tv", got.Media)
	}
	if got.ArchiveMessage.String != "保持するメッセージ" {
		t.Errorf("ArchiveMessage = %q、期待値 = 保持するメッセージ", got.ArchiveMessage.String)
	}
}

// TestAnimeRepository_UpdateTitleは、タイトルだけを更新し、ほかのカラムを書き戻さないことを
// 検証する。
func TestAnimeRepository_UpdateTitle(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRepository(query.New(db).WithTx(tx))

	created, err := repo.Create(context.Background(), fullCreateAnimeParams())
	if err != nil {
		t.Fatalf("Create()のエラー = %v", err)
	}
	if err := repo.UpdateTitle(context.Background(), created.ID, "妻役を確保せよ"); err != nil {
		t.Fatalf("UpdateTitle()のエラー = %v", err)
	}

	got, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID()のエラー = %v", err)
	}
	if got.Title.String != "妻役を確保せよ" {
		t.Errorf("Title = %q、期待値 = 妻役を確保せよ", got.Title.String)
	}
	if got.TitleKana.String != "てすとあにめ" || got.Synopsis.String != "あらすじ" {
		t.Errorf("(TitleKana, Synopsis) = (%q, %q)、期待値 = (てすとあにめ, あらすじ)", got.TitleKana.String, got.Synopsis.String)
	}
}

// createTestEpisodeAnimeは、作品parentIDのエピソードのanimeと分類を作る
func createTestEpisodeAnime(t *testing.T, queries *query.Queries, parentID model.AnimeID, sortNumber int32, status model.AnimeStatus) model.AnimeID {
	t.Helper()
	anime, err := repository.NewAnimeRepository(queries).Create(context.Background(), repository.CreateAnimeParams{
		Title:  nullStr("エピソード"),
		Status: status,
	})
	if err != nil {
		t.Fatalf("エピソードのanimeの作成に失敗しました: %v", err)
	}
	if _, err := repository.NewAnimeClassificationRepository(queries).Create(context.Background(), repository.CreateAnimeClassificationParams{
		AnimeID:       anime.ID,
		Kind:          model.AnimeClassificationKindEpisode,
		ParentAnimeID: &parentID,
		SortNumber:    sql.NullInt32{Int32: sortNumber, Valid: true},
	}); err != nil {
		t.Fatalf("エピソードの分類の作成に失敗しました: %v", err)
	}
	return anime.ID
}

// createTestWorkAnimeは作品のanimeと分類を作り、seasonsを代表の季節から順に付ける
func createTestWorkAnime(t *testing.T, queries *query.Queries, title string, status model.AnimeStatus, seasons ...model.SeasonBucket) model.AnimeID {
	t.Helper()
	ctx := context.Background()
	anime, err := repository.NewAnimeRepository(queries).Create(ctx, repository.CreateAnimeParams{
		Title:  nullStr(title),
		Status: status,
	})
	if err != nil {
		t.Fatalf("作品のanimeの作成に失敗しました: %v", err)
	}
	if _, err := repository.NewAnimeClassificationRepository(queries).Create(ctx, repository.CreateAnimeClassificationParams{
		AnimeID: anime.ID,
		Kind:    model.AnimeClassificationKindWork,
	}); err != nil {
		t.Fatalf("作品の分類の作成に失敗しました: %v", err)
	}
	for i, s := range seasons {
		if _, err := repository.NewAnimeSeasonRepository(queries).Create(ctx, repository.CreateAnimeSeasonParams{
			AnimeID: anime.ID, Year: s.Year, Name: s.Name, IsPrimary: i == 0,
		}); err != nil {
			t.Fatalf("季節の作成に失敗しました: %v", err)
		}
	}
	return anime.ID
}

func season(year int32, name model.SeasonName) model.SeasonBucket {
	return model.SeasonBucket{Year: year, Name: &name}
}

func animeIDsOf(animes []*model.Anime) []model.AnimeID {
	ids := make([]model.AnimeID, len(animes))
	for i, a := range animes {
		ids[i] = a.ID
	}
	return ids
}

func TestAnimeRepository_ListAPIWorks(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	repo := repository.NewAnimeRepository(queries)

	fall2025 := createTestWorkAnime(t, queries, "秋の作品", model.AnimeStatusPublished, season(2025, model.SeasonNameFall))
	unknown2025 := createTestWorkAnime(t, queries, "季節不明の作品", model.AnimeStatusPublished, model.SeasonBucket{Year: 2025})
	winter2026 := createTestWorkAnime(t, queries, "100%_の作品", model.AnimeStatusPublished, season(2026, model.SeasonNameWinter), season(2025, model.SeasonNameFall))
	noSeason := createTestWorkAnime(t, queries, "季節の無い作品", model.AnimeStatusPublished)
	// 同じ季節キーの作品どうしと、代表の季節の無い作品どうしは、IDの向きで並ぶ
	sameFall2025 := createTestWorkAnime(t, queries, "同じ秋の作品", model.AnimeStatusPublished, season(2025, model.SeasonNameFall))
	otherNoSeason := createTestWorkAnime(t, queries, "もう1つの季節の無い作品", model.AnimeStatusPublished)
	archived := createTestWorkAnime(t, queries, "非公開の作品", model.AnimeStatusArchived, season(2025, model.SeasonNameFall))
	episode := createTestEpisodeAnime(t, queries, fall2025, 1, model.AnimeStatusPublished)
	candidates := []model.AnimeID{fall2025, unknown2025, winter2026, noSeason, sameFall2025, otherNoSeason, archived, episode}

	tests := []struct {
		name  string
		order repository.APIWorkOrder
		want  []model.AnimeID
	}{
		{name: "IDの順。非公開とエピソードは含めない", order: repository.APIWorkOrderID, want: []model.AnimeID{fall2025, unknown2025, winter2026, noSeason, sameFall2025, otherNoSeason}},
		{name: "IDの逆順", order: repository.APIWorkOrderIDDesc, want: []model.AnimeID{otherNoSeason, sameFall2025, noSeason, winter2026, unknown2025, fall2025}},
		{
			name:  "代表の季節の順。季節不明は同じ年の秋の後、季節の無い作品は最後。どちらのグループの中もIDの昇順",
			order: repository.APIWorkOrderSeason,
			want:  []model.AnimeID{fall2025, sameFall2025, unknown2025, winter2026, noSeason, otherNoSeason},
		},
		{
			name:  "代表の季節の逆順でも、季節の無い作品は最後。どちらのグループの中もIDの降順",
			order: repository.APIWorkOrderSeasonDesc,
			want:  []model.AnimeID{winter2026, unknown2025, sameFall2025, fall2025, otherNoSeason, noSeason},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ListAPIWorks(ctx, repository.ListAPIWorksParams{IDs: candidates, Order: tt.order, Limit: 10})
			if err != nil {
				t.Fatalf("ListAPIWorks()のエラー = %v", err)
			}
			assertAnimeIDs(t, animeIDsOf(got), tt.want)

			// 1件ずつカーソルで送っても、同じ並びになる
			var paged []model.AnimeID
			params := repository.ListAPIWorksParams{IDs: candidates, Order: tt.order, Limit: 1}
			for range len(tt.want) + 1 {
				page, err := repo.ListAPIWorks(ctx, params)
				if err != nil {
					t.Fatalf("ListAPIWorks()のエラー = %v", err)
				}
				if len(page) == 0 {
					break
				}
				last := page[0]
				paged = append(paged, last.ID)
				params.AfterID = &last.ID
				params.AfterSeasonKey = nil
				seasons, err := repository.NewAnimeSeasonRepository(queries).ListByAnimeIDs(ctx, []model.AnimeID{last.ID})
				if err != nil {
					t.Fatalf("季節の取得に失敗しました: %v", err)
				}
				for _, s := range seasons {
					if s.IsPrimary {
						key := s.SortKey()
						params.AfterSeasonKey = &key
					}
				}
			}
			assertAnimeIDs(t, paged, tt.want)
		})
	}

	t.Run("タイトルの部分一致はLIKEの特殊文字をそのままの文字として扱う", func(t *testing.T) {
		title := "%_"
		got, err := repo.ListAPIWorks(ctx, repository.ListAPIWorksParams{IDs: candidates, Title: &title, Order: repository.APIWorkOrderID, Limit: 10})
		if err != nil {
			t.Fatalf("ListAPIWorks()のエラー = %v", err)
		}
		assertAnimeIDs(t, animeIDsOf(got), []model.AnimeID{winter2026})
	})

	t.Run("候補が空なら何も返さない", func(t *testing.T) {
		got, err := repo.ListAPIWorks(ctx, repository.ListAPIWorksParams{IDs: []model.AnimeID{}, Order: repository.APIWorkOrderID, Limit: 10})
		if err != nil {
			t.Fatalf("ListAPIWorks()のエラー = %v", err)
		}
		assertAnimeIDs(t, animeIDsOf(got), nil)
	})
}

func TestAnimeRepository_ListAPIEpisodes(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	repo := repository.NewAnimeRepository(queries)

	workID := createTestAnime(t, repo, "エピソードの一覧の作品")
	second := createTestEpisodeAnime(t, queries, workID, 20, model.AnimeStatusPublished)
	first := createTestEpisodeAnime(t, queries, workID, 10, model.AnimeStatusPublished)
	createTestEpisodeAnime(t, queries, workID, 15, model.AnimeStatusArchived)
	// sort_numberが同じエピソードは、anime_idの順に並ぶ
	tied := createTestEpisodeAnime(t, queries, workID, 20, model.AnimeStatusPublished)

	got, err := repo.ListAPIEpisodes(ctx, workID, nil, 10)
	if err != nil {
		t.Fatalf("ListAPIEpisodes()のエラー = %v", err)
	}
	assertAnimeIDs(t, animeIDsOf(got), []model.AnimeID{first, second, tied})

	got, err = repo.ListAPIEpisodes(ctx, workID, &repository.AfterEpisode{SortNumber: 20, AnimeID: second}, 10)
	if err != nil {
		t.Fatalf("カーソルを渡したListAPIEpisodes()のエラー = %v", err)
	}
	assertAnimeIDs(t, animeIDsOf(got), []model.AnimeID{tied})
}

func assertAnimeIDs(t *testing.T, got, want []model.AnimeID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("anime = %v、期待値 = %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("anime = %v、期待値 = %v", got, want)
		}
	}
}
