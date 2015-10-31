class AddFetchSyobocalToDraftEpisodes < ActiveRecord::Migration
  def change
    add_column :draft_episodes, :fetch_syobocal, :boolean, null: false, default: false
  end
end
