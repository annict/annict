# frozen_string_literal: true

class CreateTwitterUsers < ActiveRecord::Migration[5.1]
  def change
    create_table :twitter_users do |t|
      t.integer :work_id
      t.string :screen_name, null: false
      t.string :user_id
      t.string :aasm_state, null: false, default: "published"
      t.datetime :followed_at
      t.timestamps null: false
    end

    add_index :twitter_users, :work_id
    add_index :twitter_users, :screen_name, unique: true
    add_index :twitter_users, :user_id, unique: true

    add_foreign_key :twitter_users, :works
  end
end
