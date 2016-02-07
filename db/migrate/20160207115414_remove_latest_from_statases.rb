class RemoveLatestFromStatases < ActiveRecord::Migration
  def change
    remove_column :statuses, :latest
  end
end
