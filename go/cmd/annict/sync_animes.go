package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
)

// newSyncAnimesUsecase wires the phase 2 full-reconciliation batch usecase from a
// DB handle and its sqlc queries. Both `serve` (which registers it as an hourly
// periodic job) and `sync-animes` (which runs it once, synchronously) build it
// through this helper so the dependency wiring shared by both entry points lives in
// one place. The works-before-episodes ordering the batch depends on is enforced by
// SyncAnimesUsecase.Execute, not here.
//
// [Ja] newSyncAnimesUsecase は DB ハンドルとその sqlc クエリから、フェーズ 2 のフル・
// リコンシリエーションバッチ UseCase を組み立てる。`serve` (毎時の定期ジョブとして登録)
// と `sync-animes` (1 回だけ同期実行) の双方が本ヘルパー経由で組み立てるため、両エントリ
// ポイントが共有する依存配線が 1 箇所にまとまる。バッチが依存する works → episodes の
// 順序自体は SyncAnimesUsecase.Execute が担い、本ヘルパーでは決めない。
func newSyncAnimesUsecase(db *sql.DB, queries *query.Queries) *usecase.SyncAnimesUsecase {
	workRepo := repository.NewWorkRepository(queries)
	episodeRepo := repository.NewEpisodeRepository(queries)
	animeRepo := repository.NewAnimeRepository(queries)
	animeClassificationRepo := repository.NewAnimeClassificationRepository(queries)
	syncWorksToAnimesUC := usecase.NewSyncWorksToAnimesUsecase(db, workRepo, animeRepo, animeClassificationRepo)
	syncEpisodesToAnimesUC := usecase.NewSyncEpisodesToAnimesUsecase(db, episodeRepo, animeRepo, animeClassificationRepo)
	return usecase.NewSyncAnimesUsecase(workRepo, episodeRepo, syncWorksToAnimesUC, syncEpisodesToAnimesUC, usecase.DefaultSyncAnimesBatchSize)
}

// runSyncAnimes runs the phase 2 full-reconciliation batch once, synchronously, and
// exits with the result. Unlike the hourly periodic job that `serve` registers, this
// does not go through River: it builds the same SyncAnimesUsecase and calls Execute
// directly, so an operator can re-sync on demand (e.g. after a data fix) without
// waiting for the schedule. The per-table counts are logged inside Execute, so this
// only handles errors and the exit code.
//
// [Ja] runSyncAnimes はフェーズ 2 のフル・リコンシリエーションバッチを 1 回だけ同期実行し、
// 結果に応じて終了する。`serve` が登録する毎時の定期ジョブと違い River を介さず、同じ
// SyncAnimesUsecase を組み立てて直接 Execute を呼ぶ。これにより運用者は (データ修正後など)
// スケジュールを待たずに任意のタイミングで再同期できる。テーブルごとの件数は Execute 内で
// ログ出力されるため、本関数はエラー処理と終了コードのみを扱う。
func runSyncAnimes() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("設定の読み込みに失敗しました", "error", err)
		os.Exit(1)
	}

	// `serve` initializes Sentry here; this one-off command intentionally does not.
	// It is a manual ad-hoc run (e.g. `dokku run`) where the operator reads failures
	// straight from stderr, so routing transient CLI failures to Sentry adds only noise.
	//
	// [Ja] `serve` はここで Sentry を初期化するが、本 one-off コマンドは意図的に行わない。
	// `dokku run` 等で運用者が stderr から直接失敗を確認する手動アドホック実行のため、
	// 一過性の CLI 失敗を Sentry に流してもノイズになるだけ。

	db, err := sql.Open("postgres", cfg.DatabaseDSN())
	if err != nil {
		slog.Error("データベースへの接続に失敗しました", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Warn("データベース接続のクローズに失敗しました", "error", err)
		}
	}()

	if err := db.Ping(); err != nil {
		slog.Error("データベースへの疎通確認に失敗しました", "error", err)
		os.Exit(1)
	}

	queries := query.New(db)
	syncAnimesUC := newSyncAnimesUsecase(db, queries)

	if _, err := syncAnimesUC.Execute(context.Background()); err != nil {
		slog.Error("animes 同期バッチの実行に失敗しました", "error", err)
		os.Exit(1)
	}
}
