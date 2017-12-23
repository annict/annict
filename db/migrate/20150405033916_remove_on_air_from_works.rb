class RemoveOnAirFromWorks < ActiveRecord::Migration[4.2]
  def change
    remove_column :works, :on_air
  end
end
