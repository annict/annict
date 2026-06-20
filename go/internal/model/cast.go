// Package model はドメインモデルを提供します
package model

// Cast はキャスト情報を表します
type Cast struct {
	ID              CastID
	WorkID          WorkID
	Name            string
	NameEn          string
	CharacterName   string
	CharacterNameEn string
	PersonName      string
	PersonNameEn    string
}
