class AddFetchSyobocalToDraftEpisodes < ActiveRecord::Migration[4.2]
  def change
    add_column :draft_episodes, :fetch_syobocal, :boolean, null: false, default: false
  end
end
