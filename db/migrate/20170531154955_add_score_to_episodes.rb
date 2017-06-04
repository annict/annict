# frozen_string_literal: true

class AddScoreToEpisodes < ActiveRecord::Migration[5.1]
  def change
    add_column :episodes, :score, :float, null: false, default: 50
    add_index :episodes, :score
  end
end
