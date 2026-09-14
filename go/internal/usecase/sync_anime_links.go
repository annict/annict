package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedLinkKindsはworksがsourceとするanime_linkのkind集合
// (official_site / wikipedia)。otherは編集者が直接足すリンク向けに予約する。
var workSourcedLinkKinds = map[model.AnimeLinkKind]bool{
	model.AnimeLinkKindOfficialSite: true,
	model.AnimeLinkKindWikipedia:    true,
}

// workSourcedLinkLanguagesはworksがsourceとする言語集合 (ja / en)。otherは
// 編集者が直接足すリンク向けに予約する。
var workSourcedLinkLanguages = map[model.Language]bool{
	model.LanguageJa: true,
	model.LanguageEn: true,
}

// animeLinkKeyはreconcileSatelliteが (worksから導出した) あるべき行を既存行と
// 突合する自然キー (anime_id, kind, language)。リコンサイラはworksのページ全体を一括で
// 突合するため、anime_idを含めないとworksをまたいでキーが衝突する。
type animeLinkKey struct {
	animeID  model.AnimeID
	kind     model.AnimeLinkKind
	language model.Language
}

// SyncAnimeLinksUsecaseはanime_linksテーブルに対するフェーズ2の別表リコンサイラ。
// SyncWorkSatellitesUsecaseに登録され、anime解決済みの各workの4つのリンクカラム
// (official_site_url / official_site_url_en / wikipedia_url / wikipedia_url_en) を
// そのanimeの (kind, language) リンク行に写像し、一致するよう行を作成 / 更新 / 削除する。
// SyncWorksToAnimesUsecaseを写した自己完結の書き込みUseCaseで、既存行を読み、
// reconcileSatelliteで差分を計画し、自前のトランザクションで適用する。Executeではなく
// ReconcileメソッドがsatelliteReconcilerインターフェースを満たす。
type SyncAnimeLinksUsecase struct {
	db   *sql.DB
	repo *repository.AnimeLinkRepository
}

// NewSyncAnimeLinksUsecaseはSyncAnimeLinksUsecaseを生成する。
func NewSyncAnimeLinksUsecase(db *sql.DB, repo *repository.AnimeLinkRepository) *SyncAnimeLinksUsecase {
	return &SyncAnimeLinksUsecase{db: db, repo: repo}
}

// Reconcileは指定されたanime解決済みworksについてanime_links行をリコンサイルする。
// 書き込みUseCaseのルールに従い、あるべき行の導出と既存行の取得はapplyPlanが
// トランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeLinksUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存anime_linksの取得に失敗: %w", err)
	}

	return uc.applyPlan(ctx, planAnimeLinks(works, existing))
}

// planAnimeLinksは指定されたanime解決済みworksのanime_links行について、
// 既存行に対するリコンサイル計画を組み立てる。フェーズ2のバッチリコンサイラ (Reconcile) と
// フェーズ3の作品 作成 / 更新 の両書き (planWorkSatellites) で共有し、あるべき行の導出・
// 自然キー・削除限定・変更検出を両経路で単一の正本に保つ。
func planAnimeLinks(works []*model.Work, existing []*model.AnimeLink) satelliteReconcilePlan[repository.CreateAnimeLinkParams, *model.AnimeLink] {
	return reconcileSatellite(
		desiredAnimeLinks(works),
		existing,
		func(d repository.CreateAnimeLinkParams) animeLinkKey {
			return animeLinkKey{animeID: d.AnimeID, kind: d.Kind, language: d.Language}
		},
		func(e *model.AnimeLink) animeLinkKey {
			return animeLinkKey{animeID: e.AnimeID, kind: e.Kind, language: e.Language}
		},
		// 削除をworks管理下のキー空間の行に限定する: kindが {official_site,
		// wikipedia} かつlanguageが {ja, en}。その外の行 (例: kind=otherや
		// language=other) は編集者が足したものなので決して削除しない。
		func(e *model.AnimeLink) bool {
			return workSourcedLinkKinds[e.Kind] && workSourcedLinkLanguages[e.Language]
		},
		// worksはURLのみをsourceとするため、URLが異なる場合だけ変更扱いにする。
		// label / label_enは触らず、編集者の編集を保全する。
		func(d repository.CreateAnimeLinkParams, e *model.AnimeLink) bool {
			return e.URL != d.URL
		},
	)
}

// desiredAnimeLinksはworksのバッチが持つべきリンク行を、空でないソースURLごとに
// 1行導出する。official_site_url -> (official_site, ja)、official_site_url_en ->
// (official_site, en)、wikipedia_url -> (wikipedia, ja)、wikipedia_url_en ->
// (wikipedia, en)。urlカラムはNOT NULLで空文字列が既定値のため、空文字列は「リンクなし」を
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

// appendDesiredLinkは空でないurlについてあるべきリンクを追加し、空のurl
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

// applyPlanはリコンサイル計画を1トランザクションで永続化し、テーブルごとの件数を
// 返す。書き込むものが無ければトランザクションを開かずに早期returnするため、既に同期済みの
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
			return satelliteReconcileCounts{}, fmt.Errorf("anime_linkの作成に失敗 (anime_id=%d, kind=%s, language=%s): %w", create.AnimeID, create.Kind, create.Language, err)
		}
		counts.Created++
	}

	for _, update := range plan.updates {
		if err := repo.Update(ctx, repository.UpdateAnimeLinkParams{
			ID:  update.existing.ID,
			URL: update.desired.URL,
		}); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_linkの更新に失敗 (id=%d): %w", update.existing.ID, err)
		}
		counts.Updated++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_linkの削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}
