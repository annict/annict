create_table "syobocal_alerts", force: true do |t|
  t.integer  "work_id"
  t.integer  "kind",            null: false
  t.integer  "sc_prog_item_id"
  t.string   "sc_sub_title"
  t.string   "sc_prog_comment"
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "syobocal_alerts", ["kind"], name: "index_syobocal_alerts_on_kind", using: :btree
add_index "syobocal_alerts", ["sc_prog_item_id"], name: "index_syobocal_alerts_on_sc_prog_item_id", using: :btree

add_foreign_key "syobocal_alerts", "works", name: "syobocal_alerts_work_id_fk", dependent: :delete
