package usecase

// satelliteUpdateは (workから導出した) あるべき行と、それが更新する既存行を対にする。
// 適用ステップが更新対象の既存行のidentity (主キー) を得られるようにするため。
type satelliteUpdate[D any, E any] struct {
	desired  D
	existing E
}

// satelliteReconcilePlanは単一の別表のリコンサイルが生む作成 / 更新 / 削除の計画。
// I/Oを持たず、テーブルごとのリコンサイラがトランザクション内で書き込みに変換する。
type satelliteReconcilePlan[D any, E any] struct {
	creates   []D
	updates   []satelliteUpdate[D, E]
	deletes   []E
	unchanged int
}

// reconcileSatelliteは (同期中のworksから導出した) あるべき行をanime上の既存行と
// 自然キーで突合し、I/Oを行わずに作成 / 更新 / 削除の計画を返す。テーブルごとの別表
// リコンサイラ (タスク2-8以降) が共有する中核で、「あるべき集合vs既存集合」の差分
// ロジックを1箇所に置き、6つの別表で一貫させる。
//
// 各行に対して返すdesiredKey / existingKeyは、リコンサイラが渡すバッチ内で一意でなければ
// ならない。リコンサイラはworksのページ全体を一括でリコンサイルするため、自然キーには通常
// anime_idを含める (例: anime_external_idsなら (anime_id, service))。行をまたいでキーが
// 衝突すると静かに誤った計画になる — desiredキーの重複は二重createを生み、existingキーの
// 重複は隠れた行が未リコンサイルのまま残る。
//
// 削除はworkが源泉とする既存行だけを対象にする。existingManagedは既存行がworks管理下の
// キー空間 (例: kindが {official_site, wikipedia} のリンク) に入るかを返す。その空間の外の
// 行 (編集者が足したkind='other' のリンクや2つ目の公式アカウントなど) は決して削除しない
// ため、編集者が後続フェーズでanimeを直接編集しても同期にその行を壊されない。本移行期間は
// worksが唯一の源泉のため、そうした行はまだ存在しない。
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
