create_table "checkins", force: true do |t|
  t.integer  "user_id",                                          null: false
  t.integer  "episode_id",                                       null: false
  t.text     "comment"
  t.boolean  "spoil",                            default: false, null: false
  t.boolean  "modify_comment",                   default: false, null: false
  t.string   "twitter_url_hash",     limit: 510
  t.string   "facebook_url_hash",    limit: 510
  t.integer  "twitter_click_count",              default: 0,     null: false
  t.integer  "facebook_click_count",             default: 0,     null: false
  t.integer  "comments_count",                   default: 0,     null: false
  t.integer  "likes_count",                      default: 0,     null: false
  t.datetime "created_at"
  t.datetime "updated_at"
  t.boolean  "shared_twitter",                   default: false, null: false
  t.boolean  "shared_facebook",                  default: false, null: false
end

add_index "checkins", ["episode_id"], name: "checkins_episode_id_idx", using: :btree
add_index "checkins", ["facebook_url_hash"], name: "checkins_facebook_url_hash_key", unique: true, using: :btree
add_index "checkins", ["twitter_url_hash"], name: "checkins_twitter_url_hash_key", unique: true, using: :btree
add_index "checkins", ["user_id"], name: "checkins_user_id_idx", using: :btree

add_foreign_key "checkins", "episodes", name: "checkins_episode_id_fk", dependent: :delete
add_foreign_key "checkins", "users", name: "checkins_user_id_fk", dependent: :delete
