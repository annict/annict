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

ActiveRecord::Schema.define(version: 20151013145503) do

  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"

  create_table "activities", force: :cascade do |t|
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

  create_table "channel_groups", force: :cascade do |t|
    t.string   "sc_chgid",    null: false
    t.string   "name",        null: false
    t.integer  "sort_number"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "channel_groups", ["sc_chgid"], name: "index_channel_groups_on_sc_chgid", unique: true, using: :btree

  create_table "channel_works", force: :cascade do |t|
    t.integer  "user_id",    null: false
    t.integer  "work_id",    null: false
    t.integer  "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "channel_works", ["user_id", "work_id", "channel_id"], name: "index_channel_works_on_user_id_and_work_id_and_channel_id", unique: true, using: :btree
  add_index "channel_works", ["user_id", "work_id"], name: "index_channel_works_on_user_id_and_work_id", using: :btree

  create_table "channels", force: :cascade do |t|
    t.integer  "channel_group_id",                null: false
    t.integer  "sc_chid",                         null: false
    t.string   "name",                            null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean  "published",        default: true, null: false
  end

  add_index "channels", ["published"], name: "index_channels_on_published", using: :btree
  add_index "channels", ["sc_chid"], name: "index_channels_on_sc_chid", unique: true, using: :btree

  create_table "checkins", force: :cascade do |t|
    t.integer  "user_id",                              null: false
    t.integer  "episode_id",                           null: false
    t.text     "comment"
    t.string   "twitter_url_hash"
    t.integer  "twitter_click_count",  default: 0,     null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string   "facebook_url_hash"
    t.integer  "facebook_click_count", default: 0,     null: false
    t.integer  "comments_count",       default: 0,     null: false
    t.integer  "likes_count",          default: 0,     null: false
    t.boolean  "modify_comment",       default: false, null: false
    t.boolean  "shared_twitter",       default: false, null: false
    t.boolean  "shared_facebook",      default: false, null: false
    t.integer  "work_id"
  end

  add_index "checkins", ["facebook_url_hash"], name: "index_checkins_on_facebook_url_hash", unique: true, using: :btree
  add_index "checkins", ["twitter_url_hash"], name: "index_checkins_on_twitter_url_hash", unique: true, using: :btree
  add_index "checkins", ["work_id"], name: "index_checkins_on_work_id", using: :btree

  create_table "checks", force: :cascade do |t|
    t.integer  "user_id",                          null: false
    t.integer  "work_id",                          null: false
    t.integer  "episode_id"
    t.integer  "skipped_episode_ids", default: [], null: false, array: true
    t.integer  "position",            default: 0,  null: false
    t.datetime "created_at",                       null: false
    t.datetime "updated_at",                       null: false
  end

  add_index "checks", ["user_id", "position"], name: "index_checks_on_user_id_and_position", using: :btree
  add_index "checks", ["user_id", "work_id"], name: "index_checks_on_user_id_and_work_id", unique: true, using: :btree
  add_index "checks", ["user_id"], name: "index_checks_on_user_id", using: :btree

  create_table "comments", force: :cascade do |t|
    t.integer  "user_id",                 null: false
    t.integer  "checkin_id",              null: false
    t.text     "body",                    null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "likes_count", default: 0, null: false
    t.integer  "work_id"
  end

  add_index "comments", ["work_id"], name: "index_comments_on_work_id", using: :btree

  create_table "cover_images", force: :cascade do |t|
    t.integer  "work_id",    null: false
    t.string   "file_name",  null: false
    t.string   "location",   null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  create_table "db_activities", force: :cascade do |t|
    t.integer  "user_id",        null: false
    t.integer  "recipient_id"
    t.string   "recipient_type"
    t.integer  "trackable_id",   null: false
    t.string   "trackable_type", null: false
    t.string   "action",         null: false
    t.json     "parameters"
    t.datetime "created_at",     null: false
    t.datetime "updated_at",     null: false
  end

  add_index "db_activities", ["recipient_id", "recipient_type"], name: "index_db_activities_on_recipient_id_and_recipient_type", using: :btree
  add_index "db_activities", ["trackable_id", "trackable_type"], name: "index_db_activities_on_trackable_id_and_trackable_type", using: :btree

  create_table "delayed_jobs", force: :cascade do |t|
    t.integer  "priority",   default: 0, null: false
    t.integer  "attempts",   default: 0, null: false
    t.text     "handler",                null: false
    t.text     "last_error"
    t.datetime "run_at"
    t.datetime "locked_at"
    t.datetime "failed_at"
    t.string   "locked_by"
    t.string   "queue"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "delayed_jobs", ["priority", "run_at"], name: "delayed_jobs_priority", using: :btree

  create_table "draft_episodes", force: :cascade do |t|
    t.integer  "episode_id",                  null: false
    t.integer  "work_id",                     null: false
    t.string   "number"
    t.integer  "sort_number",     default: 0, null: false
    t.string   "title"
    t.datetime "created_at",                  null: false
    t.datetime "updated_at",                  null: false
    t.integer  "prev_episode_id"
  end

  add_index "draft_episodes", ["episode_id"], name: "index_draft_episodes_on_episode_id", using: :btree
  add_index "draft_episodes", ["prev_episode_id"], name: "index_draft_episodes_on_prev_episode_id", using: :btree
  add_index "draft_episodes", ["work_id"], name: "index_draft_episodes_on_work_id", using: :btree

  create_table "draft_items", force: :cascade do |t|
    t.integer  "item_id"
    t.integer  "work_id",                                  null: false
    t.string   "name",                                     null: false
    t.string   "url",                                      null: false
    t.boolean  "main",                     default: false, null: false
    t.string   "tombo_image_file_name",                    null: false
    t.string   "tombo_image_content_type",                 null: false
    t.integer  "tombo_image_file_size",                    null: false
    t.datetime "tombo_image_updated_at",                   null: false
    t.datetime "created_at",                               null: false
    t.datetime "updated_at",                               null: false
  end

  add_index "draft_items", ["item_id"], name: "index_draft_items_on_item_id", using: :btree
  add_index "draft_items", ["work_id"], name: "index_draft_items_on_work_id", using: :btree

  create_table "draft_multiple_episodes", force: :cascade do |t|
    t.integer  "work_id",    null: false
    t.text     "body",       null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  add_index "draft_multiple_episodes", ["work_id"], name: "index_draft_multiple_episodes_on_work_id", using: :btree

  create_table "draft_programs", force: :cascade do |t|
    t.integer  "program_id"
    t.integer  "channel_id", null: false
    t.integer  "episode_id", null: false
    t.integer  "work_id",    null: false
    t.datetime "started_at", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  add_index "draft_programs", ["channel_id"], name: "index_draft_programs_on_channel_id", using: :btree
  add_index "draft_programs", ["episode_id"], name: "index_draft_programs_on_episode_id", using: :btree
  add_index "draft_programs", ["program_id"], name: "index_draft_programs_on_program_id", using: :btree
  add_index "draft_programs", ["work_id"], name: "index_draft_programs_on_work_id", using: :btree

  create_table "draft_works", force: :cascade do |t|
    t.integer  "work_id"
    t.integer  "season_id"
    t.integer  "sc_tid"
    t.string   "title",                          null: false
    t.integer  "media",                          null: false
    t.string   "official_site_url", default: "", null: false
    t.string   "wikipedia_url",     default: "", null: false
    t.date     "released_at"
    t.string   "twitter_username"
    t.string   "twitter_hashtag"
    t.string   "released_at_about"
    t.datetime "created_at",                     null: false
    t.datetime "updated_at",                     null: false
  end

  add_index "draft_works", ["sc_tid"], name: "index_draft_works_on_sc_tid", using: :btree
  add_index "draft_works", ["season_id"], name: "index_draft_works_on_season_id", using: :btree
  add_index "draft_works", ["work_id"], name: "index_draft_works_on_work_id", using: :btree

  create_table "edit_request_comments", force: :cascade do |t|
    t.integer  "edit_request_id", null: false
    t.integer  "user_id",         null: false
    t.text     "body",            null: false
    t.datetime "created_at",      null: false
    t.datetime "updated_at",      null: false
  end

  add_index "edit_request_comments", ["edit_request_id"], name: "index_edit_request_comments_on_edit_request_id", using: :btree
  add_index "edit_request_comments", ["user_id"], name: "index_edit_request_comments_on_user_id", using: :btree

  create_table "edit_request_participants", force: :cascade do |t|
    t.integer  "edit_request_id", null: false
    t.integer  "user_id",         null: false
    t.datetime "created_at",      null: false
    t.datetime "updated_at",      null: false
  end

  add_index "edit_request_participants", ["edit_request_id", "user_id"], name: "index_edit_request_participants_on_edit_request_id_and_user_id", unique: true, using: :btree
  add_index "edit_request_participants", ["edit_request_id"], name: "index_edit_request_participants_on_edit_request_id", using: :btree
  add_index "edit_request_participants", ["user_id"], name: "index_edit_request_participants_on_user_id", using: :btree

  create_table "edit_requests", force: :cascade do |t|
    t.integer  "user_id",                                null: false
    t.integer  "draft_resource_id",                      null: false
    t.string   "draft_resource_type",                    null: false
    t.string   "title",                                  null: false
    t.text     "body"
    t.string   "aasm_state",          default: "opened", null: false
    t.datetime "published_at"
    t.datetime "closed_at"
    t.datetime "created_at",                             null: false
    t.datetime "updated_at",                             null: false
  end

  add_index "edit_requests", ["draft_resource_id", "draft_resource_type"], name: "index_er_on_drid_and_drtype", using: :btree
  add_index "edit_requests", ["user_id"], name: "index_edit_requests_on_user_id", using: :btree

  create_table "episodes", force: :cascade do |t|
    t.integer  "work_id",                               null: false
    t.string   "number"
    t.integer  "sort_number",     default: 0,           null: false
    t.string   "title"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "checkins_count",  default: 0,           null: false
    t.integer  "sc_count"
    t.integer  "prev_episode_id"
    t.string   "aasm_state",      default: "published", null: false
  end

  add_index "episodes", ["aasm_state"], name: "index_episodes_on_aasm_state", using: :btree
  add_index "episodes", ["checkins_count"], name: "index_episodes_on_checkins_count", using: :btree
  add_index "episodes", ["prev_episode_id"], name: "index_episodes_on_prev_episode_id", using: :btree
  add_index "episodes", ["work_id", "sc_count"], name: "index_episodes_on_work_id_and_sc_count", unique: true, using: :btree

  create_table "finished_tips", force: :cascade do |t|
    t.integer  "user_id",    null: false
    t.integer  "tip_id",     null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  add_index "finished_tips", ["user_id", "tip_id"], name: "index_finished_tips_on_user_id_and_tip_id", unique: true, using: :btree

  create_table "follows", force: :cascade do |t|
    t.integer  "user_id",      null: false
    t.integer  "following_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "follows", ["user_id", "following_id"], name: "index_follows_on_user_id_and_following_id", unique: true, using: :btree

  create_table "items", force: :cascade do |t|
    t.integer  "work_id"
    t.string   "name",                     null: false
    t.string   "url",                      null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string   "tombo_image_file_name"
    t.string   "tombo_image_content_type"
    t.integer  "tombo_image_file_size"
    t.datetime "tombo_image_updated_at"
  end

  add_index "items", ["work_id"], name: "index_items_on_work_id", unique: true, using: :btree

  create_table "likes", force: :cascade do |t|
    t.integer  "user_id",        null: false
    t.integer  "recipient_id",   null: false
    t.string   "recipient_type", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "likes", ["recipient_id", "recipient_type"], name: "index_likes_on_recipient_id_and_recipient_type", using: :btree

  create_table "notifications", force: :cascade do |t|
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

  create_table "oauth_access_grants", force: :cascade do |t|
    t.integer  "resource_owner_id", null: false
    t.integer  "application_id",    null: false
    t.string   "token",             null: false
    t.integer  "expires_in",        null: false
    t.text     "redirect_uri",      null: false
    t.datetime "created_at",        null: false
    t.datetime "revoked_at"
    t.string   "scopes"
  end

  add_index "oauth_access_grants", ["token"], name: "index_oauth_access_grants_on_token", unique: true, using: :btree

  create_table "oauth_access_tokens", force: :cascade do |t|
    t.integer  "resource_owner_id"
    t.integer  "application_id"
    t.string   "token",             null: false
    t.string   "refresh_token"
    t.integer  "expires_in"
    t.datetime "revoked_at"
    t.datetime "created_at",        null: false
    t.string   "scopes"
  end

  add_index "oauth_access_tokens", ["refresh_token"], name: "index_oauth_access_tokens_on_refresh_token", unique: true, using: :btree
  add_index "oauth_access_tokens", ["resource_owner_id"], name: "index_oauth_access_tokens_on_resource_owner_id", using: :btree
  add_index "oauth_access_tokens", ["token"], name: "index_oauth_access_tokens_on_token", unique: true, using: :btree

  create_table "oauth_applications", force: :cascade do |t|
    t.string   "name",                      null: false
    t.string   "uid",                       null: false
    t.string   "secret",                    null: false
    t.text     "redirect_uri",              null: false
    t.string   "scopes",       default: "", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "oauth_applications", ["uid"], name: "index_oauth_applications_on_uid", unique: true, using: :btree

  create_table "profiles", force: :cascade do |t|
    t.integer  "user_id",                                             null: false
    t.string   "name",                                default: "",    null: false
    t.string   "description",                         default: "",    null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean  "background_image_animated",           default: false, null: false
    t.string   "tombo_avatar_file_name"
    t.string   "tombo_avatar_content_type"
    t.integer  "tombo_avatar_file_size"
    t.datetime "tombo_avatar_updated_at"
    t.string   "tombo_background_image_file_name"
    t.string   "tombo_background_image_content_type"
    t.integer  "tombo_background_image_file_size"
    t.datetime "tombo_background_image_updated_at"
  end

  add_index "profiles", ["user_id"], name: "index_profiles_on_user_id", unique: true, using: :btree

  create_table "programs", force: :cascade do |t|
    t.integer  "channel_id",     null: false
    t.integer  "episode_id",     null: false
    t.integer  "work_id",        null: false
    t.datetime "started_at",     null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.datetime "sc_last_update"
  end

  add_index "programs", ["channel_id", "episode_id"], name: "index_programs_on_channel_id_and_episode_id", unique: true, using: :btree

  create_table "providers", force: :cascade do |t|
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

  create_table "receptions", force: :cascade do |t|
    t.integer  "user_id",    null: false
    t.integer  "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "receptions", ["user_id", "channel_id"], name: "index_receptions_on_user_id_and_channel_id", unique: true, using: :btree

  create_table "seasons", force: :cascade do |t|
    t.string   "name",       null: false
    t.string   "slug",       null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "seasons", ["slug"], name: "index_seasons_on_slug", unique: true, using: :btree

  create_table "sessions", force: :cascade do |t|
    t.string   "session_id", null: false
    t.text     "data"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "sessions", ["session_id"], name: "index_sessions_on_session_id", unique: true, using: :btree
  add_index "sessions", ["updated_at"], name: "index_sessions_on_updated_at", using: :btree

  create_table "settings", force: :cascade do |t|
    t.integer  "user_id",                                  null: false
    t.boolean  "hide_checkin_comment",     default: true,  null: false
    t.datetime "created_at",                               null: false
    t.datetime "updated_at",                               null: false
    t.boolean  "share_record_to_twitter",  default: false
    t.boolean  "share_record_to_facebook", default: false
  end

  add_index "settings", ["user_id"], name: "index_settings_on_user_id", using: :btree

  create_table "statuses", force: :cascade do |t|
    t.integer  "user_id",                     null: false
    t.integer  "work_id",                     null: false
    t.integer  "kind",                        null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean  "latest",      default: false, null: false
    t.integer  "likes_count", default: 0,     null: false
  end

  add_index "statuses", ["user_id", "latest"], name: "index_statuses_on_user_id_and_latest", using: :btree

  create_table "syobocal_alerts", force: :cascade do |t|
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

  create_table "tips", force: :cascade do |t|
    t.integer  "target",       null: false
    t.string   "partial_name", null: false
    t.string   "title",        null: false
    t.string   "icon_name",    null: false
    t.datetime "created_at",   null: false
    t.datetime "updated_at",   null: false
  end

  add_index "tips", ["partial_name"], name: "index_tips_on_partial_name", unique: true, using: :btree

  create_table "twitter_bots", force: :cascade do |t|
    t.string   "name",       null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "twitter_bots", ["name"], name: "index_twitter_bots_on_name", unique: true, using: :btree

  create_table "users", force: :cascade do |t|
    t.string   "username",                          null: false
    t.string   "email",                             null: false
    t.string   "encrypted_password",   default: "", null: false
    t.datetime "remember_created_at"
    t.integer  "sign_in_count",        default: 0,  null: false
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
    t.integer  "role",                              null: false
    t.integer  "checkins_count",       default: 0,  null: false
    t.integer  "notifications_count",  default: 0,  null: false
  end

  add_index "users", ["confirmation_token"], name: "index_users_on_confirmation_token", unique: true, using: :btree
  add_index "users", ["email"], name: "index_users_on_email", unique: true, using: :btree
  add_index "users", ["role"], name: "index_users_on_role", using: :btree
  add_index "users", ["username"], name: "index_users_on_username", unique: true, using: :btree

  create_table "versions", force: :cascade do |t|
    t.string   "item_type",  null: false
    t.integer  "item_id",    null: false
    t.string   "event",      null: false
    t.string   "whodunnit"
    t.text     "object"
    t.datetime "created_at"
  end

  add_index "versions", ["item_type", "item_id"], name: "index_versions_on_item_type_and_item_id", using: :btree

  create_table "works", force: :cascade do |t|
    t.string   "title",                                   null: false
    t.integer  "media",                                   null: false
    t.string   "official_site_url", default: "",          null: false
    t.string   "wikipedia_url",     default: "",          null: false
    t.date     "released_at"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "episodes_count",    default: 0,           null: false
    t.integer  "season_id"
    t.string   "twitter_username"
    t.string   "twitter_hashtag"
    t.integer  "watchers_count",    default: 0,           null: false
    t.integer  "sc_tid"
    t.boolean  "fetch_syobocal",    default: false,       null: false
    t.string   "released_at_about"
    t.string   "aasm_state",        default: "published", null: false
  end

  add_index "works", ["aasm_state"], name: "index_works_on_aasm_state", using: :btree
  add_index "works", ["episodes_count"], name: "index_works_on_episodes_count", using: :btree
  add_index "works", ["media"], name: "index_works_on_media", using: :btree
  add_index "works", ["released_at"], name: "index_works_on_released_at", using: :btree
  add_index "works", ["sc_tid"], name: "index_works_on_sc_tid", unique: true, using: :btree
  add_index "works", ["watchers_count"], name: "index_works_on_watchers_count", using: :btree

  add_foreign_key "activities", "users"
  add_foreign_key "channel_works", "channels"
  add_foreign_key "channel_works", "users"
  add_foreign_key "channel_works", "works"
  add_foreign_key "channels", "channel_groups"
  add_foreign_key "checkins", "episodes"
  add_foreign_key "checkins", "users"
  add_foreign_key "checkins", "works"
  add_foreign_key "checks", "episodes"
  add_foreign_key "checks", "users"
  add_foreign_key "checks", "works"
  add_foreign_key "comments", "checkins"
  add_foreign_key "comments", "users"
  add_foreign_key "comments", "works"
  add_foreign_key "cover_images", "works"
  add_foreign_key "db_activities", "users"
  add_foreign_key "draft_episodes", "episodes"
  add_foreign_key "draft_episodes", "episodes", column: "prev_episode_id"
  add_foreign_key "draft_episodes", "works"
  add_foreign_key "draft_items", "items"
  add_foreign_key "draft_items", "works"
  add_foreign_key "draft_multiple_episodes", "works"
  add_foreign_key "draft_programs", "channels"
  add_foreign_key "draft_programs", "episodes"
  add_foreign_key "draft_programs", "programs"
  add_foreign_key "draft_programs", "works"
  add_foreign_key "draft_works", "seasons"
  add_foreign_key "draft_works", "works"
  add_foreign_key "edit_request_comments", "edit_requests", on_delete: :cascade
  add_foreign_key "edit_request_comments", "users", on_delete: :cascade
  add_foreign_key "edit_request_participants", "edit_requests"
  add_foreign_key "edit_request_participants", "users"
  add_foreign_key "edit_requests", "users", on_delete: :cascade
  add_foreign_key "episodes", "episodes", column: "prev_episode_id"
  add_foreign_key "episodes", "works"
  add_foreign_key "finished_tips", "tips"
  add_foreign_key "finished_tips", "users"
  add_foreign_key "follows", "users"
  add_foreign_key "follows", "users", column: "following_id"
  add_foreign_key "items", "works"
  add_foreign_key "likes", "users"
  add_foreign_key "notifications", "users"
  add_foreign_key "notifications", "users", column: "action_user_id"
  add_foreign_key "profiles", "users"
  add_foreign_key "programs", "channels"
  add_foreign_key "programs", "episodes"
  add_foreign_key "programs", "works"
  add_foreign_key "providers", "users"
  add_foreign_key "receptions", "channels"
  add_foreign_key "receptions", "users"
  add_foreign_key "settings", "users"
  add_foreign_key "statuses", "users"
  add_foreign_key "statuses", "works"
  add_foreign_key "syobocal_alerts", "works"
  add_foreign_key "works", "seasons"
end
