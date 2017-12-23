class RemoveNextEpisodeIdOnEpisodes < ActiveRecord::Migration[4.2]
  def change
    remove_column :episodes, :next_episode_id
  end
end
