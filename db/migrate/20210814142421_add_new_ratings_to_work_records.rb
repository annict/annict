# frozen_string_literal: true

class AddNewRatingsToWorkRecords < ActiveRecord::Migration[6.1]
  def change
    add_column :work_records, :animation_rating, :integer
    add_column :work_records, :character_rating, :integer
    add_column :work_records, :music_rating, :integer
    add_column :work_records, :story_rating, :integer
    add_column :work_records, :migrated_at, :datetime

    add_index :work_records, :migrated_at
  end
end
