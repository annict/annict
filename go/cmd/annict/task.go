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
// 自前で読み込んだり開いたりせず、対応する引数を無名で受ける。argsはタスク名の後ろに続く
// 位置引数で、個数はrunTaskがtaskDef.paramsと照合済みのため、本体は値の解釈だけを行う。
type taskBody func(ctx context.Context, cfg *config.Config, db *sql.DB, queries *query.Queries, args []string) error

// taskGuardは設定の読み込み後、データベースを開く前にタスクを実行してよいか確認する。
// タスク固有のガードが不要な場合はnilのままとする。
type taskGuard func(cfg *config.Config) error

// taskDefは運用タスク1件を表す。タスク一覧に表示する1行の説明、受け付ける位置引数の名前、
// タスクを実行する前処理の任意ガード、本体、そしてディスパッチャが呼ぶ関数を持つ。runはnewTasks
// がguardとbodyから導出するため、登録が実行経路を書くのは1回で済み、名前と別タスクの
// 本体を結び付けてしまう中継関数が入り込まない。newTasksを通さずに組み立てた定義 (テストの
// スタブなど) はrunを直接持ち、guardとbodyはnilのままとする。
type taskDef struct {
	desc string
	// paramsは位置引数の名前を受け付ける順に並べたもの。一覧とusageに表示し、runTaskが
	// 引数の個数をこれと照合する。引数を取らないタスクはnilのままとする。
	params []string
	guard  taskGuard
	body   taskBody
	run    func(ctx context.Context, args []string) error
}

// tasksはコマンドラインで指定するタスク名と、それが実行するタスクの対応。タスク
// 一覧もusageも本mapから生成するため、タスクの追加は本レジストリへの登録1箇所で
// 済み、ヘルプ文言を別途書き足す必要は無い。
var tasks = newTasks(map[string]taskDef{
	"cleanup-anonymous-sessions": {
		desc: "最終アクセスから1日を過ぎた未ログインのセッションを削除する",
		body: cleanupAnonymousSessions,
	},
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
	"delete-user": {
		desc:   "論理削除済みのユーザーのデータを最後まで削除する (退会・利用停止したユーザーが対象)",
		params: []string{"user_id"},
		body:   deleteUser,
	},
	"seed": {
		desc:  "開発用のシードデータを生成する (dev / testでのみ実行できる)",
		guard: guardSeed,
		body:  seed,
	},
	"seed-master": {
		desc:  "マスターデータをCSVから投入する (dev / testでのみ実行できる)",
		guard: guardSeed,
		body:  seedMaster,
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

// taskWithDBはタスクのガードと本体を、ディスパッチャが呼ぶ関数に変える。
// 全タスクに共通するデータベースの前処理を束ねる。
func taskWithDB(guard taskGuard, body taskBody) func(ctx context.Context, args []string) error {
	return func(ctx context.Context, args []string) error {
		return runWithDB(ctx, guard, body, args)
	}
}

// listTaskNameは組み込みの一覧表示コマンドに予約した名前。タスクを実行せず
// 呼び出し元のWriterへ出力するためレジストリには登録しないが、一覧が `annict task`
// の受け付ける名前を網羅するよう、登録済みタスクと並べて描画する。
const listTaskName = "list"

// runTaskは `annict task <name> [<args>...]` を振り分ける。名前の解決を引数のチェックより
// 先に行い、打ち間違えた名前が「引数の誤り」ではなく「未知のタスク」として報告される
// ようにする。引数の個数がタスクの受け付ける数と異なる場合は、余分な引数を無視せず拒否する。
// 無視すると、打ち間違えたフラグが効いたように見えてしまうため。個数の照合はデータベースを
// 開く前に行い、引数を忘れた実行が接続まで進まないようにする。
func runTask(ctx context.Context, out io.Writer, args []string, registry map[string]taskDef) error {
	if len(args) == 0 {
		return fmt.Errorf("タスク名を指定してください\n\n%s", taskUsage(registry))
	}

	name := args[0]
	task, ok := registry[name]
	if !ok && name != listTaskName {
		return fmt.Errorf("不明なタスクです: %q\n\n%s", name, taskUsage(registry))
	}

	taskArgs := args[1:]
	if len(taskArgs) != len(task.params) {
		if len(task.params) == 0 {
			return fmt.Errorf("タスク %qは引数を取りません\n\n%s", name, taskUsage(registry))
		}
		return fmt.Errorf("タスク %qには引数 %sを指定してください\n\n%s", name, formatParams(task.params), taskUsage(registry))
	}

	if name == listTaskName {
		_, err := io.WriteString(out, taskListing(registry))
		return err
	}

	return task.run(ctx, taskArgs)
}

// taskUsageはtaskサブコマンドのusageテキストを返す。
func taskUsage(registry map[string]taskDef) string {
	return "使い方: annict task <name> [<args>...]\n\nタスク:\n" + taskListing(registry)
}

// formatParamsは位置引数の名前を `<user_id>` の形で空白区切りに並べる。
func formatParams(params []string) string {
	formatted := make([]string, len(params))
	for i, param := range params {
		formatted[i] = "<" + param + ">"
	}
	return strings.Join(formatted, " ")
}

// taskListingは利用可能なタスク名 (位置引数があれば併記) とその説明を1行ずつ描画する。
// レジストリはmapであるため、実行ごとに出力が揺れないよう名前を並べ替える。
func taskListing(registry map[string]taskDef) string {
	names := slices.Sorted(maps.Keys(registry))
	synopses := make(map[string]string, len(names))

	width := len(listTaskName)
	for _, name := range names {
		synopsis := name
		if params := registry[name].params; len(params) > 0 {
			synopsis += " " + formatParams(params)
		}
		synopses[name] = synopsis
		width = max(width, len(synopsis))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "  %-*s  %s\n", width, listTaskName, "実行可能なタスクを一覧表示する")
	for _, name := range names {
		fmt.Fprintf(&b, "  %-*s  %s\n", width, synopses[name], registry[name].desc)
	}

	return b.String()
}

// runWithDBは設定を読み込み、タスク固有のガードを適用してからデータベースを開き、
// ハンドルとそのsqlcクエリを位置引数と一緒にbodyへ渡して、bodyが戻ったらハンドルを閉じる。全タスクに
// 共通する前処理である。ガードはsql.OpenとPingContextより先に実行するため、現在の環境で
// 許可されていないタスクが接続することは無い。コネクションプールはドライバの既定値のままと
// する。タスクはデータベースを開いて1度処理したら終了するため、`serve` が長時間動作する
// プールに対して行うサイズ調整は、ここでは働く先が無い。
//
// `serve` と違いSentryは意図的に初期化しない。タスクは `dokku run` 等で運用者がstderrから
// 直接失敗を確認する手動アドホック実行であり、同じ処理の定期実行はRiverのSentryミドル
// ウェアが引き続き捕捉する。one-offのCLI失敗を報告しても、監視の穴を塞がずノイズが増える
// だけになる。
func runWithDB(ctx context.Context, guard taskGuard, body taskBody, args []string) error {
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

	return body(ctx, cfg, db, query.New(db), args)
}
