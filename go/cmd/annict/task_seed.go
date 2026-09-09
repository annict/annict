package main

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/seeder"
)

// guardSeed rejects seed generation outside development and test before the shared
// task preamble opens or pings the database. seeder.Run repeats the same check so the
// destructive operation stays protected when called outside this CLI entry point.
//
// [Ja] guardSeed は共通のタスク前処理がデータベースを開いたり ping したりする前に、開発・
// テスト以外の環境でのシード生成を拒否する。seeder.Run も同じ確認を繰り返し、この CLI
// 以外から呼ばれた場合も破壊的な処理を保護する。
func guardSeed(cfg *config.Config) error {
	return seeder.EnsureSeedableEnv(cfg)
}

// seed is the body of the seed task: it empties the seeded tables and regenerates the
// development data. Whether it may run at all is decided by guardSeed before the shared
// preamble opens the database; seeder.Run repeats the check for callers outside this CLI.
// The CLI hands it the configuration runWithDB already loaded rather than reading APP_ENV
// again here. The generation goes straight through the seed usecases instead of River,
// since seeding is a local development step with no scheduled counterpart. What is
// generated, and in which order, belongs to seeder.
//
// [Ja] seed は seed タスクの本体で、シード対象のテーブルを空にして開発用データを生成し直す。
// 実行してよいかどうかは、共通の前処理がデータベースを開く前に guardSeed が判定する。
// seeder.Run も同じ確認を繰り返し、この CLI 以外の呼び出し元を防御する。CLI はここで APP_ENV
// を読み直さず、runWithDB が読み込み済みの設定をそのまま渡す。生成は River を介さずシードの
// UseCase を直接駆動する。シード生成はローカル開発の手順であり、対応する定期実行が無いため。
// 何をどの順で生成するかは seeder 側の関心事。
func seed(ctx context.Context, cfg *config.Config, db *sql.DB, _ *query.Queries) error {
	return seeder.Run(ctx, cfg, db)
}
