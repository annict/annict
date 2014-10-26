create_table "sessions", force: true do |t|
  t.string   "session_id", limit: 510, null: false
  t.text     "data"
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "sessions", ["session_id"], name: "sessions_session_id_key", unique: true, using: :btree
