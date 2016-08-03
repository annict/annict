# frozen_string_literal: true

class AddAvgRatingToEpisodes < ActiveRecord::Migration[5.0]
  def change
    add_column :episodes, :avg_rating, :float
  end
end
