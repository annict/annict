package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// mainはサブコマンドへ振り分ける。本バイナリはwebサーバー (serve) と運用用の
// one-offタスク (task) をサブコマンドとして束ね、バイナリ名が役割に依存しないように
// する (無関係なタスクも担う `server` バイナリではなく `annict serve` とする)。
// サブコマンドが無い / 未知の場合はserveに既定せずusageを表示して非ゼロ終了し、
// 各起動箇所が意図を明示するようにする。
func main() {
	if err := run(context.Background(), os.Stdout, os.Args[1:], tasks); err != nil {
		// usageテキストを含むエラーは改行で終わるが、タスクの失敗は改行で終わらない。
		// 先に落としておくことで、どちらの場合も末尾に空行が出ないようにする。
		fmt.Fprintln(os.Stderr, strings.TrimRight(err.Error(), "\n"))
		os.Exit(1)
	}
}

// runは指定されたサブコマンドへ振り分け、失敗をerrorとして返す。ディスパッチと
// 各タスクの本体をテストから呼べる状態に保つため。失敗を終了コードに変換するのはmain
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

// usageはトップレベルのusageテキストを返す。引数なしの `annict` だけで実行可能な
// ものが分かるよう、タスク一覧を埋め込む。
func usage(registry map[string]taskDef) string {
	return fmt.Sprintf(`使い方: annict <command>

コマンド:
  serve   HTTPサーバーを起動する
  task    運用タスクを1回実行する

タスク:
%s`, taskListing(registry))
}
