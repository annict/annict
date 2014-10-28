create_table "profiles", force: true do |t|
  t.integer  "user_id",                                       null: false
  t.string   "name",                 limit: 510, default: "", null: false
  t.string   "description",          limit: 510, default: "", null: false
  t.string   "avatar_uid",           limit: 510
  t.string   "background_image_uid", limit: 510
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "profiles", ["user_id"], name: "profiles_user_id_idx", using: :btree
add_index "profiles", ["user_id"], name: "profiles_user_id_key", unique: true, using: :btree

add_foreign_key "profiles", "users", name: "profiles_user_id_fk", dependent: :delete
