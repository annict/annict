create_table "finished_tips", force: true do |t|
  t.integer  "user_id",    null: false
  t.integer  "tip_id",     null: false
  t.datetime "created_at", null: false
  t.datetime "updated_at", null: false
end
