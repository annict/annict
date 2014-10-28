create_table "providers", force: true do |t|
  t.integer  "user_id",                      null: false
  t.string   "name",             limit: 510, null: false
  t.string   "uid",              limit: 510, null: false
  t.string   "token",            limit: 510, null: false
  t.integer  "token_expires_at"
  t.string   "token_secret",     limit: 510
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "providers", ["name", "uid"], name: "providers_name_uid_key", unique: true, using: :btree
add_index "providers", ["user_id"], name: "providers_user_id_idx", using: :btree

add_foreign_key "providers", "users", name: "providers_user_id_fk", dependent: :delete
