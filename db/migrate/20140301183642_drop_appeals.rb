class DropAppeals < ActiveRecord::Migration[4.2]
  def change
    drop_table :appeals
  end
end
