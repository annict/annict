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

	// 両方の日付を持つworks管理下のbroadcastイベント1つと、開始日だけの
	// revival_screeningを作る。2つのkindは異なるので (anime_id, kind) のUNIQUE
	// インデックスを満たし、NULL許容のended_onがround-tripする。
	started := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)
	ended := time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC)
	if _, err := repo.Create(context.Background(), repository.CreateAnimeEventParams{
		AnimeID: animeID, Kind: model.AnimeEventKindBroadcast, StartedOn: started, EndedOn: &ended,
	}); err != nil {
		t.Fatalf("Create(broadcast)のエラー = %v", err)
	}
	if _, err := repo.Create(context.Background(), repository.CreateAnimeEventParams{
		AnimeID: animeID, Kind: model.AnimeEventKindRevivalScreening, StartedOn: started, EndedOn: nil,
	}); err != nil {
		t.Fatalf("Create(revival_screening)のエラー = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d、期待値 = 2", len(got))
	}

	byKind := map[model.AnimeEventKind]*model.AnimeEvent{}
	for _, e := range got {
		if e.ID == 0 {
			t.Error("IDが採番されていない")
		}
		if e.AnimeID != animeID {
			t.Errorf("AnimeID = %d、期待値 = %d", e.AnimeID, animeID)
		}
		byKind[e.Kind] = e
	}

	broadcast := byKind[model.AnimeEventKindBroadcast]
	if broadcast == nil || !broadcast.StartedOn.Equal(started) || broadcast.EndedOn == nil || !broadcast.EndedOn.Equal(ended) {
		t.Errorf("broadcastの行 = %+v、期待値 = started_on %v / ended_on %v", broadcast, started, ended)
	}
	revival := byKind[model.AnimeEventKindRevivalScreening]
	if revival == nil || revival.EndedOn != nil {
		t.Errorf("revivalの行 = %+v、期待値 = ended_onがNULL", revival)
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
		t.Fatalf("Create()のエラー = %v", err)
	}

	// Updateは日付をその場で上書きする。終了日が追加される。
	newStarted := time.Date(2024, 4, 6, 0, 0, 0, 0, time.UTC)
	newEnded := time.Date(2024, 6, 29, 0, 0, 0, 0, time.UTC)
	if err := repo.Update(context.Background(), repository.UpdateAnimeEventParams{
		ID: created.ID, StartedOn: newStarted, EndedOn: &newEnded,
	}); err != nil {
		t.Fatalf("Update()のエラー = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d、期待値 = 1", len(got))
	}
	if !got[0].StartedOn.Equal(newStarted) || got[0].EndedOn == nil || !got[0].EndedOn.Equal(newEnded) {
		t.Errorf("更新後の行 = %+v、期待値 = started_on %v / ended_on %v", got[0], newStarted, newEnded)
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
		t.Fatalf("Create()のエラー = %v", err)
	}

	if err := repo.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete()のエラー = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("削除後のlen(got) = %d、期待値 = 0", len(got))
	}
}

func TestAnimeEventRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeEventRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("空入力時のlen(got) = %d、期待値 = 0", len(got))
	}
}
