package usecase

import "testing"

// reconcileTestDesired / reconcileTestExistingは別表の (work由来の) あるべき行と
// 既存行の最小の代役で、具体的なテーブル (タスク2-8以降) に依存せず汎用リコンサイル
// ヘルパーを動かすのに足るだけの構造。
type reconcileTestDesired struct {
	key   string
	value string
}

type reconcileTestExisting struct {
	id      int
	key     string
	value   string
	managed bool
}

func reconcileTestPlan(desired []reconcileTestDesired, existing []reconcileTestExisting) satelliteReconcilePlan[reconcileTestDesired, reconcileTestExisting] {
	return reconcileSatellite(
		desired,
		existing,
		func(d reconcileTestDesired) string { return d.key },
		func(e reconcileTestExisting) string { return e.key },
		func(e reconcileTestExisting) bool { return e.managed },
		func(d reconcileTestDesired, e reconcileTestExisting) bool { return d.value != e.value },
	)
}

func TestReconcileSatellite_Classifies(t *testing.T) {
	t.Parallel()

	desired := []reconcileTestDesired{
		{key: "a", value: "new"},    // 既存の "a" と値が異なる -> update
		{key: "b", value: "same"},   // 既存の "b" と値が一致する -> unchanged
		{key: "c", value: "create"}, // 既存の行が無い -> create
	}
	existing := []reconcileTestExisting{
		{id: 1, key: "a", value: "old", managed: true},
		{id: 2, key: "b", value: "same", managed: true},
		{id: 3, key: "d", value: "stale", managed: true},   // 管理対象だがdesiredに無い -> delete
		{id: 4, key: "e", value: "editor", managed: false}, // 管理対象外でdesiredにも無い -> 保持
	}

	plan := reconcileTestPlan(desired, existing)

	if len(plan.creates) != 1 || plan.creates[0].key != "c" {
		t.Errorf("creates = %+v、期待値 = [c]のみ", plan.creates)
	}

	// 更新は既存行を伴い、適用ステップがそのidentity (id 1) を持てるようにする。
	if len(plan.updates) != 1 {
		t.Fatalf("updates = %d、期待値 = 1 (%+v)", len(plan.updates), plan.updates)
	}
	if plan.updates[0].desired.key != "a" || plan.updates[0].existing.id != 1 {
		t.Errorf("update = %+v、期待値 = desiredのキーがa / existingのidが1", plan.updates[0])
	}

	// 管理下でもう不要になった "d" だけを削除し、編集者が足した "e" (管理外) は残す。
	if len(plan.deletes) != 1 || plan.deletes[0].key != "d" {
		t.Errorf("deletes = %+v、期待値 = [d]のみ", plan.deletes)
	}

	if plan.unchanged != 1 {
		t.Errorf("unchanged = %d、期待値 = 1", plan.unchanged)
	}
}

func TestReconcileSatellite_EmptyDesiredDeletesOnlyManaged(t *testing.T) {
	t.Parallel()

	existing := []reconcileTestExisting{
		{id: 1, key: "a", value: "x", managed: true},
		{id: 2, key: "b", value: "y", managed: true},
		{id: 3, key: "c", value: "z", managed: false},
	}

	plan := reconcileTestPlan(nil, existing)

	if len(plan.creates) != 0 || len(plan.updates) != 0 || plan.unchanged != 0 {
		t.Errorf("creates/updates/unchanged = %d/%d/%d、期待値 = 0/0/0", len(plan.creates), len(plan.updates), plan.unchanged)
	}
	// 管理下の2行は削除され (sourceが何も出さない)、管理外の行は残る。
	if len(plan.deletes) != 2 {
		t.Fatalf("deletes = %d、期待値 = 2 (%+v)", len(plan.deletes), plan.deletes)
	}
	for _, d := range plan.deletes {
		if !d.managed {
			t.Errorf("削除対象の管理対象外行 = %+v、期待値 = 削除対象に含まれないこと", d)
		}
	}
}

func TestReconcileSatellite_AllCreatesAgainstEmptyExisting(t *testing.T) {
	t.Parallel()

	desired := []reconcileTestDesired{{key: "a", value: "1"}, {key: "b", value: "2"}}

	plan := reconcileTestPlan(desired, nil)

	if len(plan.creates) != 2 {
		t.Errorf("creates = %d、期待値 = 2", len(plan.creates))
	}
	if len(plan.updates) != 0 || len(plan.deletes) != 0 || plan.unchanged != 0 {
		t.Errorf("updates/deletes/unchanged = %d/%d/%d、期待値 = 0/0/0", len(plan.updates), len(plan.deletes), plan.unchanged)
	}
}
