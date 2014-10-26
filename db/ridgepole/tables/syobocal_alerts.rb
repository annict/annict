create_table "syobocal_alerts", force: true do |t|
  t.integer  "work_id"
  t.integer  "kind",            null: false
  t.integer  "sc_prog_item_id"
  t.string   "sc_sub_title"
  t.string   "sc_prog_comment"
  t.datetime "created_at"
  t.datetime "updated_at"
end
