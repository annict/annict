class AddFetchSyobocalToEpisodes < ActiveRecord::Migration
  def change
    add_column :episodes, :fetch_syobocal, :boolean, null: false, default: false
  end
end
