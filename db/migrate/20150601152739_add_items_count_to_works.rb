class AddItemsCountToWorks < ActiveRecord::Migration
  def change
    add_column :works, :items_count, :integer, null: false, default: 0
    add_index :works, :items_count
  end
end
