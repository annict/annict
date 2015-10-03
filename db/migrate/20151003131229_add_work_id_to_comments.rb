class AddWorkIdToComments < ActiveRecord::Migration
  def change
    add_column :comments, :work_id, :integer
    add_index :comments, :work_id
    add_foreign_key :comments, :works
  end
end
