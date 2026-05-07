package model

// Staff はスタッフ情報を表します
type Staff struct {
	ID          StaffID
	WorkID      WorkID
	Name        string
	NameEn      string
	Role        string
	RoleOther   string
	RoleOtherEn string
}
