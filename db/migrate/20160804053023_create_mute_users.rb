# frozen_string_literal: true

class CreateMuteUsers < ActiveRecord::Migration[5.0]
  def change
    create_table :mute_users do |t|
      t.integer :user_id, null: false
      t.integer :muted_user_id, null: false
      t.timestamps
    end

    add_index :mute_users, :user_id
    add_foreign_key :mute_users, :users
    add_index :mute_users, :muted_user_id
    add_foreign_key :mute_users, :users, column: :muted_user_id

    add_index :mute_users, [:user_id, :muted_user_id], unique: true
  end
end
