package model

import "time"

// PersonStatusは人物のライフサイクル状態で、PostgreSQLのperson_status enumと対応する。
// archived (非公開) は編集者も設定でき、delisted (本人の申出による掲載停止) とdeletedは管理者に
// 限る。Rails版の人物ページをGo版に移すまでは、archivedの人物も作品の非公開と同じく
// unpublished_atを立て、Rails版の表示から除外する。
type PersonStatus string

const (
	PersonStatusPublished PersonStatus = "published"
	PersonStatusArchived  PersonStatus = "archived"
	PersonStatusDelisted  PersonStatus = "delisted"
	PersonStatusDeleted   PersonStatus = "deleted"
)

// Stringはステータスの文字列表現を返す。
func (s PersonStatus) String() string { return string(s) }

// IsViewableは人物がAnnict DB管理画面の一覧に誰にでも表示される状態 (公開中と非公開) かを返す。
// 編集・削除・掲載停止ができるのも、この状態の人物に限る。
func (s PersonStatus) IsViewable() bool {
	switch s {
	case PersonStatusPublished, PersonStatusArchived:
		return true
	default:
		return false
	}
}

// PersonGenderは人物の性別を表す。値はRails版のPerson.gender (enumerize) と同じ。
type PersonGender string

const (
	PersonGenderMale   PersonGender = "male"
	PersonGenderFemale PersonGender = "female"
)

// PersonGendersは編集フォームが選択肢として示す性別を、表示する順で並べる。
var PersonGenders = []PersonGender{PersonGenderMale, PersonGenderFemale}

// IsValidは性別がRails版の許可リストにある値かを返す。
func (g PersonGender) IsValid() bool {
	switch g {
	case PersonGenderMale, PersonGenderFemale:
		return true
	default:
		return false
	}
}

// PersonBloodTypeは人物の血液型を表す。値はRails版のPerson.blood_type (enumerize) と同じ。
type PersonBloodType string

const (
	PersonBloodTypeA  PersonBloodType = "a"
	PersonBloodTypeB  PersonBloodType = "b"
	PersonBloodTypeAB PersonBloodType = "ab"
	PersonBloodTypeO  PersonBloodType = "o"
)

// PersonBloodTypesは編集フォームが選択肢として示す血液型を、表示する順で並べる。
var PersonBloodTypes = []PersonBloodType{PersonBloodTypeA, PersonBloodTypeB, PersonBloodTypeAB, PersonBloodTypeO}

// IsValidは血液型がRails版の許可リストにある値かを返す。
func (b PersonBloodType) IsValid() bool {
	switch b {
	case PersonBloodTypeA, PersonBloodTypeB, PersonBloodTypeAB, PersonBloodTypeO:
		return true
	default:
		return false
	}
}

// Personは人物情報を表します
type Person struct {
	ID       PersonID
	Name     string
	NameKana string
	NameEn   string

	// 以下はAnnict DB管理画面の編集フォームが扱う項目。値が入るのは編集用のローダー
	// (GetForEditByID) のみ。NULL許容のカラムはポインタで持つ。
	Nickname          *string
	NicknameEn        string
	Gender            *PersonGender
	BloodType         *PersonBloodType
	PrefectureID      *PrefectureID
	Birthday          *time.Time
	Height            *int32
	URL               *string
	URLEn             string
	WikipediaURL      *string
	WikipediaURLEn    string
	TwitterUsername   *string
	TwitterUsernameEn string

	// Statusは人物の状態の正本。値が入るのはAnnict DB管理画面のローダー (ListForDB・
	// GetForEditByID・GetForStateChangeByID・GetForDelistingByID・ListNameHolders・
	// GetNameHolderExcludingID) のみ。
	Status PersonStatus

	// DelistMessageは掲載停止 (delisted) の対応記録で、管理者だけが参照する。値が入るのは掲載停止・
	// 掲載停止の解除の確認画面のローダー (GetForDelistingByID) のみ。
	DelistMessage *string

	// UpdatedAtは編集フォームと掲載停止・解除の確認画面がhiddenで持ち回る版。値が入るのは
	// 編集用と掲載停止用のローダー (GetForEditByID・GetForDelistingByID) のみ。
	UpdatedAt time.Time
}
