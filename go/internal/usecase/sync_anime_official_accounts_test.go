package usecase

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func strPtr(v string) *string { return &v }

// newSyncAnimeOfficialAccountsUsecase builds the reconciler and its repository over the
// shared test DB. The usecase opens its own transaction, so the test commits its setup
// (animes) directly via GetTestDB rather than wrapping it in an outer tx.
//
// [Ja] newSyncAnimeOfficialAccountsUsecase は共有テスト DB 上にリコンサイラとそのリポジトリを
// 組み立てる。本 UseCase は自前でトランザクションを開くため、テストは前提データ (animes) を
// アウター tx で包まず GetTestDB 経由で直接コミットする。
func newSyncAnimeOfficialAccountsUsecase(db *sql.DB) (*SyncAnimeOfficialAccountsUsecase, *repository.AnimeOfficialAccountRepository) {
	repo := repository.NewAnimeOfficialAccountRepository(query.New(db))
	return NewSyncAnimeOfficialAccountsUsecase(db, repo), repo
}

// workForOfficialAccountSync builds an anime-resolved work carrying only the column the
// official-account reconciler reads (twitter_username).
//
// [Ja] workForOfficialAccountSync は公式アカウントリコンサイラが読むカラム (twitter_username)
// だけを持つ anime 解決済みの work を組み立てる。
func workForOfficialAccountSync(animeID model.AnimeID, twitterUsername *string) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, TwitterUsername: twitterUsername}
}

// accountsByService reads back an anime's official accounts keyed by service.
//
// [Ja] accountsByService は anime の公式アカウントをサービスをキーに読み戻す。
func accountsByService(t *testing.T, repo *repository.AnimeOfficialAccountRepository, animeID model.AnimeID) map[model.AnimeAccountService]string {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	byService := make(map[model.AnimeAccountService]string, len(rows))
	for _, r := range rows {
		byService[r.Service] = r.Account
	}
	return byService
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_CreatesRowFromWorkColumn(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeOfficialAccountsUsecase(db)

	animeID := insertBareAnime(t, db)
	work := workForOfficialAccountSync(animeID, strPtr("rezero_official"))

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 1 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Created 1 only", counts)
	}

	if got := accountsByService(t, repo, animeID)[model.AnimeAccountServiceX]; got != "rezero_official" {
		t.Errorf("x account = %q, want rezero_official", got)
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeOfficialAccountsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForOfficialAccountSync(animeID, strPtr("rezero_official"))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// Re-running with the same source must detect no diff and write nothing. This is
	// the invariant the cutover decision depends on (a synced page reports zero diff).
	//
	// [Ja] 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が
	// 依拠する不変条件 (同期済みのページは差分ゼロを報告する)。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 1 {
		t.Errorf("counts = %+v, want Unchanged 1 only", counts)
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_UpdatesChangedAccount(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeOfficialAccountsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, strPtr("old_handle"))}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, strPtr("new_handle"))})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Updated 1 only", counts)
	}

	if got := accountsByService(t, repo, animeID)[model.AnimeAccountServiceX]; got != "new_handle" {
		t.Errorf("x account = %q, want new_handle after update", got)
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeOfficialAccountsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, strPtr("rezero_official"))}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// twitter_username is gone (NULL); the x row should be deleted.
	//
	// [Ja] twitter_username が消えた (NULL)。x の行は削除されるべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, nil)})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Deleted 1 only", counts)
	}

	if _, ok := accountsByService(t, repo, animeID)[model.AnimeAccountServiceX]; ok {
		t.Error("x row should have been deleted")
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_TreatsNullAndEmptyAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeOfficialAccountsUsecase(db)

	// A NULL twitter_username and an empty-string twitter_username are both "absent":
	// no row is created for either.
	//
	// [Ja] NULL の twitter_username と空文字列の twitter_username はどちらも「欠損」:
	// どちらも行は作られない。
	nullAnimeID := insertBareAnime(t, db)
	emptyAnimeID := insertBareAnime(t, db)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{
		workForOfficialAccountSync(nullAnimeID, nil),
		workForOfficialAccountSync(emptyAnimeID, strPtr("")),
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want all zero", counts)
	}

	if rows := accountsByService(t, repo, nullAnimeID); len(rows) != 0 {
		t.Errorf("rows = %v, want none for NULL source", rows)
	}
	if rows := accountsByService(t, repo, emptyAnimeID); len(rows) != 0 {
		t.Errorf("rows = %v, want none for empty source", rows)
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeOfficialAccountsUsecase(db)

	animeID := insertBareAnime(t, db)

	// Seed two existing rows: one works-managed (x) and one outside the works-managed
	// key space (a youtube account an editor might add directly).
	//
	// [Ja] 既存行を 2 つ用意する: works 管理下の 1 つ (x) と、works 管理下のキー空間の外の
	// 1 つ (編集者が直接足しうる youtube アカウント)。
	seed := []repository.CreateAnimeOfficialAccountParams{
		{AnimeID: animeID, Service: model.AnimeAccountServiceX, Account: "managed_x"},
		{AnimeID: animeID, Service: model.AnimeAccountServiceYoutube, Account: "editor_youtube"},
	}
	for _, s := range seed {
		if _, err := repo.Create(context.Background(), s); err != nil {
			t.Fatalf("seed Create(%+v) error = %v", s, err)
		}
	}

	// The work sources nothing, so the managed x row is deleted while the editor-added
	// youtube row outside the key space is preserved.
	//
	// [Ja] work は何も source しないため、管理下の x は削除されるが、キー空間の外の編集者追加の
	// youtube 行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, nil)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Deleted 1 only", counts)
	}

	byService := accountsByService(t, repo, animeID)
	if _, ok := byService[model.AnimeAccountServiceX]; ok {
		t.Error("managed x row should have been deleted")
	}
	if byService[model.AnimeAccountServiceYoutube] != "editor_youtube" {
		t.Error("editor-added youtube account should be preserved")
	}
}
