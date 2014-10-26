create_table "statuses", force: true do |t|
  t.integer  "user_id",                     null: false
  t.integer  "work_id",                     null: false
  t.integer  "kind",                        null: false
  t.boolean  "latest",      default: false, null: false
  t.integer  "likes_count", default: 0,     null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
