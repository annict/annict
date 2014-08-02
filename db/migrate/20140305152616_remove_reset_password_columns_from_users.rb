class RemoveResetPasswordColumnsFromUsers < ActiveRecord::Migration
  def change
    remove_column :users, :reset_password_token
    remove_column :users, :reset_password_sent_at
  end
end
