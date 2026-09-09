package usecase

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// newCreateEpisodesUsecase wires the bulk-create usecase against the shared test DB. The
// usecase opens its own transaction, so its tests use GetTestDB (not SetupTx) so the
// committed rows are visible to the usecase's inner transaction and to the follow-up sync
// invariant check.
//
// [Ja] newCreateEpisodesUsecase は共有テスト DB に対して一括作成 UseCase を組み立てる。
// 本 UseCase は内部で自前のトランザクションを開くため、テストは SetupTx ではなく GetTestDB を
// 使い、コミット済みの行が UseCase の内側トランザクションと後続の同期不変条件チェックから
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

// createEpisodesSeasonYear keeps the works these tests commit out of the "no season" bucket
// the work list counts globally. The rows are committed to the shared test DB and outlive the
// test, so they are visible to the sibling packages running at the same time; a work with no
// season would be counted by the list's no-season filter there.
//
// [Ja] createEpisodesSeasonYear は、本テスト群がコミットする作品を、作品一覧が全体に対して
// 数える「シーズン未設定」の集合から外すためのもの。行は共有テスト DB にコミットされテストの
// 寿命を超えて残るため、同時に走る他パッケージからも見える。シーズンの無い作品は、そちらの
// 一覧のシーズン未設定フィルタに数えられてしまう。
const createEpisodesSeasonYear = 1903

// insertCreateActor inserts the user a bulk-create test submits as, committed to the shared
// pool so the use case's own transaction can attribute its rows to it.
//
// [Ja] insertCreateActor は一括作成テストの送信者となるユーザーを、UseCase 自身の
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

// unsavedCreateActor returns an editor for submits that are refused before anything is
// written, so the tests covering those paths do not have to commit a user to the shared DB.
//
// [Ja] unsavedCreateActor は、何も書かれる前に却下される送信のための編集者を返す。その経路を
// 検証するテストが共有 DB にユーザーをコミットせずに済むようにする。
func unsavedCreateActor() *model.User {
	return &model.User{ID: 1, Role: model.RoleEditor}
}

// deleteCreateActor removes a user inserted by insertCreateActor. The activities the submits
// recorded reference the user, so they go first; the rest of the user's rows are removed by
// the builder's own teardown.
//
// [Ja] deleteCreateActor は insertCreateActor が挿入したユーザーを削除する。送信が記録した
// 活動履歴がユーザーを参照するため先に消し、残りのユーザーの行は builder 側の後始末に任せる。
func deleteCreateActor(t *testing.T, db *sql.DB, userID model.UserID) {
	t.Helper()

	if _, err := db.Exec(`DELETE FROM db_activities WHERE user_id = $1`, int64(userID)); err != nil {
		t.Errorf("DB 活動履歴の後始末に失敗: %v", err)
	}
	testutil.DeleteUser(t, db, userID)
}

// insertCreateTargetWork inserts the parent work a bulk-create test submits rows under and
// attempts to remove it and its episodes when the test ends. If another package has committed
// references to those episodes, the failed cleanup is logged and the rows remain until the next
// test DB reset. animeID is the works.anime_id mapping column; pass an invalid NullInt64 for a
// work that is not mapped yet.
//
// [Ja] insertCreateTargetWork は一括作成テストが行を送信する親作品を挿入し、テスト終了時に
// その作品とエピソードの削除を試みる。他パッケージがエピソードへの参照をコミット済みの場合は、
// 失敗をログに残し、行は次回のテスト DB リセットまで残す。animeID は works.anime_id の
// マッピングカラムで、まだマッピングされていない作品には無効な NullInt64 を渡す。
func insertCreateTargetWork(t *testing.T, db *sql.DB, animeID sql.NullInt64) model.WorkID {
	t.Helper()

	var id int64
	if err := db.QueryRow(`
		INSERT INTO works (title, media, start_episode_raw_number, season_year, season_name, anime_id, created_at, updated_at)
		VALUES ($1, $2, 1.0, $3, 1, $4, NOW(), NOW() - INTERVAL '1 day')
		RETURNING id
	`, "一括作成テストアニメ_"+t.Name(), workMediaTV, createEpisodesSeasonYear, animeID).Scan(&id); err != nil {
		t.Fatalf("works の挿入に失敗: %v", err)
	}

	t.Cleanup(func() { deleteCreateTargetWork(t, db, id) })

	return model.WorkID(id)
}

// deleteCreateTargetWork attempts to remove a work inserted by insertCreateTargetWork together
// with rows created by submissions targeting it. A failed statement is logged instead of
// failing the test:
// internal/usecase/seed picks episodes at random across the whole database and commits
// episode_records / activities that reference them, so while the two packages run side by side
// this delete can legitimately lose to those foreign keys. Logging makes an incomplete cleanup
// observable without turning ordinary concurrency into a flaky failure.
//
// [Ja] deleteCreateTargetWork は insertCreateTargetWork が挿入した作品と、その作品への送信が
// 作った行の削除を試みる。失敗した文はテストを落とさずログに残す。internal/usecase/seed は
// DB 全体からエピソードをランダムに選び、それを参照する episode_records / activities を
// コミットするため、両パッケージが並行して走る間はこの削除が外部キーに阻まれることが正当に
// 起こりうる。ログに残すことで、完了できなかった後始末を、通常の並行実行を flaky な失敗に
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

// insertMappedCreateTargetWork inserts a parent work already mapped to a fresh anime, so the
// episodes created under it can be dual-written into the reference model.
//
// [Ja] insertMappedCreateTargetWork は新規 anime にマッピング済みの親作品を挿入し、その配下に
// 作られるエピソードを参照モデルへ両書きできるようにする。
func insertMappedCreateTargetWork(t *testing.T, db *sql.DB) (model.WorkID, model.AnimeID) {
	t.Helper()

	parentAnimeID := insertBareAnime(t, db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{Int64: int64(parentAnimeID), Valid: true})

	return workID, parentAnimeID
}

// createdEpisodeRow is what the assertions read back about a created episode: the columns the
// form fills in, plus the two mapping / navigation columns the create also writes.
//
// [Ja] createdEpisodeRow は作成されたエピソードについてアサーションが読み戻す内容。フォームが
// 入力するカラムと、作成が併せて書く 2 つのマッピング / 導線のカラム。
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

// TestCreateEpisodesUsecase_Execute_CreatesRowsWithAnimeAndClassification covers the create
// path of a work that is already mapped to an anime: every submitted line becomes an episode
// plus its own anime and kind='episode' classification, and the numbering starts past the
// episodes the work already has.
//
// [Ja] TestCreateEpisodesUsecase_Execute_CreatesRowsWithAnimeAndClassification は、既に anime に
// マッピング済みの作品での作成経路を検証する。送信された各行がエピソードと、その anime および
// kind='episode' の分類になり、採番は作品が既に持つエピソードの先から始まる。
func TestCreateEpisodesUsecase_Execute_CreatesRowsWithAnimeAndClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateEpisodesUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	existingID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))
	user := insertCreateActor(t, db, model.RoleEditor)

	output, err := uc.Execute(context.Background(), CreateEpisodesInput{
		WorkID: workID,
		User:   user,
		Rows:   "#2,2,もう、お婿にいけません\n#3,3.5,まずいよ☆先生",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(output.EpisodeIDs) != 2 {
		t.Fatalf("len(EpisodeIDs) = %d, want 2", len(output.EpisodeIDs))
	}

	first := readCreatedEpisode(t, db, output.EpisodeIDs[0])
	if first.number.String != "#2" || first.title.String != "もう、お婿にいけません" {
		t.Errorf("1 行目 = (%q, %q), want (\"#2\", \"もう、お婿にいけません\")", first.number.String, first.title.String)
	}
	if first.rawNumber.Float64 != 2 {
		t.Errorf("1 行目の raw_number = %v, want 2", first.rawNumber)
	}
	// The work already holds one episode, so the first created row lands on 200 and the
	// second one step further.
	//
	// [Ja] 作品は既に 1 件のエピソードを持つため、最初に作られる行は 200 に、2 件目はその
	// 1 ステップ先に着地する。
	if first.sortNumber != 200 {
		t.Errorf("1 行目の sort_number = %d, want 200", first.sortNumber)
	}
	// The first created row names the work's existing episode as its preceding one, and the
	// second names the first.
	//
	// [Ja] 最初に作られる行は作品の既存エピソードを直前のエピソードとして名指しし、2 件目は
	// 1 件目を名指しする。
	if first.prevEpisodeID.Int64 != int64(existingID) {
		t.Errorf("1 行目の prev_episode_id = %+v, want %d", first.prevEpisodeID, int64(existingID))
	}

	second := readCreatedEpisode(t, db, output.EpisodeIDs[1])
	if second.sortNumber != 300 {
		t.Errorf("2 行目の sort_number = %d, want 300", second.sortNumber)
	}
	if second.rawNumber.Float64 != 3.5 {
		t.Errorf("2 行目の raw_number = %v, want 3.5", second.rawNumber)
	}
	if second.prevEpisodeID.Int64 != int64(output.EpisodeIDs[0]) {
		t.Errorf("2 行目の prev_episode_id = %+v, want %d", second.prevEpisodeID, int64(output.EpisodeIDs[0]))
	}

	// Each row is mapped to an anime of its own, whose classification hangs under the parent
	// work's anime.
	//
	// [Ja] 各行はそれぞれ固有の anime にマッピングされ、その分類は親作品の anime の下に付く。
	if !second.animeID.Valid {
		t.Fatal("episodes.anime_id が書かれていません")
	}
	if second.animeID.Int64 == first.animeID.Int64 {
		t.Error("行ごとに別の anime が作られるべきです")
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), model.AnimeID(second.animeID.Int64))
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.Kind != model.AnimeClassificationKindEpisode {
		t.Errorf("classification.Kind = %q, want episode", classification.Kind)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v, want %d", classification.ParentAnimeID, int64(parentAnimeID))
	}
	if !classification.SortNumber.Valid || classification.SortNumber.Int32 != 300 {
		t.Errorf("classification.SortNumber = %+v, want {300 true}", classification.SortNumber)
	}
	if classification.NumberText.String != "#3" {
		t.Errorf("classification.NumberText = %q, want #3", classification.NumberText.String)
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
		t.Errorf("works.episodes_count = %d, want 2", episodesCount)
	}
	if !workTouched {
		t.Error("works.updated_at が更新されていません")
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
		t.Fatalf("DB 活動履歴件数の読み込みに失敗: %v", err)
	}
	if activityCount != 2 {
		t.Errorf("DB 活動履歴 = %d 件, want 2", activityCount)
	}
}

// TestCreateEpisodesUsecase_Execute_SkipsAnimeForUnmappedWork covers a parent work that has no
// anime yet: an episode classification requires the parent's anime, so only the episodes rows
// are written and the phase 2 sync creates their animes once the work is synced.
//
// [Ja] TestCreateEpisodesUsecase_Execute_SkipsAnimeForUnmappedWork は、まだ anime を持たない
// 親作品を検証する。エピソードの分類は親の anime を要するため episodes の行だけを書き、その
// anime は作品が同期された後にフェーズ 2 の同期が作る。
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
		t.Fatalf("Execute() error = %v", err)
	}
	if len(output.EpisodeIDs) != 1 {
		t.Fatalf("len(EpisodeIDs) = %d, want 1", len(output.EpisodeIDs))
	}

	created := readCreatedEpisode(t, db, output.EpisodeIDs[0])
	if created.animeID.Valid {
		t.Errorf("episodes.anime_id = %+v, want NULL (親作品が未マッピング)", created.animeID)
	}
	// The work has no episodes yet, so the numbering starts at the first step and there is no
	// preceding episode to name.
	//
	// [Ja] 作品はまだエピソードを持たないため、採番は最初のステップから始まり、直前の
	// エピソードとして名指しする相手もいない。
	if created.sortNumber != 100 {
		t.Errorf("sort_number = %d, want 100", created.sortNumber)
	}
	if created.prevEpisodeID.Valid {
		t.Errorf("prev_episode_id = %+v, want NULL", created.prevEpisodeID)
	}
}

// TestCreateEpisodesUsecase_Execute_ProducesSyncConsistentMapping is the invariant that
// justifies reusing the sync mapping helpers: a sync run right after a create must detect no
// diff (Unchanged), proving create and sync derive the same anime / classification from the
// episode and the create path never inflates the diff metric.
//
// [Ja] TestCreateEpisodesUsecase_Execute_ProducesSyncConsistentMapping は同期の写像ヘルパー
// 再利用を正当化する不変条件。作成直後の同期実行は差分なし (Unchanged) を検出しなければ
// ならず、create と同期が同じ anime / 分類をエピソードから導出していること、create 経路が
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
		Rows:   "#1,1,はじまり\n,,タイトルだけの話\n#3,,数値話数なし",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	syncUC := newSyncEpisodesUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: output.EpisodeIDs})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Processed != 3 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 3 {
		t.Fatalf("sync result = %+v, want {Processed:3 Created:0 Updated:0 Unchanged:3}", result)
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
			t.Fatalf("concurrent Execute() error = %v", err)
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

		wantSortNumber := int32(index+1) * episodeSortNumberStep // #nosec G115 -- index is bounded by createCount.
		if sortNumber != wantSortNumber {
			t.Errorf("%d 件目の sort_number = %d, want %d", index+1, sortNumber, wantSortNumber)
		}
		if index == 0 && prevEpisodeID.Valid {
			t.Errorf("先頭の prev_episode_id = %+v, want NULL", prevEpisodeID)
		}
		if index > 0 && (!prevEpisodeID.Valid || prevEpisodeID.Int64 != previousID) {
			t.Errorf("%d 件目の prev_episode_id = %+v, want %d", index+1, prevEpisodeID, previousID)
		}
		previousID = id
		index++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("作成されたエピソードの走査に失敗: %v", err)
	}
	if index != createCount {
		t.Errorf("作成されたエピソード = %d 件, want %d", index, createCount)
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
		t.Fatalf("予定話数の更新に失敗: %v", err)
	}
	insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))

	input := CreateEpisodesInput{WorkID: workID, User: editor, Rows: "#2,2,つづき"}
	_, err := uc.Execute(context.Background(), input)
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("editor Execute() error = %v, want *model.ValidationError", err)
	}
	// The restriction is a condition of the work, not of the submitted lines, so it is
	// reported for the form as a whole.
	//
	// [Ja] 制限は送信された行ではなく作品の状態に由来するため、フォーム全体に対して報告する。
	if len(ve.Global) != 1 || len(ve.GetFieldErrors("rows")) != 0 {
		t.Errorf("editor のエラー = Global:%v Fields:%v, want グローバルに 1 件のみ", ve.Global, ve.GetFieldErrors("rows"))
	}

	input.User = admin
	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("admin Execute() error = %v", err)
	}
	if len(output.EpisodeIDs) != 1 {
		t.Errorf("admin len(EpisodeIDs) = %d, want 1", len(output.EpisodeIDs))
	}
}

// TestCreateEpisodesUsecase_Execute_ValidationError covers a submit with a bad line: the whole
// submit fails, so not even the valid lines reach the database.
//
// [Ja] TestCreateEpisodesUsecase_Execute_ValidationError は不正な行を含む送信を検証する。送信
// 全体が失敗するため、正常な行も DB に届かない。
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
		t.Errorf("output = %+v, want nil", output)
	}
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("err = %v, want *model.ValidationError", err)
	}
	if len(ve.GetFieldErrors("rows")) != 1 {
		t.Errorf("rows のエラー = %v, want 1 件", ve.GetFieldErrors("rows"))
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM episodes WHERE work_id = $1`, int64(workID)).Scan(&count); err != nil {
		t.Fatalf("エピソード件数の取得に失敗: %v", err)
	}
	if count != 0 {
		t.Errorf("作成されたエピソード = %d 件, want 0 (送信全体が失敗するため)", count)
	}
}

// TestCreateEpisodesUsecase_Execute_WorkNotFound covers submits against a work the form may
// not be shown for: an unknown id and a deleted work both fail as not found, before the lines
// are even parsed.
//
// [Ja] TestCreateEpisodesUsecase_Execute_WorkNotFound はフォームを出せない作品への送信を検証
// する。存在しない id と削除済みの作品はいずれも、行のパースにも進まず未存在として失敗する。
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
				t.Fatalf("err = %v, want *model.AppError", err)
			}
			if ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("ae.Code = %v, want %v", ae.Code, model.AppErrCodeResourceNotFound)
			}
		})
	}
}

// TestCreateEpisodesUsecase_Execute_RequiresCommitter covers unauthenticated and ordinary-user
// submits. Episode creation belongs to committers, and keeping that check in the usecase stops
// another entry point from bypassing the web middleware.
//
// [Ja] TestCreateEpisodesUsecase_Execute_RequiresCommitter は未認証と一般ユーザーの送信を検証
// する。エピソード作成は committer に限られ、この確認を UseCase に置くことで別の入口が web
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
				t.Fatalf("err = %v, want *model.AppError", err)
			}
			if ae.Code != model.AppErrCodeForbidden {
				t.Errorf("ae.Code = %v, want %v", ae.Code, model.AppErrCodeForbidden)
			}
		})
	}
}

// TestNextSortAnchor covers which episode the next created row names as its preceding one.
// The numbering starts from the work's episode count rather than from its greatest
// sort_number, so a work whose existing episodes were spaced further apart keeps naming that
// episode, as the Rails callback does.
//
// [Ja] TestNextSortAnchor は、次に作る行がどのエピソードを直前のエピソードとして名指しするか
// を検証する。採番の起点が作品の最大 sort_number ではなくエピソード数であるため、既存の
// エピソードがより広い間隔で並んでいる作品ではそのエピソードが名指しされ続ける (Rails の
// コールバックと同じ)。
func TestNextSortAnchor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		previous   *repository.DBEpisodeSortAnchor
		episodeID  model.EpisodeID
		sortNumber int32
		want       repository.DBEpisodeSortAnchor
	}{
		{
			name:       "直前のエピソードが無いとき",
			previous:   nil,
			episodeID:  7,
			sortNumber: 100,
			want:       repository.DBEpisodeSortAnchor{ID: 7, SortNumber: 100},
		},
		{
			name:       "新規のほうが後ろに並ぶとき",
			previous:   &repository.DBEpisodeSortAnchor{ID: 3, SortNumber: 100},
			episodeID:  7,
			sortNumber: 200,
			want:       repository.DBEpisodeSortAnchor{ID: 7, SortNumber: 200},
		},
		{
			name:       "並び順が同値のときは新規を採る",
			previous:   &repository.DBEpisodeSortAnchor{ID: 3, SortNumber: 200},
			episodeID:  7,
			sortNumber: 200,
			want:       repository.DBEpisodeSortAnchor{ID: 7, SortNumber: 200},
		},
		{
			name:       "既存のほうが後ろに並ぶとき",
			previous:   &repository.DBEpisodeSortAnchor{ID: 3, SortNumber: 5000},
			episodeID:  7,
			sortNumber: 200,
			want:       repository.DBEpisodeSortAnchor{ID: 3, SortNumber: 5000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := nextSortAnchor(tt.previous, tt.episodeID, tt.sortNumber)
			if *got != tt.want {
				t.Errorf("nextSortAnchor() = %+v, want %+v", *got, tt.want)
			}
		})
	}
}
