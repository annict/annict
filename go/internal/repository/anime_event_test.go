package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestAnimeEventRepository_CreateAndListByAnimeIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeEventRepository(queries)

	animeID := createTestAnime(t, animeRepo, "イベント同期アニメ")

	// One works-managed broadcast event with both dates, plus a name-less
	// revival_screening with only a start date. The two kinds are distinct so the UNIQUE
	// index on (anime_id, kind) is satisfied, and the nullable ended_on round-trips.
	//
	// [Ja] 両方の日付を持つ works 管理下の broadcast イベント 1 つと、開始日だけの
	// revival_screening を作る。2 つの kind は異なるので (anime_id, kind) の UNIQUE
	// インデックスを満たし、NULL 許容の ended_on が round-trip する。
	started := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)
	ended := time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC)
	if _, err := repo.Create(context.Background(), repository.CreateAnimeEventParams{
		AnimeID: animeID, Kind: model.AnimeEventKindBroadcast, StartedOn: started, EndedOn: &ended,
	}); err != nil {
		t.Fatalf("Create(broadcast) error = %v", err)
	}
	if _, err := repo.Create(context.Background(), repository.CreateAnimeEventParams{
		AnimeID: animeID, Kind: model.AnimeEventKindRevivalScreening, StartedOn: started, EndedOn: nil,
	}); err != nil {
		t.Fatalf("Create(revival_screening) error = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}

	byKind := map[model.AnimeEventKind]*model.AnimeEvent{}
	for _, e := range got {
		if e.ID == 0 {
			t.Error("ID should be assigned")
		}
		if e.AnimeID != animeID {
			t.Errorf("AnimeID = %d, want %d", e.AnimeID, animeID)
		}
		byKind[e.Kind] = e
	}

	broadcast := byKind[model.AnimeEventKindBroadcast]
	if broadcast == nil || !broadcast.StartedOn.Equal(started) || broadcast.EndedOn == nil || !broadcast.EndedOn.Equal(ended) {
		t.Errorf("broadcast row = %+v, want started %v ended %v", broadcast, started, ended)
	}
	revival := byKind[model.AnimeEventKindRevivalScreening]
	if revival == nil || revival.EndedOn != nil {
		t.Errorf("revival row = %+v, want NULL ended_on", revival)
	}
}

func TestAnimeEventRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeEventRepository(queries)

	animeID := createTestAnime(t, animeRepo, "イベント更新アニメ")

	started := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)
	created, err := repo.Create(context.Background(), repository.CreateAnimeEventParams{
		AnimeID: animeID, Kind: model.AnimeEventKindBroadcast, StartedOn: started, EndedOn: nil,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update overwrites the dates in place; an end date is added.
	//
	// [Ja] Update は日付をその場で上書きする。終了日が追加される。
	newStarted := time.Date(2024, 4, 6, 0, 0, 0, 0, time.UTC)
	newEnded := time.Date(2024, 6, 29, 0, 0, 0, 0, time.UTC)
	if err := repo.Update(context.Background(), repository.UpdateAnimeEventParams{
		ID: created.ID, StartedOn: newStarted, EndedOn: &newEnded,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if !got[0].StartedOn.Equal(newStarted) || got[0].EndedOn == nil || !got[0].EndedOn.Equal(newEnded) {
		t.Errorf("updated row = %+v, want started %v ended %v", got[0], newStarted, newEnded)
	}
}

func TestAnimeEventRepository_Delete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeEventRepository(queries)

	animeID := createTestAnime(t, animeRepo, "イベント削除アニメ")

	started := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)
	created, err := repo.Create(context.Background(), repository.CreateAnimeEventParams{
		AnimeID: animeID, Kind: model.AnimeEventKindBroadcast, StartedOn: started, EndedOn: nil,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(got) = %d, want 0 after delete", len(got))
	}
}

func TestAnimeEventRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeEventRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(got) = %d, want 0 for empty input", len(got))
	}
}
