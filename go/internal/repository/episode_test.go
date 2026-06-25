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

// insertEpisodeSyncWork inserts a minimal works row and returns its ID. animeID is
// the works.anime_id mapping column; pass an invalid NullInt64 for an unsynced
// parent.
//
// [Ja] insertEpisodeSyncWork は最小の works 行を挿入し ID を返す。animeID は
// works.anime_id マッピングカラムで、未同期の親には無効な NullInt64 を渡す。
func insertEpisodeSyncWork(t *testing.T, tx *sql.Tx, animeID sql.NullInt64) model.WorkID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(
		`INSERT INTO works (title, media, anime_id) VALUES ($1, $2, $3) RETURNING id`,
		"親作品", 1, animeID,
	).Scan(&id); err != nil {
		t.Fatalf("works の挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

// insertEpisodeSyncParentAnime inserts a bare anime row to stand in for a synced
// parent work's anime and returns its ID.
//
// [Ja] insertEpisodeSyncParentAnime は同期済みの親作品の anime に見立てた素の anime 行を
// 挿入し ID を返す。
func insertEpisodeSyncParentAnime(t *testing.T, tx *sql.Tx) model.AnimeID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("animes の挿入に失敗: %v", err)
	}
	return model.AnimeID(id)
}

// episodeSyncRow holds the episodes columns relevant to the episodes -> animes sync.
//
// [Ja] episodeSyncRow は episodes -> animes 同期に関係する episodes カラムを保持する。
type episodeSyncRow struct {
	workID         model.WorkID
	title          sql.NullString
	titleRo        string
	titleEn        string
	number         sql.NullString
	sortNumber     int32
	rawNumber      sql.NullFloat64
	status         string
	archiveMessage sql.NullString
	animeID        sql.NullInt64
}

func insertEpisodeSyncEpisode(t *testing.T, tx *sql.Tx, in episodeSyncRow) model.EpisodeID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO episodes (
			work_id, title, title_ro, title_en, number, sort_number,
			raw_number, status, archive_message, anime_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		int64(in.workID), in.title, in.titleRo, in.titleEn, in.number, in.sortNumber,
		in.rawNumber, in.status, in.archiveMessage, in.animeID,
	).Scan(&id); err != nil {
		t.Fatalf("episodes の挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

func TestEpisodeRepository_ListForAnimeSyncByIDs(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 全カラムを射影し親 anime_id を JOIN で解決する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		parentAnimeID := insertEpisodeSyncParentAnime(t, tx)
		workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{Int64: int64(parentAnimeID), Valid: true})

		episodeID := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{
			workID:     workID,
			title:      sql.NullString{String: "第3話タイトル", Valid: true},
			titleRo:    "Episode 3",
			titleEn:    "Episode Three",
			number:     sql.NullString{String: "第3話", Valid: true},
			sortNumber: 3,
			rawNumber:  sql.NullFloat64{Float64: 3.5, Valid: true},
			status:     "published",
		})

		episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
		if err != nil {
			t.Fatalf("ListForAnimeSyncByIDs() error = %v", err)
		}
		if len(episodes) != 1 {
			t.Fatalf("len(episodes) = %d, want 1", len(episodes))
		}
		e := episodes[0]

		if e.ID != episodeID {
			t.Errorf("ID = %d, want %d", e.ID, episodeID)
		}
		if e.WorkID != workID {
			t.Errorf("WorkID = %d, want %d", e.WorkID, workID)
		}
		if e.Title == nil || *e.Title != "第3話タイトル" {
			t.Errorf("Title = %v, want 第3話タイトル", e.Title)
		}
		if e.TitleRo != "Episode 3" {
			t.Errorf("TitleRo = %q, want Episode 3", e.TitleRo)
		}
		if e.TitleEn != "Episode Three" {
			t.Errorf("TitleEn = %q, want Episode Three", e.TitleEn)
		}
		if e.Number == nil || *e.Number != "第3話" {
			t.Errorf("Number = %v, want 第3話", e.Number)
		}
		if e.SortNumber != 3 {
			t.Errorf("SortNumber = %d, want 3", e.SortNumber)
		}
		if e.RawNumber == nil || *e.RawNumber != 3.5 {
			t.Errorf("RawNumber = %v, want 3.5", e.RawNumber)
		}
		if e.Status != model.EpisodeStatusPublished {
			t.Errorf("Status = %q, want published", e.Status)
		}
		if e.ParentAnimeID == nil || *e.ParentAnimeID != parentAnimeID {
			t.Errorf("ParentAnimeID = %v, want %d", e.ParentAnimeID, parentAnimeID)
		}
		// The episode itself is not yet mapped to an anime.
		//
		// [Ja] episode 自体はまだ anime にマッピングされていない。
		if e.AnimeID != nil {
			t.Errorf("AnimeID = %v, want nil", e.AnimeID)
		}
	})

	t.Run("正常系: NULL 許容カラムは nil、未同期の親は ParentAnimeID nil", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		// Parent work has no anime_id (unsynced), and the episode leaves the
		// nullable title / number / raw_number columns NULL.
		//
		// [Ja] 親作品は anime_id を持たず (未同期)、episode は NULL 許容の
		// title / number / raw_number カラムを NULL のままにする。
		workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{})
		episodeID := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{
			workID:     workID,
			sortNumber: 1,
			status:     "published",
		})

		episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
		if err != nil {
			t.Fatalf("ListForAnimeSyncByIDs() error = %v", err)
		}
		if len(episodes) != 1 {
			t.Fatalf("len(episodes) = %d, want 1", len(episodes))
		}
		e := episodes[0]

		if e.Title != nil {
			t.Errorf("Title = %v, want nil", e.Title)
		}
		if e.Number != nil {
			t.Errorf("Number = %v, want nil", e.Number)
		}
		if e.RawNumber != nil {
			t.Errorf("RawNumber = %v, want nil", e.RawNumber)
		}
		if e.ParentAnimeID != nil {
			t.Errorf("ParentAnimeID = %v, want nil (unsynced parent)", e.ParentAnimeID)
		}
	})

	t.Run("正常系: 空入力はクエリせず空スライスを返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), nil)
		if err != nil {
			t.Fatalf("ListForAnimeSyncByIDs() error = %v", err)
		}
		if len(episodes) != 0 {
			t.Errorf("len(episodes) = %d, want 0", len(episodes))
		}
	})
}

func TestEpisodeRepository_UpdateAnimeID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

	parentAnimeID := insertEpisodeSyncParentAnime(t, tx)
	workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{Int64: int64(parentAnimeID), Valid: true})
	episodeID := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{
		workID:     workID,
		sortNumber: 1,
		status:     "published",
	})

	// The anime the episode gets mapped to (its own identity row).
	//
	// [Ja] episode がマッピングされる anime (episode 自身の同一性の行)。
	episodeAnimeID := insertEpisodeSyncParentAnime(t, tx)

	if err := repo.UpdateAnimeID(context.Background(), episodeID, episodeAnimeID); err != nil {
		t.Fatalf("UpdateAnimeID() error = %v", err)
	}

	episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
	if err != nil {
		t.Fatalf("ListForAnimeSyncByIDs() error = %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("len(episodes) = %d, want 1", len(episodes))
	}
	if episodes[0].AnimeID == nil || *episodes[0].AnimeID != episodeAnimeID {
		t.Errorf("AnimeID = %v, want %d", episodes[0].AnimeID, episodeAnimeID)
	}
}
