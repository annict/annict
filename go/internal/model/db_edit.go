package model

import "time"

// DBEditActorはAnnict DBの編集履歴の操作を行った主体で、PostgreSQLのdb_edit_actor enumと対応する。
type DBEditActor string

const (
	// DBEditActorUserは編集者の編集。編集者が退会すると、操作のUserIDは空になる。
	DBEditActorUser DBEditActor = "user"
	// DBEditActorSystemはエピソードの自動生成のような、アプリケーションの書き込み。
	DBEditActorSystem DBEditActor = "system"
)

// Stringは主体の文字列表現を返す。
func (a DBEditActor) String() string { return string(a) }

// DBEditActionはAnnict DBの編集履歴の操作の種類で、PostgreSQLのdb_edit_action enumと対応する。
type DBEditAction string

const (
	DBEditActionCreate    DBEditAction = "create"
	DBEditActionUpdate    DBEditAction = "update"
	DBEditActionArchive   DBEditAction = "archive"
	DBEditActionUnarchive DBEditAction = "unarchive"
	DBEditActionDelete    DBEditAction = "delete"
	// DBEditActionDelist / DBEditActionUndelistは人物の掲載停止と掲載再開。
	DBEditActionDelist   DBEditAction = "delist"
	DBEditActionUndelist DBEditAction = "undelist"
)

// Stringは操作の種類の文字列表現を返す。
func (a DBEditAction) String() string { return string(a) }

// DBEditはAnnict DBの編集履歴の操作 (1回の保存) を表す。リソースごとの変更後の状態は、
// 操作に従属する版が持つ。
type DBEdit struct {
	ID    DBEditID
	Actor DBEditActor
	// UserIDは編集した編集者。システムの書き込みと、退会したユーザーの編集では空になる。
	UserID    *UserID
	Action    DBEditAction
	CreatedAt time.Time

	// Userは編集者。ID・Username・ProfileImageDataだけを持ち、値が入るのは版の履歴のローダーと操作の一覧のローダー
	// (DBEditRepository.List) で編集者がいるときのみ。
	User *User

	// AnimeRevisionsとPersonRevisionsは操作に属する作品・エピソードと人物の版で、閲覧者に見せる版だけを
	// IDの昇順に持つ。値が入るのはAnnict DB全体の編集履歴の読み取りユースケースのみ。
	AnimeRevisions  []*AnimeRevision
	PersonRevisions []*PersonRevision
}
