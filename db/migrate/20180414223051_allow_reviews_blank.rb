# frozen_string_literal: true

class AllowReviewsBlank < ActiveRecord::Migration[5.1]
  def change
    add_column :records, :impressions_count, :integer, null: false, default: 0
    add_index :records, :impressions_count

    add_column :works, :review_comments_count, :integer, null: false, default: 0
  end
end
