package validator

import "time"

// FormNullVersionは、保存済みのupdated_atがNULLのとき編集フォームが運ぶ明示的な版。
// Annict DB管理画面が編集する各テーブルでは共有DBが同カラムをNULL許容にしているため、
// 「タイムスタンプが無い」ことをフォームが表明できる必要がある。リクエストの値が空であることは、
// そもそも版が示されていないことを意味し、拒否する。
//
// フォームが値を描画し、本パッケージが読み戻すため、リテラルは、送信された版を受け入れるかを
// 判断する側であるここに置く。
const FormNullVersion = "null"

// FormVersionLayoutは版を往復させる書式。同一秒内の2つの書き込みを区別する秒未満の桁を
// 保ち、元の時刻へパースし直せるため、更新側は保存済みのカラムと照合できる。
const FormVersionLayout = time.RFC3339Nano

// parseFormVersionは送信が示す版を読み取り、フィールドが空か、往復書式のタイムスタンプでも
// nullのセンチネルでもない場合にfalseを返す。nilの時刻とtrueの組はセンチネルで、保存済みの
// updated_atがNULLであることを表す。
//
// Annict DBのどの編集フォームも同じ形で版を運ぶため、版を運ぶ2つのフォームはこの読み取りを
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
