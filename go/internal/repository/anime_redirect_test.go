package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestAnimeRedirectRepository_FindCanonicalAnimeID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeRedirectRepository(queries)

	oldID := createTestAnime(t, animeRepo, "統合元")
	canonicalID := createTestAnime(t, animeRepo, "統合先")
	if _, err := tx.ExecContext(context.Background(),
		"INSERT INTO anime_redirects (old_anime_id, canonical_anime_id) VALUES ($1, $2)", oldID, canonicalID); err != nil {
		t.Fatalf("統合の記録の作成に失敗しました: %v", err)
	}

	got, err := repo.FindCanonicalAnimeID(context.Background(), oldID)
	if err != nil {
		t.Fatalf("FindCanonicalAnimeID()のエラー = %v", err)
	}
	if got == nil || *got != canonicalID {
		t.Errorf("統合先 = %v、期待値 = %d", got, canonicalID)
	}

	got, err = repo.FindCanonicalAnimeID(context.Background(), canonicalID)
	if err != nil {
		t.Fatalf("記録の無いanimeでのFindCanonicalAnimeID()のエラー = %v", err)
	}
	if got != nil {
		t.Errorf("記録の無いanimeの統合先 = %d、期待値 = nil", *got)
	}
}

// 統合先が見つからないanimeのIDも、記録が無いものとして扱う
func TestAnimeRedirectRepository_FindCanonicalAnimeID_Unknown(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRedirectRepository(query.New(db).WithTx(tx))

	got, err := repo.FindCanonicalAnimeID(context.Background(), model.AnimeID(0))
	if err != nil || got != nil {
		t.Errorf("統合先 = %v、エラー = %v、期待値 = nil", got, err)
	}
}

func TestAnimeRedirectRepository_ListCanonicalAnimeIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeRedirectRepository(queries)

	oldID := createTestAnime(t, animeRepo, "統合元")
	canonicalID := createTestAnime(t, animeRepo, "統合先")
	if _, err := tx.ExecContext(context.Background(),
		"INSERT INTO anime_redirects (old_anime_id, canonical_anime_id) VALUES ($1, $2)", oldID, canonicalID); err != nil {
		t.Fatalf("統合の記録の作成に失敗しました: %v", err)
	}

	got, err := repo.ListCanonicalAnimeIDs(context.Background(), []model.AnimeID{oldID, canonicalID})
	if err != nil {
		t.Fatalf("ListCanonicalAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 1 || got[oldID] != canonicalID {
		t.Errorf("統合先 = %v、期待値 = 統合元 %d から統合先 %d だけ", got, oldID, canonicalID)
	}

	empty, err := repo.ListCanonicalAnimeIDs(context.Background(), nil)
	if err != nil || len(empty) != 0 {
		t.Errorf("空入力の結果 = %v、エラー = %v、期待値 = 空", empty, err)
	}
}
