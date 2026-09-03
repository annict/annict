package validator

import "time"

// FormNullVersion is the explicit version an edit form carries when the stored updated_at is
// NULL. The shared database leaves the column nullable on the tables the Annict DB admin edits,
// so "no timestamp" is a state a form has to be able to state; an empty request field means
// instead that no version was stated at all and is refused.
//
// The forms render the value and this package reads it back, so the literal lives here, on the
// side that decides whether a submitted version is acceptable.
//
// [Ja] FormNullVersion は、保存済みの updated_at が NULL のとき編集フォームが運ぶ明示的な版。
// Annict DB 管理画面が編集する各テーブルでは共有 DB が同カラムを NULL 許容にしているため、
// 「タイムスタンプが無い」ことをフォームが表明できる必要がある。リクエストの値が空であることは、
// そもそも版が示されていないことを意味し、拒否する。
//
// フォームが値を描画し、本パッケージが読み戻すため、リテラルは、送信された版を受け入れるかを
// 判断する側であるここに置く。
const FormNullVersion = "null"

// FormVersionLayout is the format a version travels in. It keeps the sub-second digits that
// separate two writes made within the same second, and parses back to the instant it came from
// so the update can match it against the stored column.
//
// [Ja] FormVersionLayout は版を往復させる書式。同一秒内の 2 つの書き込みを区別する秒未満の桁を
// 保ち、元の時刻へパースし直せるため、更新側は保存済みのカラムと照合できる。
const FormVersionLayout = time.RFC3339Nano

// parseFormVersion reads the version a submit states, reporting false when the field is empty or
// holds neither the null sentinel nor a timestamp in the round-trip format. A nil time with true
// is the sentinel, which stands for a stored updated_at of NULL.
//
// Every Annict DB edit form carries its version the same way, so the two forms that do share
// this reading rather than each deciding what an acceptable version looks like.
//
// [Ja] parseFormVersion は送信が示す版を読み取り、フィールドが空か、往復書式のタイムスタンプでも
// null のセンチネルでもない場合に false を返す。nil の時刻と true の組はセンチネルで、保存済みの
// updated_at が NULL であることを表す。
//
// Annict DB のどの編集フォームも同じ形で版を運ぶため、版を運ぶ 2 つのフォームはこの読み取りを
// 共有し、受け入れられる版の形をそれぞれで決めないようにする。
func parseFormVersion(value string) (*time.Time, bool) {
	switch value {
	case "":
		return nil, false
	case FormNullVersion:
		return nil, true
	}

	parsed, err := time.Parse(FormVersionLayout, value)
	if err != nil {
		return nil, false
	}

	return &parsed, true
}
