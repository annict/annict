package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/annict/annict/go/internal/model"
)

// DefaultSyncAnimesBatchSize is the page size SyncAnimesUsecase walks the works /
// episodes tables with when the caller does not specify one. It is a starting
// point; the plan calls for measuring the production data size before settling on
// a final value, so it is kept tunable via the constructor.
//
// [Ja] DefaultSyncAnimesBatchSize は呼び出し側が指定しない場合に SyncAnimesUsecase が
// works / episodes テーブルを走査するページサイズ。あくまで初期値で、計画では本番データ
// 規模を実測してから最終値を決めるため、コンストラクタで調整可能にしている。
const DefaultSyncAnimesBatchSize = 1000

// workIDPager yields work IDs page by page (keyset pagination by id). Implemented
// by *repository.WorkRepository; kept as a narrow interface so the batch
// orchestration can be tested with a fake that yields only the test's own IDs,
// instead of scanning the shared test database.
//
// [Ja] workIDPager は work ID をページ単位で返す (id による keyset ページネーション)。
// 実体は *repository.WorkRepository。共有テスト DB を走査せずに、テスト自身の ID だけを
// 返すフェイクでバッチのオーケストレーションをテストできるよう、狭いインターフェースに
// している。
type workIDPager interface {
	ListIDsAfter(ctx context.Context, afterID model.WorkID, batchSize int) ([]model.WorkID, error)
}

// episodeIDPager yields episode IDs page by page (keyset pagination by id).
// Implemented by *repository.EpisodeRepository.
//
// [Ja] episodeIDPager は episode ID をページ単位で返す (id による keyset ページネーション)。
// 実体は *repository.EpisodeRepository。
type episodeIDPager interface {
	ListIDsAfter(ctx context.Context, afterID model.EpisodeID, batchSize int) ([]model.EpisodeID, error)
}

// worksToAnimesSyncer reconciles one page of works into animes. Implemented by
// *SyncWorksToAnimesUsecase.
//
// [Ja] worksToAnimesSyncer は works の 1 ページを animes へリコンサイルする。
// 実体は *SyncWorksToAnimesUsecase。
type worksToAnimesSyncer interface {
	Execute(ctx context.Context, input SyncWorksToAnimesInput) (*SyncWorksToAnimesResult, error)
}

// episodesToAnimesSyncer reconciles one page of episodes into animes. Implemented
// by *SyncEpisodesToAnimesUsecase.
//
// [Ja] episodesToAnimesSyncer は episodes の 1 ページを animes へリコンサイルする。
// 実体は *SyncEpisodesToAnimesUsecase。
type episodesToAnimesSyncer interface {
	Execute(ctx context.Context, input SyncEpisodesToAnimesInput) (*SyncEpisodesToAnimesResult, error)
}

// SyncAnimesUsecase is the phase 2 full-reconciliation batch. It walks the whole
// works table and then the whole episodes table page by page, handing each page to
// the per-page sync usecases, and aggregates the diff-detection counts for logging.
//
// Works are synced before episodes on purpose: an episode's classification needs
// its parent work's anime_id (parent_anime_id is NOT NULL), so finishing the works
// pass first lets the episodes pass resolve nearly every parent. Episodes whose
// parent is still unmapped are deferred by the episode sync (SkippedNoParent) and
// picked up on a later run; the reconciliation is idempotent.
//
// [Ja] SyncAnimesUsecase はフェーズ 2 のフル・リコンシリエーションバッチ。works テーブル
// 全体、続いて episodes テーブル全体をページ単位で走査し、各ページをページ単位の同期
// UseCase に渡して、差分検出件数を集計してログに出す。
//
// works を episodes より先に同期するのは意図的。episode の分類は親作品の anime_id を
// 必要とする (parent_anime_id が NOT NULL) ため、works の走査を先に終えると episodes の
// 走査でほぼすべての親を解決できる。親が未マッピングのままの episode は episode 同期が
// 繰り延べ (SkippedNoParent)、後続の実行で取り込む。リコンサイルは冪等。
type SyncAnimesUsecase struct {
	workIDPager    workIDPager
	episodeIDPager episodeIDPager
	worksSyncer    worksToAnimesSyncer
	episodesSyncer episodesToAnimesSyncer
	batchSize      int
}

// NewSyncAnimesUsecase constructs a SyncAnimesUsecase. A batchSize of zero or less
// falls back to DefaultSyncAnimesBatchSize so the keyset loop always makes progress.
//
// [Ja] NewSyncAnimesUsecase は SyncAnimesUsecase を生成する。batchSize が 0 以下の場合は
// DefaultSyncAnimesBatchSize にフォールバックし、keyset ループが必ず前進するようにする。
func NewSyncAnimesUsecase(
	workIDPager workIDPager,
	episodeIDPager episodeIDPager,
	worksSyncer worksToAnimesSyncer,
	episodesSyncer episodesToAnimesSyncer,
	batchSize int,
) *SyncAnimesUsecase {
	if batchSize <= 0 {
		batchSize = DefaultSyncAnimesBatchSize
	}
	return &SyncAnimesUsecase{
		workIDPager:    workIDPager,
		episodeIDPager: episodeIDPager,
		worksSyncer:    worksSyncer,
		episodesSyncer: episodesSyncer,
		batchSize:      batchSize,
	}
}

// SyncAnimesResult reports the aggregated reconciliation counts of one batch run,
// split by source table. Created / Updated together are the diff-detection metric
// the cutover decision depends on; a run reporting both as zero means the new
// schema already matches works / episodes.
//
// [Ja] SyncAnimesResult は 1 回のバッチ実行の集計リコンサイル件数を、源泉テーブルごとに
// 報告する。Created / Updated の合計が、正本切り替え判定が依拠する差分検出メトリクス。
// 両方が 0 の実行は、新スキーマが既に works / episodes と一致していることを意味する。
type SyncAnimesResult struct {
	Works    SyncWorksToAnimesResult
	Episodes SyncEpisodesToAnimesResult
}

// Execute runs the full reconciliation: all works first, then all episodes.
//
// [Ja] Execute はフル・リコンシリエーションを実行する。まず全 works、続いて全 episodes。
func (uc *SyncAnimesUsecase) Execute(ctx context.Context) (*SyncAnimesResult, error) {
	slog.InfoContext(ctx, "animes 同期バッチを開始します", "batch_size", uc.batchSize)

	works, err := uc.syncAllWorks(ctx)
	if err != nil {
		return nil, err
	}

	episodes, err := uc.syncAllEpisodes(ctx)
	if err != nil {
		return nil, err
	}

	result := &SyncAnimesResult{Works: works, Episodes: episodes}

	slog.InfoContext(ctx, "animes 同期バッチが完了しました",
		"works_processed", result.Works.Processed,
		"works_created", result.Works.Created,
		"works_updated", result.Works.Updated,
		"works_unchanged", result.Works.Unchanged,
		"episodes_processed", result.Episodes.Processed,
		"episodes_created", result.Episodes.Created,
		"episodes_updated", result.Episodes.Updated,
		"episodes_unchanged", result.Episodes.Unchanged,
		"episodes_skipped_no_parent", result.Episodes.SkippedNoParent,
	)

	return result, nil
}

// syncAllWorks walks the works table by keyset pagination, reconciling one page at
// a time and accumulating the counts. The cursor (afterID) strictly increases each
// iteration, so the loop terminates; an empty page marks the end.
//
// [Ja] syncAllWorks は works テーブルを keyset ページネーションで走査し、1 ページずつ
// リコンサイルして件数を積算する。カーソル (afterID) は反復ごとに厳密に増加するため
// ループは必ず終了し、空ページが終端を表す。
func (uc *SyncAnimesUsecase) syncAllWorks(ctx context.Context) (SyncWorksToAnimesResult, error) {
	var total SyncWorksToAnimesResult
	var afterID model.WorkID

	for {
		ids, err := uc.workIDPager.ListIDsAfter(ctx, afterID, uc.batchSize)
		if err != nil {
			return total, fmt.Errorf("works ID ページの取得に失敗: %w", err)
		}
		if len(ids) == 0 {
			break
		}

		res, err := uc.worksSyncer.Execute(ctx, SyncWorksToAnimesInput{WorkIDs: ids})
		if err != nil {
			return total, fmt.Errorf("works 同期ページの処理に失敗: %w", err)
		}

		total.Processed += res.Processed
		total.Created += res.Created
		total.Updated += res.Updated
		total.Unchanged += res.Unchanged

		afterID = ids[len(ids)-1]
	}

	return total, nil
}

// syncAllEpisodes walks the episodes table by keyset pagination, reconciling one
// page at a time and accumulating the counts (including SkippedNoParent).
//
// [Ja] syncAllEpisodes は episodes テーブルを keyset ページネーションで走査し、1 ページ
// ずつリコンサイルして件数 (SkippedNoParent を含む) を積算する。
func (uc *SyncAnimesUsecase) syncAllEpisodes(ctx context.Context) (SyncEpisodesToAnimesResult, error) {
	var total SyncEpisodesToAnimesResult
	var afterID model.EpisodeID

	for {
		ids, err := uc.episodeIDPager.ListIDsAfter(ctx, afterID, uc.batchSize)
		if err != nil {
			return total, fmt.Errorf("episodes ID ページの取得に失敗: %w", err)
		}
		if len(ids) == 0 {
			break
		}

		res, err := uc.episodesSyncer.Execute(ctx, SyncEpisodesToAnimesInput{EpisodeIDs: ids})
		if err != nil {
			return total, fmt.Errorf("episodes 同期ページの処理に失敗: %w", err)
		}

		total.Processed += res.Processed
		total.Created += res.Created
		total.Updated += res.Updated
		total.Unchanged += res.Unchanged
		total.SkippedNoParent += res.SkippedNoParent

		afterID = ids[len(ids)-1]
	}

	return total, nil
}
