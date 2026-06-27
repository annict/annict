package main

import (
	"fmt"
	"os"
)

// main routes to a subcommand. The binary hosts the web server (serve) and
// operational one-off tasks as subcommands, so the binary name stays role-neutral
// (`annict serve` rather than a `server` binary that also runs unrelated tasks).
// Running with no/unknown subcommand prints usage and exits non-zero rather than
// defaulting to serve, so every invocation site states its intent explicitly.
//
// [Ja] main はサブコマンドへ振り分ける。本バイナリは web サーバー (serve) と運用用の
// one-off タスクをサブコマンドとして束ね、バイナリ名が役割に依存しないようにする
// (無関係なタスクも担う `server` バイナリではなく `annict serve` とする)。サブコマンド
// が無い / 未知の場合は serve に既定せず usage を表示して非ゼロ終了し、各起動箇所が
// 意図を明示するようにする。
func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "serve":
		runServe()
	case "sync-animes":
		runSyncAnimes()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

// usage prints the available subcommands to stderr.
//
// [Ja] usage は利用可能なサブコマンドを標準エラーに出力する。
func usage() {
	fmt.Fprint(os.Stderr, `usage: annict <command>

commands:
  serve         start the HTTP server
  sync-animes   run the works/episodes -> animes reconciliation once
`)
}
