package model

// Staff はスタッフ情報を表します
type Staff struct {
	ID          int64
	WorkID      int64
	Name        string
	NameEn      string
	Role        string
	RoleOther   string
	RoleOtherEn string
}
