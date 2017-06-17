# frozen_string_literal: true

class AddManualEpisodesCountToWorks < ActiveRecord::Migration[5.1]
  def change
    add_column :works, :manual_episodes_count, :integer
    rename_column :works, :episodes_count, :auto_episodes_count
  end
end
