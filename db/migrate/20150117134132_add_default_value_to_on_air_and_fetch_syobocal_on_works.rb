class AddDefaultValueToOnAirAndFetchSyobocalOnWorks < ActiveRecord::Migration
  def change
    change_column_default :works, :on_air, false
    change_column_default :works, :fetch_syobocal, false
  end
end
