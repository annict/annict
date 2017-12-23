class DropTableShots < ActiveRecord::Migration[4.2]
  def change
    drop_table :shots
  end
end
