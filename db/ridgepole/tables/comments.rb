create_table "comments", force: true do |t|
  t.integer  "user_id",                 null: false
  t.integer  "checkin_id",              null: false
  t.text     "body",                    null: false
  t.integer  "likes_count", default: 0, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
