class RemoveItemsWorkIdIdx < ActiveRecord::Migration
  def change
    remove_index :items, name: "items_work_id_idx"
  end
end
