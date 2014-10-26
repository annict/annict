create_table "likes", force: true do |t|
  t.integer  "user_id",                    null: false
  t.integer  "recipient_id",               null: false
  t.string   "recipient_type", limit: 510, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
