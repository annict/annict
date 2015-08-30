class RemoveNextEpisodeIdOnEpisodes < ActiveRecord::Migration
  def change
    remove_column :episodes, :next_episode_id
  end
end
