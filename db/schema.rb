# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema.define(version: 2021_09_19_175411) do

  # These are extensions that must be enabled in order to support this database
  enable_extension "citext"
  enable_extension "pg_stat_statements"
  enable_extension "plpgsql"

  create_table "activities", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "recipient_id"
    t.string "recipient_type", limit: 510
    t.bigint "trackable_id", null: false
    t.string "trackable_type", limit: 510, null: false
    t.string "action", limit: 510
    t.datetime "created_at"
    t.datetime "updated_at"
    t.bigint "work_id"
    t.bigint "episode_id"
    t.bigint "status_id"
    t.bigint "episode_record_id"
    t.bigint "multiple_episode_record_id"
    t.bigint "work_record_id"
    t.bigint "activity_group_id", null: false
    t.datetime "migrated_at"
    t.datetime "mer_processed_at"
    t.index ["activity_group_id", "created_at"], name: "index_activities_on_activity_group_id_and_created_at"
    t.index ["activity_group_id"], name: "index_activities_on_activity_group_id"
    t.index ["created_at"], name: "index_activities_on_created_at"
    t.index ["episode_id"], name: "index_activities_on_episode_id"
    t.index ["episode_record_id"], name: "index_activities_on_episode_record_id"
    t.index ["multiple_episode_record_id"], name: "index_activities_on_multiple_episode_record_id"
    t.index ["status_id"], name: "index_activities_on_status_id"
    t.index ["trackable_id", "trackable_type"], name: "index_activities_on_trackable_id_and_trackable_type"
    t.index ["user_id"], name: "activities_user_id_idx"
    t.index ["work_id"], name: "index_activities_on_work_id"
    t.index ["work_record_id"], name: "index_activities_on_work_record_id"
  end

  create_table "activity_groups", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.string "itemable_type", null: false
    t.boolean "single", default: false, null: false
    t.integer "activities_count", default: 0, null: false
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.index ["created_at"], name: "index_activity_groups_on_created_at"
    t.index ["user_id"], name: "index_activity_groups_on_user_id"
  end

  create_table "casts", force: :cascade do |t|
    t.bigint "person_id", null: false
    t.bigint "work_id", null: false
    t.string "name", null: false
    t.string "aasm_state", default: "published", null: false
    t.integer "sort_number", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.bigint "character_id", null: false
    t.string "name_en", default: "", null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["aasm_state"], name: "index_casts_on_aasm_state"
    t.index ["character_id"], name: "index_casts_on_character_id"
    t.index ["deleted_at"], name: "index_casts_on_deleted_at"
    t.index ["person_id"], name: "index_casts_on_person_id"
    t.index ["sort_number"], name: "index_casts_on_sort_number"
    t.index ["unpublished_at"], name: "index_casts_on_unpublished_at"
    t.index ["work_id"], name: "index_casts_on_work_id"
  end

  create_table "channel_groups", force: :cascade do |t|
    t.string "sc_chgid", limit: 510
    t.string "name", limit: 510, null: false
    t.integer "sort_number", default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["sc_chgid"], name: "channel_groups_sc_chgid_key", unique: true
    t.index ["unpublished_at"], name: "index_channel_groups_on_unpublished_at"
  end

  create_table "channel_works", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "work_id", null: false
    t.bigint "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["channel_id"], name: "channel_works_channel_id_idx"
    t.index ["user_id", "work_id", "channel_id"], name: "channel_works_user_id_work_id_channel_id_key", unique: true
    t.index ["user_id"], name: "channel_works_user_id_idx"
    t.index ["work_id"], name: "channel_works_work_id_idx"
  end

  create_table "channels", force: :cascade do |t|
    t.bigint "channel_group_id", null: false
    t.integer "sc_chid"
    t.string "name", null: false, collation: "C"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean "vod", default: false
    t.string "aasm_state", default: "published", null: false
    t.integer "sort_number", default: 0, null: false
    t.datetime "deleted_at"
    t.string "name_alter", default: "", null: false
    t.datetime "unpublished_at"
    t.index ["channel_group_id"], name: "channels_channel_group_id_idx"
    t.index ["deleted_at"], name: "index_channels_on_deleted_at"
    t.index ["sc_chid"], name: "channels_sc_chid_key", unique: true
    t.index ["sort_number"], name: "index_channels_on_sort_number"
    t.index ["unpublished_at"], name: "index_channels_on_unpublished_at"
    t.index ["vod"], name: "index_channels_on_vod"
  end

  create_table "character_favorites", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "character_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["character_id"], name: "index_character_favorites_on_character_id"
    t.index ["user_id", "character_id"], name: "index_character_favorites_on_user_id_and_character_id", unique: true
    t.index ["user_id"], name: "index_character_favorites_on_user_id"
  end

  create_table "character_images", force: :cascade do |t|
    t.bigint "character_id", null: false
    t.bigint "user_id", null: false
    t.string "attachment_file_name", null: false
    t.integer "attachment_file_size", null: false
    t.string "attachment_content_type", null: false
    t.datetime "attachment_updated_at", null: false
    t.string "copyright", default: "", null: false
    t.string "asin", default: "", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["character_id"], name: "index_character_images_on_character_id"
    t.index ["user_id"], name: "index_character_images_on_user_id"
  end

  create_table "characters", force: :cascade do |t|
    t.string "name", null: false
    t.string "name_kana", default: "", null: false
    t.string "name_en", default: "", null: false
    t.string "nickname", default: "", null: false
    t.string "nickname_en", default: "", null: false
    t.string "birthday", default: "", null: false
    t.string "birthday_en", default: "", null: false
    t.string "age", default: "", null: false
    t.string "age_en", default: "", null: false
    t.string "blood_type", default: "", null: false
    t.string "blood_type_en", default: "", null: false
    t.string "height", default: "", null: false
    t.string "height_en", default: "", null: false
    t.string "weight", default: "", null: false
    t.string "weight_en", default: "", null: false
    t.string "nationality", default: "", null: false
    t.string "nationality_en", default: "", null: false
    t.string "occupation", default: "", null: false
    t.string "occupation_en", default: "", null: false
    t.text "description", default: "", null: false
    t.text "description_en", default: "", null: false
    t.string "aasm_state", default: "published", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "description_source", default: "", null: false
    t.string "description_source_en", default: "", null: false
    t.integer "favorite_users_count", default: 0, null: false
    t.bigint "series_id"
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["deleted_at"], name: "index_characters_on_deleted_at"
    t.index ["favorite_users_count"], name: "index_characters_on_favorite_users_count"
    t.index ["name", "series_id"], name: "index_characters_on_name_and_series_id", unique: true
    t.index ["series_id"], name: "index_characters_on_series_id"
    t.index ["unpublished_at"], name: "index_characters_on_unpublished_at"
  end

  create_table "collection_items", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "collection_id", null: false
    t.bigint "work_id", null: false
    t.text "body", default: "", null: false
    t.integer "likes_count", default: 0, null: false
    t.integer "position", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at"
    t.index ["collection_id", "work_id"], name: "index_collection_items_on_collection_id_and_work_id", unique: true
    t.index ["collection_id"], name: "index_collection_items_on_collection_id"
    t.index ["deleted_at"], name: "index_collection_items_on_deleted_at"
    t.index ["user_id"], name: "index_collection_items_on_user_id"
    t.index ["work_id"], name: "index_collection_items_on_work_id"
  end

  create_table "collections", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.string "title", null: false
    t.string "description", default: "", null: false
    t.integer "likes_count", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at"
    t.integer "collection_items_count", default: 0, null: false
    t.index ["deleted_at"], name: "index_collections_on_deleted_at"
    t.index ["user_id"], name: "index_collections_on_user_id"
  end

  create_table "comments", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "episode_record_id", null: false
    t.text "body", null: false
    t.integer "likes_count", default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.bigint "work_id"
    t.string "locale", default: "other", null: false
    t.index ["episode_record_id"], name: "comments_checkin_id_idx"
    t.index ["locale"], name: "index_comments_on_locale"
    t.index ["user_id"], name: "comments_user_id_idx"
    t.index ["work_id"], name: "index_comments_on_work_id"
  end

  create_table "db_activities", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "trackable_id", null: false
    t.string "trackable_type", null: false
    t.string "action", null: false
    t.json "parameters"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.bigint "root_resource_id"
    t.string "root_resource_type"
    t.bigint "object_id"
    t.string "object_type"
    t.index ["object_id", "object_type"], name: "index_db_activities_on_object_id_and_object_type"
    t.index ["root_resource_id", "root_resource_type"], name: "index_db_activities_on_root_resource_id_and_root_resource_type"
    t.index ["trackable_id", "trackable_type"], name: "index_db_activities_on_trackable_id_and_trackable_type"
  end

  create_table "db_comments", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "resource_id", null: false
    t.string "resource_type", null: false
    t.text "body", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "locale", default: "other", null: false
    t.index ["locale"], name: "index_db_comments_on_locale"
    t.index ["resource_id", "resource_type"], name: "index_db_comments_on_resource_id_and_resource_type"
    t.index ["user_id"], name: "index_db_comments_on_user_id"
  end

  create_table "delayed_jobs", force: :cascade do |t|
    t.integer "priority", default: 0, null: false
    t.integer "attempts", default: 0, null: false
    t.text "handler", null: false
    t.text "last_error"
    t.datetime "run_at"
    t.datetime "locked_at"
    t.datetime "failed_at"
    t.string "locked_by"
    t.string "queue"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["priority", "run_at"], name: "delayed_jobs_priority"
  end

  create_table "email_confirmations", force: :cascade do |t|
    t.bigint "user_id"
    t.citext "email", null: false
    t.string "event", null: false
    t.string "token", null: false
    t.string "back"
    t.datetime "expires_at", null: false
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.index ["token"], name: "index_email_confirmations_on_token", unique: true
    t.index ["user_id"], name: "index_email_confirmations_on_user_id"
  end

  create_table "email_notifications", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.string "unsubscription_key", null: false
    t.boolean "event_followed_user", default: true, null: false
    t.boolean "event_liked_episode_record", default: true, null: false
    t.boolean "event_friends_joined", default: true, null: false
    t.boolean "event_next_season_came", default: true, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.boolean "event_favorite_works_added", default: true, null: false
    t.boolean "event_related_works_added", default: true, null: false
    t.index ["unsubscription_key"], name: "index_email_notifications_on_unsubscription_key", unique: true
    t.index ["user_id"], name: "index_email_notifications_on_user_id", unique: true
  end

  create_table "episode_records", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "episode_id", null: false
    t.text "body"
    t.boolean "modify_body", default: false, null: false
    t.string "twitter_url_hash", limit: 510
    t.string "facebook_url_hash", limit: 510
    t.integer "twitter_click_count", default: 0, null: false
    t.integer "facebook_click_count", default: 0, null: false
    t.integer "comments_count", default: 0, null: false
    t.integer "likes_count", default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean "shared_twitter", default: false, null: false
    t.boolean "shared_facebook", default: false, null: false
    t.bigint "work_id", null: false
    t.float "rating"
    t.bigint "multiple_episode_record_id"
    t.bigint "oauth_application_id"
    t.string "rating_state"
    t.bigint "review_id"
    t.string "aasm_state", default: "published", null: false
    t.string "locale", default: "other", null: false
    t.bigint "record_id", null: false
    t.datetime "deleted_at"
    t.index ["episode_id", "deleted_at"], name: "index_episode_records_on_episode_id_and_deleted_at"
    t.index ["facebook_url_hash"], name: "checkins_facebook_url_hash_key", unique: true
    t.index ["locale"], name: "index_episode_records_on_locale"
    t.index ["multiple_episode_record_id"], name: "index_episode_records_on_multiple_episode_record_id"
    t.index ["oauth_application_id"], name: "index_episode_records_on_oauth_application_id"
    t.index ["rating_state"], name: "index_episode_records_on_rating_state"
    t.index ["record_id"], name: "index_episode_records_on_record_id", unique: true
    t.index ["review_id"], name: "index_episode_records_on_review_id"
    t.index ["twitter_url_hash"], name: "checkins_twitter_url_hash_key", unique: true
    t.index ["user_id"], name: "checkins_user_id_idx"
    t.index ["work_id"], name: "index_episode_records_on_work_id"
  end

  create_table "episodes", force: :cascade do |t|
    t.bigint "work_id", null: false
    t.string "number", limit: 510
    t.integer "sort_number", default: 0, null: false
    t.integer "sc_count"
    t.string "title", limit: 510
    t.integer "episode_records_count", default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.bigint "prev_episode_id"
    t.string "aasm_state", default: "published", null: false
    t.boolean "fetch_syobocal", default: false, null: false
    t.float "raw_number"
    t.string "title_ro", default: "", null: false
    t.string "title_en", default: "", null: false
    t.integer "episode_record_bodies_count", default: 0, null: false
    t.float "score"
    t.integer "ratings_count", default: 0, null: false
    t.float "satisfaction_rate"
    t.string "number_en", default: "", null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["aasm_state"], name: "index_episodes_on_aasm_state"
    t.index ["deleted_at"], name: "index_episodes_on_deleted_at"
    t.index ["prev_episode_id"], name: "index_episodes_on_prev_episode_id"
    t.index ["ratings_count"], name: "index_episodes_on_ratings_count"
    t.index ["satisfaction_rate", "ratings_count"], name: "index_episodes_on_satisfaction_rate_and_ratings_count"
    t.index ["satisfaction_rate"], name: "index_episodes_on_satisfaction_rate"
    t.index ["score"], name: "index_episodes_on_score"
    t.index ["unpublished_at"], name: "index_episodes_on_unpublished_at"
    t.index ["work_id", "sc_count"], name: "episodes_work_id_sc_count_key", unique: true
    t.index ["work_id"], name: "episodes_work_id_idx"
  end

  create_table "faq_categories", force: :cascade do |t|
    t.string "name", null: false
    t.string "locale", null: false
    t.integer "sort_number", default: 0, null: false
    t.string "aasm_state", default: "published", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at"
    t.index ["deleted_at"], name: "index_faq_categories_on_deleted_at"
    t.index ["locale"], name: "index_faq_categories_on_locale"
  end

  create_table "faq_contents", force: :cascade do |t|
    t.bigint "faq_category_id", null: false
    t.string "question", null: false
    t.text "answer", null: false
    t.string "locale", null: false
    t.integer "sort_number", default: 0, null: false
    t.string "aasm_state", default: "published", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at"
    t.index ["deleted_at"], name: "index_faq_contents_on_deleted_at"
    t.index ["faq_category_id"], name: "index_faq_contents_on_faq_category_id"
    t.index ["locale"], name: "index_faq_contents_on_locale"
  end

  create_table "finished_tips", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "tip_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "tip_id"], name: "index_finished_tips_on_user_id_and_tip_id", unique: true
  end

  create_table "flashes", force: :cascade do |t|
    t.string "client_uuid", null: false
    t.json "data"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["client_uuid"], name: "index_flashes_on_client_uuid", unique: true
  end

  create_table "follows", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "following_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["following_id"], name: "follows_following_id_idx"
    t.index ["user_id", "following_id"], name: "follows_user_id_following_id_key", unique: true
    t.index ["user_id"], name: "follows_user_id_idx"
  end

  create_table "forum_categories", force: :cascade do |t|
    t.string "slug", null: false
    t.string "name", null: false
    t.string "name_en", null: false
    t.string "description", null: false
    t.string "description_en", null: false
    t.string "postable_role", null: false
    t.integer "sort_number", null: false
    t.integer "forum_posts_count", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["slug"], name: "index_forum_categories_on_slug", unique: true
  end

  create_table "forum_comments", force: :cascade do |t|
    t.bigint "user_id"
    t.bigint "forum_post_id", null: false
    t.text "body", null: false
    t.datetime "edited_at", comment: "The datetime which user has changed body."
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "locale", default: "other", null: false
    t.index ["forum_post_id"], name: "index_forum_comments_on_forum_post_id"
    t.index ["locale"], name: "index_forum_comments_on_locale"
    t.index ["user_id"], name: "index_forum_comments_on_user_id"
  end

  create_table "forum_post_participants", force: :cascade do |t|
    t.bigint "forum_post_id", null: false
    t.bigint "user_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["forum_post_id", "user_id"], name: "index_forum_post_participants_on_forum_post_id_and_user_id", unique: true
    t.index ["forum_post_id"], name: "index_forum_post_participants_on_forum_post_id"
    t.index ["user_id"], name: "index_forum_post_participants_on_user_id"
  end

  create_table "forum_posts", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "forum_category_id", null: false
    t.string "title", null: false
    t.text "body", default: "", null: false
    t.integer "forum_comments_count", default: 0, null: false
    t.datetime "edited_at", comment: "The datetime which user has changed title, body and so on."
    t.datetime "last_commented_at", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "locale", default: "other", null: false
    t.index ["forum_category_id"], name: "index_forum_posts_on_forum_category_id"
    t.index ["locale"], name: "index_forum_posts_on_locale"
    t.index ["user_id"], name: "index_forum_posts_on_user_id"
  end

  create_table "gumroad_subscribers", force: :cascade do |t|
    t.string "gumroad_id", null: false
    t.string "gumroad_product_id", null: false
    t.string "gumroad_product_name", null: false
    t.string "gumroad_user_id"
    t.string "gumroad_user_email"
    t.string "gumroad_purchase_ids", null: false, array: true
    t.datetime "gumroad_created_at", null: false
    t.datetime "gumroad_cancelled_at"
    t.datetime "gumroad_user_requested_cancellation_at"
    t.datetime "gumroad_charge_occurrence_count"
    t.datetime "gumroad_ended_at"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["gumroad_id"], name: "index_gumroad_subscribers_on_gumroad_id", unique: true
    t.index ["gumroad_product_id"], name: "index_gumroad_subscribers_on_gumroad_product_id"
    t.index ["gumroad_user_id"], name: "index_gumroad_subscribers_on_gumroad_user_id"
  end

  create_table "internal_statistics", force: :cascade do |t|
    t.string "key", null: false
    t.float "value", null: false
    t.date "date", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["key", "date"], name: "index_internal_statistics_on_key_and_date", unique: true
    t.index ["key"], name: "index_internal_statistics_on_key"
  end

  create_table "library_entries", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "work_id", null: false
    t.bigint "next_episode_id"
    t.integer "kind"
    t.bigint "watched_episode_ids", default: [], null: false, array: true
    t.integer "position", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.bigint "status_id"
    t.bigint "program_id"
    t.bigint "next_slot_id"
    t.text "note", default: "", null: false
    t.index ["next_episode_id"], name: "index_library_entries_on_next_episode_id"
    t.index ["next_slot_id"], name: "index_library_entries_on_next_slot_id"
    t.index ["program_id"], name: "index_library_entries_on_program_id"
    t.index ["status_id"], name: "index_library_entries_on_status_id"
    t.index ["user_id", "position"], name: "index_library_entries_on_user_id_and_position"
    t.index ["user_id", "program_id"], name: "index_library_entries_on_user_id_and_program_id", unique: true
    t.index ["user_id", "work_id"], name: "index_library_entries_on_user_id_and_work_id", unique: true
    t.index ["user_id"], name: "index_library_entries_on_user_id"
    t.index ["work_id"], name: "index_library_entries_on_work_id"
  end

  create_table "likes", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "recipient_id", null: false
    t.string "recipient_type", limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["user_id"], name: "likes_user_id_idx"
  end

  create_table "multiple_episode_records", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "work_id", null: false
    t.integer "likes_count", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id"], name: "index_multiple_episode_records_on_user_id"
    t.index ["work_id"], name: "index_multiple_episode_records_on_work_id"
  end

  create_table "mute_users", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "muted_user_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["muted_user_id"], name: "index_mute_users_on_muted_user_id"
    t.index ["user_id", "muted_user_id"], name: "index_mute_users_on_user_id_and_muted_user_id", unique: true
    t.index ["user_id"], name: "index_mute_users_on_user_id"
  end

  create_table "notifications", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "action_user_id", null: false
    t.bigint "trackable_id", null: false
    t.string "trackable_type", limit: 510, null: false
    t.string "action", limit: 510, null: false
    t.boolean "read", default: false, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["action_user_id"], name: "notifications_action_user_id_idx"
    t.index ["user_id"], name: "notifications_user_id_idx"
  end

  create_table "number_formats", force: :cascade do |t|
    t.string "name", null: false
    t.string "data", default: [], null: false, array: true
    t.integer "sort_number", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "format", default: "", null: false
    t.index ["name"], name: "index_number_formats_on_name", unique: true
  end

  create_table "oauth_access_grants", force: :cascade do |t|
    t.bigint "resource_owner_id", null: false
    t.bigint "application_id", null: false
    t.string "token", null: false
    t.integer "expires_in", null: false
    t.text "redirect_uri", null: false
    t.datetime "created_at", null: false
    t.datetime "revoked_at"
    t.string "scopes"
    t.index ["token"], name: "index_oauth_access_grants_on_token", unique: true
  end

  create_table "oauth_access_tokens", force: :cascade do |t|
    t.bigint "resource_owner_id", null: false
    t.bigint "application_id"
    t.string "token", null: false
    t.string "refresh_token"
    t.integer "expires_in"
    t.datetime "revoked_at"
    t.datetime "created_at", null: false
    t.string "scopes"
    t.string "previous_refresh_token", default: "", null: false
    t.string "description", default: "", null: false
    t.index ["refresh_token"], name: "index_oauth_access_tokens_on_refresh_token", unique: true
    t.index ["resource_owner_id"], name: "index_oauth_access_tokens_on_resource_owner_id"
    t.index ["token"], name: "index_oauth_access_tokens_on_token", unique: true
  end

  create_table "oauth_applications", force: :cascade do |t|
    t.string "name", null: false
    t.string "uid", null: false
    t.string "secret", null: false
    t.text "redirect_uri", null: false
    t.string "scopes", default: "", null: false
    t.string "aasm_state", default: "published", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.bigint "owner_id"
    t.string "owner_type"
    t.boolean "confidential", default: true, null: false
    t.datetime "deleted_at"
    t.boolean "hide_social_login", default: false, null: false
    t.index ["deleted_at"], name: "index_oauth_applications_on_deleted_at"
    t.index ["owner_id", "owner_type"], name: "index_oauth_applications_on_owner_id_and_owner_type"
    t.index ["uid"], name: "index_oauth_applications_on_uid", unique: true
  end

  create_table "organization_favorites", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "organization_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "watched_works_count", default: 0, null: false
    t.index ["organization_id"], name: "index_organization_favorites_on_organization_id"
    t.index ["user_id", "organization_id"], name: "index_organization_favorites_on_user_id_and_organization_id", unique: true
    t.index ["user_id"], name: "index_organization_favorites_on_user_id"
    t.index ["watched_works_count"], name: "index_organization_favorites_on_watched_works_count"
  end

  create_table "organizations", force: :cascade do |t|
    t.string "name", null: false
    t.string "url"
    t.string "wikipedia_url"
    t.string "twitter_username"
    t.string "aasm_state", default: "published", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "name_kana", default: "", null: false
    t.string "name_en", default: "", null: false
    t.string "url_en", default: "", null: false
    t.string "wikipedia_url_en", default: "", null: false
    t.string "twitter_username_en", default: "", null: false
    t.integer "favorite_users_count", default: 0, null: false
    t.integer "staffs_count", default: 0, null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["aasm_state"], name: "index_organizations_on_aasm_state"
    t.index ["deleted_at"], name: "index_organizations_on_deleted_at"
    t.index ["favorite_users_count"], name: "index_organizations_on_favorite_users_count"
    t.index ["name"], name: "index_organizations_on_name", unique: true
    t.index ["staffs_count"], name: "index_organizations_on_staffs_count"
    t.index ["unpublished_at"], name: "index_organizations_on_unpublished_at"
  end

  create_table "people", force: :cascade do |t|
    t.bigint "prefecture_id"
    t.string "name", null: false
    t.string "name_kana", default: "", null: false
    t.string "nickname"
    t.string "gender"
    t.string "url"
    t.string "wikipedia_url"
    t.string "twitter_username"
    t.date "birthday"
    t.string "blood_type"
    t.integer "height"
    t.string "aasm_state", default: "published", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "name_en", default: "", null: false
    t.string "nickname_en", default: "", null: false
    t.string "url_en", default: "", null: false
    t.string "wikipedia_url_en", default: "", null: false
    t.string "twitter_username_en", default: "", null: false
    t.integer "favorite_users_count", default: 0, null: false
    t.integer "casts_count", default: 0, null: false
    t.integer "staffs_count", default: 0, null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["aasm_state"], name: "index_people_on_aasm_state"
    t.index ["casts_count"], name: "index_people_on_casts_count"
    t.index ["deleted_at"], name: "index_people_on_deleted_at"
    t.index ["favorite_users_count"], name: "index_people_on_favorite_users_count"
    t.index ["name"], name: "index_people_on_name", unique: true
    t.index ["prefecture_id"], name: "index_people_on_prefecture_id"
    t.index ["staffs_count"], name: "index_people_on_staffs_count"
    t.index ["unpublished_at"], name: "index_people_on_unpublished_at"
  end

  create_table "person_favorites", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "person_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "watched_works_count", default: 0, null: false
    t.index ["person_id"], name: "index_person_favorites_on_person_id"
    t.index ["user_id", "person_id"], name: "index_person_favorites_on_user_id_and_person_id", unique: true
    t.index ["user_id"], name: "index_person_favorites_on_user_id"
    t.index ["watched_works_count"], name: "index_person_favorites_on_watched_works_count"
  end

  create_table "prefectures", force: :cascade do |t|
    t.string "name", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["name"], name: "index_prefectures_on_name", unique: true
  end

  create_table "profiles", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.string "name", limit: 510, default: "", null: false
    t.string "description", limit: 510, default: "", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.boolean "background_image_animated", default: false, null: false
    t.string "tombo_avatar_file_name"
    t.string "tombo_avatar_content_type"
    t.integer "tombo_avatar_file_size"
    t.datetime "tombo_avatar_updated_at"
    t.string "tombo_background_image_file_name"
    t.string "tombo_background_image_content_type"
    t.integer "tombo_background_image_file_size"
    t.datetime "tombo_background_image_updated_at"
    t.string "url"
    t.text "image_data"
    t.text "background_image_data"
    t.index ["user_id"], name: "profiles_user_id_idx"
    t.index ["user_id"], name: "profiles_user_id_key", unique: true
  end

  create_table "programs", force: :cascade do |t|
    t.bigint "channel_id", null: false
    t.bigint "work_id", null: false
    t.string "url"
    t.datetime "started_at"
    t.string "aasm_state", default: "published", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "vod_title_code", default: "", null: false
    t.string "vod_title_name", default: "", null: false
    t.boolean "rebroadcast", default: false, null: false
    t.integer "minimum_episode_generatable_number", default: 1, null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["channel_id"], name: "index_programs_on_channel_id"
    t.index ["deleted_at"], name: "index_programs_on_deleted_at"
    t.index ["unpublished_at"], name: "index_programs_on_unpublished_at"
    t.index ["vod_title_code"], name: "index_programs_on_vod_title_code"
    t.index ["work_id"], name: "index_programs_on_work_id"
  end

  create_table "providers", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.string "name", limit: 510, null: false
    t.string "uid", limit: 510, null: false
    t.string "token", limit: 510, null: false
    t.integer "token_expires_at"
    t.string "token_secret", limit: 510
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["name", "uid"], name: "providers_name_uid_key", unique: true
    t.index ["user_id"], name: "providers_user_id_idx"
  end

  create_table "reactions", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "target_user_id", null: false
    t.string "kind", null: false
    t.bigint "collection_item_id"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["collection_item_id"], name: "index_reactions_on_collection_item_id"
    t.index ["target_user_id"], name: "index_reactions_on_target_user_id"
    t.index ["user_id"], name: "index_reactions_on_user_id"
  end

  create_table "receptions", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "channel_id", null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["channel_id"], name: "receptions_channel_id_idx"
    t.index ["user_id", "channel_id"], name: "receptions_user_id_channel_id_key", unique: true
    t.index ["user_id"], name: "receptions_user_id_idx"
  end

  create_table "records", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "work_id", null: false
    t.string "aasm_state", default: "published", null: false
    t.integer "impressions_count", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at"
    t.index ["deleted_at"], name: "index_records_on_deleted_at"
    t.index ["user_id"], name: "index_records_on_user_id"
    t.index ["work_id"], name: "index_records_on_work_id"
  end

  create_table "seasons", force: :cascade do |t|
    t.string "name", limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer "sort_number", null: false
    t.integer "year", null: false
    t.index ["sort_number"], name: "index_seasons_on_sort_number", unique: true
    t.index ["year", "name"], name: "index_seasons_on_year_and_name", unique: true
    t.index ["year"], name: "index_seasons_on_year"
  end

  create_table "series", force: :cascade do |t|
    t.string "name", null: false
    t.string "name_ro", default: "", null: false
    t.string "name_en", default: "", null: false
    t.string "aasm_state", default: "published", null: false
    t.integer "series_works_count", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.string "name_alter", default: "", null: false
    t.string "name_alter_en", default: "", null: false
    t.index ["deleted_at"], name: "index_series_on_deleted_at"
    t.index ["name"], name: "index_series_on_name", unique: true
    t.index ["unpublished_at"], name: "index_series_on_unpublished_at"
  end

  create_table "series_works", force: :cascade do |t|
    t.bigint "series_id", null: false
    t.bigint "work_id", null: false
    t.string "summary", default: "", null: false
    t.string "summary_en", default: "", null: false
    t.string "aasm_state", default: "published", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["deleted_at"], name: "index_series_works_on_deleted_at"
    t.index ["series_id", "work_id"], name: "index_series_works_on_series_id_and_work_id", unique: true
    t.index ["series_id"], name: "index_series_works_on_series_id"
    t.index ["unpublished_at"], name: "index_series_works_on_unpublished_at"
    t.index ["work_id"], name: "index_series_works_on_work_id"
  end

  create_table "sessions", force: :cascade do |t|
    t.string "session_id", null: false
    t.jsonb "data", default: "{}", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["session_id"], name: "index_sessions_on_session_id", unique: true
    t.index ["updated_at"], name: "index_sessions_on_updated_at"
  end

  create_table "settings", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.boolean "hide_record_body", default: true, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.boolean "share_record_to_twitter", default: false
    t.boolean "share_record_to_facebook", default: false
    t.string "slots_sort_type", default: "", null: false
    t.string "display_option_work_list", default: "list_detailed", null: false
    t.string "display_option_user_work_list", default: "grid_detailed", null: false
    t.string "records_sort_type", default: "created_at_desc", null: false
    t.string "display_option_record_list", default: "all_comments", null: false
    t.boolean "share_review_to_twitter", default: false, null: false
    t.boolean "share_review_to_facebook", default: false, null: false
    t.boolean "hide_supporter_badge", default: false, null: false
    t.boolean "share_status_to_twitter", default: false, null: false
    t.boolean "share_status_to_facebook", default: false, null: false
    t.boolean "privacy_policy_agreed", default: false, null: false
    t.string "timeline_mode", default: "following", null: false
    t.index ["user_id"], name: "index_settings_on_user_id"
  end

  create_table "slots", force: :cascade do |t|
    t.bigint "channel_id", null: false
    t.bigint "episode_id"
    t.bigint "work_id", null: false
    t.datetime "started_at", null: false
    t.datetime "sc_last_update"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.integer "sc_pid"
    t.boolean "rebroadcast", default: false, null: false
    t.string "aasm_state", default: "published", null: false
    t.bigint "program_id"
    t.integer "number"
    t.boolean "irregular", default: false, null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["aasm_state"], name: "index_slots_on_aasm_state"
    t.index ["channel_id"], name: "programs_channel_id_idx"
    t.index ["deleted_at"], name: "index_slots_on_deleted_at"
    t.index ["episode_id"], name: "programs_episode_id_idx"
    t.index ["program_id", "episode_id"], name: "index_slots_on_program_id_and_episode_id", unique: true
    t.index ["program_id", "number"], name: "index_slots_on_program_id_and_number", unique: true
    t.index ["program_id"], name: "index_slots_on_program_id"
    t.index ["sc_pid"], name: "index_slots_on_sc_pid", unique: true
    t.index ["unpublished_at"], name: "index_slots_on_unpublished_at"
    t.index ["work_id"], name: "programs_work_id_idx"
  end

  create_table "staffs", force: :cascade do |t|
    t.bigint "work_id", null: false
    t.string "name", null: false
    t.string "role", null: false
    t.string "role_other"
    t.string "aasm_state", default: "published", null: false
    t.integer "sort_number", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.bigint "resource_id", null: false
    t.string "resource_type", null: false
    t.string "name_en", default: "", null: false
    t.string "role_other_en", default: "", null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["aasm_state"], name: "index_staffs_on_aasm_state"
    t.index ["deleted_at"], name: "index_staffs_on_deleted_at"
    t.index ["resource_id", "resource_type"], name: "index_staffs_on_resource_id_and_resource_type"
    t.index ["sort_number"], name: "index_staffs_on_sort_number"
    t.index ["unpublished_at"], name: "index_staffs_on_unpublished_at"
    t.index ["work_id"], name: "index_staffs_on_work_id"
  end

  create_table "statuses", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "work_id", null: false
    t.integer "kind", null: false
    t.integer "likes_count", default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.bigint "oauth_application_id"
    t.index ["oauth_application_id"], name: "index_statuses_on_oauth_application_id"
    t.index ["user_id"], name: "statuses_user_id_idx"
    t.index ["work_id"], name: "statuses_work_id_idx"
  end

  create_table "syobocal_alerts", force: :cascade do |t|
    t.bigint "work_id"
    t.integer "kind", null: false
    t.integer "sc_prog_item_id"
    t.string "sc_sub_title", limit: 255
    t.string "sc_prog_comment", limit: 255
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["kind"], name: "index_syobocal_alerts_on_kind"
    t.index ["sc_prog_item_id"], name: "index_syobocal_alerts_on_sc_prog_item_id"
  end

  create_table "tips", force: :cascade do |t|
    t.integer "target", null: false
    t.string "slug", limit: 255, null: false
    t.string "title", limit: 255, null: false
    t.string "icon_name", limit: 255, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "locale", default: "other", null: false
    t.index ["locale"], name: "index_tips_on_locale"
    t.index ["slug", "locale"], name: "index_tips_on_slug_and_locale", unique: true
  end

  create_table "trailers", force: :cascade do |t|
    t.bigint "work_id", null: false
    t.string "url", null: false
    t.string "title", null: false
    t.string "thumbnail_file_name"
    t.string "thumbnail_content_type"
    t.integer "thumbnail_file_size"
    t.datetime "thumbnail_updated_at"
    t.integer "sort_number", default: 0, null: false
    t.string "aasm_state", default: "published", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "title_en", default: "", null: false
    t.text "image_data"
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["deleted_at"], name: "index_trailers_on_deleted_at"
    t.index ["unpublished_at"], name: "index_trailers_on_unpublished_at"
    t.index ["work_id"], name: "index_trailers_on_work_id"
  end

  create_table "twitter_bots", force: :cascade do |t|
    t.string "name", limit: 510, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.index ["name"], name: "twitter_bots_name_key", unique: true
  end

  create_table "twitter_watching_lists", force: :cascade do |t|
    t.string "username", null: false
    t.string "name", null: false
    t.string "since_id"
    t.string "discord_webhook_url", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  create_table "userland_categories", force: :cascade do |t|
    t.string "name", null: false
    t.string "name_en", null: false
    t.integer "sort_number", default: 0, null: false
    t.integer "userland_projects_count", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  create_table "userland_project_members", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "userland_project_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "userland_project_id"], name: "index_userland_pm_on_uid_and_userland_pid", unique: true
    t.index ["user_id"], name: "index_userland_project_members_on_user_id"
    t.index ["userland_project_id"], name: "index_userland_project_members_on_userland_project_id"
  end

  create_table "userland_projects", force: :cascade do |t|
    t.bigint "userland_category_id", null: false
    t.string "name", null: false
    t.string "summary", null: false
    t.text "description", null: false
    t.string "url", null: false
    t.string "icon_file_name"
    t.string "icon_content_type"
    t.integer "icon_file_size"
    t.datetime "icon_updated_at"
    t.boolean "available", default: false, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "locale", default: "other", null: false
    t.text "image_data"
    t.index ["locale"], name: "index_userland_projects_on_locale"
    t.index ["userland_category_id"], name: "index_userland_projects_on_userland_category_id"
  end

  create_table "users", force: :cascade do |t|
    t.citext "username", null: false
    t.citext "email", null: false
    t.integer "role", null: false
    t.string "encrypted_password", limit: 510, default: "", null: false
    t.datetime "remember_created_at"
    t.integer "sign_in_count", default: 0, null: false
    t.datetime "current_sign_in_at"
    t.datetime "last_sign_in_at"
    t.string "current_sign_in_ip", limit: 510
    t.string "last_sign_in_ip", limit: 510
    t.string "confirmation_token", limit: 510
    t.datetime "confirmed_at"
    t.datetime "confirmation_sent_at"
    t.string "unconfirmed_email", limit: 510
    t.integer "episode_records_count", default: 0, null: false
    t.integer "notifications_count", default: 0, null: false
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string "time_zone", null: false
    t.string "locale", null: false
    t.string "reset_password_token"
    t.datetime "reset_password_sent_at"
    t.datetime "record_cache_expired_at"
    t.datetime "status_cache_expired_at"
    t.datetime "work_tag_cache_expired_at"
    t.datetime "work_comment_cache_expired_at"
    t.bigint "gumroad_subscriber_id"
    t.string "allowed_locales", array: true
    t.integer "records_count", default: 0, null: false
    t.string "aasm_state", default: "published", null: false
    t.datetime "deleted_at"
    t.integer "character_favorites_count", default: 0, null: false
    t.integer "person_favorites_count", default: 0, null: false
    t.integer "organization_favorites_count", default: 0, null: false
    t.integer "plan_to_watch_works_count", default: 0, null: false
    t.integer "watching_works_count", default: 0, null: false
    t.integer "completed_works_count", default: 0, null: false
    t.integer "on_hold_works_count", default: 0, null: false
    t.integer "dropped_works_count", default: 0, null: false
    t.integer "following_count", default: 0, null: false
    t.integer "followers_count", default: 0, null: false
    t.index ["aasm_state"], name: "index_users_on_aasm_state"
    t.index ["allowed_locales"], name: "index_users_on_allowed_locales", using: :gin
    t.index ["confirmation_token"], name: "users_confirmation_token_key", unique: true
    t.index ["deleted_at"], name: "index_users_on_deleted_at"
    t.index ["email"], name: "users_email_key", unique: true
    t.index ["gumroad_subscriber_id"], name: "index_users_on_gumroad_subscriber_id"
    t.index ["username"], name: "users_username_key", unique: true
  end

  create_table "versions", force: :cascade do |t|
    t.string "item_type", limit: 510, null: false
    t.integer "item_id", null: false
    t.string "event", limit: 510, null: false
    t.string "whodunnit", limit: 510
    t.text "object"
    t.datetime "created_at"
  end

  create_table "vod_titles", force: :cascade do |t|
    t.bigint "channel_id", null: false
    t.bigint "work_id"
    t.string "code", null: false
    t.string "name", null: false
    t.string "aasm_state", default: "published", null: false
    t.datetime "mail_sent_at"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.datetime "deleted_at"
    t.datetime "unpublished_at"
    t.index ["channel_id"], name: "index_vod_titles_on_channel_id"
    t.index ["deleted_at"], name: "index_vod_titles_on_deleted_at"
    t.index ["mail_sent_at"], name: "index_vod_titles_on_mail_sent_at"
    t.index ["unpublished_at"], name: "index_vod_titles_on_unpublished_at"
    t.index ["work_id"], name: "index_vod_titles_on_work_id"
  end

  create_table "work_comments", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "work_id", null: false
    t.string "body", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "locale", default: "other", null: false
    t.index ["locale"], name: "index_work_comments_on_locale"
    t.index ["user_id", "work_id"], name: "index_work_comments_on_user_id_and_work_id", unique: true
    t.index ["user_id"], name: "index_work_comments_on_user_id"
    t.index ["work_id"], name: "index_work_comments_on_work_id"
  end

  create_table "work_images", force: :cascade do |t|
    t.bigint "work_id", null: false
    t.bigint "user_id", null: false
    t.string "attachment_file_name"
    t.integer "attachment_file_size"
    t.string "attachment_content_type"
    t.datetime "attachment_updated_at"
    t.string "copyright", default: "", null: false
    t.string "asin", default: "", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "color_rgb", default: "255,255,255", null: false
    t.text "image_data", null: false
    t.index ["user_id"], name: "index_work_images_on_user_id"
    t.index ["work_id"], name: "index_work_images_on_work_id"
  end

  create_table "work_records", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "work_id", null: false
    t.string "title", default: ""
    t.text "body", null: false
    t.string "rating_animation_state"
    t.string "rating_music_state"
    t.string "rating_story_state"
    t.string "rating_character_state"
    t.string "rating_overall_state"
    t.integer "likes_count", default: 0, null: false
    t.integer "impressions_count", default: 0, null: false
    t.string "aasm_state", default: "published", null: false
    t.datetime "modified_at"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.bigint "oauth_application_id"
    t.string "locale", default: "other", null: false
    t.bigint "record_id", null: false
    t.datetime "deleted_at"
    t.index ["deleted_at"], name: "index_work_records_on_deleted_at"
    t.index ["locale"], name: "index_work_records_on_locale"
    t.index ["oauth_application_id"], name: "index_work_records_on_oauth_application_id"
    t.index ["record_id"], name: "index_work_records_on_record_id", unique: true
    t.index ["user_id"], name: "index_work_records_on_user_id"
    t.index ["work_id"], name: "index_work_records_on_work_id"
  end

  create_table "work_taggables", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "work_tag_id", null: false
    t.string "description"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "locale", default: "other", null: false
    t.index ["locale"], name: "index_work_taggables_on_locale"
    t.index ["user_id", "work_tag_id"], name: "index_work_taggables_on_user_id_and_work_tag_id", unique: true
    t.index ["user_id"], name: "index_work_taggables_on_user_id"
    t.index ["work_tag_id"], name: "index_work_taggables_on_work_tag_id"
  end

  create_table "work_taggings", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "work_id", null: false
    t.bigint "work_tag_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "work_id", "work_tag_id"], name: "index_work_taggings_on_user_id_and_work_id_and_work_tag_id", unique: true
    t.index ["user_id"], name: "index_work_taggings_on_user_id"
    t.index ["work_id"], name: "index_work_taggings_on_work_id"
    t.index ["work_tag_id"], name: "index_work_taggings_on_work_tag_id"
  end

  create_table "work_tags", force: :cascade do |t|
    t.string "name", null: false
    t.string "aasm_state", default: "published", null: false
    t.integer "work_taggings_count", default: 0, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "locale", default: "other", null: false
    t.datetime "deleted_at"
    t.index ["deleted_at"], name: "index_work_tags_on_deleted_at"
    t.index ["locale"], name: "index_work_tags_on_locale"
    t.index ["name"], name: "index_work_tags_on_name", unique: true
    t.index ["work_taggings_count"], name: "index_work_tags_on_work_taggings_count"
  end

  create_table "works", force: :cascade do |t|
    t.bigint "season_id"
    t.integer "sc_tid"
    t.string "title", limit: 510, null: false
    t.integer "media", null: false
    t.string "official_site_url", limit: 510, default: "", null: false
    t.string "wikipedia_url", limit: 510, default: "", null: false
    t.integer "episodes_count", default: 0, null: false
    t.integer "watchers_count", default: 0, null: false
    t.date "released_at"
    t.datetime "created_at"
    t.datetime "updated_at"
    t.string "twitter_username", limit: 510
    t.string "twitter_hashtag", limit: 510
    t.string "released_at_about"
    t.string "aasm_state", default: "published", null: false
    t.bigint "number_format_id"
    t.string "title_kana", default: "", null: false
    t.string "title_ro", default: "", null: false
    t.string "title_en", default: "", null: false
    t.string "official_site_url_en", default: "", null: false
    t.string "wikipedia_url_en", default: "", null: false
    t.text "synopsis", default: "", null: false
    t.text "synopsis_en", default: "", null: false
    t.string "synopsis_source", default: "", null: false
    t.string "synopsis_source_en", default: "", null: false
    t.integer "mal_anime_id"
    t.string "facebook_og_image_url", default: "", null: false
    t.string "twitter_image_url", default: "", null: false
    t.string "recommended_image_url", default: "", null: false
    t.integer "season_year"
    t.integer "season_name"
    t.bigint "key_pv_id"
    t.integer "manual_episodes_count"
    t.boolean "no_episodes", default: false, null: false
    t.integer "work_records_count", default: 0, null: false
    t.date "started_on"
    t.date "ended_on"
    t.float "score"
    t.integer "ratings_count", default: 0, null: false
    t.float "satisfaction_rate"
    t.integer "records_count", default: 0, null: false
    t.integer "work_records_with_body_count", default: 0, null: false
    t.float "start_episode_raw_number", default: 1.0, null: false
    t.datetime "deleted_at"
    t.string "title_alter", default: "", null: false
    t.string "title_alter_en", default: "", null: false
    t.datetime "unpublished_at"
    t.index ["aasm_state"], name: "index_works_on_aasm_state"
    t.index ["deleted_at"], name: "index_works_on_deleted_at"
    t.index ["key_pv_id"], name: "index_works_on_key_pv_id"
    t.index ["number_format_id"], name: "index_works_on_number_format_id"
    t.index ["ratings_count"], name: "index_works_on_ratings_count"
    t.index ["satisfaction_rate", "ratings_count"], name: "index_works_on_satisfaction_rate_and_ratings_count"
    t.index ["satisfaction_rate"], name: "index_works_on_satisfaction_rate"
    t.index ["score"], name: "index_works_on_score"
    t.index ["season_id"], name: "works_season_id_idx"
    t.index ["season_year", "season_name"], name: "index_works_on_season_year_and_season_name"
    t.index ["season_year"], name: "index_works_on_season_year"
    t.index ["unpublished_at"], name: "index_works_on_unpublished_at"
  end

  add_foreign_key "activities", "activity_groups"
  add_foreign_key "activities", "episode_records"
  add_foreign_key "activities", "episodes"
  add_foreign_key "activities", "multiple_episode_records"
  add_foreign_key "activities", "statuses"
  add_foreign_key "activities", "users", name: "activities_user_id_fk", on_delete: :cascade
  add_foreign_key "activities", "work_records"
  add_foreign_key "activities", "works"
  add_foreign_key "activity_groups", "users"
  add_foreign_key "casts", "characters"
  add_foreign_key "casts", "people"
  add_foreign_key "casts", "works"
  add_foreign_key "channel_works", "channels", name: "channel_works_channel_id_fk", on_delete: :cascade
  add_foreign_key "channel_works", "users", name: "channel_works_user_id_fk", on_delete: :cascade
  add_foreign_key "channel_works", "works", name: "channel_works_work_id_fk", on_delete: :cascade
  add_foreign_key "channels", "channel_groups", name: "channels_channel_group_id_fk", on_delete: :cascade
  add_foreign_key "character_favorites", "characters"
  add_foreign_key "character_favorites", "users"
  add_foreign_key "character_images", "characters"
  add_foreign_key "character_images", "users"
  add_foreign_key "characters", "series"
  add_foreign_key "collection_items", "collections"
  add_foreign_key "collection_items", "users"
  add_foreign_key "collection_items", "works"
  add_foreign_key "collections", "users"
  add_foreign_key "comments", "episode_records", name: "comments_checkin_id_fk", on_delete: :cascade
  add_foreign_key "comments", "users", name: "comments_user_id_fk", on_delete: :cascade
  add_foreign_key "comments", "works"
  add_foreign_key "db_activities", "users"
  add_foreign_key "db_comments", "users"
  add_foreign_key "email_confirmations", "users"
  add_foreign_key "email_notifications", "users"
  add_foreign_key "episode_records", "episodes", name: "checkins_episode_id_fk", on_delete: :cascade
  add_foreign_key "episode_records", "multiple_episode_records"
  add_foreign_key "episode_records", "oauth_applications"
  add_foreign_key "episode_records", "records"
  add_foreign_key "episode_records", "users", name: "checkins_user_id_fk", on_delete: :cascade
  add_foreign_key "episode_records", "work_records", column: "review_id"
  add_foreign_key "episode_records", "works", name: "checkins_work_id_fk"
  add_foreign_key "episodes", "episodes", column: "prev_episode_id"
  add_foreign_key "episodes", "works", name: "episodes_work_id_fk", on_delete: :cascade
  add_foreign_key "faq_contents", "faq_categories"
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
  add_foreign_key "library_entries", "episodes", column: "next_episode_id"
  add_foreign_key "library_entries", "programs"
  add_foreign_key "library_entries", "slots", column: "next_slot_id"
  add_foreign_key "library_entries", "statuses"
  add_foreign_key "library_entries", "users"
  add_foreign_key "library_entries", "works"
  add_foreign_key "likes", "users", name: "likes_user_id_fk", on_delete: :cascade
  add_foreign_key "multiple_episode_records", "users"
  add_foreign_key "multiple_episode_records", "works"
  add_foreign_key "mute_users", "users"
  add_foreign_key "mute_users", "users", column: "muted_user_id"
  add_foreign_key "notifications", "users", column: "action_user_id", name: "notifications_action_user_id_fk", on_delete: :cascade
  add_foreign_key "notifications", "users", name: "notifications_user_id_fk", on_delete: :cascade
  add_foreign_key "oauth_access_grants", "oauth_applications", column: "application_id"
  add_foreign_key "oauth_access_grants", "users", column: "resource_owner_id"
  add_foreign_key "oauth_access_tokens", "oauth_applications", column: "application_id"
  add_foreign_key "oauth_access_tokens", "users", column: "resource_owner_id"
  add_foreign_key "organization_favorites", "organizations"
  add_foreign_key "organization_favorites", "users"
  add_foreign_key "people", "prefectures"
  add_foreign_key "person_favorites", "people"
  add_foreign_key "person_favorites", "users"
  add_foreign_key "profiles", "users", name: "profiles_user_id_fk", on_delete: :cascade
  add_foreign_key "programs", "channels"
  add_foreign_key "programs", "works"
  add_foreign_key "providers", "users", name: "providers_user_id_fk", on_delete: :cascade
  add_foreign_key "reactions", "collection_items"
  add_foreign_key "reactions", "users"
  add_foreign_key "reactions", "users", column: "target_user_id"
  add_foreign_key "receptions", "channels", name: "receptions_channel_id_fk", on_delete: :cascade
  add_foreign_key "receptions", "users", name: "receptions_user_id_fk", on_delete: :cascade
  add_foreign_key "records", "users"
  add_foreign_key "records", "works"
  add_foreign_key "series_works", "series"
  add_foreign_key "series_works", "works"
  add_foreign_key "settings", "users"
  add_foreign_key "slots", "channels", name: "programs_channel_id_fk", on_delete: :cascade
  add_foreign_key "slots", "episodes", name: "programs_episode_id_fk", on_delete: :cascade
  add_foreign_key "slots", "programs"
  add_foreign_key "slots", "works", name: "programs_work_id_fk", on_delete: :cascade
  add_foreign_key "staffs", "works"
  add_foreign_key "statuses", "oauth_applications"
  add_foreign_key "statuses", "users", name: "statuses_user_id_fk", on_delete: :cascade
  add_foreign_key "statuses", "works", name: "statuses_work_id_fk", on_delete: :cascade
  add_foreign_key "syobocal_alerts", "works", name: "syobocal_alerts_work_id_fk", on_delete: :cascade
  add_foreign_key "trailers", "works"
  add_foreign_key "userland_project_members", "userland_projects"
  add_foreign_key "userland_project_members", "users"
  add_foreign_key "userland_projects", "userland_categories"
  add_foreign_key "users", "gumroad_subscribers"
  add_foreign_key "vod_titles", "channels"
  add_foreign_key "vod_titles", "works"
  add_foreign_key "work_comments", "users"
  add_foreign_key "work_comments", "works"
  add_foreign_key "work_images", "users"
  add_foreign_key "work_images", "works"
  add_foreign_key "work_records", "oauth_applications"
  add_foreign_key "work_records", "records"
  add_foreign_key "work_records", "users"
  add_foreign_key "work_records", "works"
  add_foreign_key "work_taggables", "users"
  add_foreign_key "work_taggables", "work_tags"
  add_foreign_key "work_taggings", "users"
  add_foreign_key "work_taggings", "work_tags"
  add_foreign_key "work_taggings", "works"
  add_foreign_key "works", "number_formats"
  add_foreign_key "works", "seasons", name: "works_season_id_fk", on_delete: :cascade
  add_foreign_key "works", "trailers", column: "key_pv_id"
end
