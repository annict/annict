class AddShareCheckinToUsers < ActiveRecord::Migration
  def change
    add_column :users, :share_checkin, :boolean, default: false, after: :unconfirmed_email
  end
end
