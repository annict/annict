package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedLinkKinds is the set of anime_link kinds works are the source of:
// official_site and wikipedia. other is reserved for links an editor adds directly.
//
// [Ja] workSourcedLinkKinds は works が source とする anime_link の kind 集合
// (official_site / wikipedia)。other は編集者が直接足すリンク向けに予約する。
var workSourcedLinkKinds = map[model.AnimeLinkKind]bool{
	model.AnimeLinkKindOfficialSite: true,
	model.AnimeLinkKindWikipedia:    true,
}

// workSourcedLinkLanguages is the set of languages works are the source of: ja and en.
// other is reserved for links an editor adds directly.
//
// [Ja] workSourcedLinkLanguages は works が source とする言語集合 (ja / en)。other は
// 編集者が直接足すリンク向けに予約する。
var workSourcedLinkLanguages = map[model.Language]bool{
	model.LanguageJa: true,
	model.LanguageEn: true,
}

// animeLinkKey is the natural key reconcileSatellite matches desired rows (derived
// from works) against existing rows on: (anime_id, kind, language). It includes the
// anime_id because a reconciler diffs a whole page of works at once, so a key without
// it would collide across works.
//
// [Ja] animeLinkKey は reconcileSatellite が (works から導出した) あるべき行を既存行と
// 突合する自然キー (anime_id, kind, language)。リコンサイラは works のページ全体を一括で
// 突合するため、anime_id を含めないと works をまたいでキーが衝突する。
type animeLinkKey struct {
	animeID  model.AnimeID
	kind     model.AnimeLinkKind
	language model.Language
}

// SyncAnimeLinksUsecase is the phase 2 satellite reconciler for the anime_links table.
// Registered into SyncWorkSatellitesUsecase, it maps each anime-resolved work's four
// link columns (official_site_url / official_site_url_en / wikipedia_url /
// wikipedia_url_en) onto the (kind, language) link rows of its anime, then creates /
// updates / deletes the rows so they match. It is a self-contained write usecase
// mirroring SyncWorksToAnimesUsecase: it reads the existing rows, plans the diff via
// reconcileSatellite, and applies the plan in its own transaction. The Reconcile
// method (not Execute) satisfies the satelliteReconciler interface.
//
// [Ja] SyncAnimeLinksUsecase は anime_links テーブルに対するフェーズ 2 の別表リコンサイラ。
// SyncWorkSatellitesUsecase に登録され、anime 解決済みの各 work の 4 つのリンクカラム
// (official_site_url / official_site_url_en / wikipedia_url / wikipedia_url_en) を
// その anime の (kind, language) リンク行に写像し、一致するよう行を作成 / 更新 / 削除する。
// SyncWorksToAnimesUsecase を写した自己完結の書き込み UseCase で、既存行を読み、
// reconcileSatellite で差分を計画し、自前のトランザクションで適用する。Execute ではなく
// Reconcile メソッドが satelliteReconciler インターフェースを満たす。
type SyncAnimeLinksUsecase struct {
	db   *sql.DB
	repo *repository.AnimeLinkRepository
}

// NewSyncAnimeLinksUsecase constructs a SyncAnimeLinksUsecase.
//
// [Ja] NewSyncAnimeLinksUsecase は SyncAnimeLinksUsecase を生成する。
func NewSyncAnimeLinksUsecase(db *sql.DB, repo *repository.AnimeLinkRepository) *SyncAnimeLinksUsecase {
	return &SyncAnimeLinksUsecase{db: db, repo: repo}
}

// Reconcile reconciles the anime_links rows for the given anime-resolved works.
// Following the write-usecase rule, the desired rows are derived and the existing rows
// are read before applyPlan opens a transaction; the transaction performs persistence
// only.
//
// [Ja] Reconcile は指定された anime 解決済み works について anime_links 行をリコンサイルする。
// 書き込み UseCase のルールに従い、あるべき行の導出と既存行の取得は applyPlan が
// トランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeLinksUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	desired := desiredAnimeLinks(works)

	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存 anime_links の取得に失敗: %w", err)
	}

	plan := reconcileSatellite(
		desired,
		existing,
		func(d repository.CreateAnimeLinkParams) animeLinkKey {
			return animeLinkKey{animeID: d.AnimeID, kind: d.Kind, language: d.Language}
		},
		func(e *model.AnimeLink) animeLinkKey {
			return animeLinkKey{animeID: e.AnimeID, kind: e.Kind, language: e.Language}
		},
		// Limit deletes to rows in the works-managed key space: kind in
		// {official_site, wikipedia} AND language in {ja, en}. A row outside it
		// (e.g. kind=other or language=other) is editor-added and never deleted.
		//
		// [Ja] 削除を works 管理下のキー空間の行に限定する: kind が {official_site,
		// wikipedia} かつ language が {ja, en}。その外の行 (例: kind=other や
		// language=other) は編集者が足したものなので決して削除しない。
		func(e *model.AnimeLink) bool {
			return workSourcedLinkKinds[e.Kind] && workSourcedLinkLanguages[e.Language]
		},
		// Works only source the URL, so a row is changed iff its URL differs. label /
		// label_en are left untouched, preserving any editor edits to them.
		//
		// [Ja] works は URL のみを source とするため、URL が異なる場合だけ変更扱いにする。
		// label / label_en は触らず、編集者の編集を保全する。
		func(d repository.CreateAnimeLinkParams, e *model.AnimeLink) bool {
			return e.URL != d.URL
		},
	)

	return uc.applyPlan(ctx, plan)
}

// desiredAnimeLinks derives the link rows a batch of works should have, one per
// non-empty source URL: official_site_url -> (official_site, ja), official_site_url_en
// -> (official_site, en), wikipedia_url -> (wikipedia, ja), wikipedia_url_en ->
// (wikipedia, en). The url columns are NOT NULL with an empty-string default, so an
// empty string means "no link" and yields no row; any existing row for it is later
// deleted.
//
// [Ja] desiredAnimeLinks は works のバッチが持つべきリンク行を、空でないソース URL ごとに
// 1 行導出する。official_site_url -> (official_site, ja)、official_site_url_en ->
// (official_site, en)、wikipedia_url -> (wikipedia, ja)、wikipedia_url_en ->
// (wikipedia, en)。url カラムは NOT NULL で空文字列が既定値のため、空文字列は「リンクなし」を
// 意味して行を作らず、既存行があれば後段で削除される。
func desiredAnimeLinks(works []*model.Work) []repository.CreateAnimeLinkParams {
	desired := make([]repository.CreateAnimeLinkParams, 0, len(works))
	for _, w := range works {
		if w.AnimeID == nil {
			continue
		}
		desired = appendDesiredLink(desired, *w.AnimeID, model.AnimeLinkKindOfficialSite, model.LanguageJa, w.OfficialSiteURL)
		desired = appendDesiredLink(desired, *w.AnimeID, model.AnimeLinkKindOfficialSite, model.LanguageEn, w.OfficialSiteURLEn)
		desired = appendDesiredLink(desired, *w.AnimeID, model.AnimeLinkKindWikipedia, model.LanguageJa, w.WikipediaURL)
		desired = appendDesiredLink(desired, *w.AnimeID, model.AnimeLinkKindWikipedia, model.LanguageEn, w.WikipediaURLEn)
	}
	return desired
}

// appendDesiredLink appends a desired link for a non-empty url, and is a no-op for an
// empty url (the "no link" case).
//
// [Ja] appendDesiredLink は空でない url についてあるべきリンクを追加し、空の url
// (「リンクなし」) では何もしない。
func appendDesiredLink(desired []repository.CreateAnimeLinkParams, animeID model.AnimeID, kind model.AnimeLinkKind, language model.Language, url string) []repository.CreateAnimeLinkParams {
	if url == "" {
		return desired
	}
	return append(desired, repository.CreateAnimeLinkParams{
		AnimeID:  animeID,
		Kind:     kind,
		Language: language,
		URL:      url,
	})
}

// applyPlan persists the reconcile plan in a single transaction and returns the
// per-table counts. It returns early without opening a transaction when there is
// nothing to write, so an already-synced page costs no write.
//
// [Ja] applyPlan はリコンサイル計画を 1 トランザクションで永続化し、テーブルごとの件数を
// 返す。書き込むものが無ければトランザクションを開かずに早期 return するため、既に同期済みの
// ページは書き込みコストがかからない。
func (uc *SyncAnimeLinksUsecase) applyPlan(ctx context.Context, plan satelliteReconcilePlan[repository.CreateAnimeLinkParams, *model.AnimeLink]) (satelliteReconcileCounts, error) {
	counts := satelliteReconcileCounts{Unchanged: plan.unchanged}
	if len(plan.creates) == 0 && len(plan.updates) == 0 && len(plan.deletes) == 0 {
		return counts, nil
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	repo := uc.repo.WithTx(tx)

	for _, create := range plan.creates {
		if _, err := repo.Create(ctx, create); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_link の作成に失敗 (anime_id=%d, kind=%s, language=%s): %w", create.AnimeID, create.Kind, create.Language, err)
		}
		counts.Created++
	}

	for _, update := range plan.updates {
		if err := repo.Update(ctx, repository.UpdateAnimeLinkParams{
			ID:  update.existing.ID,
			URL: update.desired.URL,
		}); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_link の更新に失敗 (id=%d): %w", update.existing.ID, err)
		}
		counts.Updated++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_link の削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}
