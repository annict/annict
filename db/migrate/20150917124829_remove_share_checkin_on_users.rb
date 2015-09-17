class RemoveShareCheckinOnUsers < ActiveRecord::Migration
  def change
    remove_column :users, :share_checkin
  end
end
