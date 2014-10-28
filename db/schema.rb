# encoding: UTF-8
# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# Note that this schema.rb definition is the authoritative source for your
# database schema. If you need to create the application database on another
# system, you should be using db:schema:load, not running all the migrations
# from scratch. The latter is a flawed and unsustainable approach (the more migrations
# you'll amass, the slower it'll run and the greater likelihood for issues).
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema.define(version: 20141021143930) do

  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"

  create_table "activities", force: true do |t|
    t.integer  "user_id",        null: false
    t.integer  "recipient_id",   null: false
    t.string   "recipient_type", null: false
    t.integer  "trackable_id",   null: false
    t.string   "trackable_type", null: false
    t.string   "action",         null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "activities", ["recipient_id", "recipient_type"], name: "index_activities_on_recipient_id_and_recipient_type", using: :btree
  add_index "activities", ["trackable_id", "trackable_type"], name: "index_activities_on_trackable_id_and_trackable_type", using: :btree

  create_table "channel_groups", force: true do |t|
    t.string   "sc_chgid",    null: false
    t.string   "name",        null: false
    t.integer  "sort_number"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "channel_groups", ["sc_chgid"], name: "index_channel_groups_on_sc_chgid", unique: true, using: :btree

  create_table "channel_works", force: true do |t|
    t.integer  "user_id",    null: false
    t.integer  "work_id",    null: false
    t.integer  "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "channel_works", ["user_id", "work_id", "channel_id"], name: "index_channel_works_on_user_id_and_work_id_and_channel_id", unique: true, using: :btree
  add_index "channel_works", ["user_id", "work_id"], name: "index_channel_works_on_user_id_and_work_id", using: :btree

  create_table "channels", force: true do |t|
    t.integer  "channel_group_id",                null: false
    t.integer  "sc_chid",                         null: false
    t.string   "name",                            null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean  "published",        default: true, null: false
  end

  add_index "channels", ["published"], name: "index_channels_on_published", using: :btree
  add_index "channels", ["sc_chid"], name: "index_channels_on_sc_chid", unique: true, using: :btree

  create_table "checkins", force: true do |t|
    t.integer  "user_id",                              null: false
    t.integer  "episode_id",                           null: false
    t.text     "comment"
    t.string   "twitter_url_hash"
    t.integer  "twitter_click_count",  default: 0,     null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean  "spoil",                default: false, null: false
    t.string   "facebook_url_hash"
    t.integer  "facebook_click_count", default: 0,     null: false
    t.integer  "comments_count",       default: 0,     null: false
    t.integer  "likes_count",          default: 0,     null: false
    t.boolean  "modify_comment",       default: false, null: false
    t.boolean  "shared_twitter",       default: false, null: false
    t.boolean  "shared_facebook",      default: false, null: false
  end

  add_index "checkins", ["facebook_url_hash"], name: "index_checkins_on_facebook_url_hash", unique: true, using: :btree
  add_index "checkins", ["twitter_url_hash"], name: "index_checkins_on_twitter_url_hash", unique: true, using: :btree

  create_table "comments", force: true do |t|
    t.integer  "user_id",                 null: false
    t.integer  "checkin_id",              null: false
    t.text     "body",                    null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "likes_count", default: 0, null: false
  end

  create_table "cover_images", force: true do |t|
    t.integer  "work_id",    null: false
    t.string   "file_name",  null: false
    t.string   "location",   null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  create_table "episodes", force: true do |t|
    t.integer  "work_id",                        null: false
    t.string   "number"
    t.integer  "sort_number",    default: 0,     null: false
    t.string   "title"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "checkins_count", default: 0,     null: false
    t.boolean  "single",         default: false
    t.integer  "sc_count"
  end

  add_index "episodes", ["checkins_count"], name: "index_episodes_on_checkins_count", using: :btree
  add_index "episodes", ["work_id", "sc_count"], name: "index_episodes_on_work_id_and_sc_count", unique: true, using: :btree

  create_table "finished_tips", force: true do |t|
    t.integer  "user_id",    null: false
    t.integer  "tip_id",     null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  add_index "finished_tips", ["user_id", "tip_id"], name: "index_finished_tips_on_user_id_and_tip_id", unique: true, using: :btree

  create_table "follows", force: true do |t|
    t.integer  "user_id",      null: false
    t.integer  "following_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "follows", ["user_id", "following_id"], name: "index_follows_on_user_id_and_following_id", unique: true, using: :btree

  create_table "items", force: true do |t|
    t.integer  "work_id"
    t.string   "name",                       null: false
    t.string   "url",                        null: false
    t.string   "image_uid",                  null: false
    t.boolean  "main",       default: false, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  create_table "likes", force: true do |t|
    t.integer  "user_id",        null: false
    t.integer  "recipient_id",   null: false
    t.string   "recipient_type", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "likes", ["recipient_id", "recipient_type"], name: "index_likes_on_recipient_id_and_recipient_type", using: :btree

  create_table "notifications", force: true do |t|
    t.integer  "user_id",                        null: false
    t.integer  "action_user_id",                 null: false
    t.integer  "trackable_id",                   null: false
    t.string   "trackable_type",                 null: false
    t.string   "action",                         null: false
    t.boolean  "read",           default: false, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "notifications", ["read"], name: "index_notifications_on_read", using: :btree
  add_index "notifications", ["trackable_id", "trackable_type"], name: "index_notifications_on_trackable_id_and_trackable_type", using: :btree

  create_table "profiles", force: true do |t|
    t.integer  "user_id",                           null: false
    t.string   "name",                 default: "", null: false
    t.string   "description",          default: "", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string   "avatar_uid"
    t.string   "background_image_uid"
  end

  add_index "profiles", ["user_id"], name: "index_profiles_on_user_id", unique: true, using: :btree

  create_table "programs", force: true do |t|
    t.integer  "channel_id",     null: false
    t.integer  "episode_id",     null: false
    t.integer  "work_id",        null: false
    t.datetime "started_at",     null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.datetime "sc_last_update"
  end

  add_index "programs", ["channel_id", "episode_id"], name: "index_programs_on_channel_id_and_episode_id", unique: true, using: :btree

  create_table "providers", force: true do |t|
    t.integer  "user_id",          null: false
    t.string   "name",             null: false
    t.string   "uid",              null: false
    t.string   "token",            null: false
    t.integer  "token_expires_at"
    t.string   "token_secret"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "providers", ["name", "uid"], name: "index_providers_on_name_and_uid", unique: true, using: :btree

  create_table "receptions", force: true do |t|
    t.integer  "user_id",    null: false
    t.integer  "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "receptions", ["user_id", "channel_id"], name: "index_receptions_on_user_id_and_channel_id", unique: true, using: :btree

  create_table "seasons", force: true do |t|
    t.string   "name",       null: false
    t.string   "slug",       null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "seasons", ["slug"], name: "index_seasons_on_slug", unique: true, using: :btree

  create_table "sessions", force: true do |t|
    t.string   "session_id", null: false
    t.text     "data"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "sessions", ["session_id"], name: "index_sessions_on_session_id", unique: true, using: :btree
  add_index "sessions", ["updated_at"], name: "index_sessions_on_updated_at", using: :btree

  create_table "staffs", force: true do |t|
    t.string   "email",              default: "", null: false
    t.string   "encrypted_password", default: "", null: false
    t.integer  "sign_in_count",      default: 0,  null: false
    t.datetime "current_sign_in_at"
    t.datetime "last_sign_in_at"
    t.string   "current_sign_in_ip"
    t.string   "last_sign_in_ip"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "staffs", ["email"], name: "index_staffs_on_email", unique: true, using: :btree

  create_table "statuses", force: true do |t|
    t.integer  "user_id",                     null: false
    t.integer  "work_id",                     null: false
    t.integer  "kind",                        null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean  "latest",      default: false, null: false
    t.integer  "likes_count", default: 0,     null: false
  end

  add_index "statuses", ["user_id", "latest"], name: "index_statuses_on_user_id_and_latest", using: :btree

  create_table "syobocal_alerts", force: true do |t|
    t.integer  "work_id"
    t.integer  "kind",            null: false
    t.integer  "sc_prog_item_id"
    t.string   "sc_sub_title"
    t.string   "sc_prog_comment"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "syobocal_alerts", ["kind"], name: "index_syobocal_alerts_on_kind", using: :btree
  add_index "syobocal_alerts", ["sc_prog_item_id"], name: "index_syobocal_alerts_on_sc_prog_item_id", using: :btree

  create_table "tips", force: true do |t|
    t.integer  "target",       null: false
    t.string   "partial_name", null: false
    t.string   "title",        null: false
    t.string   "icon_name",    null: false
    t.datetime "created_at",   null: false
    t.datetime "updated_at",   null: false
  end

  add_index "tips", ["partial_name"], name: "index_tips_on_partial_name", unique: true, using: :btree

  create_table "twitter_bots", force: true do |t|
    t.string   "name",       null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "twitter_bots", ["name"], name: "index_twitter_bots_on_name", unique: true, using: :btree

  create_table "users", force: true do |t|
    t.string   "username",                             null: false
    t.string   "email",                                null: false
    t.string   "encrypted_password",   default: "",    null: false
    t.datetime "remember_created_at"
    t.integer  "sign_in_count",        default: 0,     null: false
    t.datetime "current_sign_in_at"
    t.datetime "last_sign_in_at"
    t.string   "current_sign_in_ip"
    t.string   "last_sign_in_ip"
    t.string   "confirmation_token"
    t.datetime "confirmed_at"
    t.datetime "confirmation_sent_at"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string   "unconfirmed_email"
    t.integer  "role",                                 null: false
    t.integer  "checkins_count",       default: 0,     null: false
    t.integer  "notifications_count",  default: 0,     null: false
    t.boolean  "share_checkin",        default: false
  end

  add_index "users", ["confirmation_token"], name: "index_users_on_confirmation_token", unique: true, using: :btree
  add_index "users", ["email"], name: "index_users_on_email", unique: true, using: :btree
  add_index "users", ["role"], name: "index_users_on_role", using: :btree
  add_index "users", ["username"], name: "index_users_on_username", unique: true, using: :btree

  create_table "versions", force: true do |t|
    t.string   "item_type",  null: false
    t.integer  "item_id",    null: false
    t.string   "event",      null: false
    t.string   "whodunnit"
    t.text     "object"
    t.datetime "created_at"
  end

  add_index "versions", ["item_type", "item_id"], name: "index_versions_on_item_type_and_item_id", using: :btree

  create_table "works", force: true do |t|
    t.string   "title",                             null: false
    t.integer  "media",                             null: false
    t.string   "official_site_url", default: "",    null: false
    t.string   "wikipedia_url",     default: "",    null: false
    t.date     "released_at"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "episodes_count",    default: 0,     null: false
    t.integer  "season_id"
    t.boolean  "on_air",            default: false, null: false
    t.string   "twitter_username"
    t.string   "twitter_hashtag"
    t.integer  "watchers_count",    default: 0,     null: false
    t.integer  "sc_tid"
    t.boolean  "fetch_syobocal",    default: false, null: false
    t.datetime "nicoch_started_at"
  end

  add_index "works", ["episodes_count"], name: "index_works_on_episodes_count", using: :btree
  add_index "works", ["media"], name: "index_works_on_media", using: :btree
  add_index "works", ["on_air"], name: "index_works_on_on_air", using: :btree
  add_index "works", ["released_at"], name: "index_works_on_released_at", using: :btree
  add_index "works", ["sc_tid"], name: "index_works_on_sc_tid", unique: true, using: :btree
  add_index "works", ["watchers_count"], name: "index_works_on_watchers_count", using: :btree

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

  add_foreign_key "finished_tips", "tips", name: "finished_tips_tip_id_fk", dependent: :delete
  add_foreign_key "finished_tips", "users", name: "finished_tips_user_id_fk", dependent: :delete

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

end
