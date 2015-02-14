class RemoveSpoilFromCheckins < ActiveRecord::Migration
  def change
    remove_column :checkins, :spoil
  end
end
