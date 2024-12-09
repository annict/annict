# frozen_string_literal: true

class AddEpisodeAndWorkRecordColumnsToRecords < ActiveRecord::Migration[7.0]
  def change
    add_column :records, :trackable_id, :bigint
    add_column :records, :trackable_type, :string
    add_column :records, :oauth_application_id, :bigint
    add_column :records, :body, :text, default: "", null: false
    add_column :records, :comments_count, :integer, default: 0, null: false
    add_column :records, :likes_count, :integer, default: 0, null: false
    add_column :records, :locale, :string
    add_column :records, :overall_rating, :string
    add_column :records, :animation_rating, :string
    add_column :records, :character_rating, :string
    add_column :records, :music_rating, :string
    add_column :records, :story_rating, :string
    add_column :records, :advanced_overall_rating, :float
    add_column :records, :modified_at, :datetime

    add_index :records, %i[trackable_id trackable_type]
    add_index :records, :oauth_application_id

    add_foreign_key :records, :oauth_applications
  end
end
