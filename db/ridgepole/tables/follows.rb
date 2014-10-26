create_table "follows", force: true do |t|
  t.integer  "user_id",      null: false
  t.integer  "following_id", null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
