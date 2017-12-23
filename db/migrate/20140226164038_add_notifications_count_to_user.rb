class AddNotificationsCountToUser < ActiveRecord::Migration[4.2]
  def change
    add_column :users, :notifications_count, :integer, null: false, default: 0, after: :checkins_count
  end
end
