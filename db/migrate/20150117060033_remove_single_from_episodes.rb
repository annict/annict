class RemoveSingleFromEpisodes < ActiveRecord::Migration
  def change
    remove_column :episodes, :single
  end
end
