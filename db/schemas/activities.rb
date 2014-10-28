create_table "activities", force: true do |t|
  t.integer  "user_id",                    null: false
  t.integer  "recipient_id",               null: false
  t.string   "recipient_type", limit: 510, null: false
  t.integer  "trackable_id",               null: false
  t.string   "trackable_type", limit: 510, null: false
  t.string   "action",         limit: 510, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "activities", ["user_id"], name: "activities_user_id_idx", using: :btree

add_foreign_key "activities", "users", name: "activities_user_id_fk", dependent: :delete
