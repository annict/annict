class AddFetchSyobocalToWorks < ActiveRecord::Migration
  def change
    add_column :works, :fetch_syobocal, :boolean, null: false, default: false, after: :on_air
  end
end
