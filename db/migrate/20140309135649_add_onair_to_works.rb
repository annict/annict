class AddOnairToWorks < ActiveRecord::Migration[4.2]
  def change
    add_column :works, :on_air, :boolean, null: false, default: false
    add_index :works, :on_air
  end
end
