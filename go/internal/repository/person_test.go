package repository_test

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

// 他のテストがコミットした人物もSetupTxのトランザクションから見えるため、一覧の検証は
// 自身が作成した人物の有無と相対的な順序に絞る。

// personIndexesは一覧の中での人物IDの位置を返す。一覧に無いIDは含めない。
func personIndexes(people []*model.Person) map[model.PersonID]int {
	indexes := make(map[model.PersonID]int, len(people))
	for i, p := range people {
		indexes[p.ID] = i
	}
	return indexes
}

// TestPersonRepository_ListForDBは、公開中とアーカイブの人物を状態付きでIDの降順に返し、
// 削除と掲載停止の人物を返さないことを検証する。
func TestPersonRepository_ListForDB(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))

	publishedID := testutil.NewPersonBuilder(t, tx).
		WithNameKana("こうかい").
		WithNameEn("Published").
		Build()
	archivedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
	delistedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).WithUnpublishedAt(time.Now()).Build()
	deletedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).WithDeletedAt(time.Now()).Build()
	newestID := testutil.NewPersonBuilder(t, tx).Build()

	people, err := repo.ListForDB(context.Background(), repository.DBPersonListParams{
		Page:    1,
		PerPage: 100,
	})
	if err != nil {
		t.Fatalf("ListForDB()のエラー = %v", err)
	}

	indexes := personIndexes(people)
	wantPresent := []model.PersonID{newestID, archivedID, publishedID}
	for i, id := range wantPresent {
		index, ok := indexes[id]
		if !ok {
			t.Fatalf("人物 %d が一覧に含まれていません", id)
		}
		// wantPresentはIDの降順に並べてあるため、一覧でも前の要素より後ろに来る。
		if i > 0 && index <= indexes[wantPresent[i-1]] {
			t.Errorf("人物 %d が人物 %d より前に並んでいます", id, wantPresent[i-1])
		}
	}
	for _, id := range []model.PersonID{delistedID, deletedID} {
		if _, ok := indexes[id]; ok {
			t.Errorf("人物 %d が一覧に含まれてはいけません", id)
		}
	}

	published := people[indexes[publishedID]]
	if published.NameKana != "こうかい" || published.NameEn != "Published" {
		t.Errorf("NameKana / NameEn = %q / %q、期待値 = こうかい / Published", published.NameKana, published.NameEn)
	}
	if published.Status != model.PersonStatusPublished {
		t.Errorf("公開中の人物のStatus = %q", published.Status)
	}
	if archived := people[indexes[archivedID]]; archived.Status != model.PersonStatusArchived {
		t.Errorf("アーカイブの人物のStatus = %q", archived.Status)
	}
}

// TestPersonRepository_ListForDB_Paginationは、PerPageとPageでオフセットを取ることを検証する。
func TestPersonRepository_ListForDB_Pagination(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))

	for range 3 {
		testutil.NewPersonBuilder(t, tx).Build()
	}

	first, err := repo.ListForDB(context.Background(), repository.DBPersonListParams{Page: 1, PerPage: 1})
	if err != nil {
		t.Fatalf("ListForDB()のエラー = %v", err)
	}
	second, err := repo.ListForDB(context.Background(), repository.DBPersonListParams{Page: 2, PerPage: 1})
	if err != nil {
		t.Fatalf("ListForDB()のエラー = %v", err)
	}

	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("len(first) = %d、len(second) = %d、期待値 = 1、1", len(first), len(second))
	}
	// 2ページ目は1ページ目の続きのため、IDの降順で1ページ目より小さいIDから始まる。
	if first[0].ID <= second[0].ID {
		t.Errorf("1ページ目の人物 %d が2ページ目の人物 %d より新しくありません", first[0].ID, second[0].ID)
	}
}

// TestPersonRepository_CountForDBは、総件数がListForDBと同じ条件で数えられることを検証する。
func TestPersonRepository_CountForDB(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	count := func() int64 {
		t.Helper()
		count, err := repo.CountForDB(ctx)
		if err != nil {
			t.Fatalf("CountForDB()のエラー = %v", err)
		}
		return count
	}
	before := count()

	testutil.NewPersonBuilder(t, tx).Build()
	testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
	testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()

	// 同じトランザクションの中で前後の差を取ることで、他のテストがコミットした人物の
	// 影響を受けずに増分だけを検証する。
	if got := count() - before; got != 2 {
		t.Errorf("増分 = %d、期待値 = 2", got)
	}
}

// TestPersonRepository_ListTakenNamesは、一意インデックスと同じく削除した人物を除き、
// アーカイブ・掲載停止の人物の名前も使用中として返すことを検証する。
func TestPersonRepository_ListTakenNames(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))

	publishedName := "使用中の人物-" + t.Name()
	archivedName := "アーカイブの人物-" + t.Name()
	delistedName := "掲載停止の人物-" + t.Name()
	deletedName := "削除済みの人物-" + t.Name()
	unusedName := "未登録の人物-" + t.Name()
	testutil.NewPersonBuilder(t, tx).WithName(publishedName).Build()
	testutil.NewPersonBuilder(t, tx).WithName(archivedName).WithStatus(model.PersonStatusArchived).Build()
	testutil.NewPersonBuilder(t, tx).WithName(delistedName).WithStatus(model.PersonStatusDelisted).Build()
	testutil.NewPersonBuilder(t, tx).WithName(deletedName).WithStatus(model.PersonStatusDeleted).Build()

	taken, err := repo.ListTakenNames(context.Background(), []string{publishedName, archivedName, delistedName, deletedName, unusedName})
	if err != nil {
		t.Fatalf("ListTakenNamesでエラー: %v", err)
	}

	got := make(map[string]bool, len(taken))
	for _, name := range taken {
		got[name] = true
	}
	for _, name := range []string{publishedName, archivedName, delistedName} {
		if !got[name] {
			t.Errorf("%qが使用中として返されていません", name)
		}
	}
	for _, name := range []string{deletedName, unusedName} {
		if got[name] {
			t.Errorf("%qが使用中として返されました", name)
		}
	}
	if len(taken) != 3 {
		t.Errorf("使用中の名前 = %d件、期待値 = 3", len(taken))
	}
}

// TestPersonRepository_ListTakenNames_Emptyは、照合する名前が無いときにクエリを
// 発行せず空を返すことを検証する。
func TestPersonRepository_ListTakenNames_Empty(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))

	taken, err := repo.ListTakenNames(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTakenNamesでエラー: %v", err)
	}
	if len(taken) != 0 {
		t.Errorf("使用中の名前 = %v、期待値 = 空", taken)
	}
}

// TestPersonRepository_Createは、人物を公開の状態で作成し、Rails版の
// save_and_create_activity! と同じ形でpeople.createの編集履歴を記録することを検証する。
func TestPersonRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()

	name := "作成する人物-" + t.Name()
	id, err := repo.Create(context.Background(), repository.CreatePersonParams{
		Name:     name,
		NameKana: "さくせいするじんぶつ",
		UserID:   userID,
	})
	if err != nil {
		t.Fatalf("Createでエラー: %v", err)
	}

	var (
		gotName, gotNameKana, gotAasmState string
		gotStatus                          string
		unpublished, deleted               bool
	)
	if err := tx.QueryRow(`
		SELECT name, name_kana, aasm_state, status, unpublished_at IS NOT NULL, deleted_at IS NOT NULL
		FROM people WHERE id = $1
	`, int64(id)).Scan(&gotName, &gotNameKana, &gotAasmState, &gotStatus, &unpublished, &deleted); err != nil {
		t.Fatalf("作成した人物の取得に失敗: %v", err)
	}
	if gotName != name || gotNameKana != "さくせいするじんぶつ" {
		t.Errorf("name / name_kana = %q / %q、期待値 = %q / %q", gotName, gotNameKana, name, "さくせいするじんぶつ")
	}
	if gotStatus != "published" || gotAasmState != "published" || unpublished || deleted {
		t.Errorf("作成した人物が公開の状態ではありません (status = %q, aasm_state = %q, unpublished = %v, deleted = %v)", gotStatus, gotAasmState, unpublished, deleted)
	}

	var (
		activityUserID, rootResourceID                   int64
		trackableType, action, rootResourceType, newName string
		newID                                            int64
		hasOld                                           bool
	)
	if err := tx.QueryRow(`
		SELECT user_id, trackable_type, action, root_resource_id, root_resource_type,
			parameters->'new'->>'name', (parameters->'new'->>'id')::bigint, parameters->'old' IS NOT NULL
		FROM db_activities WHERE trackable_type = 'Person' AND trackable_id = $1
	`, int64(id)).Scan(&activityUserID, &trackableType, &action, &rootResourceID, &rootResourceType, &newName, &newID, &hasOld); err != nil {
		t.Fatalf("編集履歴の取得に失敗: %v", err)
	}
	if activityUserID != int64(userID) {
		t.Errorf("user_id = %d、期待値 = %d", activityUserID, int64(userID))
	}
	if trackableType != "Person" || action != "people.create" {
		t.Errorf("trackable_type / action = %q / %q、期待値 = Person / people.create", trackableType, action)
	}
	if rootResourceType != "Person" || rootResourceID != int64(id) {
		t.Errorf("root_resource = %s %d、期待値 = Person %d", rootResourceType, rootResourceID, int64(id))
	}
	if newName != name || newID != int64(id) {
		t.Errorf("parameters.new = {id: %d, name: %q}、期待値 = {id: %d, name: %q}", newID, newName, int64(id), name)
	}
	if hasOld {
		t.Error("作成の編集履歴にparameters.oldが含まれています")
	}
}

// personTestVersionは版の照合を検証するテストが人物に設定する過去のupdated_at。マイクロ秒
// まで値を持たせ、フォームとの往復で精度が落ちないことも確かめられるようにする。
var personTestVersion = time.Date(2024, 1, 2, 3, 4, 5, 123456000, time.UTC)

// fullUpdatePersonParamsはすべての項目に値を入れた更新パラメータを返す。
func fullUpdatePersonParams(id model.PersonID, userID model.UserID, prefectureID model.PrefectureID, name string) repository.UpdatePersonParams {
	nickname := "ニックネーム"
	gender := model.PersonGenderFemale
	bloodType := model.PersonBloodTypeAB
	birthday := time.Date(1990, 4, 5, 0, 0, 0, 0, time.UTC)
	height := int32(158)
	url := "https://example.com/person"
	wikipediaURL := "https://ja.wikipedia.org/wiki/person"
	twitterUsername := "person_x"

	return repository.UpdatePersonParams{
		ID:                id,
		Name:              name,
		NameKana:          "へんしゅうごのよみ",
		NameEn:            "Edited Person",
		Nickname:          &nickname,
		NicknameEn:        "Nickname",
		Gender:            &gender,
		BloodType:         &bloodType,
		PrefectureID:      &prefectureID,
		Birthday:          &birthday,
		Height:            &height,
		URL:               &url,
		URLEn:             "https://example.com/en/person",
		WikipediaURL:      &wikipediaURL,
		WikipediaURLEn:    "https://en.wikipedia.org/wiki/person",
		TwitterUsername:   &twitterUsername,
		TwitterUsernameEn: "person_x_en",
		Version:           personTestVersion,
		UserID:            userID,
	}
}

// TestPersonRepository_GetForEditByIDは、編集フォームが扱う項目と版を読み、アーカイブの人物は
// 返し、掲載停止・削除の人物を除くことを検証する。
func TestPersonRepository_GetForEditByID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()
	prefectureID := testutil.CreateTestPrefecture(t, tx, "東京都")

	ctx := context.Background()
	id := testutil.NewPersonBuilder(t, tx).WithUpdatedAt(personTestVersion).Build()
	params := fullUpdatePersonParams(id, userID, prefectureID, "編集用の人物-"+t.Name())
	if _, err := repo.Update(ctx, params); err != nil {
		t.Fatalf("Updateでエラー: %v", err)
	}

	person, err := repo.GetForEditByID(ctx, id)
	if err != nil {
		t.Fatalf("GetForEditByIDでエラー: %v", err)
	}
	if person == nil {
		t.Fatal("人物が見つかりません")
	}
	if person.Name != params.Name || person.NameKana != params.NameKana || person.NameEn != params.NameEn {
		t.Errorf("名前 = %q / %q / %q", person.Name, person.NameKana, person.NameEn)
	}
	if person.Nickname == nil || *person.Nickname != *params.Nickname || person.NicknameEn != params.NicknameEn {
		t.Errorf("ニックネーム = %v / %q", person.Nickname, person.NicknameEn)
	}
	if person.Gender == nil || *person.Gender != model.PersonGenderFemale {
		t.Errorf("Gender = %v、期待値 = female", person.Gender)
	}
	if person.BloodType == nil || *person.BloodType != model.PersonBloodTypeAB {
		t.Errorf("BloodType = %v、期待値 = ab", person.BloodType)
	}
	if person.PrefectureID == nil || *person.PrefectureID != prefectureID {
		t.Errorf("PrefectureID = %v、期待値 = %d", person.PrefectureID, prefectureID)
	}
	if person.Birthday == nil || person.Birthday.Format("2006-01-02") != "1990-04-05" {
		t.Errorf("Birthday = %v、期待値 = 1990-04-05", person.Birthday)
	}
	if person.Height == nil || *person.Height != 158 {
		t.Errorf("Height = %v、期待値 = 158", person.Height)
	}
	if person.URL == nil || *person.URL != *params.URL || person.URLEn != params.URLEn {
		t.Errorf("URL = %v / %q", person.URL, person.URLEn)
	}
	if person.WikipediaURL == nil || *person.WikipediaURL != *params.WikipediaURL || person.WikipediaURLEn != params.WikipediaURLEn {
		t.Errorf("WikipediaURL = %v / %q", person.WikipediaURL, person.WikipediaURLEn)
	}
	if person.TwitterUsername == nil || *person.TwitterUsername != *params.TwitterUsername || person.TwitterUsernameEn != params.TwitterUsernameEn {
		t.Errorf("TwitterUsername = %v / %q", person.TwitterUsername, person.TwitterUsernameEn)
	}

	if person.Status != model.PersonStatusPublished {
		t.Errorf("Status = %q、期待値 = published", person.Status)
	}

	t.Run("アーカイブの人物は返す", func(t *testing.T) {
		archivedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
		person, err := repo.GetForEditByID(ctx, archivedID)
		if err != nil {
			t.Fatalf("GetForEditByIDでエラー: %v", err)
		}
		if person == nil || person.Status != model.PersonStatusArchived {
			t.Errorf("アーカイブの人物 = %v、期待値 = アーカイブの人物", person)
		}
	})

	for _, status := range []model.PersonStatus{model.PersonStatusDelisted, model.PersonStatusDeleted} {
		t.Run(status.String()+"の人物はnil", func(t *testing.T) {
			hiddenID := testutil.NewPersonBuilder(t, tx).WithStatus(status).Build()
			person, err := repo.GetForEditByID(ctx, hiddenID)
			if err != nil {
				t.Fatalf("GetForEditByIDでエラー: %v", err)
			}
			if person != nil {
				t.Errorf("%sの人物 = %v、期待値 = nil", status, person)
			}
		})
	}
}

// TestPersonRepository_GetForEditByID_VersionRoundTripは、timestamp without time zoneの
// updated_atを、精度とタイムゾーンを落とさずに版として読み、その版で更新できることを検証する。
// タイムゾーンを持たないカラムを読み書きするため、ずれると古くない送信が競合として却下される。
func TestPersonRepository_GetForEditByID_VersionRoundTrip(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()

	ctx := context.Background()
	id := testutil.NewPersonBuilder(t, tx).WithUpdatedAt(personTestVersion).Build()

	person, err := repo.GetForEditByID(ctx, id)
	if err != nil {
		t.Fatalf("GetForEditByIDでエラー: %v", err)
	}
	if !person.UpdatedAt.Equal(personTestVersion) {
		t.Fatalf("UpdatedAt = %v、期待値 = %v", person.UpdatedAt, personTestVersion)
	}

	// フォームはUTC以外のタイムゾーンで表した時刻を受け取ることもある。同じ時点を表す版
	// なら一致とみなされることを、UTC+9で表した版で確かめる。
	version := person.UpdatedAt.In(time.FixedZone("JST", 9*60*60))
	params := repository.UpdatePersonParams{
		ID:      id,
		Name:    "版の往復の人物-" + t.Name(),
		Version: version,
		UserID:  userID,
	}
	updated, err := repo.Update(ctx, params)
	if err != nil {
		t.Fatalf("Updateでエラー: %v", err)
	}
	if !updated {
		t.Error("読み取った版での更新が競合として却下されました")
	}
}

// TestPersonRepository_Updateは、版が一致する送信だけを適用し、内容が変わったときに
// Rails版と同じ形でpeople.updateの編集履歴を記録することを検証する。
func TestPersonRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()
	prefectureID := testutil.CreateTestPrefecture(t, tx, "大阪府")

	ctx := context.Background()
	oldName := "編集前の人物-" + t.Name()
	newName := "編集後の人物-" + t.Name()
	id := testutil.NewPersonBuilder(t, tx).WithName(oldName).WithUpdatedAt(personTestVersion).Build()

	updated, err := repo.Update(ctx, fullUpdatePersonParams(id, userID, prefectureID, newName))
	if err != nil {
		t.Fatalf("Updateでエラー: %v", err)
	}
	if !updated {
		t.Fatal("版が一致する更新が適用されませんでした")
	}

	var (
		activityUserID, rootResourceID                   int64
		action, rootResourceType, gotOldName, gotNewName string
		gotNewGender                                     string
	)
	if err := tx.QueryRow(`
		SELECT user_id, action, root_resource_id, root_resource_type,
			parameters->'old'->>'name', parameters->'new'->>'name', parameters->'new'->>'gender'
		FROM db_activities WHERE trackable_type = 'Person' AND trackable_id = $1
	`, int64(id)).Scan(&activityUserID, &action, &rootResourceID, &rootResourceType, &gotOldName, &gotNewName, &gotNewGender); err != nil {
		t.Fatalf("編集履歴の取得に失敗: %v", err)
	}
	if activityUserID != int64(userID) || action != "people.update" {
		t.Errorf("user_id / action = %d / %q、期待値 = %d / people.update", activityUserID, action, int64(userID))
	}
	if rootResourceType != "Person" || rootResourceID != int64(id) {
		t.Errorf("root_resource = %s %d、期待値 = Person %d", rootResourceType, rootResourceID, int64(id))
	}
	if gotOldName != oldName || gotNewName != newName || gotNewGender != "female" {
		t.Errorf("parameters = {old.name: %q, new.name: %q, new.gender: %q}", gotOldName, gotNewName, gotNewGender)
	}

	t.Run("古い版からの更新はfalseで、何も書かない", func(t *testing.T) {
		params := fullUpdatePersonParams(id, userID, prefectureID, "古い版からの人物-"+t.Name())
		updated, err := repo.Update(ctx, params)
		if err != nil {
			t.Fatalf("Updateでエラー: %v", err)
		}
		if updated {
			t.Error("古い版からの更新が適用されました")
		}

		var name string
		var activities int
		if err := tx.QueryRow(`
			SELECT p.name, (SELECT COUNT(*) FROM db_activities a WHERE a.trackable_type = 'Person' AND a.trackable_id = p.id)
			FROM people p WHERE p.id = $1
		`, int64(id)).Scan(&name, &activities); err != nil {
			t.Fatalf("人物の取得に失敗: %v", err)
		}
		if name != newName || activities != 1 {
			t.Errorf("name / 編集履歴 = %q / %d件、期待値 = %q / 1件", name, activities, newName)
		}
	})

	t.Run("アーカイブの人物は更新する", func(t *testing.T) {
		archivedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).WithUpdatedAt(personTestVersion).Build()
		params := fullUpdatePersonParams(archivedID, userID, prefectureID, "アーカイブの人物-"+t.Name())
		updated, err := repo.Update(ctx, params)
		if err != nil {
			t.Fatalf("Updateでエラー: %v", err)
		}
		if !updated {
			t.Error("アーカイブの人物が更新されませんでした")
		}
	})

	for _, status := range []model.PersonStatus{model.PersonStatusDelisted, model.PersonStatusDeleted} {
		t.Run(status.String()+"の人物はfalse", func(t *testing.T) {
			hiddenID := testutil.NewPersonBuilder(t, tx).WithStatus(status).WithUpdatedAt(personTestVersion).Build()
			params := fullUpdatePersonParams(hiddenID, userID, prefectureID, status.String()+"の人物-"+t.Name())
			updated, err := repo.Update(ctx, params)
			if err != nil {
				t.Fatalf("Updateでエラー: %v", err)
			}
			if updated {
				t.Errorf("%sの人物が更新されました", status)
			}
		})
	}
}

// TestPersonRepository_Update_NoChangeは、内容の変わらない送信では編集履歴を記録しない
// ことを検証する。Rails版が空欄を空文字列で保存した行に空欄を送っても (Go版はNULLで書く)、
// 変更とはみなさない。
func TestPersonRepository_Update_NoChange(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()

	ctx := context.Background()
	name := "変更の無い人物-" + t.Name()
	id := testutil.NewPersonBuilder(t, tx).
		WithName(name).
		WithNameKana("へんこうのないじんぶつ").
		WithNickname("").
		WithUpdatedAt(personTestVersion).
		Build()

	updated, err := repo.Update(ctx, repository.UpdatePersonParams{
		ID:       id,
		Name:     name,
		NameKana: "へんこうのないじんぶつ",
		Version:  personTestVersion,
		UserID:   userID,
	})
	if err != nil {
		t.Fatalf("Updateでエラー: %v", err)
	}
	if !updated {
		t.Fatal("版が一致する更新が適用されませんでした")
	}

	var activities int
	if err := tx.QueryRow(`
		SELECT COUNT(*) FROM db_activities WHERE trackable_type = 'Person' AND trackable_id = $1
	`, int64(id)).Scan(&activities); err != nil {
		t.Fatalf("編集履歴の取得に失敗: %v", err)
	}
	if activities != 0 {
		t.Errorf("編集履歴 = %d件、期待値 = 0件", activities)
	}
}

// TestPersonRepository_ExistsByNameExcludingIDは、編集中の人物と削除した人物を除き、
// アーカイブ・掲載停止の人物も含めて名前の重複を判定することを検証する。
func TestPersonRepository_ExistsByNameExcludingID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))

	ctx := context.Background()
	selfName := "自分の名前-" + t.Name()
	archivedName := "アーカイブの人物の名前-" + t.Name()
	delistedName := "掲載停止の人物の名前-" + t.Name()
	deletedName := "削除済みの人物の名前-" + t.Name()
	selfID := testutil.NewPersonBuilder(t, tx).WithName(selfName).Build()
	testutil.NewPersonBuilder(t, tx).WithName(archivedName).WithStatus(model.PersonStatusArchived).Build()
	testutil.NewPersonBuilder(t, tx).WithName(delistedName).WithStatus(model.PersonStatusDelisted).Build()
	testutil.NewPersonBuilder(t, tx).WithName(deletedName).WithStatus(model.PersonStatusDeleted).Build()

	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{name: "自分自身の名前は重複しない", arg: selfName, want: false},
		{name: "アーカイブの人物の名前は重複する", arg: archivedName, want: true},
		{name: "掲載停止の人物の名前は重複する", arg: delistedName, want: true},
		{name: "削除済みの人物の名前は重複しない", arg: deletedName, want: false},
		{name: "未使用の名前は重複しない", arg: "未使用の名前-" + t.Name(), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ExistsByNameExcludingID(ctx, tt.arg, selfID)
			if err != nil {
				t.Fatalf("ExistsByNameExcludingIDでエラー: %v", err)
			}
			if got != tt.want {
				t.Errorf("ExistsByNameExcludingID(%q) = %v、期待値 = %v", tt.arg, got, tt.want)
			}
		})
	}
}

// personStateは人物の状態 (status) と、Rails版への橋渡しの状態カラム (unpublished_at /
// deleted_at) が立っているかを表す。
type personState struct {
	status      model.PersonStatus
	unpublished bool
	deleted     bool
}

// readPersonStateは人物の状態を読む。
func readPersonState(t *testing.T, tx *sql.Tx, id model.PersonID) personState {
	t.Helper()

	var state personState
	if err := tx.QueryRow(`
		SELECT status, unpublished_at IS NOT NULL, deleted_at IS NOT NULL FROM people WHERE id = $1
	`, int64(id)).Scan(&state.status, &state.unpublished, &state.deleted); err != nil {
		t.Fatalf("人物 %d の状態の取得に失敗: %v", id, err)
	}
	return state
}

// TestPersonRepository_GetForStateChangeByIDは、確認画面が使う名前と状態を、削除済みの人物も
// 含めて返し、存在しない人物ではnilを返すことを検証する。
func TestPersonRepository_GetForStateChangeByID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	name := "状態変更の人物-" + t.Name()
	id := testutil.NewPersonBuilder(t, tx).WithName(name).WithStatus(model.PersonStatusDeleted).Build()

	person, err := repo.GetForStateChangeByID(ctx, id)
	if err != nil {
		t.Fatalf("GetForStateChangeByIDでエラー: %v", err)
	}
	if person == nil || person.ID != id || person.Name != name {
		t.Fatalf("人物 = %+v、期待値 = ID %d の %q", person, id, name)
	}
	if person.Status != model.PersonStatusDeleted {
		t.Errorf("Status = %q、期待値 = deleted", person.Status)
	}

	missing, err := repo.GetForStateChangeByID(ctx, model.PersonID(999999999))
	if err != nil {
		t.Fatalf("GetForStateChangeByIDでエラー: %v", err)
	}
	if missing != nil {
		t.Errorf("存在しない人物 = %+v、期待値 = nil", missing)
	}
}

// TestPersonRepository_StateChangesは、アーカイブ・アーカイブ解除・削除が、その操作を受け付ける
// 状態の人物だけを書き換え、それ以外の状態 (と存在しない人物) ではfalseを返して何も書かないことを
// 検証する。状態の確認と書き込みを1文で行うため、古い確認画面からの送信が状態を打ち直さない。
// アーカイブは閲覧できる状態のため橋渡しのunpublished_atを書かず、削除は橋渡しのdeleted_atを書く。
func TestPersonRepository_StateChanges(t *testing.T) {
	t.Parallel()

	type operation func(repo *repository.PersonRepository, id model.PersonID) (bool, error)
	archive := func(repo *repository.PersonRepository, id model.PersonID) (bool, error) {
		return repo.Archive(context.Background(), id)
	}
	unarchive := func(repo *repository.PersonRepository, id model.PersonID) (bool, error) {
		return repo.Unarchive(context.Background(), id)
	}
	deletePerson := func(repo *repository.PersonRepository, id model.PersonID) (bool, error) {
		return repo.Delete(context.Background(), id)
	}

	published := model.PersonStatusPublished
	archived := model.PersonStatusArchived
	delisted := model.PersonStatusDelisted
	deleted := model.PersonStatusDeleted

	tests := []struct {
		name        string
		op          operation
		status      model.PersonStatus
		wantWritten bool
		wantState   personState
	}{
		{name: "公開中の人物をアーカイブする", op: archive, status: published, wantWritten: true, wantState: personState{status: archived}},
		{name: "アーカイブの人物はアーカイブしない", op: archive, status: archived, wantState: personState{status: archived}},
		{name: "掲載停止の人物はアーカイブしない", op: archive, status: delisted, wantState: personState{status: delisted}},
		{name: "削除済みの人物はアーカイブしない", op: archive, status: deleted, wantState: personState{status: deleted}},
		{name: "アーカイブの人物のアーカイブを解除する", op: unarchive, status: archived, wantWritten: true, wantState: personState{status: published}},
		{name: "公開中の人物はアーカイブ解除しない", op: unarchive, status: published, wantState: personState{status: published}},
		{name: "掲載停止の人物はアーカイブ解除しない", op: unarchive, status: delisted, wantState: personState{status: delisted}},
		{name: "削除済みの人物はアーカイブ解除しない", op: unarchive, status: deleted, wantState: personState{status: deleted}},
		{name: "公開中の人物を削除する", op: deletePerson, status: published, wantWritten: true, wantState: personState{status: deleted, deleted: true}},
		{name: "アーカイブの人物を削除する", op: deletePerson, status: archived, wantWritten: true, wantState: personState{status: deleted, deleted: true}},
		{name: "掲載停止の人物は削除しない", op: deletePerson, status: delisted, wantState: personState{status: delisted}},
		{name: "削除済みの人物は削除しない", op: deletePerson, status: deleted, wantState: personState{status: deleted}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			repo := repository.NewPersonRepository(query.New(db).WithTx(tx))

			id := testutil.NewPersonBuilder(t, tx).WithStatus(tt.status).WithUpdatedAt(personTestVersion).Build()

			written, err := tt.op(repo, id)
			if err != nil {
				t.Fatalf("状態の変更でエラー: %v", err)
			}
			if written != tt.wantWritten {
				t.Errorf("書き込み = %v、期待値 = %v", written, tt.wantWritten)
			}

			if got := readPersonState(t, tx, id); got != tt.wantState {
				t.Errorf("状態 = %+v、期待値 = %+v", got, tt.wantState)
			}

			var updatedAt time.Time
			if err := tx.QueryRow(`SELECT updated_at FROM people WHERE id = $1`, int64(id)).Scan(&updatedAt); err != nil {
				t.Fatalf("updated_atの取得に失敗: %v", err)
			}
			if advanced := !updatedAt.Equal(personTestVersion); advanced != tt.wantWritten {
				t.Errorf("updated_at = %v、書き込んだときだけ進むこと (書き込み = %v)", updatedAt, tt.wantWritten)
			}
		})
	}

	t.Run("存在しない人物", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewPersonRepository(query.New(db).WithTx(tx))

		for _, op := range []operation{archive, unarchive, deletePerson} {
			written, err := op(repo, model.PersonID(999999999))
			if err != nil {
				t.Fatalf("状態の変更でエラー: %v", err)
			}
			if written {
				t.Error("存在しない人物への書き込みがtrueを返しました")
			}
		}
	})
}

// TestPersonRepository_Delete_KeepsRecordは、削除が論理削除で、人物の行を残すことを
// 検証する (Rails版の物理削除は人物とそのキャスト・スタッフ・編集履歴まで消していた)。
func TestPersonRepository_Delete_KeepsRecord(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRepository(query.New(db).WithTx(tx))
	id := testutil.NewPersonBuilder(t, tx).Build()

	if _, err := repo.Delete(context.Background(), id); err != nil {
		t.Fatalf("Deleteでエラー: %v", err)
	}

	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM people WHERE id = $1)`, int64(id)).Scan(&exists); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	if !exists {
		t.Error("削除した人物の行が残っていません")
	}
}

// TestPersonRepository_Create_NameUniquenessは、名前の一意インデックスが削除した人物を対象から
// 外し、掲載停止の人物は対象に残すことを検証する。削除した人物と同じ名前は登録し直せ、掲載停止
// した人物と同じ名前は登録できない。
func TestPersonRepository_Create_NameUniqueness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  model.PersonStatus
		wantErr bool
	}{
		{name: "削除した人物と同じ名前は登録できる", status: model.PersonStatusDeleted, wantErr: false},
		{name: "掲載停止の人物と同じ名前は登録できない", status: model.PersonStatusDelisted, wantErr: true},
		{name: "アーカイブの人物と同じ名前は登録できない", status: model.PersonStatusArchived, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			repo := repository.NewPersonRepository(query.New(db).WithTx(tx))
			userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()

			name := "一意性の人物-" + t.Name()
			testutil.NewPersonBuilder(t, tx).WithName(name).WithStatus(tt.status).Build()

			_, err := repo.Create(context.Background(), repository.CreatePersonParams{Name: name, UserID: userID})
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("Create()のエラー = %v、エラーを期待するか = %v", err, tt.wantErr)
			}
		})
	}
}
