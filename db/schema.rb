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

ActiveRecord::Schema.define(version: 20160320200438) do

  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"
  enable_extension "pg_stat_statements"

  create_table "activities", force: :cascade do |t|
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

  create_table "casts", force: :cascade do |t|
    t.integer  "person_id",                         null: false
    t.integer  "work_id",                           null: false
    t.string   "name",                              null: false
    t.string   "part",                              null: false
    t.string   "aasm_state",  default: "published", null: false
    t.integer  "sort_number", default: 0,           null: false
    t.datetime "created_at",                        null: false
    t.datetime "updated_at",                        null: false
  end

  add_index "casts", ["aasm_state"], name: "index_casts_on_aasm_state", using: :btree
  add_index "casts", ["person_id"], name: "index_casts_on_person_id", using: :btree
  add_index "casts", ["sort_number"], name: "index_casts_on_sort_number", using: :btree
  add_index "casts", ["work_id"], name: "index_casts_on_work_id", using: :btree

  create_table "channel_groups", force: :cascade do |t|
    t.string   "sc_chgid",    limit: 510, null: false
    t.string   "name",        limit: 510, null: false
    t.integer  "sort_number"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "channel_groups", ["sc_chgid"], name: "channel_groups_sc_chgid_key", unique: true, using: :btree

  create_table "channel_works", force: :cascade do |t|
    t.integer  "user_id",    null: false
    t.integer  "work_id",    null: false
    t.integer  "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "channel_works", ["channel_id"], name: "channel_works_channel_id_idx", using: :btree
  add_index "channel_works", ["user_id", "work_id", "channel_id"], name: "channel_works_user_id_work_id_channel_id_key", unique: true, using: :btree
  add_index "channel_works", ["user_id"], name: "channel_works_user_id_idx", using: :btree
  add_index "channel_works", ["work_id"], name: "channel_works_work_id_idx", using: :btree

  create_table "channels", force: :cascade do |t|
    t.integer  "channel_group_id",                null: false
    t.integer  "sc_chid",                         null: false
    t.string   "name",                            null: false
    t.boolean  "published",        default: true, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "channels", ["channel_group_id"], name: "channels_channel_group_id_idx", using: :btree
  add_index "channels", ["sc_chid"], name: "channels_sc_chid_key", unique: true, using: :btree

  create_table "checkins", force: :cascade do |t|
    t.integer  "user_id",                                          null: false
    t.integer  "episode_id",                                       null: false
    t.text     "comment"
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
    t.integer  "work_id"
    t.float    "rating"
    t.integer  "multiple_record_id"
  end

  add_index "checkins", ["episode_id"], name: "checkins_episode_id_idx", using: :btree
  add_index "checkins", ["facebook_url_hash"], name: "checkins_facebook_url_hash_key", unique: true, using: :btree
  add_index "checkins", ["multiple_record_id"], name: "index_checkins_on_multiple_record_id", using: :btree
  add_index "checkins", ["twitter_url_hash"], name: "checkins_twitter_url_hash_key", unique: true, using: :btree
  add_index "checkins", ["user_id"], name: "checkins_user_id_idx", using: :btree
  add_index "checkins", ["work_id"], name: "index_checkins_on_work_id", using: :btree

  create_table "comments", force: :cascade do |t|
    t.integer  "user_id",                 null: false
    t.integer  "checkin_id",              null: false
    t.text     "body",                    null: false
    t.integer  "likes_count", default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "work_id"
  end

  add_index "comments", ["checkin_id"], name: "comments_checkin_id_idx", using: :btree
  add_index "comments", ["user_id"], name: "comments_user_id_idx", using: :btree
  add_index "comments", ["work_id"], name: "index_comments_on_work_id", using: :btree

  create_table "cover_images", force: :cascade do |t|
    t.integer  "work_id",                null: false
    t.string   "file_name",  limit: 510, null: false
    t.string   "location",   limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "cover_images", ["work_id"], name: "cover_images_work_id_idx", using: :btree

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

  create_table "draft_casts", force: :cascade do |t|
    t.integer  "cast_id"
    t.integer  "person_id",               null: false
    t.integer  "work_id",                 null: false
    t.string   "name",                    null: false
    t.string   "part",                    null: false
    t.integer  "sort_number", default: 0, null: false
    t.datetime "created_at",              null: false
    t.datetime "updated_at",              null: false
  end

  add_index "draft_casts", ["cast_id"], name: "index_draft_casts_on_cast_id", using: :btree
  add_index "draft_casts", ["person_id"], name: "index_draft_casts_on_person_id", using: :btree
  add_index "draft_casts", ["sort_number"], name: "index_draft_casts_on_sort_number", using: :btree
  add_index "draft_casts", ["work_id"], name: "index_draft_casts_on_work_id", using: :btree

  create_table "draft_episodes", force: :cascade do |t|
    t.integer  "episode_id",                      null: false
    t.integer  "work_id",                         null: false
    t.string   "number"
    t.integer  "sort_number",     default: 0,     null: false
    t.string   "title"
    t.datetime "created_at",                      null: false
    t.datetime "updated_at",                      null: false
    t.integer  "prev_episode_id"
    t.boolean  "fetch_syobocal",  default: false, null: false
    t.string   "raw_number"
    t.integer  "sc_count"
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

  create_table "draft_organizations", force: :cascade do |t|
    t.integer  "organization_id"
    t.string   "name",             null: false
    t.string   "url"
    t.string   "wikipedia_url"
    t.string   "twitter_username"
    t.datetime "created_at",       null: false
    t.datetime "updated_at",       null: false
  end

  add_index "draft_organizations", ["name"], name: "index_draft_organizations_on_name", using: :btree
  add_index "draft_organizations", ["organization_id"], name: "index_draft_organizations_on_organization_id", using: :btree

  create_table "draft_people", force: :cascade do |t|
    t.integer  "person_id"
    t.integer  "prefecture_id"
    t.string   "name",             null: false
    t.string   "name_kana"
    t.string   "nickname"
    t.string   "gender"
    t.string   "url"
    t.string   "wikipedia_url"
    t.string   "twitter_username"
    t.date     "birthday"
    t.string   "blood_type"
    t.integer  "height"
    t.datetime "created_at",       null: false
    t.datetime "updated_at",       null: false
  end

  add_index "draft_people", ["name"], name: "index_draft_people_on_name", using: :btree
  add_index "draft_people", ["person_id"], name: "index_draft_people_on_person_id", using: :btree
  add_index "draft_people", ["prefecture_id"], name: "index_draft_people_on_prefecture_id", using: :btree

  create_table "draft_programs", force: :cascade do |t|
    t.integer  "program_id"
    t.integer  "channel_id",                  null: false
    t.integer  "episode_id",                  null: false
    t.integer  "work_id",                     null: false
    t.datetime "started_at",                  null: false
    t.datetime "created_at",                  null: false
    t.datetime "updated_at",                  null: false
    t.boolean  "rebroadcast", default: false, null: false
  end

  add_index "draft_programs", ["channel_id"], name: "index_draft_programs_on_channel_id", using: :btree
  add_index "draft_programs", ["episode_id"], name: "index_draft_programs_on_episode_id", using: :btree
  add_index "draft_programs", ["program_id"], name: "index_draft_programs_on_program_id", using: :btree
  add_index "draft_programs", ["work_id"], name: "index_draft_programs_on_work_id", using: :btree

  create_table "draft_staffs", force: :cascade do |t|
    t.integer  "staff_id"
    t.integer  "work_id",                   null: false
    t.string   "name",                      null: false
    t.string   "role",                      null: false
    t.string   "role_other"
    t.integer  "sort_number",   default: 0, null: false
    t.datetime "created_at",                null: false
    t.datetime "updated_at",                null: false
    t.integer  "resource_id"
    t.string   "resource_type"
  end

  add_index "draft_staffs", ["resource_id", "resource_type"], name: "index_draft_staffs_on_resource_id_and_resource_type", using: :btree
  add_index "draft_staffs", ["sort_number"], name: "index_draft_staffs_on_sort_number", using: :btree
  add_index "draft_staffs", ["staff_id"], name: "index_draft_staffs_on_staff_id", using: :btree
  add_index "draft_staffs", ["work_id"], name: "index_draft_staffs_on_work_id", using: :btree

  create_table "draft_work_organizations", force: :cascade do |t|
    t.integer  "work_organization_id"
    t.integer  "work_id",                          null: false
    t.integer  "organization_id",                  null: false
    t.string   "role",                             null: false
    t.string   "role_other"
    t.integer  "sort_number",          default: 0, null: false
    t.datetime "created_at",                       null: false
    t.datetime "updated_at",                       null: false
  end

  add_index "draft_work_organizations", ["organization_id"], name: "index_draft_work_organizations_on_organization_id", using: :btree
  add_index "draft_work_organizations", ["sort_number"], name: "index_draft_work_organizations_on_sort_number", using: :btree
  add_index "draft_work_organizations", ["work_id"], name: "index_draft_work_organizations_on_work_id", using: :btree
  add_index "draft_work_organizations", ["work_organization_id"], name: "index_draft_work_organizations_on_work_organization_id", using: :btree

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
    t.integer  "number_format_id"
  end

  add_index "draft_works", ["number_format_id"], name: "index_draft_works_on_number_format_id", using: :btree
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
    t.integer  "work_id",                                           null: false
    t.string   "number",          limit: 510
    t.integer  "sort_number",                 default: 0,           null: false
    t.integer  "sc_count"
    t.string   "title",           limit: 510
    t.integer  "checkins_count",              default: 0,           null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "prev_episode_id"
    t.string   "aasm_state",                  default: "published", null: false
    t.boolean  "fetch_syobocal",              default: false,       null: false
    t.string   "raw_number"
  end

  add_index "episodes", ["aasm_state"], name: "index_episodes_on_aasm_state", using: :btree
  add_index "episodes", ["prev_episode_id"], name: "index_episodes_on_prev_episode_id", using: :btree
  add_index "episodes", ["work_id", "sc_count"], name: "episodes_work_id_sc_count_key", unique: true, using: :btree
  add_index "episodes", ["work_id"], name: "episodes_work_id_idx", using: :btree

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

  add_index "follows", ["following_id"], name: "follows_following_id_idx", using: :btree
  add_index "follows", ["user_id", "following_id"], name: "follows_user_id_following_id_key", unique: true, using: :btree
  add_index "follows", ["user_id"], name: "follows_user_id_idx", using: :btree

  create_table "items", force: :cascade do |t|
    t.integer  "work_id"
    t.string   "name",                     limit: 510, null: false
    t.string   "url",                      limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string   "tombo_image_file_name"
    t.string   "tombo_image_content_type"
    t.integer  "tombo_image_file_size"
    t.datetime "tombo_image_updated_at"
  end

  add_index "items", ["work_id"], name: "index_items_on_work_id", unique: true, using: :btree
  add_index "items", ["work_id"], name: "items_work_id_idx", using: :btree

  create_table "latest_statuses", force: :cascade do |t|
    t.integer  "user_id",                          null: false
    t.integer  "work_id",                          null: false
    t.integer  "next_episode_id"
    t.integer  "kind",                             null: false
    t.integer  "watched_episode_ids", default: [], null: false, array: true
    t.integer  "position",            default: 0,  null: false
    t.datetime "created_at",                       null: false
    t.datetime "updated_at",                       null: false
  end

  add_index "latest_statuses", ["next_episode_id"], name: "index_latest_statuses_on_next_episode_id", using: :btree
  add_index "latest_statuses", ["user_id", "position"], name: "index_latest_statuses_on_user_id_and_position", using: :btree
  add_index "latest_statuses", ["user_id", "work_id"], name: "index_latest_statuses_on_user_id_and_work_id", unique: true, using: :btree
  add_index "latest_statuses", ["user_id"], name: "index_latest_statuses_on_user_id", using: :btree
  add_index "latest_statuses", ["work_id"], name: "index_latest_statuses_on_work_id", using: :btree

  create_table "likes", force: :cascade do |t|
    t.integer  "user_id",                    null: false
    t.integer  "recipient_id",               null: false
    t.string   "recipient_type", limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "likes", ["user_id"], name: "likes_user_id_idx", using: :btree

  create_table "multiple_records", force: :cascade do |t|
    t.integer  "user_id",                 null: false
    t.integer  "work_id",                 null: false
    t.integer  "likes_count", default: 0, null: false
    t.datetime "created_at",              null: false
    t.datetime "updated_at",              null: false
  end

  add_index "multiple_records", ["user_id"], name: "index_multiple_records_on_user_id", using: :btree
  add_index "multiple_records", ["work_id"], name: "index_multiple_records_on_work_id", using: :btree

  create_table "notifications", force: :cascade do |t|
    t.integer  "user_id",                                    null: false
    t.integer  "action_user_id",                             null: false
    t.integer  "trackable_id",                               null: false
    t.string   "trackable_type", limit: 510,                 null: false
    t.string   "action",         limit: 510,                 null: false
    t.boolean  "read",                       default: false, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "notifications", ["action_user_id"], name: "notifications_action_user_id_idx", using: :btree
  add_index "notifications", ["user_id"], name: "notifications_user_id_idx", using: :btree

  create_table "number_formats", force: :cascade do |t|
    t.string   "name",                     null: false
    t.string   "data",        default: [], null: false, array: true
    t.integer  "sort_number", default: 0,  null: false
    t.datetime "created_at",               null: false
    t.datetime "updated_at",               null: false
  end

  add_index "number_formats", ["name"], name: "index_number_formats_on_name", unique: true, using: :btree

  create_table "organizations", force: :cascade do |t|
    t.string   "name",                                   null: false
    t.string   "url"
    t.string   "wikipedia_url"
    t.string   "twitter_username"
    t.string   "aasm_state",       default: "published", null: false
    t.datetime "created_at",                             null: false
    t.datetime "updated_at",                             null: false
  end

  add_index "organizations", ["aasm_state"], name: "index_organizations_on_aasm_state", using: :btree
  add_index "organizations", ["name"], name: "index_organizations_on_name", unique: true, using: :btree

  create_table "people", force: :cascade do |t|
    t.integer  "prefecture_id"
    t.string   "name",                                   null: false
    t.string   "name_kana"
    t.string   "nickname"
    t.string   "gender"
    t.string   "url"
    t.string   "wikipedia_url"
    t.string   "twitter_username"
    t.date     "birthday"
    t.string   "blood_type"
    t.integer  "height"
    t.string   "aasm_state",       default: "published", null: false
    t.datetime "created_at",                             null: false
    t.datetime "updated_at",                             null: false
  end

  add_index "people", ["aasm_state"], name: "index_people_on_aasm_state", using: :btree
  add_index "people", ["name"], name: "index_people_on_name", unique: true, using: :btree
  add_index "people", ["prefecture_id"], name: "index_people_on_prefecture_id", using: :btree

  create_table "prefectures", force: :cascade do |t|
    t.string   "name",       null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  add_index "prefectures", ["name"], name: "index_prefectures_on_name", unique: true, using: :btree

  create_table "profiles", force: :cascade do |t|
    t.integer  "user_id",                                                         null: false
    t.string   "name",                                limit: 510, default: "",    null: false
    t.string   "description",                         limit: 510, default: "",    null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean  "background_image_animated",                       default: false, null: false
    t.string   "tombo_avatar_file_name"
    t.string   "tombo_avatar_content_type"
    t.integer  "tombo_avatar_file_size"
    t.datetime "tombo_avatar_updated_at"
    t.string   "tombo_background_image_file_name"
    t.string   "tombo_background_image_content_type"
    t.integer  "tombo_background_image_file_size"
    t.datetime "tombo_background_image_updated_at"
    t.string   "url"
  end

  add_index "profiles", ["user_id"], name: "profiles_user_id_idx", using: :btree
  add_index "profiles", ["user_id"], name: "profiles_user_id_key", unique: true, using: :btree

  create_table "programs", force: :cascade do |t|
    t.integer  "channel_id",                     null: false
    t.integer  "episode_id",                     null: false
    t.integer  "work_id",                        null: false
    t.datetime "started_at",                     null: false
    t.datetime "sc_last_update"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "sc_pid"
    t.boolean  "rebroadcast",    default: false, null: false
  end

  add_index "programs", ["channel_id"], name: "programs_channel_id_idx", using: :btree
  add_index "programs", ["episode_id"], name: "programs_episode_id_idx", using: :btree
  add_index "programs", ["sc_pid"], name: "index_programs_on_sc_pid", unique: true, using: :btree
  add_index "programs", ["work_id"], name: "programs_work_id_idx", using: :btree

  create_table "providers", force: :cascade do |t|
    t.integer  "user_id",                      null: false
    t.string   "name",             limit: 510, null: false
    t.string   "uid",              limit: 510, null: false
    t.string   "token",            limit: 510, null: false
    t.integer  "token_expires_at"
    t.string   "token_secret",     limit: 510
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "providers", ["name", "uid"], name: "providers_name_uid_key", unique: true, using: :btree
  add_index "providers", ["user_id"], name: "providers_user_id_idx", using: :btree

  create_table "receptions", force: :cascade do |t|
    t.integer  "user_id",    null: false
    t.integer  "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "receptions", ["channel_id"], name: "receptions_channel_id_idx", using: :btree
  add_index "receptions", ["user_id", "channel_id"], name: "receptions_user_id_channel_id_key", unique: true, using: :btree
  add_index "receptions", ["user_id"], name: "receptions_user_id_idx", using: :btree

  create_table "seasons", force: :cascade do |t|
    t.string   "name",        limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "sort_number",             null: false
    t.integer  "year",                    null: false
  end

  add_index "seasons", ["sort_number"], name: "index_seasons_on_sort_number", unique: true, using: :btree
  add_index "seasons", ["year", "name"], name: "index_seasons_on_year_and_name", unique: true, using: :btree
  add_index "seasons", ["year"], name: "index_seasons_on_year", using: :btree

  create_table "sessions", force: :cascade do |t|
    t.string   "session_id", limit: 510, null: false
    t.text     "data"
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "sessions", ["session_id"], name: "sessions_session_id_key", unique: true, using: :btree

  create_table "settings", force: :cascade do |t|
    t.integer  "user_id",                                  null: false
    t.boolean  "hide_checkin_comment",     default: true,  null: false
    t.datetime "created_at",                               null: false
    t.datetime "updated_at",                               null: false
    t.boolean  "share_record_to_twitter",  default: false
    t.boolean  "share_record_to_facebook", default: false
  end

  add_index "settings", ["user_id"], name: "index_settings_on_user_id", using: :btree

  create_table "staffs", force: :cascade do |t|
    t.integer  "work_id",                             null: false
    t.string   "name",                                null: false
    t.string   "role",                                null: false
    t.string   "role_other"
    t.string   "aasm_state",    default: "published", null: false
    t.integer  "sort_number",   default: 0,           null: false
    t.datetime "created_at",                          null: false
    t.datetime "updated_at",                          null: false
    t.integer  "resource_id",                         null: false
    t.string   "resource_type",                       null: false
  end

  add_index "staffs", ["aasm_state"], name: "index_staffs_on_aasm_state", using: :btree
  add_index "staffs", ["resource_id", "resource_type"], name: "index_staffs_on_resource_id_and_resource_type", using: :btree
  add_index "staffs", ["sort_number"], name: "index_staffs_on_sort_number", using: :btree
  add_index "staffs", ["work_id"], name: "index_staffs_on_work_id", using: :btree

  create_table "statuses", force: :cascade do |t|
    t.integer  "user_id",                 null: false
    t.integer  "work_id",                 null: false
    t.integer  "kind",                    null: false
    t.integer  "likes_count", default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "statuses", ["user_id"], name: "statuses_user_id_idx", using: :btree
  add_index "statuses", ["work_id"], name: "statuses_work_id_idx", using: :btree

  create_table "syobocal_alerts", force: :cascade do |t|
    t.integer  "work_id"
    t.integer  "kind",                        null: false
    t.integer  "sc_prog_item_id"
    t.string   "sc_sub_title",    limit: 255
    t.string   "sc_prog_comment", limit: 255
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "syobocal_alerts", ["kind"], name: "index_syobocal_alerts_on_kind", using: :btree
  add_index "syobocal_alerts", ["sc_prog_item_id"], name: "index_syobocal_alerts_on_sc_prog_item_id", using: :btree

  create_table "tips", force: :cascade do |t|
    t.integer  "target",                   null: false
    t.string   "partial_name", limit: 255, null: false
    t.string   "title",        limit: 255, null: false
    t.string   "icon_name",    limit: 255, null: false
    t.datetime "created_at",               null: false
    t.datetime "updated_at",               null: false
  end

  add_index "tips", ["partial_name"], name: "index_tips_on_partial_name", unique: true, using: :btree

  create_table "twitter_bots", force: :cascade do |t|
    t.string   "name",       limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "twitter_bots", ["name"], name: "twitter_bots_name_key", unique: true, using: :btree

  create_table "users", force: :cascade do |t|
    t.string   "username",             limit: 510,              null: false
    t.string   "email",                limit: 510,              null: false
    t.integer  "role",                                          null: false
    t.string   "encrypted_password",   limit: 510, default: "", null: false
    t.datetime "remember_created_at"
    t.integer  "sign_in_count",                    default: 0,  null: false
    t.datetime "current_sign_in_at"
    t.datetime "last_sign_in_at"
    t.string   "current_sign_in_ip",   limit: 510
    t.string   "last_sign_in_ip",      limit: 510
    t.string   "confirmation_token",   limit: 510
    t.datetime "confirmed_at"
    t.datetime "confirmation_sent_at"
    t.string   "unconfirmed_email",    limit: 510
    t.integer  "checkins_count",                   default: 0,  null: false
    t.integer  "notifications_count",              default: 0,  null: false
    t.datetime "created_at"
    t.datetime "updated_at"
  end

  add_index "users", ["confirmation_token"], name: "users_confirmation_token_key", unique: true, using: :btree
  add_index "users", ["email"], name: "users_email_key", unique: true, using: :btree
  add_index "users", ["username"], name: "users_username_key", unique: true, using: :btree

  create_table "versions", force: :cascade do |t|
    t.string   "item_type",  limit: 510, null: false
    t.integer  "item_id",                null: false
    t.string   "event",      limit: 510, null: false
    t.string   "whodunnit",  limit: 510
    t.text     "object"
    t.datetime "created_at"
  end

  create_table "work_organizations", force: :cascade do |t|
    t.integer  "work_id",                               null: false
    t.integer  "organization_id",                       null: false
    t.string   "role",                                  null: false
    t.string   "role_other"
    t.string   "aasm_state",      default: "published", null: false
    t.integer  "sort_number",     default: 0,           null: false
    t.datetime "created_at",                            null: false
    t.datetime "updated_at",                            null: false
  end

  add_index "work_organizations", ["aasm_state"], name: "index_work_organizations_on_aasm_state", using: :btree
  add_index "work_organizations", ["organization_id"], name: "index_work_organizations_on_organization_id", using: :btree
  add_index "work_organizations", ["sort_number"], name: "index_work_organizations_on_sort_number", using: :btree
  add_index "work_organizations", ["work_id", "organization_id"], name: "index_work_organizations_on_work_id_and_organization_id", unique: true, using: :btree
  add_index "work_organizations", ["work_id"], name: "index_work_organizations_on_work_id", using: :btree

  create_table "works", force: :cascade do |t|
    t.integer  "season_id"
    t.integer  "sc_tid"
    t.string   "title",             limit: 510,                       null: false
    t.integer  "media",                                               null: false
    t.string   "official_site_url", limit: 510, default: "",          null: false
    t.string   "wikipedia_url",     limit: 510, default: "",          null: false
    t.integer  "episodes_count",                default: 0,           null: false
    t.integer  "watchers_count",                default: 0,           null: false
    t.date     "released_at"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string   "twitter_username",  limit: 510
    t.string   "twitter_hashtag",   limit: 510
    t.string   "released_at_about"
    t.string   "aasm_state",                    default: "published", null: false
    t.integer  "number_format_id"
  end

  add_index "works", ["aasm_state"], name: "index_works_on_aasm_state", using: :btree
  add_index "works", ["number_format_id"], name: "index_works_on_number_format_id", using: :btree
  add_index "works", ["sc_tid"], name: "works_sc_tid_key", unique: true, using: :btree
  add_index "works", ["season_id"], name: "works_season_id_idx", using: :btree

  add_foreign_key "activities", "users", name: "activities_user_id_fk", on_delete: :cascade
  add_foreign_key "casts", "people"
  add_foreign_key "casts", "works"
  add_foreign_key "channel_works", "channels", name: "channel_works_channel_id_fk", on_delete: :cascade
  add_foreign_key "channel_works", "users", name: "channel_works_user_id_fk", on_delete: :cascade
  add_foreign_key "channel_works", "works", name: "channel_works_work_id_fk", on_delete: :cascade
  add_foreign_key "channels", "channel_groups", name: "channels_channel_group_id_fk", on_delete: :cascade
  add_foreign_key "checkins", "episodes", name: "checkins_episode_id_fk", on_delete: :cascade
  add_foreign_key "checkins", "multiple_records"
  add_foreign_key "checkins", "users", name: "checkins_user_id_fk", on_delete: :cascade
  add_foreign_key "checkins", "works", name: "checkins_work_id_fk"
  add_foreign_key "comments", "checkins", name: "comments_checkin_id_fk", on_delete: :cascade
  add_foreign_key "comments", "users", name: "comments_user_id_fk", on_delete: :cascade
  add_foreign_key "comments", "works"
  add_foreign_key "cover_images", "works", name: "cover_images_work_id_fk", on_delete: :cascade
  add_foreign_key "db_activities", "users"
  add_foreign_key "draft_casts", "casts"
  add_foreign_key "draft_casts", "people"
  add_foreign_key "draft_casts", "works"
  add_foreign_key "draft_episodes", "episodes"
  add_foreign_key "draft_episodes", "episodes", column: "prev_episode_id"
  add_foreign_key "draft_episodes", "works"
  add_foreign_key "draft_items", "items"
  add_foreign_key "draft_items", "works"
  add_foreign_key "draft_multiple_episodes", "works"
  add_foreign_key "draft_organizations", "organizations"
  add_foreign_key "draft_people", "people"
  add_foreign_key "draft_people", "prefectures"
  add_foreign_key "draft_programs", "channels"
  add_foreign_key "draft_programs", "episodes"
  add_foreign_key "draft_programs", "programs"
  add_foreign_key "draft_programs", "works"
  add_foreign_key "draft_staffs", "staffs"
  add_foreign_key "draft_staffs", "works"
  add_foreign_key "draft_work_organizations", "organizations"
  add_foreign_key "draft_work_organizations", "work_organizations"
  add_foreign_key "draft_work_organizations", "works"
  add_foreign_key "draft_works", "number_formats"
  add_foreign_key "draft_works", "seasons"
  add_foreign_key "draft_works", "works"
  add_foreign_key "edit_request_comments", "edit_requests", on_delete: :cascade
  add_foreign_key "edit_request_comments", "users", on_delete: :cascade
  add_foreign_key "edit_request_participants", "edit_requests"
  add_foreign_key "edit_request_participants", "users"
  add_foreign_key "edit_requests", "users", on_delete: :cascade
  add_foreign_key "episodes", "episodes", column: "prev_episode_id"
  add_foreign_key "episodes", "works", name: "episodes_work_id_fk", on_delete: :cascade
  add_foreign_key "finished_tips", "tips", name: "finished_tips_tip_id_fk", on_delete: :cascade
  add_foreign_key "finished_tips", "users", name: "finished_tips_user_id_fk", on_delete: :cascade
  add_foreign_key "follows", "users", column: "following_id", name: "follows_following_id_fk", on_delete: :cascade
  add_foreign_key "follows", "users", name: "follows_user_id_fk", on_delete: :cascade
  add_foreign_key "items", "works", name: "items_work_id_fk", on_delete: :cascade
  add_foreign_key "latest_statuses", "episodes", column: "next_episode_id"
  add_foreign_key "latest_statuses", "users"
  add_foreign_key "latest_statuses", "works"
  add_foreign_key "likes", "users", name: "likes_user_id_fk", on_delete: :cascade
  add_foreign_key "multiple_records", "users"
  add_foreign_key "multiple_records", "works"
  add_foreign_key "notifications", "users", column: "action_user_id", name: "notifications_action_user_id_fk", on_delete: :cascade
  add_foreign_key "notifications", "users", name: "notifications_user_id_fk", on_delete: :cascade
  add_foreign_key "people", "prefectures"
  add_foreign_key "profiles", "users", name: "profiles_user_id_fk", on_delete: :cascade
  add_foreign_key "programs", "channels", name: "programs_channel_id_fk", on_delete: :cascade
  add_foreign_key "programs", "episodes", name: "programs_episode_id_fk", on_delete: :cascade
  add_foreign_key "programs", "works", name: "programs_work_id_fk", on_delete: :cascade
  add_foreign_key "providers", "users", name: "providers_user_id_fk", on_delete: :cascade
  add_foreign_key "receptions", "channels", name: "receptions_channel_id_fk", on_delete: :cascade
  add_foreign_key "receptions", "users", name: "receptions_user_id_fk", on_delete: :cascade
  add_foreign_key "settings", "users"
  add_foreign_key "staffs", "works"
  add_foreign_key "statuses", "users", name: "statuses_user_id_fk", on_delete: :cascade
  add_foreign_key "statuses", "works", name: "statuses_work_id_fk", on_delete: :cascade
  add_foreign_key "syobocal_alerts", "works", name: "syobocal_alerts_work_id_fk", on_delete: :cascade
  add_foreign_key "work_organizations", "organizations"
  add_foreign_key "work_organizations", "works"
  add_foreign_key "works", "number_formats"
  add_foreign_key "works", "seasons", name: "works_season_id_fk", on_delete: :cascade
end
