create_table "cover_images", force: true do |t|
  t.integer  "work_id",                null: false
  t.string   "file_name",  limit: 510, null: false
  t.string   "location",   limit: 510, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
