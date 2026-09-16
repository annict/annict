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

// newSyncAnimeOfficialAccountsUsecaseは共有テストDB上にリコンサイラとそのリポジトリを
// 組み立てる。本UseCaseは自前でトランザクションを開くため、テストは前提データ (animes) を
// アウターtxで包まずGetTestDB経由で直接コミットする。
func newSyncAnimeOfficialAccountsUsecase(db *sql.DB) (*SyncAnimeOfficialAccountsUsecase, *repository.AnimeOfficialAccountRepository) {
	repo := repository.NewAnimeOfficialAccountRepository(query.New(db))
	return NewSyncAnimeOfficialAccountsUsecase(db, repo), repo
}

// workForOfficialAccountSyncは公式アカウントリコンサイラが読むカラム (twitter_username)
// だけを持つanime解決済みのworkを組み立てる。
func workForOfficialAccountSync(animeID model.AnimeID, twitterUsername *string) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, TwitterUsername: twitterUsername}
}

// accountsByServiceはanimeの公式アカウントをサービスをキーに読み戻す。
func accountsByService(t *testing.T, repo *repository.AnimeOfficialAccountRepository, animeID model.AnimeID) map[model.AnimeAccountService]string {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
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
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 1 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Createdのみ1", counts)
	}

	if got := accountsByService(t, repo, animeID)[model.AnimeAccountServiceX]; got != "rezero_official" {
		t.Errorf("x account = %q、期待値 = rezero_official", got)
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeOfficialAccountsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForOfficialAccountSync(animeID, strPtr("rezero_official"))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が
	// 依拠する不変条件 (同期済みのページは差分ゼロを報告する)。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 1 {
		t.Errorf("counts = %+v、期待値 = Unchangedのみ1", counts)
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_UpdatesChangedAccount(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeOfficialAccountsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, strPtr("old_handle"))}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, strPtr("new_handle"))})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Updatedのみ1", counts)
	}

	if got := accountsByService(t, repo, animeID)[model.AnimeAccountServiceX]; got != "new_handle" {
		t.Errorf("更新後のxのアカウント = %q、期待値 = new_handle", got)
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeOfficialAccountsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, strPtr("rezero_official"))}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// twitter_usernameが消えた (NULL)。xの行は削除されるべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, nil)})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Deletedのみ1", counts)
	}

	if _, ok := accountsByService(t, repo, animeID)[model.AnimeAccountServiceX]; ok {
		t.Error("xの行が削除されなかった")
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_TreatsNullAndEmptyAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeOfficialAccountsUsecase(db)

	// NULLのtwitter_usernameと空文字列のtwitter_usernameはどちらも「欠損」:
	// どちらも行は作られない。
	nullAnimeID := insertBareAnime(t, db)
	emptyAnimeID := insertBareAnime(t, db)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{
		workForOfficialAccountSync(nullAnimeID, nil),
		workForOfficialAccountSync(emptyAnimeID, strPtr("")),
	})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = すべて0", counts)
	}

	if rows := accountsByService(t, repo, nullAnimeID); len(rows) != 0 {
		t.Errorf("同期元がNULLのときのrows = %v、期待値 = 0件", rows)
	}
	if rows := accountsByService(t, repo, emptyAnimeID); len(rows) != 0 {
		t.Errorf("同期元が空のときのrows = %v、期待値 = 0件", rows)
	}
}

func TestSyncAnimeOfficialAccountsUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeOfficialAccountsUsecase(db)

	animeID := insertBareAnime(t, db)

	// 既存行を2つ用意する: works管理下の1つ (x) と、works管理下のキー空間の外の
	// 1つ (編集者が直接足しうるyoutubeアカウント)。
	seed := []repository.CreateAnimeOfficialAccountParams{
		{AnimeID: animeID, Service: model.AnimeAccountServiceX, Account: "managed_x"},
		{AnimeID: animeID, Service: model.AnimeAccountServiceYoutube, Account: "editor_youtube"},
	}
	for _, s := range seed {
		if _, err := repo.Create(context.Background(), s); err != nil {
			t.Fatalf("シードのCreate(%+v)のエラー = %v", s, err)
		}
	}

	// workは何もsourceしないため、管理下のxは削除されるが、キー空間の外の編集者追加の
	// youtube行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForOfficialAccountSync(animeID, nil)})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Deletedのみ1", counts)
	}

	byService := accountsByService(t, repo, animeID)
	if _, ok := byService[model.AnimeAccountServiceX]; ok {
		t.Error("管理対象のxの行が削除されなかった")
	}
	if byService[model.AnimeAccountServiceYoutube] != "editor_youtube" {
		t.Error("編集者が追加したyoutubeのアカウントが保持されなかった")
	}
}
