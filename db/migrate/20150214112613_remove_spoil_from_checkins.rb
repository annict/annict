class RemoveSpoilFromCheckins < ActiveRecord::Migration[4.2]
  def change
    remove_column :checkins, :spoil
  end
end
