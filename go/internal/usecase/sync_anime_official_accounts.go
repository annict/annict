package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedAccountServicesはworksがsourceとするanime_official_accountsの
// サービス集合。twitter_usernameがxのアカウントに写像される。リコンサイルは削除をこの
// キー空間の行に限定するため、編集者が他のサービス (例: youtube / instagram) に足しうる
// アカウントが壊されない (reconcileSatelliteを参照)。
var workSourcedAccountServices = map[model.AnimeAccountService]bool{
	model.AnimeAccountServiceX: true,
}

// animeOfficialAccountKeyはreconcileSatelliteが (worksから導出した) あるべき行を
// 既存行と突合する自然キー (anime_id, service)。リコンサイラはworksのページ全体を一括で
// 突合するため、anime_idを含めないとworksをまたいでキーが衝突する。
type animeOfficialAccountKey struct {
	animeID model.AnimeID
	service model.AnimeAccountService
}

// SyncAnimeOfficialAccountsUsecaseはanime_official_accountsテーブルに対する
// フェーズ2の別表リコンサイラ。SyncWorkSatellitesUsecaseに登録され、anime解決済みの
// 各workのtwitter_usernameカラムをそのanimeのx公式アカウント行に写像し、一致する
// よう行を作成 / 更新 / 削除する。SyncWorksToAnimesUsecaseを写した自己完結の書き込み
// UseCaseで、既存行を読み、reconcileSatelliteで差分を計画し、自前のトランザクションで
// 適用する。ExecuteではなくReconcileメソッドがsatelliteReconcilerインターフェースを
// 満たす。
type SyncAnimeOfficialAccountsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeOfficialAccountRepository
}

// NewSyncAnimeOfficialAccountsUsecaseはSyncAnimeOfficialAccountsUsecaseを生成する。
func NewSyncAnimeOfficialAccountsUsecase(db *sql.DB, repo *repository.AnimeOfficialAccountRepository) *SyncAnimeOfficialAccountsUsecase {
	return &SyncAnimeOfficialAccountsUsecase{db: db, repo: repo}
}

// Reconcileは指定されたanime解決済みworksについてanime_official_accounts行を
// リコンサイルする。書き込みUseCaseのルールに従い、あるべき行の導出と既存行の取得は
// applyPlanがトランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeOfficialAccountsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存anime_official_accountsの取得に失敗: %w", err)
	}

	return uc.applyPlan(ctx, planAnimeOfficialAccounts(works, existing))
}

// planAnimeOfficialAccountsは指定されたanime解決済みworksの
// anime_official_accounts行について、既存行に対するリコンサイル計画を組み立てる。
// フェーズ2のバッチリコンサイラ (Reconcile) とフェーズ3の作品 作成 / 更新 の両書き
// (planWorkSatellites) で共有し、あるべき行の導出・自然キー・削除限定・変更検出を両経路で
// 単一の正本に保つ。
func planAnimeOfficialAccounts(works []*model.Work, existing []*model.AnimeOfficialAccount) satelliteReconcilePlan[repository.CreateAnimeOfficialAccountParams, *model.AnimeOfficialAccount] {
	return reconcileSatellite(
		desiredAnimeOfficialAccounts(works),
		existing,
		func(d repository.CreateAnimeOfficialAccountParams) animeOfficialAccountKey {
			return animeOfficialAccountKey{animeID: d.AnimeID, service: d.Service}
		},
		func(e *model.AnimeOfficialAccount) animeOfficialAccountKey {
			return animeOfficialAccountKey{animeID: e.AnimeID, service: e.Service}
		},
		func(e *model.AnimeOfficialAccount) bool { return workSourcedAccountServices[e.Service] },
		// worksはaccountのハンドルのみをsourceとするため、accountが異なる場合だけ
		// 変更扱いにする。label / label_en / sort_numberは触らず、編集者の編集を保全する。
		func(d repository.CreateAnimeOfficialAccountParams, e *model.AnimeOfficialAccount) bool {
			return e.Account != d.Account
		},
	)
}

// desiredAnimeOfficialAccountsはworksのバッチが持つべき公式アカウント行を導出する。
// twitter_username -> xのアカウント。ハンドルはそのまま使う (Railsは先頭の '@' を付けずに
// 保持しており、anime_official_accountsが持つ素の文字列の形と一致する)。twitter_usernameが
// NULLまたは空のときは行を作らないため、それを持たないworkは何も寄与せず、既存のx行が
// あれば後段で削除される。
func desiredAnimeOfficialAccounts(works []*model.Work) []repository.CreateAnimeOfficialAccountParams {
	desired := make([]repository.CreateAnimeOfficialAccountParams, 0, len(works))
	for _, w := range works {
		if w.AnimeID == nil {
			continue
		}
		if account := accountFromStringPtr(w.TwitterUsername); account != "" {
			desired = append(desired, repository.CreateAnimeOfficialAccountParams{
				AnimeID: *w.AnimeID,
				Service: model.AnimeAccountServiceX,
				Account: account,
			})
		}
	}
	return desired
}

// accountFromStringPtrはworksのNULL許容ハンドル列を読み、NULL (nil) と空文字列の
// 両方を "" (行なし) に写像する。Railsは欠損のハンドルをNULLまたは "" で持つため、ここでは
// どちらも欠損として扱う。
func accountFromStringPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// applyPlanはリコンサイル計画を1トランザクションで永続化し、テーブルごとの件数を
// 返す。書き込むものが無ければトランザクションを開かずに早期returnするため、既に同期済みの
// ページは書き込みコストがかからない。
func (uc *SyncAnimeOfficialAccountsUsecase) applyPlan(ctx context.Context, plan satelliteReconcilePlan[repository.CreateAnimeOfficialAccountParams, *model.AnimeOfficialAccount]) (satelliteReconcileCounts, error) {
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
			return satelliteReconcileCounts{}, fmt.Errorf("anime_official_accountの作成に失敗 (anime_id=%d, service=%s): %w", create.AnimeID, create.Service, err)
		}
		counts.Created++
	}

	for _, update := range plan.updates {
		if err := repo.Update(ctx, repository.UpdateAnimeOfficialAccountParams{
			ID:      update.existing.ID,
			Account: update.desired.Account,
		}); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_official_accountの更新に失敗 (id=%d): %w", update.existing.ID, err)
		}
		counts.Updated++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_official_accountの削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}
