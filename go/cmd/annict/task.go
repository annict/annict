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

// taskBody runs one task once, against a loaded configuration and an open database.
// The configuration, the connection and its teardown are supplied by runWithDB, so a
// task file holds only the wiring of its own usecase and the call into it. A task that
// does not use one of these dependencies declares that parameter blank rather than
// loading or opening the dependency on its own.
//
// [Ja] taskBody はタスクを 1 回、読み込み済みの設定と開かれたデータベースに対して実行する。
// 設定・接続・その後始末は runWithDB が与えるため、タスクごとのファイルにはそのタスク自身の
// UseCase の配線と呼び出しだけが残る。これらの依存のいずれかを使わないタスクは、その依存を
// 自前で読み込んだり開いたりせず、対応する引数を無名で受ける。
type taskBody func(ctx context.Context, cfg *config.Config, db *sql.DB, queries *query.Queries) error

// taskGuard checks whether a task may run after the configuration is loaded and before
// the database is opened. Tasks without a task-specific guard leave it nil.
//
// [Ja] taskGuard は設定の読み込み後、データベースを開く前にタスクを実行してよいか確認する。
// タスク固有のガードが不要な場合は nil のままとする。
type taskGuard func(cfg *config.Config) error

// taskDef describes one operational task: the one-line description shown in the task
// listing, the optional pre-database guard, the body that runs the task, and the
// argument-less function the dispatcher calls. run is derived from guard and body by
// newTasks, so a registration names its execution path once and no intermediate
// function can pair a name with another task's body. A definition built outside
// newTasks (a test stub, say) sets run directly and leaves guard and body nil.
//
// [Ja] taskDef は運用タスク 1 件を表す。タスク一覧に表示する 1 行の説明、タスクを実行する
// 前処理の任意ガード、本体、そしてディスパッチャが呼ぶ引数なしの関数を持つ。run は newTasks
// が guard と body から導出するため、登録が実行経路を書くのは 1 回で済み、名前と別タスクの
// 本体を結び付けてしまう中継関数が入り込まない。newTasks を通さずに組み立てた定義 (テストの
// スタブなど) は run を直接持ち、guard と body は nil のままとする。
type taskDef struct {
	desc  string
	guard taskGuard
	body  taskBody
	run   func(ctx context.Context) error
}

// tasks maps the name given on the command line to the task it runs. Both the task
// listing and the usage text are generated from this map, so adding a task takes a
// single registration here and needs no separate help text.
//
// [Ja] tasks はコマンドラインで指定するタスク名と、それが実行するタスクの対応。タスク
// 一覧も usage も本 map から生成するため、タスクの追加は本レジストリへの登録 1 箇所で
// 済み、ヘルプ文言を別途書き足す必要は無い。
var tasks = newTasks(map[string]taskDef{
	"cleanup-expired-sessions": {
		desc: "最終アクセスから 30 日を過ぎたセッションを削除する",
		body: cleanupExpiredSessions,
	},
	"cleanup-expired-sign-in-codes": {
		desc: "24 時間以上前に期限切れ・使用済みになったログインコードを削除する",
		body: cleanupExpiredSignInCodes,
	},
	"cleanup-expired-tokens": {
		desc: "24 時間以上前に期限切れ・使用済みになったパスワードリセットトークンを削除する",
		body: cleanupExpiredTokens,
	},
	"seed": {
		desc:  "開発用のシードデータを生成する (dev / test でのみ実行できる)",
		guard: guardSeed,
		body:  seed,
	},
	"sync-animes": {
		desc: "works/episodes → animes のリコンサイルを 1 回実行する",
		body: syncAnimes,
	},
})

// newTasks completes every registration by deriving its run function from its guard
// and body.
//
// [Ja] newTasks は各登録の run を guard と body から導出して補完する。
func newTasks(defs map[string]taskDef) map[string]taskDef {
	registry := make(map[string]taskDef, len(defs))
	for name, def := range defs {
		def.run = taskWithDB(def.guard, def.body)
		registry[name] = def
	}

	return registry
}

// taskWithDB turns a task guard and body into the argument-less function the dispatcher
// calls, binding them to the database preamble that every task shares.
//
// [Ja] taskWithDB はタスクのガードと本体を、ディスパッチャが呼ぶ引数なしの関数に変える。
// 全タスクに共通するデータベースの前処理を束ねる。
func taskWithDB(guard taskGuard, body taskBody) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return runWithDB(ctx, guard, body)
	}
}

// listTaskName is the reserved name of the built-in listing command. It stays out of
// the registry because it writes to the caller's writer instead of running a task,
// but it is rendered together with the registered tasks so that the listing covers
// every name `annict task` accepts.
//
// [Ja] listTaskName は組み込みの一覧表示コマンドに予約した名前。タスクを実行せず
// 呼び出し元の Writer へ出力するためレジストリには登録しないが、一覧が `annict task`
// の受け付ける名前を網羅するよう、登録済みタスクと並べて描画する。
const listTaskName = "list"

// runTask dispatches `annict task <name>`. The name is resolved before the arguments
// are checked, so that a mistyped name is reported as unknown rather than as a task
// that takes no arguments. Tasks take no arguments, so anything after the name is
// rejected rather than ignored: a mistyped flag would otherwise look like it took
// effect.
//
// [Ja] runTask は `annict task <name>` を振り分ける。名前の解決を引数のチェックより先に
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
		return fmt.Errorf("タスク %q は引数を取りません\n\n%s", name, taskUsage(registry))
	}

	if name == listTaskName {
		_, err := io.WriteString(out, taskListing(registry))
		return err
	}

	return task.run(ctx)
}

// taskUsage returns the usage text of the task subcommand.
//
// [Ja] taskUsage は task サブコマンドの usage テキストを返す。
func taskUsage(registry map[string]taskDef) string {
	return "使い方: annict task <name>\n\nタスク:\n" + taskListing(registry)
}

// taskListing renders one line per available task name with its description. The
// names are sorted so that the output stays stable across runs even though the
// registry is a map.
//
// [Ja] taskListing は利用可能なタスク名とその説明を 1 行ずつ描画する。レジストリは map
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

// runWithDB loads the configuration, applies the task-specific guard, opens the
// database and hands the handle together with its sqlc queries to body, closing the
// handle once body returns. It is the preamble every task shares. The guard runs before
// sql.Open and PingContext so a task that is not permitted in the current environment
// cannot connect. The connection pool is left at the driver defaults: a task opens the
// database, makes one pass and exits, so the sizing `serve` applies to a long-lived pool
// has nothing to act on here.
//
// Sentry is deliberately not initialized, unlike in `serve`. Tasks are manual ad-hoc
// runs (e.g. `dokku run`) whose failures the operator reads straight from stderr, and
// the scheduled runs of the same work stay covered by River's Sentry middleware, so
// reporting one-off CLI failures would add noise without closing a gap in monitoring.
//
// [Ja] runWithDB は設定を読み込み、タスク固有のガードを適用してからデータベースを開き、
// ハンドルとその sqlc クエリを body へ渡して、body が戻ったらハンドルを閉じる。全タスクに
// 共通する前処理である。ガードは sql.Open と PingContext より先に実行するため、現在の環境で
// 許可されていないタスクが接続することは無い。コネクションプールはドライバの既定値のままと
// する。タスクはデータベースを開いて 1 度処理したら終了するため、`serve` が長時間動作する
// プールに対して行うサイズ調整は、ここでは働く先が無い。
//
// `serve` と違い Sentry は意図的に初期化しない。タスクは `dokku run` 等で運用者が stderr から
// 直接失敗を確認する手動アドホック実行であり、同じ処理の定期実行は River の Sentry ミドル
// ウェアが引き続き捕捉する。one-off の CLI 失敗を報告しても、監視の穴を塞がずノイズが増える
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
