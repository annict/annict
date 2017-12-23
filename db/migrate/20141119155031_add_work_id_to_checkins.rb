class AddWorkIdToCheckins < ActiveRecord::Migration[4.2]
  def change
    add_column :checkins, :work_id, :integer
    add_index :checkins, :work_id
    add_foreign_key :checkins, :works
  end
end
