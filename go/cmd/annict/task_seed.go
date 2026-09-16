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

// seedはseedタスクの本体で、マスターデータを投入したうえで、シード対象のテーブルを空に
// して開発用データを生成し直す。マスターデータはseeder.Runが空にするテーブルには含まれず、
// 作り直しの対象でもないが、開発用データの作り直しがマスターデータを欠いた状態で終わらない
// よう、seed-masterと同じ処理をここでも呼ぶ。
//
// 実行してよいかどうかは、共通の前処理がデータベースを開く前にguardSeedが判定する。
// master.Runとseeder.Runも同じ確認を繰り返し、このCLI以外の呼び出し元を防御する。CLIは
// ここでAPP_ENVを読み直さず、runWithDBが読み込み済みの設定をそのまま渡す。生成はRiverを
// 介さずシードのUseCaseを直接駆動する。シード生成はローカル開発の手順であり、対応する
// 定期実行が無いため。何をどの順で生成するかはseeder側の関心事。
func seed(ctx context.Context, cfg *config.Config, db *sql.DB, queries *query.Queries) error {
	if err := seedMaster(ctx, cfg, db, queries); err != nil {
		return err
	}

	return seeder.Run(ctx, cfg, db)
}
