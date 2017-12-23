class RemoveSingleFromEpisodes < ActiveRecord::Migration[4.2]
  def change
    remove_column :episodes, :single
  end
end
