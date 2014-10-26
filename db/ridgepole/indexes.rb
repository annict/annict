# ==========================================
# add_index
# ==========================================

add_index "activities", ["user_id"], name: "activities_user_id_idx", using: :btree

add_index "channel_groups", ["sc_chgid"], name: "channel_groups_sc_chgid_key", unique: true, using: :btree

add_index "channel_works", ["channel_id"], name: "channel_works_channel_id_idx", using: :btree
add_index "channel_works", ["user_id", "work_id", "channel_id"], name: "channel_works_user_id_work_id_channel_id_key", unique: true, using: :btree
add_index "channel_works", ["user_id"], name: "channel_works_user_id_idx", using: :btree
add_index "channel_works", ["work_id"], name: "channel_works_work_id_idx", using: :btree

add_index "channels", ["channel_group_id"], name: "channels_channel_group_id_idx", using: :btree
add_index "channels", ["sc_chid"], name: "channels_sc_chid_key", unique: true, using: :btree

add_index "checkins", ["episode_id"], name: "checkins_episode_id_idx", using: :btree
add_index "checkins", ["facebook_url_hash"], name: "checkins_facebook_url_hash_key", unique: true, using: :btree
add_index "checkins", ["twitter_url_hash"], name: "checkins_twitter_url_hash_key", unique: true, using: :btree
add_index "checkins", ["user_id"], name: "checkins_user_id_idx", using: :btree

add_index "comments", ["checkin_id"], name: "comments_checkin_id_idx", using: :btree
add_index "comments", ["user_id"], name: "comments_user_id_idx", using: :btree

add_index "cover_images", ["work_id"], name: "cover_images_work_id_idx", using: :btree

add_index "episodes", ["work_id", "sc_count"], name: "episodes_work_id_sc_count_key", unique: true, using: :btree
add_index "episodes", ["work_id"], name: "episodes_work_id_idx", using: :btree

add_index "follows", ["following_id"], name: "follows_following_id_idx", using: :btree
add_index "follows", ["user_id", "following_id"], name: "follows_user_id_following_id_key", unique: true, using: :btree
add_index "follows", ["user_id"], name: "follows_user_id_idx", using: :btree

add_index "items", ["work_id"], name: "items_work_id_idx", using: :btree

add_index "likes", ["user_id"], name: "likes_user_id_idx", using: :btree

add_index "notifications", ["action_user_id"], name: "notifications_action_user_id_idx", using: :btree
add_index "notifications", ["user_id"], name: "notifications_user_id_idx", using: :btree

add_index "profiles", ["user_id"], name: "profiles_user_id_idx", using: :btree
add_index "profiles", ["user_id"], name: "profiles_user_id_key", unique: true, using: :btree

add_index "programs", ["channel_id", "episode_id"], name: "programs_channel_id_episode_id_key", unique: true, using: :btree
add_index "programs", ["channel_id"], name: "programs_channel_id_idx", using: :btree
add_index "programs", ["episode_id"], name: "programs_episode_id_idx", using: :btree
add_index "programs", ["work_id"], name: "programs_work_id_idx", using: :btree

add_index "providers", ["name", "uid"], name: "providers_name_uid_key", unique: true, using: :btree
add_index "providers", ["user_id"], name: "providers_user_id_idx", using: :btree

add_index "receptions", ["channel_id"], name: "receptions_channel_id_idx", using: :btree
add_index "receptions", ["user_id", "channel_id"], name: "receptions_user_id_channel_id_key", unique: true, using: :btree
add_index "receptions", ["user_id"], name: "receptions_user_id_idx", using: :btree

add_index "seasons", ["slug"], name: "seasons_slug_key", unique: true, using: :btree

add_index "staffs", ["email"], name: "staffs_email_key", unique: true, using: :btree

add_index "statuses", ["user_id"], name: "statuses_user_id_idx", using: :btree
add_index "statuses", ["work_id"], name: "statuses_work_id_idx", using: :btree

add_index "syobocal_alerts", ["kind"], name: "index_syobocal_alerts_on_kind", using: :btree
add_index "syobocal_alerts", ["sc_prog_item_id"], name: "index_syobocal_alerts_on_sc_prog_item_id", using: :btree

add_index "twitter_bots", ["name"], name: "twitter_bots_name_key", unique: true, using: :btree

add_index "users", ["confirmation_token"], name: "users_confirmation_token_key", unique: true, using: :btree
add_index "users", ["email"], name: "users_email_key", unique: true, using: :btree
add_index "users", ["username"], name: "users_username_key", unique: true, using: :btree

add_index "works", ["sc_tid"], name: "works_sc_tid_key", unique: true, using: :btree
add_index "works", ["season_id"], name: "works_season_id_idx", using: :btree


# ==========================================
# add_foreign_key
# ==========================================

add_foreign_key "activities", "users", name: "activities_user_id_fk", dependent: :delete

add_foreign_key "channel_works", "channels", name: "channel_works_channel_id_fk", dependent: :delete
add_foreign_key "channel_works", "users", name: "channel_works_user_id_fk", dependent: :delete
add_foreign_key "channel_works", "works", name: "channel_works_work_id_fk", dependent: :delete

add_foreign_key "channels", "channel_groups", name: "channels_channel_group_id_fk", dependent: :delete

add_foreign_key "checkins", "episodes", name: "checkins_episode_id_fk", dependent: :delete
add_foreign_key "checkins", "users", name: "checkins_user_id_fk", dependent: :delete

add_foreign_key "comments", "checkins", name: "comments_checkin_id_fk", dependent: :delete
add_foreign_key "comments", "users", name: "comments_user_id_fk", dependent: :delete

add_foreign_key "cover_images", "works", name: "cover_images_work_id_fk", dependent: :delete

add_foreign_key "episodes", "works", name: "episodes_work_id_fk", dependent: :delete

add_foreign_key "follows", "users", name: "follows_following_id_fk", column: "following_id", dependent: :delete
add_foreign_key "follows", "users", name: "follows_user_id_fk", dependent: :delete

add_foreign_key "items", "works", name: "items_work_id_fk", dependent: :delete

add_foreign_key "likes", "users", name: "likes_user_id_fk", dependent: :delete

add_foreign_key "notifications", "users", name: "notifications_action_user_id_fk", column: "action_user_id", dependent: :delete
add_foreign_key "notifications", "users", name: "notifications_user_id_fk", dependent: :delete

add_foreign_key "profiles", "users", name: "profiles_user_id_fk", dependent: :delete

add_foreign_key "programs", "channels", name: "programs_channel_id_fk", dependent: :delete
add_foreign_key "programs", "episodes", name: "programs_episode_id_fk", dependent: :delete
add_foreign_key "programs", "works", name: "programs_work_id_fk", dependent: :delete

add_foreign_key "providers", "users", name: "providers_user_id_fk", dependent: :delete

add_foreign_key "receptions", "channels", name: "receptions_channel_id_fk", dependent: :delete
add_foreign_key "receptions", "users", name: "receptions_user_id_fk", dependent: :delete

add_foreign_key "statuses", "users", name: "statuses_user_id_fk", dependent: :delete
add_foreign_key "statuses", "works", name: "statuses_work_id_fk", dependent: :delete

add_foreign_key "syobocal_alerts", "works", name: "syobocal_alerts_work_id_fk", dependent: :delete

add_foreign_key "works", "seasons", name: "works_season_id_fk", dependent: :delete
