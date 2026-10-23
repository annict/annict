package viewmodel

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// TestNewDBPersonListItemsは人物を一覧の行へ詰め替え、状態をそのまま射影することを検証する。
func TestNewDBPersonListItems(t *testing.T) {
	t.Parallel()

	items := NewDBPersonListItems([]*model.Person{
		{ID: 2, Name: "人物2", NameKana: "じんぶつに", NameEn: "Person 2", Status: model.PersonStatusPublished},
		{ID: 1, Name: "人物1", Status: model.PersonStatusArchived},
	})

	want := []DBPersonListItem{
		{ID: 2, Name: "人物2", NameKana: "じんぶつに", NameEn: "Person 2", Status: DBPersonStatusPublished},
		{ID: 1, Name: "人物1", Status: DBPersonStatusArchived},
	}
	if len(items) != len(want) {
		t.Fatalf("len(items) = %d、期待値 = %d", len(items), len(want))
	}
	for i := range want {
		if items[i] != want[i] {
			t.Errorf("items[%d] = %+v、期待値 = %+v", i, items[i], want[i])
		}
	}
}

// TestDBPersonStatus_MirrorsDomainConstantsは一覧に表示する人物の状態のドメイン定数が、
// 同名のDBPersonStatus定数へ射影されることを検証する。NewDBPersonListItemは状態を直接変換する
// ため、値がずれたドメイン定数は一覧のテンプレートが描画しない値に黙って着地してしまう。
func TestDBPersonStatus_MirrorsDomainConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  DBPersonStatus
		want DBPersonStatus
	}{
		{name: "公開中", got: DBPersonStatus(model.PersonStatusPublished), want: DBPersonStatusPublished},
		{name: "アーカイブ", got: DBPersonStatus(model.PersonStatusArchived), want: DBPersonStatusArchived},
		{name: "掲載停止", got: DBPersonStatus(model.PersonStatusDelisted), want: DBPersonStatusDelisted},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: DBPersonStatus = %q、期待値 = %q", tt.name, tt.got, tt.want)
		}
	}
}

// TestNewDBPersonDelistingTargetは人物を確認画面の対象へ詰め替え、未記録の対応記録を空文字列に
// することを検証する。
func TestNewDBPersonDelistingTarget(t *testing.T) {
	t.Parallel()

	message := "本人から申出"
	updatedAt := time.Date(2026, 10, 18, 12, 34, 56, 123456789, time.UTC)
	tests := []struct {
		name   string
		person *model.Person
		want   DBPersonDelistingTarget
	}{
		{
			name:   "対応記録あり",
			person: &model.Person{ID: 1, Name: "人物1", DelistMessage: &message, UpdatedAt: updatedAt},
			want:   DBPersonDelistingTarget{ID: 1, Name: "人物1", DelistMessage: "本人から申出", UpdatedAt: "2026-10-18T12:34:56.123456789Z"},
		},
		{
			name:   "対応記録なし",
			person: &model.Person{ID: 2, Name: "人物2", UpdatedAt: updatedAt},
			want:   DBPersonDelistingTarget{ID: 2, Name: "人物2", UpdatedAt: "2026-10-18T12:34:56.123456789Z"},
		},
	}

	for _, tt := range tests {
		if got := NewDBPersonDelistingTarget(tt.person); got != tt.want {
			t.Errorf("%s: NewDBPersonDelistingTarget() = %+v、期待値 = %+v", tt.name, got, tt.want)
		}
	}
}

// TestNewDBPersonFormInputFromPersonは保存済みの人物を編集フォームの値へ射影し、未設定の
// カラムを空欄にし、版をUTCの往復書式で書き出すことを検証する。
func TestNewDBPersonFormInputFromPerson(t *testing.T) {
	t.Parallel()

	nickname := "あだな"
	gender := model.PersonGenderMale
	bloodType := model.PersonBloodTypeB
	prefectureID := model.PrefectureID(13)
	birthday := time.Date(1990, 4, 5, 0, 0, 0, 0, time.UTC)
	height := int32(170)
	url := "https://example.com"

	got := NewDBPersonFormInputFromPerson(&model.Person{
		ID:                1,
		Name:              "人物",
		NameKana:          "じんぶつ",
		NameEn:            "Person",
		Nickname:          &nickname,
		NicknameEn:        "Nick",
		Gender:            &gender,
		BloodType:         &bloodType,
		PrefectureID:      &prefectureID,
		Birthday:          &birthday,
		Height:            &height,
		URL:               &url,
		TwitterUsernameEn: "x_en",
		UpdatedAt:         time.Date(2026, 10, 15, 10, 2, 3, 123456000, time.FixedZone("JST", 9*60*60)),
	})

	want := DBPersonFormInput{
		Name:              "人物",
		NameKana:          "じんぶつ",
		NameEn:            "Person",
		Nickname:          "あだな",
		NicknameEn:        "Nick",
		Gender:            "male",
		BloodType:         "b",
		PrefectureID:      "13",
		Birthday:          "1990-04-05",
		Height:            "170",
		URL:               "https://example.com",
		TwitterUsernameEn: "x_en",
		UpdatedAt:         "2026-10-15T01:02:03.123456Z",
	}
	if got != want {
		t.Errorf("NewDBPersonFormInputFromPerson() = %+v、期待値 = %+v", got, want)
	}
}

// TestNewDBPersonFormOptionsは性別・血液型・出身地の選択肢を、値とラベルの組で並べることを
// 検証する。
func TestNewDBPersonFormOptions(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	options := NewDBPersonFormOptions(ctx, []model.Prefecture{{ID: 1, Name: "北海道"}, {ID: 48, Name: "海外"}})

	wantGender := []SelectOption{{Value: "male", Label: "男性"}, {Value: "female", Label: "女性"}}
	wantBloodType := []SelectOption{
		{Value: "a", Label: "A型"},
		{Value: "b", Label: "B型"},
		{Value: "ab", Label: "AB型"},
		{Value: "o", Label: "O型"},
	}
	wantPrefecture := []SelectOption{{Value: "1", Label: "北海道"}, {Value: "48", Label: "海外"}}

	for _, tc := range []struct {
		name string
		got  []SelectOption
		want []SelectOption
	}{
		{name: "性別", got: options.GenderOptions, want: wantGender},
		{name: "血液型", got: options.BloodTypeOptions, want: wantBloodType},
		{name: "出身地", got: options.PrefectureOptions, want: wantPrefecture},
	} {
		if len(tc.got) != len(tc.want) {
			t.Errorf("%sの選択肢 = %+v、期待値 = %+v", tc.name, tc.got, tc.want)
			continue
		}
		for i := range tc.want {
			if tc.got[i] != tc.want[i] {
				t.Errorf("%sの選択肢[%d] = %+v、期待値 = %+v", tc.name, i, tc.got[i], tc.want[i])
			}
		}
	}
}
