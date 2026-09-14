package model

import "time"

// EpisodeStatusはunpublished_at / deleted_atから導出したエピソードのライフサイクル
// 状態 (published / archived / deleted) を表す。
type EpisodeStatus string

const (
	EpisodeStatusPublished EpisodeStatus = "published"
	EpisodeStatusArchived  EpisodeStatus = "archived"
	EpisodeStatusDeleted   EpisodeStatus = "deleted"
)

// Stringはステータスの文字列表現を返す。
func (s EpisodeStatus) String() string { return string(s) }

// Episodeはエピソードのドメインエンティティ (ページに依存しない汎用的な構造)。
// Domain層に属し、Presentation層に依存しない。
//
// AnimeIDとParentAnimeIDに値が入るのはanime同期ローダー
// (ListForAnimeSyncByIDs) のみ。フェーズ2のリコンシリエーションでanimes /
// anime_classificationsに写像するepisodesカラムを射影したもの。AnimeIDは
// episodes.anime_idのマッピングカラム (nilは未同期 = anime未作成)。ParentAnimeIDは
// episodes.work_id経由で解決した親作品のanime_idで、nilは親作品が未同期であることを
// 表し、その場合は後続の実行までエピソードをリコンサイルできない。
type Episode struct {
	ID            EpisodeID
	WorkID        WorkID
	Title         *string
	TitleRo       string
	TitleEn       string
	Number        *string
	SortNumber    int32
	RawNumber     *float64
	AnimeID       *AnimeID
	ParentAnimeID *AnimeID

	// UnpublishedAt / DeletedAtはエピソードの状態を表す正本カラム (Unpublishable /
	// SoftDeletable)。フェーズ2のリコンシリエーションはDerivedStatusを通じてこれらから
	// anime.statusを導出する。値が入るのはanime同期ローダー (ListForAnimeSyncByIDs) と
	// Annict DB一覧のローダー (ListForDB)。どちらのカラムも選択しないローダーで取得した
	// エピソードでは両方がnilのまま残るため、DerivedStatusはpublishedを返す。
	UnpublishedAt *time.Time
	DeletedAt     *time.Time

	// UpdatedAtは行が最後に書かれた時刻。値を入れるのはAnnict DB編集フォームの
	// ローダー (GetForEditByID) で、フォームは送信が前提とする版としてこれを持ち帰る。
	// 他者の変更を黙って上書きする送信を、更新側で却下できるようにするため。カラムを
	// 選択しないローダーと、保存済みのupdated_atがNULLの行ではnilのまま残る。
	UpdatedAt *time.Time

	// EpisodeRecordsCountはepisodes.episode_records_countのカウンターキャッシュ。
	// 値が入るのはAnnict DB一覧のローダー (ListForDB) のみで、他のロード経路では0の
	// まま残る。
	EpisodeRecordsCount int32

	// PrevNumberとPrevRawNumberは、sort_number順でこのエピソードの直前に来る
	// エピソードの表示用話数と数値話数。Annict DB一覧のローダー (ListForDB) が隣接行から
	// 導出する。作品の最初のエピソードと、導出しないローダーではいずれもnilのまま残る。
	PrevNumber    *string
	PrevRawNumber *float64
}

// DerivedStatusはepisodeの状態の正本であるUnpublishable / SoftDeletable
// タイムスタンプからエピソードのライフサイクル状態を導出する。deleted_atがunpublished_at
// より優先される (削除済みのエピソードは公開状態に関わらずdeleted)。これはRailsの可視性
// scope only_kept = without_deleted.published (公開は両方がNULLのとき) に揃う。
// timestampsからstatusへの優先順位を定めるのはこの1箇所で、これにより呼び出し側が
// 優先順位を再実装せずに済む。
func (e *Episode) DerivedStatus() EpisodeStatus {
	switch {
	case e.DeletedAt != nil:
		return EpisodeStatusDeleted
	case e.UnpublishedAt != nil:
		return EpisodeStatusArchived
	default:
		return EpisodeStatusPublished
	}
}

// ManualEpisodeCreationRestrictionは編集者が作品のエピソードを手動作成できない理由を
// 表す。
type ManualEpisodeCreationRestriction string

const (
	ManualEpisodeCreationAllowed        ManualEpisodeCreationRestriction = ""
	ManualEpisodeCreationEpisodesFilled ManualEpisodeCreationRestriction = "episodes_filled"
	ManualEpisodeCreationSlotsExist     ManualEpisodeCreationRestriction = "slots_exist"
)

// ManualEpisodeCreationStateはRailsが管理者以外の手動エピソード作成を止める条件を
// 保持する。作品が予定話数までエピソードを持っている (Work#episodes_filled?) か、開始時刻を
// 持つ放送枠があってエピソードが自動生成される (Work#slots_exists?) かのいずれか。
type ManualEpisodeCreationState struct {
	EpisodesFilled bool
	SlotsExist     bool
}

// Restrictionはどの理由が当てはまるかを返し、いずれも当てはまらない場合は
// ManualEpisodeCreationAllowedを返す。両方に当てはまる作品は予定話数到達を報告する
// (Railsのフォームが理由を述べる順序と同じ)。順序をここで決めることで、送信の却下と
// ページの警告が同じ理由を名指しする。
func (s ManualEpisodeCreationState) Restriction() ManualEpisodeCreationRestriction {
	switch {
	case s.EpisodesFilled:
		return ManualEpisodeCreationEpisodesFilled
	case s.SlotsExist:
		return ManualEpisodeCreationSlotsExist
	default:
		return ManualEpisodeCreationAllowed
	}
}

// Allowedは通常のコミッターがエピソードを手動作成できるかを返す。
func (s ManualEpisodeCreationState) Allowed() bool {
	return s.Restriction() == ManualEpisodeCreationAllowed
}
