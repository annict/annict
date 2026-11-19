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

// newDelistPersonUsecaseは共有テストDBに対してUseCaseを組み立てる。本UseCaseは編集履歴と同じトランザクションで
// 書くため自前のトランザクションを開く。前提の人物はtestutil.CreateCommittedPersonでコミットする。
func newDelistPersonUsecase(t *testing.T) (*DelistPersonUsecase, *sql.DB) {
	t.Helper()

	db := testutil.GetTestDB()
	queries := query.New(db)
	return NewDelistPersonUsecase(db, repository.NewPersonRepository(queries), newTestDBEditRecordRepos(queries)), db
}

// readDelistedStateは人物が掲載停止で橋渡しのunpublished_atが立っているかと、対応記録を返す。
func readDelistedState(t *testing.T, db *sql.DB, id model.PersonID) (bool, sql.NullString) {
	t.Helper()

	var delisted bool
	var message sql.NullString
	if err := db.QueryRow(`
		SELECT status = 'delisted' AND unpublished_at IS NOT NULL, delist_message FROM people WHERE id = $1
	`, int64(id)).Scan(&delisted, &message); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	return delisted, message
}

// TestDelistPersonUsecase_Executeは、掲載を停止できる状態の人物の掲載を停止し、前後の空白を除いた
// 対応記録を書き、空の対応記録はNULLとして書くことを検証する。
func TestDelistPersonUsecase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		status      model.PersonStatus
		message     string
		wantMessage sql.NullString
	}{
		{name: "公開中の人物", status: model.PersonStatusPublished, message: "  本人から申出  \n", wantMessage: sql.NullString{String: "本人から申出", Valid: true}},
		{name: "アーカイブの人物", status: model.PersonStatusArchived, message: "本人から申出", wantMessage: sql.NullString{String: "本人から申出", Valid: true}},
		{name: "対応記録が空白だけ", status: model.PersonStatusPublished, message: " \n "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, db := newDelistPersonUsecase(t)
			personID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder { return b.WithStatus(tt.status) })

			output, err := uc.Execute(context.Background(), DelistPersonInput{
				User:          committedEditorForTest(t, model.RoleAdmin),
				PersonID:      personID,
				UpdatedAt:     testutil.PersonVersion(t, db, int64(personID)),
				DelistMessage: tt.message,
			})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.PersonID != personID {
				t.Errorf("PersonID = %d、期待値 = %d", output.PersonID, personID)
			}

			delisted, message := readDelistedState(t, db, personID)
			if !delisted {
				t.Error("statusがdelistedでないか、unpublished_atが書かれていません")
			}
			if message != tt.wantMessage {
				t.Errorf("delist_message = %+v、期待値 = %+v", message, tt.wantMessage)
			}
		})
	}
}

// TestDelistPersonUsecase_Execute_NotFoundは、存在しない人物と掲載を停止できない状態の人物を、
// not foundのAppErrorとして報告することを検証する。
func TestDelistPersonUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(b *testutil.PersonBuilder) *testutil.PersonBuilder
	}{
		{name: "存在しない人物"},
		{name: "掲載停止の人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
			return b.WithStatus(model.PersonStatusDelisted)
		}},
		{name: "削除済みの人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
			return b.WithStatus(model.PersonStatusDeleted)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, db := newDelistPersonUsecase(t)
			personID := model.PersonID(999999999)
			if tt.prepare != nil {
				personID = testutil.CreateCommittedPerson(t, tt.prepare)
			}
			version := time.Now().UTC().Format(time.RFC3339Nano)
			if tt.prepare != nil {
				version = testutil.PersonVersion(t, db, int64(personID))
			}

			output, err := uc.Execute(context.Background(), DelistPersonInput{
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

// TestDelistPersonUsecase_Execute_Conflictは、確認後に版が進んだ人物への送信を競合のAppErrorとして
// 報告し、掲載を停止せず対応記録も書かないことを検証する。
func TestDelistPersonUsecase_Execute_Conflict(t *testing.T) {
	t.Parallel()

	uc, db := newDelistPersonUsecase(t)
	personID := testutil.CreateCommittedPerson(t, nil)
	version := testutil.PersonVersion(t, db, int64(personID))
	if _, err := db.Exec(`UPDATE people SET updated_at = updated_at + interval '1 second' WHERE id = $1`, int64(personID)); err != nil {
		t.Fatalf("人物の更新に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), DelistPersonInput{
		User:          committedEditorForTest(t, model.RoleAdmin),
		PersonID:      personID,
		UpdatedAt:     version,
		DelistMessage: "古い確認画面からの記録",
	})
	if output != nil {
		t.Errorf("Execute() = %+v、期待値 = nil", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeConflict", err)
	}

	delisted, message := readDelistedState(t, db, personID)
	if delisted || message.Valid {
		t.Errorf("競合した送信が書き込まれました (掲載停止 = %v、delist_message = %+v)", delisted, message)
	}
}

// TestDelistPersonUsecase_Execute_MissingVersionは、確認時の版を持たない送信をValidationErrorとして
// 報告し、人物を書き換えないことを検証する。
func TestDelistPersonUsecase_Execute_MissingVersion(t *testing.T) {
	t.Parallel()

	uc, db := newDelistPersonUsecase(t)
	personID := testutil.CreateCommittedPerson(t, nil)

	output, err := uc.Execute(context.Background(), DelistPersonInput{
		User:          committedEditorForTest(t, model.RoleAdmin),
		PersonID:      personID,
		DelistMessage: "記録",
	})
	if output != nil {
		t.Errorf("Execute() = %+v、期待値 = nil", output)
	}
	if model.AsValidationError(err) == nil {
		t.Errorf("Execute()のエラー = %v、期待値 = *model.ValidationError", err)
	}

	if delisted, _ := readDelistedState(t, db, personID); delisted {
		t.Error("版の無い送信が人物の掲載を停止しました")
	}
}

// TestDelistPersonUsecase_Execute_RequiresAdminは、管理者でない呼び出しを、人物を書き換える前に
// 拒否することを検証する。
func TestDelistPersonUsecase_Execute_RequiresAdmin(t *testing.T) {
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

			uc, db := newDelistPersonUsecase(t)
			personID := testutil.CreateCommittedPerson(t, nil)

			_, err := uc.Execute(context.Background(), DelistPersonInput{User: tt.user, PersonID: personID, DelistMessage: "記録"})
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
			}

			if delisted, _ := readDelistedState(t, db, personID); delisted {
				t.Error("拒否した呼び出しが人物の掲載を停止しました")
			}
		})
	}
}
