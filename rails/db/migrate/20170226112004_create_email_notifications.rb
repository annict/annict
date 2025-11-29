# frozen_string_literal: true

class CreateEmailNotifications < ActiveRecord::Migration[5.0]
  def change
    create_table :email_notifications do |t|
      t.integer :user_id, null: false
      t.string :unsubscription_key, null: false
      t.boolean :event_followed_user, null: false, default: true
      t.boolean :event_liked_record, null: false, default: true
      t.boolean :event_friends_joined, null: false, default: true
      t.boolean :event_next_season_came, null: false, default: true
      t.timestamps null: false
    end

    add_index :email_notifications, :user_id, unique: true
    add_index :email_notifications, :unsubscription_key, unique: true

    add_foreign_key :email_notifications, :users
  end
end
