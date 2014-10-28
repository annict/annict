class AddRoleToUsers < ActiveRecord::Migration
  def change
    add_column :users, :role, :integer, null: false, after: :email
    add_index  :users, :role
  end
end
