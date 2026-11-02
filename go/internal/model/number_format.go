package model

import (
	"fmt"
	"strings"
)

// NumberFormatはエピソード番号のフォーマット定義です
type NumberFormat struct {
	ID         NumberFormatID
	Name       string
	SortNumber int32
	// Formatは話数を埋め込む書式 (例: "第%d話")。空のときはDataから話数番目の文字列を引く。
	// 値が入るのは自動生成のローダー (GetForEpisodeGenerationByID) のみ。
	Format string
	// Dataは話数ごとの表示 (例: "第一話", "第二話")。値が入るのはFormatと同じローダーのみ。
	Data []string
}

// EpisodeNumberは話数を書式に当てはめた表示の番号を返す。当てはめられないときは空文字列を
// 返し、呼び出し側が既定の表示にする。RailsのWork#formatted_numberと同じく、話数は整数部
// だけを使う。
func (f *NumberFormat) EpisodeNumber(rawNumber float64) string {
	number := int(rawNumber)

	if f.Format == "" {
		// Rubyの配列は負の添字で末尾から引くが、0話以下に末尾の表示を当てる意味は無いため、
		// 範囲外として既定の表示に任せる。
		if number < 1 || number > len(f.Data) {
			return ""
		}
		return f.Data[number-1]
	}

	// 書式に話数の指定子が無いと、fmt.Sprintfは余った引数を本文に書き出してしまう。Rubyの
	// String#% は余った引数を黙って捨てるため、書式をそのまま返して揃える。
	if !strings.Contains(f.Format, "%") {
		return f.Format
	}
	return fmt.Sprintf(f.Format, number)
}
