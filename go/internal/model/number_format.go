package model

// NumberFormat はエピソード番号のフォーマット定義です
type NumberFormat struct {
	ID         NumberFormatID
	Name       string
	SortNumber int32
}
