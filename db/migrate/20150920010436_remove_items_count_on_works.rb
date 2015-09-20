class RemoveItemsCountOnWorks < ActiveRecord::Migration
  def change
    remove_column :works, :items_count
  end
end
