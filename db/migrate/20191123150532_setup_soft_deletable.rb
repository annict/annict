# frozen_string_literal: true

class SetupSoftDeletable < ActiveRecord::Migration[6.0]
  def change
    add_column :casts, :deleted_at, :datetime
    add_column :channels, :deleted_at, :datetime
    add_column :characters, :deleted_at, :datetime
    add_column :collection_items, :deleted_at, :datetime
    add_column :collections, :deleted_at, :datetime
    add_column :episode_records, :deleted_at, :datetime
    add_column :episodes, :deleted_at, :datetime
    add_column :faq_categories, :deleted_at, :datetime
    add_column :faq_contents, :deleted_at, :datetime
    add_column :oauth_applications, :deleted_at, :datetime
    add_column :organizations, :deleted_at, :datetime
    add_column :people, :deleted_at, :datetime
    add_column :programs, :deleted_at, :datetime
    add_column :records, :deleted_at, :datetime
    add_column :series, :deleted_at, :datetime
    add_column :series_works, :deleted_at, :datetime
    add_column :slots, :deleted_at, :datetime
    add_column :staffs, :deleted_at, :datetime
    add_column :trailers, :deleted_at, :datetime
    add_column :users, :deleted_at, :datetime
    add_column :vod_titles, :deleted_at, :datetime
    add_column :work_records, :deleted_at, :datetime
    add_column :work_tags, :deleted_at, :datetime
    add_column :works, :deleted_at, :datetime

    add_index :casts, :deleted_at
    add_index :channels, :deleted_at
    add_index :characters, :deleted_at
    add_index :collection_items, :deleted_at
    add_index :collections, :deleted_at
    add_index :episode_records, :deleted_at
    add_index :episodes, :deleted_at
    add_index :faq_categories, :deleted_at
    add_index :faq_contents, :deleted_at
    add_index :oauth_applications, :deleted_at
    add_index :organizations, :deleted_at
    add_index :people, :deleted_at
    add_index :programs, :deleted_at
    add_index :records, :deleted_at
    add_index :series, :deleted_at
    add_index :series_works, :deleted_at
    add_index :slots, :deleted_at
    add_index :staffs, :deleted_at
    add_index :trailers, :deleted_at
    add_index :users, :deleted_at
    add_index :vod_titles, :deleted_at
    add_index :work_records, :deleted_at
    add_index :work_tags, :deleted_at
    add_index :works, :deleted_at
  end
end
