class AddRoleToUsers < ActiveRecord::Migration[4.2]
  def change
    add_column :users, :role, :integer, null: false, after: :email
    add_index  :users, :role
  end
end
