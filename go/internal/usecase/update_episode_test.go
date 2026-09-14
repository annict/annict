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

// newUpdateEpisodeUsecaseは共有テストDBに対して更新UseCaseを組み立てる。本UseCaseは
// 内部で自前のトランザクションを開くため、テストはSetupTxではなくGetTestDBを使い、コミット
// 済みの行がUseCaseの内側トランザクションと後続の同期不変条件チェックから見えるようにする。
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

// insertUpdateTargetEpisodeは更新テストが編集するエピソードを指定作品の配下に挿入する。
// animeを渡した場合は自身のanimeにマッピング済みにする。タイムスタンプはDBから取るため、
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
		t.Fatalf("episodesの挿入に失敗: %v", err)
	}

	return model.EpisodeID(id)
}

// insertMappedUpdateTargetEpisodeは、kind='episode' の分類とともにanimeへマッピング済み
// のエピソードを挿入する。フェーズ2の同期が残す形であり、更新が両書きする先でもある。
func insertMappedUpdateTargetEpisode(t *testing.T, db *sql.DB, workID model.WorkID, parentAnimeID model.AnimeID, sortNumber int32) (model.EpisodeID, model.AnimeID) {
	t.Helper()

	episodeAnimeID := insertBareAnime(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{Int64: int64(episodeAnimeID), Valid: true}, sortNumber)

	if _, err := db.Exec(`
		INSERT INTO anime_classifications (anime_id, kind, parent_anime_id, number, number_text, sort_number, standalone)
		VALUES ($1, 'episode', $2, 1, '#1', $3, false)`,
		int64(episodeAnimeID), int64(parentAnimeID), sortNumber,
	); err != nil {
		t.Fatalf("anime_classificationsの挿入に失敗: %v", err)
	}

	return episodeID, episodeAnimeID
}

// readUpdateTargetVersionはエピソードのフォームが運ぶ版を返す。送信が受理されるには、この
// 版を名乗る必要がある。
func readUpdateTargetVersion(t *testing.T, db *sql.DB, episodeID model.EpisodeID) string {
	t.Helper()

	var updatedAt sql.NullTime
	if err := db.QueryRow(`SELECT updated_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&updatedAt); err != nil {
		t.Fatalf("エピソードの版の読み込みに失敗: %v", err)
	}
	if !updatedAt.Valid {
		return validator.FormNullVersion
	}

	return updatedAt.Time.UTC().Format(validator.FormVersionLayout)
}

// updateEpisodeSubmitは編集できる全フィールドを変更する送信を、エピソードが現在持つ版と
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

// TestUpdateEpisodeUsecase_Execute_DualWritesAnimeAndClassificationは、既にマッピング済み
// のエピソードの更新経路を検証する。送信された値がepisodesの行と、そこから参照モデルが導出する
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
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.EpisodeID != episodeID || output.WorkID != workID {
		t.Errorf("output = %+v、期待値 = {EpisodeID:%d WorkID:%d}", output, int64(episodeID), int64(workID))
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.number.String != "第2話" || stored.title.String != "もう、お婿にいけません" {
		t.Errorf("(number, title) = (%q, %q)、期待値 = (\"第2話\", \"もう、お婿にいけません\")", stored.number.String, stored.title.String)
	}
	if stored.rawNumber.Float64 != 2.5 {
		t.Errorf("raw_number = %v、期待値 = 2.5", stored.rawNumber)
	}
	if stored.sortNumber != 250 {
		t.Errorf("sort_number = %d、期待値 = 250", stored.sortNumber)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.String != "もう、お婿にいけません" {
		t.Errorf("anime.Title = %q、期待値 = %q", anime.Title.String, "もう、お婿にいけません")
	}
	if anime.TitleEn.String != "No Longer Marriageable" {
		t.Errorf("anime.TitleEn = %q、期待値 = %q", anime.TitleEn.String, "No Longer Marriageable")
	}
	// エピソードはtitle_roをsourceとしないため、更新は保存済みの値を引き継ぐ。フォームに
	// 欄の無いカラムを空にしないようにするため。
	if anime.TitleRo.String != "Before" {
		t.Errorf("anime.TitleRo = %q、期待値 = %q", anime.TitleRo.String, "Before")
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), episodeAnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.NumberText.String != "第2話" {
		t.Errorf("classification.NumberText = %q、期待値 = %q", classification.NumberText.String, "第2話")
	}
	if classification.Number.String != "2.5" {
		t.Errorf("classification.Number = %q、期待値 = %q", classification.Number.String, "2.5")
	}
	if !classification.SortNumber.Valid || classification.SortNumber.Int32 != 250 {
		t.Errorf("classification.SortNumber = %+v、期待値 = {250 true}", classification.SortNumber)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v、期待値 = %d", classification.ParentAnimeID, int64(parentAnimeID))
	}
}

// TestUpdateEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisodeは、作品がまだanimeを
// 持たないエピソードを検証する。episodesの行だけを書き、そのanimeは作品が同期された後に
// フェーズ2の同期が作る。
func TestUpdateEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	user := insertCreateActor(t, db, model.RoleEditor)

	if _, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user)); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "もう、お婿にいけません" {
		t.Errorf("title = %q、期待値 = %q", stored.title.String, "もう、お婿にいけません")
	}
	if stored.animeID.Valid {
		t.Errorf("episodes.anime_id = %+v、期待値 = NULL (未マッピングのまま)", stored.animeID)
	}
}

// TestUpdateEpisodeUsecase_Execute_SkipsAnimeWhenParentMappingIsMissingは部分的に古い写像を
// 検証する。episodeはanimeを指したままだが、親作品はanimeを指していない。episodesの行は
// 編集できる一方、親が再度マッピングされてフェーズ2同期が有効なparent_anime_idを導出できる
// まではanimeとその分類に触れない。
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
		t.Fatalf("Execute()のエラー = %v", err)
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "もう、お婿にいけません" {
		t.Errorf("episodes.title = %q、期待値 = %q", stored.title.String, "もう、お婿にいけません")
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.Valid {
		t.Errorf("anime.Title = %+v、期待値 = NULLのまま", anime.Title)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), episodeAnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.NumberText.String != "#1" || classification.SortNumber.Int32 != 100 {
		t.Errorf("classification = %+v、期待値 = 更新前のnumber_text=#1 sort_number=100", classification)
	}
}

// TestUpdateEpisodeUsecase_Execute_RecreatesMissingClassificationは、分類だけが独立して
// 削除されたマッピング済みエピソードを検証する。編集はepisodes / animeと同じトランザクションで
// 分類を再作成し、直後の同期はUnchangedを報告する。
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
		t.Fatalf("Execute()のエラー = %v", err)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), episodeAnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v、期待値 = %d", classification.ParentAnimeID, int64(parentAnimeID))
	}
	if classification.NumberText.String != "第2話" || classification.Number.String != "2.5" {
		t.Errorf("classification = %+v、期待値 = 送信した話数情報", classification)
	}
	if !classification.SortNumber.Valid || classification.SortNumber.Int32 != 250 {
		t.Errorf("classification.SortNumber = %+v、期待値 = {250 true}", classification.SortNumber)
	}

	syncUC := newSyncEpisodesUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("同期の結果 = %+v、期待値 = {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestUpdateEpisodeUsecase_Execute_ProducesSyncConsistentMappingは同期の写像ヘルパー再利用
// を正当化する不変条件。更新直後の同期実行は差分なし (Unchanged) を検出しなければならず、update
// と同期が同じanime / 分類をエピソードから導出していること、update経路が差分メトリクスを
// 水増ししないことを示す。
func TestUpdateEpisodeUsecase_Execute_ProducesSyncConsistentMapping(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, _ := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	user := insertCreateActor(t, db, model.RoleEditor)

	if _, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user)); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	syncUC := newSyncEpisodesUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("同期の結果 = %+v、期待値 = {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestUpdateEpisodeUsecase_Execute_KeepsArchivedAnimeArchivedは、非公開エピソードの内容
// 編集を検証する。フォームが触れない状態のタイムスタンプが引き継がれるため、両書きが編集者の
// 知らないうちにanimeを再公開することはない。
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
		t.Fatalf("animeの非公開化に失敗: %v", err)
	}
	user := insertCreateActor(t, db, model.RoleEditor)

	if _, err := uc.Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, user)); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q、期待値 = %q", anime.Status, model.AnimeStatusArchived)
	}
}

// TestUpdateEpisodeUsecase_Execute_RejectsStaleVersionは、2人の編集者が同じフォームから
// 送信する場合を検証する。2件目は競合として報告され、1件目の値を上書きせずに残す。
func TestUpdateEpisodeUsecase_Execute_RejectsStaleVersion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	user := insertCreateActor(t, db, model.RoleEditor)

	// 2人の編集者は同じ版でフォームを開く。
	shared := updateEpisodeSubmit(t, db, episodeID, user)

	first := shared
	first.Title = "先に保存したタイトル"
	if _, err := uc.Execute(context.Background(), first); err != nil {
		t.Fatalf("1件目のExecute()のエラー = %v", err)
	}

	second := shared
	second.Title = "後から届いたタイトル"
	_, err := uc.Execute(context.Background(), second)
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Fatalf("2件目のExecute()のエラー = %v、期待値 = AppErrCodeConflict", err)
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "先に保存したタイトル" {
		t.Errorf("title = %q、期待値 = %q (後の送信は上書きしない)", stored.title.String, "先に保存したタイトル")
	}
}

// TestUpdateEpisodeUsecase_Execute_NullVersionは、updated_atが埋まる前に書かれた
// エピソードを検証する。NULLの版は1度だけ受理され、カラムを進める書き込みによって同じ
// フォームからの2件目は競合する。
func TestUpdateEpisodeUsecase_Execute_NullVersion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	if _, err := db.Exec(`UPDATE episodes SET updated_at = NULL WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("updated_atのNULL化に失敗: %v", err)
	}
	user := insertCreateActor(t, db, model.RoleEditor)

	submit := updateEpisodeSubmit(t, db, episodeID, user)
	if submit.UpdatedAt != validator.FormNullVersion {
		t.Fatalf("フォームが運ぶ版 = %q、期待値 = %q", submit.UpdatedAt, validator.FormNullVersion)
	}

	if _, err := uc.Execute(context.Background(), submit); err != nil {
		t.Fatalf("1件目のExecute()のエラー = %v", err)
	}

	_, err := uc.Execute(context.Background(), submit)
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Fatalf("2件目のExecute()のエラー = %v、期待値 = AppErrCodeConflict", err)
	}
}

// TestUpdateEpisodeUsecase_Execute_RejectsEmptyVersionは、版をまったく示さない送信を検証
// する。これはNULLのセンチネルとは別物で、受理すると、ある編集者が別の編集者を上書きするのを
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
		t.Fatalf("Execute()のエラー = %v、期待値 = *model.ValidationError", err)
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "編集前のタイトル" {
		t.Errorf("title = %q、期待値 = %q (却下された送信は行を書かない)", stored.title.String, "編集前のタイトル")
	}
}

// TestUpdateEpisodeUsecase_Execute_NotFoundは、編集フォームを持たないエピソードへの送信を
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
				t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", err)
			}
		})
	}
}

// TestUpdateEpisodeUsecase_Execute_RequiresCommitterはロールの規則をルートではなく
// UseCase側に置いていることを検証する。後から増える経路がロール確認を経ずに到達できないように
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
				t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}

	stored := readCreatedEpisode(t, db, episodeID)
	if stored.title.String != "編集前のタイトル" {
		t.Errorf("title = %q、期待値 = %q (拒否された呼び出しは行を書かない)", stored.title.String, "編集前のタイトル")
	}
}

// TestUpdateEpisodeUsecase_Execute_RecordsRailsSaveSideEffectsは、Railsの更新が行と一緒に
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
		t.Fatalf("Execute()のエラー = %v", err)
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
		t.Fatalf("DB活動履歴件数の読み込みに失敗: %v", err)
	}
	if activityCount != 1 {
		t.Errorf("DB活動履歴 = %d件、期待値 = 1", activityCount)
	}

	var workUpdatedAt time.Time
	if err := db.QueryRow(`SELECT updated_at FROM works WHERE id = $1`, int64(workID)).Scan(&workUpdatedAt); err != nil {
		t.Fatalf("作品の保存副作用の読み込みに失敗: %v", err)
	}
	if workUpdatedAt.Before(before) {
		t.Errorf("works.updated_at = %v、期待値 = %v以降", workUpdatedAt, before)
	}
}

// newRetryOnlyUpdateEpisodeUsecaseは、再試行の上限は本番と同じまま、backoffだけを
// ユニットテストで回せる長さにしたUseCaseを組み立てる。再試行のヘルパーだけを対象にするため、
// Repositoryは組み込まない。
func newRetryOnlyUpdateEpisodeUsecase() *UpdateEpisodeUsecase {
	return &UpdateEpisodeUsecase{
		lockRetryLimit:     defaultUpdateEpisodeLockRetryLimit,
		lockRetryBaseDelay: time.Microsecond,
	}
}

// TestRetryEpisodeUpdateLockは、RepositoryのNOWAITロック取得失敗だけが渡された試行全体を
// やり直すこと、やり直しが設定された上限で止まること、それ以外のエラーは1回目でそのまま返る
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
			t.Fatalf("retryEpisodeUpdateLock()のエラー = %v", err)
		}
		if got != want {
			t.Errorf("retryEpisodeUpdateLock() = %+v、期待値 = %+v", got, want)
		}
		if attempts != 3 {
			t.Errorf("attempts = %d、期待値 = 3", attempts)
		}
	})

	// 上限は、ロックを取れない送信がリクエストを開いたままにするのを防ぐためのもので、
	// 返るエラーはupdateEpisodeが編集者に見せる応答へ変換するものである。
	t.Run("ロックを取れないままなら上限で打ち切り、最後のエラーを返す", func(t *testing.T) {
		t.Parallel()

		uc := newRetryOnlyUpdateEpisodeUsecase()
		attempts := 0
		_, err := uc.retryEpisodeUpdateLock(context.Background(), func() (*UpdateEpisodeOutput, error) {
			attempts++
			return nil, fmt.Errorf("エピソードの更新に失敗しました: %w", repository.ErrEpisodeLockUnavailable)
		})
		if !errors.Is(err, repository.ErrEpisodeLockUnavailable) {
			t.Errorf("retryEpisodeUpdateLock()のエラー = %v、期待値 = ErrEpisodeLockUnavailable", err)
		}
		if attempts != defaultUpdateEpisodeLockRetryLimit {
			t.Errorf("attempts = %d、期待値 = %d", attempts, defaultUpdateEpisodeLockRetryLimit)
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
			t.Errorf("retryEpisodeUpdateLock()のエラー = %v、期待値 = %v", err, wantErr)
		}
		if attempts != 1 {
			t.Errorf("attempts = %d、期待値 = 1", attempts)
		}
	})
}

// TestUpdateEpisodeUsecase_Execute_ReportsBusyWhenLockNeverFreesは、どの試行でも必要な行が
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

	// Railsの書き込みと同じ形で対象行を保持する。行は別トランザクションの間ずっとロック
	// されるため、UseCaseのどの試行もNOWAITに当たる。
	holdTx, err := db.Begin()
	if err != nil {
		t.Fatalf("ロック保持トランザクションのBegin()に失敗: %v", err)
	}
	defer func() { _ = holdTx.Rollback() }()
	if _, err := holdTx.Exec("UPDATE episodes SET title = title WHERE id = $1", int64(episodeID)); err != nil {
		t.Fatalf("対象行のロック取得に失敗: %v", err)
	}

	_, err = uc.Execute(context.Background(), submit)

	appErr := model.AsAppError(err)
	if appErr == nil {
		t.Fatalf("Execute()のエラー = %v、期待値 = *model.AppError", err)
	}
	if appErr.Code != model.AppErrCodeBusy {
		t.Errorf("Execute() AppError.Code = %v、期待値 = AppErrCodeBusy (%v)", appErr.Code, model.AppErrCodeBusy)
	}
	if appErr.UserMsg != i18n.T(context.Background(), "validation_record_busy") {
		t.Errorf("Execute()のAppError.UserMsg = %q、期待値 = レコードの競合を伝えるメッセージ", appErr.UserMsg)
	}

	// 保存済みの行は触られていないこと。メッセージは同じ送信をもう一度送るよう伝えるが、
	// それが成立するのは編集者が持つ版が一致したままの間だけである。
	var storedTitle string
	var storedSortNumber int32
	if err := db.QueryRow(
		"SELECT title, sort_number FROM episodes WHERE id = $1", int64(episodeID),
	).Scan(&storedTitle, &storedSortNumber); err != nil {
		t.Fatalf("保存済みのエピソードの読み込みに失敗: %v", err)
	}
	if storedTitle != "編集前のタイトル" || storedSortNumber != 100 {
		t.Errorf("保存済みの行 = (%q, %d)、期待値 = (%q, %d)", storedTitle, storedSortNumber, "編集前のタイトル", 100)
	}
}
