create_table "comments", force: true do |t|
  t.integer  "user_id",                 null: false
  t.integer  "checkin_id",              null: false
  t.text     "body",                    null: false
  t.integer  "likes_count", default: 0, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "comments", ["checkin_id"], name: "comments_checkin_id_idx", using: :btree
add_index "comments", ["user_id"], name: "comments_user_id_idx", using: :btree

add_foreign_key "comments", "checkins", name: "comments_checkin_id_fk", dependent: :delete
add_foreign_key "comments", "users", name: "comments_user_id_fk", dependent: :delete
