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
