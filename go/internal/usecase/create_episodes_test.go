package usecase

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// newCreateEpisodesUsecaseは共有テストDBに対して一括作成UseCaseを組み立てる。
// 本UseCaseは内部で自前のトランザクションを開くため、テストはSetupTxではなくGetTestDBを
// 使い、コミット済みの行がUseCaseの内側トランザクションと後続の同期不変条件チェックから
// 見えるようにする。
func newCreateEpisodesUsecase(db *sql.DB) *CreateEpisodesUsecase {
	queries := query.New(db)
	return NewCreateEpisodesUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
		repository.NewAnimeClassificationRepository(queries),
		validator.NewDBEpisodeCreateValidator(),
	)
}

// createEpisodesSeasonYearは、本テスト群がコミットする作品を、作品一覧が全体に対して
// 数える「シーズン未設定」の集合から外すためのもの。行は共有テストDBにコミットされテストの
// 寿命を超えて残るため、同時に走る他パッケージからも見える。シーズンの無い作品は、そちらの
// 一覧のシーズン未設定フィルタに数えられてしまう。
const createEpisodesSeasonYear = 1903

// insertCreateActorは一括作成テストの送信者となるユーザーを、UseCase自身の
// トランザクションから行を帰属させられるよう共有プールにコミットして挿入する。
func insertCreateActor(t *testing.T, db *sql.DB, role int32) *model.User {
	t.Helper()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("ユーザー作成トランザクションの開始に失敗: %v", err)
	}
	userID := testutil.NewUserBuilder(t, tx).WithRole(role).Build()
	if err := tx.Commit(); err != nil {
		t.Fatalf("ユーザー作成トランザクションのコミットに失敗: %v", err)
	}

	t.Cleanup(func() { deleteCreateActor(t, db, userID) })

	return &model.User{ID: userID, Role: role}
}

// unsavedCreateActorは、何も書かれる前に却下される送信のための編集者を返す。その経路を
// 検証するテストが共有DBにユーザーをコミットせずに済むようにする。
func unsavedCreateActor() *model.User {
	return &model.User{ID: 1, Role: model.RoleEditor}
}

// deleteCreateActorはinsertCreateActorが挿入したユーザーを削除する。送信が記録した
// 活動履歴がユーザーを参照するため先に消し、残りのユーザーの行はbuilder側の後始末に任せる。
func deleteCreateActor(t *testing.T, db *sql.DB, userID model.UserID) {
	t.Helper()

	if _, err := db.Exec(`DELETE FROM db_activities WHERE user_id = $1`, int64(userID)); err != nil {
		t.Errorf("DB活動履歴の後始末に失敗: %v", err)
	}
	testutil.DeleteUser(t, db, userID)
}

// insertCreateTargetWorkは一括作成テストが行を送信する親作品を挿入し、テスト終了時に
// その作品とエピソードの削除を試みる。他パッケージがエピソードへの参照をコミット済みの場合は、
// 失敗をログに残し、行は次回のテストDBリセットまで残す。animeIDはworks.anime_idの
// マッピングカラムで、まだマッピングされていない作品には無効なNullInt64を渡す。
func insertCreateTargetWork(t *testing.T, db *sql.DB, animeID sql.NullInt64) model.WorkID {
	t.Helper()

	var id int64
	if err := db.QueryRow(`
		INSERT INTO works (title, media, start_episode_raw_number, season_year, season_name, anime_id, created_at, updated_at)
		VALUES ($1, $2, 1.0, $3, 1, $4, NOW(), NOW() - INTERVAL '1 day')
		RETURNING id
	`, "一括作成テストアニメ_"+t.Name(), workMediaTV, createEpisodesSeasonYear, animeID).Scan(&id); err != nil {
		t.Fatalf("worksの挿入に失敗: %v", err)
	}

	t.Cleanup(func() { deleteCreateTargetWork(t, db, id) })

	return model.WorkID(id)
}

// deleteCreateTargetWorkはinsertCreateTargetWorkが挿入した作品と、その作品への送信が
// 作った行の削除を試みる。失敗した文はテストを落とさずログに残す。internal/usecase/seedは
// DB全体からエピソードをランダムに選び、それを参照するepisode_records / activitiesを
// コミットするため、両パッケージが並行して走る間はこの削除が外部キーに阻まれることが正当に
// 起こりうる。ログに残すことで、完了できなかった後始末を、通常の並行実行をflakyな失敗に
// 変えずに観測できる。
func deleteCreateTargetWork(t *testing.T, db *sql.DB, workID int64) {
	t.Helper()

	statements := []string{
		`DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1`,
		`DELETE FROM episodes WHERE work_id = $1`,
		`DELETE FROM works WHERE id = $1`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement, workID); err != nil {
			t.Logf("作品の後始末に失敗 (%s): %v", statement, err)
		}
	}
}

// insertMappedCreateTargetWorkは新規animeにマッピング済みの親作品を挿入し、その配下に
// 作られるエピソードを参照モデルへ両書きできるようにする。
func insertMappedCreateTargetWork(t *testing.T, db *sql.DB) (model.WorkID, model.AnimeID) {
	t.Helper()

	parentAnimeID := insertBareAnime(t, db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{Int64: int64(parentAnimeID), Valid: true})

	return workID, parentAnimeID
}

// createdEpisodeRowは作成されたエピソードについてアサーションが読み戻す内容。フォームが
// 入力するカラムと、作成が併せて書く2つのマッピング / 導線のカラム。
type createdEpisodeRow struct {
	number        sql.NullString
	rawNumber     sql.NullFloat64
	title         sql.NullString
	sortNumber    int32
	prevEpisodeID sql.NullInt64
	animeID       sql.NullInt64
}

func readCreatedEpisode(t *testing.T, db *sql.DB, id model.EpisodeID) createdEpisodeRow {
	t.Helper()

	var row createdEpisodeRow
	if err := db.QueryRow(`
		SELECT number, raw_number, title, sort_number, prev_episode_id, anime_id
		FROM episodes
		WHERE id = $1
	`, int64(id)).Scan(&row.number, &row.rawNumber, &row.title, &row.sortNumber, &row.prevEpisodeID, &row.animeID); err != nil {
		t.Fatalf("作成されたエピソードの読み込みに失敗: %v", err)
	}
	return row
}

// TestCreateEpisodesUsecase_Execute_CreatesRowsWithAnimeAndClassificationは、既にanimeに
// マッピング済みの作品での作成経路を検証する。送信された各行がエピソードと、そのanimeおよび
// kind='episode' の分類になり、採番は作品が既に持つエピソードの先から始まる。
func TestCreateEpisodesUsecase_Execute_CreatesRowsWithAnimeAndClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	existingID := insertSyncEpisode(t, db, generationEpisodeInput(workID, 1, 100))
	user := insertCreateActor(t, db, model.RoleEditor)

	output, err := uc.Execute(context.Background(), CreateEpisodesInput{
		WorkID: workID,
		User:   user,
		Rows:   "#2,2,もう、お婿にいけません\n#3,3.5,まずいよ☆先生",
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if len(output.EpisodeIDs) != 2 {
		t.Fatalf("len(EpisodeIDs) = %d、期待値 = 2", len(output.EpisodeIDs))
	}

	first := readCreatedEpisode(t, db, output.EpisodeIDs[0])
	if first.number.String != "#2" || first.title.String != "もう、お婿にいけません" {
		t.Errorf("1行目 = (%q, %q)、期待値 = (\"#2\", \"もう、お婿にいけません\")", first.number.String, first.title.String)
	}
	if first.rawNumber.Float64 != 2 {
		t.Errorf("1行目のraw_number = %v、期待値 = 2", first.rawNumber)
	}
	// 作品の最大のsort_numberは100のため、最初に作られる行は200に、2件目はその
	// 1ステップ先に着地する。
	if first.sortNumber != 200 {
		t.Errorf("1行目のsort_number = %d、期待値 = 200", first.sortNumber)
	}
	// 最初に作られる行は作品の既存エピソードを直前のエピソードとして名指しし、2件目は
	// 1件目を名指しする。
	if first.prevEpisodeID.Int64 != int64(existingID) {
		t.Errorf("1行目のprev_episode_id = %+v、期待値 = %d", first.prevEpisodeID, int64(existingID))
	}

	second := readCreatedEpisode(t, db, output.EpisodeIDs[1])
	if second.sortNumber != 300 {
		t.Errorf("2行目のsort_number = %d、期待値 = 300", second.sortNumber)
	}
	if second.rawNumber.Float64 != 3.5 {
		t.Errorf("2行目のraw_number = %v、期待値 = 3.5", second.rawNumber)
	}
	if second.prevEpisodeID.Int64 != int64(output.EpisodeIDs[0]) {
		t.Errorf("2行目のprev_episode_id = %+v、期待値 = %d", second.prevEpisodeID, int64(output.EpisodeIDs[0]))
	}

	// 各行はそれぞれ固有のanimeにマッピングされ、その分類は親作品のanimeの下に付く。
	if !second.animeID.Valid {
		t.Fatal("episodes.anime_idが書かれていません")
	}
	if second.animeID.Int64 == first.animeID.Int64 {
		t.Error("行ごとに別のanimeが作られるべきです")
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), model.AnimeID(second.animeID.Int64))
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.Kind != model.AnimeClassificationKindEpisode {
		t.Errorf("classification.Kind = %q、期待値 = episode", classification.Kind)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v、期待値 = %d", classification.ParentAnimeID, int64(parentAnimeID))
	}
	if !classification.SortNumber.Valid || classification.SortNumber.Int32 != 300 {
		t.Errorf("classification.SortNumber = %+v、期待値 = {300 true}", classification.SortNumber)
	}
	if classification.NumberText.String != "#3" {
		t.Errorf("classification.NumberText = %q、期待値 = #3", classification.NumberText.String)
	}

	var (
		episodesCount int32
		workTouched   bool
	)
	if err := db.QueryRow(`
		SELECT episodes_count, updated_at >= created_at
		FROM works
		WHERE id = $1
	`, int64(workID)).Scan(&episodesCount, &workTouched); err != nil {
		t.Fatalf("作品の保存副作用の読み込みに失敗: %v", err)
	}
	if episodesCount != 2 {
		t.Errorf("works.episodes_count = %d、期待値 = 2", episodesCount)
	}
	if !workTouched {
		t.Error("works.updated_atが更新されていません")
	}

	var activityCount int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM db_activities
		WHERE user_id = $1
			AND root_resource_id = $2
			AND root_resource_type = 'Work'
			AND trackable_type = 'Episode'
			AND action = 'episodes.create'
	`, int64(user.ID), int64(workID)).Scan(&activityCount); err != nil {
		t.Fatalf("DB活動履歴件数の読み込みに失敗: %v", err)
	}
	if activityCount != 2 {
		t.Errorf("DB活動履歴 = %d件、期待値 = 2", activityCount)
	}
}

// TestCreateEpisodesUsecase_Execute_SkipsAnimeForUnmappedWorkは、まだanimeを持たない
// 親作品を検証する。エピソードの分類は親のanimeを要するためepisodesの行だけを書き、その
// animeは作品が同期された後にフェーズ2の同期が作る。
func TestCreateEpisodesUsecase_Execute_SkipsAnimeForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})

	user := insertCreateActor(t, db, model.RoleEditor)
	output, err := uc.Execute(context.Background(), CreateEpisodesInput{
		WorkID: workID,
		User:   user,
		Rows:   "#1,1,はじまり",
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if len(output.EpisodeIDs) != 1 {
		t.Fatalf("len(EpisodeIDs) = %d、期待値 = 1", len(output.EpisodeIDs))
	}

	created := readCreatedEpisode(t, db, output.EpisodeIDs[0])
	if created.animeID.Valid {
		t.Errorf("episodes.anime_id = %+v、期待値 = NULL (親作品が未マッピング)", created.animeID)
	}
	// 作品はまだエピソードを持たないため、採番は最初のステップから始まり、直前の
	// エピソードとして名指しする相手もいない。
	if created.sortNumber != 100 {
		t.Errorf("sort_number = %d、期待値 = 100", created.sortNumber)
	}
	if created.prevEpisodeID.Valid {
		t.Errorf("prev_episode_id = %+v、期待値 = NULL", created.prevEpisodeID)
	}
}

// TestCreateEpisodesUsecase_Execute_ProducesSyncConsistentMappingは同期の写像ヘルパー
// 再利用を正当化する不変条件。作成直後の同期実行は差分なし (Unchanged) を検出しなければ
// ならず、createと同期が同じanime / 分類をエピソードから導出していること、create経路が
// 差分メトリクスを水増ししないことを示す。
func TestCreateEpisodesUsecase_Execute_ProducesSyncConsistentMapping(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)

	workID, _ := insertMappedCreateTargetWork(t, db)
	user := insertCreateActor(t, db, model.RoleEditor)

	output, err := uc.Execute(context.Background(), CreateEpisodesInput{
		WorkID: workID,
		User:   user,
		Rows:   "#1,1,はじまり\n,,タイトルだけの話\n#3,,話数なし",
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	syncUC := newSyncEpisodesUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: output.EpisodeIDs})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Processed != 3 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 3 {
		t.Fatalf("同期の結果 = %+v、期待値 = {Processed:3 Created:0 Updated:0 Unchanged:3}", result)
	}
}

func TestCreateEpisodesUsecase_Execute_SerializesConcurrentCreates(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	user := insertCreateActor(t, db, model.RoleEditor)

	const createCount = 8
	start := make(chan struct{})
	errs := make(chan error, createCount)
	var wg sync.WaitGroup
	for range createCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := uc.Execute(context.Background(), CreateEpisodesInput{
				WorkID: workID,
				User:   user,
				Rows:   "#1,1,はじまり",
			})
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("並行実行したExecute()のエラー = %v", err)
		}
	}

	rows, err := db.Query(`
		SELECT id, sort_number, prev_episode_id
		FROM episodes
		WHERE work_id = $1
		ORDER BY sort_number, id
	`, int64(workID))
	if err != nil {
		t.Fatalf("作成されたエピソードの読み込みに失敗: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var (
		index      int
		previousID int64
	)
	for rows.Next() {
		var (
			id            int64
			sortNumber    int32
			prevEpisodeID sql.NullInt64
		)
		if err := rows.Scan(&id, &sortNumber, &prevEpisodeID); err != nil {
			t.Fatalf("作成されたエピソードの走査に失敗: %v", err)
		}

		wantSortNumber := int32(index+1) * episodeSortNumberStep // #nosec G115 -- indexはcreateCountで上限が決まる
		if sortNumber != wantSortNumber {
			t.Errorf("%d件目のsort_number = %d、期待値 = %d", index+1, sortNumber, wantSortNumber)
		}
		if index == 0 && prevEpisodeID.Valid {
			t.Errorf("先頭のprev_episode_id = %+v、期待値 = NULL", prevEpisodeID)
		}
		if index > 0 && (!prevEpisodeID.Valid || prevEpisodeID.Int64 != previousID) {
			t.Errorf("%d件目のprev_episode_id = %+v、期待値 = %d", index+1, prevEpisodeID, previousID)
		}
		previousID = id
		index++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("作成されたエピソードの走査に失敗: %v", err)
	}
	if index != createCount {
		t.Errorf("作成されたエピソード = %d件、期待値 = %d", index, createCount)
	}
}

func TestCreateEpisodesUsecase_Execute_EnforcesManualCreationRestriction(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	editor := insertCreateActor(t, db, model.RoleEditor)
	admin := insertCreateActor(t, db, model.RoleAdmin)
	if _, err := db.Exec(`UPDATE works SET manual_episodes_count = 1 WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("予定エピソード数の更新に失敗: %v", err)
	}
	insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))

	input := CreateEpisodesInput{WorkID: workID, User: editor, Rows: "#2,2,つづき"}
	_, err := uc.Execute(context.Background(), input)
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("編集者のExecute()のエラー = %v、期待値 = *model.ValidationError", err)
	}
	// 制限は送信された行ではなく作品の状態に由来するため、フォーム全体に対して報告する。
	if len(ve.Global) != 1 || len(ve.GetFieldErrors("rows")) != 0 {
		t.Errorf("editorのエラー = Global:%v Fields:%v、期待値 = グローバルに1件のみ", ve.Global, ve.GetFieldErrors("rows"))
	}

	input.User = admin
	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("管理者のExecute()のエラー = %v", err)
	}
	if len(output.EpisodeIDs) != 1 {
		t.Errorf("管理者のlen(EpisodeIDs) = %d、期待値 = 1", len(output.EpisodeIDs))
	}
}

// TestCreateEpisodesUsecase_Execute_ValidationErrorは不正な行を含む送信を検証する。送信
// 全体が失敗するため、正常な行もDBに届かない。
func TestCreateEpisodesUsecase_Execute_ValidationError(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)

	workID, _ := insertMappedCreateTargetWork(t, db)

	output, err := uc.Execute(context.Background(), CreateEpisodesInput{
		WorkID: workID,
		User:   unsavedCreateActor(),
		Rows:   "#1,1,はじまり\n#2,いち,つづき",
	})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil", output)
	}
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("err = %v、期待値 = *model.ValidationError", err)
	}
	if len(ve.GetFieldErrors("rows")) != 1 {
		t.Errorf("rowsのエラー = %v、期待値 = 1件", ve.GetFieldErrors("rows"))
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM episodes WHERE work_id = $1`, int64(workID)).Scan(&count); err != nil {
		t.Fatalf("エピソード件数の取得に失敗: %v", err)
	}
	if count != 0 {
		t.Errorf("作成されたエピソード = %d件、期待値 = 0 (送信全体が失敗するため)", count)
	}
}

// TestCreateEpisodesUsecase_Execute_WorkNotFoundはフォームを出せない作品への送信を検証
// する。存在しないidと削除済みの作品はいずれも、行のパースにも進まず未存在として失敗する。
func TestCreateEpisodesUsecase_Execute_WorkNotFound(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)

	deletedWorkID := insertCreateTargetWork(t, db, sql.NullInt64{})
	if _, err := db.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(deletedWorkID)); err != nil {
		t.Fatalf("作品の削除に失敗: %v", err)
	}

	tests := []struct {
		name   string
		workID model.WorkID
	}{
		{name: "存在しない作品", workID: model.WorkID(999999999)},
		{name: "削除済みの作品", workID: deletedWorkID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), CreateEpisodesInput{
				WorkID: tt.workID,
				User:   unsavedCreateActor(),
				Rows:   "#1,1,はじまり",
			})

			ae := model.AsAppError(err)
			if ae == nil {
				t.Fatalf("err = %v、期待値 = *model.AppError", err)
			}
			if ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("ae.Code = %v、期待値 = %v", ae.Code, model.AppErrCodeResourceNotFound)
			}
		})
	}
}

// TestCreateEpisodesUsecase_Execute_RequiresCommitterは未認証と一般ユーザーの送信を検証
// する。エピソード作成はcommitterに限られ、この確認をUseCaseに置くことで別の入口がweb
// ミドルウェアを迂回できないようにする。
func TestCreateEpisodesUsecase_Execute_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "ユーザーなし"},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), CreateEpisodesInput{
				WorkID: workID,
				User:   tt.user,
				Rows:   "#1,1,はじまり",
			})

			ae := model.AsAppError(err)
			if ae == nil {
				t.Fatalf("err = %v、期待値 = *model.AppError", err)
			}
			if ae.Code != model.AppErrCodeForbidden {
				t.Errorf("ae.Code = %v、期待値 = %v", ae.Code, model.AppErrCodeForbidden)
			}
		})
	}
}

// TestCreateEpisodesUsecase_Execute_AppendsAfterMaxSortNumberは、作品の最大のsort_number
// (削除済みを含む) の先から振り、その最大の行を最初の行の前のエピソードにすることを検証する。
// 開始話数が13の作品で自動生成が1300に置いた第13話より前に並ばないようにするため。
func TestCreateEpisodesUsecase_Execute_AppendsAfterMaxSortNumber(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	insertSyncEpisode(t, db, generationEpisodeInput(workID, 13, 1300))
	deleted := generationEpisodeInput(workID, 14, 1400)
	deleted.deletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	deletedID := insertSyncEpisode(t, db, deleted)
	user := insertCreateActor(t, db, model.RoleEditor)

	output, err := uc.Execute(context.Background(), CreateEpisodesInput{
		WorkID: workID,
		User:   user,
		Rows:   "総集編,13.5,これまでのお話\n#14,14,つづき",
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if len(output.EpisodeIDs) != 2 {
		t.Fatalf("len(EpisodeIDs) = %d、期待値 = 2", len(output.EpisodeIDs))
	}

	first := readCreatedEpisode(t, db, output.EpisodeIDs[0])
	if first.sortNumber != 1500 || first.prevEpisodeID.Int64 != int64(deletedID) {
		t.Errorf("1行目 = (sort_number %d, prev_episode_id %+v)、期待値 = (1500, %d)", first.sortNumber, first.prevEpisodeID, deletedID)
	}
	second := readCreatedEpisode(t, db, output.EpisodeIDs[1])
	if second.sortNumber != 1600 || second.prevEpisodeID.Int64 != int64(output.EpisodeIDs[0]) {
		t.Errorf("2行目 = (sort_number %d, prev_episode_id %+v)、期待値 = (1600, %d)", second.sortNumber, second.prevEpisodeID, output.EpisodeIDs[0])
	}
}

// TestCreateEpisodesUsecase_Execute_RequiresRawNumberForWorkWithSlotsは、有効な放送枠がある
// 作品でも編集者が一括登録でき、そのときはすべての行に話数を求めることを検証する。
func TestCreateEpisodesUsecase_Execute_RequiresRawNumberForWorkWithSlots(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	program := insertGenerationProgram(t, db, workID, 1)
	program.insertSlot(1, time.Now().Add(24*time.Hour), nil)
	editor := insertCreateActor(t, db, model.RoleEditor)
	admin := insertCreateActor(t, db, model.RoleAdmin)

	for _, user := range []*model.User{editor, admin} {
		_, err := uc.Execute(context.Background(), CreateEpisodesInput{
			WorkID: workID,
			User:   user,
			Rows:   "#1,1,はじまり\n総集編,,これまでのお話",
		})
		ve := model.AsValidationError(err)
		if ve == nil {
			t.Fatalf("ロール%dのExecute()のエラー = %v、期待値 = *model.ValidationError", user.Role, err)
		}
		if errs := ve.GetFieldErrors("rows"); len(errs) != 1 || !strings.Contains(errs[0], "2 行目") {
			t.Errorf("ロール%dのrowsのエラー = %v、期待値 = 2行目の話数の不足", user.Role, errs)
		}
	}

	output, err := uc.Execute(context.Background(), CreateEpisodesInput{
		WorkID: workID,
		User:   editor,
		Rows:   "#1,1,はじまり\n総集編,1.5,これまでのお話",
	})
	if err != nil {
		t.Fatalf("話数を入れたExecute()のエラー = %v", err)
	}
	if len(output.EpisodeIDs) != 2 {
		t.Errorf("len(EpisodeIDs) = %d、期待値 = 2", len(output.EpisodeIDs))
	}
}

// TestCreateEpisodesUsecase_Execute_RechecksRawNumberAfterWorkLockは、ロック待ち中に
// 放送枠が追加されても、元の行番号で話数の不足を報告し、一括登録全体を止めることを検証する。
func TestCreateEpisodesUsecase_Execute_RechecksRawNumberAfterWorkLock(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		role int32
	}{
		{name: "編集者", role: model.RoleEditor},
		{name: "管理者", role: model.RoleAdmin},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutil.GetTestDB()
			workID := insertCreateTargetWork(t, db, sql.NullInt64{})
			program := insertGenerationProgram(t, db, workID, 1)
			user := insertCreateActor(t, db, tt.role)
			blocker, err := db.Begin()
			if err != nil {
				t.Fatalf("ロック保持用トランザクションの開始に失敗: %v", err)
			}
			defer func() { _ = blocker.Rollback() }()
			if _, err := blocker.Exec(`SELECT id FROM works WHERE id = $1 FOR UPDATE`, int64(workID)); err != nil {
				t.Fatalf("作品のロックに失敗: %v", err)
			}
			var blockerPID int
			if err := blocker.QueryRow(`SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
				t.Fatalf("ロック保持側のPID取得に失敗: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			type result struct {
				output *CreateEpisodesOutput
				err    error
			}
			done := make(chan result, 1)
			go func() {
				output, err := newCreateEpisodesUsecase(db).Execute(ctx, CreateEpisodesInput{
					WorkID: workID,
					User:   user,
					Rows:   "#1,1,はじまり\n\n\n総集編,,これまでのお話",
				})
				done <- result{output: output, err: err}
			}()

			// 事前の検証が終わり、作品ロックを実際に待っていることをDBで確認する。
			for {
				var blocked bool
				if err := db.QueryRowContext(ctx, `
					SELECT EXISTS (
						SELECT 1 FROM pg_stat_activity WHERE $1 = ANY(pg_blocking_pids(pid))
					)
				`, blockerPID).Scan(&blocked); err != nil {
					t.Fatalf("作品ロック待ちの確認に失敗: %v", err)
				}
				if blocked {
					break
				}
				select {
				case got := <-done:
					t.Fatalf("ロック待ち前に終了しました: %v", got.err)
				default:
				}
				time.Sleep(10 * time.Millisecond)
			}

			testutil.NewSlotBuilder(t, blocker).
				WithWorkID(workID).
				WithChannelID(program.channelID).
				WithProgramID(program.programID).
				WithNumber(1).
				WithStartedAt(time.Now()).Build()
			if err := blocker.Commit(); err != nil {
				t.Fatalf("放送枠追加のコミットに失敗: %v", err)
			}

			got := <-done
			if got.output != nil {
				t.Errorf("Execute()の出力 = %+v、期待値 = nil", got.output)
			}
			ve := model.AsValidationError(got.err)
			if ve == nil {
				t.Fatalf("Execute()のエラー = %v、期待値 = *model.ValidationError", got.err)
			}
			if errs := ve.GetFieldErrors("rows"); len(errs) != 1 || !strings.Contains(errs[0], "4 行目") {
				t.Errorf("rowsのエラー = %v、期待値 = 元の4行目の話数の不足", errs)
			}
			if count := countWorkEpisodes(t, db, workID); count != 0 {
				t.Errorf("作成エピソード数 = %d、期待値 = 0", count)
			}
		})
	}
}
