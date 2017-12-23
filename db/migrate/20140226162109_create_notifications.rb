class CreateNotifications < ActiveRecord::Migration[4.2]
  def change
    create_table :notifications do |t|
      t.integer :user_id,        null: false
      t.integer :action_user_id, null: false
      t.integer :trackable_id,   null: false
      t.string  :trackable_type, null: false
      t.string  :action,         null: false
      t.boolean :read,           null: false, default: false
      t.timestamps
    end

    add_index :notifications, [:trackable_id, :trackable_type]
    add_index :notifications, :read

    add_foreign_key :notifications, :users
    add_foreign_key :notifications, :users, column: :action_user_id

  end
end
