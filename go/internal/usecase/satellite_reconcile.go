package usecase

// satelliteUpdate pairs a desired row (derived from a work) with the existing row it
// updates, so the apply step has the existing row's identity (primary key) to target.
//
// [Ja] satelliteUpdate は (work から導出した) あるべき行と、それが更新する既存行を対にする。
// 適用ステップが更新対象の既存行の identity (主キー) を得られるようにするため。
type satelliteUpdate[D any, E any] struct {
	desired  D
	existing E
}

// satelliteReconcilePlan is the create / update / delete plan a single satellite
// table's reconcile produces. It carries no I/O; the per-table reconciler turns it
// into writes inside its transaction.
//
// [Ja] satelliteReconcilePlan は単一の別表のリコンサイルが生む作成 / 更新 / 削除の計画。
// I/O を持たず、テーブルごとのリコンサイラがトランザクション内で書き込みに変換する。
type satelliteReconcilePlan[D any, E any] struct {
	creates   []D
	updates   []satelliteUpdate[D, E]
	deletes   []E
	unchanged int
}

// reconcileSatellite matches the desired rows (derived from the works being synced)
// against the existing rows on the anime by natural key, and returns the create /
// update / delete plan without performing any I/O. It is the shared core every
// per-table satellite reconciler (tasks 2-8 onward) builds on, so the "desired set vs
// existing set" diff logic lives in one place and stays consistent across the six
// satellite tables.
//
// The desiredKey / existingKey returned for a row must be unique within the batch a
// reconciler hands in. A reconciler reconciles a whole page of works at once, so the
// natural key normally includes the anime_id (e.g. (anime_id, service) for
// anime_external_ids); a key that collides across rows is mis-planned silently — a
// duplicate desired key double-creates, and a duplicate existing key leaves the
// shadowed row unreconciled.
//
// Deletes are limited to existing rows the works are the source of: existingManaged
// reports whether an existing row falls within the works-managed key space (e.g. links
// with kind in {official_site, wikipedia}). Rows outside that space (editor-added rows
// such as kind='other' links or a secondary official account) are never deleted, so a
// later phase where editors edit animes directly does not have its rows clobbered by
// the sync. During this migration works are the only source, so no such rows exist yet.
//
// [Ja] reconcileSatellite は (同期中の works から導出した) あるべき行を anime 上の既存行と
// 自然キーで突合し、I/O を行わずに作成 / 更新 / 削除の計画を返す。テーブルごとの別表
// リコンサイラ (タスク 2-8 以降) が共有する中核で、「あるべき集合 vs 既存集合」の差分
// ロジックを 1 箇所に置き、6 つの別表で一貫させる。
//
// 各行に対して返す desiredKey / existingKey は、リコンサイラが渡すバッチ内で一意でなければ
// ならない。リコンサイラは works のページ全体を一括でリコンサイルするため、自然キーには通常
// anime_id を含める (例: anime_external_ids なら (anime_id, service))。行をまたいでキーが
// 衝突すると静かに誤った計画になる — desired キーの重複は二重 create を生み、existing キーの
// 重複は隠れた行が未リコンサイルのまま残る。
//
// 削除は work が源泉とする既存行だけを対象にする。existingManaged は既存行が works 管理下の
// キー空間 (例: kind が {official_site, wikipedia} のリンク) に入るかを返す。その空間の外の
// 行 (編集者が足した kind='other' のリンクや 2 つ目の公式アカウントなど) は決して削除しない
// ため、編集者が後続フェーズで anime を直接編集しても同期にその行を壊されない。本移行期間は
// works が唯一の源泉のため、そうした行はまだ存在しない。
func reconcileSatellite[D any, E any, K comparable](
	desired []D,
	existing []E,
	desiredKey func(D) K,
	existingKey func(E) K,
	existingManaged func(E) bool,
	changed func(desired D, existing E) bool,
) satelliteReconcilePlan[D, E] {
	existingByKey := make(map[K]E, len(existing))
	for _, e := range existing {
		existingByKey[existingKey(e)] = e
	}

	desiredKeys := make(map[K]struct{}, len(desired))
	var plan satelliteReconcilePlan[D, E]

	for _, d := range desired {
		key := desiredKey(d)
		desiredKeys[key] = struct{}{}

		e, ok := existingByKey[key]
		if !ok {
			plan.creates = append(plan.creates, d)
			continue
		}
		if changed(d, e) {
			plan.updates = append(plan.updates, satelliteUpdate[D, E]{desired: d, existing: e})
		} else {
			plan.unchanged++
		}
	}

	for _, e := range existing {
		if !existingManaged(e) {
			continue
		}
		if _, ok := desiredKeys[existingKey(e)]; ok {
			continue
		}
		plan.deletes = append(plan.deletes, e)
	}

	return plan
}
