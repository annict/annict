package model

// Programはチャンネルで作品を放送する番組を表す。
type Program struct {
	ID ProgramID
	// MinimumEpisodeGeneratableNumberは、自動生成が話数の計算で引く不定期放送の数に
	// 数える回数の下限。話数のずれ (途中の話数から放送する番組など) には使わない。
	MinimumEpisodeGeneratableNumber int32
}
