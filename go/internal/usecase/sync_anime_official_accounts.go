package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedAccountServices is the set of anime_official_accounts services that works
// are the source of: twitter_username maps to the x account. The reconcile limits
// deletes to rows in this key space, so an account an editor might add for some other
// service (e.g. youtube, instagram) is never clobbered (see reconcileSatellite).
//
// [Ja] workSourcedAccountServices は works が source とする anime_official_accounts の
// サービス集合。twitter_username が x のアカウントに写像される。リコンサイルは削除をこの
// キー空間の行に限定するため、編集者が他のサービス (例: youtube / instagram) に足しうる
// アカウントが壊されない (reconcileSatellite を参照)。
var workSourcedAccountServices = map[model.AnimeAccountService]bool{
	model.AnimeAccountServiceX: true,
}

// animeOfficialAccountKey is the natural key reconcileSatellite matches desired rows
// (derived from works) against existing rows on: (anime_id, service). It includes the
// anime_id because a reconciler diffs a whole page of works at once, so a key without
// it would collide across works.
//
// [Ja] animeOfficialAccountKey は reconcileSatellite が (works から導出した) あるべき行を
// 既存行と突合する自然キー (anime_id, service)。リコンサイラは works のページ全体を一括で
// 突合するため、anime_id を含めないと works をまたいでキーが衝突する。
type animeOfficialAccountKey struct {
	animeID model.AnimeID
	service model.AnimeAccountService
}

// SyncAnimeOfficialAccountsUsecase is the phase 2 satellite reconciler for the
// anime_official_accounts table. Registered into SyncWorkSatellitesUsecase, it maps
// each anime-resolved work's twitter_username column onto the x official-account row of
// its anime, then creates / updates / deletes the rows so they match. It is a
// self-contained write usecase mirroring SyncWorksToAnimesUsecase: it reads the
// existing rows, plans the diff via reconcileSatellite, and applies the plan in its own
// transaction. The Reconcile method (not Execute) satisfies the satelliteReconciler
// interface.
//
// [Ja] SyncAnimeOfficialAccountsUsecase は anime_official_accounts テーブルに対する
// フェーズ 2 の別表リコンサイラ。SyncWorkSatellitesUsecase に登録され、anime 解決済みの
// 各 work の twitter_username カラムをその anime の x 公式アカウント行に写像し、一致する
// よう行を作成 / 更新 / 削除する。SyncWorksToAnimesUsecase を写した自己完結の書き込み
// UseCase で、既存行を読み、reconcileSatellite で差分を計画し、自前のトランザクションで
// 適用する。Execute ではなく Reconcile メソッドが satelliteReconciler インターフェースを
// 満たす。
type SyncAnimeOfficialAccountsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeOfficialAccountRepository
}

// NewSyncAnimeOfficialAccountsUsecase constructs a SyncAnimeOfficialAccountsUsecase.
//
// [Ja] NewSyncAnimeOfficialAccountsUsecase は SyncAnimeOfficialAccountsUsecase を生成する。
func NewSyncAnimeOfficialAccountsUsecase(db *sql.DB, repo *repository.AnimeOfficialAccountRepository) *SyncAnimeOfficialAccountsUsecase {
	return &SyncAnimeOfficialAccountsUsecase{db: db, repo: repo}
}

// Reconcile reconciles the anime_official_accounts rows for the given anime-resolved
// works. Following the write-usecase rule, the desired rows are derived and the existing
// rows are read before applyPlan opens a transaction; the transaction performs
// persistence only.
//
// [Ja] Reconcile は指定された anime 解決済み works について anime_official_accounts 行を
// リコンサイルする。書き込み UseCase のルールに従い、あるべき行の導出と既存行の取得は
// applyPlan がトランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeOfficialAccountsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	desired := desiredAnimeOfficialAccounts(works)

	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存 anime_official_accounts の取得に失敗: %w", err)
	}

	plan := reconcileSatellite(
		desired,
		existing,
		func(d repository.CreateAnimeOfficialAccountParams) animeOfficialAccountKey {
			return animeOfficialAccountKey{animeID: d.AnimeID, service: d.Service}
		},
		func(e *model.AnimeOfficialAccount) animeOfficialAccountKey {
			return animeOfficialAccountKey{animeID: e.AnimeID, service: e.Service}
		},
		func(e *model.AnimeOfficialAccount) bool { return workSourcedAccountServices[e.Service] },
		// Works only source the account handle, so a row is changed iff its account
		// differs. label / label_en / sort_number are left untouched, preserving any
		// editor edits to them.
		//
		// [Ja] works は account のハンドルのみを source とするため、account が異なる場合だけ
		// 変更扱いにする。label / label_en / sort_number は触らず、編集者の編集を保全する。
		func(d repository.CreateAnimeOfficialAccountParams, e *model.AnimeOfficialAccount) bool {
			return e.Account != d.Account
		},
	)

	return uc.applyPlan(ctx, plan)
}

// desiredAnimeOfficialAccounts derives the official-account rows a batch of works
// should have: twitter_username -> the x account. The handle is taken verbatim (Rails
// already stores it without a leading '@', matching the bare-string form
// anime_official_accounts holds). A NULL or empty twitter_username yields no row, so a
// work without one contributes nothing and any existing x row for it is later deleted.
//
// [Ja] desiredAnimeOfficialAccounts は works のバッチが持つべき公式アカウント行を導出する。
// twitter_username -> x のアカウント。ハンドルはそのまま使う (Rails は先頭の '@' を付けずに
// 保持しており、anime_official_accounts が持つ素の文字列の形と一致する)。twitter_username が
// NULL または空のときは行を作らないため、それを持たない work は何も寄与せず、既存の x 行が
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

// accountFromStringPtr reads a works nullable handle column, mapping both NULL (nil)
// and the empty string to "" (no row). Rails stores an absent handle as NULL or "", so
// both are treated as missing here.
//
// [Ja] accountFromStringPtr は works の NULL 許容ハンドル列を読み、NULL (nil) と空文字列の
// 両方を "" (行なし) に写像する。Rails は欠損のハンドルを NULL または "" で持つため、ここでは
// どちらも欠損として扱う。
func accountFromStringPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// applyPlan persists the reconcile plan in a single transaction and returns the
// per-table counts. It returns early without opening a transaction when there is
// nothing to write, so an already-synced page costs no write.
//
// [Ja] applyPlan はリコンサイル計画を 1 トランザクションで永続化し、テーブルごとの件数を
// 返す。書き込むものが無ければトランザクションを開かずに早期 return するため、既に同期済みの
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
			return satelliteReconcileCounts{}, fmt.Errorf("anime_official_account の作成に失敗 (anime_id=%d, service=%s): %w", create.AnimeID, create.Service, err)
		}
		counts.Created++
	}

	for _, update := range plan.updates {
		if err := repo.Update(ctx, repository.UpdateAnimeOfficialAccountParams{
			ID:      update.existing.ID,
			Account: update.desired.Account,
		}); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_official_account の更新に失敗 (id=%d): %w", update.existing.ID, err)
		}
		counts.Updated++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_official_account の削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}
