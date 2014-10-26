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
