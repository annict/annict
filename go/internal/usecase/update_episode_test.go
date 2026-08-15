package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// newUpdateEpisodeUsecase wires the update usecase against the shared test DB. The usecase
// opens its own transaction, so its tests use GetTestDB (not SetupTx) so the committed rows are
// visible to the usecase's inner transaction and to the follow-up sync invariant check.
//
// [Ja] newUpdateEpisodeUsecase は共有テスト DB に対して更新 UseCase を組み立てる。本 UseCase は
// 内部で自前のトランザクションを開くため、テストは SetupTx ではなく GetTestDB を使い、コミット
// 済みの行が UseCase の内側トランザクションと後続の同期不変条件チェックから見えるようにする。
func newUpdateEpisodeUsecase(db *sql.DB) *UpdateEpisodeUsecase {
	queries := query.New(db)
	return NewUpdateEpisodeUsecase(
		db,
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
		repository.NewAnimeClassificationRepository(queries),
		validator.NewDBEpisodeUpdateValidator(),
	)
}

// insertUpdateTargetEpisode inserts the episode an update test edits under the given work,
// already mapped to its own anime when one is passed. Its timestamps come from the database so
// the version the form reads back is the one the update compares against.
//
// [Ja] insertUpdateTargetEpisode は更新テストが編集するエピソードを指定作品の配下に挿入する。
// anime を渡した場合は自身の anime にマッピング済みにする。タイムスタンプは DB から取るため、
// フォームが読み戻す版と更新が照合する版が一致する。
func insertUpdateTargetEpisode(t *testing.T, db *sql.DB, workID model.WorkID, animeID sql.NullInt64, sortNumber int32) model.EpisodeID {
	t.Helper()

	var id int64
	if err := db.QueryRow(`
		INSERT INTO episodes (
			work_id, number, raw_number, sort_number, title, title_ro, title_en,
			anime_id, created_at, updated_at
		) VALUES ($1, '#1', 1, $2, '編集前のタイトル', 'Before', 'Before EN', $3, NOW(), NOW())
		RETURNING id`,
		int64(workID), sortNumber, animeID,
	).Scan(&id); err != nil {
		t.Fatalf("episodes の挿入に失敗: %v", err)
	}

	return model.EpisodeID(id)
}

// insertMappedUpdateTargetEpisode inserts an episode already mapped to an anime with its
// kind='episode' classification, the shape the phase 2 sync leaves behind and the one the
// update dual-writes into.
//
// [Ja] insertMappedUpdateTargetEpisode は、kind='episode' の分類とともに anime へマッピング済み
// のエピソードを挿入する。フェーズ 2 の同期が残す形であり、更新が両書きする先でもある。
func insertMappedUpdateTargetEpisode(t *testing.T, db *sql.DB, workID model.WorkID, parentAnimeID model.AnimeID, sortNumber int32) (model.EpisodeID, model.AnimeID) {
	t.Helper()

	episodeAnimeID := insertBareAnime(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{Int64: int64(episodeAnimeID), Valid: true}, sortNumber)

	if _, err := db.Exec(`
		INSERT INTO anime_classifications (anime_id, kind, parent_anime_id, number, number_text, sort_number, standalone)
		VALUES ($1, 'episode', $2, 1, '#1', $3, false)`,
		int64(episodeAnimeID), int64(parentAnimeID), sortNumber,
	); err != nil {
		t.Fatalf("anime_classifications の挿入に失敗: %v", err)
	}

	return episodeID, episodeAnimeID
}

// readUpdateTargetVersion returns the version an episode's form would carry, which a submit has
// to state to be accepted.
//
// [Ja] readUpdateTargetVersion はエピソードのフォームが運ぶ版を返す。送信が受理されるには、この
// 版を名乗る必要がある。
func readUpdateTargetVersion(t *testing.T, db *sql.DB, episodeID model.EpisodeID) string {
	t.Helper()

	var updatedAt sql.NullTime
	if err := db.QueryRow(`SELECT updated_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&updatedAt); err != nil {
		t.Fatalf("エピソードの版の読み込みに失敗: %v", err)
	}
	if !updatedAt.Valid {
		return validator.DBEpisodeNullVersion
	}

	return updatedAt.Time.UTC().Format(validator.DBEpisodeVersionLayout)
}

// updateEpisodeSubmit returns a submit that changes every editable field, with the version the
// episode currently carries.
//
// [Ja] updateEpisodeSubmit は編集できる全フィールドを変更する送信を、エピソードが現在持つ版と
// ともに返す。
func updateEpisodeSubmit(t *testing.T, db *sql.DB, episodeID model.EpisodeID, user *model.User) UpdateEpisodeInput {
	t.Helper()

	return UpdateEpisodeInput{
		EpisodeID:  episodeID,
		User:       user,
		Number:     "第2話",
		RawNumber:  "2.5",
		SortNumber: "250",
		Title:      "もう、お婿にいけません",
		TitleEn:    "No Longer Marriageable",
		UpdatedAt:  readUpdateTargetVersion(t, db, episodeID),
	}
}

// TestUpdateEpisodeUsecase_Execute_DualWritesAnimeAndClassification covers the update path of an
// episode that is already mapped: the submitted values reach both the episodes row and the
// anime / classification the reference model derives from it.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_DualWritesAnimeAndClassification は、既にマッピング済み
// のエピソードの更新経路を検証する。送信された値が episodes の行と、そこから参照モデルが導出する
// anime / 分類の双方に届く。
func TestUpdateEpisodeUsecase_Execute_DualWritesAnimeAndClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	user := insertCreateActor(t, db, model.RoleEditor)

	output, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.EpisodeID != episodeID || output.WorkID != workID {
		t.Errorf("output = %+v, want {EpisodeID:%d WorkID:%d}", output, int64(episodeID), int64(workID))
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.number.String != "第2話" || stored.title.String != "もう、お婿にいけません" {
		t.Errorf("(number, title) = (%q, %q), want (\"第2話\", \"もう、お婿にいけません\")", stored.number.String, stored.title.String)
	}
	if stored.rawNumber.Float64 != 2.5 {
		t.Errorf("raw_number = %v, want 2.5", stored.rawNumber)
	}
	if stored.sortNumber != 250 {
		t.Errorf("sort_number = %d, want 250", stored.sortNumber)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Title.String != "もう、お婿にいけません" {
		t.Errorf("anime.Title = %q, want %q", anime.Title.String, "もう、お婿にいけません")
	}
	if anime.TitleEn.String != "No Longer Marriageable" {
		t.Errorf("anime.TitleEn = %q, want %q", anime.TitleEn.String, "No Longer Marriageable")
	}
	// The episode does not source title_ro, so the update carries the stored one over instead
	// of blanking the column the form has no field for.
	//
	// [Ja] エピソードは title_ro を source としないため、更新は保存済みの値を引き継ぐ。フォームに
	// 欄の無いカラムを空にしないようにするため。
	if anime.TitleRo.String != "Before" {
		t.Errorf("anime.TitleRo = %q, want %q", anime.TitleRo.String, "Before")
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), episodeAnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.NumberText.String != "第2話" {
		t.Errorf("classification.NumberText = %q, want %q", classification.NumberText.String, "第2話")
	}
	if classification.Number.String != "2.5" {
		t.Errorf("classification.Number = %q, want %q", classification.Number.String, "2.5")
	}
	if !classification.SortNumber.Valid || classification.SortNumber.Int32 != 250 {
		t.Errorf("classification.SortNumber = %+v, want {250 true}", classification.SortNumber)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v, want %d", classification.ParentAnimeID, int64(parentAnimeID))
	}
}

// TestUpdateEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode covers an episode whose work has
// no anime yet: only the episodes row is written, and the phase 2 sync creates the anime once
// the work is synced.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode は、作品がまだ anime を
// 持たないエピソードを検証する。episodes の行だけを書き、その anime は作品が同期された後に
// フェーズ 2 の同期が作る。
func TestUpdateEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	user := insertCreateActor(t, db, model.RoleEditor)

	if _, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user)); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "もう、お婿にいけません" {
		t.Errorf("title = %q, want %q", stored.title.String, "もう、お婿にいけません")
	}
	if stored.animeID.Valid {
		t.Errorf("episodes.anime_id = %+v, want NULL (未マッピングのまま)", stored.animeID)
	}
}

// TestUpdateEpisodeUsecase_Execute_SkipsAnimeWhenParentMappingIsMissing covers a partially
// stale mapping: the episode still points at an anime, but its parent work no longer does. The
// episodes row remains editable, while the anime and its classification stay untouched until
// the parent is mapped and phase 2 sync can derive a valid parent_anime_id.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_SkipsAnimeWhenParentMappingIsMissing は部分的に古い写像を
// 検証する。episode は anime を指したままだが、親作品は anime を指していない。episodes の行は
// 編集できる一方、親が再度マッピングされてフェーズ 2 同期が有効な parent_anime_id を導出できる
// までは anime とその分類に触れない。
func TestUpdateEpisodeUsecase_Execute_SkipsAnimeWhenParentMappingIsMissing(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	if _, err := db.Exec(`UPDATE works SET anime_id = NULL WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("親作品のマッピング解除に失敗: %v", err)
	}
	user := insertCreateActor(t, db, model.RoleEditor)

	if _, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user)); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "もう、お婿にいけません" {
		t.Errorf("episodes.title = %q, want %q", stored.title.String, "もう、お婿にいけません")
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Title.Valid {
		t.Errorf("anime.Title = %+v, want unchanged NULL", anime.Title)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), episodeAnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.NumberText.String != "#1" || classification.SortNumber.Int32 != 100 {
		t.Errorf("classification = %+v, want pre-update number_text=#1 sort_number=100", classification)
	}
}

// TestUpdateEpisodeUsecase_Execute_RecreatesMissingClassification covers a mapped episode whose
// classification was removed independently. The edit recreates the classification in the same
// transaction as the episodes / anime writes, and an immediate sync then reports Unchanged.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_RecreatesMissingClassification は、分類だけが独立して
// 削除されたマッピング済みエピソードを検証する。編集は episodes / anime と同じトランザクションで
// 分類を再作成し、直後の同期は Unchanged を報告する。
func TestUpdateEpisodeUsecase_Execute_RecreatesMissingClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	if _, err := db.Exec(`DELETE FROM anime_classifications WHERE anime_id = $1`, int64(episodeAnimeID)); err != nil {
		t.Fatalf("分類の削除に失敗: %v", err)
	}
	user := insertCreateActor(t, db, model.RoleEditor)

	if _, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user)); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), episodeAnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v, want %d", classification.ParentAnimeID, int64(parentAnimeID))
	}
	if classification.NumberText.String != "第2話" || classification.Number.String != "2.5" {
		t.Errorf("classification = %+v, want submitted numbering", classification)
	}
	if !classification.SortNumber.Valid || classification.SortNumber.Int32 != 250 {
		t.Errorf("classification.SortNumber = %+v, want {250 true}", classification.SortNumber)
	}

	syncUC := newSyncEpisodesUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("sync result = %+v, want {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestUpdateEpisodeUsecase_Execute_ProducesSyncConsistentMapping is the invariant that justifies
// reusing the sync mapping helpers: a sync run right after an update must detect no diff
// (Unchanged), proving update and sync derive the same anime / classification from the episode
// and the update path never inflates the diff metric.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_ProducesSyncConsistentMapping は同期の写像ヘルパー再利用
// を正当化する不変条件。更新直後の同期実行は差分なし (Unchanged) を検出しなければならず、update
// と同期が同じ anime / 分類をエピソードから導出していること、update 経路が差分メトリクスを
// 水増ししないことを示す。
func TestUpdateEpisodeUsecase_Execute_ProducesSyncConsistentMapping(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, _ := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	user := insertCreateActor(t, db, model.RoleEditor)

	if _, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user)); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	syncUC := newSyncEpisodesUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("sync result = %+v, want {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestUpdateEpisodeUsecase_Execute_KeepsArchivedAnimeArchived covers a content edit of an
// archived episode: the state timestamps the form does not touch are carried over, so the
// dual-write does not republish the anime behind the editor's back.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_KeepsArchivedAnimeArchived は、非公開エピソードの内容
// 編集を検証する。フォームが触れない状態のタイムスタンプが引き継がれるため、両書きが編集者の
// 知らないうちに anime を再公開することはない。
func TestUpdateEpisodeUsecase_Execute_KeepsArchivedAnimeArchived(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	if _, err := db.Exec(`UPDATE episodes SET unpublished_at = NOW() WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("エピソードの非公開化に失敗: %v", err)
	}
	if _, err := db.Exec(`UPDATE animes SET status = 'archived' WHERE id = $1`, int64(episodeAnimeID)); err != nil {
		t.Fatalf("anime の非公開化に失敗: %v", err)
	}
	user := insertCreateActor(t, db, model.RoleEditor)

	if _, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user)); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q, want %q", anime.Status, model.AnimeStatusArchived)
	}
}

// TestUpdateEpisodeUsecase_Execute_RejectsStaleVersion covers two editors submitting from the
// same form: the second submit is reported as a conflict and leaves the first one's values in
// place instead of overwriting them.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_RejectsStaleVersion は、2 人の編集者が同じフォームから
// 送信する場合を検証する。2 件目は競合として報告され、1 件目の値を上書きせずに残す。
func TestUpdateEpisodeUsecase_Execute_RejectsStaleVersion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	user := insertCreateActor(t, db, model.RoleEditor)

	// Both editors open the form at the same version.
	//
	// [Ja] 2 人の編集者は同じ版でフォームを開く。
	shared := updateEpisodeSubmit(t, db, episodeID, user)

	first := shared
	first.Title = "先に保存したタイトル"
	if _, err := uc.Execute(context.Background(), first); err != nil {
		t.Fatalf("1 件目の Execute() error = %v", err)
	}

	second := shared
	second.Title = "後から届いたタイトル"
	_, err := uc.Execute(context.Background(), second)
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Fatalf("2 件目の Execute() error = %v, want AppErrCodeConflict", err)
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "先に保存したタイトル" {
		t.Errorf("title = %q, want %q (後の送信は上書きしない)", stored.title.String, "先に保存したタイトル")
	}
}

// TestUpdateEpisodeUsecase_Execute_NullVersion covers an episode written before updated_at was
// populated: the NULL version is accepted once, and the write that advances the column makes a
// second submit from the same form conflict.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_NullVersion は、updated_at が埋まる前に書かれた
// エピソードを検証する。NULL の版は 1 度だけ受理され、カラムを進める書き込みによって同じ
// フォームからの 2 件目は競合する。
func TestUpdateEpisodeUsecase_Execute_NullVersion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	if _, err := db.Exec(`UPDATE episodes SET updated_at = NULL WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("updated_at の NULL 化に失敗: %v", err)
	}
	user := insertCreateActor(t, db, model.RoleEditor)

	submit := updateEpisodeSubmit(t, db, episodeID, user)
	if submit.UpdatedAt != validator.DBEpisodeNullVersion {
		t.Fatalf("フォームが運ぶ版 = %q, want %q", submit.UpdatedAt, validator.DBEpisodeNullVersion)
	}

	if _, err := uc.Execute(context.Background(), submit); err != nil {
		t.Fatalf("1 件目の Execute() error = %v", err)
	}

	_, err := uc.Execute(context.Background(), submit)
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Fatalf("2 件目の Execute() error = %v, want AppErrCodeConflict", err)
	}
}

// TestUpdateEpisodeUsecase_Execute_RejectsEmptyVersion covers a submit that states no version at
// all, which is not the same as the NULL sentinel: accepting it would let a crafted request skip
// the check that stops one editor from overwriting another.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_RejectsEmptyVersion は、版をまったく示さない送信を検証
// する。これは NULL のセンチネルとは別物で、受理すると、ある編集者が別の編集者を上書きするのを
// 止める検査を、改変されたリクエストが素通りできてしまう。
func TestUpdateEpisodeUsecase_Execute_RejectsEmptyVersion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	user := insertCreateActor(t, db, model.RoleEditor)

	submit := updateEpisodeSubmit(t, db, episodeID, user)
	submit.UpdatedAt = ""

	_, err := uc.Execute(context.Background(), submit)
	if ve := model.AsValidationError(err); ve == nil {
		t.Fatalf("Execute() error = %v, want *model.ValidationError", err)
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "編集前のタイトル" {
		t.Errorf("title = %q, want %q (却下された送信は行を書かない)", stored.title.String, "編集前のタイトル")
	}
}

// TestUpdateEpisodeUsecase_Execute_NotFound covers submits against episodes that have no edit
// form: one that never existed, a deleted one, and one whose work was deleted.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_NotFound は、編集フォームを持たないエピソードへの送信を
// 検証する。存在しないもの、削除済みのもの、作品が削除済みのもの。
func TestUpdateEpisodeUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)
	user := insertCreateActor(t, db, model.RoleEditor)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	deletedEpisodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	deletedEpisodeSubmit := updateEpisodeSubmit(t, db, deletedEpisodeID, user)
	if _, err := db.Exec(`UPDATE episodes SET deleted_at = NOW() WHERE id = $1`, int64(deletedEpisodeID)); err != nil {
		t.Fatalf("エピソードの削除に失敗: %v", err)
	}

	deletedWorkID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeOfDeletedWorkID := insertUpdateTargetEpisode(t, db, deletedWorkID, sql.NullInt64{}, 100)
	episodeOfDeletedWorkSubmit := updateEpisodeSubmit(t, db, episodeOfDeletedWorkID, user)
	if _, err := db.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(deletedWorkID)); err != nil {
		t.Fatalf("作品の削除に失敗: %v", err)
	}

	missingSubmit := deletedEpisodeSubmit
	missingSubmit.EpisodeID = model.EpisodeID(999999999)

	for _, tt := range []struct {
		name  string
		input UpdateEpisodeInput
	}{
		{name: "存在しないエピソード", input: missingSubmit},
		{name: "削除済みのエピソード", input: deletedEpisodeSubmit},
		{name: "削除済み作品のエピソード", input: episodeOfDeletedWorkSubmit},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), tt.input)
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Fatalf("Execute() error = %v, want AppErrCodeResourceNotFound", err)
			}
		})
	}
}

// TestUpdateEpisodeUsecase_Execute_RequiresCommitter keeps the role rule with the use case
// rather than with the route, so an entry point added later cannot reach it without one.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_RequiresCommitter はロールの規則をルートではなく
// UseCase 側に置いていることを検証する。後から増える経路がロール確認を経ずに到達できないように
// するため。
func TestUpdateEpisodeUsecase_Execute_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	submit := updateEpisodeSubmit(t, db, episodeID, nil)

	for _, tt := range []struct {
		name string
		user *model.User
	}{
		{name: "ユーザーを伴わない呼び出し", user: nil},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := submit
			input.User = tt.user

			_, err := uc.Execute(context.Background(), input)
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("Execute() error = %v, want AppErrCodeForbidden", err)
			}
		})
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "編集前のタイトル" {
		t.Errorf("title = %q, want %q (拒否された呼び出しは行を書かない)", stored.title.String, "編集前のタイトル")
	}
}

// TestUpdateEpisodeUsecase_Execute_RecordsRailsSaveSideEffects covers the side effects the Rails
// update performs alongside the row: the change history the shared admin screen reads, and the
// parent work's timestamp.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_RecordsRailsSaveSideEffects は、Rails の更新が行と一緒に
// 行う副作用を検証する。共有の管理画面が読む変更履歴と、親作品のタイムスタンプ。
func TestUpdateEpisodeUsecase_Execute_RecordsRailsSaveSideEffects(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	user := insertCreateActor(t, db, model.RoleEditor)

	before := time.Now()
	if _, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user)); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var activityCount int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM db_activities
		WHERE user_id = $1
			AND trackable_type = 'Episode'
			AND trackable_id = $2
			AND action = 'episodes.update'
			AND root_resource_type = 'Work'
			AND root_resource_id = $3
	`, int64(user.ID), int64(episodeID), int64(workID)).Scan(&activityCount); err != nil {
		t.Fatalf("DB 活動履歴件数の読み込みに失敗: %v", err)
	}
	if activityCount != 1 {
		t.Errorf("DB 活動履歴 = %d 件, want 1", activityCount)
	}

	var workUpdatedAt time.Time
	if err := db.QueryRow(`SELECT updated_at FROM works WHERE id = $1`, int64(workID)).Scan(&workUpdatedAt); err != nil {
		t.Fatalf("作品の保存副作用の読み込みに失敗: %v", err)
	}
	if workUpdatedAt.Before(before) {
		t.Errorf("works.updated_at = %v, want >= %v", workUpdatedAt, before)
	}
}

// newRetryOnlyUpdateEpisodeUsecase builds a usecase whose retry bound is the production one but
// whose backoff is short enough to run in a unit test. Only the retry helpers are exercised, so
// no repositories are wired.
//
// [Ja] newRetryOnlyUpdateEpisodeUsecase は、再試行の上限は本番と同じまま、backoff だけを
// ユニットテストで回せる長さにした UseCase を組み立てる。再試行のヘルパーだけを対象にするため、
// Repository は組み込まない。
func newRetryOnlyUpdateEpisodeUsecase() *UpdateEpisodeUsecase {
	return &UpdateEpisodeUsecase{
		lockRetryLimit:     defaultUpdateEpisodeLockRetryLimit,
		lockRetryBaseDelay: time.Microsecond,
	}
}

// TestRetryEpisodeUpdateLock verifies that only the repository's NOWAIT lock miss reruns the
// whole supplied attempt, that the reruns stop at the configured limit, and that any other
// error is returned from the first attempt.
//
// [Ja] TestRetryEpisodeUpdateLock は、Repository の NOWAIT ロック取得失敗だけが渡された試行全体を
// やり直すこと、やり直しが設定された上限で止まること、それ以外のエラーは 1 回目でそのまま返る
// ことを検証する。
func TestRetryEpisodeUpdateLock(t *testing.T) {
	t.Parallel()

	t.Run("ロック取得失敗は試行全体を再実行する", func(t *testing.T) {
		t.Parallel()

		uc := newRetryOnlyUpdateEpisodeUsecase()
		attempts := 0
		want := &UpdateEpisodeOutput{EpisodeID: 10, WorkID: 20}
		got, err := uc.retryEpisodeUpdateLock(context.Background(), func() (*UpdateEpisodeOutput, error) {
			attempts++
			if attempts < 3 {
				return nil, fmt.Errorf("エピソードの更新に失敗しました: %w", repository.ErrEpisodeLockUnavailable)
			}
			return want, nil
		})
		if err != nil {
			t.Fatalf("retryEpisodeUpdateLock() error = %v", err)
		}
		if got != want {
			t.Errorf("retryEpisodeUpdateLock() = %+v, want %+v", got, want)
		}
		if attempts != 3 {
			t.Errorf("attempts = %d, want 3", attempts)
		}
	})

	// The bound is what keeps a submit that can never get the lock from holding the request
	// open indefinitely, and the returned error is what updateEpisode turns into the response
	// the editor sees.
	//
	// [Ja] 上限は、ロックを取れない送信がリクエストを開いたままにするのを防ぐためのもので、
	// 返るエラーは updateEpisode が編集者に見せる応答へ変換するものである。
	t.Run("ロックを取れないままなら上限で打ち切り、最後のエラーを返す", func(t *testing.T) {
		t.Parallel()

		uc := newRetryOnlyUpdateEpisodeUsecase()
		attempts := 0
		_, err := uc.retryEpisodeUpdateLock(context.Background(), func() (*UpdateEpisodeOutput, error) {
			attempts++
			return nil, fmt.Errorf("エピソードの更新に失敗しました: %w", repository.ErrEpisodeLockUnavailable)
		})
		if !errors.Is(err, repository.ErrEpisodeLockUnavailable) {
			t.Errorf("retryEpisodeUpdateLock() error = %v, want ErrEpisodeLockUnavailable", err)
		}
		if attempts != defaultUpdateEpisodeLockRetryLimit {
			t.Errorf("attempts = %d, want %d", attempts, defaultUpdateEpisodeLockRetryLimit)
		}
	})

	t.Run("ロック取得失敗以外は再実行しない", func(t *testing.T) {
		t.Parallel()

		uc := newRetryOnlyUpdateEpisodeUsecase()
		attempts := 0
		wantErr := errors.New("保存に失敗")
		_, err := uc.retryEpisodeUpdateLock(context.Background(), func() (*UpdateEpisodeOutput, error) {
			attempts++
			return nil, wantErr
		})
		if !errors.Is(err, wantErr) {
			t.Errorf("retryEpisodeUpdateLock() error = %v, want %v", err, wantErr)
		}
		if attempts != 1 {
			t.Errorf("attempts = %d, want 1", attempts)
		}
	})
}

// TestUpdateEpisodeUsecase_Execute_ReportsBusyWhenLockNeverFrees fixes what a submit gets when
// every attempt finds a row it needs locked. Nothing is written and the version the form
// carries still matches, so this must not be reported as the version conflict a concurrent edit
// produces: the editor would be sent to compare their input against stored values that have not
// moved.
//
// [Ja] TestUpdateEpisodeUsecase_Execute_ReportsBusyWhenLockNeverFrees は、どの試行でも必要な行が
// ロックされていた送信が何を受け取るかを固定する。何も書かれず、フォームが運ぶ版も一致したまま
// のため、並行編集が生む版の競合として報告してはならない。そうすると編集者は、動いていない
// 保存済みの値と自分の入力を見比べに行かされてしまう。
func TestUpdateEpisodeUsecase_Execute_ReportsBusyWhenLockNeverFrees(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)
	uc.lockRetryBaseDelay = time.Millisecond

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	user := insertCreateActor(t, db, model.RoleEditor)
	submit := updateEpisodeSubmit(t, db, episodeID, user)

	// Hold the target the way a Rails write does: the row is locked for the whole of another
	// transaction, so every attempt the usecase makes hits NOWAIT.
	//
	// [Ja] Rails の書き込みと同じ形で対象行を保持する。行は別トランザクションの間ずっとロック
	// されるため、UseCase のどの試行も NOWAIT に当たる。
	holdTx, err := db.Begin()
	if err != nil {
		t.Fatalf("ロック保持トランザクションの Begin() に失敗: %v", err)
	}
	defer func() { _ = holdTx.Rollback() }()
	if _, err := holdTx.Exec("UPDATE episodes SET title = title WHERE id = $1", int64(episodeID)); err != nil {
		t.Fatalf("対象行のロック取得に失敗: %v", err)
	}

	_, err = uc.Execute(context.Background(), submit)

	appErr := model.AsAppError(err)
	if appErr == nil {
		t.Fatalf("Execute() error = %v, want *model.AppError", err)
	}
	if appErr.Code != model.AppErrCodeBusy {
		t.Errorf("Execute() AppError.Code = %v, want AppErrCodeBusy (%v)", appErr.Code, model.AppErrCodeBusy)
	}
	if appErr.UserMsg != i18n.T(context.Background(), "validation_record_busy") {
		t.Errorf("Execute() AppError.UserMsg = %q, want the record-busy message", appErr.UserMsg)
	}

	// The stored row must be untouched: the message tells the editor to send the same submit
	// again, which only works while the version they hold still matches.
	//
	// [Ja] 保存済みの行は触られていないこと。メッセージは同じ送信をもう一度送るよう伝えるが、
	// それが成立するのは編集者が持つ版が一致したままの間だけである。
	var storedTitle string
	var storedSortNumber int32
	if err := db.QueryRow(
		"SELECT title, sort_number FROM episodes WHERE id = $1", int64(episodeID),
	).Scan(&storedTitle, &storedSortNumber); err != nil {
		t.Fatalf("保存済みのエピソードの読み込みに失敗: %v", err)
	}
	if storedTitle != "編集前のタイトル" || storedSortNumber != 100 {
		t.Errorf("保存済みの行 = (%q, %d), want (%q, %d)", storedTitle, storedSortNumber, "編集前のタイトル", 100)
	}
}
