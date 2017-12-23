class AddWorkIdToComments < ActiveRecord::Migration[4.2]
  def change
    add_column :comments, :work_id, :integer
    add_index :comments, :work_id
    add_foreign_key :comments, :works
  end
end
