create_table "episodes", force: true do |t|
  t.integer  "work_id",                                null: false
  t.string   "number",         limit: 510
  t.integer  "sort_number",                default: 0, null: false
  t.integer  "sc_count"
  t.string   "title",          limit: 510
  t.boolean  "single"
  t.integer  "checkins_count",             default: 0, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "episodes", ["work_id", "sc_count"], name: "episodes_work_id_sc_count_key", unique: true, using: :btree
add_index "episodes", ["work_id"], name: "episodes_work_id_idx", using: :btree

add_foreign_key "episodes", "works", name: "episodes_work_id_fk", dependent: :delete
