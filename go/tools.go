//go:build tools

package tools

// このファイルは開発ツールの依存関係を go.mod に記録するためのものです。
// ビルドタグ "tools" により、本番ビルドには含まれません。
import (
	_ "github.com/a-h/templ/cmd/templ"
	_ "github.com/mfridman/tparse"
	_ "golang.org/x/tools/cmd/goimports"
)
