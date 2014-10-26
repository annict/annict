create_table "seasons", force: true do |t|
  t.string   "name",       limit: 510, null: false
  t.string   "slug",       limit: 510, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
