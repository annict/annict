class AddOnairToWorks < ActiveRecord::Migration
  def change
    add_column :works, :on_air, :boolean, null: false, default: false
    add_index :works, :on_air
  end
end
