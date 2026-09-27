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

// newSyncAnimesUsecaseはDBハンドルとそのsqlcクエリから、フェーズ2のフル・
// リコンシリエーションバッチUseCaseを組み立てる。`serve` (毎時の定期ジョブとして登録)
// と `task sync-animes` (1回だけ同期実行) の双方が本ヘルパー経由で組み立てるため、両エントリ
// ポイントが共有する依存配線が1箇所にまとまる。バッチが依存するworks → episodesの
// 順序自体はSyncAnimesUsecase.Executeが担い、本ヘルパーでは決めない。
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
	// 別表リコンサイラはテーブルごとに1つずつ登録する (タスク2-8以降)。各リコンサイラ
	// はanime_id解決済みのworksについて第3パスで走り、anime未解決のworkは後続の実行へ
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

// syncAnimesはsync-animesタスクの本体で、フェーズ2のフル・リコンシリエーション
// バッチを実行する。`serve` が登録する毎時の定期ジョブと違いRiverを介さず、同じ
// SyncAnimesUsecaseを組み立てて直接Executeを呼ぶ。これにより運用者は (データ修正後など)
// スケジュールを待たずに任意のタイミングで再同期できる。テーブルごとの件数はExecute内で
// ログ出力されるため、本関数は失敗の報告のみを扱う。
func syncAnimes(ctx context.Context, _ *config.Config, db *sql.DB, queries *query.Queries) error {
	if _, err := newSyncAnimesUsecase(db, queries).Execute(ctx); err != nil {
		return fmt.Errorf("animes同期バッチの実行に失敗しました: %w", err)
	}

	return nil
}
