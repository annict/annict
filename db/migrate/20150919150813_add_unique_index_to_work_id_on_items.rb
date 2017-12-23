class AddUniqueIndexToWorkIdOnItems < ActiveRecord::Migration[4.2]
  def change
    add_index :items, :work_id, unique: true
  end
end
