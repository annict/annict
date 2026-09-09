package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
)

// satelliteWorkLoader loads works with their satellite source columns projected.
// Implemented by *repository.WorkRepository; kept as a narrow interface so the
// orchestration can be tested with a fake loader that returns the test's own works,
// instead of touching the database (the loader SQL is covered by the repository test).
//
// [Ja] satelliteWorkLoader は別表ソース列を射影した works をロードする。実体は
// *repository.WorkRepository。DB に触れず、テスト自身の works を返すフェイクローダーで
// オーケストレーションをテストできるよう狭いインターフェースにしている (ローダー SQL 自体は
// リポジトリテストでカバーする)。
type satelliteWorkLoader interface {
	ListForSatelliteSyncByIDs(ctx context.Context, workIDs []model.WorkID) ([]*model.Work, error)
}

// satelliteReconciler reconciles one satellite table (e.g. anime_external_ids) for a
// batch of anime-resolved works, returning the per-table reconcile counts. Tasks 2-8
// onward register one reconciler per table; each is a self-contained write usecase that
// reads the existing rows, plans the diff via reconcileSatellite, and applies the
// creates / updates / deletes in its own transaction (mirroring SyncWorksToAnimesUsecase).
// Each registered reconciler reads its existing rows and persists its diff; with none
// registered this pass only loads the works and counts how many are deferred for a
// missing anime_id.
//
// The works slice handed to Reconcile is the same one shared across every registered
// reconciler, so a reconciler must treat it as read-only: mutating the slice or the
// fields reachable through its *model.Work pointers would leak into the reconcilers
// that run later in the same page.
//
// [Ja] satelliteReconciler は anime 解決済みの works のバッチに対して 1 つの別表
// (例: anime_external_ids) をリコンサイルし、テーブルごとの件数を返す。タスク 2-8 以降が
// テーブルごとに 1 つずつ登録する。各リコンサイラは既存行を読み、reconcileSatellite で差分を
// 計画し、自前のトランザクションで作成 / 更新 / 削除を適用する自己完結した書き込み UseCase
// (SyncWorksToAnimesUsecase を写したもの)。登録された各リコンサイラは既存行を読み差分を
// 永続化する。1 つも登録されていなければ、本パスは works をロードし anime_id 未解決で
// 繰り延べた件数を数えるだけになる。
//
// Reconcile に渡される works スライスは登録済みの全リコンサイラで共有される同一の
// ものなので、リコンサイラはこれを read-only として扱う。スライスや *model.Work
// ポインタ経由で到達できるフィールドを変更すると、同じページで後に走るリコンサイラに
// その変更が漏れる。
type satelliteReconciler interface {
	Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error)
}

// satelliteReconcileCounts is the reconcile outcome of one satellite table, used both
// as a reconciler's return value and as the aggregate accumulated across reconcilers.
//
// [Ja] satelliteReconcileCounts は 1 つの別表のリコンサイル結果。リコンサイラの戻り値と、
// リコンサイラ横断で積算する集計の双方に使う。
type satelliteReconcileCounts struct {
	Created   int
	Updated   int
	Deleted   int
	Unchanged int
}

// SyncWorkSatellitesUsecase is the phase 2 third pass: after works and episodes are
// reconciled into animes / anime_classifications, it reconciles the six satellite
// tables that works (but not episodes) are the source of. It runs as a separate pass,
// not folded into the works sync (task 2-2), so the works-sync core (mapping, diff,
// anime_id write-back) stays small and per-table reconcilers can be added one at a time.
//
// Only works whose anime_id is already resolved are reconciled; an unmapped work is
// deferred (counted in SkippedNoAnime) and picked up on a later run once the works pass
// has written its anime_id back. The reconciliation is idempotent, so deferring is safe.
//
// [Ja] SyncWorkSatellitesUsecase はフェーズ 2 の第 3 パス。works と episodes を animes /
// anime_classifications にリコンサイルした後、works (episodes ではなく) が source とする
// 6 つの別表をリコンサイルする。works 同期 (タスク 2-2) に織り込まず独立したパスとして走らせ、
// works 同期の中核 (写像・差分・anime_id 書き戻し) を小さく保ち、テーブルごとのリコンサイラを
// 1 つずつ足せるようにする。
//
// anime_id が解決済みの work だけをリコンサイルし、未マッピングの work は繰り延べる
// (SkippedNoAnime に数える)。works パスが anime_id を書き戻した後の実行で取り込まれる。
// リコンサイルは冪等なので繰り延べは安全。
type SyncWorkSatellitesUsecase struct {
	workLoader  satelliteWorkLoader
	reconcilers []satelliteReconciler
}

// NewSyncWorkSatellitesUsecase constructs a SyncWorkSatellitesUsecase. Reconcilers are
// passed variadically so tasks 2-8 onward can register one per satellite table; zero
// reconcilers is also valid (the pass then only loads works and defers unmapped ones).
//
// [Ja] NewSyncWorkSatellitesUsecase は SyncWorkSatellitesUsecase を生成する。リコンサイラは
// 可変長で渡し、タスク 2-8 以降が別表ごとに 1 つずつ登録できるようにする。0 個でも有効で、
// その場合は本パスが works のロードと未マッピング work の繰り延べだけを行う。
func NewSyncWorkSatellitesUsecase(
	workLoader satelliteWorkLoader,
	reconcilers ...satelliteReconciler,
) *SyncWorkSatellitesUsecase {
	return &SyncWorkSatellitesUsecase{
		workLoader:  workLoader,
		reconcilers: reconcilers,
	}
}

// SyncWorkSatellitesInput carries the work IDs to reconcile in this run (typically one
// page of the full-table scan driven by the batch job).
//
// [Ja] SyncWorkSatellitesInput は今回のリコンサイル対象の work ID を保持する
// (通常はバッチジョブが駆動する全件スキャンの 1 ページ分)。
type SyncWorkSatellitesInput struct {
	WorkIDs []model.WorkID
}

// SyncWorkSatellitesResult reports the reconciliation outcome counts aggregated across
// all satellite tables. Created / Updated / Deleted together form the diff-detection
// metric the cutover decision depends on; SkippedNoAnime counts works deferred because
// their anime_id is not resolved yet.
//
// [Ja] SyncWorkSatellitesResult は全別表で集計したリコンサイル結果の件数を報告する。
// Created / Updated / Deleted の合計が、正本切り替え判定が依拠する差分検出メトリクス。
// SkippedNoAnime は anime_id 未解決で繰り延べた work の件数。
type SyncWorkSatellitesResult struct {
	Processed      int
	Created        int
	Updated        int
	Deleted        int
	Unchanged      int
	SkippedNoAnime int
}

// Execute reconciles the satellite tables for the input works. Following the
// write-usecase rule, the works are loaded before any reconciler opens a transaction;
// each registered reconciler then reads its existing rows and persists its diff.
//
// [Ja] Execute は入力 works について別表をリコンサイルする。書き込み UseCase のルールに従い、
// works はどのリコンサイラがトランザクションを開くより前にロードする。登録された各リコンサイラ
// が既存行を読んで差分を永続化する。
func (uc *SyncWorkSatellitesUsecase) Execute(ctx context.Context, input SyncWorkSatellitesInput) (*SyncWorkSatellitesResult, error) {
	works, err := uc.workLoader.ListForSatelliteSyncByIDs(ctx, input.WorkIDs)
	if err != nil {
		return nil, fmt.Errorf("別表同期対象 works の取得に失敗: %w", err)
	}

	resolved := worksWithAnime(works)
	result := &SyncWorkSatellitesResult{
		Processed:      len(works),
		SkippedNoAnime: len(works) - len(resolved),
	}
	if len(resolved) == 0 {
		return result, nil
	}

	for _, reconciler := range uc.reconcilers {
		counts, err := reconciler.Reconcile(ctx, resolved)
		if err != nil {
			return nil, fmt.Errorf("別表のリコンサイルに失敗: %w", err)
		}
		result.Created += counts.Created
		result.Updated += counts.Updated
		result.Deleted += counts.Deleted
		result.Unchanged += counts.Unchanged
	}

	return result, nil
}

// worksWithAnime returns the works whose anime_id is already resolved. Works still
// pending an anime (anime_id is nil) cannot have their satellite rows anchored yet and
// are deferred to a later run.
//
// [Ja] worksWithAnime は anime_id が解決済みの works を返す。anime 未解決 (anime_id が nil)
// の work は別表行をまだ紐付けられないため、後続の実行へ繰り延べる。
func worksWithAnime(works []*model.Work) []*model.Work {
	resolved := make([]*model.Work, 0, len(works))
	for _, w := range works {
		if w.AnimeID != nil {
			resolved = append(resolved, w)
		}
	}
	return resolved
}
