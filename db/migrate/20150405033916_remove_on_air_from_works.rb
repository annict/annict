class RemoveOnAirFromWorks < ActiveRecord::Migration
  def change
    remove_column :works, :on_air
  end
end
