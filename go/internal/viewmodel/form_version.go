package viewmodel

import (
	"time"

	"github.com/annict/annict/go/internal/validator"
)

// FormNullVersionは保存済みupdated_atがNULLのとき、編集フォームが運ぶ明示的な版。
// 更新側はupdated_at IS NULLで照合し、最初に成功した書き込みがカラムをtimestampへ進めるため、
// 同じNULL版からの次の送信は競合する。空文字列は版を指定していない要求のために残す。
//
// 本定数と下の書式はいずれもvalidator側の定数である。送信された版を読み戻して受理するかを
// 判断するのはvalidatorのため、往復の書式の正本をそちらに1つ置き、ここでは書き出すフォームの
// ために名前を与える。
const FormNullVersion = validator.FormNullVersion

const formVersionLayout = validator.FormVersionLayout

// formatFormVersionは保存済みのupdated_atを、その編集フォームが運ぶ版として描画し、
// updated_atを持たない行をセンチネルに写像する。版を運ぶ2つのフォームがこれを共有することで、
// 書き出す値とvalidatorが読み戻す値を一致させる。
func formatFormVersion(updatedAt *time.Time) string {
	if updatedAt == nil {
		return FormNullVersion
	}

	return updatedAt.UTC().Format(formVersionLayout)
}
