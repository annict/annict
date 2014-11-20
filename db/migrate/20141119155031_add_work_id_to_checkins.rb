class AddWorkIdToCheckins < ActiveRecord::Migration
  def change
    add_column :checkins, :work_id, :integer
    add_index :checkins, :work_id
    add_foreign_key :checkins, :works
  end
end
