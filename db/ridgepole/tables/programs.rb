create_table "programs", force: true do |t|
  t.integer  "channel_id",     null: false
  t.integer  "episode_id",     null: false
  t.integer  "work_id",        null: false
  t.datetime "started_at",     null: false
  t.datetime "sc_last_update"
  t.datetime "created_at"
  t.datetime "updated_at"
end
