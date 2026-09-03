package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// main routes to a subcommand. The binary hosts the web server (serve) and
// operational one-off tasks (task) as subcommands, so the binary name stays
// role-neutral (`annict serve` rather than a `server` binary that also runs
// unrelated tasks). Running with no/unknown subcommand prints usage and exits
// non-zero rather than defaulting to serve, so every invocation site states its
// intent explicitly.
//
// [Ja] main はサブコマンドへ振り分ける。本バイナリは web サーバー (serve) と運用用の
// one-off タスク (task) をサブコマンドとして束ね、バイナリ名が役割に依存しないように
// する (無関係なタスクも担う `server` バイナリではなく `annict serve` とする)。
// サブコマンドが無い / 未知の場合は serve に既定せず usage を表示して非ゼロ終了し、
// 各起動箇所が意図を明示するようにする。
func main() {
	if err := run(context.Background(), os.Stdout, os.Args[1:], tasks); err != nil {
		// An error carrying the usage text ends with a newline, while a task failure
		// does not. Trimming first keeps either case from printing a trailing blank line.
		//
		// [Ja] usage テキストを含むエラーは改行で終わるが、タスクの失敗は改行で終わらない。
		// 先に落としておくことで、どちらの場合も末尾に空行が出ないようにする。
		fmt.Fprintln(os.Stderr, strings.TrimRight(err.Error(), "\n"))
		os.Exit(1)
	}
}

// run dispatches to the requested subcommand and reports failures as an error, so
// that the dispatch and the task bodies stay callable from tests. The failure is
// turned into an exit code by main alone. `serve` is the exception: it owns the
// process lifecycle (signal handling and shutdown) and exits on its own.
//
// [Ja] run は指定されたサブコマンドへ振り分け、失敗を error として返す。ディスパッチと
// 各タスクの本体をテストから呼べる状態に保つため。失敗を終了コードに変換するのは main
// だけが行う。`serve` は例外で、プロセスのライフサイクル (シグナル処理とシャットダウン)
// を自身で持ち、自ら終了する。
func run(ctx context.Context, out io.Writer, args []string, registry map[string]taskDef) error {
	if len(args) == 0 {
		return fmt.Errorf("サブコマンドを指定してください\n\n%s", usage(registry))
	}

	switch args[0] {
	case "serve":
		runServe()
		return nil
	case "task":
		return runTask(ctx, out, args[1:], registry)
	default:
		return fmt.Errorf("不明なサブコマンドです: %q\n\n%s", args[0], usage(registry))
	}
}

// usage returns the top-level usage text. It embeds the task listing so that a bare
// `annict` shows what can be run without a second invocation.
//
// [Ja] usage はトップレベルの usage テキストを返す。引数なしの `annict` だけで実行可能な
// ものが分かるよう、タスク一覧を埋め込む。
func usage(registry map[string]taskDef) string {
	return fmt.Sprintf(`使い方: annict <command>

コマンド:
  serve   HTTP サーバーを起動する
  task    運用タスクを 1 回実行する

タスク:
%s`, taskListing(registry))
}
