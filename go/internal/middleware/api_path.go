package middleware

import "strings"

// APIPathPrefixはREST API V2のパスの接頭辞。APIのルーターをマウントするパスでもある
const APIPathPrefix = "/api/v2"

// IsAPIPathはpathがREST API V2のパスかを返す。
// メンテナンス中やパニック時の応答をProblem Detailsにするかどうかと、
// APIに共通の応答ヘッダーを付けるかどうかの判定に使う
func IsAPIPath(path string) bool {
	return path == APIPathPrefix || strings.HasPrefix(path, APIPathPrefix+"/")
}
