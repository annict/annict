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
)

// newUndelistPersonUsecaseは共有テストDBに対してUseCaseを組み立てる。本UseCaseは編集履歴と同じトランザクションで
// 書くため自前のトランザクションを開く。前提の人物はtestutil.CreateCommittedPersonでコミットする。
func newUndelistPersonUsecase(t *testing.T) (*UndelistPersonUsecase, *sql.DB) {
	t.Helper()

	db := testutil.GetTestDB()
	queries := query.New(db)
	return NewUndelistPersonUsecase(db, repository.NewPersonRepository(queries), newTestDBEditRecordRepos(queries)), db
}

// newDelistedPersonは橋渡しのunpublished_atと対応記録を持つ、掲載停止の人物を作る。
func newDelistedPerson(t *testing.T, db *sql.DB) model.PersonID {
	t.Helper()

	return testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
		return b.
			WithStatus(model.PersonStatusDelisted).
			WithUnpublishedAt(time.Now()).
			WithDelistMessage("対応記録")
	})
}

// TestUndelistPersonUsecase_Executeは、掲載停止の人物を公開中に戻し、橋渡しのunpublished_atを
// NULLに戻して、対応記録を残すことを検証する。
func TestUndelistPersonUsecase_Execute(t *testing.T) {
	t.Parallel()

	uc, db := newUndelistPersonUsecase(t)
	personID := newDelistedPerson(t, db)

	output, err := uc.Execute(context.Background(), UndelistPersonInput{
		User:      committedEditorForTest(t, model.RoleAdmin),
		PersonID:  personID,
		UpdatedAt: testutil.PersonVersion(t, db, int64(personID)),
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.PersonID != personID {
		t.Errorf("PersonID = %d、期待値 = %d", output.PersonID, personID)
	}

	var published bool
	var message sql.NullString
	if err := db.QueryRow(`
		SELECT status = 'published' AND unpublished_at IS NULL, delist_message FROM people WHERE id = $1
	`, int64(personID)).Scan(&published, &message); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	if !published {
		t.Error("statusがpublishedでないか、unpublished_atが残っています")
	}
	if message.String != "対応記録" {
		t.Errorf("delist_message = %+v、期待値 = 対応記録", message)
	}
}

// TestUndelistPersonUsecase_Execute_NotFoundは、存在しない人物と掲載停止でない人物を、not foundの
// AppErrorとして報告することを検証する。
func TestUndelistPersonUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(b *testutil.PersonBuilder) *testutil.PersonBuilder
	}{
		{name: "存在しない人物"},
		{name: "公開中の人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder { return b }},
		{name: "アーカイブの人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
			return b.WithStatus(model.PersonStatusArchived)
		}},
		{name: "削除済みの人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
			return b.WithStatus(model.PersonStatusDeleted)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, db := newUndelistPersonUsecase(t)
			personID := model.PersonID(999999999)
			if tt.prepare != nil {
				personID = testutil.CreateCommittedPerson(t, tt.prepare)
			}
			version := time.Now().UTC().Format(time.RFC3339Nano)
			if tt.prepare != nil {
				version = testutil.PersonVersion(t, db, int64(personID))
			}

			output, err := uc.Execute(context.Background(), UndelistPersonInput{
				User:      committedEditorForTest(t, model.RoleAdmin),
				PersonID:  personID,
				UpdatedAt: version,
			})
			if output != nil {
				t.Errorf("Execute() = %+v、期待値 = nil", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", err)
			}
		})
	}
}

// readUndelistedは人物が公開中に戻り、橋渡しのunpublished_atがNULLになっているかを返す。
func readUndelisted(t *testing.T, db *sql.DB, id model.PersonID) bool {
	t.Helper()

	var published bool
	if err := db.QueryRow(`
		SELECT status = 'published' AND unpublished_at IS NULL FROM people WHERE id = $1
	`, int64(id)).Scan(&published); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	return published
}

// TestUndelistPersonUsecase_Execute_Conflictは、確認後に版が進んだ人物への解除を競合のAppErrorとして
// 報告し、掲載停止のまま残すことを検証する。
func TestUndelistPersonUsecase_Execute_Conflict(t *testing.T) {
	t.Parallel()

	uc, db := newUndelistPersonUsecase(t)
	personID := newDelistedPerson(t, db)
	version := testutil.PersonVersion(t, db, int64(personID))
	if _, err := db.Exec(`UPDATE people SET updated_at = updated_at + interval '1 second' WHERE id = $1`, int64(personID)); err != nil {
		t.Fatalf("人物の更新に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), UndelistPersonInput{
		User:      committedEditorForTest(t, model.RoleAdmin),
		PersonID:  personID,
		UpdatedAt: version,
	})
	if output != nil {
		t.Errorf("Execute() = %+v、期待値 = nil", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeConflict", err)
	}
	if readUndelisted(t, db, personID) {
		t.Error("競合した解除が人物を公開中に戻しました")
	}
}

// TestUndelistPersonUsecase_Execute_MissingVersionは、確認時の版を持たない解除をValidationErrorとして
// 報告し、掲載停止のまま残すことを検証する。
func TestUndelistPersonUsecase_Execute_MissingVersion(t *testing.T) {
	t.Parallel()

	uc, db := newUndelistPersonUsecase(t)
	personID := newDelistedPerson(t, db)

	output, err := uc.Execute(context.Background(), UndelistPersonInput{
		User:     committedEditorForTest(t, model.RoleAdmin),
		PersonID: personID,
	})
	if output != nil {
		t.Errorf("Execute() = %+v、期待値 = nil", output)
	}
	if model.AsValidationError(err) == nil {
		t.Errorf("Execute()のエラー = %v、期待値 = *model.ValidationError", err)
	}
	if readUndelisted(t, db, personID) {
		t.Error("版の無い解除が人物を公開中に戻しました")
	}
}

// TestUndelistPersonUsecase_Execute_RequiresAdminは、管理者でない呼び出しを、人物を書き換える前に
// 拒否することを検証する。
func TestUndelistPersonUsecase_Execute_RequiresAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "ユーザー無し", user: nil},
		{name: "一般ユーザー", user: committedEditorForTest(t, model.RoleUser)},
		{name: "編集者", user: committedEditorForTest(t, model.RoleEditor)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, db := newUndelistPersonUsecase(t)
			personID := newDelistedPerson(t, db)

			_, err := uc.Execute(context.Background(), UndelistPersonInput{User: tt.user, PersonID: personID})
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
			}

			var unchanged bool
			if err := db.QueryRow(`SELECT status = 'delisted' AND unpublished_at IS NOT NULL FROM people WHERE id = $1`, int64(personID)).Scan(&unchanged); err != nil {
				t.Fatalf("人物の取得に失敗: %v", err)
			}
			if !unchanged {
				t.Error("拒否した呼び出しが人物の掲載停止を解除しました")
			}
		})
	}
}
