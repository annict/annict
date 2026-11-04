package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// apiAnimeFixtureは、作品とエピソードのテストデータを作るためのリポジトリの束
type apiAnimeFixture struct {
	t                 *testing.T
	tx                *sql.Tx
	queries           *query.Queries
	animeRepo         *repository.AnimeRepository
	classificationRep *repository.AnimeClassificationRepository
	satelliteRepos    WorkSatelliteRepos
}

func newAPIAnimeFixture(t *testing.T) (*apiAnimeFixture, *GetAPIAnimeUsecase) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	f := &apiAnimeFixture{
		t:                 t,
		tx:                tx,
		queries:           queries,
		animeRepo:         repository.NewAnimeRepository(queries),
		classificationRep: repository.NewAnimeClassificationRepository(queries),
		satelliteRepos: WorkSatelliteRepos{
			ExternalID:      repository.NewAnimeExternalIDRepository(queries),
			Link:            repository.NewAnimeLinkRepository(queries),
			OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
			Hashtag:         repository.NewAnimeHashtagRepository(queries),
			Season:          repository.NewAnimeSeasonRepository(queries),
			Event:           repository.NewAnimeEventRepository(queries),
		},
	}
	uc := NewGetAPIAnimeUsecase(
		f.animeRepo,
		f.classificationRep,
		repository.NewAnimeRedirectRepository(queries),
		repository.NewAnimeLegacyIDRepository(queries),
		f.satelliteRepos,
	)
	return f, uc
}

func (f *apiAnimeFixture) createWork(title string, status model.AnimeStatus) model.AnimeID {
	f.t.Helper()
	anime, err := f.animeRepo.Create(context.Background(), repository.CreateAnimeParams{
		Title:  sql.NullString{String: title, Valid: true},
		Status: status,
	})
	if err != nil {
		f.t.Fatalf("作品のanimeの作成に失敗しました: %v", err)
	}
	if _, err := f.classificationRep.Create(context.Background(), repository.CreateAnimeClassificationParams{
		AnimeID:               anime.ID,
		Kind:                  model.AnimeClassificationKindWork,
		ExpectedEpisodesCount: sql.NullInt32{Int32: 12, Valid: true},
	}); err != nil {
		f.t.Fatalf("作品の分類の作成に失敗しました: %v", err)
	}
	return anime.ID
}

func (f *apiAnimeFixture) createEpisode(parentID model.AnimeID, sortNumber int32, status model.AnimeStatus) model.AnimeID {
	f.t.Helper()
	anime, err := f.animeRepo.Create(context.Background(), repository.CreateAnimeParams{
		Title:  sql.NullString{String: "エピソード", Valid: true},
		Status: status,
	})
	if err != nil {
		f.t.Fatalf("エピソードのanimeの作成に失敗しました: %v", err)
	}
	if _, err := f.classificationRep.Create(context.Background(), repository.CreateAnimeClassificationParams{
		AnimeID:       anime.ID,
		Kind:          model.AnimeClassificationKindEpisode,
		ParentAnimeID: &parentID,
		Number:        sql.NullString{String: "3.5", Valid: true},
		SortNumber:    sql.NullInt32{Int32: sortNumber, Valid: true},
	}); err != nil {
		f.t.Fatalf("エピソードの分類の作成に失敗しました: %v", err)
	}
	return anime.ID
}

func (f *apiAnimeFixture) redirect(oldID, canonicalID model.AnimeID) {
	f.t.Helper()
	if _, err := f.tx.ExecContext(context.Background(),
		"INSERT INTO anime_redirects (old_anime_id, canonical_anime_id) VALUES ($1, $2)", oldID, canonicalID); err != nil {
		f.t.Fatalf("統合の記録の作成に失敗しました: %v", err)
	}
}

func TestGetAPIAnimeUsecase_Execute_Work(t *testing.T) {
	t.Parallel()

	f, uc := newAPIAnimeFixture(t)
	ctx := context.Background()

	workID := f.createWork("作品", model.AnimeStatusPublished)
	f.createEpisode(workID, 1, model.AnimeStatusPublished)
	f.createEpisode(workID, 2, model.AnimeStatusPublished)
	f.createEpisode(workID, 3, model.AnimeStatusArchived)

	spring := model.SeasonNameSpring
	if _, err := f.satelliteRepos.Season.Create(ctx, repository.CreateAnimeSeasonParams{AnimeID: workID, Year: 2026, Name: &spring, IsPrimary: true}); err != nil {
		t.Fatalf("季節の作成に失敗しました: %v", err)
	}
	// 予定は作成した順ではなくsort_numberの順に並ぶ
	for _, e := range []struct {
		sortNumber int32
		kind       model.AnimeEventKind
	}{{sortNumber: 2, kind: model.AnimeEventKindBroadcast}, {sortNumber: 1, kind: model.AnimeEventKindOther}} {
		event, err := f.satelliteRepos.Event.Create(ctx, repository.CreateAnimeEventParams{
			AnimeID: workID, Kind: e.kind, StartedOn: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		})
		if err != nil {
			t.Fatalf("予定の作成に失敗しました: %v", err)
		}
		if _, err := f.tx.ExecContext(ctx, "UPDATE anime_events SET sort_number = $1 WHERE id = $2", e.sortNumber, event.ID); err != nil {
			t.Fatalf("予定の並び順の設定に失敗しました: %v", err)
		}
	}

	legacyWorkID := testutil.NewWorkBuilder(t, f.tx).WithTitle("旧作品").Build()
	if err := repository.NewWorkRepository(f.queries).UpdateAnimeID(ctx, legacyWorkID, workID); err != nil {
		t.Fatalf("旧作品のanime_idの設定に失敗しました: %v", err)
	}

	output, err := uc.Execute(ctx, GetAPIAnimeInput{AnimeID: workID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	a := output.Anime

	if a.Anime.ID != workID || a.Classification.Kind != model.AnimeClassificationKindWork {
		t.Errorf("anime = %d (%s)、期待値 = %d (work)", a.Anime.ID, a.Classification.Kind, workID)
	}
	if a.EpisodesCount != 2 {
		t.Errorf("EpisodesCount = %d、期待値 = 2", a.EpisodesCount)
	}
	if len(a.Seasons) != 1 || a.Seasons[0].Year != 2026 {
		t.Errorf("Seasons = %+v、期待値 = 2026年の1件", a.Seasons)
	}
	if len(a.Events) != 2 || a.Events[0].SortNumber != 1 || a.Events[1].SortNumber != 2 {
		t.Errorf("Eventsの並び = %+v、期待値 = sort_numberの順", a.Events)
	}
	if a.LegacyWorkID == nil || *a.LegacyWorkID != legacyWorkID || a.LegacyEpisodeID != nil {
		t.Errorf("旧ID = %v / %v、期待値 = 旧作品 %d だけ", a.LegacyWorkID, a.LegacyEpisodeID, legacyWorkID)
	}
	if a.Work != nil || a.PrevEpisode != nil || a.NextEpisode != nil {
		t.Error("作品にエピソードの情報が入っている")
	}
}

func TestGetAPIAnimeUsecase_Execute_Episode(t *testing.T) {
	t.Parallel()

	f, uc := newAPIAnimeFixture(t)

	workID := f.createWork("親の作品", model.AnimeStatusPublished)
	prevID := f.createEpisode(workID, 1, model.AnimeStatusPublished)
	f.createEpisode(workID, 2, model.AnimeStatusDeleted)
	episodeID := f.createEpisode(workID, 3, model.AnimeStatusPublished)

	output, err := uc.Execute(context.Background(), GetAPIAnimeInput{AnimeID: episodeID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	a := output.Anime

	if a.Classification.Kind != model.AnimeClassificationKindEpisode {
		t.Errorf("Kind = %q、期待値 = episode", a.Classification.Kind)
	}
	if a.Work == nil || a.Work.ID != workID {
		t.Errorf("親の作品 = %+v、期待値 = %d", a.Work, workID)
	}
	if a.PrevEpisode == nil || a.PrevEpisode.ID != prevID {
		t.Errorf("前のエピソード = %+v、期待値 = %d (削除済みを飛ばす)", a.PrevEpisode, prevID)
	}
	if a.NextEpisode != nil {
		t.Errorf("次のエピソード = %d、期待値 = nil", a.NextEpisode.ID)
	}
	if len(a.Seasons) != 0 || len(a.Events) != 0 {
		t.Errorf("別表の無いエピソードの季節 = %v・予定 = %v、期待値 = 空", a.Seasons, a.Events)
	}
}

func TestGetAPIAnimeUsecase_Execute_EpisodeParentStatus(t *testing.T) {
	t.Parallel()

	for _, status := range []model.AnimeStatus{model.AnimeStatusPublished, model.AnimeStatusArchived, model.AnimeStatusDeleted} {
		t.Run(string(status), func(t *testing.T) {
			t.Parallel()

			f, uc := newAPIAnimeFixture(t)
			parentID := f.createWork("親の作品", status)
			episodeID := f.createEpisode(parentID, 1, model.AnimeStatusPublished)
			output, err := uc.Execute(context.Background(), GetAPIAnimeInput{AnimeID: episodeID})
			if status != model.AnimeStatusPublished {
				ae := model.AsAppError(err)
				if ae == nil || ae.Code != model.AppErrCodeResourceNotFound || output != nil {
					t.Fatalf("Execute() = %+v, %v、期待値 = nil, 未存在のAppError", output, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.Anime.Work == nil || output.Anime.Work.ID != parentID {
				t.Errorf("親の作品 = %+v、期待値 = %d", output.Anime.Work, parentID)
			}
		})
	}
}

func TestGetAPIAnimeUsecase_Execute_Status(t *testing.T) {
	t.Parallel()

	f, uc := newAPIAnimeFixture(t)

	canonicalID := f.createWork("統合先", model.AnimeStatusPublished)
	mergedID := f.createWork("統合元", model.AnimeStatusMerged)
	f.redirect(mergedID, canonicalID)
	mergedWithoutRedirectID := f.createWork("記録の無い統合元", model.AnimeStatusMerged)
	archivedCanonicalID := f.createWork("非公開の統合先", model.AnimeStatusArchived)
	mergedToArchivedID := f.createWork("非公開に統合した統合元", model.AnimeStatusMerged)
	f.redirect(mergedToArchivedID, archivedCanonicalID)
	archivedID := f.createWork("非公開", model.AnimeStatusArchived)
	deletedID := f.createWork("削除済み", model.AnimeStatusDeleted)

	tests := []struct {
		name   string
		id     model.AnimeID
		wantID *model.AnimeID
	}{
		{name: "統合したanimeは統合先を返す", id: mergedID, wantID: &canonicalID},
		{name: "統合の記録が無ければ404", id: mergedWithoutRedirectID},
		{name: "統合先が公開中でなければ404", id: mergedToArchivedID},
		{name: "非公開は404", id: archivedID},
		{name: "削除済みは404", id: deletedID},
		{name: "存在しないIDは404", id: model.AnimeID(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetAPIAnimeInput{AnimeID: tt.id})

			if tt.wantID == nil {
				ae := model.AsAppError(err)
				if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
					t.Fatalf("エラー = %v、期待値 = 未存在のAppError", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.Anime.Anime.ID != *tt.wantID {
				t.Errorf("返したanime = %d、期待値 = %d", output.Anime.Anime.ID, *tt.wantID)
			}
		})
	}
}

func TestGetAPIAnimeUsecase_Execute_MissingClassification(t *testing.T) {
	t.Parallel()

	f, uc := newAPIAnimeFixture(t)
	anime, err := f.animeRepo.Create(context.Background(), repository.CreateAnimeParams{Status: model.AnimeStatusPublished})
	if err != nil {
		t.Fatalf("animeの作成に失敗しました: %v", err)
	}

	// 分類の無いanimeはデータの不整合のため、存在しないとは伝えず、500になるエラーにする
	output, err := uc.Execute(context.Background(), GetAPIAnimeInput{AnimeID: anime.ID})
	if err == nil || model.AsAppError(err) != nil || output != nil {
		t.Errorf("Execute() = %+v, %v、期待値 = nil, AppErrorでないエラー", output, err)
	}
}
