package auth

import (
	"crypto/rand"
	"math/big"
	"strings"

	"golang.org/x/text/width"
)

const (
	// RecoveryCodeCountは二要素認証を有効にしたときに発行するリカバリーコードの数。
	RecoveryCodeCount = 10

	// recoveryCodeGroupLengthはリカバリーコードをハイフンで区切る1組の文字数。
	// 4文字 + 4文字の `xxxx-xxxx` の形にして、書き写すときに読み上げやすくする。
	recoveryCodeGroupLength = 4

	// recoveryCodeAlphabetはリカバリーコードに使う文字。
	// 紙に書き写す・読み上げることを考え、見間違えやすい `0` / `o`、`1` / `l` / `i` を除いた小文字と数字にする。
	// 31文字から8文字を選ぶため、1つのコードはおよそ40ビットのエントロピーを持つ。
	recoveryCodeAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"
)

// GenerateRecoveryCodesは互いに異なるRecoveryCodeCount個のリカバリーコードを `xxxx-xxxx` の形で返す。
// 保存するダイジェストは、NormalizeRecoveryCodeを通した値から作る。
//
// 重複を引き直すのは、ダイジェストの (user_id, code_digest) の一意制約で保存に失敗しないようにするため。
func GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, 0, RecoveryCodeCount)
	seen := make(map[string]bool, RecoveryCodeCount)
	for len(codes) < RecoveryCodeCount {
		first, err := randomRecoveryCodeGroup()
		if err != nil {
			return nil, err
		}
		second, err := randomRecoveryCodeGroup()
		if err != nil {
			return nil, err
		}
		code := first + "-" + second
		if seen[code] {
			continue
		}
		seen[code] = true
		codes = append(codes, code)
	}

	return codes, nil
}

// NormalizeRecoveryCodeは入力されたリカバリーコードを、ダイジェストを作る形 (区切りの無い小文字) にする。
// 大文字小文字・ハイフン・空白の有無を問わず、全角で打った文字も受け付ける。
func NormalizeRecoveryCode(input string) string {
	return strings.ToLower(stripSeparators(width.Narrow.String(input)))
}

// IsRecoveryCodeFormatは、NormalizeRecoveryCodeを通したコードが、発行するコードの形 (区切りを除いた8文字) かを返す。
// 発行に使う文字と長さから判定し、発行するコードを入力の検証で拒むことが無いようにする。
func IsRecoveryCodeFormat(normalizedCode string) bool {
	if len(normalizedCode) != recoveryCodeGroupLength*2 {
		return false
	}
	for _, r := range normalizedCode {
		if !strings.ContainsRune(recoveryCodeAlphabet, r) {
			return false
		}
	}

	return true
}

// randomRecoveryCodeGroupはrecoveryCodeAlphabetから一様に選んだrecoveryCodeGroupLength文字を返す。
func randomRecoveryCodeGroup() (string, error) {
	b := make([]byte, recoveryCodeGroupLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(recoveryCodeAlphabet))))
		if err != nil {
			return "", err
		}
		b[i] = recoveryCodeAlphabet[n.Int64()]
	}

	return string(b), nil
}
