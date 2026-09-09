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

// newSyncAnimeHashtagsUsecase builds the reconciler and its repository over the shared
// test DB. The usecase opens its own transaction, so the test commits its setup (animes)
// directly via GetTestDB rather than wrapping it in an outer tx.
//
// [Ja] newSyncAnimeHashtagsUsecase は共有テスト DB 上にリコンサイラとそのリポジトリを
// 組み立てる。本 UseCase は自前でトランザクションを開くため、テストは前提データ (animes) を
// アウター tx で包まず GetTestDB 経由で直接コミットする。
func newSyncAnimeHashtagsUsecase(db *sql.DB) (*SyncAnimeHashtagsUsecase, *repository.AnimeHashtagRepository) {
	repo := repository.NewAnimeHashtagRepository(query.New(db))
	return NewSyncAnimeHashtagsUsecase(db, repo), repo
}

// workForHashtagSync builds an anime-resolved work carrying only the column the hashtag
// reconciler reads (twitter_hashtag).
//
// [Ja] workForHashtagSync はハッシュタグリコンサイラが読むカラム (twitter_hashtag) だけを
// 持つ anime 解決済みの work を組み立てる。
func workForHashtagSync(animeID model.AnimeID, twitterHashtag *string) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, TwitterHashtag: twitterHashtag}
}

// hashtagsOf reads back an anime's hashtags as a plain slice of tag values.
//
// [Ja] hashtagsOf は anime のハッシュタグをタグ値の素のスライスとして読み戻す。
func hashtagsOf(t *testing.T, repo *repository.AnimeHashtagRepository, animeID model.AnimeID) []string {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
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
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 1 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Created 1 only", counts)
	}

	if got := hashtagsOf(t, repo, animeID); len(got) != 1 || got[0] != "rezero" {
		t.Errorf("hashtags = %v, want [rezero]", got)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeHashtagsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForHashtagSync(animeID, strPtr("rezero"))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// Re-running with the same source must detect no diff and write nothing. This is the
	// invariant the cutover decision depends on (a synced page reports zero diff).
	//
	// [Ja] 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が依拠する
	// 不変条件 (同期済みのページは差分ゼロを報告する)。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 1 {
		t.Errorf("counts = %+v, want Unchanged 1 only", counts)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_ReplacesChangedHashtag(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeHashtagsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, strPtr("old_tag"))}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// The hashtag is the natural key, so a changed tag is a delete (old) plus a create
	// (new), not an in-place update.
	//
	// [Ja] hashtag は自然キーのため、タグの変更は更新ではなく削除 (old) + 作成 (new) になる。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, strPtr("new_tag"))})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Created != 1 || counts.Deleted != 1 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Created 1 and Deleted 1", counts)
	}

	if got := hashtagsOf(t, repo, animeID); len(got) != 1 || got[0] != "new_tag" {
		t.Errorf("hashtags = %v, want [new_tag] after replace", got)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeHashtagsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, strPtr("rezero"))}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// twitter_hashtag is gone (NULL); the works-managed row should be deleted.
	//
	// [Ja] twitter_hashtag が消えた (NULL)。works 管理下の行は削除されるべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, nil)})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Deleted 1 only", counts)
	}

	if got := hashtagsOf(t, repo, animeID); len(got) != 0 {
		t.Errorf("hashtags = %v, want none after source removed", got)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_TreatsNullAndEmptyAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeHashtagsUsecase(db)

	// A NULL twitter_hashtag and an empty-string twitter_hashtag are both "absent": no
	// row is created for either.
	//
	// [Ja] NULL の twitter_hashtag と空文字列の twitter_hashtag はどちらも「欠損」:
	// どちらも行は作られない。
	nullAnimeID := insertBareAnime(t, db)
	emptyAnimeID := insertBareAnime(t, db)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{
		workForHashtagSync(nullAnimeID, nil),
		workForHashtagSync(emptyAnimeID, strPtr("")),
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want all zero", counts)
	}

	if got := hashtagsOf(t, repo, nullAnimeID); len(got) != 0 {
		t.Errorf("hashtags = %v, want none for NULL source", got)
	}
	if got := hashtagsOf(t, repo, emptyAnimeID); len(got) != 0 {
		t.Errorf("hashtags = %v, want none for empty source", got)
	}
}

func TestSyncAnimeHashtagsUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeHashtagsUsecase(db)

	animeID := insertBareAnime(t, db)

	// Seed a works-managed hashtag (sort_number 0, created via the repo) and an
	// editor-added one (sort_number 1, inserted directly since works never source a
	// non-zero slot). anime_hashtags has no kind / service discriminator, so the
	// sort_number 0 slot is what marks a row as works-managed and thus deletable.
	//
	// [Ja] works 管理下のハッシュタグ (sort_number 0、リポジトリ経由で作成) と編集者追加の
	// ハッシュタグ (sort_number 1、works は非ゼロのスロットを source しないため直接挿入) を
	// 用意する。anime_hashtags には kind / service の判別列が無いため、sort_number 0 の
	// スロットが「works 管理下 (= 削除対象)」であることを示す。
	if _, err := repo.Create(context.Background(), repository.CreateAnimeHashtagParams{AnimeID: animeID, Hashtag: "managed_tag"}); err != nil {
		t.Fatalf("seed managed Create() error = %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO anime_hashtags (anime_id, hashtag, sort_number, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())`,
		int64(animeID), "editor_tag", 1,
	); err != nil {
		t.Fatalf("seed editor INSERT error = %v", err)
	}

	// The work sources nothing, so the managed sort_number 0 row is deleted while the
	// editor-added sort_number 1 row outside the works-managed slot is preserved.
	//
	// [Ja] work は何も source しないため、管理下の sort_number 0 の行は削除されるが、works
	// 管理下のスロットの外にある編集者追加の sort_number 1 の行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForHashtagSync(animeID, nil)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Deleted 1 only", counts)
	}

	if got := hashtagsOf(t, repo, animeID); len(got) != 1 || got[0] != "editor_tag" {
		t.Errorf("hashtags = %v, want [editor_tag] preserved", got)
	}
}
