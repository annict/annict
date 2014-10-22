create_table "programs", force: true do |t|
  t.integer  "channel_id",     null: false
  t.integer  "episode_id",     null: false
  t.integer  "work_id",        null: false
  t.datetime "started_at",     null: false
  t.datetime "sc_last_update"
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "programs", ["channel_id", "episode_id"], name: "programs_channel_id_episode_id_key", unique: true, using: :btree
add_index "programs", ["channel_id"], name: "programs_channel_id_idx", using: :btree
add_index "programs", ["episode_id"], name: "programs_episode_id_idx", using: :btree
add_index "programs", ["work_id"], name: "programs_work_id_idx", using: :btree

add_foreign_key "programs", "channels", name: "programs_channel_id_fk", dependent: :delete
add_foreign_key "programs", "episodes", name: "programs_episode_id_fk", dependent: :delete
add_foreign_key "programs", "works", name: "programs_work_id_fk", dependent: :delete
