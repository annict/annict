package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// ErrPersonNameTakenは更新の直前に別の書き込みが同じ名前を取得したことを表す。
var ErrPersonNameTaken = errors.New("人物名が既に使われています")

// PersonRepositoryは人物 (people) のデータアクセスを担当します
type PersonRepository struct {
	queries *query.Queries
}

// NewPersonRepositoryはPersonRepositoryを作成します
func NewPersonRepository(queries *query.Queries) *PersonRepository {
	return &PersonRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *PersonRepository) WithTx(tx *sql.Tx) *PersonRepository {
	return &PersonRepository{queries: r.queries.WithTx(tx)}
}

// DBPersonListParamsはAnnict DB管理画面の人物一覧の取得条件を保持します。
type DBPersonListParams struct {
	Page    int32
	PerPage int32
}

// ListForDBはAnnict DB管理画面の人物一覧 (公開中とアーカイブの人物) を、IDの降順に
// 1ページ分取得します。
func (r *PersonRepository) ListForDB(ctx context.Context, params DBPersonListParams) ([]*model.Person, error) {
	// 乗算の前に幅を広げる。ページ番号がint32の範囲で大きいとint32同士の積が負に折り返し、
	// PostgreSQLがそのOFFSETを拒否するため (WorkRepository.ListForDBと同じ)。
	offset := int64(params.Page-1) * int64(params.PerPage)

	rows, err := r.queries.ListDBPeople(ctx, query.ListDBPeopleParams{
		PerPage:    params.PerPage,
		PageOffset: offset,
	})
	if err != nil {
		return nil, err
	}

	people := make([]*model.Person, len(rows))
	for i, row := range rows {
		people[i] = &model.Person{
			ID:       model.PersonID(row.ID),
			Name:     row.Name,
			NameKana: row.NameKana,
			NameEn:   row.NameEn,
			Status:   model.PersonStatus(row.Status),
		}
	}
	return people, nil
}

// CountForDBはAnnict DB管理画面の人物一覧の総件数を取得します。
func (r *PersonRepository) CountForDB(ctx context.Context) (int64, error) {
	return r.queries.CountDBPeople(ctx)
}

// ListTakenNamesはnamesのうち、すでに人物の名前として使われているものを返します。
// people.nameの一意インデックスと同じく、削除した人物の名前は含めません。
func (r *PersonRepository) ListTakenNames(ctx context.Context, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	return r.queries.ListTakenPersonNames(ctx, names)
}

// CreatePersonParamsは人物の作成に使うパラメータを保持します。
type CreatePersonParams struct {
	Name     string
	NameKana string
	// UserIDは作成を編集履歴 (db_activities) に帰属させる編集者です。
	UserID model.UserID
}

// Createは人物を作成し、同じ文で編集履歴へpeople.createを記録します。
func (r *PersonRepository) Create(ctx context.Context, params CreatePersonParams) (model.PersonID, error) {
	id, err := r.queries.CreatePerson(ctx, query.CreatePersonParams{
		Name:     params.Name,
		NameKana: params.NameKana,
		UserID:   int64(params.UserID),
	})
	if err != nil {
		return 0, err
	}

	return model.PersonID(id), nil
}

// GetForEditByIDはAnnict DB管理画面の編集フォームが扱う人物を取得します。編集できるのは
// 公開中とアーカイブの人物で、それ以外 (削除・掲載停止) の人物と存在しない人物ではnilを
// 返します。
func (r *PersonRepository) GetForEditByID(ctx context.Context, id model.PersonID) (*model.Person, error) {
	row, err := r.queries.GetPersonForEditByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	person := &model.Person{
		ID:                model.PersonID(row.ID),
		Name:              row.Name,
		NameKana:          row.NameKana,
		NameEn:            row.NameEn,
		Nickname:          stringPtrFromNull(row.Nickname),
		NicknameEn:        row.NicknameEn,
		URL:               stringPtrFromNull(row.Url),
		URLEn:             row.UrlEn,
		WikipediaURL:      stringPtrFromNull(row.WikipediaUrl),
		WikipediaURLEn:    row.WikipediaUrlEn,
		TwitterUsername:   stringPtrFromNull(row.TwitterUsername),
		TwitterUsernameEn: row.TwitterUsernameEn,
		Birthday:          timePtrFromNull(row.Birthday),
		Status:            model.PersonStatus(row.Status),
		UpdatedAt:         row.UpdatedAt,
	}
	if row.Gender.Valid {
		gender := model.PersonGender(row.Gender.String)
		person.Gender = &gender
	}
	if row.BloodType.Valid {
		bloodType := model.PersonBloodType(row.BloodType.String)
		person.BloodType = &bloodType
	}
	if row.PrefectureID.Valid {
		prefectureID := model.PrefectureID(row.PrefectureID.Int64)
		person.PrefectureID = &prefectureID
	}
	if row.Height.Valid {
		height := row.Height.Int32
		person.Height = &height
	}

	return person, nil
}

// ExistsByNameExcludingIDは、excludeIDの人物以外に、nameを名前に持つ人物がいるかを返します。
// people.nameの一意インデックスと同じく、削除した人物は含めません。
func (r *PersonRepository) ExistsByNameExcludingID(ctx context.Context, name string, excludeID model.PersonID) (bool, error) {
	return r.queries.ExistsPersonByNameExcludingID(ctx, query.ExistsPersonByNameExcludingIDParams{
		Name:      name,
		ExcludeID: int64(excludeID),
	})
}

// UpdatePersonParamsは人物の更新に使うパラメータを保持します。NULL許容のカラムは
// ポインタで持ち、nilをNULLとして書きます。
type UpdatePersonParams struct {
	ID                model.PersonID
	Name              string
	NameKana          string
	NameEn            string
	Nickname          *string
	NicknameEn        string
	Gender            *model.PersonGender
	BloodType         *model.PersonBloodType
	PrefectureID      *model.PrefectureID
	Birthday          *time.Time
	Height            *int32
	URL               *string
	URLEn             string
	WikipediaURL      *string
	WikipediaURLEn    string
	TwitterUsername   *string
	TwitterUsernameEn string
	// Versionは送信が前提とする版 (フォームを開いた時点のupdated_at)。
	Version time.Time
	// UserIDは更新を編集履歴 (db_activities) に帰属させる編集者です。
	UserID model.UserID
}

// Updateは人物編集の1回の送信を適用し、内容が変わった場合は同じ文で編集履歴へ
// people.updateを記録します。どの行も一致しなかった場合 (人物が編集できない状態になったか、
// フォームを開いてからupdated_atが進んだ場合) はfalseを返します。
func (r *PersonRepository) Update(ctx context.Context, params UpdatePersonParams) (bool, error) {
	queryParams := query.UpdateDBPersonParams{
		ID:                int64(params.ID),
		Name:              params.Name,
		NameKana:          params.NameKana,
		NameEn:            params.NameEn,
		Nickname:          nullStringFromPtr(params.Nickname),
		NicknameEn:        params.NicknameEn,
		Birthday:          nullTimeFromPtr(params.Birthday),
		Height:            nullInt32FromPtr(params.Height),
		Url:               nullStringFromPtr(params.URL),
		UrlEn:             params.URLEn,
		WikipediaUrl:      nullStringFromPtr(params.WikipediaURL),
		WikipediaUrlEn:    params.WikipediaURLEn,
		TwitterUsername:   nullStringFromPtr(params.TwitterUsername),
		TwitterUsernameEn: params.TwitterUsernameEn,
		// people.updated_atはtimestamp without time zoneで、Rails版はUTCの時刻を書いている。
		// timestampへのキャストは送った値のタイムゾーンを読み捨てるため、UTCに揃えてから送る。
		Version: sql.NullTime{Time: params.Version.UTC(), Valid: true},
		UserID:  int64(params.UserID),
	}
	if params.Gender != nil {
		queryParams.Gender = sql.NullString{String: string(*params.Gender), Valid: true}
	}
	if params.BloodType != nil {
		queryParams.BloodType = sql.NullString{String: string(*params.BloodType), Valid: true}
	}
	if params.PrefectureID != nil {
		queryParams.PrefectureID = sql.NullInt64{Int64: int64(*params.PrefectureID), Valid: true}
	}

	if _, err := r.queries.UpdateDBPerson(ctx, queryParams); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "index_people_on_name" {
			return false, ErrPersonNameTaken
		}
		return false, err
	}

	return true, nil
}

// GetForStateChangeByIDはAnnict DB管理画面の人物の状態変更 (アーカイブ・アーカイブ解除・
// 削除) の確認画面が必要とするカラム (表示する名前と状態) を読み込みます。呼び出し側が現在の
// 状態でその画面が操作できない人物を弾けるよう、どの状態の人物も返します。存在しない人物では
// nilを返します。
func (r *PersonRepository) GetForStateChangeByID(ctx context.Context, id model.PersonID) (*model.Person, error) {
	row, err := r.queries.GetPersonForStateChangeByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &model.Person{
		ID:     model.PersonID(row.ID),
		Name:   row.Name,
		Status: model.PersonStatus(row.Status),
	}, nil
}

// Archiveは公開中の人物をアーカイブします。人物が存在しないか、公開中でない場合は何も書かずに
// falseを返します。
func (r *PersonRepository) Archive(ctx context.Context, id model.PersonID) (bool, error) {
	return rowAffected(r.queries.ArchiveDBPerson(ctx, int64(id)))
}

// Unarchiveはアーカイブの人物を公開中に戻します。人物が存在しないか、アーカイブでない場合は
// 何も書かずにfalseを返します。
func (r *PersonRepository) Unarchive(ctx context.Context, id model.PersonID) (bool, error) {
	return rowAffected(r.queries.UnarchiveDBPerson(ctx, int64(id)))
}

// Deleteは人物を論理削除します。キャスト・スタッフ・編集履歴は残します。人物が存在しないか、
// 公開中とアーカイブのどちらでもない場合は何も書かずにfalseを返します。
func (r *PersonRepository) Delete(ctx context.Context, id model.PersonID) (bool, error) {
	return rowAffected(r.queries.DeleteDBPerson(ctx, int64(id)))
}

// rowAffectedは条件付きの更新文 (RETURNING id) の結果を、行を書いたかどうかに変換します。
// 条件に一致する行が無かったことはエラーではなくfalseとして返します。
func rowAffected(_ int64, err error) (bool, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// stringPtrFromNullはNULL許容の文字列カラムをポインタに変換し、NULLをnilにします。
func stringPtrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	s := value.String
	return &s
}

// timePtrFromNullはNULL許容の時刻カラムをポインタに変換し、NULLをnilにします。
func timePtrFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}
