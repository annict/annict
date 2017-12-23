class RemoveLatestFromStatases < ActiveRecord::Migration[4.2]
  def change
    remove_column :statuses, :latest
  end
end
