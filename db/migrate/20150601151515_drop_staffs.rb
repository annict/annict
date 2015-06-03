class DropStaffs < ActiveRecord::Migration
  def change
    drop_table :staffs
  end
end
