class AddFetchSyobocalToWorks < ActiveRecord::Migration[4.2]
  def change
    add_column :works, :fetch_syobocal, :boolean, null: false, default: false, after: :on_air
  end
end
