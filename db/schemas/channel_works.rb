create_table "channel_works", force: true do |t|
  t.integer  "user_id",    null: false
  t.integer  "work_id",    null: false
  t.integer  "channel_id", null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "channel_works", ["channel_id"], name: "channel_works_channel_id_idx", using: :btree
add_index "channel_works", ["user_id", "work_id", "channel_id"], name: "channel_works_user_id_work_id_channel_id_key", unique: true, using: :btree
add_index "channel_works", ["user_id"], name: "channel_works_user_id_idx", using: :btree
add_index "channel_works", ["work_id"], name: "channel_works_work_id_idx", using: :btree

add_foreign_key "channel_works", "channels", name: "channel_works_channel_id_fk", dependent: :delete
add_foreign_key "channel_works", "users", name: "channel_works_user_id_fk", dependent: :delete
add_foreign_key "channel_works", "works", name: "channel_works_work_id_fk", dependent: :delete
