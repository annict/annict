class DropTableShots < ActiveRecord::Migration
  def change
    drop_table :shots
  end
end
