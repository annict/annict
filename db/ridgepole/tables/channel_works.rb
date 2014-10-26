create_table "channel_works", force: true do |t|
  t.integer  "user_id",    null: false
  t.integer  "work_id",    null: false
  t.integer  "channel_id", null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
