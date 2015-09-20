class AddUniqueIndexToWorkIdOnItems < ActiveRecord::Migration
  def change
    add_index :items, :work_id, unique: true
  end
end
