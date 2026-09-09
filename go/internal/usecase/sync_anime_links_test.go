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

// newSyncAnimeLinksUsecase builds the reconciler and its repository over the shared
// test DB. The usecase opens its own transaction, so the test commits its setup
// (animes) directly via GetTestDB rather than wrapping it in an outer tx.
//
// [Ja] newSyncAnimeLinksUsecase は共有テスト DB 上にリコンサイラとそのリポジトリを
// 組み立てる。本 UseCase は自前でトランザクションを開くため、テストは前提データ (animes) を
// アウター tx で包まず GetTestDB 経由で直接コミットする。
func newSyncAnimeLinksUsecase(db *sql.DB) (*SyncAnimeLinksUsecase, *repository.AnimeLinkRepository) {
	repo := repository.NewAnimeLinkRepository(query.New(db))
	return NewSyncAnimeLinksUsecase(db, repo), repo
}

// workForLinkSync builds an anime-resolved work carrying only the columns the link
// reconciler reads (the four official_site / wikipedia url columns).
//
// [Ja] workForLinkSync はリンクリコンサイラが読むカラム (official_site / wikipedia の
// 4 つの url カラム) だけを持つ anime 解決済みの work を組み立てる。
func workForLinkSync(animeID model.AnimeID, officialJa, officialEn, wikiJa, wikiEn string) *model.Work {
	aid := animeID
	return &model.Work{
		AnimeID:           &aid,
		OfficialSiteURL:   officialJa,
		OfficialSiteURLEn: officialEn,
		WikipediaURL:      wikiJa,
		WikipediaURLEn:    wikiEn,
	}
}

// linkKey identifies a link by (kind, language) within one anime, for assertions.
//
// [Ja] linkKey は 1 つの anime 内でリンクを (kind, language) で識別する (アサーション用)。
type linkKey struct {
	kind     model.AnimeLinkKind
	language model.Language
}

// linksByKey reads back an anime's links keyed by (kind, language).
//
// [Ja] linksByKey は anime のリンクを (kind, language) をキーに読み戻す。
func linksByKey(t *testing.T, repo *repository.AnimeLinkRepository, animeID model.AnimeID) map[linkKey]string {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	byKey := make(map[linkKey]string, len(rows))
	for _, r := range rows {
		byKey[linkKey{r.Kind, r.Language}] = r.URL
	}
	return byKey
}

func TestSyncAnimeLinksUsecase_Reconcile_CreatesRowsFromWorkColumns(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)
	work := workForLinkSync(animeID,
		"https://example.dev/official",
		"https://example.dev/official-en",
		"https://ja.wikipedia.org/wiki/Example",
		"https://en.wikipedia.org/wiki/Example",
	)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 4 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Created 4 only", counts)
	}

	byKey := linksByKey(t, repo, animeID)
	wantURLs := map[linkKey]string{
		{model.AnimeLinkKindOfficialSite, model.LanguageJa}: "https://example.dev/official",
		{model.AnimeLinkKindOfficialSite, model.LanguageEn}: "https://example.dev/official-en",
		{model.AnimeLinkKindWikipedia, model.LanguageJa}:    "https://ja.wikipedia.org/wiki/Example",
		{model.AnimeLinkKindWikipedia, model.LanguageEn}:    "https://en.wikipedia.org/wiki/Example",
	}
	for k, wantURL := range wantURLs {
		if byKey[k] != wantURL {
			t.Errorf("link %+v = %q, want %q", k, byKey[k], wantURL)
		}
	}
}

func TestSyncAnimeLinksUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForLinkSync(animeID, "https://example.dev/o", "", "https://example.dev/w", "")}

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
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 2 {
		t.Errorf("counts = %+v, want Unchanged 2 only", counts)
	}
}

func TestSyncAnimeLinksUsecase_Reconcile_UpdatesChangedURL(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "https://example.dev/old", "", "", "")}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "https://example.dev/new", "", "", "")})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Updated 1 only", counts)
	}

	if got := linksByKey(t, repo, animeID)[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}]; got != "https://example.dev/new" {
		t.Errorf("official_site/ja = %q, want https://example.dev/new after update", got)
	}
}

func TestSyncAnimeLinksUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "https://example.dev/o", "", "https://example.dev/w", "")}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// The wikipedia url is gone (empty); only the official_site/ja row should survive.
	//
	// [Ja] wikipedia の url が消えた (空)。official_site/ja の行だけが残るべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "https://example.dev/o", "", "", "")})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Unchanged != 1 || counts.Created != 0 || counts.Updated != 0 {
		t.Errorf("counts = %+v, want Deleted 1 / Unchanged 1", counts)
	}

	byKey := linksByKey(t, repo, animeID)
	if _, ok := byKey[linkKey{model.AnimeLinkKindWikipedia, model.LanguageJa}]; ok {
		t.Error("wikipedia/ja row should have been deleted")
	}
	if byKey[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}] != "https://example.dev/o" {
		t.Errorf("official_site/ja = %q, want https://example.dev/o (kept)", byKey[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}])
	}
}

func TestSyncAnimeLinksUsecase_Reconcile_TreatsEmptyURLAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)
	// All four url columns are empty: no rows are created.
	//
	// [Ja] 4 つの url カラムがすべて空: 行は作られない。
	work := workForLinkSync(animeID, "", "", "", "")

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want all zero", counts)
	}

	if rows := linksByKey(t, repo, animeID); len(rows) != 0 {
		t.Errorf("rows = %v, want none for empty urls", rows)
	}
}

func TestSyncAnimeLinksUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)

	// Seed three existing rows: one works-managed (official_site/ja) and two outside
	// the works-managed key space (a kind=other link and a language=other link an
	// editor might add directly).
	//
	// [Ja] 既存行を 3 つ用意する: works 管理下の 1 つ (official_site/ja) と、works 管理下の
	// キー空間の外の 2 つ (編集者が直接足しうる kind=other のリンクと language=other のリンク)。
	seed := []repository.CreateAnimeLinkParams{
		{AnimeID: animeID, Kind: model.AnimeLinkKindOfficialSite, Language: model.LanguageJa, URL: "https://example.dev/managed"},
		{AnimeID: animeID, Kind: model.AnimeLinkKindOther, Language: model.LanguageJa, URL: "https://example.dev/other-kind"},
		{AnimeID: animeID, Kind: model.AnimeLinkKindOfficialSite, Language: model.LanguageOther, URL: "https://example.dev/other-lang"},
	}
	for _, s := range seed {
		if _, err := repo.Create(context.Background(), s); err != nil {
			t.Fatalf("seed Create(%+v) error = %v", s, err)
		}
	}

	// The work sources nothing, so the managed official_site/ja row is deleted while
	// the editor-added rows outside the key space are preserved.
	//
	// [Ja] work は何も source しないため、管理下の official_site/ja は削除されるが、キー空間の
	// 外の編集者追加行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "", "", "", "")})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Deleted 1 only", counts)
	}

	byKey := linksByKey(t, repo, animeID)
	if _, ok := byKey[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}]; ok {
		t.Error("managed official_site/ja row should have been deleted")
	}
	if byKey[linkKey{model.AnimeLinkKindOther, model.LanguageJa}] != "https://example.dev/other-kind" {
		t.Error("editor-added kind=other link should be preserved")
	}
	if byKey[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageOther}] != "https://example.dev/other-lang" {
		t.Error("editor-added language=other link should be preserved")
	}
}
