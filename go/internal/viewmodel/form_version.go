package viewmodel

import (
	"time"

	"github.com/annict/annict/go/internal/validator"
)

// FormNullVersion is the explicit version an edit form carries when the stored updated_at is
// NULL. The update side matches it with updated_at IS NULL; the first successful write advances
// the column to a timestamp, so another submit from the same NULL version conflicts. An empty
// value remains reserved for a request that stated no version.
//
// Both this and the layout below are the validator's constants: it reads the submitted version
// back and decides whether to accept it, so the round-trip format is single-sourced there and
// named here for the forms that write it.
//
// [Ja] FormNullVersion は保存済み updated_at が NULL のとき、編集フォームが運ぶ明示的な版。
// 更新側は updated_at IS NULL で照合し、最初に成功した書き込みがカラムを timestamp へ進めるため、
// 同じ NULL 版からの次の送信は競合する。空文字列は版を指定していない要求のために残す。
//
// 本定数と下の書式はいずれも validator 側の定数である。送信された版を読み戻して受理するかを
// 判断するのは validator のため、往復の書式の正本をそちらに 1 つ置き、ここでは書き出すフォームの
// ために名前を与える。
const FormNullVersion = validator.FormNullVersion

const formVersionLayout = validator.FormVersionLayout

// formatFormVersion renders a stored updated_at as the version its edit form carries, mapping a
// row without one onto the null sentinel. The two forms that carry a version share this so the
// value they write is the one the validator reads back.
//
// [Ja] formatFormVersion は保存済みの updated_at を、その編集フォームが運ぶ版として描画し、
// updated_at を持たない行をセンチネルに写像する。版を運ぶ 2 つのフォームがこれを共有することで、
// 書き出す値と validator が読み戻す値を一致させる。
func formatFormVersion(updatedAt *time.Time) string {
	if updatedAt == nil {
		return FormNullVersion
	}

	return updatedAt.UTC().Format(formVersionLayout)
}
