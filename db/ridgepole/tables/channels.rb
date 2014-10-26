create_table "channels", force: true do |t|
  t.integer  "channel_group_id",                            null: false
  t.integer  "sc_chid",                                     null: false
  t.string   "name",             limit: 510,                null: false
  t.boolean  "published",                    default: true, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
