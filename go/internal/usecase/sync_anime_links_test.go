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

// newSyncAnimeLinksUsecaseは共有テストDB上にリコンサイラとそのリポジトリを
// 組み立てる。本UseCaseは自前でトランザクションを開くため、テストは前提データ (animes) を
// アウターtxで包まずGetTestDB経由で直接コミットする。
func newSyncAnimeLinksUsecase(db *sql.DB) (*SyncAnimeLinksUsecase, *repository.AnimeLinkRepository) {
	repo := repository.NewAnimeLinkRepository(query.New(db))
	return NewSyncAnimeLinksUsecase(db, repo), repo
}

// workForLinkSyncはリンクリコンサイラが読むカラム (official_site / wikipediaの
// 4つのurlカラム) だけを持つanime解決済みのworkを組み立てる。
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

// linkKeyは1つのanime内でリンクを (kind, language) で識別する (アサーション用)。
type linkKey struct {
	kind     model.AnimeLinkKind
	language model.Language
}

// linksByKeyはanimeのリンクを (kind, language) をキーに読み戻す。
func linksByKey(t *testing.T, repo *repository.AnimeLinkRepository, animeID model.AnimeID) map[linkKey]string {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
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
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 4 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Createdのみ4", counts)
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
			t.Errorf("link %+v = %q、期待値 = %q", k, byKey[k], wantURL)
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
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が
	// 依拠する不変条件 (同期済みのページは差分ゼロを報告する)。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 2 {
		t.Errorf("counts = %+v、期待値 = Unchangedのみ2", counts)
	}
}

func TestSyncAnimeLinksUsecase_Reconcile_UpdatesChangedURL(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "https://example.dev/old", "", "", "")}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "https://example.dev/new", "", "", "")})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Updatedのみ1", counts)
	}

	if got := linksByKey(t, repo, animeID)[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}]; got != "https://example.dev/new" {
		t.Errorf("更新後のofficial_site/ja = %q、期待値 = https://example.dev/new", got)
	}
}

func TestSyncAnimeLinksUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "https://example.dev/o", "", "https://example.dev/w", "")}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// wikipediaのurlが消えた (空)。official_site/jaの行だけが残るべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "https://example.dev/o", "", "", "")})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Unchanged != 1 || counts.Created != 0 || counts.Updated != 0 {
		t.Errorf("counts = %+v、期待値 = Deleted 1 / Unchanged 1", counts)
	}

	byKey := linksByKey(t, repo, animeID)
	if _, ok := byKey[linkKey{model.AnimeLinkKindWikipedia, model.LanguageJa}]; ok {
		t.Error("wikipedia/jaの行が削除されなかった")
	}
	if byKey[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}] != "https://example.dev/o" {
		t.Errorf("official_site/ja = %q、期待値 = https://example.dev/o (据え置き)", byKey[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}])
	}
}

func TestSyncAnimeLinksUsecase_Reconcile_TreatsEmptyURLAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)
	// 4つのurlカラムがすべて空: 行は作られない。
	work := workForLinkSync(animeID, "", "", "", "")

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = すべて0", counts)
	}

	if rows := linksByKey(t, repo, animeID); len(rows) != 0 {
		t.Errorf("urlが空のときのrows = %v、期待値 = 0件", rows)
	}
}

func TestSyncAnimeLinksUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeLinksUsecase(db)

	animeID := insertBareAnime(t, db)

	// 既存行を3つ用意する: works管理下の1つ (official_site/ja) と、works管理下の
	// キー空間の外の2つ (編集者が直接足しうるkind=otherのリンクとlanguage=otherのリンク)。
	seed := []repository.CreateAnimeLinkParams{
		{AnimeID: animeID, Kind: model.AnimeLinkKindOfficialSite, Language: model.LanguageJa, URL: "https://example.dev/managed"},
		{AnimeID: animeID, Kind: model.AnimeLinkKindOther, Language: model.LanguageJa, URL: "https://example.dev/other-kind"},
		{AnimeID: animeID, Kind: model.AnimeLinkKindOfficialSite, Language: model.LanguageOther, URL: "https://example.dev/other-lang"},
	}
	for _, s := range seed {
		if _, err := repo.Create(context.Background(), s); err != nil {
			t.Fatalf("シードのCreate(%+v)のエラー = %v", s, err)
		}
	}

	// workは何もsourceしないため、管理下のofficial_site/jaは削除されるが、キー空間の
	// 外の編集者追加行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForLinkSync(animeID, "", "", "", "")})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Deletedのみ1", counts)
	}

	byKey := linksByKey(t, repo, animeID)
	if _, ok := byKey[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}]; ok {
		t.Error("管理対象のofficial_site/jaの行が削除されなかった")
	}
	if byKey[linkKey{model.AnimeLinkKindOther, model.LanguageJa}] != "https://example.dev/other-kind" {
		t.Error("編集者が追加したkind=otherのリンクが保持されなかった")
	}
	if byKey[linkKey{model.AnimeLinkKindOfficialSite, model.LanguageOther}] != "https://example.dev/other-lang" {
		t.Error("編集者が追加したlanguage=otherのリンクが保持されなかった")
	}
}
