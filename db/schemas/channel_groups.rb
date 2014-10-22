create_table "channel_groups", force: true do |t|
  t.string   "sc_chgid",    limit: 510, null: false
  t.string   "name",        limit: 510, null: false
  t.integer  "sort_number"
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "channel_groups", ["sc_chgid"], name: "channel_groups_sc_chgid_key", unique: true, using: :btree
