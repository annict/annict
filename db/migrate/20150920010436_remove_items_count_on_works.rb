class RemoveItemsCountOnWorks < ActiveRecord::Migration[4.2]
  def change
    remove_column :works, :items_count
  end
end
