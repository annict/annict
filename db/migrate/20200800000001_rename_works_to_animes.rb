# frozen_string_literal: true

class RenameWorksToAnimes < ActiveRecord::Migration[6.0]
  def change
    rename_table :works, :animes
    rename_table :channel_works, :channel_animes
    rename_table :series_works, :series_animes
    rename_table :work_comments, :anime_comments
    rename_table :work_images, :anime_images
    rename_table :work_records, :anime_records
    rename_table :work_taggables, :anime_taggables
    rename_table :work_taggings, :anime_taggings
    rename_table :work_tags, :anime_tags

    %i(
      anime_comments
      anime_images
      anime_records
      anime_taggings
      casts
      channel_animes
      collection_items
      comments
      episodes
      episode_records
      library_entries
      multiple_episode_records
      programs
      records
      series_animes
      slots
      staffs
      statuses
      trailers
      vod_titles
    ).each do |table_name|
      rename_column table_name, :work_id, :anime_id
    end
  end
end
