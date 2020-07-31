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
  end
end
