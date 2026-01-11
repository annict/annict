package stripe

import (
	"database/sql"
	"time"
)

// NullTimeFromUnix はUnixタイムスタンプからsql.NullTimeを作成します
// 値が0の場合はValidがfalseになります
func NullTimeFromUnix(ts int64) sql.NullTime {
	if ts == 0 {
		return sql.NullTime{}
	}
	return sql.NullTime{
		Time:  time.Unix(ts, 0),
		Valid: true,
	}
}
