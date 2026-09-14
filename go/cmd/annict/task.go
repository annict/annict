package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
)

// taskBodyはタスクを1回、読み込み済みの設定と開かれたデータベースに対して実行する。
// 設定・接続・その後始末はrunWithDBが与えるため、タスクごとのファイルにはそのタスク自身の
// UseCaseの配線と呼び出しだけが残る。これらの依存のいずれかを使わないタスクは、その依存を
// 自前で読み込んだり開いたりせず、対応する引数を無名で受ける。
type taskBody func(ctx context.Context, cfg *config.Config, db *sql.DB, queries *query.Queries) error

// taskGuardは設定の読み込み後、データベースを開く前にタスクを実行してよいか確認する。
// タスク固有のガードが不要な場合はnilのままとする。
type taskGuard func(cfg *config.Config) error

// taskDefは運用タスク1件を表す。タスク一覧に表示する1行の説明、タスクを実行する
// 前処理の任意ガード、本体、そしてディスパッチャが呼ぶ引数なしの関数を持つ。runはnewTasks
// がguardとbodyから導出するため、登録が実行経路を書くのは1回で済み、名前と別タスクの
// 本体を結び付けてしまう中継関数が入り込まない。newTasksを通さずに組み立てた定義 (テストの
// スタブなど) はrunを直接持ち、guardとbodyはnilのままとする。
type taskDef struct {
	desc  string
	guard taskGuard
	body  taskBody
	run   func(ctx context.Context) error
}

// tasksはコマンドラインで指定するタスク名と、それが実行するタスクの対応。タスク
// 一覧もusageも本mapから生成するため、タスクの追加は本レジストリへの登録1箇所で
// 済み、ヘルプ文言を別途書き足す必要は無い。
var tasks = newTasks(map[string]taskDef{
	"cleanup-expired-sessions": {
		desc: "最終アクセスから30日を過ぎたセッションを削除する",
		body: cleanupExpiredSessions,
	},
	"cleanup-expired-sign-in-codes": {
		desc: "24時間以上前に期限切れ・使用済みになったログインコードを削除する",
		body: cleanupExpiredSignInCodes,
	},
	"cleanup-expired-tokens": {
		desc: "24時間以上前に期限切れ・使用済みになったパスワードリセットトークンを削除する",
		body: cleanupExpiredTokens,
	},
	"seed": {
		desc:  "開発用のシードデータを生成する (dev / testでのみ実行できる)",
		guard: guardSeed,
		body:  seed,
	},
	"sync-animes": {
		desc: "works/episodes → animesのリコンサイルを1回実行する",
		body: syncAnimes,
	},
})

// newTasksは各登録のrunをguardとbodyから導出して補完する。
func newTasks(defs map[string]taskDef) map[string]taskDef {
	registry := make(map[string]taskDef, len(defs))
	for name, def := range defs {
		def.run = taskWithDB(def.guard, def.body)
		registry[name] = def
	}

	return registry
}

// taskWithDBはタスクのガードと本体を、ディスパッチャが呼ぶ引数なしの関数に変える。
// 全タスクに共通するデータベースの前処理を束ねる。
func taskWithDB(guard taskGuard, body taskBody) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return runWithDB(ctx, guard, body)
	}
}

// listTaskNameは組み込みの一覧表示コマンドに予約した名前。タスクを実行せず
// 呼び出し元のWriterへ出力するためレジストリには登録しないが、一覧が `annict task`
// の受け付ける名前を網羅するよう、登録済みタスクと並べて描画する。
const listTaskName = "list"

// runTaskは `annict task <name>` を振り分ける。名前の解決を引数のチェックより先に
// 行い、打ち間違えた名前が「引数を取らない」ではなく「未知のタスク」として報告される
// ようにする。タスクは引数を取らないため、名前の後ろに続く引数は無視せず拒否する。
// 無視すると、打ち間違えたフラグが効いたように見えてしまうため。
func runTask(ctx context.Context, out io.Writer, args []string, registry map[string]taskDef) error {
	if len(args) == 0 {
		return fmt.Errorf("タスク名を指定してください\n\n%s", taskUsage(registry))
	}

	name := args[0]
	task, ok := registry[name]
	if !ok && name != listTaskName {
		return fmt.Errorf("不明なタスクです: %q\n\n%s", name, taskUsage(registry))
	}

	if len(args) > 1 {
		return fmt.Errorf("タスク %qは引数を取りません\n\n%s", name, taskUsage(registry))
	}

	if name == listTaskName {
		_, err := io.WriteString(out, taskListing(registry))
		return err
	}

	return task.run(ctx)
}

// taskUsageはtaskサブコマンドのusageテキストを返す。
func taskUsage(registry map[string]taskDef) string {
	return "使い方: annict task <name>\n\nタスク:\n" + taskListing(registry)
}

// taskListingは利用可能なタスク名とその説明を1行ずつ描画する。レジストリはmap
// であるため、実行ごとに出力が揺れないよう名前を並べ替える。
func taskListing(registry map[string]taskDef) string {
	names := slices.Sorted(maps.Keys(registry))

	width := len(listTaskName)
	for _, name := range names {
		width = max(width, len(name))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "  %-*s  %s\n", width, listTaskName, "実行可能なタスクを一覧表示する")
	for _, name := range names {
		fmt.Fprintf(&b, "  %-*s  %s\n", width, name, registry[name].desc)
	}

	return b.String()
}

// runWithDBは設定を読み込み、タスク固有のガードを適用してからデータベースを開き、
// ハンドルとそのsqlcクエリをbodyへ渡して、bodyが戻ったらハンドルを閉じる。全タスクに
// 共通する前処理である。ガードはsql.OpenとPingContextより先に実行するため、現在の環境で
// 許可されていないタスクが接続することは無い。コネクションプールはドライバの既定値のままと
// する。タスクはデータベースを開いて1度処理したら終了するため、`serve` が長時間動作する
// プールに対して行うサイズ調整は、ここでは働く先が無い。
//
// `serve` と違いSentryは意図的に初期化しない。タスクは `dokku run` 等で運用者がstderrから
// 直接失敗を確認する手動アドホック実行であり、同じ処理の定期実行はRiverのSentryミドル
// ウェアが引き続き捕捉する。one-offのCLI失敗を報告しても、監視の穴を塞がずノイズが増える
// だけになる。
func runWithDB(ctx context.Context, guard taskGuard, body taskBody) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("設定の読み込みに失敗しました: %w", err)
	}
	if guard != nil {
		if err := guard(cfg); err != nil {
			return err
		}
	}

	db, err := sql.Open("postgres", cfg.DatabaseDSN())
	if err != nil {
		return fmt.Errorf("データベースへの接続に失敗しました: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.WarnContext(ctx, "データベース接続のクローズに失敗しました", "error", err)
		}
	}()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("データベースへの疎通確認に失敗しました: %w", err)
	}

	return body(ctx, cfg, db, query.New(db))
}
