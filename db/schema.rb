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

ActiveRecord::Schema.define(version: 20170504111113) do

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
    t.index ["user_id"], name: "activities_user_id_idx", using: :btree
  end

  create_table "casts", force: :cascade do |t|
    t.integer  "person_id",                          null: false
    t.integer  "work_id",                            null: false
    t.string   "name",                               null: false
    t.string   "part",                               null: false
    t.string   "aasm_state",   default: "published", null: false
    t.integer  "sort_number",  default: 0,           null: false
    t.datetime "created_at",                         null: false
    t.datetime "updated_at",                         null: false
    t.integer  "character_id",                       null: false
    t.string   "name_en",      default: "",          null: false
    t.index ["aasm_state"], name: "index_casts_on_aasm_state", using: :btree
    t.index ["character_id"], name: "index_casts_on_character_id", using: :btree
    t.index ["person_id"], name: "index_casts_on_person_id", using: :btree
    t.index ["sort_number"], name: "index_casts_on_sort_number", using: :btree
    t.index ["work_id"], name: "index_casts_on_work_id", using: :btree
  end

  create_table "channel_groups", force: :cascade do |t|
    t.string   "sc_chgid",    limit: 510, null: false
    t.string   "name",        limit: 510, null: false
    t.integer  "sort_number"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["sc_chgid"], name: "channel_groups_sc_chgid_key", unique: true, using: :btree
  end

  create_table "channel_works", force: :cascade do |t|
    t.integer  "user_id",    null: false
    t.integer  "work_id",    null: false
    t.integer  "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["channel_id"], name: "channel_works_channel_id_idx", using: :btree
    t.index ["user_id", "work_id", "channel_id"], name: "channel_works_user_id_work_id_channel_id_key", unique: true, using: :btree
    t.index ["user_id"], name: "channel_works_user_id_idx", using: :btree
    t.index ["work_id"], name: "channel_works_work_id_idx", using: :btree
  end

  create_table "channels", force: :cascade do |t|
    t.integer  "channel_group_id",                null: false
    t.integer  "sc_chid",                         null: false
    t.string   "name",                            null: false, collation: "C"
    t.boolean  "published",        default: true, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["channel_group_id"], name: "channels_channel_group_id_idx", using: :btree
    t.index ["sc_chid"], name: "channels_sc_chid_key", unique: true, using: :btree
  end

  create_table "character_images", force: :cascade do |t|
    t.integer  "character_id",                         null: false
    t.integer  "user_id",                              null: false
    t.string   "attachment_file_name",                 null: false
    t.integer  "attachment_file_size",                 null: false
    t.string   "attachment_content_type",              null: false
    t.datetime "attachment_updated_at",                null: false
    t.string   "copyright",               default: "", null: false
    t.string   "asin",                    default: "", null: false
    t.datetime "created_at",                           null: false
    t.datetime "updated_at",                           null: false
    t.index ["character_id"], name: "index_character_images_on_character_id", using: :btree
    t.index ["user_id"], name: "index_character_images_on_user_id", using: :btree
  end

  create_table "characters", force: :cascade do |t|
    t.string   "name",                                            null: false
    t.string   "name_kana",                 default: "",          null: false
    t.string   "name_en",                   default: "",          null: false
    t.string   "kind",                      default: "",          null: false
    t.string   "nickname",                  default: "",          null: false
    t.string   "nickname_en",               default: "",          null: false
    t.string   "birthday",                  default: "",          null: false
    t.string   "birthday_en",               default: "",          null: false
    t.string   "age",                       default: "",          null: false
    t.string   "age_en",                    default: "",          null: false
    t.string   "blood_type",                default: "",          null: false
    t.string   "blood_type_en",             default: "",          null: false
    t.string   "height",                    default: "",          null: false
    t.string   "height_en",                 default: "",          null: false
    t.string   "weight",                    default: "",          null: false
    t.string   "weight_en",                 default: "",          null: false
    t.string   "nationality",               default: "",          null: false
    t.string   "nationality_en",            default: "",          null: false
    t.string   "occupation",                default: "",          null: false
    t.string   "occupation_en",             default: "",          null: false
    t.text     "description",               default: "",          null: false
    t.text     "description_en",            default: "",          null: false
    t.string   "aasm_state",                default: "published", null: false
    t.datetime "created_at",                                      null: false
    t.datetime "updated_at",                                      null: false
    t.string   "description_source",        default: "",          null: false
    t.string   "description_source_en",     default: "",          null: false
    t.integer  "favorite_characters_count", default: 0,           null: false
    t.integer  "series_id"
    t.index ["favorite_characters_count"], name: "index_characters_on_favorite_characters_count", using: :btree
    t.index ["name", "series_id"], name: "index_characters_on_name_and_series_id", unique: true, using: :btree
    t.index ["series_id"], name: "index_characters_on_series_id", using: :btree
  end

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
    t.integer  "work_id",                                          null: false
    t.float    "rating"
    t.integer  "multiple_record_id"
    t.integer  "oauth_application_id"
    t.index ["episode_id"], name: "checkins_episode_id_idx", using: :btree
    t.index ["facebook_url_hash"], name: "checkins_facebook_url_hash_key", unique: true, using: :btree
    t.index ["multiple_record_id"], name: "index_checkins_on_multiple_record_id", using: :btree
    t.index ["oauth_application_id"], name: "index_checkins_on_oauth_application_id", using: :btree
    t.index ["twitter_url_hash"], name: "checkins_twitter_url_hash_key", unique: true, using: :btree
    t.index ["user_id"], name: "checkins_user_id_idx", using: :btree
    t.index ["work_id"], name: "index_checkins_on_work_id", using: :btree
  end

  create_table "collection_item_agreements", force: :cascade do |t|
    t.integer  "user_id",            null: false
    t.integer  "collection_item_id", null: false
    t.datetime "created_at",         null: false
    t.datetime "updated_at",         null: false
    t.index ["collection_item_id"], name: "index_collection_item_agreements_on_collection_item_id", using: :btree
    t.index ["user_id", "collection_item_id"], name: "index_cia_on_uid_and_ciid", unique: true, using: :btree
    t.index ["user_id"], name: "index_collection_item_agreements_on_user_id", using: :btree
  end

  create_table "collection_items", force: :cascade do |t|
    t.integer  "collection_id", null: false
    t.integer  "work_id",       null: false
    t.string   "title",         null: false
    t.text     "comment"
    t.datetime "created_at",    null: false
    t.datetime "updated_at",    null: false
    t.index ["collection_id", "work_id"], name: "index_collection_items_on_collection_id_and_work_id", unique: true, using: :btree
    t.index ["collection_id"], name: "index_collection_items_on_collection_id", using: :btree
    t.index ["work_id"], name: "index_collection_items_on_work_id", using: :btree
  end

  create_table "collections", force: :cascade do |t|
    t.integer  "user_id",                        null: false
    t.string   "title",                          null: false
    t.string   "description"
    t.string   "aasm_state",   default: "draft", null: false
    t.integer  "likes_count",  default: 0,       null: false
    t.datetime "published_at"
    t.datetime "created_at",                     null: false
    t.datetime "updated_at",                     null: false
    t.index ["user_id"], name: "index_collections_on_user_id", using: :btree
  end

  create_table "comments", force: :cascade do |t|
    t.integer  "user_id",                 null: false
    t.integer  "checkin_id",              null: false
    t.text     "body",                    null: false
    t.integer  "likes_count", default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "work_id"
    t.index ["checkin_id"], name: "comments_checkin_id_idx", using: :btree
    t.index ["user_id"], name: "comments_user_id_idx", using: :btree
    t.index ["work_id"], name: "index_comments_on_work_id", using: :btree
  end

  create_table "cover_images", force: :cascade do |t|
    t.integer  "work_id",                null: false
    t.string   "file_name",  limit: 510, null: false
    t.string   "location",   limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["work_id"], name: "cover_images_work_id_idx", using: :btree
  end

  create_table "db_activities", force: :cascade do |t|
    t.integer  "user_id",            null: false
    t.integer  "trackable_id",       null: false
    t.string   "trackable_type",     null: false
    t.string   "action",             null: false
    t.json     "parameters"
    t.datetime "created_at",         null: false
    t.datetime "updated_at",         null: false
    t.integer  "root_resource_id"
    t.string   "root_resource_type"
    t.integer  "object_id"
    t.string   "object_type"
    t.index ["object_id", "object_type"], name: "index_db_activities_on_object_id_and_object_type", using: :btree
    t.index ["root_resource_id", "root_resource_type"], name: "index_db_activities_on_root_resource_id_and_root_resource_type", using: :btree
    t.index ["trackable_id", "trackable_type"], name: "index_db_activities_on_trackable_id_and_trackable_type", using: :btree
  end

  create_table "db_comments", force: :cascade do |t|
    t.integer  "user_id",       null: false
    t.integer  "resource_id",   null: false
    t.string   "resource_type", null: false
    t.text     "body",          null: false
    t.datetime "created_at",    null: false
    t.datetime "updated_at",    null: false
    t.index ["resource_id", "resource_type"], name: "index_db_comments_on_resource_id_and_resource_type", using: :btree
    t.index ["user_id"], name: "index_db_comments_on_user_id", using: :btree
  end

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
    t.index ["priority", "run_at"], name: "delayed_jobs_priority", using: :btree
  end

  create_table "draft_casts", force: :cascade do |t|
    t.integer  "cast_id"
    t.integer  "person_id",               null: false
    t.integer  "work_id",                 null: false
    t.string   "name",                    null: false
    t.string   "part",                    null: false
    t.integer  "sort_number", default: 0, null: false
    t.datetime "created_at",              null: false
    t.datetime "updated_at",              null: false
    t.index ["cast_id"], name: "index_draft_casts_on_cast_id", using: :btree
    t.index ["person_id"], name: "index_draft_casts_on_person_id", using: :btree
    t.index ["sort_number"], name: "index_draft_casts_on_sort_number", using: :btree
    t.index ["work_id"], name: "index_draft_casts_on_work_id", using: :btree
  end

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
    t.index ["episode_id"], name: "index_draft_episodes_on_episode_id", using: :btree
    t.index ["prev_episode_id"], name: "index_draft_episodes_on_prev_episode_id", using: :btree
    t.index ["work_id"], name: "index_draft_episodes_on_work_id", using: :btree
  end

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
    t.index ["item_id"], name: "index_draft_items_on_item_id", using: :btree
    t.index ["work_id"], name: "index_draft_items_on_work_id", using: :btree
  end

  create_table "draft_multiple_episodes", force: :cascade do |t|
    t.integer  "work_id",    null: false
    t.text     "body",       null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["work_id"], name: "index_draft_multiple_episodes_on_work_id", using: :btree
  end

  create_table "draft_organizations", force: :cascade do |t|
    t.integer  "organization_id"
    t.string   "name",                          null: false
    t.string   "url"
    t.string   "wikipedia_url"
    t.string   "twitter_username"
    t.datetime "created_at",                    null: false
    t.datetime "updated_at",                    null: false
    t.string   "name_kana",        default: "", null: false
    t.index ["name"], name: "index_draft_organizations_on_name", using: :btree
    t.index ["organization_id"], name: "index_draft_organizations_on_organization_id", using: :btree
  end

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
    t.index ["name"], name: "index_draft_people_on_name", using: :btree
    t.index ["person_id"], name: "index_draft_people_on_person_id", using: :btree
    t.index ["prefecture_id"], name: "index_draft_people_on_prefecture_id", using: :btree
  end

  create_table "draft_programs", force: :cascade do |t|
    t.integer  "program_id"
    t.integer  "channel_id",                  null: false
    t.integer  "episode_id",                  null: false
    t.integer  "work_id",                     null: false
    t.datetime "started_at",                  null: false
    t.datetime "created_at",                  null: false
    t.datetime "updated_at",                  null: false
    t.boolean  "rebroadcast", default: false, null: false
    t.index ["channel_id"], name: "index_draft_programs_on_channel_id", using: :btree
    t.index ["episode_id"], name: "index_draft_programs_on_episode_id", using: :btree
    t.index ["program_id"], name: "index_draft_programs_on_program_id", using: :btree
    t.index ["work_id"], name: "index_draft_programs_on_work_id", using: :btree
  end

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
    t.index ["resource_id", "resource_type"], name: "index_draft_staffs_on_resource_id_and_resource_type", using: :btree
    t.index ["sort_number"], name: "index_draft_staffs_on_sort_number", using: :btree
    t.index ["staff_id"], name: "index_draft_staffs_on_staff_id", using: :btree
    t.index ["work_id"], name: "index_draft_staffs_on_work_id", using: :btree
  end

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
    t.string   "title_kana",        default: "", null: false
    t.index ["number_format_id"], name: "index_draft_works_on_number_format_id", using: :btree
    t.index ["sc_tid"], name: "index_draft_works_on_sc_tid", using: :btree
    t.index ["season_id"], name: "index_draft_works_on_season_id", using: :btree
    t.index ["work_id"], name: "index_draft_works_on_work_id", using: :btree
  end

  create_table "edit_request_comments", force: :cascade do |t|
    t.integer  "edit_request_id", null: false
    t.integer  "user_id",         null: false
    t.text     "body",            null: false
    t.datetime "created_at",      null: false
    t.datetime "updated_at",      null: false
    t.index ["edit_request_id"], name: "index_edit_request_comments_on_edit_request_id", using: :btree
    t.index ["user_id"], name: "index_edit_request_comments_on_user_id", using: :btree
  end

  create_table "edit_request_participants", force: :cascade do |t|
    t.integer  "edit_request_id", null: false
    t.integer  "user_id",         null: false
    t.datetime "created_at",      null: false
    t.datetime "updated_at",      null: false
    t.index ["edit_request_id", "user_id"], name: "index_edit_request_participants_on_edit_request_id_and_user_id", unique: true, using: :btree
    t.index ["edit_request_id"], name: "index_edit_request_participants_on_edit_request_id", using: :btree
    t.index ["user_id"], name: "index_edit_request_participants_on_user_id", using: :btree
  end

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
    t.index ["draft_resource_id", "draft_resource_type"], name: "index_er_on_drid_and_drtype", using: :btree
    t.index ["user_id"], name: "index_edit_requests_on_user_id", using: :btree
  end

  create_table "email_notifications", force: :cascade do |t|
    t.integer  "user_id",                                   null: false
    t.string   "unsubscription_key",                        null: false
    t.boolean  "event_followed_user",        default: true, null: false
    t.boolean  "event_liked_record",         default: true, null: false
    t.boolean  "event_friends_joined",       default: true, null: false
    t.boolean  "event_next_season_came",     default: true, null: false
    t.datetime "created_at",                                null: false
    t.datetime "updated_at",                                null: false
    t.boolean  "event_favorite_works_added", default: true, null: false
    t.index ["unsubscription_key"], name: "index_email_notifications_on_unsubscription_key", unique: true, using: :btree
    t.index ["user_id"], name: "index_email_notifications_on_user_id", unique: true, using: :btree
  end

  create_table "episodes", force: :cascade do |t|
    t.integer  "work_id",                                                 null: false
    t.string   "number",                limit: 510
    t.integer  "sort_number",                       default: 0,           null: false
    t.integer  "sc_count"
    t.string   "title",                 limit: 510
    t.integer  "checkins_count",                    default: 0,           null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "prev_episode_id"
    t.string   "aasm_state",                        default: "published", null: false
    t.boolean  "fetch_syobocal",                    default: false,       null: false
    t.string   "raw_number"
    t.float    "avg_rating"
    t.string   "title_ro",                          default: "",          null: false
    t.string   "title_en",                          default: "",          null: false
    t.integer  "record_comments_count",             default: 0,           null: false
    t.index ["aasm_state"], name: "index_episodes_on_aasm_state", using: :btree
    t.index ["prev_episode_id"], name: "index_episodes_on_prev_episode_id", using: :btree
    t.index ["work_id", "sc_count"], name: "episodes_work_id_sc_count_key", unique: true, using: :btree
    t.index ["work_id"], name: "episodes_work_id_idx", using: :btree
  end

  create_table "favorite_characters", force: :cascade do |t|
    t.integer  "user_id",      null: false
    t.integer  "character_id", null: false
    t.datetime "created_at",   null: false
    t.datetime "updated_at",   null: false
    t.index ["character_id"], name: "index_favorite_characters_on_character_id", using: :btree
    t.index ["user_id", "character_id"], name: "index_favorite_characters_on_user_id_and_character_id", unique: true, using: :btree
    t.index ["user_id"], name: "index_favorite_characters_on_user_id", using: :btree
  end

  create_table "favorite_organizations", force: :cascade do |t|
    t.integer  "user_id",                         null: false
    t.integer  "organization_id",                 null: false
    t.datetime "created_at",                      null: false
    t.datetime "updated_at",                      null: false
    t.integer  "watched_works_count", default: 0, null: false
    t.index ["organization_id"], name: "index_favorite_organizations_on_organization_id", using: :btree
    t.index ["user_id", "organization_id"], name: "index_favorite_organizations_on_user_id_and_organization_id", unique: true, using: :btree
    t.index ["user_id"], name: "index_favorite_organizations_on_user_id", using: :btree
    t.index ["watched_works_count"], name: "index_favorite_organizations_on_watched_works_count", using: :btree
  end

  create_table "favorite_people", force: :cascade do |t|
    t.integer  "user_id",                         null: false
    t.integer  "person_id",                       null: false
    t.datetime "created_at",                      null: false
    t.datetime "updated_at",                      null: false
    t.integer  "watched_works_count", default: 0, null: false
    t.index ["person_id"], name: "index_favorite_people_on_person_id", using: :btree
    t.index ["user_id", "person_id"], name: "index_favorite_people_on_user_id_and_person_id", unique: true, using: :btree
    t.index ["user_id"], name: "index_favorite_people_on_user_id", using: :btree
    t.index ["watched_works_count"], name: "index_favorite_people_on_watched_works_count", using: :btree
  end

  create_table "finished_tips", force: :cascade do |t|
    t.integer  "user_id",    null: false
    t.integer  "tip_id",     null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "tip_id"], name: "index_finished_tips_on_user_id_and_tip_id", unique: true, using: :btree
  end

  create_table "follows", force: :cascade do |t|
    t.integer  "user_id",      null: false
    t.integer  "following_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["following_id"], name: "follows_following_id_idx", using: :btree
    t.index ["user_id", "following_id"], name: "follows_user_id_following_id_key", unique: true, using: :btree
    t.index ["user_id"], name: "follows_user_id_idx", using: :btree
  end

  create_table "forum_categories", force: :cascade do |t|
    t.string   "slug",                          null: false
    t.string   "name",                          null: false
    t.string   "name_en",                       null: false
    t.string   "description",                   null: false
    t.string   "description_en",                null: false
    t.string   "postable_role",                 null: false
    t.integer  "sort_number",                   null: false
    t.integer  "forum_posts_count", default: 0, null: false
    t.datetime "created_at",                    null: false
    t.datetime "updated_at",                    null: false
    t.index ["slug"], name: "index_forum_categories_on_slug", unique: true, using: :btree
  end

  create_table "forum_comments", force: :cascade do |t|
    t.integer  "user_id",       null: false
    t.integer  "forum_post_id", null: false
    t.text     "body",          null: false
    t.datetime "edited_at",                  comment: "The datetime which user has changed body."
    t.datetime "created_at",    null: false
    t.datetime "updated_at",    null: false
    t.index ["forum_post_id"], name: "index_forum_comments_on_forum_post_id", using: :btree
    t.index ["user_id"], name: "index_forum_comments_on_user_id", using: :btree
  end

  create_table "forum_post_participants", force: :cascade do |t|
    t.integer  "forum_post_id", null: false
    t.integer  "user_id",       null: false
    t.datetime "created_at",    null: false
    t.datetime "updated_at",    null: false
    t.index ["forum_post_id", "user_id"], name: "index_forum_post_participants_on_forum_post_id_and_user_id", unique: true, using: :btree
    t.index ["forum_post_id"], name: "index_forum_post_participants_on_forum_post_id", using: :btree
    t.index ["user_id"], name: "index_forum_post_participants_on_user_id", using: :btree
  end

  create_table "forum_posts", force: :cascade do |t|
    t.integer  "user_id",                           null: false
    t.integer  "forum_category_id",                 null: false
    t.string   "title",                             null: false
    t.text     "body",                 default: "", null: false
    t.integer  "forum_comments_count", default: 0,  null: false
    t.datetime "edited_at",                                      comment: "The datetime which user has changed title, body and so on."
    t.datetime "last_commented_at",                 null: false
    t.datetime "created_at",                        null: false
    t.datetime "updated_at",                        null: false
    t.index ["forum_category_id"], name: "index_forum_posts_on_forum_category_id", using: :btree
    t.index ["user_id"], name: "index_forum_posts_on_user_id", using: :btree
  end

  create_table "impressions", force: :cascade do |t|
    t.string   "impressionable_type"
    t.integer  "impressionable_id"
    t.integer  "user_id"
    t.string   "controller_name"
    t.string   "action_name"
    t.string   "view_name"
    t.string   "request_hash"
    t.string   "ip_address"
    t.string   "session_hash"
    t.text     "message"
    t.text     "referrer"
    t.text     "params"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["controller_name", "action_name", "ip_address"], name: "controlleraction_ip_index", using: :btree
    t.index ["controller_name", "action_name", "request_hash"], name: "controlleraction_request_index", using: :btree
    t.index ["controller_name", "action_name", "session_hash"], name: "controlleraction_session_index", using: :btree
    t.index ["impressionable_type", "impressionable_id", "ip_address"], name: "poly_ip_index", using: :btree
    t.index ["impressionable_type", "impressionable_id", "params"], name: "poly_params_request_index", using: :btree
    t.index ["impressionable_type", "impressionable_id", "request_hash"], name: "poly_request_index", using: :btree
    t.index ["impressionable_type", "impressionable_id", "session_hash"], name: "poly_session_index", using: :btree
    t.index ["impressionable_type", "message", "impressionable_id"], name: "impressionable_type_message_index", using: :btree
    t.index ["user_id"], name: "index_impressions_on_user_id", using: :btree
  end

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
    t.index ["work_id"], name: "index_items_on_work_id", unique: true, using: :btree
    t.index ["work_id"], name: "items_work_id_idx", using: :btree
  end

  create_table "latest_statuses", force: :cascade do |t|
    t.integer  "user_id",                          null: false
    t.integer  "work_id",                          null: false
    t.integer  "next_episode_id"
    t.integer  "kind",                             null: false
    t.integer  "watched_episode_ids", default: [], null: false, array: true
    t.integer  "position",            default: 0,  null: false
    t.datetime "created_at",                       null: false
    t.datetime "updated_at",                       null: false
    t.index ["next_episode_id"], name: "index_latest_statuses_on_next_episode_id", using: :btree
    t.index ["user_id", "position"], name: "index_latest_statuses_on_user_id_and_position", using: :btree
    t.index ["user_id", "work_id"], name: "index_latest_statuses_on_user_id_and_work_id", unique: true, using: :btree
    t.index ["user_id"], name: "index_latest_statuses_on_user_id", using: :btree
    t.index ["work_id"], name: "index_latest_statuses_on_work_id", using: :btree
  end

  create_table "likes", force: :cascade do |t|
    t.integer  "user_id",                    null: false
    t.integer  "recipient_id",               null: false
    t.string   "recipient_type", limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["user_id"], name: "likes_user_id_idx", using: :btree
  end

  create_table "multiple_records", force: :cascade do |t|
    t.integer  "user_id",                 null: false
    t.integer  "work_id",                 null: false
    t.integer  "likes_count", default: 0, null: false
    t.datetime "created_at",              null: false
    t.datetime "updated_at",              null: false
    t.index ["user_id"], name: "index_multiple_records_on_user_id", using: :btree
    t.index ["work_id"], name: "index_multiple_records_on_work_id", using: :btree
  end

  create_table "mute_users", force: :cascade do |t|
    t.integer  "user_id",       null: false
    t.integer  "muted_user_id", null: false
    t.datetime "created_at",    null: false
    t.datetime "updated_at",    null: false
    t.index ["muted_user_id"], name: "index_mute_users_on_muted_user_id", using: :btree
    t.index ["user_id", "muted_user_id"], name: "index_mute_users_on_user_id_and_muted_user_id", unique: true, using: :btree
    t.index ["user_id"], name: "index_mute_users_on_user_id", using: :btree
  end

  create_table "notifications", force: :cascade do |t|
    t.integer  "user_id",                                    null: false
    t.integer  "action_user_id",                             null: false
    t.integer  "trackable_id",                               null: false
    t.string   "trackable_type", limit: 510,                 null: false
    t.string   "action",         limit: 510,                 null: false
    t.boolean  "read",                       default: false, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["action_user_id"], name: "notifications_action_user_id_idx", using: :btree
    t.index ["user_id"], name: "notifications_user_id_idx", using: :btree
  end

  create_table "number_formats", force: :cascade do |t|
    t.string   "name",                     null: false
    t.string   "data",        default: [], null: false, array: true
    t.integer  "sort_number", default: 0,  null: false
    t.datetime "created_at",               null: false
    t.datetime "updated_at",               null: false
    t.string   "format",      default: "", null: false
    t.index ["name"], name: "index_number_formats_on_name", unique: true, using: :btree
  end

  create_table "oauth_access_grants", force: :cascade do |t|
    t.integer  "resource_owner_id", null: false
    t.integer  "application_id",    null: false
    t.string   "token",             null: false
    t.integer  "expires_in",        null: false
    t.text     "redirect_uri",      null: false
    t.datetime "created_at",        null: false
    t.datetime "revoked_at"
    t.string   "scopes"
    t.index ["token"], name: "index_oauth_access_grants_on_token", unique: true, using: :btree
  end

  create_table "oauth_access_tokens", force: :cascade do |t|
    t.integer  "resource_owner_id",                   null: false
    t.integer  "application_id"
    t.string   "token",                               null: false
    t.string   "refresh_token"
    t.integer  "expires_in"
    t.datetime "revoked_at"
    t.datetime "created_at",                          null: false
    t.string   "scopes"
    t.string   "previous_refresh_token", default: "", null: false
    t.string   "description",            default: "", null: false
    t.index ["refresh_token"], name: "index_oauth_access_tokens_on_refresh_token", unique: true, using: :btree
    t.index ["resource_owner_id"], name: "index_oauth_access_tokens_on_resource_owner_id", using: :btree
    t.index ["token"], name: "index_oauth_access_tokens_on_token", unique: true, using: :btree
  end

  create_table "oauth_applications", force: :cascade do |t|
    t.string   "name",                               null: false
    t.string   "uid",                                null: false
    t.string   "secret",                             null: false
    t.text     "redirect_uri",                       null: false
    t.string   "scopes",       default: "",          null: false
    t.string   "aasm_state",   default: "published", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "owner_id"
    t.string   "owner_type"
    t.index ["owner_id", "owner_type"], name: "index_oauth_applications_on_owner_id_and_owner_type", using: :btree
    t.index ["uid"], name: "index_oauth_applications_on_uid", unique: true, using: :btree
  end

  create_table "organizations", force: :cascade do |t|
    t.string   "name",                                               null: false
    t.string   "url"
    t.string   "wikipedia_url"
    t.string   "twitter_username"
    t.string   "aasm_state",                   default: "published", null: false
    t.datetime "created_at",                                         null: false
    t.datetime "updated_at",                                         null: false
    t.string   "name_kana",                    default: "",          null: false
    t.string   "name_en",                      default: "",          null: false
    t.string   "url_en",                       default: "",          null: false
    t.string   "wikipedia_url_en",             default: "",          null: false
    t.string   "twitter_username_en",          default: "",          null: false
    t.integer  "favorite_organizations_count", default: 0,           null: false
    t.integer  "staffs_count",                 default: 0,           null: false
    t.index ["aasm_state"], name: "index_organizations_on_aasm_state", using: :btree
    t.index ["favorite_organizations_count"], name: "index_organizations_on_favorite_organizations_count", using: :btree
    t.index ["name"], name: "index_organizations_on_name", unique: true, using: :btree
    t.index ["staffs_count"], name: "index_organizations_on_staffs_count", using: :btree
  end

  create_table "people", force: :cascade do |t|
    t.integer  "prefecture_id"
    t.string   "name",                                        null: false
    t.string   "name_kana",             default: "",          null: false
    t.string   "nickname"
    t.string   "gender"
    t.string   "url"
    t.string   "wikipedia_url"
    t.string   "twitter_username"
    t.date     "birthday"
    t.string   "blood_type"
    t.integer  "height"
    t.string   "aasm_state",            default: "published", null: false
    t.datetime "created_at",                                  null: false
    t.datetime "updated_at",                                  null: false
    t.string   "name_en",               default: "",          null: false
    t.string   "nickname_en",           default: "",          null: false
    t.string   "url_en",                default: "",          null: false
    t.string   "wikipedia_url_en",      default: "",          null: false
    t.string   "twitter_username_en",   default: "",          null: false
    t.integer  "favorite_people_count", default: 0,           null: false
    t.integer  "casts_count",           default: 0,           null: false
    t.integer  "staffs_count",          default: 0,           null: false
    t.index ["aasm_state"], name: "index_people_on_aasm_state", using: :btree
    t.index ["casts_count"], name: "index_people_on_casts_count", using: :btree
    t.index ["favorite_people_count"], name: "index_people_on_favorite_people_count", using: :btree
    t.index ["name"], name: "index_people_on_name", unique: true, using: :btree
    t.index ["prefecture_id"], name: "index_people_on_prefecture_id", using: :btree
    t.index ["staffs_count"], name: "index_people_on_staffs_count", using: :btree
  end

  create_table "prefectures", force: :cascade do |t|
    t.string   "name",       null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["name"], name: "index_prefectures_on_name", unique: true, using: :btree
  end

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
    t.index ["user_id"], name: "profiles_user_id_idx", using: :btree
    t.index ["user_id"], name: "profiles_user_id_key", unique: true, using: :btree
  end

  create_table "programs", force: :cascade do |t|
    t.integer  "channel_id",                           null: false
    t.integer  "episode_id",                           null: false
    t.integer  "work_id",                              null: false
    t.datetime "started_at",                           null: false
    t.datetime "sc_last_update"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "sc_pid"
    t.boolean  "rebroadcast",    default: false,       null: false
    t.string   "aasm_state",     default: "published", null: false
    t.index ["aasm_state"], name: "index_programs_on_aasm_state", using: :btree
    t.index ["channel_id"], name: "programs_channel_id_idx", using: :btree
    t.index ["episode_id"], name: "programs_episode_id_idx", using: :btree
    t.index ["sc_pid"], name: "index_programs_on_sc_pid", unique: true, using: :btree
    t.index ["work_id"], name: "programs_work_id_idx", using: :btree
  end

  create_table "providers", force: :cascade do |t|
    t.integer  "user_id",                      null: false
    t.string   "name",             limit: 510, null: false
    t.string   "uid",              limit: 510, null: false
    t.string   "token",            limit: 510, null: false
    t.integer  "token_expires_at"
    t.string   "token_secret",     limit: 510
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["name", "uid"], name: "providers_name_uid_key", unique: true, using: :btree
    t.index ["user_id"], name: "providers_user_id_idx", using: :btree
  end

  create_table "receptions", force: :cascade do |t|
    t.integer  "user_id",    null: false
    t.integer  "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["channel_id"], name: "receptions_channel_id_idx", using: :btree
    t.index ["user_id", "channel_id"], name: "receptions_user_id_channel_id_key", unique: true, using: :btree
    t.index ["user_id"], name: "receptions_user_id_idx", using: :btree
  end

  create_table "seasons", force: :cascade do |t|
    t.string   "name",        limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "sort_number",             null: false
    t.integer  "year",                    null: false
    t.index ["sort_number"], name: "index_seasons_on_sort_number", unique: true, using: :btree
    t.index ["year", "name"], name: "index_seasons_on_year_and_name", unique: true, using: :btree
    t.index ["year"], name: "index_seasons_on_year", using: :btree
  end

  create_table "series", force: :cascade do |t|
    t.string   "name",                                     null: false
    t.string   "name_ro",            default: "",          null: false
    t.string   "name_en",            default: "",          null: false
    t.string   "aasm_state",         default: "published", null: false
    t.integer  "series_works_count", default: 0,           null: false
    t.datetime "created_at",                               null: false
    t.datetime "updated_at",                               null: false
    t.index ["name"], name: "index_series_on_name", unique: true, using: :btree
  end

  create_table "series_works", force: :cascade do |t|
    t.integer  "series_id",                        null: false
    t.integer  "work_id",                          null: false
    t.string   "summary",    default: "",          null: false
    t.string   "summary_en", default: "",          null: false
    t.string   "aasm_state", default: "published", null: false
    t.datetime "created_at",                       null: false
    t.datetime "updated_at",                       null: false
    t.index ["series_id", "work_id"], name: "index_series_works_on_series_id_and_work_id", unique: true, using: :btree
    t.index ["series_id"], name: "index_series_works_on_series_id", using: :btree
    t.index ["work_id"], name: "index_series_works_on_work_id", using: :btree
  end

  create_table "sessions", force: :cascade do |t|
    t.string   "session_id", limit: 510, null: false
    t.text     "data"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["session_id"], name: "sessions_session_id_key", unique: true, using: :btree
  end

  create_table "settings", force: :cascade do |t|
    t.integer  "user_id",                                        null: false
    t.boolean  "hide_checkin_comment",          default: true,   null: false
    t.datetime "created_at",                                     null: false
    t.datetime "updated_at",                                     null: false
    t.boolean  "share_record_to_twitter",       default: false
    t.boolean  "share_record_to_facebook",      default: false
    t.string   "programs_sort_type",            default: "",     null: false
    t.string   "display_option_work_list",      default: "list", null: false
    t.string   "display_option_user_work_list", default: "list", null: false
    t.index ["user_id"], name: "index_settings_on_user_id", using: :btree
  end

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
    t.string   "name_en",       default: "",          null: false
    t.string   "role_other_en", default: "",          null: false
    t.index ["aasm_state"], name: "index_staffs_on_aasm_state", using: :btree
    t.index ["resource_id", "resource_type"], name: "index_staffs_on_resource_id_and_resource_type", using: :btree
    t.index ["sort_number"], name: "index_staffs_on_sort_number", using: :btree
    t.index ["work_id"], name: "index_staffs_on_work_id", using: :btree
  end

  create_table "statuses", force: :cascade do |t|
    t.integer  "user_id",                          null: false
    t.integer  "work_id",                          null: false
    t.integer  "kind",                             null: false
    t.integer  "likes_count",          default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer  "oauth_application_id"
    t.index ["oauth_application_id"], name: "index_statuses_on_oauth_application_id", using: :btree
    t.index ["user_id"], name: "statuses_user_id_idx", using: :btree
    t.index ["work_id"], name: "statuses_work_id_idx", using: :btree
  end

  create_table "syobocal_alerts", force: :cascade do |t|
    t.integer  "work_id"
    t.integer  "kind",                        null: false
    t.integer  "sc_prog_item_id"
    t.string   "sc_sub_title",    limit: 255
    t.string   "sc_prog_comment", limit: 255
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["kind"], name: "index_syobocal_alerts_on_kind", using: :btree
    t.index ["sc_prog_item_id"], name: "index_syobocal_alerts_on_sc_prog_item_id", using: :btree
  end

  create_table "tips", force: :cascade do |t|
    t.integer  "target",                              null: false
    t.string   "slug",       limit: 255,              null: false
    t.string   "title",      limit: 255,              null: false
    t.string   "icon_name",  limit: 255,              null: false
    t.datetime "created_at",                          null: false
    t.datetime "updated_at",                          null: false
    t.string   "title_en",               default: "", null: false
    t.index ["slug"], name: "index_tips_on_slug", unique: true, using: :btree
  end

  create_table "twitter_bots", force: :cascade do |t|
    t.string   "name",       limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["name"], name: "twitter_bots_name_key", unique: true, using: :btree
  end

  create_table "userland_categories", force: :cascade do |t|
    t.string   "name",                                null: false
    t.string   "name_en",                             null: false
    t.integer  "sort_number",             default: 0, null: false
    t.integer  "userland_projects_count", default: 0, null: false
    t.datetime "created_at",                          null: false
    t.datetime "updated_at",                          null: false
  end

  create_table "userland_project_members", force: :cascade do |t|
    t.integer  "user_id",             null: false
    t.integer  "userland_project_id", null: false
    t.datetime "created_at",          null: false
    t.datetime "updated_at",          null: false
    t.index ["user_id", "userland_project_id"], name: "index_userland_pm_on_uid_and_userland_pid", unique: true, using: :btree
    t.index ["user_id"], name: "index_userland_project_members_on_user_id", using: :btree
    t.index ["userland_project_id"], name: "index_userland_project_members_on_userland_project_id", using: :btree
  end

  create_table "userland_projects", force: :cascade do |t|
    t.integer  "userland_category_id",                 null: false
    t.string   "name",                                 null: false
    t.string   "summary",                              null: false
    t.text     "description",                          null: false
    t.string   "url",                                  null: false
    t.string   "icon_file_name"
    t.string   "icon_content_type"
    t.integer  "icon_file_size"
    t.datetime "icon_updated_at"
    t.boolean  "available",            default: false, null: false
    t.datetime "created_at",                           null: false
    t.datetime "updated_at",                           null: false
    t.index ["userland_category_id"], name: "index_userland_projects_on_userland_category_id", using: :btree
  end

  create_table "users", force: :cascade do |t|
    t.string   "username",                limit: 510,              null: false
    t.string   "email",                   limit: 510,              null: false
    t.integer  "role",                                             null: false
    t.string   "encrypted_password",      limit: 510, default: "", null: false
    t.datetime "remember_created_at"
    t.integer  "sign_in_count",                       default: 0,  null: false
    t.datetime "current_sign_in_at"
    t.datetime "last_sign_in_at"
    t.string   "current_sign_in_ip",      limit: 510
    t.string   "last_sign_in_ip",         limit: 510
    t.string   "confirmation_token",      limit: 510
    t.datetime "confirmed_at"
    t.datetime "confirmation_sent_at"
    t.string   "unconfirmed_email",       limit: 510
    t.integer  "checkins_count",                      default: 0,  null: false
    t.integer  "notifications_count",                 default: 0,  null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string   "time_zone",                                        null: false
    t.string   "locale",                                           null: false
    t.string   "reset_password_token"
    t.datetime "reset_password_sent_at"
    t.datetime "record_cache_expired_at"
    t.datetime "status_cache_expired_at"
    t.index ["confirmation_token"], name: "users_confirmation_token_key", unique: true, using: :btree
    t.index ["email"], name: "users_email_key", unique: true, using: :btree
    t.index ["username"], name: "users_username_key", unique: true, using: :btree
  end

  create_table "versions", force: :cascade do |t|
    t.string   "item_type",  limit: 510, null: false
    t.integer  "item_id",                null: false
    t.string   "event",      limit: 510, null: false
    t.string   "whodunnit",  limit: 510
    t.text     "object"
    t.datetime "created_at"
  end

  create_table "work_images", force: :cascade do |t|
    t.integer  "work_id",                              null: false
    t.integer  "user_id",                              null: false
    t.string   "attachment_file_name",                 null: false
    t.integer  "attachment_file_size",                 null: false
    t.string   "attachment_content_type",              null: false
    t.datetime "attachment_updated_at",                null: false
    t.string   "copyright",               default: "", null: false
    t.string   "asin",                    default: "", null: false
    t.datetime "created_at",                           null: false
    t.datetime "updated_at",                           null: false
    t.index ["user_id"], name: "index_work_images_on_user_id", using: :btree
    t.index ["work_id"], name: "index_work_images_on_work_id", using: :btree
  end

  create_table "works", force: :cascade do |t|
    t.integer  "season_id"
    t.integer  "sc_tid"
    t.string   "title",                 limit: 510,                       null: false
    t.integer  "media",                                                   null: false
    t.string   "official_site_url",     limit: 510, default: "",          null: false
    t.string   "wikipedia_url",         limit: 510, default: "",          null: false
    t.integer  "episodes_count",                    default: 0,           null: false
    t.integer  "watchers_count",                    default: 0,           null: false
    t.date     "released_at"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string   "twitter_username",      limit: 510
    t.string   "twitter_hashtag",       limit: 510
    t.string   "released_at_about"
    t.string   "aasm_state",                        default: "published", null: false
    t.integer  "number_format_id"
    t.string   "title_kana",                        default: "",          null: false
    t.string   "title_ro",                          default: "",          null: false
    t.string   "title_en",                          default: "",          null: false
    t.string   "official_site_url_en",              default: "",          null: false
    t.string   "wikipedia_url_en",                  default: "",          null: false
    t.text     "synopsis",                          default: "",          null: false
    t.text     "synopsis_en",                       default: "",          null: false
    t.string   "synopsis_source",                   default: "",          null: false
    t.string   "synopsis_source_en",                default: "",          null: false
    t.integer  "mal_anime_id"
    t.string   "facebook_og_image_url",             default: "",          null: false
    t.string   "twitter_image_url",                 default: "",          null: false
    t.string   "recommended_image_url",             default: "",          null: false
    t.integer  "season_year"
    t.integer  "season_name"
    t.index ["aasm_state"], name: "index_works_on_aasm_state", using: :btree
    t.index ["number_format_id"], name: "index_works_on_number_format_id", using: :btree
    t.index ["sc_tid"], name: "works_sc_tid_key", unique: true, using: :btree
    t.index ["season_id"], name: "works_season_id_idx", using: :btree
    t.index ["season_year", "season_name"], name: "index_works_on_season_year_and_season_name", using: :btree
    t.index ["season_year"], name: "index_works_on_season_year", using: :btree
  end

  add_foreign_key "activities", "users", name: "activities_user_id_fk", on_delete: :cascade
  add_foreign_key "casts", "characters"
  add_foreign_key "casts", "people"
  add_foreign_key "casts", "works"
  add_foreign_key "channel_works", "channels", name: "channel_works_channel_id_fk", on_delete: :cascade
  add_foreign_key "channel_works", "users", name: "channel_works_user_id_fk", on_delete: :cascade
  add_foreign_key "channel_works", "works", name: "channel_works_work_id_fk", on_delete: :cascade
  add_foreign_key "channels", "channel_groups", name: "channels_channel_group_id_fk", on_delete: :cascade
  add_foreign_key "character_images", "characters"
  add_foreign_key "character_images", "users"
  add_foreign_key "characters", "series"
  add_foreign_key "checkins", "episodes", name: "checkins_episode_id_fk", on_delete: :cascade
  add_foreign_key "checkins", "multiple_records"
  add_foreign_key "checkins", "oauth_applications"
  add_foreign_key "checkins", "users", name: "checkins_user_id_fk", on_delete: :cascade
  add_foreign_key "checkins", "works", name: "checkins_work_id_fk"
  add_foreign_key "collection_item_agreements", "collection_items"
  add_foreign_key "collection_item_agreements", "users"
  add_foreign_key "collection_items", "collections"
  add_foreign_key "collection_items", "works"
  add_foreign_key "collections", "users"
  add_foreign_key "comments", "checkins", name: "comments_checkin_id_fk", on_delete: :cascade
  add_foreign_key "comments", "users", name: "comments_user_id_fk", on_delete: :cascade
  add_foreign_key "comments", "works"
  add_foreign_key "cover_images", "works", name: "cover_images_work_id_fk", on_delete: :cascade
  add_foreign_key "db_activities", "users"
  add_foreign_key "db_comments", "users"
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
  add_foreign_key "draft_works", "number_formats"
  add_foreign_key "draft_works", "seasons"
  add_foreign_key "draft_works", "works"
  add_foreign_key "edit_request_comments", "edit_requests", on_delete: :cascade
  add_foreign_key "edit_request_comments", "users", on_delete: :cascade
  add_foreign_key "edit_request_participants", "edit_requests"
  add_foreign_key "edit_request_participants", "users"
  add_foreign_key "edit_requests", "users", on_delete: :cascade
  add_foreign_key "email_notifications", "users"
  add_foreign_key "episodes", "episodes", column: "prev_episode_id"
  add_foreign_key "episodes", "works", name: "episodes_work_id_fk", on_delete: :cascade
  add_foreign_key "favorite_characters", "characters"
  add_foreign_key "favorite_characters", "users"
  add_foreign_key "favorite_organizations", "organizations"
  add_foreign_key "favorite_organizations", "users"
  add_foreign_key "favorite_people", "people"
  add_foreign_key "favorite_people", "users"
  add_foreign_key "finished_tips", "tips", name: "finished_tips_tip_id_fk", on_delete: :cascade
  add_foreign_key "finished_tips", "users", name: "finished_tips_user_id_fk", on_delete: :cascade
  add_foreign_key "follows", "users", column: "following_id", name: "follows_following_id_fk", on_delete: :cascade
  add_foreign_key "follows", "users", name: "follows_user_id_fk", on_delete: :cascade
  add_foreign_key "forum_comments", "forum_posts"
  add_foreign_key "forum_comments", "users"
  add_foreign_key "forum_post_participants", "forum_posts"
  add_foreign_key "forum_post_participants", "users"
  add_foreign_key "forum_posts", "forum_categories"
  add_foreign_key "forum_posts", "users"
  add_foreign_key "items", "works", name: "items_work_id_fk", on_delete: :cascade
  add_foreign_key "latest_statuses", "episodes", column: "next_episode_id"
  add_foreign_key "latest_statuses", "users"
  add_foreign_key "latest_statuses", "works"
  add_foreign_key "likes", "users", name: "likes_user_id_fk", on_delete: :cascade
  add_foreign_key "multiple_records", "users"
  add_foreign_key "multiple_records", "works"
  add_foreign_key "mute_users", "users"
  add_foreign_key "mute_users", "users", column: "muted_user_id"
  add_foreign_key "notifications", "users", column: "action_user_id", name: "notifications_action_user_id_fk", on_delete: :cascade
  add_foreign_key "notifications", "users", name: "notifications_user_id_fk", on_delete: :cascade
  add_foreign_key "oauth_access_grants", "oauth_applications", column: "application_id"
  add_foreign_key "oauth_access_grants", "users", column: "resource_owner_id"
  add_foreign_key "oauth_access_tokens", "oauth_applications", column: "application_id"
  add_foreign_key "oauth_access_tokens", "users", column: "resource_owner_id"
  add_foreign_key "people", "prefectures"
  add_foreign_key "profiles", "users", name: "profiles_user_id_fk", on_delete: :cascade
  add_foreign_key "programs", "channels", name: "programs_channel_id_fk", on_delete: :cascade
  add_foreign_key "programs", "episodes", name: "programs_episode_id_fk", on_delete: :cascade
  add_foreign_key "programs", "works", name: "programs_work_id_fk", on_delete: :cascade
  add_foreign_key "providers", "users", name: "providers_user_id_fk", on_delete: :cascade
  add_foreign_key "receptions", "channels", name: "receptions_channel_id_fk", on_delete: :cascade
  add_foreign_key "receptions", "users", name: "receptions_user_id_fk", on_delete: :cascade
  add_foreign_key "series_works", "series"
  add_foreign_key "series_works", "works"
  add_foreign_key "settings", "users"
  add_foreign_key "staffs", "works"
  add_foreign_key "statuses", "oauth_applications"
  add_foreign_key "statuses", "users", name: "statuses_user_id_fk", on_delete: :cascade
  add_foreign_key "statuses", "works", name: "statuses_work_id_fk", on_delete: :cascade
  add_foreign_key "syobocal_alerts", "works", name: "syobocal_alerts_work_id_fk", on_delete: :cascade
  add_foreign_key "userland_project_members", "userland_projects"
  add_foreign_key "userland_project_members", "users"
  add_foreign_key "userland_projects", "userland_categories"
  add_foreign_key "work_images", "users"
  add_foreign_key "work_images", "works"
  add_foreign_key "works", "number_formats"
  add_foreign_key "works", "seasons", name: "works_season_id_fk", on_delete: :cascade
end
