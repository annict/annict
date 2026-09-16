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

// newSyncAnimeHashtagsUsecaseは共有テストDB上にリコンサイラとそのリポジトリを
// 組み立てる。本UseCaseは自前でトランザクションを開くため、テストは前提データ (animes) を
// アウターtxで包まずGetTestDB経由で直接コミットする。
func newSyncAnimeHashtagsUsecase(db *sql.DB) (*SyncAnimeHashtagsUsecase, *repository.AnimeHashtagRepository) {
	repo := repository.NewAnimeHashtagRepository(query.New(db))
	return NewSyncAnimeHashtagsUsecase(db, repo), repo
}

// workForHashtagSyncはハッシュタグリコンサイラが読むカラム (twitter_hashtag) だけを
// 持つanime解決済みのworkを組み立てる。
func workForHashtagSync(animeID model.AnimeID, twitterHashtag *string) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, TwitterHashtag: twitterHashtag}
}

// hashtagsOfはanimeのハッシュタグをタグ値の素のスライスとして読み戻す。
func hashtagsOf(t *testing.T, repo *repository.AnimeHashtagRepository, animeID model.AnimeID) []string {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	tags := make([]string, len(rows))
	for i, r := range rows {
		tags[i] = r.Hashtag
	}
	return tags
}

func TestSyncAnimeHashtagsUsecase_Reconcile_CreatesRowFromWorkColumn(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeHashtagsUsecase(db)

	animeID := insertBareAnime(t, db)
	work := workForHashtagSync(animeID, strPtr("rezero"))

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 1 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Createdのみ1", counts)
	}

	if got := hashtagsOf(t, repo, animeID); len(got) != 1 || got[0] != "rezero" {
		t.Errorf("hashtags = %v、期待値 = [rezero]", got)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeHashtagsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForHashtagSync(animeID, strPtr("rezero"))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が依拠する
	// 不変条件 (同期済みのページは差分ゼロを報告する)。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 1 {
		t.Errorf("counts = %+v、期待値 = Unchangedのみ1", counts)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_ReplacesChangedHashtag(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeHashtagsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, strPtr("old_tag"))}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// hashtagは自然キーのため、タグの変更は更新ではなく削除 (old) + 作成 (new) になる。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, strPtr("new_tag"))})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Created != 1 || counts.Deleted != 1 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Created 1とDeleted 1", counts)
	}

	if got := hashtagsOf(t, repo, animeID); len(got) != 1 || got[0] != "new_tag" {
		t.Errorf("置き換え後のhashtags = %v、期待値 = [new_tag]", got)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeHashtagsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, strPtr("rezero"))}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// twitter_hashtagが消えた (NULL)。works管理下の行は削除されるべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, nil)})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Deletedのみ1", counts)
	}

	if got := hashtagsOf(t, repo, animeID); len(got) != 0 {
		t.Errorf("同期元の削除後のhashtags = %v、期待値 = 0件", got)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_TreatsNullAndEmptyAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeHashtagsUsecase(db)

	// NULLのtwitter_hashtagと空文字列のtwitter_hashtagはどちらも「欠損」:
	// どちらも行は作られない。
	nullAnimeID := insertBareAnime(t, db)
	emptyAnimeID := insertBareAnime(t, db)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{
		workForHashtagSync(nullAnimeID, nil),
		workForHashtagSync(emptyAnimeID, strPtr("")),
	})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = すべて0", counts)
	}

	if got := hashtagsOf(t, repo, nullAnimeID); len(got) != 0 {
		t.Errorf("同期元がNULLのときのhashtags = %v、期待値 = 0件", got)
	}
	if got := hashtagsOf(t, repo, emptyAnimeID); len(got) != 0 {
		t.Errorf("同期元が空のときのhashtags = %v、期待値 = 0件", got)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeHashtagsUsecase(db)

	animeID := insertBareAnime(t, db)

	// works管理下のハッシュタグ (sort_number 0、リポジトリ経由で作成) と編集者追加の
	// ハッシュタグ (sort_number 1、worksは非ゼロのスロットをsourceしないため直接挿入) を
	// 用意する。anime_hashtagsにはkind / serviceの判別列が無いため、sort_number 0の
	// スロットが「works管理下 (= 削除対象)」であることを示す。
	if _, err := repo.Create(context.Background(), repository.CreateAnimeHashtagParams{AnimeID: animeID, Hashtag: "managed_tag"}); err != nil {
		t.Fatalf("シードの管理対象のCreate()のエラー = %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO anime_hashtags (anime_id, hashtag, sort_number, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())`,
		int64(animeID), "editor_tag", 1,
	); err != nil {
		t.Fatalf("シードの編集者追加行のINSERTのエラー = %v", err)
	}

	// workは何もsourceしないため、管理下のsort_number 0の行は削除されるが、works
	// 管理下のスロットの外にある編集者追加のsort_number 1の行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, nil)})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Deletedのみ1", counts)
	}

	if got := hashtagsOf(t, repo, animeID); len(got) != 1 || got[0] != "editor_tag" {
		t.Errorf("hashtags = %v、期待値 = [editor_tag]が保持されること", got)
	}
}
