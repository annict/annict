class SetDefaultValueToSpoilOnCheckins < ActiveRecord::Migration
  def change
    change_column_default :checkins, :spoil, :false
  end
end
