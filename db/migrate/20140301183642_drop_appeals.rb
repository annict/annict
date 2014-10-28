class DropAppeals < ActiveRecord::Migration
  def change
    drop_table :appeals
  end
end
