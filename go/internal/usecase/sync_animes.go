package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/annict/annict/go/internal/model"
)

// DefaultSyncAnimesBatchSizeは呼び出し側が指定しない場合にSyncAnimesUsecaseが
// works / episodesテーブルを走査するページサイズ。あくまで初期値で、計画では本番データ
// 規模を実測してから最終値を決めるため、コンストラクタで調整可能にしている。
const DefaultSyncAnimesBatchSize = 1000

// workIDPagerはwork IDをページ単位で返す (idによるkeysetページネーション)。
// 実体は *repository.WorkRepository。共有テストDBを走査せずに、テスト自身のIDだけを
// 返すフェイクでバッチのオーケストレーションをテストできるよう、狭いインターフェースに
// している。
type workIDPager interface {
	ListIDsAfter(ctx context.Context, afterID model.WorkID, batchSize int) ([]model.WorkID, error)
}

// episodeIDPagerはepisode IDをページ単位で返す (idによるkeysetページネーション)。
// 実体は *repository.EpisodeRepository。
type episodeIDPager interface {
	ListIDsAfter(ctx context.Context, afterID model.EpisodeID, batchSize int) ([]model.EpisodeID, error)
}

// worksToAnimesSyncerはworksの1ページをanimesへリコンサイルする。
// 実体は *SyncWorksToAnimesUsecase。
type worksToAnimesSyncer interface {
	Execute(ctx context.Context, input SyncWorksToAnimesInput) (*SyncWorksToAnimesResult, error)
}

// episodesToAnimesSyncerはepisodesの1ページをanimesへリコンサイルする。
// 実体は *SyncEpisodesToAnimesUsecase。
type episodesToAnimesSyncer interface {
	Execute(ctx context.Context, input SyncEpisodesToAnimesInput) (*SyncEpisodesToAnimesResult, error)
}

// workSatellitesSyncerはworksの1ページを別表へリコンサイルする。
// 実体は *SyncWorkSatellitesUsecase。
type workSatellitesSyncer interface {
	Execute(ctx context.Context, input SyncWorkSatellitesInput) (*SyncWorkSatellitesResult, error)
}

// SyncAnimesUsecaseはフェーズ2のフル・リコンシリエーションバッチ。worksテーブル
// 全体、続いてepisodesテーブル全体をページ単位で走査し、各ページをページ単位の同期
// UseCaseに渡して、差分検出件数を集計してログに出す。
//
// worksをepisodesより先に同期するのは意図的。episodeの分類は親作品のanime_idを
// 必要とする (parent_anime_idがNOT NULL) ため、worksの走査を先に終えるとepisodesの
// 走査でほぼすべての親を解決できる。親が未マッピングのままのepisodeはepisode同期が
// 繰り延べ (SkippedNoParent)、後続の実行で取り込む。リコンサイルは冪等。
//
// 第3のパスはworksテーブルを再度走査し、別表 (外部ID / リンク / 公式アカウント /
// ハッシュタグ / 季節 / イベント) をリコンサイルする。works同期パスがworks.anime_idを
// 書き戻すことに依存するため最後に走らせる。anime_idが未解決のままのworkは繰り延べる
// (SkippedNoAnime)。
type SyncAnimesUsecase struct {
	workIDPager      workIDPager
	episodeIDPager   episodeIDPager
	worksSyncer      worksToAnimesSyncer
	episodesSyncer   episodesToAnimesSyncer
	satellitesSyncer workSatellitesSyncer
	batchSize        int
}

// NewSyncAnimesUsecaseはSyncAnimesUsecaseを生成する。batchSizeが0以下の場合は
// DefaultSyncAnimesBatchSizeにフォールバックし、keysetループが必ず前進するようにする。
func NewSyncAnimesUsecase(
	workIDPager workIDPager,
	episodeIDPager episodeIDPager,
	worksSyncer worksToAnimesSyncer,
	episodesSyncer episodesToAnimesSyncer,
	satellitesSyncer workSatellitesSyncer,
	batchSize int,
) *SyncAnimesUsecase {
	if batchSize <= 0 {
		batchSize = DefaultSyncAnimesBatchSize
	}
	return &SyncAnimesUsecase{
		workIDPager:      workIDPager,
		episodeIDPager:   episodeIDPager,
		worksSyncer:      worksSyncer,
		episodesSyncer:   episodesSyncer,
		satellitesSyncer: satellitesSyncer,
		batchSize:        batchSize,
	}
}

// SyncAnimesResultは1回のバッチ実行の集計リコンサイル件数を、源泉テーブルごとに
// 報告する。Created / Updatedの合計が、正本切り替え判定が依拠する差分検出メトリクス。
// 両方が0の実行は、新スキーマが既にworks / episodesと一致していることを意味する。
type SyncAnimesResult struct {
	Works      SyncWorksToAnimesResult
	Episodes   SyncEpisodesToAnimesResult
	Satellites SyncWorkSatellitesResult
}

// Executeはフル・リコンシリエーションを実行する。まず全works、続いて全episodes、
// 最後にworksの別表。別表はworksパスがworks.anime_idを書き戻すことに依存するため
// 最後に走らせる。
func (uc *SyncAnimesUsecase) Execute(ctx context.Context) (*SyncAnimesResult, error) {
	slog.InfoContext(ctx, "animes同期バッチを開始します", "batch_size", uc.batchSize)

	works, err := uc.syncAllWorks(ctx)
	if err != nil {
		return nil, err
	}

	episodes, err := uc.syncAllEpisodes(ctx)
	if err != nil {
		return nil, err
	}

	satellites, err := uc.syncAllWorkSatellites(ctx)
	if err != nil {
		return nil, err
	}

	result := &SyncAnimesResult{Works: works, Episodes: episodes, Satellites: satellites}

	slog.InfoContext(ctx, "animes同期バッチが完了しました",
		"works_processed", result.Works.Processed,
		"works_created", result.Works.Created,
		"works_updated", result.Works.Updated,
		"works_unchanged", result.Works.Unchanged,
		"episodes_processed", result.Episodes.Processed,
		"episodes_created", result.Episodes.Created,
		"episodes_updated", result.Episodes.Updated,
		"episodes_unchanged", result.Episodes.Unchanged,
		"episodes_skipped_no_parent", result.Episodes.SkippedNoParent,
		"satellites_processed", result.Satellites.Processed,
		"satellites_created", result.Satellites.Created,
		"satellites_updated", result.Satellites.Updated,
		"satellites_deleted", result.Satellites.Deleted,
		"satellites_unchanged", result.Satellites.Unchanged,
		"satellites_skipped_no_anime", result.Satellites.SkippedNoAnime,
	)

	return result, nil
}

// syncAllWorksはworksテーブルをkeysetページネーションで走査し、1ページずつ
// リコンサイルして件数を積算する。カーソル (afterID) は反復ごとに厳密に増加するため
// ループは必ず終了し、空ページが終端を表す。
func (uc *SyncAnimesUsecase) syncAllWorks(ctx context.Context) (SyncWorksToAnimesResult, error) {
	var total SyncWorksToAnimesResult
	var afterID model.WorkID

	for {
		ids, err := uc.workIDPager.ListIDsAfter(ctx, afterID, uc.batchSize)
		if err != nil {
			return total, fmt.Errorf("works IDページの取得に失敗: %w", err)
		}
		if len(ids) == 0 {
			break
		}

		res, err := uc.worksSyncer.Execute(ctx, SyncWorksToAnimesInput{WorkIDs: ids})
		if err != nil {
			return total, fmt.Errorf("works同期ページの処理に失敗: %w", err)
		}

		total.Processed += res.Processed
		total.Created += res.Created
		total.Updated += res.Updated
		total.Unchanged += res.Unchanged

		afterID = ids[len(ids)-1]
	}

	return total, nil
}

// syncAllEpisodesはepisodesテーブルをkeysetページネーションで走査し、1ページ
// ずつリコンサイルして件数 (SkippedNoParentを含む) を積算する。
func (uc *SyncAnimesUsecase) syncAllEpisodes(ctx context.Context) (SyncEpisodesToAnimesResult, error) {
	var total SyncEpisodesToAnimesResult
	var afterID model.EpisodeID

	for {
		ids, err := uc.episodeIDPager.ListIDsAfter(ctx, afterID, uc.batchSize)
		if err != nil {
			return total, fmt.Errorf("episodes IDページの取得に失敗: %w", err)
		}
		if len(ids) == 0 {
			break
		}

		res, err := uc.episodesSyncer.Execute(ctx, SyncEpisodesToAnimesInput{EpisodeIDs: ids})
		if err != nil {
			return total, fmt.Errorf("episodes同期ページの処理に失敗: %w", err)
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

// syncAllWorkSatellitesはworksテーブルを再度keysetページネーションで走査し、各
// ページの別表をリコンサイルして件数 (SkippedNoAnimeを含む) を積算する。別表パスは同じ
// worksに紐づくためwork IDページャを再利用し、works.anime_idが大半の行で解決済みに
// なるようsyncAllWorksの後に走らせる。
func (uc *SyncAnimesUsecase) syncAllWorkSatellites(ctx context.Context) (SyncWorkSatellitesResult, error) {
	var total SyncWorkSatellitesResult
	var afterID model.WorkID

	for {
		ids, err := uc.workIDPager.ListIDsAfter(ctx, afterID, uc.batchSize)
		if err != nil {
			return total, fmt.Errorf("別表同期のworks IDページの取得に失敗: %w", err)
		}
		if len(ids) == 0 {
			break
		}

		res, err := uc.satellitesSyncer.Execute(ctx, SyncWorkSatellitesInput{WorkIDs: ids})
		if err != nil {
			return total, fmt.Errorf("別表同期ページの処理に失敗: %w", err)
		}

		total.Processed += res.Processed
		total.Created += res.Created
		total.Updated += res.Updated
		total.Deleted += res.Deleted
		total.Unchanged += res.Unchanged
		total.SkippedNoAnime += res.SkippedNoAnime

		afterID = ids[len(ids)-1]
	}

	return total, nil
}
