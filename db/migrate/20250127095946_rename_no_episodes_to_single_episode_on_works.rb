# typed: false
# frozen_string_literal: true

class RenameNoEpisodesToSingleEpisodeOnWorks < ActiveRecord::Migration[7.0]
  def change
    rename_column :works, :no_episodes, :single_episode
  end
end
