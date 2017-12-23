class DropStaffs < ActiveRecord::Migration[4.2]
  def change
    drop_table :staffs
  end
end
