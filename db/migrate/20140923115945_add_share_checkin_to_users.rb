class AddShareCheckinToUsers < ActiveRecord::Migration[4.2]
  def change
    add_column :users, :share_checkin, :boolean, default: false, after: :unconfirmed_email
  end
end
