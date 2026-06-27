package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedExternalServices is the set of anime_external_ids services that works
// are the source of: sc_tid maps to syobocal and mal_anime_id maps to mal. The
// reconcile limits deletes to rows in this key space, so a row an editor might add
// for some future service outside it is never clobbered (see reconcileSatellite).
// Today every service in the enum is works-sourced, so the predicate currently
// always holds; it exists to keep the delete limit correct as services are added.
//
// [Ja] workSourcedExternalServices は works が source とする anime_external_ids の
// サービス集合。sc_tid が syobocal に、mal_anime_id が mal に写像される。リコンサイルは
// 削除をこのキー空間の行に限定するため、編集者が将来このキー空間の外のサービスに足しうる
// 行が壊されない (reconcileSatellite を参照)。現状 enum のすべてのサービスが works 由来の
// ため本判定は常に真。サービスが増えても削除限定を正しく保つために存在する。
var workSourcedExternalServices = map[model.AnimeExternalService]bool{
	model.AnimeExternalServiceSyobocal: true,
	model.AnimeExternalServiceMal:      true,
}

// animeExternalIDKey is the natural key reconcileSatellite matches desired rows
// (derived from works) against existing rows on: (anime_id, service). It includes
// the anime_id because a reconciler diffs a whole page of works at once, so a key
// without it would collide across works.
//
// [Ja] animeExternalIDKey は reconcileSatellite が (works から導出した) あるべき行を
// 既存行と突合する自然キー (anime_id, service)。リコンサイラは works のページ全体を一括で
// 突合するため、anime_id を含めないと works をまたいでキーが衝突する。
type animeExternalIDKey struct {
	animeID model.AnimeID
	service model.AnimeExternalService
}

// SyncAnimeExternalIDsUsecase is the phase 2 satellite reconciler for the
// anime_external_ids table. Registered into SyncWorkSatellitesUsecase, it maps each
// anime-resolved work's sc_tid / mal_anime_id columns onto the syobocal / mal
// external-ID rows of its anime, then creates / updates / deletes the rows so they
// match. It is a self-contained write usecase mirroring SyncWorksToAnimesUsecase:
// it reads the existing rows, plans the diff via reconcileSatellite, and applies
// the plan in its own transaction. The Reconcile method (not Execute) satisfies the
// satelliteReconciler interface.
//
// [Ja] SyncAnimeExternalIDsUsecase は anime_external_ids テーブルに対するフェーズ 2 の
// 別表リコンサイラ。SyncWorkSatellitesUsecase に登録され、anime 解決済みの各 work の
// sc_tid / mal_anime_id カラムをその anime の syobocal / mal の外部 ID 行に写像し、一致する
// よう行を作成 / 更新 / 削除する。SyncWorksToAnimesUsecase を写した自己完結の書き込み
// UseCase で、既存行を読み、reconcileSatellite で差分を計画し、自前のトランザクションで適用
// する。Execute ではなく Reconcile メソッドが satelliteReconciler インターフェースを満たす。
type SyncAnimeExternalIDsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeExternalIDRepository
}

// NewSyncAnimeExternalIDsUsecase constructs a SyncAnimeExternalIDsUsecase.
//
// [Ja] NewSyncAnimeExternalIDsUsecase は SyncAnimeExternalIDsUsecase を生成する。
func NewSyncAnimeExternalIDsUsecase(db *sql.DB, repo *repository.AnimeExternalIDRepository) *SyncAnimeExternalIDsUsecase {
	return &SyncAnimeExternalIDsUsecase{db: db, repo: repo}
}

// Reconcile reconciles the anime_external_ids rows for the given anime-resolved
// works. Following the write-usecase rule, the desired rows are derived and the
// existing rows are read before applyPlan opens a transaction; the transaction
// performs persistence only.
//
// [Ja] Reconcile は指定された anime 解決済み works について anime_external_ids 行を
// リコンサイルする。書き込み UseCase のルールに従い、あるべき行の導出と既存行の取得は
// applyPlan がトランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeExternalIDsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	desired := desiredAnimeExternalIDs(works)

	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存 anime_external_ids の取得に失敗: %w", err)
	}

	plan := reconcileSatellite(
		desired,
		existing,
		func(d repository.CreateAnimeExternalIDParams) animeExternalIDKey {
			return animeExternalIDKey{animeID: d.AnimeID, service: d.Service}
		},
		func(e *model.AnimeExternalID) animeExternalIDKey {
			return animeExternalIDKey{animeID: e.AnimeID, service: e.Service}
		},
		func(e *model.AnimeExternalID) bool { return workSourcedExternalServices[e.Service] },
		func(d repository.CreateAnimeExternalIDParams, e *model.AnimeExternalID) bool {
			return e.ExternalID != d.ExternalID
		},
	)

	return uc.applyPlan(ctx, plan)
}

// desiredAnimeExternalIDs derives the external-ID rows a batch of works should
// have: sc_tid -> syobocal, mal_anime_id -> mal. A source that is NULL or 0 yields
// no row (per the migration inventory), so an empty work contributes nothing and
// any existing row for it is later deleted.
//
// [Ja] desiredAnimeExternalIDs は works のバッチが持つべき外部 ID 行を導出する。
// sc_tid -> syobocal、mal_anime_id -> mal。ソースが NULL または 0 のときは (移行
// インベントリのとおり) 行を作らないため、空の work は何も寄与せず、既存行があれば後段で
// 削除される。
func desiredAnimeExternalIDs(works []*model.Work) []repository.CreateAnimeExternalIDParams {
	desired := make([]repository.CreateAnimeExternalIDParams, 0, len(works))
	for _, w := range works {
		if w.AnimeID == nil {
			continue
		}
		if externalID := externalIDFromInt32Ptr(w.ScTid); externalID != "" {
			desired = append(desired, repository.CreateAnimeExternalIDParams{
				AnimeID:    *w.AnimeID,
				Service:    model.AnimeExternalServiceSyobocal,
				ExternalID: externalID,
			})
		}
		if externalID := externalIDFromInt32Ptr(w.MalAnimeID); externalID != "" {
			desired = append(desired, repository.CreateAnimeExternalIDParams{
				AnimeID:    *w.AnimeID,
				Service:    model.AnimeExternalServiceMal,
				ExternalID: externalID,
			})
		}
	}
	return desired
}

// externalIDFromInt32Ptr stringifies a works integer external-id column, mapping
// both NULL (nil) and 0 to "" (no row). Rails stores an absent external ID as NULL
// or 0, so both are treated as missing here.
//
// [Ja] externalIDFromInt32Ptr は works の integer 外部 ID カラムを文字列化し、NULL (nil)
// と 0 の両方を "" (行なし) に写像する。Rails は欠損の外部 ID を NULL または 0 で持つため、
// ここではどちらも欠損として扱う。
func externalIDFromInt32Ptr(p *int32) string {
	if p == nil || *p == 0 {
		return ""
	}
	return strconv.FormatInt(int64(*p), 10)
}

// applyPlan persists the reconcile plan in a single transaction and returns the
// per-table counts. It returns early without opening a transaction when there is
// nothing to write, so an already-synced page costs no write.
//
// [Ja] applyPlan はリコンサイル計画を 1 トランザクションで永続化し、テーブルごとの件数を
// 返す。書き込むものが無ければトランザクションを開かずに早期 return するため、既に同期済みの
// ページは書き込みコストがかからない。
func (uc *SyncAnimeExternalIDsUsecase) applyPlan(ctx context.Context, plan satelliteReconcilePlan[repository.CreateAnimeExternalIDParams, *model.AnimeExternalID]) (satelliteReconcileCounts, error) {
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
			return satelliteReconcileCounts{}, fmt.Errorf("anime_external_id の作成に失敗 (anime_id=%d, service=%s): %w", create.AnimeID, create.Service, err)
		}
		counts.Created++
	}

	for _, update := range plan.updates {
		if err := repo.Update(ctx, repository.UpdateAnimeExternalIDParams{
			ID:         update.existing.ID,
			ExternalID: update.desired.ExternalID,
		}); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_external_id の更新に失敗 (id=%d): %w", update.existing.ID, err)
		}
		counts.Updated++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_external_id の削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}
