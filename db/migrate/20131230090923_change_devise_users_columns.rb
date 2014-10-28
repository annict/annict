class ChangeDeviseUsersColumns < ActiveRecord::Migration
  def change
    change_column :users, :encrypted_password, :string, default: ''
    add_column :users, :unconfirmed_email, :string, after: :confirmation_sent_at
  end
end
