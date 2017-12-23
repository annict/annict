class RemoveShareCheckinOnUsers < ActiveRecord::Migration[4.2]
  def change
    remove_column :users, :share_checkin
  end
end
