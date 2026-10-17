// Package modelはドメインモデルを提供します
package model

// Castはキャスト情報を表します
type Cast struct {
	ID     CastID
	WorkID WorkID
	Name   string
	NameEn string
	// Personはキャストを演じる人物です
	Person *Person
	// Characterはキャストが演じるキャラクターです (キャラクターのレコードが見つからない場合はnil)
	Character *Character
}
