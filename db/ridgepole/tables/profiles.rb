create_table "profiles", force: true do |t|
  t.integer  "user_id",                                       null: false
  t.string   "name",                 limit: 510, default: "", null: false
  t.string   "description",          limit: 510, default: "", null: false
  t.string   "avatar_uid",           limit: 510
  t.string   "background_image_uid", limit: 510
  t.datetime "created_at"
  t.datetime "updated_at"
end
