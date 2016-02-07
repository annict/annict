class RemoveFetchSyobocalFromWorks < ActiveRecord::Migration
  def change
    remove_column :works, :fetch_syobocal
  end
end
