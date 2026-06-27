package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedEventKinds is the set of anime_events kinds that works are the source of:
// started_on / ended_on map to the broadcast event. The reconcile limits deletes to rows
// in this key space, so an event an editor might add for some other kind (e.g. a
// revival_screening) is never clobbered (see reconcileSatellite).
//
// [Ja] workSourcedEventKinds は works が source とする anime_events の種別集合。
// started_on / ended_on が broadcast イベントに写像される。リコンサイルは削除をこのキー空間の
// 行に限定するため、編集者が他の種別 (例: revival_screening) に足しうるイベントが壊されない
// (reconcileSatellite を参照)。
var workSourcedEventKinds = map[model.AnimeEventKind]bool{
	model.AnimeEventKindBroadcast: true,
}

// animeEventKey is the natural key reconcileSatellite matches desired rows (derived from
// works) against existing rows on: (anime_id, kind). It includes the anime_id because a
// reconciler diffs a whole page of works at once, so a key without it would collide
// across works.
//
// [Ja] animeEventKey は reconcileSatellite が (works から導出した) あるべき行を既存行と
// 突合する自然キー (anime_id, kind)。リコンサイラは works のページ全体を一括で突合するため、
// anime_id を含めないと works をまたいでキーが衝突する。
type animeEventKey struct {
	animeID model.AnimeID
	kind    model.AnimeEventKind
}

// SyncAnimeEventsUsecase is the phase 2 satellite reconciler for the anime_events table.
// Registered into SyncWorkSatellitesUsecase, it maps each anime-resolved work's
// started_on / ended_on columns onto the broadcast event row of its anime, then creates
// / updates / deletes the rows so they match. It is a self-contained write usecase
// mirroring SyncWorksToAnimesUsecase: it reads the existing rows, plans the diff via
// reconcileSatellite, and applies the plan in its own transaction. The Reconcile method
// (not Execute) satisfies the satelliteReconciler interface.
//
// Unlike anime_hashtags / anime_seasons there is an update path: kind is the natural key
// while started_on / ended_on are mutable non-key columns works source, so a changed
// broadcast period is an in-place update (same key) rather than a delete plus a create
// (mirroring anime_official_accounts, whose account is its mutable non-key column).
//
// [Ja] SyncAnimeEventsUsecase は anime_events テーブルに対するフェーズ 2 の別表リコンサイラ。
// SyncWorkSatellitesUsecase に登録され、anime 解決済みの各 work の started_on / ended_on
// カラムをその anime の broadcast イベント行に写像し、一致するよう行を作成 / 更新 / 削除する。
// SyncWorksToAnimesUsecase を写した自己完結の書き込み UseCase で、既存行を読み、
// reconcileSatellite で差分を計画し、自前のトランザクションで適用する。Execute ではなく
// Reconcile メソッドが satelliteReconciler インターフェースを満たす。
//
// anime_hashtags / anime_seasons と違い更新パスを持つ。kind が自然キーで、started_on /
// ended_on は works が source する可変の非キー列のため、放送期間の変更は削除 + 作成ではなく
// その場の更新 (同じキー) になる (account が可変の非キー列である anime_official_accounts を
// 写したもの)。
type SyncAnimeEventsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeEventRepository
}

// NewSyncAnimeEventsUsecase constructs a SyncAnimeEventsUsecase.
//
// [Ja] NewSyncAnimeEventsUsecase は SyncAnimeEventsUsecase を生成する。
func NewSyncAnimeEventsUsecase(db *sql.DB, repo *repository.AnimeEventRepository) *SyncAnimeEventsUsecase {
	return &SyncAnimeEventsUsecase{db: db, repo: repo}
}

// Reconcile reconciles the anime_events rows for the given anime-resolved works.
// Following the write-usecase rule, the desired rows are derived and the existing rows
// are read before applyPlan opens a transaction; the transaction performs persistence
// only.
//
// [Ja] Reconcile は指定された anime 解決済み works について anime_events 行をリコンサイル
// する。書き込み UseCase のルールに従い、あるべき行の導出と既存行の取得は applyPlan が
// トランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeEventsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	desired := desiredAnimeEvents(works)

	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存 anime_events の取得に失敗: %w", err)
	}

	plan := reconcileSatellite(
		desired,
		existing,
		func(d repository.CreateAnimeEventParams) animeEventKey {
			return animeEventKey{animeID: d.AnimeID, kind: d.Kind}
		},
		func(e *model.AnimeEvent) animeEventKey {
			return animeEventKey{animeID: e.AnimeID, kind: e.Kind}
		},
		func(e *model.AnimeEvent) bool { return workSourcedEventKinds[e.Kind] },
		// Works source the broadcast period (started_on / ended_on), so a row is changed
		// iff either date differs. title / description / sort_number are left untouched,
		// preserving any editor edits to them.
		//
		// [Ja] works は放送期間 (started_on / ended_on) を source するため、どちらかの日付が
		// 異なる場合だけ変更扱いにする。title / description / sort_number は触らず、編集者の
		// 編集を保全する。
		func(d repository.CreateAnimeEventParams, e *model.AnimeEvent) bool {
			return !sameDate(e.StartedOn, d.StartedOn) || !sameNullableDate(e.EndedOn, d.EndedOn)
		},
	)

	return uc.applyPlan(ctx, plan)
}

// desiredAnimeEvents derives the event rows a batch of works should have:
// started_on / ended_on -> the broadcast event. A row is created only when started_on is
// present (anime_events.started_on is NOT NULL), so a work without a start date
// contributes nothing and any existing works-managed broadcast row for it is later
// deleted. ended_on is carried verbatim and may be nil (an open-ended broadcast).
//
// [Ja] desiredAnimeEvents は works のバッチが持つべきイベント行を導出する。
// started_on / ended_on -> broadcast イベント。行は started_on が存在するときだけ作る
// (anime_events.started_on は NOT NULL)。開始日を持たない work は何も寄与せず、既存の works
// 管理下の broadcast 行があれば後段で削除される。ended_on はそのまま使い、nil (終了未定の放送)
// もありうる。
func desiredAnimeEvents(works []*model.Work) []repository.CreateAnimeEventParams {
	desired := make([]repository.CreateAnimeEventParams, 0, len(works))
	for _, w := range works {
		if w.AnimeID == nil || w.StartedOn == nil {
			continue
		}
		desired = append(desired, repository.CreateAnimeEventParams{
			AnimeID:   *w.AnimeID,
			Kind:      model.AnimeEventKindBroadcast,
			StartedOn: *w.StartedOn,
			EndedOn:   w.EndedOn,
		})
	}
	return desired
}

// sameDate reports whether two times fall on the same calendar date. anime_events
// started_on / ended_on are date columns, so only the calendar date is meaningful: a
// works-sourced date and the round-tripped stored date can carry different clock times /
// zones from the driver, and comparing the instants would spuriously report a diff (and
// break the idempotency the cutover decision depends on).
//
// [Ja] sameDate は 2 つの時刻が同じ暦日かを返す。anime_events の started_on / ended_on は
// date 列なので意味があるのは暦日だけ。works が source した日付と、保存して読み戻した日付は
// ドライバ次第で異なる時刻 / ゾーンを持ちうるため、瞬間を比較すると誤って差分扱いになり
// (正本切り替え判定が依拠する冪等性を壊す)。
func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// sameNullableDate reports whether two nullable dates are equal: both nil counts as
// equal, one nil as different, and two present dates are compared by calendar date.
//
// [Ja] sameNullableDate は NULL 許容の 2 つの日付が等しいかを返す。両方 nil は等しい、
// 片方 nil は異なる、両方ある場合は暦日で比較する。
func sameNullableDate(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return sameDate(*a, *b)
}

// applyPlan persists the reconcile plan in a single transaction and returns the
// per-table counts. It returns early without opening a transaction when there is nothing
// to write, so an already-synced page costs no write.
//
// [Ja] applyPlan はリコンサイル計画を 1 トランザクションで永続化し、テーブルごとの件数を返す。
// 書き込むものが無ければトランザクションを開かずに早期 return するため、既に同期済みのページは
// 書き込みコストがかからない。
func (uc *SyncAnimeEventsUsecase) applyPlan(ctx context.Context, plan satelliteReconcilePlan[repository.CreateAnimeEventParams, *model.AnimeEvent]) (satelliteReconcileCounts, error) {
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
			return satelliteReconcileCounts{}, fmt.Errorf("anime_event の作成に失敗 (anime_id=%d, kind=%s): %w", create.AnimeID, create.Kind, err)
		}
		counts.Created++
	}

	for _, update := range plan.updates {
		if err := repo.Update(ctx, repository.UpdateAnimeEventParams{
			ID:        update.existing.ID,
			StartedOn: update.desired.StartedOn,
			EndedOn:   update.desired.EndedOn,
		}); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_event の更新に失敗 (id=%d): %w", update.existing.ID, err)
		}
		counts.Updated++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_event の削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}
