# frozen_string_literal: true

class MergeReviewsWithRecords < ActiveRecord::Migration[5.1]
  def change
    add_column :reviews, :record_id, :integer
    add_index :reviews, :record_id
    add_foreign_key :reviews, :records

    add_column :records, :impressions_count, :integer, null: false, default: 0
    add_index :records, :impressions_count

    change_column_null :records, :episode_id, true

    add_column :works, :record_comments_count, :integer, null: false, default: 0
  end
end
