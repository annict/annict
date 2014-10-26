create_table "notifications", force: true do |t|
  t.integer  "user_id",                                    null: false
  t.integer  "action_user_id",                             null: false
  t.integer  "trackable_id",                               null: false
  t.string   "trackable_type", limit: 510,                 null: false
  t.string   "action",         limit: 510,                 null: false
  t.boolean  "read",                       default: false, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end
