# frozen_string_literal: true

class UpdateForIrregularEpisodes < ActiveRecord::Migration[5.2]
  def change
    add_column :programs, :irregular, :boolean, default: false, null: false
    add_column :program_details, :minimum_episode_generatable_number, :integer, default: 1, null: false
  end
end
