# frozen_string_literal: true

class AddEpisodeAndWorkRecordColumnsToRecords < ActiveRecord::Migration[6.1]
  def change
    add_column :records, :watchable_id, :bigint
    add_column :records, :watchable_type, :string
    add_column :records, :oauth_application_id, :bigint
    add_column :records, :body, :text, default: "", null: false
    add_column :records, :comments_count, :integer, default: 0, null: false
    add_column :records, :likes_count, :integer, default: 0, null: false
    add_column :records, :locale, :integer, default: 0, null: false
    add_column :records, :rating, :integer
    add_column :records, :advanced_rating, :float
    add_column :records, :animation_rating, :integer
    add_column :records, :character_rating, :integer
    add_column :records, :music_rating, :integer
    add_column :records, :story_rating, :integer
    add_column :records, :twitter_url_hash, :string
    add_column :records, :facebook_url_hash, :string
    add_column :records, :watched_at, :datetime
    add_column :records, :modified_at, :datetime

    add_index :records, %i[watchable_id watchable_type]
    add_index :records, :oauth_application_id

    add_foreign_key :records, :oauth_applications
  end
end
