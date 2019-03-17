# frozen_string_literal: true

class UpdateForIrregularEpisodes < ActiveRecord::Migration[5.2]
  def change
    add_column :works, :irregular_episodes_count, :integer, default: 0, null: false
  end
end
