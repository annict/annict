package viewmodel

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// DBPersonStatusはAnnict DB管理画面の人物一覧に表示する人物の状態。一覧は公開中と非公開の
// 人物と、管理者には掲載停止の人物を表示するため、その3値を持つ。PublishingStatusに無い
// 掲載停止を含むため、PublishingStatusとは別の型にする。射影が保たれていることは、ドメイン定数を変換して
// 比較するテストで担保する。
type DBPersonStatus string

const (
	DBPersonStatusPublished DBPersonStatus = "published"
	DBPersonStatusArchived  DBPersonStatus = "archived"
	DBPersonStatusDelisted  DBPersonStatus = "delisted"
)

// DBPersonListItemはAnnict DB管理画面の人物一覧で1行ごとに表示する整形済みデータ。
type DBPersonListItem struct {
	ID   PersonID
	Name string
	// 名前の下に並べる読みと英語名。未設定なら空文字列で、テンプレート側で "-" の
	// プレースホルダーを表示する。
	NameKana string
	NameEn   string
	Status   DBPersonStatus
}

// NewDBPersonListItemsは人物の一覧を一覧画面の行に変換する。
func NewDBPersonListItems(people []*model.Person) []DBPersonListItem {
	result := make([]DBPersonListItem, len(people))
	for i, person := range people {
		result[i] = NewDBPersonListItem(person)
	}
	return result
}

// NewDBPersonListItemは人物を一覧画面の1行に変換する。
func NewDBPersonListItem(person *model.Person) DBPersonListItem {
	return DBPersonListItem{
		ID:       PersonID(person.ID),
		Name:     person.Name,
		NameKana: person.NameKana,
		NameEn:   person.NameEn,
		Status:   DBPersonStatus(person.Status),
	}
}

// DBPersonEditTargetは人物編集ページが見出しとリンクで名指しする、編集対象の人物。
type DBPersonEditTarget struct {
	ID   PersonID
	Name string
}

// NewDBPersonEditTargetは保存済みの人物を、編集ページが名指しする対象に変換する。見出しは
// 送信された名前ではなく保存済みの名前を示す。却下された送信で名前を空にしても、どの人物の
// ページかが変わらないようにするため。
func NewDBPersonEditTarget(person *model.Person) DBPersonEditTarget {
	return DBPersonEditTarget{
		ID:   PersonID(person.ID),
		Name: person.Name,
	}
}

// DBPersonDelistingTargetは人物の掲載停止・掲載停止の解除の確認画面が名指しする人物と、
// その対応記録。対応記録は管理者だけが参照する内部記録で、未記録なら空文字列になる。
type DBPersonDelistingTarget struct {
	ID            PersonID
	Name          string
	DelistMessage string
	UpdatedAt     string
}

// NewDBPersonDelistingTargetは人物を、掲載停止・掲載停止の解除の確認画面が名指しする対象に変換する。
func NewDBPersonDelistingTarget(person *model.Person) DBPersonDelistingTarget {
	target := DBPersonDelistingTarget{
		ID:        PersonID(person.ID),
		Name:      person.Name,
		UpdatedAt: formatFormVersion(&person.UpdatedAt),
	}
	if person.DelistMessage != nil {
		target.DelistMessage = *person.DelistMessage
	}
	return target
}

// DBPersonFormOptionsは人物編集フォームの選択欄が示す選択肢。
type DBPersonFormOptions struct {
	GenderOptions     []SelectOption
	BloodTypeOptions  []SelectOption
	PrefectureOptions []SelectOption
}

// NewDBPersonFormOptionsは人物編集フォームの選択肢を組み立てる。性別と血液型は許可リスト
// の順、出身地は登録済みの出身地の順に並べる。
func NewDBPersonFormOptions(ctx context.Context, prefectures []model.Prefecture) DBPersonFormOptions {
	genderOptions := make([]SelectOption, len(model.PersonGenders))
	for i, gender := range model.PersonGenders {
		genderOptions[i] = SelectOption{Value: string(gender), Label: personGenderLabel(ctx, gender)}
	}

	bloodTypeOptions := make([]SelectOption, len(model.PersonBloodTypes))
	for i, bloodType := range model.PersonBloodTypes {
		bloodTypeOptions[i] = SelectOption{Value: string(bloodType), Label: personBloodTypeLabel(ctx, bloodType)}
	}

	prefectureOptions := make([]SelectOption, len(prefectures))
	for i, prefecture := range prefectures {
		prefectureOptions[i] = SelectOption{Value: prefecture.ID.String(), Label: prefecture.Name}
	}

	return DBPersonFormOptions{
		GenderOptions:     genderOptions,
		BloodTypeOptions:  bloodTypeOptions,
		PrefectureOptions: prefectureOptions,
	}
}

// personGenderLabelは性別の表示名を返す。翻訳キーを使用箇所からgrepで追える状態に保つため、
// 値から組み立てずに直書きのキーで引く。
func personGenderLabel(ctx context.Context, gender model.PersonGender) string {
	switch gender {
	case model.PersonGenderMale:
		return i18n.T(ctx, "person_gender_male")
	case model.PersonGenderFemale:
		return i18n.T(ctx, "person_gender_female")
	default:
		return string(gender)
	}
}

// personBloodTypeLabelは血液型の表示名を返す。
func personBloodTypeLabel(ctx context.Context, bloodType model.PersonBloodType) string {
	switch bloodType {
	case model.PersonBloodTypeA:
		return i18n.T(ctx, "person_blood_type_a")
	case model.PersonBloodTypeB:
		return i18n.T(ctx, "person_blood_type_b")
	case model.PersonBloodTypeAB:
		return i18n.T(ctx, "person_blood_type_ab")
	case model.PersonBloodTypeO:
		return i18n.T(ctx, "person_blood_type_o")
	default:
		return string(bloodType)
	}
}

// DBPersonFormInputは人物編集フォームが各欄に描画する値を保持する。値を文字列で持つのは、
// 却下された送信を入力されたまま再描画するため (整数や日付として解釈できない入力も残す)。
type DBPersonFormInput struct {
	Name              string
	NameKana          string
	NameEn            string
	Nickname          string
	NicknameEn        string
	Gender            string
	BloodType         string
	PrefectureID      string
	Birthday          string
	Height            string
	URL               string
	URLEn             string
	WikipediaURL      string
	WikipediaURLEn    string
	TwitterUsername   string
	TwitterUsernameEn string
	// UpdatedAtは編集フォームを開いた時点の版で、hiddenで持ち回る。古い読み取りに対する
	// 送信を、間に書いた人の変更を黙って上書きせずに更新側で却下できるようにするため。
	UpdatedAt string
}

// NewDBPersonFormInputFromPersonは保存済みの人物を、編集フォームが描画するフォーム値に
// 射影する。未設定のカラムは "" になる。
func NewDBPersonFormInputFromPerson(person *model.Person) DBPersonFormInput {
	input := DBPersonFormInput{
		Name:              person.Name,
		NameKana:          person.NameKana,
		NameEn:            person.NameEn,
		Nickname:          derefString(person.Nickname),
		NicknameEn:        person.NicknameEn,
		Birthday:          formatDateInput(person.Birthday),
		Height:            formatNullableInt32(person.Height),
		URL:               derefString(person.URL),
		URLEn:             person.URLEn,
		WikipediaURL:      derefString(person.WikipediaURL),
		WikipediaURLEn:    person.WikipediaURLEn,
		TwitterUsername:   derefString(person.TwitterUsername),
		TwitterUsernameEn: person.TwitterUsernameEn,
		UpdatedAt:         formatFormVersion(&person.UpdatedAt),
	}
	if person.Gender != nil {
		input.Gender = string(*person.Gender)
	}
	if person.BloodType != nil {
		input.BloodType = string(*person.BloodType)
	}
	if person.PrefectureID != nil {
		input.PrefectureID = person.PrefectureID.String()
	}

	return input
}

// NewDBPersonFormInputFromSubmitは却下された送信の値を、入力されたまま再描画するフォームに
// 残す。版もそのまま返す (NewDBEpisodeFormInputFromSubmitと同じ)。
func NewDBPersonFormInputFromSubmit(input usecase.UpdatePersonInput) DBPersonFormInput {
	return DBPersonFormInput{
		Name:              input.Name,
		NameKana:          input.NameKana,
		NameEn:            input.NameEn,
		Nickname:          input.Nickname,
		NicknameEn:        input.NicknameEn,
		Gender:            input.Gender,
		BloodType:         input.BloodType,
		PrefectureID:      input.PrefectureID,
		Birthday:          input.Birthday,
		Height:            input.Height,
		URL:               input.URL,
		URLEn:             input.URLEn,
		WikipediaURL:      input.WikipediaURL,
		WikipediaURLEn:    input.WikipediaURLEn,
		TwitterUsername:   input.TwitterUsername,
		TwitterUsernameEn: input.TwitterUsernameEn,
		UpdatedAt:         input.UpdatedAt,
	}
}

// Valはフィールド名 (フォームのname属性) に対応する値を返す。エラー要約と競合の案内が、
// フィールドの並びから値を引くために使う。
func (d DBPersonFormInput) Val(field string) string {
	switch field {
	case "name":
		return d.Name
	case "name_kana":
		return d.NameKana
	case "name_en":
		return d.NameEn
	case "nickname":
		return d.Nickname
	case "nickname_en":
		return d.NicknameEn
	case "gender":
		return d.Gender
	case "blood_type":
		return d.BloodType
	case "prefecture_id":
		return d.PrefectureID
	case "birthday":
		return d.Birthday
	case "height":
		return d.Height
	case "url":
		return d.URL
	case "url_en":
		return d.URLEn
	case "wikipedia_url":
		return d.WikipediaURL
	case "wikipedia_url_en":
		return d.WikipediaURLEn
	case "twitter_username":
		return d.TwitterUsername
	case "twitter_username_en":
		return d.TwitterUsernameEn
	default:
		return ""
	}
}
