# frozen_string_literal: true

class AddRatingsCountToEpisodes < ActiveRecord::Migration[5.1]
  def change
    add_column :episodes, :ratings_count, :integer, null: false, default: 0
    add_index :episodes, :ratings_count

    add_column :episodes, :satisfaction_rate, :float
    add_index :episodes, :satisfaction_rate
    add_index :episodes, %i[satisfaction_rate ratings_count]
  end
end
