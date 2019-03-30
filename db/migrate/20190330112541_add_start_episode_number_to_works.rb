# frozen_string_literal: true

class AddStartEpisodeNumberToWorks < ActiveRecord::Migration[5.2]
  def change
    add_column :works, :start_episode_raw_number, :float, null: false, default: 1.0
  end
end
