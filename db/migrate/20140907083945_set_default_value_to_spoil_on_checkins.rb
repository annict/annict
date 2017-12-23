class SetDefaultValueToSpoilOnCheckins < ActiveRecord::Migration[4.2]
  def change
    change_column_default :checkins, :spoil, :false
  end
end
