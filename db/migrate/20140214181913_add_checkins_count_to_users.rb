class AddCheckinsCountToUsers < ActiveRecord::Migration[4.2]
  def change
    add_column :users, :checkins_count, :integer, null: false, default: 0, after: :unconfirmed_email
  end
end
