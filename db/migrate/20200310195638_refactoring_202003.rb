# frozen_string_literal: true

class Refactoring202003 < ActiveRecord::Migration[6.0]
  def change
    %i[
      casts
      channel_groups
      channels
      characters
      episodes
      organizations
      people
      programs
      series
      series_works
      slots
      staffs
      trailers
      vod_titles
      works
    ].each do |table_name|
      add_column table_name, :unpublished_at, :datetime
      add_index table_name, :unpublished_at
    end

    %i[
      activities
      casts
      channel_groups
      channel_works
      channels
      character_images
      characters
      comments
      db_activities
      db_comments
      delayed_jobs
      email_notifications
      episode_records
      episodes
      favorite_characters
      favorite_organizations
      favorite_people
      finished_tips
      follows
      forum_categories
      forum_comments
      forum_post_participants
      forum_posts
      library_entries
      likes
      multiple_episode_records
      mute_users
      notifications
      number_formats
      oauth_access_grants
      oauth_access_tokens
      oauth_applications
      organizations
      people
      prefectures
      profiles
      providers
      receptions
      seasons
      series
      series_works
      settings
      slots
      staffs
      statuses
      syobocal_alerts
      tips
      twitter_bots
      userland_categories
      userland_project_members
      userland_projects
      users
      versions
      work_images
      works
    ].each do |table_name|
      change_column table_name, :id, :bigint
    end

    [
      %i[activities user_id],
      %i[activities recipient_id],
      %i[activities trackable_id],
      %i[activities work_id],
      %i[activities episode_id],
      %i[activities status_id],
      %i[activities episode_record_id],
      %i[activities multiple_episode_record_id],
      %i[activities work_record_id],
      %i[casts person_id],
      %i[casts work_id],
      %i[casts character_id],
      %i[channel_works user_id],
      %i[channel_works work_id],
      %i[channel_works channel_id],
      %i[channels channel_group_id],
      %i[character_images character_id],
      %i[character_images user_id],
      %i[characters series_id],
      %i[collection_items user_id],
      %i[collection_items collection_id],
      %i[collection_items work_id],
      %i[collections user_id],
      %i[comments user_id],
      %i[comments episode_record_id],
      %i[comments work_id],
      %i[db_activities user_id],
      %i[db_activities trackable_id],
      %i[db_activities root_resource_id],
      %i[db_activities object_id],
      %i[db_comments user_id],
      %i[db_comments resource_id],
      %i[email_notifications user_id],
      %i[episode_items episode_id],
      %i[episode_items item_id],
      %i[episode_items user_id],
      %i[episode_items work_id],
      %i[episode_records user_id],
      %i[episode_records episode_id],
      %i[episode_records work_id],
      %i[episode_records multiple_episode_record_id],
      %i[episode_records oauth_application_id],
      %i[episode_records review_id],
      %i[episode_records record_id],
      %i[episodes work_id],
      %i[episodes prev_episode_id],
      %i[faq_contents faq_category_id],
      %i[favorite_characters user_id],
      %i[favorite_characters character_id],
      %i[favorite_organizations user_id],
      %i[favorite_organizations organization_id],
      %i[favorite_people user_id],
      %i[favorite_people person_id],
      %i[finished_tips user_id],
      %i[finished_tips tip_id],
      %i[follows user_id],
      %i[follows following_id],
      %i[forum_comments user_id],
      %i[forum_comments forum_post_id],
      %i[forum_post_participants forum_post_id],
      %i[forum_post_participants user_id],
      %i[forum_posts user_id],
      %i[forum_posts forum_category_id],
      %i[impressions impressionable_id],
      %i[impressions user_id],
      %i[library_entries user_id],
      %i[library_entries work_id],
      %i[library_entries next_episode_id],
      %i[library_entries status_id],
      %i[likes user_id],
      %i[likes recipient_id],
      %i[multiple_episode_records user_id],
      %i[multiple_episode_records work_id],
      %i[mute_users user_id],
      %i[mute_users muted_user_id],
      %i[notifications user_id],
      %i[notifications action_user_id],
      %i[notifications trackable_id],
      %i[oauth_access_grants resource_owner_id],
      %i[oauth_access_grants application_id],
      %i[oauth_access_tokens resource_owner_id],
      %i[oauth_access_tokens application_id],
      %i[oauth_applications owner_id],
      %i[people prefecture_id],
      %i[profiles user_id],
      %i[programs channel_id],
      %i[programs work_id],
      %i[providers user_id],
      %i[reactions user_id],
      %i[reactions target_user_id],
      %i[reactions collection_item_id],
      %i[receptions user_id],
      %i[receptions channel_id],
      %i[records user_id],
      %i[records work_id],
      %i[series_works series_id],
      %i[series_works work_id],
      %i[settings user_id],
      %i[slots channel_id],
      %i[slots episode_id],
      %i[slots work_id],
      %i[slots program_id],
      %i[staffs work_id],
      %i[staffs resource_id],
      %i[statuses user_id],
      %i[statuses work_id],
      %i[statuses oauth_application_id],
      %i[syobocal_alerts work_id],
      %i[trailers work_id],
      %i[userland_project_members user_id],
      %i[userland_project_members userland_project_id],
      %i[userland_projects userland_category_id],
      %i[users gumroad_subscriber_id],
      %i[vod_titles channel_id],
      %i[vod_titles work_id],
      %i[work_comments user_id],
      %i[work_comments work_id],
      %i[work_images work_id],
      %i[work_images user_id],
      %i[work_items work_id],
      %i[work_items item_id],
      %i[work_items user_id],
      %i[work_records user_id],
      %i[work_records work_id],
      %i[work_records oauth_application_id],
      %i[work_records record_id],
      %i[work_taggables user_id],
      %i[work_taggables work_tag_id],
      %i[work_taggings user_id],
      %i[work_taggings work_id],
      %i[work_taggings work_tag_id],
      %i[works season_id],
      %i[works number_format_id],
      %i[works key_pv_id]
    ].each do |(table_name, field_name)|
      change_column table_name, field_name, :bigint
    end

    [
      %i[library_entries watched_episode_ids]
    ].each do |(table_name, field_name)|
      change_column table_name, field_name, :bigint, array: true
    end
  end
end
