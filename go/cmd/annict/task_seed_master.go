package main

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/seeder/master"
)

// seedMasterはseed-masterタスクの本体で、マスターデータをCSVから投入する。投入は主キーを
// 指定したupsertで既存のデータを消さないため、日常的なセットアップ手順から繰り返し呼べる。
// 実行してよいかどうかは、共通の前処理がデータベースを開く前にguardSeedが判定する。
// master.Runも同じ確認を繰り返し、このCLI以外の呼び出し元を防御する。
func seedMaster(ctx context.Context, cfg *config.Config, db *sql.DB, _ *query.Queries) error {
	return master.Run(ctx, cfg, db)
}
