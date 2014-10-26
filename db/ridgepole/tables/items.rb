create_table "items", force: true do |t|
  t.integer  "work_id"
  t.string   "name",       limit: 510, null: false
  t.string   "url",        limit: 510, null: false
  t.string   "image_uid",  limit: 510, null: false
  t.boolean  "main",                   null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
