create_table "statuses", force: true do |t|
  t.integer  "user_id",                     null: false
  t.integer  "work_id",                     null: false
  t.integer  "kind",                        null: false
  t.boolean  "latest",      default: false, null: false
  t.integer  "likes_count", default: 0,     null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "statuses", ["user_id"], name: "statuses_user_id_idx", using: :btree
add_index "statuses", ["work_id"], name: "statuses_work_id_idx", using: :btree

add_foreign_key "statuses", "users", name: "statuses_user_id_fk", dependent: :delete
add_foreign_key "statuses", "works", name: "statuses_work_id_fk", dependent: :delete
