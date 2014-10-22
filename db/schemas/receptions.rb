create_table "receptions", force: true do |t|
  t.integer  "user_id",    null: false
  t.integer  "channel_id", null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "receptions", ["channel_id"], name: "receptions_channel_id_idx", using: :btree
add_index "receptions", ["user_id", "channel_id"], name: "receptions_user_id_channel_id_key", unique: true, using: :btree
add_index "receptions", ["user_id"], name: "receptions_user_id_idx", using: :btree

add_foreign_key "receptions", "channels", name: "receptions_channel_id_fk", dependent: :delete
add_foreign_key "receptions", "users", name: "receptions_user_id_fk", dependent: :delete
