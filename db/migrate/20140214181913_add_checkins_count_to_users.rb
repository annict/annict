class AddCheckinsCountToUsers < ActiveRecord::Migration
  def change
    add_column :users, :checkins_count, :integer, null: false, default: 0, after: :unconfirmed_email
  end
end
