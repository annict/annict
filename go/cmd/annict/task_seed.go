package main

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/seeder"
)

// guardSeedは共通のタスク前処理がデータベースを開いたりpingしたりする前に、開発・
// テスト以外の環境でのシード生成を拒否する。seeder.Runも同じ確認を繰り返し、このCLI
// 以外から呼ばれた場合も破壊的な処理を保護する。
func guardSeed(cfg *config.Config) error {
	return seeder.EnsureSeedableEnv(cfg)
}

// seedはseedタスクの本体で、シード対象のテーブルを空にして開発用データを生成し直す。
// 実行してよいかどうかは、共通の前処理がデータベースを開く前にguardSeedが判定する。
// seeder.Runも同じ確認を繰り返し、このCLI以外の呼び出し元を防御する。CLIはここでAPP_ENV
// を読み直さず、runWithDBが読み込み済みの設定をそのまま渡す。生成はRiverを介さずシードの
// UseCaseを直接駆動する。シード生成はローカル開発の手順であり、対応する定期実行が無いため。
// 何をどの順で生成するかはseeder側の関心事。
func seed(ctx context.Context, cfg *config.Config, db *sql.DB, _ *query.Queries) error {
	return seeder.Run(ctx, cfg, db)
}
