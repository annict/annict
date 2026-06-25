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

// nullStr is a test helper for building a non-NULL sql.NullString.
//
// [Ja] nullStr は非 NULL の sql.NullString を組み立てるテストヘルパー。
func nullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// fullCreateAnimeParams returns CreateAnimeParams with every column populated,
// so round-trip assertions can verify each field is persisted.
//
// [Ja] fullCreateAnimeParams は全カラムを埋めた CreateAnimeParams を返し、
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
			t.Fatalf("Create() error = %v", err)
		}
		if created.ID == 0 {
			t.Error("created.ID should be assigned")
		}

		// Re-fetch to verify the row was persisted, not just echoed back.
		//
		// [Ja] 行が単にエコーされたのでなく永続化されたことを確認するため再取得する。
		got, err := repo.GetByID(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got == nil {
			t.Fatal("GetByID() returned nil for an existing anime")
		}

		if got.Title.String != "テストアニメ" {
			t.Errorf("Title = %q, want テストアニメ", got.Title.String)
		}
		if got.TitleAlterOther.String != "测试动画" {
			t.Errorf("TitleAlterOther = %q, want 测试动画", got.TitleAlterOther.String)
		}
		if got.Media != model.AnimeMediaTV {
			t.Errorf("Media = %q, want tv", got.Media)
		}
		if got.ReleaseStatus != model.ReleaseStatusReleased {
			t.Errorf("ReleaseStatus = %q, want released", got.ReleaseStatus)
		}
		if got.Status != model.AnimeStatusPublished {
			t.Errorf("Status = %q, want published", got.Status)
		}
	})

	t.Run("正常系: NULL 許容の enum とステータス既定値を扱う", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewAnimeRepository(query.New(db).WithTx(tx))

		// Empty media / release_status become NULL; empty status defaults to
		// 'published' (mirrors the column default).
		//
		// [Ja] media / release_status の空値は NULL になり、status の空値は
		// 'published' に既定される (カラム既定値に一致)。
		created, err := repo.Create(context.Background(), repository.CreateAnimeParams{
			Title: nullStr("最小アニメ"),
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repo.GetByID(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got.Media != "" {
			t.Errorf("Media = %q, want empty (NULL)", got.Media)
		}
		if got.ReleaseStatus != "" {
			t.Errorf("ReleaseStatus = %q, want empty (NULL)", got.ReleaseStatus)
		}
		if got.Status != model.AnimeStatusPublished {
			t.Errorf("Status = %q, want published", got.Status)
		}
		if got.TitleEn.Valid {
			t.Errorf("TitleEn should be NULL, got %q", got.TitleEn.String)
		}
	})
}

func TestAnimeRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRepository(query.New(db).WithTx(tx))

	got, err := repo.GetByID(context.Background(), model.AnimeID(999999999))
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got != nil {
		t.Errorf("GetByID() = %+v, want nil for a missing anime", got)
	}
}

func TestAnimeRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRepository(query.New(db).WithTx(tx))

	created, err := repo.Create(context.Background(), fullCreateAnimeParams())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err = repo.Update(context.Background(), repository.UpdateAnimeParams{
		ID:             created.ID,
		Title:          nullStr("更新後タイトル"),
		Media:          model.AnimeMediaMovie,
		Status:         model.AnimeStatusArchived,
		ArchiveMessage: nullStr("凍結しました"),
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Title.String != "更新後タイトル" {
		t.Errorf("Title = %q, want 更新後タイトル", got.Title.String)
	}
	if got.Media != model.AnimeMediaMovie {
		t.Errorf("Media = %q, want movie", got.Media)
	}
	if got.Status != model.AnimeStatusArchived {
		t.Errorf("Status = %q, want archived", got.Status)
	}
	if got.ArchiveMessage.String != "凍結しました" {
		t.Errorf("ArchiveMessage = %q, want 凍結しました", got.ArchiveMessage.String)
	}
	// Fields not set in the update are overwritten to their zero/NULL value.
	//
	// [Ja] 更新で指定しなかったフィールドはゼロ値/NULL に上書きされる。
	if got.ReleaseStatus != "" {
		t.Errorf("ReleaseStatus = %q, want empty (overwritten to NULL)", got.ReleaseStatus)
	}
}
