package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedEventKindsはworksがsourceとするanime_eventsの種別集合。
// started_on / ended_onがbroadcastイベントに写像される。リコンサイルは削除をこのキー空間の
// 行に限定するため、編集者が他の種別 (例: revival_screening) に足しうるイベントが壊されない
// (reconcileSatelliteを参照)。
var workSourcedEventKinds = map[model.AnimeEventKind]bool{
	model.AnimeEventKindBroadcast: true,
}

// animeEventKeyはreconcileSatelliteが (worksから導出した) あるべき行を既存行と
// 突合する自然キー (anime_id, kind)。リコンサイラはworksのページ全体を一括で突合するため、
// anime_idを含めないとworksをまたいでキーが衝突する。
type animeEventKey struct {
	animeID model.AnimeID
	kind    model.AnimeEventKind
}

// SyncAnimeEventsUsecaseはanime_eventsテーブルに対するフェーズ2の別表リコンサイラ。
// SyncWorkSatellitesUsecaseに登録され、anime解決済みの各workのstarted_on / ended_on
// カラムをそのanimeのbroadcastイベント行に写像し、一致するよう行を作成 / 更新 / 削除する。
// SyncWorksToAnimesUsecaseを写した自己完結の書き込みUseCaseで、既存行を読み、
// reconcileSatelliteで差分を計画し、自前のトランザクションで適用する。Executeではなく
// ReconcileメソッドがsatelliteReconcilerインターフェースを満たす。
//
// anime_hashtags / anime_seasonsと違い更新パスを持つ。kindが自然キーで、started_on /
// ended_onはworksがsourceする可変の非キー列のため、放送期間の変更は削除 + 作成ではなく
// その場の更新 (同じキー) になる (accountが可変の非キー列であるanime_official_accountsを
// 写したもの)。
type SyncAnimeEventsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeEventRepository
}

// NewSyncAnimeEventsUsecaseはSyncAnimeEventsUsecaseを生成する。
func NewSyncAnimeEventsUsecase(db *sql.DB, repo *repository.AnimeEventRepository) *SyncAnimeEventsUsecase {
	return &SyncAnimeEventsUsecase{db: db, repo: repo}
}

// Reconcileは指定されたanime解決済みworksについてanime_events行をリコンサイル
// する。書き込みUseCaseのルールに従い、あるべき行の導出と既存行の取得はapplyPlanが
// トランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeEventsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存anime_eventsの取得に失敗: %w", err)
	}

	return uc.applyPlan(ctx, planAnimeEvents(works, existing))
}

// planAnimeEventsは指定されたanime解決済みworksのanime_events行について、
// 既存行に対するリコンサイル計画を組み立てる。フェーズ2のバッチリコンサイラ (Reconcile) と
// フェーズ3の作品 作成 / 更新 の両書き (planWorkSatellites) で共有し、あるべき行の導出・
// 自然キー・削除限定・変更検出を両経路で単一の正本に保つ。
func planAnimeEvents(works []*model.Work, existing []*model.AnimeEvent) satelliteReconcilePlan[repository.CreateAnimeEventParams, *model.AnimeEvent] {
	return reconcileSatellite(
		desiredAnimeEvents(works),
		existing,
		func(d repository.CreateAnimeEventParams) animeEventKey {
			return animeEventKey{animeID: d.AnimeID, kind: d.Kind}
		},
		func(e *model.AnimeEvent) animeEventKey {
			return animeEventKey{animeID: e.AnimeID, kind: e.Kind}
		},
		func(e *model.AnimeEvent) bool { return workSourcedEventKinds[e.Kind] },
		// worksは放送期間 (started_on / ended_on) をsourceするため、どちらかの日付が
		// 異なる場合だけ変更扱いにする。title / description / sort_numberは触らず、編集者の
		// 編集を保全する。
		func(d repository.CreateAnimeEventParams, e *model.AnimeEvent) bool {
			return !sameDate(e.StartedOn, d.StartedOn) || !sameNullableDate(e.EndedOn, d.EndedOn)
		},
	)
}

// desiredAnimeEventsはworksのバッチが持つべきイベント行を導出する。
// started_on / ended_on -> broadcastイベント。行はstarted_onが存在するときだけ作る
// (anime_events.started_onはNOT NULL)。開始日を持たないworkは何も寄与せず、既存のworks
// 管理下のbroadcast行があれば後段で削除される。ended_onはそのまま使い、nil (終了未定の放送)
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

// sameDateは2つの時刻が同じ暦日かを返す。anime_eventsのstarted_on / ended_onは
// date列なので意味があるのは暦日だけ。worksがsourceした日付と、保存して読み戻した日付は
// ドライバ次第で異なる時刻 / ゾーンを持ちうるため、瞬間を比較すると誤って差分扱いになり
// (正本切り替え判定が依拠する冪等性を壊す)。
func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// sameNullableDateはNULL許容の2つの日付が等しいかを返す。両方nilは等しい、
// 片方nilは異なる、両方ある場合は暦日で比較する。
func sameNullableDate(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return sameDate(*a, *b)
}

// applyPlanはリコンサイル計画を1トランザクションで永続化し、テーブルごとの件数を返す。
// 書き込むものが無ければトランザクションを開かずに早期returnするため、既に同期済みのページは
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
			return satelliteReconcileCounts{}, fmt.Errorf("anime_eventの作成に失敗 (anime_id=%d, kind=%s): %w", create.AnimeID, create.Kind, err)
		}
		counts.Created++
	}

	for _, update := range plan.updates {
		if err := repo.Update(ctx, repository.UpdateAnimeEventParams{
			ID:        update.existing.ID,
			StartedOn: update.desired.StartedOn,
			EndedOn:   update.desired.EndedOn,
		}); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_eventの更新に失敗 (id=%d): %w", update.existing.ID, err)
		}
		counts.Updated++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_eventの削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}
