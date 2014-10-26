create_table "versions", force: true do |t|
  t.string   "item_type",  limit: 510, null: false
  t.integer  "item_id",                null: false
  t.string   "event",      limit: 510, null: false
  t.string   "whodunnit",  limit: 510
  t.text     "object"
  t.datetime "created_at"
end
