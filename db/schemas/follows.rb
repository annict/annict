create_table "follows", force: true do |t|
  t.integer  "user_id",      null: false
  t.integer  "following_id", null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "follows", ["following_id"], name: "follows_following_id_idx", using: :btree
add_index "follows", ["user_id", "following_id"], name: "follows_user_id_following_id_key", unique: true, using: :btree
add_index "follows", ["user_id"], name: "follows_user_id_idx", using: :btree

add_foreign_key "follows", "users", name: "follows_following_id_fk", column: "following_id", dependent: :delete
add_foreign_key "follows", "users", name: "follows_user_id_fk", dependent: :delete
