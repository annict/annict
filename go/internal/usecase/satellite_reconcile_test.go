package usecase

import "testing"

// reconcileTestDesired / reconcileTestExisting are minimal stand-ins for a satellite
// table's desired (work-derived) and existing rows, just enough to exercise the generic
// reconcile helper without depending on any concrete table (those land in tasks 2-8+).
//
// [Ja] reconcileTestDesired / reconcileTestExisting は別表の (work 由来の) あるべき行と
// 既存行の最小の代役で、具体的なテーブル (タスク 2-8 以降) に依存せず汎用リコンサイル
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
		{key: "a", value: "new"},    // existing "a" differs -> update
		{key: "b", value: "same"},   // existing "b" matches -> unchanged
		{key: "c", value: "create"}, // no existing row -> create
	}
	existing := []reconcileTestExisting{
		{id: 1, key: "a", value: "old", managed: true},
		{id: 2, key: "b", value: "same", managed: true},
		{id: 3, key: "d", value: "stale", managed: true},   // managed, not desired -> delete
		{id: 4, key: "e", value: "editor", managed: false}, // unmanaged, not desired -> preserved
	}

	plan := reconcileTestPlan(desired, existing)

	if len(plan.creates) != 1 || plan.creates[0].key != "c" {
		t.Errorf("creates = %+v, want exactly [c]", plan.creates)
	}

	// The update carries the existing row so the apply step has its identity (id 1).
	//
	// [Ja] 更新は既存行を伴い、適用ステップがその identity (id 1) を持てるようにする。
	if len(plan.updates) != 1 {
		t.Fatalf("updates = %d, want 1 (%+v)", len(plan.updates), plan.updates)
	}
	if plan.updates[0].desired.key != "a" || plan.updates[0].existing.id != 1 {
		t.Errorf("update = %+v, want desired key a / existing id 1", plan.updates[0])
	}

	// Only the managed, no-longer-desired row "d" is deleted; the editor-added row "e"
	// (unmanaged) is preserved.
	//
	// [Ja] 管理下でもう不要になった "d" だけを削除し、編集者が足した "e" (管理外) は残す。
	if len(plan.deletes) != 1 || plan.deletes[0].key != "d" {
		t.Errorf("deletes = %+v, want exactly [d]", plan.deletes)
	}

	if plan.unchanged != 1 {
		t.Errorf("unchanged = %d, want 1", plan.unchanged)
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
		t.Errorf("creates/updates/unchanged = %d/%d/%d, want 0/0/0", len(plan.creates), len(plan.updates), plan.unchanged)
	}
	// Both managed rows are deleted (source emits nothing); the unmanaged row is kept.
	//
	// [Ja] 管理下の 2 行は削除され (source が何も出さない)、管理外の行は残る。
	if len(plan.deletes) != 2 {
		t.Fatalf("deletes = %d, want 2 (%+v)", len(plan.deletes), plan.deletes)
	}
	for _, d := range plan.deletes {
		if !d.managed {
			t.Errorf("deleted an unmanaged row: %+v", d)
		}
	}
}

func TestReconcileSatellite_AllCreatesAgainstEmptyExisting(t *testing.T) {
	t.Parallel()

	desired := []reconcileTestDesired{{key: "a", value: "1"}, {key: "b", value: "2"}}

	plan := reconcileTestPlan(desired, nil)

	if len(plan.creates) != 2 {
		t.Errorf("creates = %d, want 2", len(plan.creates))
	}
	if len(plan.updates) != 0 || len(plan.deletes) != 0 || plan.unchanged != 0 {
		t.Errorf("updates/deletes/unchanged = %d/%d/%d, want 0/0/0", len(plan.updates), len(plan.deletes), plan.unchanged)
	}
}
