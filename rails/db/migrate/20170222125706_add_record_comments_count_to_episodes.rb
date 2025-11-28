# frozen_string_literal: true

class AddRecordCommentsCountToEpisodes < ActiveRecord::Migration[5.0]
  def change
    add_column :episodes, :record_comments_count, :integer, null: false, default: 0
  end
end
