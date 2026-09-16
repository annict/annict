package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedExternalServicesはworksがsourceとするanime_external_idsの
// サービス集合。sc_tidがsyobocalに、mal_anime_idがmalに写像される。リコンサイルは
// 削除をこのキー空間の行に限定するため、編集者が将来このキー空間の外のサービスに足しうる
// 行が壊されない (reconcileSatelliteを参照)。現状enumのすべてのサービスがworks由来の
// ため本判定は常に真。サービスが増えても削除限定を正しく保つために存在する。
var workSourcedExternalServices = map[model.AnimeExternalService]bool{
	model.AnimeExternalServiceSyobocal: true,
	model.AnimeExternalServiceMal:      true,
}

// animeExternalIDKeyはreconcileSatelliteが (worksから導出した) あるべき行を
// 既存行と突合する自然キー (anime_id, service)。リコンサイラはworksのページ全体を一括で
// 突合するため、anime_idを含めないとworksをまたいでキーが衝突する。
type animeExternalIDKey struct {
	animeID model.AnimeID
	service model.AnimeExternalService
}

// SyncAnimeExternalIDsUsecaseはanime_external_idsテーブルに対するフェーズ2の
// 別表リコンサイラ。SyncWorkSatellitesUsecaseに登録され、anime解決済みの各workの
// sc_tid / mal_anime_idカラムをそのanimeのsyobocal / malの外部ID行に写像し、一致する
// よう行を作成 / 更新 / 削除する。SyncWorksToAnimesUsecaseを写した自己完結の書き込み
// UseCaseで、既存行を読み、reconcileSatelliteで差分を計画し、自前のトランザクションで適用
// する。ExecuteではなくReconcileメソッドがsatelliteReconcilerインターフェースを満たす。
type SyncAnimeExternalIDsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeExternalIDRepository
}

// NewSyncAnimeExternalIDsUsecaseはSyncAnimeExternalIDsUsecaseを生成する。
func NewSyncAnimeExternalIDsUsecase(db *sql.DB, repo *repository.AnimeExternalIDRepository) *SyncAnimeExternalIDsUsecase {
	return &SyncAnimeExternalIDsUsecase{db: db, repo: repo}
}

// Reconcileは指定されたanime解決済みworksについてanime_external_ids行を
// リコンサイルする。書き込みUseCaseのルールに従い、あるべき行の導出と既存行の取得は
// applyPlanがトランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeExternalIDsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存anime_external_idsの取得に失敗: %w", err)
	}

	return uc.applyPlan(ctx, planAnimeExternalIDs(works, existing))
}

// planAnimeExternalIDsは指定されたanime解決済みworksのanime_external_ids行に
// ついて、既存行に対するリコンサイル計画を組み立てる。フェーズ2のバッチリコンサイラ
// (Reconcile) とフェーズ3の作品 作成 / 更新 の両書き (planWorkSatellites) で共有し、
// あるべき行の導出・自然キー・削除限定・変更検出を両経路で単一の正本に保つ。
func planAnimeExternalIDs(works []*model.Work, existing []*model.AnimeExternalID) satelliteReconcilePlan[repository.CreateAnimeExternalIDParams, *model.AnimeExternalID] {
	return reconcileSatellite(
		desiredAnimeExternalIDs(works),
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
}

// desiredAnimeExternalIDsはworksのバッチが持つべき外部ID行を導出する。
// sc_tid -> syobocal、mal_anime_id -> mal。ソースがNULLまたは0のときは (移行
// インベントリのとおり) 行を作らないため、空のworkは何も寄与せず、既存行があれば後段で
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

// externalIDFromInt32Ptrはworksのinteger外部IDカラムを文字列化し、NULL (nil)
// と0の両方を "" (行なし) に写像する。Railsは欠損の外部IDをNULLまたは0で持つため、
// ここではどちらも欠損として扱う。
func externalIDFromInt32Ptr(p *int32) string {
	if p == nil || *p == 0 {
		return ""
	}
	return strconv.FormatInt(int64(*p), 10)
}

// applyPlanはリコンサイル計画を1トランザクションで永続化し、テーブルごとの件数を
// 返す。書き込むものが無ければトランザクションを開かずに早期returnするため、既に同期済みの
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
			return satelliteReconcileCounts{}, fmt.Errorf("anime_external_idの作成に失敗 (anime_id=%d, service=%s): %w", create.AnimeID, create.Service, err)
		}
		counts.Created++
	}

	for _, update := range plan.updates {
		if err := repo.Update(ctx, repository.UpdateAnimeExternalIDParams{
			ID:         update.existing.ID,
			ExternalID: update.desired.ExternalID,
		}); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_external_idの更新に失敗 (id=%d): %w", update.existing.ID, err)
		}
		counts.Updated++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_external_idの削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}
