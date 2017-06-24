# frozen_string_literal: true

class AddReviewsCountToWorks < ActiveRecord::Migration[5.1]
  def change
    add_column :works, :reviews_count, :integer, null: false, default: 0
  end
end
