class AddNextEpisodeIdToEpisodes < ActiveRecord::Migration[4.2]
  def change
    add_column :episodes, :next_episode_id, :integer
    add_index :episodes, :next_episode_id
    add_foreign_key :episodes, :episodes, column: :next_episode_id
  end
end
