class RemoveFetchSyobocalFromWorks < ActiveRecord::Migration[4.2]
  def change
    remove_column :works, :fetch_syobocal
  end
end
