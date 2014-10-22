create_table "works", force: true do |t|
  t.integer  "season_id"
  t.integer  "sc_tid"
  t.string   "title",             limit: 510,              null: false
  t.integer  "media",                                      null: false
  t.string   "official_site_url", limit: 510, default: "", null: false
  t.string   "wikipedia_url",     limit: 510, default: "", null: false
  t.integer  "episodes_count",                default: 0,  null: false
  t.integer  "watchers_count",                default: 0,  null: false
  t.date     "released_at"
  t.datetime "nicoch_started_at"
  t.datetime "created_at"
  t.datetime "updated_at"
  t.boolean  "on_air",                                     null: false
  t.boolean  "fetch_syobocal",                             null: false
  t.string   "twitter_username",  limit: 510
  t.string   "twitter_hashtag",   limit: 510
end

add_index "works", ["sc_tid"], name: "works_sc_tid_key", unique: true, using: :btree
add_index "works", ["season_id"], name: "works_season_id_idx", using: :btree

add_foreign_key "works", "seasons", name: "works_season_id_fk", dependent: :delete
