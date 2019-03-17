# frozen_string_literal: true

class UpdateForIrregularEpisodes < ActiveRecord::Migration[5.2]
  def change
    add_column :programs, :irregular, :boolean, default: false, null: false
  end
end
