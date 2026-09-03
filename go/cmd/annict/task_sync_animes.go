package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
)

// newSyncAnimesUsecase wires the phase 2 full-reconciliation batch usecase from a
// DB handle and its sqlc queries. Both `serve` (which registers it as an hourly
// periodic job) and `task sync-animes` (which runs it once, synchronously) build it
// through this helper so the dependency wiring shared by both entry points lives in
// one place. The works-before-episodes ordering the batch depends on is enforced by
// SyncAnimesUsecase.Execute, not here.
//
// [Ja] newSyncAnimesUsecase は DB ハンドルとその sqlc クエリから、フェーズ 2 のフル・
// リコンシリエーションバッチ UseCase を組み立てる。`serve` (毎時の定期ジョブとして登録)
// と `task sync-animes` (1 回だけ同期実行) の双方が本ヘルパー経由で組み立てるため、両エントリ
// ポイントが共有する依存配線が 1 箇所にまとまる。バッチが依存する works → episodes の
// 順序自体は SyncAnimesUsecase.Execute が担い、本ヘルパーでは決めない。
func newSyncAnimesUsecase(db *sql.DB, queries *query.Queries) *usecase.SyncAnimesUsecase {
	workRepo := repository.NewWorkRepository(queries)
	episodeRepo := repository.NewEpisodeRepository(queries)
	animeRepo := repository.NewAnimeRepository(queries)
	animeClassificationRepo := repository.NewAnimeClassificationRepository(queries)
	animeExternalIDRepo := repository.NewAnimeExternalIDRepository(queries)
	animeLinkRepo := repository.NewAnimeLinkRepository(queries)
	animeOfficialAccountRepo := repository.NewAnimeOfficialAccountRepository(queries)
	animeHashtagRepo := repository.NewAnimeHashtagRepository(queries)
	animeSeasonRepo := repository.NewAnimeSeasonRepository(queries)
	animeEventRepo := repository.NewAnimeEventRepository(queries)
	syncWorksToAnimesUC := usecase.NewSyncWorksToAnimesUsecase(db, workRepo, animeRepo, animeClassificationRepo)
	syncEpisodesToAnimesUC := usecase.NewSyncEpisodesToAnimesUsecase(db, episodeRepo, animeRepo, animeClassificationRepo)
	// Satellite reconcilers are registered one per table (tasks 2-8 onward). Each
	// runs in the third pass for works whose anime_id is already resolved; works
	// still pending an anime are deferred to a later run.
	//
	// [Ja] 別表リコンサイラはテーブルごとに 1 つずつ登録する (タスク 2-8 以降)。各リコンサイラ
	// は anime_id 解決済みの works について第 3 パスで走り、anime 未解決の work は後続の実行へ
	// 繰り延べる。
	syncAnimeExternalIDsReconciler := usecase.NewSyncAnimeExternalIDsUsecase(db, animeExternalIDRepo)
	syncAnimeLinksReconciler := usecase.NewSyncAnimeLinksUsecase(db, animeLinkRepo)
	syncAnimeOfficialAccountsReconciler := usecase.NewSyncAnimeOfficialAccountsUsecase(db, animeOfficialAccountRepo)
	syncAnimeHashtagsReconciler := usecase.NewSyncAnimeHashtagsUsecase(db, animeHashtagRepo)
	syncAnimeSeasonsReconciler := usecase.NewSyncAnimeSeasonsUsecase(db, animeSeasonRepo)
	syncAnimeEventsReconciler := usecase.NewSyncAnimeEventsUsecase(db, animeEventRepo)
	syncWorkSatellitesUC := usecase.NewSyncWorkSatellitesUsecase(workRepo, syncAnimeExternalIDsReconciler, syncAnimeLinksReconciler, syncAnimeOfficialAccountsReconciler, syncAnimeHashtagsReconciler, syncAnimeSeasonsReconciler, syncAnimeEventsReconciler)
	return usecase.NewSyncAnimesUsecase(workRepo, episodeRepo, syncWorksToAnimesUC, syncEpisodesToAnimesUC, syncWorkSatellitesUC, usecase.DefaultSyncAnimesBatchSize)
}

// syncAnimes is the body of the sync-animes task: it runs the phase 2 full-reconciliation
// batch. Unlike the hourly periodic job that `serve` registers, it does not go through
// River: it builds the same SyncAnimesUsecase and calls Execute directly, so that an
// operator can re-sync on demand (e.g. after a data fix) without waiting for the
// schedule. The per-table counts are logged inside Execute, so this only reports
// failures.
//
// [Ja] syncAnimes は sync-animes タスクの本体で、フェーズ 2 のフル・リコンシリエーション
// バッチを実行する。`serve` が登録する毎時の定期ジョブと違い River を介さず、同じ
// SyncAnimesUsecase を組み立てて直接 Execute を呼ぶ。これにより運用者は (データ修正後など)
// スケジュールを待たずに任意のタイミングで再同期できる。テーブルごとの件数は Execute 内で
// ログ出力されるため、本関数は失敗の報告のみを扱う。
func syncAnimes(ctx context.Context, _ *config.Config, db *sql.DB, queries *query.Queries) error {
	if _, err := newSyncAnimesUsecase(db, queries).Execute(ctx); err != nil {
		return fmt.Errorf("animes 同期バッチの実行に失敗しました: %w", err)
	}

	return nil
}
