package model

// Prefectureは人物の出身地の選択肢 (都道府県と「海外」) を表します
type Prefecture struct {
	ID   PrefectureID
	Name string
}
