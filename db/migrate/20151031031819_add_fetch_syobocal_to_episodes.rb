class AddFetchSyobocalToEpisodes < ActiveRecord::Migration[4.2]
  def change
    add_column :episodes, :fetch_syobocal, :boolean, null: false, default: false
  end
end
