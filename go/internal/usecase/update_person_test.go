package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// updatePersonTestVersionはテストの人物に設定する過去のupdated_atで、送信が運ぶ版と
// 往復書式で一致する。同じトランザクションの中ではNOW() が進まないため、過去の時刻を使う。
var updatePersonTestVersion = time.Date(2024, 5, 6, 7, 8, 9, 654321000, time.UTC)

// newUpdatePersonUsecaseは共有テストDBに対してUseCaseを組み立てる。本UseCaseは編集履歴と同じトランザクションで
// 書くため自前のトランザクションを開く。前提の人物はtestutil.CreateCommittedPersonでコミットする。
func newUpdatePersonUsecase(t *testing.T) (*UpdatePersonUsecase, *sql.DB) {
	t.Helper()

	db := testutil.GetTestDB()
	queries := query.New(db)
	personRepo := repository.NewPersonRepository(queries)
	prefectureRepo := repository.NewPrefectureRepository(queries)

	return NewUpdatePersonUsecase(db, personRepo, validator.NewDBPersonUpdateValidator(personRepo, prefectureRepo), newTestDBEditRecordRepos(queries)), db
}

// TestUpdatePersonUsecase_Execute_UpdatesPersonWithActivityは、送信された値を人物に適用し、
// 編集者に帰属するpeople.updateの編集履歴を記録することを検証する。
func TestUpdatePersonUsecase_Execute_UpdatesPersonWithActivity(t *testing.T) {
	t.Parallel()

	uc, db := newUpdatePersonUsecase(t)
	userID := committedEditorForTest(t, model.RoleEditor).ID
	personID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
		return b.WithUpdatedAt(updatePersonTestVersion)
	})
	newName := "編集後の人物-" + t.Name()

	output, err := uc.Execute(context.Background(), UpdatePersonInput{
		PersonID:  personID,
		User:      &model.User{ID: userID, Role: model.RoleEditor},
		Name:      newName,
		Gender:    "female",
		Height:    "160",
		UpdatedAt: updatePersonTestVersion.Format(validator.FormVersionLayout),
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.PersonID != personID {
		t.Errorf("PersonID = %d、期待値 = %d", output.PersonID, personID)
	}

	var (
		name, gender string
		height       int32
		activities   int
	)
	if err := db.QueryRow(`
		SELECT p.name, p.gender, p.height,
			(SELECT COUNT(*) FROM db_activities a
			 WHERE a.trackable_type = 'Person' AND a.trackable_id = p.id AND a.action = 'people.update' AND a.user_id = $2)
		FROM people p WHERE p.id = $1
	`, int64(personID), int64(userID)).Scan(&name, &gender, &height, &activities); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	if name != newName || gender != "female" || height != 160 {
		t.Errorf("人物 = {name: %q, gender: %q, height: %d}", name, gender, height)
	}
	if activities != 1 {
		t.Errorf("編集履歴 = %d件、期待値 = 1件", activities)
	}
}

// TestUpdatePersonUsecase_Execute_Conflictは、フォームを開いた後に他者が書いた人物への
// 送信を、上書きせずに競合として報告することを検証する。
func TestUpdatePersonUsecase_Execute_Conflict(t *testing.T) {
	t.Parallel()

	uc, db := newUpdatePersonUsecase(t)
	userID := committedEditorForTest(t, model.RoleEditor).ID
	storedName := "保存済みの人物-" + t.Name()
	personID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
		return b.WithName(storedName).WithUpdatedAt(updatePersonTestVersion)
	})

	staleVersion := updatePersonTestVersion.Add(-time.Second).Format(validator.FormVersionLayout)
	output, err := uc.Execute(context.Background(), UpdatePersonInput{
		PersonID:  personID,
		User:      &model.User{ID: userID, Role: model.RoleEditor},
		Name:      "古い版からの人物-" + t.Name(),
		UpdatedAt: staleVersion,
	})
	if output != nil {
		t.Errorf("Execute() = %+v、期待値 = nil", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeConflict", err)
	}

	var name string
	if err := db.QueryRow(`SELECT name FROM people WHERE id = $1`, int64(personID)).Scan(&name); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	if name != storedName {
		t.Errorf("name = %q、期待値 = %q (上書きしない)", name, storedName)
	}
}

// TestUpdatePersonUsecase_Execute_ConflictBeforeDuplicateは、古い版と名前の重複が
// 重なったとき、先に保存済みの版を示す競合を返すことを検証する。
func TestUpdatePersonUsecase_Execute_ConflictBeforeDuplicate(t *testing.T) {
	t.Parallel()

	uc, _ := newUpdatePersonUsecase(t)
	personID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
		return b.WithUpdatedAt(updatePersonTestVersion)
	})
	takenName := "別の人物-" + t.Name()
	testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder { return b.WithName(takenName) })

	_, err := uc.Execute(context.Background(), UpdatePersonInput{
		PersonID:  personID,
		User:      committedEditorForTest(t, model.RoleEditor),
		Name:      takenName,
		UpdatedAt: updatePersonTestVersion.Add(-time.Second).Format(validator.FormVersionLayout),
	})
	if ae := model.AsAppError(err); ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeConflict", err)
	}
}

// TestUpdatePersonUsecase_Execute_NotFoundは、存在しない人物と掲載停止・削除済みの人物への送信を
// not foundとして報告することを検証する。
func TestUpdatePersonUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	uc, _ := newUpdatePersonUsecase(t)
	delistedID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
		return b.
			WithStatus(model.PersonStatusDelisted).
			WithUpdatedAt(updatePersonTestVersion)
	})
	deletedID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
		return b.
			WithStatus(model.PersonStatusDeleted).
			WithUpdatedAt(updatePersonTestVersion)
	})

	tests := []struct {
		name     string
		personID model.PersonID
	}{
		{name: "存在しない人物", personID: model.PersonID(999999999)},
		{name: "掲載停止の人物", personID: delistedID},
		{name: "削除済みの人物", personID: deletedID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), UpdatePersonInput{
				PersonID:  tt.personID,
				User:      committedEditorForTest(t, model.RoleEditor),
				Name:      "見つからない人物-" + t.Name(),
				UpdatedAt: updatePersonTestVersion.Format(validator.FormVersionLayout),
			})
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", err)
			}
		})
	}
}

// TestUpdatePersonUsecase_Execute_ValidationErrorは、不正な送信を書き込まずに
// バリデーションエラーとして返すことを検証する。
func TestUpdatePersonUsecase_Execute_ValidationError(t *testing.T) {
	t.Parallel()

	uc, _ := newUpdatePersonUsecase(t)
	storedName := "検証される人物-" + t.Name()
	personID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
		return b.WithName(storedName).WithUpdatedAt(updatePersonTestVersion)
	})

	_, err := uc.Execute(context.Background(), UpdatePersonInput{
		PersonID:  personID,
		User:      committedEditorForTest(t, model.RoleEditor),
		Name:      storedName,
		Birthday:  "1990-13-01",
		UpdatedAt: updatePersonTestVersion.Format(validator.FormVersionLayout),
	})
	ve := model.AsValidationError(err)
	if ve == nil || !ve.HasFieldError("birthday") {
		t.Fatalf("Execute()のエラー = %v、期待値 = birthdayのバリデーションエラー", err)
	}
}

// TestUpdatePersonUsecase_Execute_RequiresCommitterは、ユーザーを伴わない呼び出しと
// 一般ユーザーの呼び出しを、書き込む前に拒否することを検証する。
func TestUpdatePersonUsecase_Execute_RequiresCommitter(t *testing.T) {
	t.Parallel()

	uc, db := newUpdatePersonUsecase(t)
	storedName := "認可される人物-" + t.Name()
	personID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
		return b.WithName(storedName).WithUpdatedAt(updatePersonTestVersion)
	})

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "ユーザー無し", user: nil},
		{name: "一般ユーザー", user: committedEditorForTest(t, model.RoleUser)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), UpdatePersonInput{
				PersonID:  personID,
				User:      tt.user,
				Name:      "書き換えられた人物-" + t.Name(),
				UpdatedAt: updatePersonTestVersion.Format(validator.FormVersionLayout),
			})
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}

	var name string
	if err := db.QueryRow(`SELECT name FROM people WHERE id = $1`, int64(personID)).Scan(&name); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	if name != storedName {
		t.Errorf("name = %q、期待値 = %q (書き込まない)", name, storedName)
	}
}
