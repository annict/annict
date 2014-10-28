class CreateNotifications < ActiveRecord::Migration
  def change
    create_table :notifications do |t|
      t.integer :user_id,        null: false
      t.integer :action_user_id, null: false
      t.integer :trackable_id,   null: false
      t.string  :trackable_type, null: false
      t.string  :action,         null: false
      t.boolean :read,           null: false, default: false
      t.timestamps

      t.foreign_key :users, dependent: :delete
      t.foreign_key :users, column: 'action_user_id', dependent: :delete
    end

    add_index :notifications, [:trackable_id, :trackable_type]
    add_index :notifications, :read
  end
end
