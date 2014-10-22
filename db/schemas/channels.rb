create_table "channels", force: true do |t|
  t.integer  "channel_group_id",                            null: false
  t.integer  "sc_chid",                                     null: false
  t.string   "name",             limit: 510,                null: false
  t.boolean  "published",                    default: true, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "channels", ["channel_group_id"], name: "channels_channel_group_id_idx", using: :btree
add_index "channels", ["sc_chid"], name: "channels_sc_chid_key", unique: true, using: :btree

add_foreign_key "channels", "channel_groups", name: "channels_channel_group_id_fk", dependent: :delete
