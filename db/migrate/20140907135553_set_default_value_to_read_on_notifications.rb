class SetDefaultValueToReadOnNotifications < ActiveRecord::Migration[4.2]
  def change
    change_column_default :notifications, :read, :false
  end
end
