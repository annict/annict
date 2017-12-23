class AddPrevEpisodeIdToEpisodes < ActiveRecord::Migration[4.2]
  def change
    add_column :episodes, :prev_episode_id, :integer
    add_index :episodes, :prev_episode_id
    add_foreign_key :episodes, :episodes, column: :prev_episode_id

    add_column :draft_episodes, :prev_episode_id, :integer
    add_index :draft_episodes, :prev_episode_id
    add_foreign_key :draft_episodes, :episodes, column: :prev_episode_id
  end
end
