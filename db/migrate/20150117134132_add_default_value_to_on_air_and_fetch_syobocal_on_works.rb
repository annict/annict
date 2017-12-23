class AddDefaultValueToOnAirAndFetchSyobocalOnWorks < ActiveRecord::Migration[4.2]
  def change
    change_column_default :works, :on_air, false
    change_column_default :works, :fetch_syobocal, false
  end
end
